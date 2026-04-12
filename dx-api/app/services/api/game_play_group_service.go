package api

import (
	"context"
	"fmt"
	"sort"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/realtime"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// --- Result types ---

// GroupPlayStartSessionResult is returned after starting or resuming a group game session.
type GroupPlayStartSessionResult struct {
	ID           string    `json:"id"`
	Degree       string    `json:"degree"`
	Pattern      *string   `json:"pattern"`
	Score        int       `json:"score"`
	Exp          int       `json:"exp"`
	MaxCombo     int       `json:"maxCombo"`
	CorrectCount int       `json:"correctCount"`
	WrongCount   int       `json:"wrongCount"`
	StartedAt    time.Time `json:"startedAt"`
	GameLevelID  string    `json:"gameLevelId"`
}

// GroupPlayCompleteLevelResult is returned after completing a level in a group game.
type GroupPlayCompleteLevelResult struct {
	ExpEarned      int     `json:"expEarned"`
	Accuracy       float64 `json:"accuracy"`
	MeetsThreshold bool    `json:"meetsThreshold"`
	NextLevelID    *string `json:"nextLevelId"`
	NextLevelName  *string `json:"nextLevelName"`
}

// GroupPlayerCompleteEvent is the SSE payload for group_player_complete.
type GroupPlayerCompleteEvent struct {
	UserID        string                 `json:"user_id"`
	UserName      string                 `json:"user_name"`
	GameLevelID   string                 `json:"game_level_id"`
	Score         int                    `json:"score"`
	Participants  []GroupParticipantInfo `json:"participants"`
	NextLevelID   *string                `json:"next_level_id"`
	NextLevelName *string                `json:"next_level_name"`
}

// GroupParticipantInfo holds a participant's score snapshot.
type GroupParticipantInfo struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

// GroupPlayerActionEvent is the SSE payload for group_player_action.
type GroupPlayerActionEvent struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Action      string `json:"action"`
	ComboStreak int    `json:"combo_streak,omitempty"`
}

// --- Session Lifecycle ---

// GroupPlayStartSession starts or resumes a group game session for a specific level.
func GroupPlayStartSession(userID, gameID, gameLevelID, degree string, pattern *string, gameGroupID string) (*GroupPlayStartSessionResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	query := facades.Orm().Query()

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

	// Check for existing active group session for this level
	existing, err := findGroupPlayActiveSession(query, userID, gameLevelID, degree, pattern, gameGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}

	if existing != nil {
		// Touch lastPlayedAt
		if _, err := query.Model(&models.GameSession{}).Where("id", existing.ID).
			Update("last_played_at", time.Now()); err != nil {
			return nil, fmt.Errorf("failed to touch session: %w", err)
		}

		return &GroupPlayStartSessionResult{
			ID:           existing.ID,
			Degree:       existing.Degree,
			Pattern:      existing.Pattern,
			Score:        existing.Score,
			Exp:          existing.Exp,
			MaxCombo:     existing.MaxCombo,
			CorrectCount: existing.CorrectCount,
			WrongCount:   existing.WrongCount,
			StartedAt:    existing.StartedAt,
			GameLevelID:  existing.GameLevelID,
		}, nil
	}

	// Count content items for this level
	totalItemsCount, err := countLevelItems(query, gameLevelID, degree)
	if err != nil {
		return nil, fmt.Errorf("failed to count content items: %w", err)
	}

	now := time.Now()
	gid := gameGroupID
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
		GameGroupID:     &gid,
		GameSubgroupID:  gameSubgroupID,
	}

	if err := query.Create(&session); err != nil {
		// Concurrent request may have already created the session
		existing, findErr := findGroupPlayActiveSession(query, userID, gameLevelID, degree, pattern, gameGroupID)
		if findErr != nil || existing == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		return &GroupPlayStartSessionResult{
			ID:           existing.ID,
			Degree:       existing.Degree,
			Pattern:      existing.Pattern,
			Score:        existing.Score,
			Exp:          existing.Exp,
			MaxCombo:     existing.MaxCombo,
			CorrectCount: existing.CorrectCount,
			WrongCount:   existing.WrongCount,
			StartedAt:    existing.StartedAt,
			GameLevelID:  existing.GameLevelID,
		}, nil
	}

	return &GroupPlayStartSessionResult{
		ID:          session.ID,
		Degree:      session.Degree,
		Pattern:     session.Pattern,
		StartedAt:   session.StartedAt,
		GameLevelID: session.GameLevelID,
	}, nil
}

