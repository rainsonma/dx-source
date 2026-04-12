package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// SyncCategoryName is the name of the top-level category shown on /hall/sync.
const SyncCategoryName = "同步练习"

// CategoryData represents a game category with hierarchy info.
type CategoryData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Depth  int    `json:"depth"`
	IsLeaf bool   `json:"isLeaf"`
}

// categoryTree holds the loaded category hierarchy for shared use.
type categoryTree struct {
	parentMap   map[string][]models.GameCategory
	hasChildren map[string]bool
	syncID      string
}

// loadCategoryTree loads all enabled categories and builds the tree structure.
func loadCategoryTree() (*categoryTree, error) {
	var categories []models.GameCategory
	if err := facades.Orm().Query().
		Where("is_enabled", true).
		Order("\"order\" ASC").
		Get(&categories); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	tree := &categoryTree{
		parentMap:   make(map[string][]models.GameCategory),
		hasChildren: make(map[string]bool),
	}

	for _, cat := range categories {
		key := ""
		if cat.ParentID != nil {
			key = *cat.ParentID
			tree.hasChildren[*cat.ParentID] = true
		}
		tree.parentMap[key] = append(tree.parentMap[key], cat)

		if cat.Name == SyncCategoryName && cat.ParentID == nil {
			tree.syncID = cat.ID
		}
	}

	return tree, nil
}

// ListCategories returns all enabled categories except the sync subtree.
func ListCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
	}

	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		for _, cat := range tree.parentMap[parentID] {
			if cat.ID == tree.syncID {
				continue
			}
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: !tree.hasChildren[cat.ID],
			})
			walk(cat.ID, depth+1)
		}
	}
	walk("", 0)

	return result, nil
}

// ListSyncCategories returns the sync subtree with depths adjusted to start at 0.
func ListSyncCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
	}

	if tree.syncID == "" {
		return []CategoryData{}, nil
	}

	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		for _, cat := range tree.parentMap[parentID] {
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: !tree.hasChildren[cat.ID],
			})
			walk(cat.ID, depth+1)
		}
	}
	walk(tree.syncID, 0)

	return result, nil
}
