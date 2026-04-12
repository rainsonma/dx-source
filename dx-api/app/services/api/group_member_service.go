package api

import (
	"context"
	"fmt"

	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/realtime"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// MemberItem represents a group member in a list response.
type MemberItem struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	IsOwner      bool   `json:"is_owner"`
	IsLastWinner bool   `json:"is_last_winner"`
	CreatedAt    string `json:"created_at"`
}

// memberRow is used to scan raw SQL results for member list queries.
type memberRow struct {
	ID           string `gorm:"column:id"`
	UserID       string `gorm:"column:user_id"`
	UserName     string `gorm:"column:user_name"`
	IsLastWinner bool   `gorm:"column:is_last_winner"`
	CreatedAt    string `gorm:"column:created_at"`
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
	if err := requireVip(userID); err != nil {
		return nil, "", false, err
	}
	if err := verifyMembership(userID, groupID); err != nil {
		return nil, "", false, err
	}

	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return nil, "", false, ErrGroupNotFound
	}

	query := `
		SELECT m.id, m.user_id,
		       COALESCE(u.nickname, u.username) AS user_name,
		       CASE WHEN m.last_won_at IS NOT NULL
		            AND m.last_won_at = (SELECT MAX(m2.last_won_at) FROM game_group_members m2 WHERE m2.game_group_id = m.game_group_id)
		       THEN true ELSE false END AS is_last_winner,
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
			ID:           r.ID,
			UserID:       r.UserID,
			UserName:     r.UserName,
			IsOwner:      r.UserID == group.OwnerID,
			IsLastWinner: r.IsLastWinner,
			CreatedAt:    r.CreatedAt,
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
	if err := removeMemberFromGroup(groupID, targetUserID); err != nil {
		return err
	}
	helpers.GroupNotifyHub.Notify(groupID, "members")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "members"}})
	helpers.GroupNotifyHub.Notify(groupID, "detail")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "detail"}})
	return nil
}

// LeaveGroup removes the current user from the group.
func LeaveGroup(userID, groupID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if err := removeMemberFromGroup(groupID, userID); err != nil {
		return err
	}
	helpers.GroupNotifyHub.Notify(groupID, "members")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "members"}})
	helpers.GroupNotifyHub.Notify(groupID, "detail")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "detail"}})
	return nil
}

// JoinByCode submits a join application for a group via invite code, returning the group ID.
func JoinByCode(userID, code string) (string, error) {
	if err := requireVip(userID); err != nil {
		return "", err
	}
	var group models.GameGroup
	if err := facades.Orm().Query().Where("invite_code", code).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return "", ErrGroupNotFound
	}

	var existing models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", group.ID).Where("user_id", userID).First(&existing); err == nil && existing.ID != "" {
		return "", ErrAlreadyMember
	}

	if _, err := ApplyToGroup(userID, group.ID); err != nil {
		return "", err
	}
	return group.ID, nil
}