// GroupPlayCompleteLevel marks a level complete — first-to-complete wins.
func GroupPlayCompleteLevel(userID, sessionID, gameLevelID string, score, maxCombo, totalItems int) (*GroupPlayCompleteLevelResult, error) {
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

	// Broadcast player completion to group (first-to-complete wins)
	if session.GameGroupID != nil {
		var user models.User
		if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user); err == nil && user.ID != "" {
			userName := user.Username
			if user.Nickname != nil && *user.Nickname != "" {
				userName = *user.Nickname
			}

			// Collect all participants' current scores for this group+level
			var participantRows []struct {
				UserID   string  `gorm:"column:user_id"`
				Username string  `gorm:"column:username"`
				Nickname *string `gorm:"column:nickname"`
				Score    int     `gorm:"column:score"`
			}
			facades.Orm().Query().Raw(
				`SELECT DISTINCT ON (gs.user_id) gs.user_id, u.username, u.nickname, gs.score
				 FROM game_sessions gs
				 JOIN users u ON u.id = gs.user_id
				 WHERE gs.game_group_id = ? AND gs.game_level_id = ?
				 ORDER BY gs.user_id, gs.created_at DESC`,
				*session.GameGroupID, gameLevelID,
			).Scan(&participantRows)

			participants := make([]GroupParticipantInfo, 0, len(participantRows))
			for _, p := range participantRows {
				name := p.Username
				if p.Nickname != nil && *p.Nickname != "" {
					name = *p.Nickname
				}
				participants = append(participants, GroupParticipantInfo{
					UserID:   p.UserID,
					UserName: name,
					Score:    p.Score,
				})
			}
			sort.Slice(participants, func(i, j int) bool {
				return participants[i].Score > participants[j].Score
			})

			nextLevelID, nextLevelName, _ := findNextLevel(session.GameID, gameLevelID)

			_ = realtime.Publish(context.Background(), realtime.GroupTopic(*session.GameGroupID), realtime.Event{Type: "group_player_complete", Data: GroupPlayerCompleteEvent{
				UserID:        userID,
				UserName:      userName,
				GameLevelID:   gameLevelID,
				Score:         score,
				Participants:  participants,
				NextLevelID:   nextLevelID,
				NextLevelName: nextLevelName,
			}})
		}

		// Force-end all other players' sessions for this level
		if session.GameSubgroupID != nil && *session.GameSubgroupID != "" {
			// Team mode: end sessions of all players not in the winning team
			_ = ForceEndGroupLosersExceptTeam(*session.GameGroupID, gameLevelID, *session.GameSubgroupID)
		} else {
			// Solo mode: end sessions of all other players
			_ = ForceEndGroupLosers(*session.GameGroupID, gameLevelID, userID)
		}

		// Update last_won_at for winner
		facades.Orm().Query().Exec(
			"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id = ?",
			time.Now(), *session.GameGroupID, userID)

		// Round is over — reset is_playing so room shows "开始游戏" again
		facades.Orm().Query().Model(&models.GameGroup{}).
			Where("id", *session.GameGroupID).Update("is_playing", false)
		_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(*session.GameGroupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "detail"}})
	}

	// Find next level
	nextLevelID, nextLevelName, _ := findNextLevel(session.GameID, gameLevelID)

	return &GroupPlayCompleteLevelResult{
		ExpEarned:      expAmount,
		Accuracy:       accuracy,
		MeetsThreshold: meetsThreshold,
		NextLevelID:    nextLevelID,
		NextLevelName:  nextLevelName,
	}, nil
}

