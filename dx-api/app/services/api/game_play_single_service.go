package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

func newID() string {
	return uuid.Must(uuid.NewV7()).String()
}

const (
	rateLimitAnswerKey  = "ratelimit:record-answer:%s"
	rateLimitSkipKey    = "ratelimit:record-skip:%s"
	rateLimitAnswer     = 30
	rateLimitSkip       = 30
	rateLimitWindowSecs = 60
)

// --- DTOs ---

// StartSessionResult is returned after starting or resuming a session.
type StartSessionResult struct {
	ID                   string    `json:"id"`
	Degree               string    `json:"degree"`
	Pattern              *string   `json:"pattern"`
	Score                int       `json:"score"`
	Exp                  int       `json:"exp"`
	MaxCombo             int       `json:"maxCombo"`
	CorrectCount         int       `json:"correctCount"`
	WrongCount           int       `json:"wrongCount"`
	StartedAt            time.Time `json:"startedAt"`
	LevelID              *string   `json:"levelId"`
	CurrentContentItemID *string   `json:"currentContentItemId"`
}

// ActiveSessionData is returned when checking for an active session.
type ActiveSessionData struct {
	ID                   string  `json:"id"`
	Degree               string  `json:"degree"`
	Pattern              *string `json:"pattern"`
	CurrentLevelID       *string `json:"currentLevelId"`
	CurrentContentItemID *string `json:"currentContentItemId"`
}

// ActiveLevelSessionData is returned when checking for an active level session.
type ActiveLevelSessionData struct {
	SessionID      string `json:"sessionId"`
	LevelSessionID string `json:"levelSessionId"`
}

// StartLevelResult is returned after starting or resuming a session level.
type StartLevelResult struct {
	ID                   string  `json:"id"`
	GameSessionTotalID   string  `json:"gameSessionTotalId"`
	GameLevelID          string  `json:"gameLevelId"`
	CurrentContentItemID *string `json:"currentContentItemId"`
}

// CompleteLevelResult is returned after completing a level.
type CompleteLevelResult struct {
	ExpEarned      int     `json:"expEarned"`
	Accuracy       float64 `json:"accuracy"`
	MeetsThreshold bool    `json:"meetsThreshold"`
}

// SessionRestoreData holds accumulated stats for restoring client state.
type SessionRestoreData struct {
	Session      *SessionStats `json:"session"`
	SessionLevel *SessionStats `json:"sessionLevel"`
}

// SessionStats holds score/combo/count/playtime fields.
type SessionStats struct {
	Score        int `json:"score"`
	MaxCombo     int `json:"maxCombo"`
	CorrectCount int `json:"correctCount"`
	WrongCount   int `json:"wrongCount"`
	SkipCount    int `json:"skipCount"`
	PlayTime     int `json:"playTime"`
}

// RecordAnswerInput holds the data needed to record an answer.
type RecordAnswerInput struct {
	GameSessionTotalID string
	GameSessionLevelID string
	GameLevelID        string
	ContentItemID      string
	IsCorrect          bool
	UserAnswer         string
	SourceAnswer       string
	BaseScore          int
	ComboScore         int
	Score              int
	MaxCombo           int
	PlayTime           int
	NextContentItemID  *string
	Duration           int
}

// RecordSkipInput holds the data needed to record a skip.
type RecordSkipInput struct {
	GameSessionTotalID string
	GameLevelID        string
	PlayTime           int
	NextContentItemID  *string
}

// EndSessionInput holds the data needed to end a session.
type EndSessionInput struct {
	GameID             string
	Score              int
	Exp                int
	MaxCombo           int
	CorrectCount       int
	WrongCount         int
	SkipCount          int
	AllLevelsCompleted bool
}

// --- Session Lifecycle ---

