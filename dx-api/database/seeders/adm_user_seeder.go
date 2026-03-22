package seeders

import (
	"fmt"
	"log"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type AdmUserSeeder struct{}

func (s *AdmUserSeeder) Signature() string {
	return "AdmUserSeeder"
}

func (s *AdmUserSeeder) Run() error {
	hashedPw, err := helpers.HashPassword("password123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []struct {
		Username string
		Nickname string
	}{
		{"admin", "Administrator"},
		{"manager", "Manager"},
		{"editor", "Editor"},
		{"moderator", "Moderator"},
		{"support", "Support Staff"},
		{"analyst", "Data Analyst"},
		{"developer", "Developer"},
		{"tester", "QA Tester"},
		{"designer", "UI Designer"},
		{"marketing", "Marketing Lead"},
		{"sales", "Sales Manager"},
		{"finance", "Finance Officer"},
		{"hr", "HR Manager"},
		{"ops", "Operations Lead"},
		{"content", "Content Writer"},
		{"reviewer", "Content Reviewer"},
		{"auditor", "System Auditor"},
		{"trainer", "Training Lead"},
		{"consultant", "Consultant"},
		{"partner", "Partner Manager"},
		{"vendor", "Vendor Manager"},
		{"inventory", "Inventory Manager"},
		{"logistics", "Logistics Lead"},
		{"quality", "Quality Manager"},
		{"compliance", "Compliance Officer"},
		{"security", "Security Admin"},
		{"backup", "Backup Admin"},
		{"network", "Network Admin"},
		{"database", "Database Admin"},
		{"sysadmin", "System Admin"},
	}

	query := facades.Orm().Query()

	for _, u := range users {
		nickname := u.Nickname
		var existing models.AdmUser
		if err := query.Where("username", u.Username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.AdmUser{
				ID:       uuid.Must(uuid.NewV7()).String(),
				Username: u.Username,
				Nickname: &nickname,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to create admin user %s: %w", u.Username, err)
			}
		} else {
			if _, err := query.Where("username", u.Username).Update(&models.AdmUser{
				Nickname: &nickname,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update admin user %s: %w", u.Username, err)
			}
		}
	}

	log.Printf("Seeded %d admin users\n", len(users))
	return nil
}
