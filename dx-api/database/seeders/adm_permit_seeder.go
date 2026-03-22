package seeders

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

type AdmPermitSeeder struct{}

func (s *AdmPermitSeeder) Signature() string {
	return "AdmPermitSeeder"
}

func (s *AdmPermitSeeder) Run() error {
	permits := []struct {
		Slug        string
		Name        string
		HttpMethods pq.StringArray
		HttpPaths   pq.StringArray
	}{
		{"*", "All permissions", pq.StringArray{}, pq.StringArray{"*"}},
		{"adm.dashboard", "Admin dashboard", pq.StringArray{"GET"}, pq.StringArray{}},
		{"auth.login", "Admin login", pq.StringArray{}, pq.StringArray{"/login", "/logout"}},
		{"adm.users", "Admin users", pq.StringArray{}, pq.StringArray{"/adm-users/*"}},
		{"adm.roles", "Admin roles", pq.StringArray{}, pq.StringArray{"/adm-roles/*"}},
		{"adm.permits", "Admin permits", pq.StringArray{}, pq.StringArray{"/adm-permits/*"}},
	}

	query := facades.Orm().Query()

	for _, p := range permits {
		var existing models.AdmPermit
		if err := query.Where("slug", p.Slug).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.AdmPermit{
				ID:          uuid.Must(uuid.NewV7()).String(),
				Slug:        p.Slug,
				Name:        p.Name,
				HttpMethods: p.HttpMethods,
				HttpPaths:   p.HttpPaths,
			}); err != nil {
				return fmt.Errorf("failed to create permit %s: %w", p.Slug, err)
			}
		} else {
			if _, err := query.Where("slug", p.Slug).Update(&models.AdmPermit{
				Name:        p.Name,
				HttpMethods: p.HttpMethods,
				HttpPaths:   p.HttpPaths,
			}); err != nil {
				return fmt.Errorf("failed to update permit %s: %w", p.Slug, err)
			}
		}
	}

	log.Printf("Seeded %d admin permits\n", len(permits))
	return nil
}
