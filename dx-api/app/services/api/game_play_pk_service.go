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
	OpponentID   string `json:"opponent_id"`
	OpponentName string `json:"opponent_name"`
}

// PkPlayerCompleteEvent is the SSE payload for pk_player_complete.
type PkPlayerCompleteEvent struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	GameLevelID string `json:"game_level_id"`
	Score       int    `json:"score"`
}

// PkPlayerActionEvent is the SSE payload for pk_player_action.
type PkPlayerActionEvent struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Action      string `json:"action"`
	ComboStreak int    `json:"combo_streak,omitempty"`
}

// PkNextLevelEvent is the SSE payload for pk_next_level.
type PkNextLevelEvent struct {
	GameLevelID string `json:"game_level_id"`
}

// PkTimeoutEvent is the SSE payload for pk_timeout / pk_timeout_warning.
type PkTimeoutEvent struct {
	GameLevelID string `json:"game_level_id"`
	SecondsLeft int    `json:"seconds_left"`
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

// StartPk starts a new PK match against a robot opponent.
func StartPk(userID, gameID, degree string, pattern *string, levelID *string, difficulty string) (*PkStartResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	// Verify game exists and is published
	var game models.Game
	if err := query.Where("id", gameID).First(&game); err != nil || game.ID == "" {
		return nil, ErrGameNotFound
	}
	if game.Status != "published" {
		return nil, ErrGameNotPublished
	}

	// Find first active level
	var firstLevel models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&firstLevel); err != nil || firstLevel.ID == "" {
		return nil, ErrNoGameLevels
	}

	// Find or create mock user
	mockUser, err := FindOrCreateMockUser()
	if err != nil {
		return nil, fmt.Errorf("failed to find mock user: %w", err)
	}

	// Resolve starting level
	resolvedLevelID := &firstLevel.ID
	if levelID != nil {
		resolvedLevelID = levelID
	}

	totalLevelsCount, err := query.Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count levels: %w", err)
	}

	// Create game_pks record
	pkID := newID()
	pk := models.GamePk{
		ID:              pkID,
		UserID:          userID,
		OpponentID:      mockUser.ID,
		GameID:          gameID,
		Degree:          degree,
		Pattern:         pattern,
		RobotDifficulty: difficulty,
		CurrentLevelID:  resolvedLevelID,
		IsPlaying:       true,
	}
	if err := query.Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create pk record: %w", err)
	}

	// Create human's session total
	now := time.Now()
	session := models.GameSessionTotal{
		ID:               newID(),
		UserID:           userID,
		GameID:           gameID,
		CurrentLevelID:   resolvedLevelID,
		Degree:           degree,
		Pattern:          pattern,
		TotalLevelsCount: int(totalLevelsCount),
		StartedAt:        now,
		LastPlayedAt:     now,
		GamePkID:         &pkID,
	}
	if err := query.Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Upsert game stats
	if err := UpsertGameStats(userID, gameID); err != nil {
		return nil, fmt.Errorf("failed to upsert game stats: %w", err)
	}

	opponentName := mockUser.Username
	if mockUser.Nickname != nil && *mockUser.Nickname != "" {
		opponentName = *mockUser.Nickname
	}

	return &PkStartResult{
		PkID:         pkID,
		SessionID:    session.ID,
		OpponentID:   mockUser.ID,
		OpponentName: opponentName,
	}, nil
}

