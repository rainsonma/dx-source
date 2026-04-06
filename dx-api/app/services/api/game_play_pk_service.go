package api

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// --- Result types ---

// PkStartResult is returned after starting a PK match.
type PkStartResult struct {
	PkID         string `json:"pk_id"`
	SessionID    string `json:"session_id"`
	GameLevelID  string `json:"game_level_id"`
	OpponentID   string `json:"opponent_id"`
	OpponentName string `json:"opponent_name"`
}

// PkPlayerCompleteEvent is the SSE payload for pk_player_complete.
type PkPlayerCompleteEvent struct {
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	GameLevelID   string  `json:"game_level_id"`
	Score         int     `json:"score"`
	NextLevelID   *string `json:"next_level_id"`
	NextLevelName *string `json:"next_level_name"`
}

// PkPlayerActionEvent is the SSE payload for pk_player_action.
type PkPlayerActionEvent struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Action      string `json:"action"`
	ComboStreak int    `json:"combo_streak,omitempty"`
}


// --- Robot state management ---

type robotState struct {
	cancel  context.CancelFunc
	pauseCh chan struct{}
	paused  bool
	mu      sync.Mutex
}

var (
	robotStates   = make(map[string]*robotState)
	robotStatesMu sync.Mutex
)

func getRobotState(pkID string) *robotState {
	robotStatesMu.Lock()
	defer robotStatesMu.Unlock()
	return robotStates[pkID]
}

func setRobotState(pkID string, state *robotState) {
	robotStatesMu.Lock()
	defer robotStatesMu.Unlock()
	robotStates[pkID] = state
}

func deleteRobotState(pkID string) {
	robotStatesMu.Lock()
	defer robotStatesMu.Unlock()
	delete(robotStates, pkID)
}

// --- PK Lifecycle ---

