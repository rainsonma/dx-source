package api

import (
	"context"
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/realtime"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// ApplicationItem represents an application in a list response.
type ApplicationItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ApplyToGroup creates a pending application for a user to join a group.
func ApplyToGroup(userID, groupID string) (string, error) {
	if err := requireVip(userID); err != nil {
		return "", err
	}

	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return "", ErrGroupNotFound
	}

	if group.MemberCount >= consts.MaxGroupMembers {
		return "", ErrGroupMembersFull
	}

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).Where("user_id", userID).First(&member); err == nil && member.ID != "" {
		return "", ErrAlreadyMember
	}

	var existing models.GameGroupApplication
	if err := facades.Orm().Query().Where("game_group_id", groupID).Where("user_id", userID).First(&existing); err == nil && existing.ID != "" {
		if existing.Status == consts.ApplicationStatusPending {
			return "", ErrAlreadyApplied
		}
		// rejected — delete old row so we can re-apply
		if _, err := facades.Orm().Query().Where("id", existing.ID).Delete(&models.GameGroupApplication{}); err != nil {
			return "", fmt.Errorf("failed to delete old application: %w", err)
		}
	}

	app := models.GameGroupApplication{
		ID:          newID(),
		GameGroupID: groupID,
		UserID:      userID,
		Status:      consts.ApplicationStatusPending,
	}
	if err := facades.Orm().Query().Create(&app); err != nil {
		return "", fmt.Errorf("failed to create application: %w", err)
	}
	helpers.GroupNotifyHub.Notify(groupID, "applications")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "applications"}})
	return app.ID, nil
}

// CancelApplication deletes a pending application for a user+group.
func CancelApplication(userID, groupID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}

	var app models.GameGroupApplication
	if err := facades.Orm().Query().
		Where("game_group_id", groupID).
		Where("user_id", userID).
		Where("status", consts.ApplicationStatusPending).
		First(&app); err != nil || app.ID == "" {
		return ErrApplicationNotFound
	}

	if _, err := facades.Orm().Query().Where("id", app.ID).Delete(&models.GameGroupApplication{}); err != nil {
		return fmt.Errorf("failed to cancel application: %w", err)
	}
	helpers.GroupNotifyHub.Notify(groupID, "applications")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "applications"}})
	return nil
}

// applicationRow is used to scan raw SQL results for application list queries.
type applicationRow struct {
	ID        string `gorm:"column:id"`
	UserID    string `gorm:"column:user_id"`
	UserName  string `gorm:"column:user_name"`
	Status    string `gorm:"column:status"`
	CreatedAt string `gorm:"column:created_at"`
}

// ListApplications returns a paginated list of pending applications for a group.
func ListApplications(userID, groupID, cursor string, limit int) ([]ApplicationItem, string, bool, error) {
	if err := requireVip(userID); err != nil {
		return nil, "", false, err
	}

	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("dismissed_at IS NULL").First(&group); err != nil || group.ID == "" {
		return nil, "", false, ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return nil, "", false, ErrNotGroupOwner
	}

	query := `
		SELECT a.id, a.user_id,
		       COALESCE(u.nickname, u.username) AS user_name,
		       a.status,
		       TO_CHAR(a.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		FROM game_group_applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.game_group_id = ? AND a.status = ?`

	args := []any{groupID, consts.ApplicationStatusPending}

	if cursor != "" {
		query += " AND a.created_at < ?"
		args = append(args, cursor)
	}

	query += " ORDER BY a.created_at DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []applicationRow
	if err := facades.Orm().Query().Raw(query, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list applications: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]ApplicationItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ApplicationItem{
			ID:        r.ID,
			UserID:    r.UserID,
			UserName:  r.UserName,
			Status:    r.Status,
			CreatedAt: r.CreatedAt,
		})
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].CreatedAt
	}

	return items, nextCursor, hasMore, nil
}

// HandleApplication accepts or rejects a pending application.
func HandleApplication(userID, groupID, appID, action string) error {
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

	var app models.GameGroupApplication
	if err := facades.Orm().Query().Where("id", appID).Where("game_group_id", groupID).First(&app); err != nil || app.ID == "" {
		return ErrApplicationNotFound
	}
	if app.Status != consts.ApplicationStatusPending {
		return ErrApplicationNotFound
	}

	if action == "accept" {
		if group.MemberCount >= consts.MaxGroupMembers {
			return ErrGroupMembersFull
		}
		if err := facades.Orm().Transaction(func(tx orm.Query) error {
			if _, err := tx.Model(&models.GameGroupApplication{}).Where("id", appID).Update("status", consts.ApplicationStatusAccepted); err != nil {
				return fmt.Errorf("failed to update application status: %w", err)
			}
			member := models.GameGroupMember{
				ID:          newID(),
				GameGroupID: groupID,
				UserID:      app.UserID,
			}
			if err := tx.Create(&member); err != nil {
				return fmt.Errorf("failed to create member: %w", err)
			}
			if _, err := tx.Exec("UPDATE game_groups SET member_count = member_count + 1 WHERE id = ?", groupID); err != nil {
				return fmt.Errorf("failed to increment member_count: %w", err)
			}
			return nil
		}); err != nil {
			return err
		}
		helpers.GroupNotifyHub.Notify(groupID, "applications")
		_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "applications"}})
		helpers.GroupNotifyHub.Notify(groupID, "members")
		_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "members"}})
		helpers.GroupNotifyHub.Notify(groupID, "detail")
		_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "detail"}})
		return nil
	}

	// reject
	if _, err := facades.Orm().Query().Model(&models.GameGroupApplication{}).Where("id", appID).Update("status", consts.ApplicationStatusRejected); err != nil {
		return fmt.Errorf("failed to reject application: %w", err)
	}
	helpers.GroupNotifyHub.Notify(groupID, "applications")
	_ = realtime.Publish(context.Background(), realtime.GroupNotifyTopic(groupID), realtime.Event{Type: "group_updated", Data: map[string]string{"scope": "applications"}})
	return nil
}