// StartSession starts or resumes a game session.
func StartSession(userID, gameID, degree string, pattern *string, levelID *string) (*StartSessionResult, error) {
	query := facades.Orm().Query()

	// Find the first active level
	var firstLevel models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&firstLevel); err != nil || firstLevel.ID == "" {
		return nil, ErrNoGameLevels
	}

	// VIP guard: non-first levels require active VIP
	targetLevelID := levelID
	if targetLevelID == nil {
		targetLevelID = &firstLevel.ID
	}
	if *targetLevelID != firstLevel.ID {
		if err := requireVip(userID); err != nil {
			return nil, err
		}
	}

	// Check for existing active session
	existing, err := findActiveSession(query, userID, gameID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}

	if existing != nil {
		// Touch lastPlayedAt
		if _, err := query.Model(&models.GameSessionTotal{}).Where("id", existing.ID).
			Update("last_played_at", time.Now()); err != nil {
			return nil, fmt.Errorf("failed to touch session: %w", err)
		}

		resolvedLevelID := existing.CurrentLevelID
		contentItemID := existing.CurrentContentItemID

		if levelID != nil && (existing.CurrentLevelID == nil || *levelID != *existing.CurrentLevelID) {
			if _, err := query.Model(&models.GameSessionTotal{}).Where("id", existing.ID).
				Update(map[string]any{
					"current_level_id":        *levelID,
					"current_content_item_id": nil,
				}); err != nil {
				return nil, fmt.Errorf("failed to update resume point: %w", err)
			}
			resolvedLevelID = levelID
			contentItemID = nil
		}

		return &StartSessionResult{
			ID:                   existing.ID,
			Degree:               existing.Degree,
			Pattern:              existing.Pattern,
			Score:                existing.Score,
			Exp:                  existing.Exp,
			MaxCombo:             existing.MaxCombo,
			CorrectCount:         existing.CorrectCount,
			WrongCount:           existing.WrongCount,
			StartedAt:            existing.StartedAt,
			LevelID:              resolvedLevelID,
			CurrentContentItemID: contentItemID,
		}, nil
	}

	// Create new session
	resolvedLevelID := &firstLevel.ID
	if levelID != nil {
		resolvedLevelID = levelID
	}

	totalLevelsCount, err := query.Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count levels: %w", err)
	}

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
	}

	if err := query.Create(&session); err != nil {
		// Unique constraint violation: concurrent request already created the session.
		existing, findErr := findActiveSession(query, userID, gameID, degree, pattern)
		if findErr != nil || existing == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		return &StartSessionResult{
			ID:                   existing.ID,
			Degree:               existing.Degree,
			Pattern:              existing.Pattern,
			Score:                existing.Score,
			Exp:                  existing.Exp,
			MaxCombo:             existing.MaxCombo,
			CorrectCount:         existing.CorrectCount,
			WrongCount:           existing.WrongCount,
			StartedAt:            existing.StartedAt,
			LevelID:              existing.CurrentLevelID,
			CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	// Upsert game stats only after successful session creation
	if err := UpsertGameStats(userID, gameID); err != nil {
		return nil, fmt.Errorf("failed to upsert game stats: %w", err)
	}

	return &StartSessionResult{
		ID:                   session.ID,
		Degree:               session.Degree,
		Pattern:              session.Pattern,
		StartedAt:            session.StartedAt,
		LevelID:              resolvedLevelID,
		CurrentContentItemID: nil,
	}, nil
}

// CheckActiveSession finds an active session for a specific degree+pattern combo.
func CheckActiveSession(userID, gameID, degree string, pattern *string) (*ActiveSessionData, error) {
	session, err := findActiveSession(facades.Orm().Query(), userID, gameID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if session == nil {
		return nil, nil
	}
	return &ActiveSessionData{
		ID:                   session.ID,
		Degree:               session.Degree,
		Pattern:              session.Pattern,
		CurrentLevelID:       session.CurrentLevelID,
		CurrentContentItemID: session.CurrentContentItemID,
	}, nil
}

// CheckAnyActiveSession finds any active single-play session for a game.
// Excludes group sessions.
func CheckAnyActiveSession(userID, gameID string) (*ActiveSessionData, error) {
	var session models.GameSessionTotal
	if err := facades.Orm().Query().Where("user_id", userID).Where("game_id", gameID).
		Where("ended_at IS NULL").Where("game_group_id IS NULL").
		Order("last_played_at desc").First(&session); err != nil || session.ID == "" {
		return nil, nil
	}
	return &ActiveSessionData{
		ID:                   session.ID,
		Degree:               session.Degree,
		Pattern:              session.Pattern,
		CurrentLevelID:       session.CurrentLevelID,
		CurrentContentItemID: session.CurrentContentItemID,
	}, nil
}

// CheckActiveLevelSession finds an active level session within an active game session.
func CheckActiveLevelSession(userID, gameID, degree string, pattern *string, gameLevelID string) (*ActiveLevelSessionData, error) {
	session, err := findActiveSession(facades.Orm().Query(), userID, gameID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if session == nil {
		return nil, nil
	}

	var levelSession models.GameSessionLevel
	if err := facades.Orm().Query().Where("game_session_total_id", session.ID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&levelSession); err != nil || levelSession.ID == "" {
		return nil, nil
	}

	return &ActiveLevelSessionData{
		SessionID:      session.ID,
		LevelSessionID: levelSession.ID,
	}, nil
}

// ForceCompleteSession marks a session as ended.
func ForceCompleteSession(userID, sessionID string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	now := time.Now()
	_, err := facades.Orm().Query().Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update("ended_at", now)
	return err
}

// RestartLevel ends the active level and resets the session's resume point.
func RestartLevel(userID, sessionID, gameLevelID string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	query := facades.Orm().Query()

	// End active level if exists
	var activeLevel models.GameSessionLevel
	if err := query.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&activeLevel); err == nil && activeLevel.ID != "" {
		now := time.Now()
		if _, err := query.Model(&models.GameSessionLevel{}).Where("id", activeLevel.ID).
			Update("ended_at", now); err != nil {
			return fmt.Errorf("failed to end level session: %w", err)
		}
	}

	// Reset resume point
	_, err := query.Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update(map[string]any{
			"current_level_id":        gameLevelID,
			"current_content_item_id": nil,
		})
	return err
}

// --- Answer / Skip Recording ---

// RecordAnswer records a single answer and updates session + level stats atomically.
func RecordAnswer(userID string, input RecordAnswerInput) error {
	if err := verifyOwnership(userID, input.GameSessionTotalID); err != nil {
		return err
	}

	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitAnswerKey, userID), rateLimitAnswer, rateLimitWindowSecs,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	safeDuration := input.Duration
	if safeDuration < 0 {
		safeDuration = 0
	}
	if safeDuration > 3600 {
		safeDuration = 3600
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Upsert game record
	var existingRecord models.GameRecord
	_ = tx.Where("game_session_level_id", input.GameSessionLevelID).
		Where("content_item_id", input.ContentItemID).First(&existingRecord)

	if existingRecord.ID == "" {
		record := models.GameRecord{
			ID:                 newID(),
			UserID:             userID,
			GameSessionTotalID: input.GameSessionTotalID,
			GameSessionLevelID: input.GameSessionLevelID,
			GameLevelID:        input.GameLevelID,
			ContentItemID:      input.ContentItemID,
			IsCorrect:          input.IsCorrect,
			UserAnswer:         input.UserAnswer,
			SourceAnswer:       input.SourceAnswer,
			BaseScore:          input.BaseScore,
			ComboScore:         input.ComboScore,
			Duration:           safeDuration,
		}
		if err := tx.Create(&record); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to create game record: %w", err)
		}
	}

	// 2. Update session level stats
	var sessionLevel models.GameSessionLevel
	if err := tx.Where("game_session_total_id", input.GameSessionTotalID).
		Where("game_level_id", input.GameLevelID).Where("ended_at IS NULL").
		First(&sessionLevel); err != nil || sessionLevel.ID == "" {
		_ = tx.Rollback()
		return ErrSessionLevelNotFound
	}

	var levelCountCol string
	if input.IsCorrect {
		levelCountCol = "correct_count = correct_count + 1"
	} else {
		levelCountCol = "wrong_count = wrong_count + 1"
	}
	if input.NextContentItemID != nil {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_session_levels SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = ?, updated_at = now() WHERE id = ?", levelCountCol),
			input.Score, input.MaxCombo, input.PlayTime, *input.NextContentItemID, sessionLevel.ID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session level stats: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_session_levels SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, updated_at = now() WHERE id = ?", levelCountCol),
			input.Score, input.MaxCombo, input.PlayTime, sessionLevel.ID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session level stats: %w", err)
		}
	}

	// 3. Update session total stats
	var sessionCountCol string
	if input.IsCorrect {
		sessionCountCol = "correct_count = correct_count + 1"
	} else {
		sessionCountCol = "wrong_count = wrong_count + 1"
	}
	if input.NextContentItemID != nil {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_session_totals SET score = ?, max_combo = ?, play_time = ?, %s, current_content_item_id = ?, updated_at = now() WHERE id = ?", sessionCountCol),
			input.Score, input.MaxCombo, input.PlayTime, *input.NextContentItemID, input.GameSessionTotalID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_session_totals SET score = ?, max_combo = ?, play_time = ?, %s, current_content_item_id = NULL, updated_at = now() WHERE id = ?", sessionCountCol),
			input.Score, input.MaxCombo, input.PlayTime, input.GameSessionTotalID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	}

	// 4. Touch user lastPlayedAt (once per day)
	if _, err := tx.Exec(
		"UPDATE users SET last_played_at = now(), updated_at = now() WHERE id = ? AND (last_played_at IS NULL OR last_played_at::date < CURRENT_DATE)",
		userID,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to touch user last_played_at: %w", err)
	}

	return tx.Commit()
}

