package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// SubgroupItem represents a subgroup in a list response.
type SubgroupItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	MemberCount int64   `json:"member_count"`
	Order       float64 `json:"order"`
}

// SubgroupMemberItem represents a subgroup member in a list response.
type SubgroupMemberItem struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// VerifyGroupOwnership checks that userID is the owner of groupID.
func VerifyGroupOwnership(userID, groupID string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return nil
}

// CreateSubgroup creates a new subgroup in the given group.
func CreateSubgroup(userID, groupID, name string) (string, error) {
	if err := VerifyGroupOwnership(userID, groupID); err != nil {
		return "", err
	}

	type maxOrderRow struct {
		MaxOrder float64 `gorm:"column:max_order"`
	}
	var result maxOrderRow
	if err := facades.Orm().Query().Raw(
		`SELECT COALESCE(MAX("order"), 0) AS max_order FROM game_subgroups WHERE game_group_id = ?`,
		groupID,
	).Scan(&result); err != nil {
		return "", fmt.Errorf("failed to get max order: %w", err)
	}

	sub := models.GameSubgroup{
		ID:          newID(),
		GameGroupID: groupID,
		Name:        name,
		Order:       result.MaxOrder + 1,
	}
	if err := facades.Orm().Query().Create(&sub); err != nil {
		return "", fmt.Errorf("failed to create subgroup: %w", err)
	}
	return sub.ID, nil
}

// subgroupRow is used to scan raw SQL results for subgroup list queries.
type subgroupRow struct {
	ID          string  `gorm:"column:id"`
	Name        string  `gorm:"column:name"`
	Description *string `gorm:"column:description"`
	MemberCount int64   `gorm:"column:member_count"`
	Order       float64 `gorm:"column:order"`
}

// ListSubgroups returns all subgroups for a group, with member counts.
func ListSubgroups(userID, groupID string) ([]SubgroupItem, error) {
	if err := verifyMembership(userID, groupID); err != nil {
		return nil, err
	}

	query := `
		SELECT s.id, s.name, s.description,
		       COUNT(sm.id) AS member_count,
		       s."order"
		FROM game_subgroups s
		LEFT JOIN game_subgroup_members sm ON sm.game_subgroup_id = s.id
		WHERE s.game_group_id = ?
		GROUP BY s.id, s.name, s.description, s."order"
		ORDER BY s."order" ASC`

	var rows []subgroupRow
	if err := facades.Orm().Query().Raw(query, groupID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to list subgroups: %w", err)
	}

	items := make([]SubgroupItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, SubgroupItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			MemberCount: r.MemberCount,
			Order:       r.Order,
		})
	}
	return items, nil
}

// UpdateSubgroup updates the name of a subgroup.
func UpdateSubgroup(userID, groupID, subgroupID, name string) error {
	if err := VerifyGroupOwnership(userID, groupID); err != nil {
		return err
	}

	var sub models.GameSubgroup
	if err := facades.Orm().Query().Where("id", subgroupID).Where("game_group_id", groupID).First(&sub); err != nil || sub.ID == "" {
		return ErrSubgroupNotFound
	}

	if _, err := facades.Orm().Query().Model(&models.GameSubgroup{}).Where("id", subgroupID).Update("name", name); err != nil {
		return fmt.Errorf("failed to update subgroup: %w", err)
	}
	return nil
}

// DeleteSubgroup removes a subgroup and all its members.
func DeleteSubgroup(userID, groupID, subgroupID string) error {
	if err := VerifyGroupOwnership(userID, groupID); err != nil {
		return err
	}

	var sub models.GameSubgroup
	if err := facades.Orm().Query().Where("id", subgroupID).Where("game_group_id", groupID).First(&sub); err != nil || sub.ID == "" {
		return ErrSubgroupNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Where("game_subgroup_id", subgroupID).Delete(&models.GameSubgroupMember{}); err != nil {
			return fmt.Errorf("failed to delete subgroup members: %w", err)
		}
		if _, err := tx.Where("id", subgroupID).Delete(&models.GameSubgroup{}); err != nil {
			return fmt.Errorf("failed to delete subgroup: %w", err)
		}
		return nil
	})
}

// subgroupMemberRow is used to scan raw SQL results for subgroup member list queries.
type subgroupMemberRow struct {
	ID       string `gorm:"column:id"`
	UserID   string `gorm:"column:user_id"`
	UserName string `gorm:"column:user_name"`
}

// ListSubgroupMembers returns all members of a subgroup.
func ListSubgroupMembers(userID, groupID, subgroupID string) ([]SubgroupMemberItem, error) {
	if err := verifyMembership(userID, groupID); err != nil {
		return nil, err
	}

	query := `
		SELECT sm.id, sm.user_id,
		       COALESCE(u.nickname, u.username) AS user_name
		FROM game_subgroup_members sm
		JOIN users u ON u.id = sm.user_id
		WHERE sm.game_subgroup_id = ?`

	var rows []subgroupMemberRow
	if err := facades.Orm().Query().Raw(query, subgroupID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to list subgroup members: %w", err)
	}

	items := make([]SubgroupMemberItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, SubgroupMemberItem{
			ID:       r.ID,
			UserID:   r.UserID,
			UserName: r.UserName,
		})
	}
	return items, nil
}

// AssignSubgroupMembers adds a list of group members to a subgroup.
func AssignSubgroupMembers(userID, groupID, subgroupID string, targetUserIDs []string) error {
	if err := VerifyGroupOwnership(userID, groupID); err != nil {
		return err
	}

	var sub models.GameSubgroup
	if err := facades.Orm().Query().Where("id", subgroupID).Where("game_group_id", groupID).First(&sub); err != nil || sub.ID == "" {
		return ErrSubgroupNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		for _, targetID := range targetUserIDs {
			// Verify target is a group member
			var gm models.GameGroupMember
			if err := tx.Where("game_group_id", groupID).Where("user_id", targetID).First(&gm); err != nil || gm.ID == "" {
				return ErrNotGroupMember
			}

			// Skip if already in this subgroup
			var existing models.GameSubgroupMember
			if err := tx.Where("game_subgroup_id", subgroupID).Where("user_id", targetID).First(&existing); err == nil && existing.ID != "" {
				continue
			}

			// Remove from any other subgroup in this group (one user = one subgroup)
			if _, err := tx.Exec(
				`DELETE FROM game_subgroup_members WHERE user_id = ? AND game_subgroup_id IN (SELECT id FROM game_subgroups WHERE game_group_id = ?)`,
				targetID, groupID,
			); err != nil {
				return fmt.Errorf("failed to remove from previous subgroup: %w", err)
			}

			sm := models.GameSubgroupMember{
				ID:             newID(),
				GameSubgroupID: subgroupID,
				UserID:         targetID,
			}
			if err := tx.Create(&sm); err != nil {
				return fmt.Errorf("failed to create subgroup member: %w", err)
			}
		}
		return nil
	})
}

// RemoveSubgroupMember removes a single member from a subgroup.
func RemoveSubgroupMember(userID, groupID, subgroupID, targetUserID string) error {
	if err := VerifyGroupOwnership(userID, groupID); err != nil {
		return err
	}

	var sub models.GameSubgroup
	if err := facades.Orm().Query().Where("id", subgroupID).Where("game_group_id", groupID).First(&sub); err != nil || sub.ID == "" {
		return ErrSubgroupNotFound
	}

	if _, err := facades.Orm().Query().Where("game_subgroup_id", subgroupID).Where("user_id", targetUserID).Delete(&models.GameSubgroupMember{}); err != nil {
		return fmt.Errorf("failed to remove subgroup member: %w", err)
	}
	return nil
}
