package seeders

import (
	"fmt"
	"log"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type GameCategorySeeder struct{}

func (s *GameCategorySeeder) Signature() string {
	return "GameCategorySeeder"
}

type gameCategoryDef struct {
	Name        string
	Alias       string
	Description string
	Order       float64
}

type gameChildCategoryDef struct {
	ParentName  string
	Name        string
	Alias       string
	Description string
	Order       float64
}

func (s *GameCategorySeeder) Run() error {
	parents := []gameCategoryDef{
		{"同步练习", "同步练习", "同步练习", 1000},
		{"应试练习", "应试练习", "应试练习", 2000},
		{"分级练习", "分级练习", "分级练习", 3000},
		{"实用英语", "实用英语", "实用英语", 4000},
	}

	children := []gameChildCategoryDef{
		{"同步练习", "一年级", "一年级", "一年级", 1000},
		{"同步练习", "二年级", "二年级", "二年级", 2000},
		{"同步练习", "三年级", "三年级", "三年级", 3000},
		{"同步练习", "四年级", "四年级", "四年级", 4000},
		{"同步练习", "五年级", "五年级", "五年级", 5000},
		{"同步练习", "六年级", "六年级", "六年级", 6000},
		{"同步练习", "七年级", "七年级", "七年级", 7000},
		{"同步练习", "八年级", "八年级", "八年级", 8000},
		{"同步练习", "九年级", "九年级", "九年级", 9000},
		{"同步练习", "高中", "高中", "高中", 10000},
		{"同步练习", "中职", "中职", "中职", 11000},
	}

	query := facades.Orm().Query()

	parentIDs := make(map[string]string)
	for _, p := range parents {
		alias := p.Alias
		desc := p.Description

		var existing models.GameCategory
		if err := query.Where("name", p.Name).WhereNull("parent_id").First(&existing); err != nil || existing.ID == "" {
			id := uuid.Must(uuid.NewV7()).String()
			if err := query.Create(&models.GameCategory{
				ID:          id,
				Name:        p.Name,
				Alias:       &alias,
				Description: &desc,
				Order:       p.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to create category %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = id
		} else {
			if _, err := query.Where("name", p.Name).WhereNull("parent_id").Update(&models.GameCategory{
				Alias:       &alias,
				Description: &desc,
				Order:       p.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to update category %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = existing.ID
		}
	}

	for _, c := range children {
		parentID, ok := parentIDs[c.ParentName]
		if !ok {
			return fmt.Errorf("parent category %s not found", c.ParentName)
		}
		alias := c.Alias
		desc := c.Description

		var existing models.GameCategory
		if err := query.Where("name", c.Name).Where("parent_id", parentID).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.GameCategory{
				ID:          uuid.Must(uuid.NewV7()).String(),
				ParentID:    &parentID,
				Name:        c.Name,
				Alias:       &alias,
				Description: &desc,
				Order:       c.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to create child category %s: %w", c.Name, err)
			}
		} else {
			if _, err := query.Where("name", c.Name).Where("parent_id", parentID).Update(&models.GameCategory{
				Alias:       &alias,
				Description: &desc,
				Order:       c.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to update child category %s: %w", c.Name, err)
			}
		}
	}

	log.Printf("Seeded %d game categories\n", len(parents)+len(children))
	return nil
}
