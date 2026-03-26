package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// --- Result types ---

// GroupPlayStartSessionResult is returned after starting or resuming a group game session.
type GroupPlayStartSessionResult struct {
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

// GroupPlayStartLevelResult is returned after starting or resuming a group level session.
type GroupPlayStartLevelResult struct {
	ID                   string  `json:"id"`
	GameSessionTotalID   string  `json:"gameSessionTotalId"`
	GameLevelID          string  `json:"gameLevelId"`
	CurrentContentItemID *string `json:"currentContentItemId"`
}

// GroupPlayCompleteLevelResult is returned after completing a level in a group game.
type GroupPlayCompleteLevelResult struct {
	ExpEarned      int     `json:"expEarned"`
	Accuracy       float64 `json:"accuracy"`
	MeetsThreshold bool    `json:"meetsThreshold"`
}

// GroupPlayRestoreSessionResult holds accumulated stats for restoring client state in a group game.
type GroupPlayRestoreSessionResult struct {
	Session      *SessionStats `json:"session"`
	SessionLevel *SessionStats `json:"sessionLevel"`
}

// --- Session Lifecycle ---

// GroupPlayStartSession starts or resumes a group game session.
// gameGroupID is always required.
func GroupPlayStartSession(userID, gameID, degree string, pattern *string, levelID *string, gameGroupID string) (*GroupPlayStartSessionResult, error) {
	query := facades.Orm().Query()

	// Find the first active level
	var firstLevel models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&firstLevel); err != nil || firstLevel.ID == "" {
		return nil, ErrNoGameLevels
	}

	// Check for existing active group session
	existing, err := findGroupPlayActiveSession(query, userID, gameID, degree, pattern, gameGroupID)
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

		return &GroupPlayStartSessionResult{
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

	// Resolve starting level
	resolvedLevelID := &firstLevel.ID
	if levelID != nil {
		resolvedLevelID = levelID
	}

	totalLevelsCount, err := query.Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count levels: %w", err)
	}

	// Verify group exists
	var group models.GameGroup
	if err := query.Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	// Resolve subgroup if team mode
	var gameSubgroupID *string
	if group.GameMode != nil && *group.GameMode == consts.GameModeTeam {
		var subMember models.GameSubgroupMember
		if err := facades.Orm().Query().Raw(
			"SELECT gsm.* FROM game_subgroup_members gsm JOIN game_subgroups gs ON gs.id = gsm.game_subgroup_id WHERE gs.game_group_id = ? AND gsm.user_id = ? LIMIT 1",
			gameGroupID, userID,
		).Scan(&subMember); err != nil || subMember.ID == "" {
			return nil, ErrNotInSubgroup
		}
		gameSubgroupID = &subMember.GameSubgroupID
	}

	now := time.Now()
	gid := gameGroupID
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
		GameGroupID:      &gid,
		GameSubgroupID:   gameSubgroupID,
	}

	if err := query.Create(&session); err != nil {
		// Concurrent request may have already created the session — return it.
		existing, findErr := findGroupPlayActiveSession(query, userID, gameID, degree, pattern, gameGroupID)
		if findErr != nil || existing == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		return &GroupPlayStartSessionResult{
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

	return &GroupPlayStartSessionResult{
		ID:                   session.ID,
		Degree:               session.Degree,
		Pattern:              session.Pattern,
		StartedAt:            session.StartedAt,
		LevelID:              resolvedLevelID,
		CurrentContentItemID: nil,
	}, nil
}

// GroupPlayStartLevel creates or resumes a level session within a group game session.
func GroupPlayStartLevel(userID, sessionID, gameLevelID, degree string, pattern *string) (*GroupPlayStartLevelResult, error) {
	if err := verifyOwnership(userID, sessionID); err != nil {
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
		return &GroupPlayStartLevelResult{
			ID:                   existing.ID,
			GameSessionTotalID:   existing.GameSessionTotalID,
			GameLevelID:          existing.GameLevelID,
			CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	// Inherit group fields from parent session
	var parentSession models.GameSessionTotal
	if err := query.Where("id", sessionID).First(&parentSession); err != nil || parentSession.ID == "" {
		return nil, ErrSessionNotFound
	}

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
		GameGroupID:        parentSession.GameGroupID,
		GameSubgroupID:     parentSession.GameSubgroupID,
	}

	if err := query.Create(&levelSession); err != nil {
		return nil, fmt.Errorf("failed to create level session: %w", err)
	}

	return &GroupPlayStartLevelResult{
		ID:                   levelSession.ID,
		GameSessionTotalID:   levelSession.GameSessionTotalID,
		GameLevelID:          levelSession.GameLevelID,
		CurrentContentItemID: nil,
	}, nil
}

// GroupPlayCompleteLevel marks a level complete and triggers group winner determination.
func GroupPlayCompleteLevel(userID, sessionID, gameLevelID string, score, maxCombo, totalItems int) (*GroupPlayCompleteLevelResult, error) {
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

	// Find parent session to get gameID and group info
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

	// 6. Check for group winner determination and broadcast SSE
	if session.GameGroupID != nil {
		result, winErr := CheckAndDetermineWinner(*session.GameGroupID, gameLevelID)
		if winErr == nil && result != nil {
			helpers.GroupSSEHub.Broadcast(*session.GameGroupID, "group_level_complete", result)

			// If this was the last level, set is_playing = false on the group
			var totalLevels int64
			totalLevels, _ = facades.Orm().Query().Model(&models.GameLevel{}).
				Where("game_id", session.GameID).Where("is_active", true).Count()
			if int64(session.PlayedLevelsCount+1) >= totalLevels {
				facades.Orm().Query().Model(&models.GameGroup{}).
					Where("id", *session.GameGroupID).Update("is_playing", false)
			}
		}
	}

	return &GroupPlayCompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
	}, nil
}

// GroupPlayRecordAnswer records a single answer and updates session + level stats atomically.
func GroupPlayRecordAnswer(userID string, input RecordAnswerInput) error {
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

// GroupPlayRecordSkip records a skip and increments skip counts atomically.
func GroupPlayRecordSkip(userID string, input RecordSkipInput) error {
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

// GroupPlaySyncPlayTime syncs playtime to both session and active level.
func GroupPlaySyncPlayTime(userID, sessionID, gameLevelID string, playTime int) error {
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

// GroupPlayRestoreSessionData fetches accumulated stats for restoring client state on resume.
func GroupPlayRestoreSessionData(userID, sessionID, gameLevelID string) (*GroupPlayRestoreSessionResult, error) {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	var session models.GameSessionTotal
	if err := query.Where("id", sessionID).First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}

	result := &GroupPlayRestoreSessionResult{
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

// GroupPlayUpdateContentItem updates the session's resume point within a level.
func GroupPlayUpdateContentItem(userID, sessionID string, contentItemID *string) error {
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Model(&models.GameSessionTotal{}).Where("id", sessionID).
		Update("current_content_item_id", contentItemID)
	return err
}

// --- Helper ---

// findGroupPlayActiveSession queries for an active session that always filters by game_group_id.
func findGroupPlayActiveSession(query orm.Query, userID, gameID, degree string, pattern *string, gameGroupID string) (*models.GameSessionTotal, error) {
	var session models.GameSessionTotal
	q := query.Where("user_id", userID).Where("game_id", gameID).
		Where("degree", degree).Where("ended_at IS NULL").
		Where("game_group_id", gameGroupID).
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
