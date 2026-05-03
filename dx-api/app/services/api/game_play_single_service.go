package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
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
	GameLevelID          string    `json:"gameLevelId"`
	Degree               string    `json:"degree"`
	Pattern              *string   `json:"pattern"`
	Score                int       `json:"score"`
	Exp                  int       `json:"exp"`
	MaxCombo             int       `json:"maxCombo"`
	CorrectCount         int       `json:"correctCount"`
	WrongCount           int       `json:"wrongCount"`
	StartedAt            time.Time `json:"startedAt"`
	CurrentContentItemID *string   `json:"currentContentItemId"`
}

// ActiveSessionData is returned when checking for an active session.
type ActiveSessionData struct {
	ID                   string  `json:"id"`
	GameLevelID          string  `json:"gameLevelId"`
	Degree               string  `json:"degree"`
	Pattern              *string `json:"pattern"`
	CurrentContentItemID *string `json:"currentContentItemId"`
}

// CompleteLevelResult is returned after completing a level.
type CompleteLevelResult struct {
	ExpEarned      int     `json:"expEarned"`
	Accuracy       float64 `json:"accuracy"`
	MeetsThreshold bool    `json:"meetsThreshold"`
	NextLevelID    *string `json:"nextLevelId"`
	NextLevelName  *string `json:"nextLevelName"`
}

// SessionRestoreData holds accumulated stats for restoring client state.
type SessionRestoreData struct {
	Score        int `json:"score"`
	MaxCombo     int `json:"maxCombo"`
	CorrectCount int `json:"correctCount"`
	WrongCount   int `json:"wrongCount"`
	SkipCount    int `json:"skipCount"`
	PlayTime     int `json:"playTime"`
}

// RecordAnswerInput holds the data needed to record an answer.
type RecordAnswerInput struct {
	GameSessionID      string
	GameLevelID        string
	ContentItemID      *string // exactly one of ContentItemID / ContentVocabID must be set
	ContentVocabID     *string
	IsCorrect          bool
	UserAnswer         string
	SourceAnswer       string
	BaseScore          int
	ComboScore         int
	Score              int
	MaxCombo           int
	PlayTime           int
	NextContentItemID  *string
	NextContentVocabID *string
	Duration           int
}

// RecordSkipInput holds the data needed to record a skip.
type RecordSkipInput struct {
	GameSessionID      string
	GameLevelID        string
	PlayTime           int
	NextContentItemID  *string
	NextContentVocabID *string
}

// EndSessionInput holds the data needed to end a session.
type EndSessionInput struct {
	Score        int
	Exp          int
	MaxCombo     int
	CorrectCount int
	WrongCount   int
	SkipCount    int
}

// --- Session Lifecycle ---

