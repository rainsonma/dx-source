package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/realtime"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// CreateGroupResult holds the result of a group creation.
type CreateGroupResult struct {
	ID         string `json:"id"`
	InviteCode string `json:"invite_code"`
}

// GroupListItem represents a group in a list response.
type GroupListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	OwnerID     string  `json:"owner_id"`
	OwnerName   string  `json:"owner_name"`
	MemberCount int     `json:"member_count"`
	InviteCode  string  `json:"invite_code"`
	IsMember    bool    `json:"is_member"`
	HasApplied  bool    `json:"has_applied"`
	IsOwner     bool    `json:"is_owner"`
	CreatedAt   string  `json:"created_at"`
}

// GroupDetail represents full group detail.
type GroupDetail struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	OwnerID            string  `json:"owner_id"`
	OwnerName          string  `json:"owner_name"`
	MemberCount        int     `json:"member_count"`
	InviteCode         string  `json:"invite_code"`
	IsOwner            bool    `json:"is_owner"`
	CreatedAt          string  `json:"created_at"`
	CurrentGameID      *string `json:"current_game_id"`
	GameMode           *string `json:"game_mode"`
	CurrentGameName    string  `json:"current_game_name"`
	InviteQrcodeURL    string  `json:"invite_qrcode_url"`
	IsPlaying          bool    `json:"is_playing"`
	StartGameLevelID   *string `json:"start_game_level_id"`
	StartGameLevelName *string `json:"start_game_level_name"`
}

// CreateGroup creates a new group with the given user as owner and first member.
func CreateGroup(userID, name string, description *string) (*CreateGroupResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	groupID := newID()
	inviteCode := helpers.GenerateInviteCode(8)

	group := models.GameGroup{
		ID:          groupID,
		Name:        name,
		Description: description,
		OwnerID:     userID,
		InviteCode:  inviteCode,
		MemberCount: 1,
	}
	member := models.GameGroupMember{
		ID:          newID(),
		GameGroupID: groupID,
		UserID:      userID,
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&group); err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}
		if err := tx.Create(&member); err != nil {
			return fmt.Errorf("failed to create group member: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Generate QR code after the transaction succeeds — non-fatal if it fails
	origin := facades.Config().Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000").(string)
	if idx := strings.Index(origin, ","); idx > 0 {
		origin = origin[:idx]
	}
	inviteURL := fmt.Sprintf("%s/g/%s", strings.TrimRight(origin, "/"), inviteCode)

	qrcodeURL, err := helpers.GenerateGroupQRCode(userID, inviteURL)
	if err == nil && qrcodeURL != "" {
		facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("invite_qrcode_url", qrcodeURL)
	}

	return &CreateGroupResult{ID: groupID, InviteCode: inviteCode}, nil
}

// groupRow is used to scan raw SQL results for list queries.
type groupRow struct {
	ID          string  `gorm:"column:id"`
	Name        string  `gorm:"column:name"`
	Description *string `gorm:"column:description"`
	OwnerID     string  `gorm:"column:owner_id"`
	OwnerName   string  `gorm:"column:owner_name"`
	MemberCount int     `gorm:"column:member_count"`
	InviteCode  string  `gorm:"column:invite_code"`
	IsMember    bool    `gorm:"column:is_member"`
	HasApplied  bool    `gorm:"column:has_applied"`
	CreatedAt   string  `gorm:"column:created_at"`
}

// ListGroups returns a paginated list of groups.
// tab "" (all) = all active groups; "created" = user's own; "joined" = user joined but not owned.
func ListGroups(userID, tab, cursor string, limit int) ([]GroupListItem, string, bool, error) {
	var base string
	var args []any

	switch tab {
	case "created":
		// Only groups owned by this user
		base = `
			SELECT g.id, g.name, g.description, g.owner_id,
			       COALESCE(u.nickname, u.username) AS owner_name,
			       g.member_count, g.invite_code,
			       true AS is_member, false AS has_applied,
			       TO_CHAR(g.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
			FROM game_groups g
			JOIN users u ON u.id = g.owner_id
			WHERE g.owner_id = ? AND g.dismissed_at IS NULL`
		args = []any{userID}
	case "joined":
		// Groups user is a member of but did not create
		base = `
			SELECT g.id, g.name, g.description, g.owner_id,
			       COALESCE(u.nickname, u.username) AS owner_name,
			       g.member_count, g.invite_code,
			       true AS is_member, false AS has_applied,
			       TO_CHAR(g.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
			FROM game_groups g
			JOIN game_group_members m ON m.game_group_id = g.id
			JOIN users u ON u.id = g.owner_id
			WHERE m.user_id = ? AND g.owner_id != ? AND g.dismissed_at IS NULL`
		args = []any{userID, userID}
	default:
		// All active groups — LEFT JOIN members + pending applications
		base = `
			SELECT g.id, g.name, g.description, g.owner_id,
			       COALESCE(u.nickname, u.username) AS owner_name,
			       g.member_count, g.invite_code,
			       (m.id IS NOT NULL) AS is_member,
			       (a.id IS NOT NULL) AS has_applied,
			       TO_CHAR(g.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
			FROM game_groups g
			JOIN users u ON u.id = g.owner_id
			LEFT JOIN game_group_members m ON m.game_group_id = g.id AND m.user_id = ?
			LEFT JOIN game_group_applications a ON a.game_group_id = g.id AND a.user_id = ? AND a.status = 'pending'
			WHERE g.dismissed_at IS NULL`
		args = []any{userID, userID}
	}

	if cursor != "" {
		base += " AND g.created_at < ?"
		args = append(args, cursor)
	}

	base += " ORDER BY g.created_at DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []groupRow
	if err := facades.Orm().Query().Raw(base, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list groups: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]GroupListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, GroupListItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			OwnerID:     r.OwnerID,
			OwnerName:   r.OwnerName,
			MemberCount: r.MemberCount,
			InviteCode:  r.InviteCode,
			IsMember:    r.IsMember,
			HasApplied:  r.HasApplied,
			IsOwner:     r.OwnerID == userID,
			CreatedAt:   r.CreatedAt,
		})
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].CreatedAt
	}

	return items, nextCursor, hasMore, nil
}