// RecordSkip records a skip and increments skip counts atomically.
func RecordSkip(userID string, input RecordSkipInput) error {
	if err := verifyOwnership(userID, input.GameSessionTotalID); err != nil {
		return err
	}

	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitSkipKey, userID), rateLimitSkip, rateLimitWindowSecs,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Increment session level skip count
	var sessionLevel models.GameSessionLevel
	if err := tx.Where("game_session_total_id", input.GameSessionTotalID).
		Where("game_level_id", input.GameLevelID).Where("ended_at IS NULL").
		First(&sessionLevel); err != nil || sessionLevel.ID == "" {
		_ = tx.Rollback()
		return ErrSessionLevelNotFound
	}

	if input.NextContentItemID != nil {
		if _, err := tx.Exec(
			"UPDATE game_session_levels SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = ?, updated_at = now() WHERE id = ?",
			input.PlayTime, *input.NextContentItemID, sessionLevel.ID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session level skip count: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			"UPDATE game_session_levels SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = NULL, updated_at = now() WHERE id = ?",
			input.PlayTime, sessionLevel.ID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session level skip count: %w", err)
		}
	}

	// 2. Increment session total skip count
	if input.NextContentItemID != nil {
		if _, err := tx.Exec(
			"UPDATE game_session_totals SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = ?, updated_at = now() WHERE id = ?",
			input.PlayTime, *input.NextContentItemID, input.GameSessionTotalID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session skip count: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			"UPDATE game_session_totals SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = NULL, updated_at = now() WHERE id = ?",
			input.PlayTime, input.GameSessionTotalID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session skip count: %w", err)
		}
	}

	return tx.Commit()
}