// GroupPlayRecordAnswer records a single answer and updates session stats atomically.
func GroupPlayRecordAnswer(userID string, input RecordAnswerInput) error {
	if err := requireVip(userID); err != nil {
		return err
	}
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
	_ = tx.Where("game_session_id", input.GameSessionID).
		Where("content_item_id", input.ContentItemID).First(&existingRecord)

	if existingRecord.ID == "" {
		record := models.GameRecord{
			ID:            newID(),
			UserID:        userID,
			GameSessionID: input.GameSessionID,
			GameLevelID:   input.GameLevelID,
			ContentItemID: input.ContentItemID,
			IsCorrect:     input.IsCorrect,
			UserAnswer:    input.UserAnswer,
			SourceAnswer:  input.SourceAnswer,
			BaseScore:     input.BaseScore,
			ComboScore:    input.ComboScore,
			Duration:      safeDuration,
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
			fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = ?, updated_at = now() WHERE id = ?", countCol),
			input.Score, input.MaxCombo, input.PlayTime, *input.NextContentItemID, input.GameSessionID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to update session stats: %w", err)
		}
	} else {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE game_sessions SET score = ?, max_combo = ?, play_time = ?, played_items_count = played_items_count + 1, %s, current_content_item_id = NULL, updated_at = now() WHERE id = ?", countCol),
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit answer transaction: %w", err)
	}

	// Broadcast player action to group (fire-and-forget)
	if input.IsCorrect {
		go broadcastGroupPlayerAction(userID, input.GameSessionID, "score", 0)
		if input.ComboScore > 0 {
			go broadcastGroupPlayerAction(userID, input.GameSessionID, "combo", input.ComboScore)
		}
	}

	return nil
}

// GroupPlaySyncPlayTime syncs playtime to the session.
func GroupPlaySyncPlayTime(userID, sessionID string, playTime int) error {
	if err := requireVip(userID); err != nil {
		return err
	}
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

// GroupPlayRestoreSessionData fetches accumulated stats for restoring client state on resume.
func GroupPlayRestoreSessionData(userID, sessionID string) (*SessionRestoreData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
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

// GroupPlayUpdateContentItem updates the session's resume point.
func GroupPlayUpdateContentItem(userID, sessionID string, contentItemID *string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	if err := verifyOwnership(userID, sessionID); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Model(&models.GameSession{}).Where("id", sessionID).
		Update("current_content_item_id", contentItemID)
	return err
}

// --- Helper ---

// broadcastGroupPlayerAction broadcasts a group_player_action SSE event.
// Runs as fire-and-forget -- errors are silently ignored.
func broadcastGroupPlayerAction(userID, sessionID, action string, comboStreak int) {
	var session models.GameSession
	if err := facades.Orm().Query().Select("id", "game_group_id").
		Where("id", sessionID).First(&session); err != nil || session.GameGroupID == nil {
		return
	}

	var user models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").
		Where("id", userID).First(&user); err != nil || user.ID == "" {
		return
	}

	userName := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		userName = *user.Nickname
	}

	_ = realtime.Publish(context.Background(), realtime.GroupTopic(*session.GameGroupID), realtime.Event{Type: "group_player_action", Data: GroupPlayerActionEvent{
		UserID:      userID,
		UserName:    userName,
		Action:      action,
		ComboStreak: comboStreak,
	}})
}

// findGroupPlayActiveSession queries for an active group session for a specific level.
func findGroupPlayActiveSession(query orm.Query, userID, gameLevelID, degree string, pattern *string, gameGroupID string) (*models.GameSession, error) {
	var session models.GameSession
	q := query.Where("user_id", userID).Where("game_level_id", gameLevelID).
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
