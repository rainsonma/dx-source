package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// GroupGameSearchItem represents a game in group game search results.
type GroupGameSearchItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	CategoryName string `json:"category_name"`
}

// SearchGamesForGroup searches published active games by name for group game selection.
func SearchGamesForGroup(userID, query string, limit int) ([]GroupGameSearchItem, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}

	q := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Order("created_at DESC").
		Limit(limit)

	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}

	var games []models.Game
	if err := q.Get(&games); err != nil {
		return nil, fmt.Errorf("failed to search games for group: %w", err)
	}

	// Batch load categories
	categoryIDs := make([]string, 0, len(games))
	for _, g := range games {
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			categoryIDs = append(categoryIDs, *g.GameCategoryID)
		}
	}
	categoryMap := make(map[string]string)
	if len(categoryIDs) > 0 {
		var categories []models.GameCategory
		if err := facades.Orm().Query().Where("id IN ?", categoryIDs).Get(&categories); err == nil {
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}
		}
	}

	result := make([]GroupGameSearchItem, 0, len(games))
	for _, g := range games {
		item := GroupGameSearchItem{
			ID:   g.ID,
			Name: g.Name,
			Mode: g.Mode,
		}
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				item.CategoryName = name
			}
		}
		result = append(result, item)
	}

	return result, nil
}

// SetGroupGame sets the current game and game mode for a group.
func SetGroupGame(userID, groupID, gameID, gameMode string, levelTimeLimit int, startGameLevelID *string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}

	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}
	if game.Status != consts.GameStatusPublished {
		return ErrGameNotPublished
	}

	if startGameLevelID != nil && *startGameLevelID != "" {
		var level models.GameLevel
		if err := facades.Orm().Query().Where("id", *startGameLevelID).Where("game_id", gameID).Where("is_active", true).First(&level); err != nil || level.ID == "" {
			return ErrLevelNotFound
		}
	}

	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update(map[string]any{
		"current_game_id":    gameID,
		"game_mode":          gameMode,
		"level_time_limit":   levelTimeLimit,
		"start_game_level_id": startGameLevelID,
	}); err != nil {
		return fmt.Errorf("failed to set group game: %w", err)
	}
	helpers.GroupNotifyHub.Notify(groupID, "detail")
	return nil
}

// ClearGroupGame clears the current game and game mode for a group.
func ClearGroupGame(userID, groupID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}

	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_groups SET current_game_id = NULL, game_mode = NULL, start_game_level_id = NULL WHERE id = ?",
		groupID,
	); err != nil {
		return fmt.Errorf("failed to clear group game: %w", err)
	}
	helpers.GroupNotifyHub.Notify(groupID, "detail")
	return nil
}

// ParticipantMember represents a connected member in the game.
type ParticipantMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// SoloParticipants is the participants payload for group_solo mode.
type SoloParticipants struct {
	Mode    string              `json:"mode"`
	Members []ParticipantMember `json:"members"`
}

// TeamParticipantGroup is one team in the participants payload for group_team mode.
type TeamParticipantGroup struct {
	SubgroupID   string              `json:"subgroup_id"`
	SubgroupName string              `json:"subgroup_name"`
	Members      []ParticipantMember `json:"members"`
}

// TeamParticipants is the participants payload for group_team mode.
type TeamParticipants struct {
	Mode  string                 `json:"mode"`
	Teams []TeamParticipantGroup `json:"teams"`
}

// GroupGameStartEvent is the SSE payload for group_game_start.
type GroupGameStartEvent struct {
	GameGroupID    string  `json:"game_group_id"`
	GameID         string  `json:"game_id"`
	GameName       string  `json:"game_name"`
	GameMode       string  `json:"game_mode"`
	Degree         string  `json:"degree"`
	Pattern        *string `json:"pattern"`
	LevelTimeLimit int     `json:"level_time_limit"`
	LevelID        *string `json:"level_id"`
	LevelName      string  `json:"level_name"`
	Participants   any     `json:"participants"`
}