// StartPkLevel creates a level session and spawns the robot goroutine.
func StartPkLevel(userID, sessionID, gameLevelID, degree string, pattern *string) (*StartLevelResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	// Verify session has GamePkID
	var parentSession models.GameSessionTotal
	if err := query.Where("id", sessionID).First(&parentSession); err != nil || parentSession.ID == "" {
		return nil, ErrSessionNotFound
	}
	if parentSession.GamePkID == nil {
		return nil, ErrPkNotFound
	}

	// Upsert level stats
	if err := UpsertLevelStats(userID, gameLevelID); err != nil {
		return nil, fmt.Errorf("failed to upsert level stats: %w", err)
	}

	// Count content items with degree filtering
	contentTypes := consts.DegreeContentTypes[degree]
	totalItemsCount, err := countActiveContentItems(gameLevelID, contentTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to count content items: %w", err)
	}

	// Check for existing incomplete level session
	var existing models.GameSessionLevel
	if err := query.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&existing); err == nil && existing.ID != "" {
		if _, err := query.Model(&models.GameSessionLevel{}).Where("id", existing.ID).
			Update("last_played_at", time.Now()); err != nil {
			return nil, fmt.Errorf("failed to touch level session: %w", err)
		}
		return &StartLevelResult{
			ID:                   existing.ID,
			GameSessionTotalID:   existing.GameSessionTotalID,
			GameLevelID:          existing.GameLevelID,
			CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	// Create human's level session
	now := time.Now()
	levelSession := models.GameSessionLevel{
		ID:                 newID(),
		GameSessionTotalID: sessionID,
		GameLevelID:        gameLevelID,
		Degree:             degree,
		Pattern:            pattern,
		TotalItemsCount:    int(totalItemsCount),
		StartedAt:          now,
		LastPlayedAt:       now,
		GamePkID:           parentSession.GamePkID,
	}
	if err := query.Create(&levelSession); err != nil {
		return nil, fmt.Errorf("failed to create level session: %w", err)
	}

	// Update session's current_level_id
	if _, err := query.Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update("current_level_id", gameLevelID); err != nil {
		return nil, fmt.Errorf("failed to update current level: %w", err)
	}

	// Fetch PK record for robot params
	pkID := *parentSession.GamePkID
	var pk models.GamePk
	if err := query.Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return nil, ErrPkNotFound
	}

	// Spawn robot goroutine
	go spawnRobotForLevel(pkID, pk.OpponentID, pk.GameID, gameLevelID, degree, pattern, pk.RobotDifficulty, int(totalItemsCount))

	return &StartLevelResult{
		ID:                   levelSession.ID,
		GameSessionTotalID:   levelSession.GameSessionTotalID,
		GameLevelID:          levelSession.GameLevelID,
		CurrentContentItemID: nil,
	}, nil
}

// CompletePkLevel marks a level complete for the human player and checks for winner.
func CompletePkLevel(userID, sessionID, gameLevelID string, score, maxCombo, totalItems int) (*CompleteLevelResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Find active session level
	var sessionLevel models.GameSessionLevel
	if err := tx.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&sessionLevel); err != nil || sessionLevel.ID == "" {
		_ = tx.Rollback()
		return nil, ErrSessionLevelNotFound
	}

	// Find parent session
	var session models.GameSessionTotal
	if err := tx.Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		_ = tx.Rollback()
		return nil, ErrSessionNotFound
	}

	// Calculate accuracy and EXP
	var accuracy float64
	if totalItems > 0 {
		accuracy = float64(sessionLevel.CorrectCount) / float64(totalItems)
	}
	meetsThreshold := accuracy >= consts.ExpAccuracyThreshold
	expAmount := 0
	if meetsThreshold {
		expAmount = consts.LevelCompleteExp
	}

	// 1. Complete session level
	now := time.Now()
	if _, err := tx.Model(&models.GameSessionLevel{}).Where("id", sessionLevel.ID).
		Update(map[string]any{
			"ended_at":  now,
			"score":     score,
			"exp":       expAmount,
			"max_combo": maxCombo,
		}); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to complete session level: %w", err)
	}

	// 2. Update session total
	if meetsThreshold {
		if _, err := tx.Exec(
			"UPDATE game_session_totals SET score = ?, max_combo = ?, played_levels_count = played_levels_count + 1, exp = exp + ?, updated_at = now() WHERE id = ?",
			score, maxCombo, expAmount, sessionID,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to update session total: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			"UPDATE game_session_totals SET score = ?, max_combo = ?, played_levels_count = played_levels_count + 1, updated_at = now() WHERE id = ?",
			score, maxCombo, sessionID,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to update session total: %w", err)
		}
	}

	// 3. Complete level stats
	if err := completeLevelStatsInTx(tx, userID, gameLevelID, score, sessionLevel.PlayTime); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to complete level stats: %w", err)
	}

	// 4. Update game stats
	if err := updateGameStatsOnLevelCompleteInTx(tx, userID, session.GameID, score, sessionLevel.PlayTime, expAmount); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to update game stats: %w", err)
	}

	// 5. Increment user EXP if threshold met
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

	// Broadcast player complete
	if session.GamePkID != nil {
		pkID := *session.GamePkID
		var user models.User
		if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user); err == nil && user.ID != "" {
			userName := user.Username
			if user.Nickname != nil && *user.Nickname != "" {
				userName = *user.Nickname
			}
			helpers.PkHub.Broadcast(pkID, "pk_player_complete", PkPlayerCompleteEvent{
				UserID:      userID,
				UserName:    userName,
				GameLevelID: gameLevelID,
				Score:       score,
			})
		}

		// Check for winner
		result, winErr := DeterminePkWinner(pkID, gameLevelID)
		if winErr != nil {
			fmt.Printf("[PK] Winner check error for pk=%s level=%s: %v\n", pkID, gameLevelID, winErr)
		}
		if winErr == nil && result != nil {
			helpers.PkHub.Broadcast(pkID, "pk_level_complete", result)

			// Cancel robot timeout since both players finished
			if rs := getRobotState(pkID); rs != nil {
				rs.cancel()
			}
		}

		// Check if all levels completed
		totalLevels, countErr := facades.Orm().Query().Model(&models.GameLevel{}).
			Where("game_id", session.GameID).Where("is_active", true).Count()
		if countErr == nil && totalLevels > 0 && int64(session.PlayedLevelsCount+1) >= totalLevels {
			facades.Orm().Query().Model(&models.GamePk{}).
				Where("id", pkID).Update("is_playing", false)
		}
	}

	return &CompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
	}, nil
}