// StartSession starts or resumes a game session for a specific level.
func StartSession(userID, gameID, gameLevelID, degree string, pattern *string) (*StartSessionResult, error) {
	query := facades.Orm().Query()

	// Find the first active level for VIP guard
	var firstLevel models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&firstLevel); err != nil || firstLevel.ID == "" {
		return nil, ErrNoGameLevels
	}

	// VIP guard: non-first levels require active VIP
	if gameLevelID != firstLevel.ID {
		if err := requireVip(userID); err != nil {
			return nil, err
		}
	}

	// Check for existing active session
	existing, err := findActiveSession(query, userID, gameLevelID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}

	if existing != nil {
		// Touch lastPlayedAt
		if _, err := query.Model(&models.GameSession{}).Where("id", existing.ID).
			Update("last_played_at", time.Now()); err != nil {
			return nil, fmt.Errorf("failed to touch session: %w", err)
		}

		return &StartSessionResult{
			ID:                   existing.ID,
			GameLevelID:          existing.GameLevelID,
			Degree:               existing.Degree,
			Pattern:              existing.Pattern,
			Score:                existing.Score,
			Exp:                  existing.Exp,
			MaxCombo:             existing.MaxCombo,
			CorrectCount:         existing.CorrectCount,
			WrongCount:           existing.WrongCount,
			StartedAt:            existing.StartedAt,
			CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	// Count content items for this level
	totalItemsCount, err := countLevelItems(query, gameLevelID, degree)
	if err != nil {
		return nil, fmt.Errorf("failed to count content items: %w", err)
	}

	// Create new session
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
	}

	if err := query.Create(&session); err != nil {
		// Unique constraint violation: concurrent request already created the session.
		existing, findErr := findActiveSession(query, userID, gameLevelID, degree, pattern)
		if findErr != nil || existing == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		return &StartSessionResult{
			ID:                   existing.ID,
			GameLevelID:          existing.GameLevelID,
			Degree:               existing.Degree,
			Pattern:              existing.Pattern,
			Score:                existing.Score,
			Exp:                  existing.Exp,
			MaxCombo:             existing.MaxCombo,
			CorrectCount:         existing.CorrectCount,
			WrongCount:           existing.WrongCount,
			StartedAt:            existing.StartedAt,
			CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	return &StartSessionResult{
		ID:          session.ID,
		GameLevelID: session.GameLevelID,
		Degree:      session.Degree,
		Pattern:     session.Pattern,
		StartedAt:   session.StartedAt,
	}, nil
}

// CheckActiveSession finds an active session for a specific gameLevelID+degree+pattern combo.
func CheckActiveSession(userID, gameLevelID, degree string, pattern *string) (*ActiveSessionData, error) {
	session, err := findActiveSession(facades.Orm().Query(), userID, gameLevelID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if session == nil {
		return nil, nil
	}
	return &ActiveSessionData{
		ID:                   session.ID,
		GameLevelID:          session.GameLevelID,
		Degree:               session.Degree,
		Pattern:              session.Pattern,
		CurrentContentItemID: session.CurrentContentItemID,
	}, nil
}

// CheckAnyActiveSession finds any active single-play session for a game.
// Excludes group and PK sessions.
func CheckAnyActiveSession(userID, gameID string) (*ActiveSessionData, error) {
	query := facades.Orm().Query()

	// First: check for an active (unfinished) session
	var active models.GameSession
	if err := query.Where("user_id", userID).Where("game_id", gameID).
		Where("ended_at IS NULL").Where("game_group_id IS NULL").Where("game_pk_id IS NULL").
		Order("last_played_at desc").First(&active); err == nil && active.ID != "" {
		return &ActiveSessionData{
			ID:                   active.ID,
			GameLevelID:          active.GameLevelID,
			Degree:               active.Degree,
			Pattern:              active.Pattern,
			CurrentContentItemID: active.CurrentContentItemID,
		}, nil
	}

	// Fallback: find the latest completed session and suggest the next level
	var latest models.GameSession
	if err := query.Where("user_id", userID).Where("game_id", gameID).
		Where("ended_at IS NOT NULL").Where("game_group_id IS NULL").Where("game_pk_id IS NULL").
		Order("last_played_at desc").First(&latest); err != nil || latest.ID == "" {
		return nil, nil
	}

	levelID := latest.GameLevelID
	if nextLevelID, _, err := findNextLevel(gameID, latest.GameLevelID); err == nil && nextLevelID != nil {
		levelID = *nextLevelID
	}

	return &ActiveSessionData{
		GameLevelID: levelID,
		Degree:      latest.Degree,
		Pattern:     latest.Pattern,
	}, nil
}

// ForceCompleteSession marks a session as ended.
func ForceCompleteSession(userID, sessionID string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	now := time.Now()
	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update("ended_at", now)
	return err
}

// RestartLevel resets the session so the user can replay the level from scratch.
func RestartLevel(userID, sessionID string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update(map[string]any{
			"current_content_item_id": nil,
			"score":                   0,
			"exp":                     0,
			"max_combo":               0,
			"correct_count":           0,
			"wrong_count":             0,
			"skip_count":              0,
			"play_time":               0,
			"played_items_count":      0,
		})
	return err
}

// --- Answer / Skip Recording ---

// RecordAnswer records a single answer and updates session stats atomically.
func RecordAnswer(userID string, input RecordAnswerInput) error {
	if err := verifyOwnership(userID, input.GameSessionID); err != nil {
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

	safeDuration := max(0, min(input.Duration, 3600))

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
	q := tx.Where("game_session_id", input.GameSessionID)
	if input.ContentItemID != nil {
		q = q.Where("content_item_id", *input.ContentItemID)
	} else if input.ContentVocabID != nil {
		q = q.Where("content_vocab_id", *input.ContentVocabID)
	}
	_ = q.First(&existingRecord)

	if existingRecord.ID == "" {
		record := models.GameRecord{
			ID:             newID(),
			UserID:         userID,
			GameSessionID:  input.GameSessionID,
			GameLevelID:    input.GameLevelID,
			ContentItemID:  input.ContentItemID,
			ContentVocabID: input.ContentVocabID,
			IsCorrect:      input.IsCorrect,
			UserAnswer:     input.UserAnswer,
			SourceAnswer:   input.SourceAnswer,
			BaseScore:      input.BaseScore,
			ComboScore:     input.ComboScore,
			Duration:       safeDuration,
		}
		if err := tx.Create(&record); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to create game record: %w", err)
		}
	}

	// 2. Update session stats
	var countCol string
	if input.IsCorrect {
		countCol = "correct_count = correct_count + 1"
	} else {
		countCol = "wrong_count = wrong_count + 1"
	}
	if input.NextContentItemID != nil {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = ?, current_content_vocab_id = NULL, updated_at = now() WHERE id = ?", countCol),
			input.Score, input.MaxCombo, input.PlayTime, *input.NextContentItemID, input.GameSessionID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	} else if input.NextContentVocabID != nil {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, current_content_vocab_id = ?, updated_at = now() WHERE id = ?", countCol),
			input.Score, input.MaxCombo, input.PlayTime, *input.NextContentVocabID, input.GameSessionID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, current_content_vocab_id = NULL, updated_at = now() WHERE id = ?", countCol),
			input.Score, input.MaxCombo, input.PlayTime, input.GameSessionID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	}

	// 3. Touch user lastPlayedAt (once per day)
	if _, err := tx.Exec(
		"UPDATE users SET last_played_at = now(), updated_at = now() WHERE id = ? AND (last_played_at IS NULL OR last_played_at::date < CURRENT_DATE)",
		userID,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to touch user last_played_at: %w", err)
	}

	return tx.Commit()
}

// RecordSkip records a skip and increments skip count.
func RecordSkip(userID string, input RecordSkipInput) error {
	if err := verifyOwnership(userID, input.GameSessionID); err != nil {
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

	if input.NextContentItemID != nil {
		_, err = facades.Orm().Query().Exec(
			"UPDATE game_sessions SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = ?, current_content_vocab_id = NULL, updated_at = now() WHERE id = ?",
			input.PlayTime, *input.NextContentItemID, input.GameSessionID,
		)
	} else if input.NextContentVocabID != nil {
		_, err = facades.Orm().Query().Exec(
			"UPDATE game_sessions SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = NULL, current_content_vocab_id = ?, updated_at = now() WHERE id = ?",
			input.PlayTime, *input.NextContentVocabID, input.GameSessionID,
		)
	} else {
		_, err = facades.Orm().Query().Exec(
			"UPDATE game_sessions SET skip_count = skip_count + 1, play_time = ?, current_content_item_id = NULL, current_content_vocab_id = NULL, updated_at = now() WHERE id = ?",
			input.PlayTime, input.GameSessionID,
		)
	}
	return err
}

// --- Level Completion ---

// CompleteLevel marks the session as complete and grants EXP if accuracy >= threshold.
func CompleteLevel(userID, sessionID string, score, maxCombo, totalItems int) (*CompleteLevelResult, error) {
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

	// Find active session
	var session models.GameSession
	if err := tx.Where("id", sessionID).Where("ended_at IS NULL").
		First(&session); err != nil || session.ID == "" {
		_ = tx.Rollback()
		return nil, ErrSessionNotFound
	}

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

	// 1. Complete session
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

	// Find next level
	nextLevelID, nextLevelName, _ := findNextLevel(session.GameID, session.GameLevelID)

	return &CompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
		NextLevelID:    nextLevelID,
		NextLevelName:  nextLevelName,
	}, nil
}

// --- Session Termination ---

// EndSession ends a game session with final stats.
func EndSession(userID, sessionID string, input EndSessionInput) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	now := time.Now()
	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update(map[string]any{
			"score":         input.Score,
			"exp":           input.Exp,
			"max_combo":     input.MaxCombo,
			"correct_count": input.CorrectCount,
			"wrong_count":   input.WrongCount,
			"skip_count":    input.SkipCount,
			"ended_at":      now,
		})
	return err
}

// SyncPlayTime syncs playtime to the session.
func SyncPlayTime(userID, sessionID string, playTime int) error {
	if playTime < 0 || playTime > 86400 {
		return ErrInvalidPlayTime
	}

	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}

	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update("play_time", playTime)
	return err
}

// UpdateCurrentContentItem updates the session's resume point.
// Pass exactly one of contentItemID or contentVocabID; the other column is cleared.
func UpdateCurrentContentItem(userID, sessionID string, contentItemID, contentVocabID *string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update(map[string]any{
			"current_content_item_id":  contentItemID,
			"current_content_vocab_id": contentVocabID,
		})
	return err
}

// RestoreSessionData fetches accumulated stats for restoring client state on resume.
func RestoreSessionData(userID, sessionID string) (*SessionRestoreData, error) {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	var session models.GameSession
	if err := facades.Orm().Query().Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}

	return &SessionRestoreData{
		Score:        session.Score,
		MaxCombo:     session.MaxCombo,
		CorrectCount: session.CorrectCount,
		WrongCount:   session.WrongCount,
		SkipCount:    session.SkipCount,
		PlayTime:     session.PlayTime,
	}, nil
}