// StartGroupGame validates and initiates a group game round, broadcasting via SSE.
func StartGroupGame(userID, groupID, degree string, pattern *string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}
	if group.CurrentGameID == nil || *group.CurrentGameID == "" {
		return ErrNoGameSet
	}
	if group.GameMode == nil || *group.GameMode == "" {
		return ErrNoGameModeSet
	}

	// Validate member/subgroup requirements
	memberCount, _ := facades.Orm().Query().Model(&models.GameGroupMember{}).Where("game_group_id", groupID).Count()
	if memberCount < 2 {
		return ErrNotEnoughMembers
	}

	if *group.GameMode == consts.GameModeTeam {
		type subgroupCount struct {
			Count int64 `gorm:"column:count"`
		}
		var subgroups []subgroupCount
		if err := facades.Orm().Query().Raw(
			"SELECT COUNT(*) AS count FROM game_subgroup_members WHERE game_subgroup_id IN (SELECT id FROM game_subgroups WHERE game_group_id = ?) GROUP BY game_subgroup_id",
			groupID).Scan(&subgroups); err != nil {
			return fmt.Errorf("failed to check subgroup members: %w", err)
		}
		if len(subgroups) < 2 {
			return ErrNotEnoughSubgroups
		}
		first := subgroups[0].Count
		for _, sg := range subgroups[1:] {
			if sg.Count != first {
				return ErrUnequalSubgroups
			}
		}
	}

	// Fetch game name for SSE payload
	var game models.Game
	if err := facades.Orm().Query().Where("id", *group.CurrentGameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}

	// Resolve starting level
	var startLevel models.GameLevel
	if group.StartGameLevelID != nil && *group.StartGameLevelID != "" {
		if err := facades.Orm().Query().Where("id", *group.StartGameLevelID).Where("game_id", *group.CurrentGameID).Where("is_active", true).First(&startLevel); err != nil || startLevel.ID == "" {
			// Fallback to first level if configured level is invalid
			facades.Orm().Query().Where("game_id", *group.CurrentGameID).Where("is_active", true).Order("\"order\" asc").First(&startLevel)
		}
	} else {
		facades.Orm().Query().Where("game_id", *group.CurrentGameID).Where("is_active", true).Order("\"order\" asc").First(&startLevel)
	}

	var levelID *string
	if startLevel.ID != "" {
		levelID = &startLevel.ID
	}

	// End any stale sessions from a previous round (auto-end doesn't clean these up)
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_session_levels SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID)
	facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID)

	// Set is_playing = true
	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).
		Update("is_playing", true); err != nil {
		return fmt.Errorf("failed to set is_playing: %w", err)
	}

	// Build participants roster from connected users
	connectedIDs := helpers.GroupSSEHub.ConnectedUserIDs(groupID)

	userMap := make(map[string]string, len(connectedIDs))
	if len(connectedIDs) > 0 {
		var users []models.User
		facades.Orm().Query().Where("id IN ?", connectedIDs).Get(&users)
		for _, u := range users {
			name := u.Username
			if u.Nickname != nil && *u.Nickname != "" {
				name = *u.Nickname
			}
			userMap[u.ID] = name
		}
	}

	var participants any
	if *group.GameMode == consts.GameModeTeam {
		var subgroups []models.GameSubgroup
		facades.Orm().Query().Where("game_group_id", groupID).Order("\"order\" ASC").Get(&subgroups)

		type sgMemberRow struct {
			GameSubgroupID string `gorm:"column:game_subgroup_id"`
			UserID         string `gorm:"column:user_id"`
		}
		sgIDs := make([]string, len(subgroups))
		for i, sg := range subgroups {
			sgIDs[i] = sg.ID
		}

		var sgMembers []sgMemberRow
		if len(sgIDs) > 0 {
			facades.Orm().Query().Raw(
				"SELECT game_subgroup_id, user_id FROM game_subgroup_members WHERE game_subgroup_id IN ? ORDER BY game_subgroup_id",
				sgIDs).Scan(&sgMembers)
		}

		// Group members by subgroup, only include connected users
		sgMemberMap := make(map[string][]ParticipantMember)
		for _, m := range sgMembers {
			if name, ok := userMap[m.UserID]; ok {
				sgMemberMap[m.GameSubgroupID] = append(sgMemberMap[m.GameSubgroupID], ParticipantMember{
					UserID:   m.UserID,
					UserName: name,
				})
			}
		}

		teams := make([]TeamParticipantGroup, 0, len(subgroups))
		for _, sg := range subgroups {
			members := sgMemberMap[sg.ID]
			if members == nil {
				members = []ParticipantMember{}
			}
			teams = append(teams, TeamParticipantGroup{
				SubgroupID:   sg.ID,
				SubgroupName: sg.Name,
				Members:      members,
			})
		}
		participants = TeamParticipants{Mode: consts.GameModeTeam, Teams: teams}
	} else {
		members := make([]ParticipantMember, 0, len(connectedIDs))
		for _, uid := range connectedIDs {
			if name, ok := userMap[uid]; ok {
				members = append(members, ParticipantMember{UserID: uid, UserName: name})
			}
		}
		participants = SoloParticipants{Mode: consts.GameModeSolo, Members: members}
	}

	// Broadcast SSE event
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_start", GroupGameStartEvent{
		GameGroupID:    groupID,
		GameID:         *group.CurrentGameID,
		GameName:       game.Name,
		GameMode:       *group.GameMode,
		Degree:         degree,
		Pattern:        pattern,
		LevelTimeLimit: group.LevelTimeLimit,
		LevelID:        levelID,
		LevelName:      startLevel.Name,
		Participants:   participants,
	})
	helpers.GroupNotifyHub.Notify(groupID, "detail")

	return nil
}

