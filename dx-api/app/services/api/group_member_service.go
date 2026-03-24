package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// MemberItem represents a group member in a list response.
type MemberItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	IsOwner   bool   `json:"is_owner"`
	CreatedAt string `json:"created_at"`
}

// memberRow is used to scan raw SQL results for member list queries.
type memberRow struct {
	ID        string `gorm:"column:id"`
	UserID    string `gorm:"column:user_id"`
	UserName  string `gorm:"column:user_name"`
	CreatedAt string `gorm:"column:created_at"`
}

// verifyMembership checks that userID is a member of groupID.
func verifyMembership(userID, groupID string) error {
	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return ErrNotGroupMember
	}
	return nil
}

// removeMemberFromGroup deletes a member from a group inside a transaction.
func removeMemberFromGroup(groupID, targetUserID string) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			"DELETE FROM game_subgroup_members WHERE game_subgroup_id IN (SELECT id FROM game_subgroups WHERE game_group_id = ?) AND user_id = ?",
			groupID, targetUserID,
		); err != nil {
			return fmt.Errorf("failed to delete subgroup memberships: %w", err)
		}
		if _, err := tx.Where("game_group_id", groupID).Where("user_id", targetUserID).Delete(&models.GameGroupMember{}); err != nil {
			return fmt.Errorf("failed to delete group member: %w", err)
		}
		if _, err := tx.Exec("UPDATE game_groups SET member_count = member_count - 1 WHERE id = ?", groupID); err != nil {
			return fmt.Errorf("failed to decrement member_count: %w", err)
		}
		return nil
	})
}

// ListGroupMembers returns a paginated list of members for a group.
func ListGroupMembers(userID, groupID, cursor string, limit int) ([]MemberItem, string, bool, error) {
	if err := verifyMembership(userID, groupID); err != nil {
		return nil, "", false, err
	}

	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return nil, "", false, ErrGroupNotFound
	}

	query := `
		SELECT m.id, m.user_id,
		       COALESCE(u.nickname, u.username) AS user_name,
		       TO_CHAR(m.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		FROM game_group_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.game_group_id = ?`

	args := []any{groupID}

	if cursor != "" {
		query += " AND m.created_at > ?"
		args = append(args, cursor)
	}

	query += " ORDER BY m.created_at ASC LIMIT ?"
	args = append(args, limit+1)

	var rows []memberRow
	if err := facades.Orm().Query().Raw(query, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list group members: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]MemberItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, MemberItem{
			ID:        r.ID,
			UserID:    r.UserID,
			UserName:  r.UserName,
			IsOwner:   r.UserID == group.OwnerID,
			CreatedAt: r.CreatedAt,
		})
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].CreatedAt
	}

	return items, nextCursor, hasMore, nil
}

// KickMember removes a member from the group (owner only).
func KickMember(userID, groupID, targetUserID string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if targetUserID == userID {
		return ErrCannotLeaveOwned
	}
	return removeMemberFromGroup(groupID, targetUserID)
}

// LeaveGroup removes the current user from the group.
func LeaveGroup(userID, groupID string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID == userID {
		return ErrCannotLeaveOwned
	}
	return removeMemberFromGroup(groupID, userID)
}

// JoinByCode adds a user to a group via invite code, returning the group ID.
func JoinByCode(userID, code string) (string, error) {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("invite_code", code).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return "", ErrGroupNotFound
	}

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", group.ID).Where("user_id", userID).First(&member); err == nil && member.ID != "" {
		return "", ErrAlreadyMember
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		newMember := models.GameGroupMember{
			ID:          newID(),
			GameGroupID: group.ID,
			UserID:      userID,
		}
		if err := tx.Create(&newMember); err != nil {
			return fmt.Errorf("failed to create member: %w", err)
		}
		if _, err := tx.Exec("UPDATE game_groups SET member_count = member_count + 1 WHERE id = ?", group.ID); err != nil {
			return fmt.Errorf("failed to increment member_count: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return group.ID, nil
}
