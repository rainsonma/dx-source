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
			ID:   ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
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
			ID:          ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
			AdmRoleID:   role.ID,
			AdmPermitID: permit.ID,
		}); err != nil {
			return fmt.Errorf("failed to create role-permit link: %w", err)
		}
	}

	log.Println("Seeded 1 admin role with permit link")
	return nil
}
