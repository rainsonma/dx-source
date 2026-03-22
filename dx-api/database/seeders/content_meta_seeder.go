package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"
)

type ContentMetaSeeder struct{}

func (s *ContentMetaSeeder) Signature() string {
	return "ContentMetaSeeder"
}

type metaDef struct {
	SourceData string
	Order      float64
	LevelName  string
}

func metaDefs() []metaDef {
	return []metaDef{
		{"The food is ready.", 1000, "第一关"},
		{"I am very hungry.", 2000, "第一关"},
		{"It is a good day.", 3000, "第一关"},
		{"A car is on the road.", 1000, "第二关"},
		{"It is a red car.", 2000, "第二关"},
		{"The driver is happy.", 3000, "第二关"},
		{"The children go to school.", 1000, "第三关"},
		{"The bell rings.", 2000, "第三关"},
		{"They go home.", 3000, "第三关"},
	}
}

func (s *ContentMetaSeeder) Run() error {
	query := facades.Orm().Query()
	metas := metaDefs()

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
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return fmt.Errorf("failed to query levels for game %s: %w", game.Name, err)
		}
		levelIDs := make(map[string]string)
		for _, l := range levels {
			levelIDs[l.Name] = l.ID
		}

		for _, m := range metas {
			levelID, ok := levelIDs[m.LevelName]
			if !ok {
				continue
			}

			var existing models.ContentMeta
			if err := query.Where("source_data", m.SourceData).Where("game_level_id", levelID).First(&existing); err != nil || existing.ID == "" {
				if err := query.Create(&models.ContentMeta{
					ID:          ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
					GameLevelID: levelID,
					SourceFrom:  "manual",
					SourceType:  "sentence",
					SourceData:  m.SourceData,
					IsBreakDone: true,
					Order:       m.Order,
				}); err != nil {
					return fmt.Errorf("failed to create meta '%s': %w", m.SourceData, err)
				}
			} else {
				if _, err := query.Where("source_data", m.SourceData).Where("game_level_id", levelID).Update(&models.ContentMeta{
					SourceFrom:  "manual",
					SourceType:  "sentence",
					IsBreakDone: true,
					Order:       m.Order,
				}); err != nil {
					return fmt.Errorf("failed to update meta '%s': %w", m.SourceData, err)
				}
			}
			count++
		}
	}

	log.Printf("Seeded %d content metas\n", count)
	return nil
}