// --- Level Operations ---

// StartLevel creates a session level entry or resumes an existing one.
func StartLevel(userID, sessionID, gameLevelID, degree string, pattern *string) (*StartLevelResult, error) {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	// VIP guard: non-first levels require active VIP
	var sessionForVip models.GameSessionTotal
	if err := facades.Orm().Query().Select("id", "game_id").Where("id", sessionID).First(&sessionForVip); err != nil || sessionForVip.ID == "" {
		return nil, ErrSessionNotFound
	}
	if err := requireVipForLevel(userID, sessionForVip.GameID, gameLevelID); err != nil {
		return nil, err
	}

	// Upsert level stats
	if err := UpsertLevelStats(userID, gameLevelID); err != nil {
		return nil, fmt.Errorf("failed to upsert level stats: %w", err)
	}

	// Count content items for this level with degree-based filtering
	contentTypes := consts.DegreeContentTypes[degree]
	totalItemsCount, err := countActiveContentItems(gameLevelID, contentTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to count content items: %w", err)
	}

	query := facades.Orm().Query()

	// Check for existing incomplete level session
	var existing models.GameSessionLevel
	if err := query.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&existing); err == nil && existing.ID != "" {
		// Touch lastPlayedAt
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

	// Create new level session
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
	}

	if err := query.Create(&levelSession); err != nil {
		return nil, fmt.Errorf("failed to create session level: %w", err)
	}

	return &StartLevelResult{
		ID:                   levelSession.ID,
		GameSessionTotalID:   levelSession.GameSessionTotalID,
		GameLevelID:          levelSession.GameLevelID,
		CurrentContentItemID: nil,
	}, nil
}

// CompleteLevel marks a level as complete and grants EXP if accuracy >= 60%.
func CompleteLevel(userID, sessionID, gameLevelID string, score, maxCombo, totalItems int) (*CompleteLevelResult, error) {
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

	// Find session to get gameID
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

	// 4. Update game stats on level complete
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

	return &CompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
	}, nil
}

// AdvanceLevel moves the session to the next level.
func AdvanceLevel(userID, sessionID, nextLevelID string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	// VIP guard: non-first levels require active VIP
	var sessionForVip models.GameSessionTotal
	if err := facades.Orm().Query().Select("id", "game_id").Where("id", sessionID).First(&sessionForVip); err != nil || sessionForVip.ID == "" {
		return ErrSessionNotFound
	}
	if err := requireVipForLevel(userID, sessionForVip.GameID, nextLevelID); err != nil {
		return err
	}

	_, err := facades.Orm().Query().Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update(map[string]any{
			"current_level_id":        nextLevelID,
			"current_content_item_id": nil,
		})
	return err
}

