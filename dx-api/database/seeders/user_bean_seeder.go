package seeders

import (
	"fmt"
	"log"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type UserBeanSeeder struct{}

func (s *UserBeanSeeder) Signature() string {
	return "UserBeanSeeder"
}

func (s *UserBeanSeeder) Run() error {
	const grantAmount = 15000

	usernames := []string{"rainson", "june"}
	query := facades.Orm().Query()

	for _, username := range usernames {
		var user models.User
		if err := query.Where("username", username).First(&user); err != nil || user.ID == "" {
			log.Printf("User %s not found, skipping bean grant", username)
			continue
		}

		// Skip if already granted
		var existing models.UserBean
		if err := query.Where("user_id", user.ID).Where("slug", consts.BeanSlugSeederGrant).First(&existing); err == nil && existing.ID != "" {
			log.Printf("User %s already has seeder grant, skipping", username)
			continue
		}

		newBalance := user.Beans + grantAmount

		if _, err := query.Model(&models.User{}).Where("id", user.ID).
			Update("beans", newBalance); err != nil {
			return fmt.Errorf("failed to update beans for %s: %w", username, err)
		}

		if err := query.Create(&models.UserBean{
			ID:     uuid.Must(uuid.NewV7()).String(),
			UserID: user.ID,
			Beans:  grantAmount,
			Origin: user.Beans,
			Result: newBalance,
			Slug:   consts.BeanSlugSeederGrant,
			Reason: consts.BeanReasonSeederGrant,
		}); err != nil {
			return fmt.Errorf("failed to create bean ledger for %s: %w", username, err)
		}

		log.Printf("Granted %d beans to %s (balance: %d → %d)", grantAmount, username, user.Beans, newBalance)
	}

	return nil
}