// StartPk starts a new PK match against a robot opponent for a single level.
func StartPk(userID, gameID, gameLevelID, degree string, pattern *string, difficulty string) (*PkStartResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	// Check for existing active PK for this user/game/level (idempotent for concurrent calls)
	var existingPk models.GamePk
	query.Where("user_id", userID).Where("game_id", gameID).
		Where("game_level_id", gameLevelID).Where("is_playing", true).First(&existingPk)
	if existingPk.ID != "" {
		var existingSession models.GameSession
		query.Where("game_pk_id", existingPk.ID).Where("user_id", userID).First(&existingSession)
		if existingSession.ID != "" {
			var opponent models.User
			query.Where("id", existingPk.OpponentID).First(&opponent)
			opName := opponent.Username
			if opponent.Nickname != nil && *opponent.Nickname != "" {
				opName = *opponent.Nickname
			}
			return &PkStartResult{
				PkID:         existingPk.ID,
				SessionID:    existingSession.ID,
				GameLevelID:  existingPk.GameLevelID,
				OpponentID:   existingPk.OpponentID,
				OpponentName: opName,
			}, nil
		}
	}

	// Verify game exists and is published
	var game models.Game
	if err := query.Where("id", gameID).First(&game); err != nil || game.ID == "" {
		return nil, ErrGameNotFound
	}
	if game.Status != "published" {
		return nil, ErrGameNotPublished
	}

	// Find or create mock user
	mockUser, err := FindOrCreateMockUser()
	if err != nil {
		return nil, fmt.Errorf("failed to find mock user: %w", err)
	}

	// Count content items for this level
	totalItemsCount, err := countLevelItems(query, gameLevelID, degree)
	if err != nil {
		return nil, fmt.Errorf("failed to count content items: %w", err)
	}

	// Create game_pks record
	pkID := newID()
	pk := models.GamePk{
		ID:              pkID,
		UserID:          userID,
		OpponentID:      mockUser.ID,
		GameID:          gameID,
		GameLevelID:     gameLevelID,
		Degree:          degree,
		Pattern:         pattern,
		RobotDifficulty: difficulty,
		IsPlaying:       true,
		PkType:          consts.PkTypeRandom,
	}
	if err := query.Create(&pk); err != nil {
		// Unique constraint violation — concurrent call already created a PK
		var fallback models.GamePk
		query.Where("user_id", userID).Where("game_id", gameID).
			Where("game_level_id", gameLevelID).Where("is_playing", true).First(&fallback)
		if fallback.ID != "" {
			var fbSession models.GameSession
			query.Where("game_pk_id", fallback.ID).Where("user_id", userID).First(&fbSession)
			if fbSession.ID != "" {
				var fbOpponent models.User
				query.Where("id", fallback.OpponentID).First(&fbOpponent)
				fbName := fbOpponent.Username
				if fbOpponent.Nickname != nil && *fbOpponent.Nickname != "" {
					fbName = *fbOpponent.Nickname
				}
				return &PkStartResult{
					PkID:         fallback.ID,
					SessionID:    fbSession.ID,
					GameLevelID:  fallback.GameLevelID,
					OpponentID:   fallback.OpponentID,
					OpponentName: fbName,
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to create pk record: %w", err)
	}

	// Create human's game session
	now := time.Now()
	session := models.GameSession{
		ID:              newID(),
		UserID:          userID,
		GameID:          gameID,
		GameLevelID:     gameLevelID,
		Degree:          degree,
		Pattern:         pattern,
		TotalItemsCount: int(totalItemsCount),
		StartedAt:       now,
		LastPlayedAt:    now,
		GamePkID:        &pkID,
	}
	if err := query.Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	opponentName := mockUser.Username
	if mockUser.Nickname != nil && *mockUser.Nickname != "" {
		opponentName = *mockUser.Nickname
	}

	// Spawn robot goroutine for this level
	go spawnRobotForLevel(pkID, mockUser.ID, gameID, gameLevelID, degree, pattern, difficulty, int(totalItemsCount))

	return &PkStartResult{
		PkID:         pkID,
		SessionID:    session.ID,
		GameLevelID:  gameLevelID,
		OpponentID:   mockUser.ID,
		OpponentName: opponentName,
	}, nil
}

// CompletePk marks the level complete for the human player. First-to-complete wins.
func CompletePk(userID, sessionID string, score, maxCombo, totalItems int) (*CompleteLevelResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	// Find the active session
	var session models.GameSession
	if err := query.Where("id", sessionID).Where("ended_at IS NULL").
		First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}
	if session.GamePkID == nil {
		return nil, ErrPkNotFound
	}
	pkID := *session.GamePkID

	// Calculate accuracy and EXP
	var accuracy float64
	if totalItems > 0 {
		accuracy = float64(session.CorrectCount) / float64(totalItems)
	}
	meetsThreshold := accuracy >= consts.ExpAccuracyThreshold
	expAmount := 0
	if meetsThreshold {
		expAmount = consts.LevelCompleteExp
	}

	tx, err := query.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Set ended_at on winner's session
	now := time.Now()
	if _, err := tx.Model(&models.GameSession{}).Where("id", sessionID).
		Update(map[string]any{
			"ended_at":  now,
			"score":     score,
			"exp":       expAmount,
			"max_combo": maxCombo,
		}); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	// 2. Increment user EXP if threshold met
	if meetsThreshold {
		if _, err := tx.Exec(
			"UPDATE users SET exp = exp + ?, updated_at = now() WHERE id = ?",
			expAmount, userID,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to increment user exp: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Find next level for result panel
	nextLevelID, nextLevelName, _ := findNextLevel(session.GameID, session.GameLevelID)

	// Broadcast player complete via SSE
	var user models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user); err == nil && user.ID != "" {
		userName := user.Username
		if user.Nickname != nil && *user.Nickname != "" {
			userName = *user.Nickname
		}
		helpers.PkHub.Broadcast(pkID, "pk_player_complete", PkPlayerCompleteEvent{
			UserID:        userID,
			UserName:      userName,
			GameLevelID:   session.GameLevelID,
			Score:         score,
			NextLevelID:   nextLevelID,
			NextLevelName: nextLevelName,
		})
	}

	// First-to-complete wins: force-end opponent's session
	if err := ForceEndPkLoser(pkID, userID); err != nil {
		fmt.Printf("[PK] Failed to force-end loser for pk=%s: %v\n", pkID, err)
	}

	// Set PK as finished and record winner
	facades.Orm().Query().Model(&models.GamePk{}).Where("id", pkID).
		Update(map[string]any{
			"is_playing":     false,
			"last_winner_id": userID,
		})

	// Cancel robot goroutine since human won
	cancelRobot(pkID)

	// nextLevelID and nextLevelName already computed above for SSE broadcast

	return &CompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
		NextLevelID:    nextLevelID,
		NextLevelName:  nextLevelName,
	}, nil
}

// NextPkLevel creates a new PK for the next level.
func NextPkLevel(userID, pkID string) (*PkStartResult, error) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return nil, ErrPkNotFound
	}

	nextLevelID, _, err := findNextLevel(pk.GameID, pk.GameLevelID)
	if err != nil || nextLevelID == nil {
		return nil, fmt.Errorf("no next level available")
	}

	// For specified PK, create a new PK directly without robot
	if pk.PkType == consts.PkTypeSpecified {
		return nextSpecifiedPkLevel(userID, pk, *nextLevelID)
	}

	return StartPk(userID, pk.GameID, *nextLevelID, pk.Degree, pk.Pattern, pk.RobotDifficulty)
}