// ForceEndGroupGame ends all active sessions and determines winners.
func ForceEndGroupGame(userID, groupID string) ([]LevelWinnerResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return nil, ErrNotGroupOwner
	}
	if !group.IsPlaying {
		return nil, ErrGroupNotPlaying
	}

	// Collect active session IDs before ending (these scope winner queries to current round)
	type sessionIDRow struct {
		ID string `gorm:"column:id"`
	}
	var activeRows []sessionIDRow
	if err := facades.Orm().Query().Raw(
		"SELECT id FROM game_session_totals WHERE game_group_id = ? AND ended_at IS NULL",
		groupID).Scan(&activeRows); err != nil {
		return nil, fmt.Errorf("failed to collect session ids: %w", err)
	}
	sessionIDs := make([]string, len(activeRows))
	for i, r := range activeRows {
		sessionIDs[i] = r.ID
	}

	now := time.Now()

	// End all active session levels
	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_session_levels SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID); err != nil {
		return nil, fmt.Errorf("failed to end session levels: %w", err)
	}

	// End all active session totals
	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID); err != nil {
		return nil, fmt.Errorf("failed to end session totals: %w", err)
	}

	// Collect completed level IDs for winner determination (current round only)
	type levelIDRow struct {
		GameLevelID string `gorm:"column:game_level_id"`
	}
	var levelIDs []levelIDRow
	if len(sessionIDs) > 0 {
		if err := facades.Orm().Query().Raw(
			"SELECT DISTINCT game_level_id FROM game_session_levels WHERE game_group_id = ? AND ended_at IS NOT NULL AND game_session_total_id IN ?",
			groupID, sessionIDs).Scan(&levelIDs); err != nil {
			return nil, fmt.Errorf("failed to query levels: %w", err)
		}
	}

	var results []LevelWinnerResult
	for _, lid := range levelIDs {
		result, err := DetermineWinnerForLevel(groupID, lid.GameLevelID, sessionIDs)
		if err == nil && result != nil {
			results = append(results, *result)
		}
	}

	// Set is_playing = false
	facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("is_playing", false)

	// Broadcast force end event
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_force_end", map[string]any{
		"results": results,
	})
	helpers.GroupNotifyHub.Notify(groupID, "detail")

	return results, nil
}

// GroupNextLevelEvent is the SSE payload for group_next_level.
type GroupNextLevelEvent struct {
	GameGroupID    string  `json:"game_group_id"`
	GameID         string  `json:"game_id"`
	LevelID        string  `json:"level_id"`
	LevelName      string  `json:"level_name"`
	Degree         string  `json:"degree"`
	Pattern        *string `json:"pattern"`
	LevelTimeLimit int     `json:"level_time_limit"`
}

// NextGroupLevel finds the next level and broadcasts it to all group members.
func NextGroupLevel(userID, groupID, currentLevelID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if !group.IsPlaying {
		return ErrGroupNotPlaying
	}

	// Verify user is a group member
	var memberCount int64
	memberCount, _ = facades.Orm().Query().Model(&models.GameGroupMember{}).Where("game_group_id", groupID).Where("user_id", userID).Count()
	if memberCount == 0 {
		return ErrNotGroupMemberForAction
	}

	// Validate current level belongs to this game
	if group.CurrentGameID == nil {
		return ErrNoGameSet
	}
	var currentLevel models.GameLevel
	if err := facades.Orm().Query().Where("id", currentLevelID).Where("game_id", *group.CurrentGameID).First(&currentLevel); err != nil || currentLevel.ID == "" {
		return ErrLevelNotFound
	}

	// Find next active level by order
	var nextLevel models.GameLevel
	if err := facades.Orm().Query().
		Where("game_id", *group.CurrentGameID).
		Where("is_active", true).
		Where("\"order\" > ?", currentLevel.Order).
		Order("\"order\" asc").
		First(&nextLevel); err != nil || nextLevel.ID == "" {
		return ErrLastLevel
	}

	// Concurrency guard: prevent duplicate SSE broadcasts
	cacheKey := fmt.Sprintf("group_next_level:%s:%s", groupID, currentLevelID)
	if facades.Cache().Store("redis").Has(cacheKey) {
		return nil // Already broadcast, return success silently
	}
	_ = facades.Cache().Store("redis").Put(cacheKey, "1", 30*time.Second)

	// Get degree/pattern from caller's most recent session
	type sessionInfo struct {
		Degree  string  `gorm:"column:degree"`
		Pattern *string `gorm:"column:pattern"`
	}
	var si sessionInfo
	facades.Orm().Query().Raw(
		`SELECT degree, pattern FROM game_session_totals
		 WHERE game_group_id = ? AND user_id = ?
		 ORDER BY last_played_at DESC LIMIT 1`,
		groupID, userID).Scan(&si)

	helpers.GroupSSEHub.Broadcast(groupID, "group_next_level", GroupNextLevelEvent{
		GameGroupID:    groupID,
		GameID:         *group.CurrentGameID,
		LevelID:        nextLevel.ID,
		LevelName:      nextLevel.Name,
		Degree:         si.Degree,
		Pattern:        si.Pattern,
		LevelTimeLimit: group.LevelTimeLimit,
	})

	return nil
}
