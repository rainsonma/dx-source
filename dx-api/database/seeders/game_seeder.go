package seeders

import (
	"fmt"
	"log"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type GameSeeder struct{}

func (s *GameSeeder) Signature() string {
	return "GameSeeder"
}

type gameDef struct {
	Name         string
	CategoryName string
	PressName    string
	Order        float64
}

func buildGameDefs() []gameDef {
	categories := []string{
		"一年级", "二年级", "三年级", "四年级", "五年级",
		"六年级", "七年级", "八年级", "九年级", "高中", "中职",
	}
	presses := []string{
		"人教版", "沪教版", "冀教版", "外研社版", "译林版",
		"北京版", "北师大版", "川教版", "教科版", "接力版",
		"科普版", "辽师大版", "鲁科版", "闽教版", "湘鲁版",
		"陕旅版", "湘少版", "粤人版", "重大版", "EEC 版",
		"牛津上海版", "清华版",
	}
	volumes := []string{"上册", "下册"}

	var defs []gameDef
	order := float64(1000)

	for _, cat := range categories {
		for _, press := range presses {
			for _, vol := range volumes {
				name := fmt.Sprintf("%s%s (%s)", cat, vol, press)
				defs = append(defs, gameDef{
					Name:         name,
					CategoryName: cat,
					PressName:    press,
					Order:        order,
				})
				order += 1000
				if len(defs) >= 50 {
					return defs
				}
			}
		}
	}
	return defs
}

func (s *GameSeeder) Run() error {
	query := facades.Orm().Query()

	// Resolve user "rainson"
	var user models.User
	if err := query.Where("username", "rainson").First(&user); err != nil || user.ID == "" {
		return fmt.Errorf("user 'rainson' not found — run UserSeeder first")
	}

	// Build category name→ID map (child categories only)
	var categories []models.GameCategory
	if err := query.WhereNotNull("parent_id").Get(&categories); err != nil {
		return fmt.Errorf("failed to query categories: %w", err)
	}
	categoryIDs := make(map[string]string)
	for _, c := range categories {
		categoryIDs[c.Name] = c.ID
	}

	// Build press name→ID map
	var presses []models.GamePress
	if err := query.Get(&presses); err != nil {
		return fmt.Errorf("failed to query presses: %w", err)
	}
	pressIDs := make(map[string]string)
	for _, p := range presses {
		pressIDs[p.Name] = p.ID
	}

	games := buildGameDefs()

	for _, g := range games {
		catID, ok := categoryIDs[g.CategoryName]
		if !ok {
			return fmt.Errorf("category %s not found", g.CategoryName)
		}
		pressID, ok := pressIDs[g.PressName]
		if !ok {
			return fmt.Errorf("press %s not found", g.PressName)
		}

		var existing models.Game
		if err := query.Where("name", g.Name).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.Game{
				ID:             uuid.Must(uuid.NewV7()).String(),
				Name:           g.Name,
				Mode:           "lsrw",
				Status:         "published",
				Order:          g.Order,
				IsActive:       true,
				UserID:         &user.ID,
				GameCategoryID: &catID,
				GamePressID:    &pressID,
			}); err != nil {
				return fmt.Errorf("failed to create game %s: %w", g.Name, err)
			}
		} else {
			if _, err := query.Where("name", g.Name).Update(&models.Game{
				Mode:           "lsrw",
				Status:         "published",
				Order:          g.Order,
				IsActive:       true,
				UserID:         &user.ID,
				GameCategoryID: &catID,
				GamePressID:    &pressID,
			}); err != nil {
				return fmt.Errorf("failed to update game %s: %w", g.Name, err)
			}
		}
	}

	log.Printf("Seeded %d games\n", len(games))
	return nil
}