// nextSpecifiedPkLevel creates a new specified PK for the next level (no robot, no re-invitation).
func nextSpecifiedPkLevel(_ string, oldPk models.GamePk, nextLevelID string) (*PkStartResult, error) {
	pkID := newID()
	statusAccepted := consts.PkInvitationAccepted

	pk := models.GamePk{
		ID:               pkID,
		UserID:           oldPk.UserID,
		OpponentID:       oldPk.OpponentID,
		GameID:           oldPk.GameID,
		GameLevelID:      nextLevelID,
		Degree:           oldPk.Degree,
		Pattern:          oldPk.Pattern,
		RobotDifficulty:  "",
		IsPlaying:        true,
		PkType:           consts.PkTypeSpecified,
		InvitationStatus: &statusAccepted,
	}
	if err := facades.Orm().Query().Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create next PK: %w", err)
	}

	// Create sessions for both players
	now := time.Now()
	totalItems, _ := countLevelItems(facades.Orm().Query(), nextLevelID, oldPk.Degree)

	for _, uid := range []string{oldPk.UserID, oldPk.OpponentID} {
		sid := newID()
		session := models.GameSession{
			ID:              sid,
			UserID:          uid,
			GameID:          oldPk.GameID,
			GameLevelID:     nextLevelID,
			Degree:          oldPk.Degree,
			Pattern:         oldPk.Pattern,
			GamePkID:        &pkID,
			StartedAt:       now,
			TotalItemsCount: int(totalItems),
		}
		if err := facades.Orm().Query().Create(&session); err != nil {
			return nil, fmt.Errorf("failed to create session for %s: %w", uid, err)
		}
	}

	var opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", oldPk.OpponentID).First(&opponent)

	return &PkStartResult{
		PkID:         pkID,
		SessionID:    "",
		GameLevelID:  nextLevelID,
		OpponentID:   oldPk.OpponentID,
		OpponentName: nickname(opponent),
	}, nil
}

// EndPk forcefully ends a PK match.
func EndPk(userID, pkID string) error {
	query := facades.Orm().Query()

	var pk models.GamePk
	if err := query.Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.UserID != userID {
		return ErrForbidden
	}

	cancelRobot(pkID)
	endPkSessions(pkID)

	updates := map[string]any{"is_playing": false}
	if pk.PkType == consts.PkTypeSpecified && pk.InvitationStatus != nil && *pk.InvitationStatus == consts.PkInvitationPending {
		expired := consts.PkInvitationExpired
		updates["invitation_status"] = expired
	}

	if _, err := query.Model(&models.GamePk{}).Where("id", pkID).
		Update(updates); err != nil {
		return fmt.Errorf("failed to end pk: %w", err)
	}

	helpers.PkHub.Broadcast(pkID, "pk_force_end", map[string]string{"pk_id": pkID})

	return nil
}

