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

type GamePressSeeder struct{}

func (s *GamePressSeeder) Signature() string {
	return "GamePressSeeder"
}

func (s *GamePressSeeder) Run() error {
	presses := []struct {
		Name  string
		Order float64
	}{
		{"人教版", 1000},
		{"沪教版", 2000},
		{"冀教版", 3000},
		{"外研社版", 4000},
		{"译林版", 5000},
		{"北京版", 6000},
		{"北师大版", 7000},
		{"川教版", 8000},
		{"教科版", 9000},
		{"接力版", 10000},
		{"科普版", 11000},
		{"辽师大版", 12000},
		{"鲁科版", 13000},
		{"闽教版", 14000},
		{"湘鲁版", 15000},
		{"陕旅版", 16000},
		{"湘少版", 17000},
		{"粤人版", 18000},
		{"重大版", 19000},
		{"EEC 版", 20000},
		{"牛津上海版", 21000},
		{"清华版", 22000},
	}

	query := facades.Orm().Query()

	for _, p := range presses {
		var existing models.GamePress
		if err := query.Where("name", p.Name).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.GamePress{
				ID:    ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				Name:  p.Name,
				Order: p.Order,
			}); err != nil {
				return fmt.Errorf("failed to create press %s: %w", p.Name, err)
			}
		} else {
			if _, err := query.Where("name", p.Name).Update(&models.GamePress{
				Order: p.Order,
			}); err != nil {
				return fmt.Errorf("failed to update press %s: %w", p.Name, err)
			}
		}
	}

	log.Printf("Seeded %d game presses\n", len(presses))
	return nil
}