// --- Session Termination ---

// EndSession ends a game session and updates game stats.
func EndSession(userID, sessionID string, input EndSessionInput) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	query := facades.Orm().Query()

	// Update final session stats and mark ended in one update
	now := time.Now()
	if _, err := query.Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update(map[string]any{
			"max_combo":     input.MaxCombo,
			"correct_count": input.CorrectCount,
			"wrong_count":   input.WrongCount,
			"skip_count":    input.SkipCount,
			"ended_at":      now,
		}); err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}

	// Update game stats after session
	if err := UpdateGameStatsAfterSession(userID, input.GameID, input.AllLevelsCompleted); err != nil {
		return fmt.Errorf("failed to update game stats after session: %w", err)
	}

	// Mark first completion
	if input.AllLevelsCompleted {
		if err := MarkGameFirstCompletion(userID, input.GameID); err != nil {
			return fmt.Errorf("failed to mark first completion: %w", err)
		}
	}

	return nil
}

// SyncPlayTime syncs playtime to both session and active level.
func SyncPlayTime(userID, sessionID, gameLevelID string, playTime int) error {
	if playTime < 0 || playTime > 86400 {
		return ErrInvalidPlayTime
	}

	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Update session total playtime
	if _, err := tx.Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update("play_time", playTime); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to sync session play time: %w", err)
	}

	// Update active session level playtime
	var sessionLevel models.GameSessionLevel
	if err := tx.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&sessionLevel); err == nil && sessionLevel.ID != "" {
		if _, err := tx.Model(&models.GameSessionLevel{}).Where("id", sessionLevel.ID).
			Update("play_time", playTime); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to sync level play time: %w", err)
		}
	}

	return tx.Commit()
}

// UpdateCurrentContentItem updates the session's resume point within a level.
func UpdateCurrentContentItem(userID, sessionID string, contentItemID *string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update("current_content_item_id", contentItemID)
	return err
}

// RestoreSessionData fetches accumulated stats for restoring client state on resume.
func RestoreSessionData(userID, sessionID, gameLevelID string) (*SessionRestoreData, error) {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	var session models.GameSessionTotal
	if err := query.Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}

	result := &SessionRestoreData{
		Session: &SessionStats{
			Score:        session.Score,
			MaxCombo:     session.MaxCombo,
			CorrectCount: session.CorrectCount,
			WrongCount:   session.WrongCount,
			SkipCount:    session.SkipCount,
			PlayTime:     session.PlayTime,
		},
	}

	var sessionLevel models.GameSessionLevel
	if err := query.Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).Where("ended_at IS NULL").
		First(&sessionLevel); err == nil && sessionLevel.ID != "" {
		result.SessionLevel = &SessionStats{
			Score:        sessionLevel.Score,
			MaxCombo:     sessionLevel.MaxCombo,
			CorrectCount: sessionLevel.CorrectCount,
			WrongCount:   sessionLevel.WrongCount,
			SkipCount:    sessionLevel.SkipCount,
			PlayTime:     sessionLevel.PlayTime,
		}
	}

	return result, nil
}

// --- Helpers ---

// findActiveSession queries for an active session with specific degree+pattern.
// Excludes group sessions — only finds single-play sessions.
func findActiveSession(query orm.Query, userID, gameID, degree string, pattern *string) (*models.GameSessionTotal, error) {
	var session models.GameSessionTotal
	q := query.Where("user_id", userID).Where("game_id", gameID).
		Where("degree", degree).Where("ended_at IS NULL").
		Where("game_group_id IS NULL").
		Order("started_at desc")

	if pattern != nil {
		q = q.Where("pattern", *pattern)
	} else {
		q = q.Where("pattern IS NULL")
	}

	if err := q.First(&session); err != nil || session.ID == "" {
		return nil, nil
	}
	return &session, nil
}

// verifyOwnership checks that a session belongs to the given user.
func verifyOwnership(userID, sessionID string) error {
	var session models.GameSessionTotal
	if err := facades.Orm().Query().Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		return ErrSessionNotFound
	}
	if session.UserID != userID {
		return ErrForbidden
	}
	return nil
}

// countActiveContentItems counts active content items for a level, optionally filtered by types.
func countActiveContentItems(gameLevelID string, contentTypes []string) (int64, error) {
	query := facades.Orm().Query().Model(&models.ContentItem{}).Where("game_level_id", gameLevelID).Where("is_active", true)
	if len(contentTypes) > 0 {
		query = query.Where("content_type IN ?", contentTypes)
	}
	return query.Count()
}
