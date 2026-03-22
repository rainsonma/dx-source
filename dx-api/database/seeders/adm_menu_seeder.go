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

type AdmMenuSeeder struct{}

func (s *AdmMenuSeeder) Signature() string {
	return "AdmMenuSeeder"
}

type admMenuDef struct {
	Name  string
	Icon  string
	Uri   string
	Order float64
}

type admChildMenuDef struct {
	ParentName string
	Name       string
	Icon       string
	Uri        string
	Order      float64
}

func (s *AdmMenuSeeder) Run() error {
	parents := []admMenuDef{
		{"Dashboard", "layout-dashboard", "/", 1000},
		{"System", "monitor-cog", "", 2000},
		{"Settings", "settings", "", 3000},
		{"Materials", "archive", "", 4000},
		{"Games", "gamepad-2", "", 5000},
		{"Users", "users", "", 6000},
	}

	children := []admChildMenuDef{
		{"System", "Administrators", "users", "/adm-users", 1000},
		{"System", "Adm roles", "user-lock", "/adm-roles", 2000},
		{"System", "Adm permits", "file-lock", "/adm-permits", 3000},
		{"System", "Adm menus", "square-library", "/adm-menus", 4000},
		{"System", "Adm configs", "cog", "/adm-configs", 5000},
		{"System", "Adm operates", "clipboard-clock", "/adm-operates", 6000},
		{"System", "Adm login logs", "clipboard-clock", "/adm-logins", 7000},
		{"System", "Failed queue jobs", "circle-x", "/adm-failed-queue-jobs", 8000},
		{"Settings", "Default settings", "bolt", "/", 1000},
		{"Settings", "User settings", "user-round-cog", "/", 2000},
		{"Materials", "Images", "image", "/images", 1000},
		{"Materials", "Audios", "file-headphone", "/audios", 2000},
		{"Games", "Categories", "gamepad-directional", "/game-categories", 1000},
		{"Games", "Presses", "book-plus", "/game-presses", 2000},
		{"Games", "Definitions", "sliders-horizontal", "/game-definitions", 3000},
		{"Games", "Templates", "square-dashed-kanban", "/game-templates", 4000},
		{"Games", "Games", "codesandbox", "/games", 5000},
		{"Games", "Levels", "arrow-big-up-dash", "/game-levels", 6000},
		{"Games", "Topics", "tag", "/game-topics", 7000},
		{"Games", "Contents", "list-todo", "/game-contents", 8000},
		{"Games", "Sessions", "clipboard-clock", "/game-sessions", 9000},
		{"Games", "Game progress", "circle-dot", "/game-progress", 11000},
		{"Games", "Level progress", "circle-ellipsis", "/game-level-progress", 12000},
		{"Games", "Records", "file-clock", "/game-records", 13000},
		{"Users", "User login logs", "clipboard-clock", "/user-logins", 1000},
		{"Users", "Customers", "users-round", "/users", 2000},
	}

	query := facades.Orm().Query()

	parentIDs := make(map[string]string)
	for _, p := range parents {
		icon := p.Icon
		var uri *string
		if p.Uri != "" {
			uri = &p.Uri
		}

		var existing models.AdmMenu
		if err := query.Where("name", p.Name).WhereNull("parent_id").First(&existing); err != nil || existing.ID == "" {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
			if err := query.Create(&models.AdmMenu{
				ID:    id,
				Name:  p.Name,
				Icon:  &icon,
				Uri:   uri,
				Order: p.Order,
			}); err != nil {
				return fmt.Errorf("failed to create menu %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = id
		} else {
			if _, err := query.Where("name", p.Name).WhereNull("parent_id").Update(&models.AdmMenu{
				Icon:  &icon,
				Uri:   uri,
				Order: p.Order,
			}); err != nil {
				return fmt.Errorf("failed to update menu %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = existing.ID
		}
	}

	for _, c := range children {
		parentID, ok := parentIDs[c.ParentName]
		if !ok {
			return fmt.Errorf("parent menu %s not found", c.ParentName)
		}
		icon := c.Icon
		uri := c.Uri

		var existing models.AdmMenu
		if err := query.Where("name", c.Name).Where("parent_id", parentID).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.AdmMenu{
				ID:       ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				ParentID: &parentID,
				Name:     c.Name,
				Icon:     &icon,
				Uri:      &uri,
				Order:    c.Order,
			}); err != nil {
				return fmt.Errorf("failed to create child menu %s: %w", c.Name, err)
			}
		} else {
			if _, err := query.Where("name", c.Name).Where("parent_id", parentID).Update(&models.AdmMenu{
				Icon:  &icon,
				Uri:   &uri,
				Order: c.Order,
			}); err != nil {
				return fmt.Errorf("failed to update child menu %s: %w", c.Name, err)
			}
		}
	}

	log.Printf("Seeded %d admin menus\n", len(parents)+len(children))
	return nil
}