// PausePkRobot pauses the robot goroutine.
func PausePkRobot(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.UserID != userID {
		return ErrForbidden
	}

	rs := getRobotState(pkID)
	if rs == nil {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if !rs.paused {
		rs.paused = true
		rs.pauseCh = make(chan struct{})
	}
	return nil
}

// ResumePkRobot resumes the robot goroutine.
func ResumePkRobot(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.UserID != userID {
		return ErrForbidden
	}

	rs := getRobotState(pkID)
	if rs == nil {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.paused {
		rs.paused = false
		close(rs.pauseCh)
	}
	return nil
}

// OnPkDisconnect handles cleanup when the human player's SSE connection drops.
func OnPkDisconnect(pkID string) {
	cancelRobot(pkID)
	endPkSessions(pkID)

	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("is_playing", false)
}

// --- Thin wrappers ---

// PkRecordAnswer records an answer in a PK match and broadcasts the action to the opponent.
func PkRecordAnswer(userID string, input RecordAnswerInput) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	if err := RecordAnswer(userID, input); err != nil {
		return err
	}

	// Broadcast pk_player_action so the opponent sees progress updates.
	// Look up the PK ID from the session (needed for the broadcast channel).
	var session models.GameSession
	facades.Orm().Query().Select("game_pk_id").Where("id", input.GameSessionID).First(&session)
	if session.GamePkID != nil {
		var user models.User
		facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user)
		action := "score"
		if !input.IsCorrect {
			action = "wrong"
		}
		go helpers.PkHub.Broadcast(*session.GamePkID, "pk_player_action", PkPlayerActionEvent{
			UserID:      userID,
			UserName:    nickname(user),
			Action:      action,
			ComboStreak: input.MaxCombo,
		})
	}
	return nil
}

// PkSyncPlayTime syncs playtime in a PK match.
func PkSyncPlayTime(userID, sessionID string, playTime int) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return SyncPlayTime(userID, sessionID, playTime)
}

// PkRestoreSessionData restores session data for a PK match.
func PkRestoreSessionData(userID, sessionID string) (*SessionRestoreData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	return RestoreSessionData(userID, sessionID)
}

// PkUpdateContentItem updates the current content item in a PK session.
func PkUpdateContentItem(userID, sessionID string, contentItemID *string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return UpdateCurrentContentItem(userID, sessionID, contentItemID)
}

// --- Robot goroutine ---

