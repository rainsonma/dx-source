package seeders

import (
	"fmt"
	"log"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type GameLevelSeeder struct{}

func (s *GameLevelSeeder) Signature() string {
	return "GameLevelSeeder"
}

func (s *GameLevelSeeder) Run() error {
	levels := []struct {
		Name         string
		Order        float64
		PassingScore int
	}{
		{"第一关", 1000, 60},
		{"第二关", 2000, 60},
		{"第三关", 3000, 60},
	}

	query := facades.Orm().Query()

	// Get only the 50 seeded games by name
	gameDefs := buildGameDefs()
	gameNames := make([]any, len(gameDefs))
	for i, g := range gameDefs {
		gameNames[i] = g.Name
	}
	var games []models.Game
	if err := query.WhereIn("name", gameNames).Get(&games); err != nil {
		return fmt.Errorf("failed to query games: %w", err)
	}

	count := 0
	for _, game := range games {
		for _, l := range levels {
			var existing models.GameLevel
			if err := query.Where("name", l.Name).Where("game_id", game.ID).First(&existing); err != nil || existing.ID == "" {
				if err := query.Create(&models.GameLevel{
					ID:           uuid.Must(uuid.NewV7()).String(),
					GameID:       game.ID,
					Name:         l.Name,
					Order:        l.Order,
					PassingScore: l.PassingScore,
					IsActive:     true,
				}); err != nil {
					return fmt.Errorf("failed to create level %s for game %s: %w", l.Name, game.Name, err)
				}
			} else {
				if _, err := query.Where("name", l.Name).Where("game_id", game.ID).Update(&models.GameLevel{
					Order:        l.Order,
					PassingScore: l.PassingScore,
					IsActive:     true,
				}); err != nil {
					return fmt.Errorf("failed to update level %s for game %s: %w", l.Name, game.Name, err)
				}
			}
			count++
		}
	}

	log.Printf("Seeded %d game levels\n", count)
	return nil
}