// NextPkLevel advances the PK match to the next level.
func NextPkLevel(userID, pkID, currentLevelID string) error {
	query := facades.Orm().Query()

	var pk models.GamePk
	if err := query.Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.UserID != userID {
		return ErrForbidden
	}
	if !pk.IsPlaying {
		return ErrPkNotPlaying
	}

	// Find next level by order
	var currentLevel models.GameLevel
	if err := query.Where("id", currentLevelID).First(&currentLevel); err != nil || currentLevel.ID == "" {
		return ErrLevelNotFound
	}

	var nextLevel models.GameLevel
	if err := query.Where("game_id", pk.GameID).Where("is_active", true).
		Where("\"order\" > ?", currentLevel.Order).
		Order("\"order\" asc").First(&nextLevel); err != nil || nextLevel.ID == "" {
		return ErrLastLevel
	}

	// Update current level
	if _, err := query.Model(&models.GamePk{}).Where("id", pkID).
		Update("current_level_id", nextLevel.ID); err != nil {
		return fmt.Errorf("failed to update pk level: %w", err)
	}

	helpers.PkHub.Broadcast(pkID, "pk_next_level", PkNextLevelEvent{
		GameLevelID: nextLevel.ID,
	})

	return nil
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

	if _, err := query.Model(&models.GamePk{}).Where("id", pkID).
		Update("is_playing", false); err != nil {
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

// PkRecordAnswer records an answer in a PK match.
func PkRecordAnswer(userID string, input RecordAnswerInput) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return RecordAnswer(userID, input)
}

// PkRecordSkip records a skip in a PK match.
func PkRecordSkip(userID string, input RecordSkipInput) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return RecordSkip(userID, input)
}

// PkSyncPlayTime syncs playtime in a PK match.
func PkSyncPlayTime(userID, sessionID, gameLevelID string, playTime int) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return SyncPlayTime(userID, sessionID, gameLevelID, playTime)
}