// GetGroupDetail returns the full detail of a group for a member.
func GetGroupDetail(userID, groupID string) (*GroupDetail, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return nil, ErrNotGroupMember
	}

	var owner models.User
	if err := facades.Orm().Query().Where("id", group.OwnerID).First(&owner); err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	ownerName := owner.Username
	if owner.Nickname != nil && *owner.Nickname != "" {
		ownerName = *owner.Nickname
	}

	var currentGameName string
	if group.CurrentGameID != nil && *group.CurrentGameID != "" {
		var game models.Game
		if err := facades.Orm().Query().Select("id", "name").Where("id", *group.CurrentGameID).First(&game); err == nil && game.ID != "" {
			currentGameName = game.Name
		}
	}

	var startGameLevelName *string
	if group.StartGameLevelID != nil && *group.StartGameLevelID != "" {
		var level models.GameLevel
		if err := facades.Orm().Query().Select("id", "name").Where("id", *group.StartGameLevelID).First(&level); err == nil && level.ID != "" {
			startGameLevelName = &level.Name
		}
	}

	var inviteQrcodeURL string
	if group.InviteQrcodeURL != nil && *group.InviteQrcodeURL != "" {
		inviteQrcodeURL = *group.InviteQrcodeURL
	}

	return &GroupDetail{
		ID:                 group.ID,
		Name:               group.Name,
		Description:        group.Description,
		OwnerID:            group.OwnerID,
		OwnerName:          ownerName,
		MemberCount:        group.MemberCount,
		InviteCode:         group.InviteCode,
		IsOwner:            group.OwnerID == userID,
		CreatedAt:          group.CreatedAt.ToDateTimeString(),
		CurrentGameID:      group.CurrentGameID,
		GameMode:           group.GameMode,
		CurrentGameName:    currentGameName,
		InviteQrcodeURL:    inviteQrcodeURL,
		IsPlaying:          group.IsPlaying,
		StartGameLevelID:   group.StartGameLevelID,
		StartGameLevelName: startGameLevelName,
	}, nil
}

// GroupInviteInfo holds public info about a group for the invite link preview.
type GroupInviteInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	MemberCount int     `json:"member_count"`
	OwnerName   string  `json:"owner_name"`
}

// GetGroupByInviteCode returns public group info for a given invite code (no auth required).
func GetGroupByInviteCode(code string) (*GroupInviteInfo, error) {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("invite_code", code).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}
	var owner models.User
	_ = facades.Orm().Query().Where("id", group.OwnerID).First(&owner)
	ownerName := owner.Username
	if owner.Nickname != nil && *owner.Nickname != "" {
		ownerName = *owner.Nickname
	}
	return &GroupInviteInfo{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		MemberCount: group.MemberCount,
		OwnerName:   ownerName,
	}, nil
}

// UpdateGroup updates the name and description of a group.
func UpdateGroup(userID, groupID, name string, description *string) error {
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

	updates := map[string]any{"name": name}
	if description != nil {
		updates["description"] = *description
	}

	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update(updates); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	return nil
}

// DismissGroup soft-dismisses a group by setting dismissed_at.
// If a game is in progress, it is force-ended first.
func DismissGroup(userID, groupID string) error {
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

	// Force-end any active game before dismissing
	if group.IsPlaying {
		if _, err := ForceEndGroupGame(userID, groupID); err != nil {
			return fmt.Errorf("failed to force-end game: %w", err)
		}
	}

	// Set dismissed_at
	now := time.Now()
	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("dismissed_at", now); err != nil {
		return fmt.Errorf("failed to dismiss group: %w", err)
	}

	_ = realtime.Publish(context.Background(), realtime.GroupTopic(groupID), realtime.Event{Type: "group_dismissed", Data: map[string]string{"group_id": groupID}})

	return nil
}