// --- Helpers ---

// findActiveSession queries for an active single-play session for a specific gameLevelID+degree+pattern.
func findActiveSession(query orm.Query, userID, gameLevelID, degree string, pattern *string) (*models.GameSession, error) {
	var session models.GameSession
	q := query.Where("user_id", userID).Where("game_level_id", gameLevelID).
		Where("degree", degree).Where("ended_at IS NULL").
		Where("game_group_id IS NULL").Where("game_pk_id IS NULL")

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
	var session models.GameSession
	if err := facades.Orm().Query().Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		return ErrSessionNotFound
	}
	if session.UserID != userID {
		return ErrForbidden
	}
	return nil
}

// countLevelItems counts content items linked to a level, mode-branched.
// Shared by single play, PK play, and group play.
func countLevelItems(query orm.Query, gameLevelID, degree string) (int64, error) {
	// We need the game's mode. Take the game_id from the level row first.
	var level models.GameLevel
	if err := query.Select("game_id").Where("id", gameLevelID).First(&level); err != nil || level.GameID == "" {
		return 0, fmt.Errorf("countLevelItems: failed to load level: %w", err)
	}
	var game models.Game
	if err := query.Select("id", "mode").Where("id", level.GameID).First(&game); err != nil || game.ID == "" {
		return 0, fmt.Errorf("countLevelItems: failed to load game: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		return query.Model(&models.GameVocab{}).Where("game_level_id", gameLevelID).Count()
	}

	q := query.Model(&models.ContentItem{}).Where("game_level_id", gameLevelID)
	if allowedTypes, ok := consts.DegreeContentTypes[degree]; ok && allowedTypes != nil {
		q = q.Where("content_type IN ?", allowedTypes)
	}
	return q.Count()
}

// findNextLevel finds the next active level after the current one.
func findNextLevel(gameID, currentLevelID string) (*string, *string, error) {
	var currentLevel models.GameLevel
	if err := facades.Orm().Query().Where("id", currentLevelID).First(&currentLevel); err != nil || currentLevel.ID == "" {
		return nil, nil, nil
	}

	var nextLevel models.GameLevel
	if err := facades.Orm().Query().Where("game_id", gameID).Where("is_active", true).
		Where("\"order\" > ?", currentLevel.Order).
		Order("\"order\" asc").First(&nextLevel); err != nil || nextLevel.ID == "" {
		return nil, nil, nil
	}

	return &nextLevel.ID, &nextLevel.Name, nil
}
