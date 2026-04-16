package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

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
	}

	return tree, nil
}

// ListCategories returns all enabled categories in hierarchical order.
func ListCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
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
	walk("", 0)

	return result, nil
}
