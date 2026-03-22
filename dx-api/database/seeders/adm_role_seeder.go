package seeders

import (
	"fmt"
	"log"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type AdmRoleSeeder struct{}

func (s *AdmRoleSeeder) Signature() string {
	return "AdmRoleSeeder"
}

func (s *AdmRoleSeeder) Run() error {
	query := facades.Orm().Query()

	// Upsert role
	var role models.AdmRole
	if err := query.Where("slug", "admin").First(&role); err != nil || role.ID == "" {
		role = models.AdmRole{
			ID:   uuid.Must(uuid.NewV7()).String(),
			Slug: "admin",
			Name: "Admin",
		}
		if err := query.Create(&role); err != nil {
			return fmt.Errorf("failed to create admin role: %w", err)
		}
	} else {
		if _, err := query.Where("slug", "admin").Update(&models.AdmRole{Name: "Admin"}); err != nil {
			return fmt.Errorf("failed to update admin role: %w", err)
		}
	}

	// Resolve permit "*"
	var permit models.AdmPermit
	if err := query.Where("slug", "*").First(&permit); err != nil || permit.ID == "" {
		return fmt.Errorf("permit '*' not found — run AdmPermitSeeder first")
	}

	// Upsert role-permit junction
	var existing models.AdmRolePermit
	if err := query.Where("adm_role_id", role.ID).Where("adm_permit_id", permit.ID).First(&existing); err != nil || existing.ID == "" {
		if err := query.Create(&models.AdmRolePermit{
			ID:          uuid.Must(uuid.NewV7()).String(),
			AdmRoleID:   role.ID,
			AdmPermitID: permit.ID,
		}); err != nil {
			return fmt.Errorf("failed to create role-permit link: %w", err)
		}
	}

	log.Println("Seeded 1 admin role with permit link")
	return nil
}
