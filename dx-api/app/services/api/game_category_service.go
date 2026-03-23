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

// ListCategories returns all enabled categories in hierarchical order.
func ListCategories() ([]CategoryData, error) {
	var categories []models.GameCategory
	if err := facades.Orm().Query().
		Where("is_enabled", true).
		Order("\"order\" ASC").
		Get(&categories); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Build parent-children map
	parentMap := make(map[string][]models.GameCategory)
	rootKey := ""
	for _, cat := range categories {
		key := rootKey
		if cat.ParentID != nil {
			key = *cat.ParentID
		}
		parentMap[key] = append(parentMap[key], cat)
	}

	// Track which IDs have children
	hasChildren := make(map[string]bool)
	for _, cat := range categories {
		if cat.ParentID != nil {
			hasChildren[*cat.ParentID] = true
		}
	}

	// Walk the tree depth-first to produce a flat list
	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		children := parentMap[parentID]
		for _, cat := range children {
			isLeaf := !hasChildren[cat.ID]
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: isLeaf,
			})
			walk(cat.ID, depth+1)
		}
	}
	walk(rootKey, 0)

	return result, nil
}