func spawnRobotForLevel(pkID, robotUserID, gameID, gameLevelID, degree string, pattern *string, difficulty string, totalItems int) {
	ctx, cancel := context.WithCancel(context.Background())

	rs := &robotState{
		cancel:  cancel,
		pauseCh: make(chan struct{}),
		paused:  false,
	}
	// Close the initial pauseCh so the robot doesn't block on first check
	close(rs.pauseCh)
	setRobotState(pkID, rs)

	defer func() {
		cancel()
		deleteRobotState(pkID)
	}()

	query := facades.Orm().Query()

	// Create robot's game session
	now := time.Now()
	robotSession := models.GameSession{
		ID:              newID(),
		UserID:          robotUserID,
		GameID:          gameID,
		GameLevelID:     gameLevelID,
		Degree:          degree,
		Pattern:         pattern,
		TotalItemsCount: totalItems,
		StartedAt:       now,
		LastPlayedAt:    now,
		GamePkID:        &pkID,
	}
	if err := query.Create(&robotSession); err != nil {
		fmt.Printf("[PK] Failed to create robot session for pk=%s: %v\n", pkID, err)
		return
	}

	// Fetch content items for the level
	contentTypes := consts.DegreeContentTypes[degree]
	contentQuery := query.Model(&models.ContentItem{}).Where("game_level_id", gameLevelID).Where("is_active", true)
	if len(contentTypes) > 0 {
		contentQuery = contentQuery.Where("content_type IN ?", contentTypes)
	}
	var items []models.ContentItem
	if err := contentQuery.Order("\"order\" asc").Get(&items); err != nil || len(items) == 0 {
		fmt.Printf("[PK] No content items for pk=%s level=%s\n", pkID, gameLevelID)
		return
	}

	// Get difficulty params
	params, ok := consts.PkDifficulties[difficulty]
	if !ok {
		params = consts.PkDifficulties[consts.PkDifficultyNormal]
	}

	// Roll accuracy for this level
	accuracy := params.AccuracyMin + rand.Float64()*(params.AccuracyMax-params.AccuracyMin)

	// Robot user info for SSE broadcasts
	var robotUser models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", robotUserID).First(&robotUser)
	robotName := robotUser.Username
	if robotUser.Nickname != nil && *robotUser.Nickname != "" {
		robotName = *robotUser.Nickname
	}

	// Simulate answering items
	combo := helpers.ComboState{}
	correctCount := 0
	for i, item := range items {
		// Check pause
		rs.mu.Lock()
		pauseCh := rs.pauseCh
		rs.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-pauseCh:
			// Not paused or just resumed — continue
		}

		// Check context again after pause
		if ctx.Err() != nil {
			return
		}

		// Random delay
		delayMs := params.MinDelayMs + rand.IntN(params.MaxDelayMs-params.MinDelayMs+1)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delayMs) * time.Millisecond):
		}

		// Roll correctness
		isCorrect := rand.Float64() < accuracy
		// Apply combo break chance
		if isCorrect && combo.Streak >= 3 && rand.Float64() < params.ComboBreakPct {
			isCorrect = false
		}

		// Process answer for scoring
		result := helpers.ProcessAnswer(combo, isCorrect)
		combo = result.State
		if isCorrect {
			correctCount++
		}

		// Write game record
		record := models.GameRecord{
			ID:            newID(),
			UserID:        robotUserID,
			GameSessionID: robotSession.ID,
			GameLevelID:   gameLevelID,
			ContentItemID: item.ID,
			IsCorrect:     isCorrect,
			SourceAnswer:  item.Content,
			UserAnswer:    item.Content,
			BaseScore:     consts.CorrectAnswer,
			ComboScore:    result.ComboBonus,
			Duration:      delayMs / 1000,
		}
		if !isCorrect {
			record.BaseScore = 0
			record.ComboScore = 0
		}
		facades.Orm().Query().Create(&record)

		// Update robot's session stats
		var countCol string
		if isCorrect {
			countCol = "correct_count = correct_count + 1"
		} else {
			countCol = "wrong_count = wrong_count + 1"
		}

		var nextItemID *string
		if i+1 < len(items) {
			nextItemID = &items[i+1].ID
		}

		if nextItemID != nil {
			facades.Orm().Query().Exec(
				fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = ?, updated_at = now() WHERE id = ?", countCol),
				combo.TotalScore, combo.MaxCombo, *nextItemID, robotSession.ID,
			)
		} else {
			facades.Orm().Query().Exec(
				fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, updated_at = now() WHERE id = ?", countCol),
				combo.TotalScore, combo.MaxCombo, robotSession.ID,
			)
		}

		// Broadcast robot action
		action := "score"
		if !isCorrect {
			action = "wrong"
		}
		helpers.PkHub.Broadcast(pkID, "pk_player_action", PkPlayerActionEvent{
			UserID:      robotUserID,
			UserName:    robotName,
			Action:      action,
			ComboStreak: combo.Streak,
		})
	}

	// Robot finished all items — first-to-complete wins
	robotAccuracy := float64(0)
	if totalItems > 0 {
		robotAccuracy = float64(correctCount) / float64(totalItems)
	}
	robotExp := 0
	if robotAccuracy >= consts.ExpAccuracyThreshold {
		robotExp = consts.LevelCompleteExp
	}

	endNow := time.Now()
	facades.Orm().Query().Model(&models.GameSession{}).Where("id", robotSession.ID).
		Update(map[string]any{
			"ended_at":  endNow,
			"score":     combo.TotalScore,
			"exp":       robotExp,
			"max_combo": combo.MaxCombo,
		})

	// Broadcast robot completion
	robotNextLevelID, robotNextLevelName, _ := findNextLevel(gameID, gameLevelID)
	helpers.PkHub.Broadcast(pkID, "pk_player_complete", PkPlayerCompleteEvent{
		UserID:        robotUserID,
		UserName:      robotName,
		GameLevelID:   gameLevelID,
		Score:         combo.TotalScore,
		NextLevelID:   robotNextLevelID,
		NextLevelName: robotNextLevelName,
	})

	// Robot won — force-end human's session
	if err := ForceEndPkLoser(pkID, robotUserID); err != nil {
		fmt.Printf("[PK] Failed to force-end human for pk=%s: %v\n", pkID, err)
	}

	// Set PK as finished and record robot as winner
	facades.Orm().Query().Model(&models.GamePk{}).Where("id", pkID).
		Update(map[string]any{
			"is_playing":     false,
			"last_winner_id": robotUserID,
		})

}

// --- Internal helpers ---

// cancelRobot cancels the robot goroutine for a PK match.
func cancelRobot(pkID string) {
	rs := getRobotState(pkID)
	if rs != nil {
		rs.cancel()
	}
}

// endPkSessions ends all active sessions for a PK match.
func endPkSessions(pkID string) {
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_sessions SET ended_at = ?, updated_at = now() WHERE game_pk_id = ? AND ended_at IS NULL",
		now, pkID,
	)
}