// PkRestoreSessionData restores session data for a PK match.
func PkRestoreSessionData(userID, sessionID, gameLevelID string) (*SessionRestoreData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	return RestoreSessionData(userID, sessionID, gameLevelID)
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

	// Find or create robot's session total
	var robotSession models.GameSessionTotal
	if err := query.Where("user_id", robotUserID).Where("game_pk_id", pkID).
		Where("ended_at IS NULL").First(&robotSession); err != nil || robotSession.ID == "" {
		totalLevels, _ := query.Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count()
		now := time.Now()
		robotSession = models.GameSessionTotal{
			ID:               newID(),
			UserID:           robotUserID,
			GameID:           gameID,
			CurrentLevelID:   &gameLevelID,
			Degree:           degree,
			Pattern:          pattern,
			TotalLevelsCount: int(totalLevels),
			StartedAt:        now,
			LastPlayedAt:     now,
			GamePkID:         &pkID,
		}
		if err := query.Create(&robotSession); err != nil {
			fmt.Printf("[PK] Failed to create robot session for pk=%s: %v\n", pkID, err)
			return
		}
	}

	// Create robot's level session
	now := time.Now()
	robotLevelSession := models.GameSessionLevel{
		ID:                 newID(),
		GameSessionTotalID: robotSession.ID,
		GameLevelID:        gameLevelID,
		Degree:             degree,
		Pattern:            pattern,
		TotalItemsCount:    totalItems,
		StartedAt:          now,
		LastPlayedAt:       now,
		GamePkID:           &pkID,
	}
	if err := query.Create(&robotLevelSession); err != nil {
		fmt.Printf("[PK] Failed to create robot level session for pk=%s: %v\n", pkID, err)
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

		// Write game record via upsert
		record := models.GameRecord{
			ID:                 newID(),
			UserID:             robotUserID,
			GameSessionTotalID: robotSession.ID,
			GameSessionLevelID: robotLevelSession.ID,
			GameLevelID:        gameLevelID,
			ContentItemID:      item.ID,
			IsCorrect:          isCorrect,
			SourceAnswer:       item.Content,
			UserAnswer:         item.Content,
			BaseScore:          consts.CorrectAnswer,
			ComboScore:         result.ComboBonus,
			Duration:           delayMs / 1000,
		}
		if !isCorrect {
			record.BaseScore = 0
			record.ComboScore = 0
		}
		facades.Orm().Query().Create(&record)

		// Update robot's level session stats
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
				fmt.Sprintf("UPDATE game_session_levels SET score = ?, max_combo = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = ?, updated_at = now() WHERE id = ?", countCol),
				combo.TotalScore, combo.MaxCombo, *nextItemID, robotLevelSession.ID,
			)
		} else {
			facades.Orm().Query().Exec(
				fmt.Sprintf("UPDATE game_session_levels SET score = ?, max_combo = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, updated_at = now() WHERE id = ?", countCol),
				combo.TotalScore, combo.MaxCombo, robotLevelSession.ID,
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

	// Robot finished all items — complete the level
	robotAccuracy := float64(0)
	if totalItems > 0 {
		robotAccuracy = float64(correctCount) / float64(totalItems)
	}
	robotExp := 0
	if robotAccuracy >= consts.ExpAccuracyThreshold {
		robotExp = consts.LevelCompleteExp
	}

	endNow := time.Now()
	facades.Orm().Query().Model(&models.GameSessionLevel{}).Where("id", robotLevelSession.ID).
		Update(map[string]any{
			"ended_at":  endNow,
			"score":     combo.TotalScore,
			"exp":       robotExp,
			"max_combo": combo.MaxCombo,
		})

	// Update robot's session total
	facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET score = ?, max_combo = ?, played_levels_count = played_levels_count + 1, updated_at = now() WHERE id = ?",
		combo.TotalScore, combo.MaxCombo, robotSession.ID,
	)

	// Broadcast robot completion
	helpers.PkHub.Broadcast(pkID, "pk_player_complete", PkPlayerCompleteEvent{
		UserID:      robotUserID,
		UserName:    robotName,
		GameLevelID: gameLevelID,
		Score:       combo.TotalScore,
	})

	// Check if human already finished
	winResult, winErr := DeterminePkWinner(pkID, gameLevelID)
	if winErr == nil && winResult != nil {
		helpers.PkHub.Broadcast(pkID, "pk_level_complete", winResult)
		return
	}

	// Human hasn't finished — start timeout
	waitDuration := time.Duration(consts.PkTimeoutDuration-consts.PkTimeoutWarning) * time.Second
	select {
	case <-ctx.Done():
		return
	case <-time.After(waitDuration):
	}

	// Broadcast timeout warning
	helpers.PkHub.Broadcast(pkID, "pk_timeout_warning", PkTimeoutEvent{
		GameLevelID: gameLevelID,
		SecondsLeft: consts.PkTimeoutWarning,
	})

	// Wait remaining seconds
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(consts.PkTimeoutWarning) * time.Second):
	}

	// Timeout — auto-end human's level
	helpers.PkHub.Broadcast(pkID, "pk_timeout", PkTimeoutEvent{
		GameLevelID: gameLevelID,
		SecondsLeft: 0,
	})

	// Force-complete human's active level session for this level
	autoEndHumanLevel(pkID, gameLevelID)
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
		"UPDATE game_session_levels SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL",
		now, pkID,
	)
	facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL",
		now, pkID,
	)
}

// autoEndHumanLevel force-completes the human's active level on timeout.
func autoEndHumanLevel(pkID, gameLevelID string) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return
	}

	// Find the human's active level session
	var levelSession models.GameSessionLevel
	err := facades.Orm().Query().Raw(
		`SELECT gsl.* FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 WHERE gsl.game_pk_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NULL
		   AND gst.user_id = ?
		 LIMIT 1`,
		pkID, gameLevelID, pk.UserID).Scan(&levelSession)
	if err != nil || levelSession.ID == "" {
		return
	}

	now := time.Now()
	facades.Orm().Query().Model(&models.GameSessionLevel{}).Where("id", levelSession.ID).
		Update(map[string]any{
			"ended_at": now,
			"score":    levelSession.Score,
		})

	// Update session total played_levels_count
	facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET played_levels_count = played_levels_count + 1, updated_at = now() WHERE id = ?",
		levelSession.GameSessionTotalID,
	)

	// Determine winner after auto-end
	winResult, winErr := DeterminePkWinner(pkID, gameLevelID)
	if winErr == nil && winResult != nil {
		helpers.PkHub.Broadcast(pkID, "pk_level_complete", winResult)
	}
}
