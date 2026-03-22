# Goravel Database Seeders Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port all 12 Prisma seed files from dx-web to Goravel seeders in dx-api, with 50 games and idempotent upserts.

**Architecture:** 12 individual seeder files + 1 DatabaseSeeder orchestrator + bootstrap registration. Each seeder uses GORM model-based upserts via `facades.Orm().Query()`. FK resolution by querying parent records by name/slug.

**Tech Stack:** Go, Goravel framework, GORM, PostgreSQL, ULID, bcrypt

**Spec:** `docs/superpowers/specs/2026-03-22-goravel-seeders-design.md`

**Goravel ORM Notes:**
- The update method is `Update(column any, value ...any)`, NOT `Updates()`. When passing a struct: `query.Where(...).Update(&model)`.
- `Update` returns `(*db.Result, error)` — always use `if _, err := ...; err != nil`.
- `WhereIn` takes `[]any`, NOT `[]string`. Convert: `ids := make([]any, len(strings)); for i, s := range strings { ids[i] = s }`.
- GORM's struct-based update skips zero-value fields (`false`, `0`, `""`). For current seed data all booleans are `true` and all orders are > 0, so this is safe.

---

## File Map

| # | File | Action | Responsibility |
|---|------|--------|---------------|
| 1 | `dx-api/bootstrap/seeders.go` | Create | Register all seeders with `WithSeeders` |
| 2 | `dx-api/bootstrap/app.go` | Modify (line 19) | Add `.WithSeeders(Seeders)` |
| 3 | `dx-api/database/seeders/database_seeder.go` | Create | Orchestrate all seeders in dependency order |
| 4 | `dx-api/database/seeders/adm_user_seeder.go` | Create | 30 admin users |
| 5 | `dx-api/database/seeders/adm_permit_seeder.go` | Create | 6 admin permissions |
| 6 | `dx-api/database/seeders/adm_role_seeder.go` | Create | 1 role + 1 role-permit link |
| 7 | `dx-api/database/seeders/adm_menu_seeder.go` | Create | 32 admin menus (6 parent + 26 child) |
| 8 | `dx-api/database/seeders/game_category_seeder.go` | Create | 15 categories (4 parent + 11 child) |
| 9 | `dx-api/database/seeders/game_press_seeder.go` | Create | 22 publishers |
| 10 | `dx-api/database/seeders/user_seeder.go` | Create | 100 users |
| 11 | `dx-api/database/seeders/game_seeder.go` | Create | 50 games |
| 12 | `dx-api/database/seeders/game_level_seeder.go` | Create | 150 levels (3 × 50) |
| 13 | `dx-api/database/seeders/content_meta_seeder.go` | Create | 450 metas (9 × 50) |
| 14 | `dx-api/database/seeders/content_item_seeder.go` | Create | ~2,150 items (43 × 50) |

---

### Task 1: Bootstrap Registration

**Files:**
- Create: `dx-api/bootstrap/seeders.go`
- Modify: `dx-api/bootstrap/app.go:19`

- [ ] **Step 1: Create `bootstrap/seeders.go`**

```go
package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"dx-api/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
		&seeders.AdmUserSeeder{},
		&seeders.AdmPermitSeeder{},
		&seeders.AdmRoleSeeder{},
		&seeders.AdmMenuSeeder{},
		&seeders.GameCategorySeeder{},
		&seeders.GamePressSeeder{},
		&seeders.UserSeeder{},
		&seeders.GameSeeder{},
		&seeders.GameLevelSeeder{},
		&seeders.ContentMetaSeeder{},
		&seeders.ContentItemSeeder{},
	}
}
```

- [ ] **Step 2: Add `WithSeeders` to `bootstrap/app.go`**

Add `.WithSeeders(Seeders)` after `.WithMigrations(Migrations)` on line 19:

```go
return foundation.Setup().
    WithMigrations(Migrations).
    WithSeeders(Seeders).
    WithRouting(func() {
```

- [ ] **Step 3: Create empty seeders directory and placeholder**

```bash
mkdir -p dx-api/database/seeders
```

Note: This won't compile until Task 2+ creates the actual seeder structs. Do NOT run `go build` yet.

---

### Task 2: DatabaseSeeder Orchestrator

**Files:**
- Create: `dx-api/database/seeders/database_seeder.go`

- [ ] **Step 1: Create `database_seeder.go`**

```go
package seeders

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"dx-api/app/facades"
)

type DatabaseSeeder struct{}

func (s *DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

func (s *DatabaseSeeder) Run() error {
	return facades.Seeder().Call([]seeder.Seeder{
		&AdmUserSeeder{},
		&AdmPermitSeeder{},
		&AdmRoleSeeder{},
		&AdmMenuSeeder{},
		&GameCategorySeeder{},
		&GamePressSeeder{},
		&UserSeeder{},
		&GameSeeder{},
		&GameLevelSeeder{},
		&ContentMetaSeeder{},
		&ContentItemSeeder{},
	})
}
```

Note: This won't compile until all 11 seeder structs exist. Continue to Task 3.

---

### Task 3: AdmUserSeeder

**Files:**
- Create: `dx-api/database/seeders/adm_user_seeder.go`

- [ ] **Step 1: Create `adm_user_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
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
				ID:       ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
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
```

---

### Task 4: AdmPermitSeeder

**Files:**
- Create: `dx-api/database/seeders/adm_permit_seeder.go`

- [ ] **Step 1: Create `adm_permit_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/models"
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
				ID:          ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
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
```

---

### Task 5: AdmRoleSeeder

**Files:**
- Create: `dx-api/database/seeders/adm_role_seeder.go`

- [ ] **Step 1: Create `adm_role_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
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
```

---

### Task 6: AdmMenuSeeder

**Files:**
- Create: `dx-api/database/seeders/adm_menu_seeder.go`

- [ ] **Step 1: Create `adm_menu_seeder.go`**

Original data ref: `git -C dx-web show 685491d^:prisma/seeds/adm-menus-seed.ts`

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
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

	// Upsert parents, collect name→ID map
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

	// Upsert children
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
```

- [ ] **Step 2: Commit admin seeders**

```bash
cd dx-api
git add bootstrap/seeders.go bootstrap/app.go database/seeders/
git commit -m "feat: add admin seeders (users, permits, roles, menus) and bootstrap registration"
```

---

### Task 7: GameCategorySeeder

**Files:**
- Create: `dx-api/database/seeders/game_category_seeder.go`

- [ ] **Step 1: Create `game_category_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/models"
)

type GameCategorySeeder struct{}

func (s *GameCategorySeeder) Signature() string {
	return "GameCategorySeeder"
}

type gameCategoryDef struct {
	Name        string
	Alias       string
	Description string
	Order       float64
}

type gameChildCategoryDef struct {
	ParentName  string
	Name        string
	Alias       string
	Description string
	Order       float64
}

func (s *GameCategorySeeder) Run() error {
	parents := []gameCategoryDef{
		{"同步练习", "同步练习", "同步练习", 1000},
		{"应试练习", "应试练习", "应试练习", 2000},
		{"分级练习", "分级练习", "分级练习", 3000},
		{"实用英语", "实用英语", "实用英语", 4000},
	}

	children := []gameChildCategoryDef{
		{"同步练习", "一年级", "一年级", "一年级", 1000},
		{"同步练习", "二年级", "二年级", "二年级", 2000},
		{"同步练习", "三年级", "三年级", "三年级", 3000},
		{"同步练习", "四年级", "四年级", "四年级", 4000},
		{"同步练习", "五年级", "五年级", "五年级", 5000},
		{"同步练习", "六年级", "六年级", "六年级", 6000},
		{"同步练习", "七年级", "七年级", "七年级", 7000},
		{"同步练习", "八年级", "八年级", "八年级", 8000},
		{"同步练习", "九年级", "九年级", "九年级", 9000},
		{"同步练习", "高中", "高中", "高中", 10000},
		{"同步练习", "中职", "中职", "中职", 11000},
	}

	query := facades.Orm().Query()

	parentIDs := make(map[string]string)
	for _, p := range parents {
		alias := p.Alias
		desc := p.Description

		var existing models.GameCategory
		if err := query.Where("name", p.Name).WhereNull("parent_id").First(&existing); err != nil || existing.ID == "" {
			id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
			if err := query.Create(&models.GameCategory{
				ID:          id,
				Name:        p.Name,
				Alias:       &alias,
				Description: &desc,
				Order:       p.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to create category %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = id
		} else {
			if _, err := query.Where("name", p.Name).WhereNull("parent_id").Update(&models.GameCategory{
				Alias:       &alias,
				Description: &desc,
				Order:       p.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to update category %s: %w", p.Name, err)
			}
			parentIDs[p.Name] = existing.ID
		}
	}

	for _, c := range children {
		parentID, ok := parentIDs[c.ParentName]
		if !ok {
			return fmt.Errorf("parent category %s not found", c.ParentName)
		}
		alias := c.Alias
		desc := c.Description

		var existing models.GameCategory
		if err := query.Where("name", c.Name).Where("parent_id", parentID).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.GameCategory{
				ID:          ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				ParentID:    &parentID,
				Name:        c.Name,
				Alias:       &alias,
				Description: &desc,
				Order:       c.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to create child category %s: %w", c.Name, err)
			}
		} else {
			if _, err := query.Where("name", c.Name).Where("parent_id", parentID).Update(&models.GameCategory{
				Alias:       &alias,
				Description: &desc,
				Order:       c.Order,
				IsEnabled:   true,
			}); err != nil {
				return fmt.Errorf("failed to update child category %s: %w", c.Name, err)
			}
		}
	}

	log.Printf("Seeded %d game categories\n", len(parents)+len(children))
	return nil
}
```

---

### Task 8: GamePressSeeder

**Files:**
- Create: `dx-api/database/seeders/game_press_seeder.go`

- [ ] **Step 1: Create `game_press_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
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
```

---

### Task 9: UserSeeder

**Files:**
- Create: `dx-api/database/seeders/user_seeder.go`

- [ ] **Step 1: Create `user_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

type UserSeeder struct{}

func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

func (s *UserSeeder) Run() error {
	hashedPw, err := helpers.HashPassword("Password123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := facades.Orm().Query()

	// Named users
	namedUsers := []struct {
		Username string
		Nickname string
		Grade    string
		Email    string
	}{
		{"rainson", "Rainson", "lifetime", "rainsonma@gmail.com"},
		{"june", "June", "lifetime", ""},
	}

	for _, u := range namedUsers {
		nickname := u.Nickname
		var email *string
		if u.Email != "" {
			email = &u.Email
		}

		var existing models.User
		if err := query.Where("username", u.Username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				Username:   u.Username,
				Nickname:   &nickname,
				Grade:      u.Grade,
				Email:      email,
				Password:   hashedPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
			}); err != nil {
				return fmt.Errorf("failed to create user %s: %w", u.Username, err)
			}
		} else {
			if _, err := query.Where("username", u.Username).Update(&models.User{
				Nickname: &nickname,
				Grade:    u.Grade,
				Email:    email,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update user %s: %w", u.Username, err)
			}
		}
	}

	// Generic users 003–100
	for i := 3; i <= 100; i++ {
		username := fmt.Sprintf("user%03d", i)
		nickname := fmt.Sprintf("用户%03d", i)

		var existing models.User
		if err := query.Where("username", username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
				Username:   username,
				Nickname:   &nickname,
				Password:   hashedPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
			}); err != nil {
				return fmt.Errorf("failed to create user %s: %w", username, err)
			}
		} else {
			if _, err := query.Where("username", username).Update(&models.User{
				Nickname: &nickname,
				Password: hashedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update user %s: %w", username, err)
			}
		}
	}

	log.Println("Seeded 100 users")
	return nil
}
```

- [ ] **Step 2: Commit game structure + user seeders**

```bash
cd dx-api
git add database/seeders/game_category_seeder.go database/seeders/game_press_seeder.go database/seeders/user_seeder.go
git commit -m "feat: add game category, game press, and user seeders"
```

---

### Task 10: GameSeeder

**Files:**
- Create: `dx-api/database/seeders/game_seeder.go`

- [ ] **Step 1: Create `game_seeder.go`**

Generates 50 games by combining categories × presses × volumes (上册/下册). Each game resolves FK references to user "rainson", a child category, and a press.

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/models"
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
				ID:             ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
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
```

---

### Task 11: GameLevelSeeder

**Files:**
- Create: `dx-api/database/seeders/game_level_seeder.go`

- [ ] **Step 1: Create `game_level_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/models"
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
					ID:           ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
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
```

- [ ] **Step 2: Commit game + level seeders**

```bash
cd dx-api
git add database/seeders/game_seeder.go database/seeders/game_level_seeder.go
git commit -m "feat: add game and game level seeders (50 games × 3 levels)"
```

---

### Task 12: ContentMetaSeeder

**Files:**
- Create: `dx-api/database/seeders/content_meta_seeder.go`

- [ ] **Step 1: Create `content_meta_seeder.go`**

```go
package seeders

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
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
		// Build level name→ID map for this game
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
```

---

### Task 13: ContentItemSeeder

**Files:**
- Create: `dx-api/database/seeders/content_item_seeder.go`

- [ ] **Step 1: Create `content_item_seeder.go`**

This is the largest seeder. It contains all 43 content items from the original seed data (per game), duplicated across all 50 games. The `items` field is stored as JSONB.

Original data ref: `git -C dx-web show 685491d^:prisma/seeds/content-items-seed.ts`

```go
package seeders

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"dx-api/app/facades"
	"dx-api/app/models"
)

type ContentItemSeeder struct{}

func (s *ContentItemSeeder) Signature() string {
	return "ContentItemSeeder"
}

type itemDef struct {
	LevelName      string
	MetaSourceData string
	Content        string
	ContentType    string
	Translation    string
	Order          float64
	Items          []map[string]any
}

func contentItemDefs() []itemDef {
	return []itemDef{
		// --- Level 1 (第一关), Meta "The food is ready." ---
		{"第一关", "The food is ready.", "The food", "phrase", "食物", 1010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
		}},
		{"第一关", "The food is ready.", "is", "word", "是", 1020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "The food is ready.", "The food is", "block", "食物是", 1030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第一关", "The food is ready.", "ready", "word", "准备好了", 1040, []map[string]any{
			{"pos": "形容词", "item": "ready", "answer": true, "phonetic": map[string]any{"uk": "/ˈredi/", "us": "/ˈredi/"}, "position": 1, "translation": "准备好的"},
		}},
		{"第一关", "The food is ready.", "The food is ready.", "sentence", "食物准备好了。", 1050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "形容词", "item": "ready", "answer": true, "phonetic": map[string]any{"uk": "/ˈredi/", "us": "/ˈredi/"}, "position": 4, "translation": "准备好的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 1 (第一关), Meta "I am very hungry." ---
		{"第一关", "I am very hungry.", "I", "word", "我", 2010, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
		}},
		{"第一关", "I am very hungry.", "am", "word", "是", 2020, []map[string]any{
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "I am very hungry.", "I am", "block", "我是", 2030, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 2, "translation": "是"},
		}},
		{"第一关", "I am very hungry.", "very hungry", "phrase", "非常饿", 2040, []map[string]any{
			{"pos": "副词", "item": "very", "answer": true, "phonetic": map[string]any{"uk": "/ˈveri/", "us": "/ˈveri/"}, "position": 1, "translation": "非常"},
			{"pos": "形容词", "item": "hungry", "answer": true, "phonetic": map[string]any{"uk": "/ˈhʌŋɡri/", "us": "/ˈhʌŋɡri/"}, "position": 2, "translation": "饥饿的"},
		}},
		{"第一关", "I am very hungry.", "I am very hungry.", "sentence", "我非常饿。", 2050, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 2, "translation": "是"},
			{"pos": "副词", "item": "very", "answer": true, "phonetic": map[string]any{"uk": "/ˈveri/", "us": "/ˈveri/"}, "position": 3, "translation": "非常"},
			{"pos": "形容词", "item": "hungry", "answer": true, "phonetic": map[string]any{"uk": "/ˈhʌŋɡri/", "us": "/ˈhʌŋɡri/"}, "position": 4, "translation": "饥饿的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 1 (第一关), Meta "It is a good day." ---
		{"第一关", "It is a good day.", "It", "word", "它", 3010, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
		}},
		{"第一关", "It is a good day.", "is", "word", "是", 3020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "It is a good day.", "It is", "block", "它是", 3030, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
		}},
		{"第一关", "It is a good day.", "a good day", "phrase", "一个好日子", 3040, []map[string]any{
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "形容词", "item": "good", "answer": true, "phonetic": map[string]any{"uk": "/ɡʊd/", "us": "/ɡʊd/"}, "position": 2, "translation": "好的"},
			{"pos": "名词", "item": "day", "answer": true, "phonetic": map[string]any{"uk": "/deɪ/", "us": "/deɪ/"}, "position": 3, "translation": "天"},
		}},
		{"第一关", "It is a good day.", "It is a good day.", "sentence", "它是一个好日子。", 3050, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 3, "translation": "一个"},
			{"pos": "形容词", "item": "good", "answer": true, "phonetic": map[string]any{"uk": "/ɡʊd/", "us": "/ɡʊd/"}, "position": 4, "translation": "好的"},
			{"pos": "名词", "item": "day", "answer": true, "phonetic": map[string]any{"uk": "/deɪ/", "us": "/deɪ/"}, "position": 5, "translation": "天"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "A car is on the road." ---
		{"第二关", "A car is on the road.", "A car", "phrase", "一辆汽车", 1010, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
		}},
		{"第二关", "A car is on the road.", "is", "word", "是", 1020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "A car is on the road.", "A car is", "block", "一辆汽车是", 1030, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第二关", "A car is on the road.", "on the road", "phrase", "在路上", 1040, []map[string]any{
			{"pos": "介词", "item": "on", "answer": true, "phonetic": map[string]any{"uk": "/ɒn/", "us": "/ɑːn/"}, "position": 1, "translation": "在...上"},
			{"pos": "冠词", "item": "the", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 2, "translation": "这"},
			{"pos": "名词", "item": "road", "answer": true, "phonetic": map[string]any{"uk": "/rəʊd/", "us": "/roʊd/"}, "position": 3, "translation": "道路"},
		}},
		{"第二关", "A car is on the road.", "A car is on the road.", "sentence", "一辆汽车在路上。", 1050, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "介词", "item": "on", "answer": true, "phonetic": map[string]any{"uk": "/ɒn/", "us": "/ɑːn/"}, "position": 4, "translation": "在...上"},
			{"pos": "冠词", "item": "the", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 5, "translation": "这"},
			{"pos": "名词", "item": "road", "answer": true, "phonetic": map[string]any{"uk": "/rəʊd/", "us": "/roʊd/"}, "position": 6, "translation": "道路"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 7, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "It is a red car." ---
		{"第二关", "It is a red car.", "It", "word", "它", 2010, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
		}},
		{"第二关", "It is a red car.", "is", "word", "是", 2020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "It is a red car.", "It is", "block", "它是", 2030, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
		}},
		{"第二关", "It is a red car.", "a red car", "phrase", "一辆红色的汽车", 2040, []map[string]any{
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "形容词", "item": "red", "answer": true, "phonetic": map[string]any{"uk": "/red/", "us": "/red/"}, "position": 2, "translation": "红色的"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 3, "translation": "汽车"},
		}},
		{"第二关", "It is a red car.", "It is a red car.", "sentence", "它是一辆红色的汽车。", 2050, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 3, "translation": "一个"},
			{"pos": "形容词", "item": "red", "answer": true, "phonetic": map[string]any{"uk": "/red/", "us": "/red/"}, "position": 4, "translation": "红色的"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 5, "translation": "汽车"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "The driver is happy." ---
		{"第二关", "The driver is happy.", "The driver", "phrase", "司机", 3010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
		}},
		{"第二关", "The driver is happy.", "is", "word", "是", 3020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "The driver is happy.", "The driver is", "block", "司机是", 3030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第二关", "The driver is happy.", "happy", "word", "高兴的", 3040, []map[string]any{
			{"pos": "形容词", "item": "happy", "answer": true, "phonetic": map[string]any{"uk": "/ˈhæpi/", "us": "/ˈhæpi/"}, "position": 1, "translation": "高兴的"},
		}},
		{"第二关", "The driver is happy.", "The driver is happy.", "sentence", "司机很高兴。", 3050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "形容词", "item": "happy", "answer": true, "phonetic": map[string]any{"uk": "/ˈhæpi/", "us": "/ˈhæpi/"}, "position": 4, "translation": "高兴的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "The children go to school." ---
		{"第三关", "The children go to school.", "The children", "phrase", "孩子们", 1010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
		}},
		{"第三关", "The children go to school.", "go", "word", "去", 1020, []map[string]any{
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 1, "translation": "去"},
		}},
		{"第三关", "The children go to school.", "The children go", "block", "孩子们去", 1030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 3, "translation": "去"},
		}},
		{"第三关", "The children go to school.", "to school", "phrase", "去上学", 1040, []map[string]any{
			{"pos": "介词", "item": "to", "answer": true, "phonetic": map[string]any{"uk": "/tuː/", "us": "/tuː/"}, "position": 1, "translation": "到"},
			{"pos": "名词", "item": "school", "answer": true, "phonetic": map[string]any{"uk": "/skuːl/", "us": "/skuːl/"}, "position": 2, "translation": "学校"},
		}},
		{"第三关", "The children go to school.", "The children go to school.", "sentence", "孩子们去上学。", 1050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 3, "translation": "去"},
			{"pos": "介词", "item": "to", "answer": true, "phonetic": map[string]any{"uk": "/tuː/", "us": "/tuː/"}, "position": 4, "translation": "到"},
			{"pos": "名词", "item": "school", "answer": true, "phonetic": map[string]any{"uk": "/skuːl/", "us": "/skuːl/"}, "position": 5, "translation": "学校"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "The bell rings." ---
		{"第三关", "The bell rings.", "The bell", "phrase", "铃", 2010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "bell", "answer": true, "phonetic": map[string]any{"uk": "/bel/", "us": "/bel/"}, "position": 2, "translation": "铃"},
		}},
		{"第三关", "The bell rings.", "rings", "word", "响", 2020, []map[string]any{
			{"pos": "动词", "item": "rings", "answer": true, "phonetic": map[string]any{"uk": "/rɪŋz/", "us": "/rɪŋz/"}, "position": 1, "translation": "响"},
		}},
		{"第三关", "The bell rings.", "The bell rings.", "sentence", "铃响了。", 2030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "bell", "answer": true, "phonetic": map[string]any{"uk": "/bel/", "us": "/bel/"}, "position": 2, "translation": "铃"},
			{"pos": "动词", "item": "rings", "answer": true, "phonetic": map[string]any{"uk": "/rɪŋz/", "us": "/rɪŋz/"}, "position": 3, "translation": "响"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 4, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "They go home." ---
		{"第三关", "They go home.", "They", "word", "他们", 3010, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
		}},
		{"第三关", "They go home.", "go", "word", "去", 3020, []map[string]any{
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 1, "translation": "去"},
		}},
		{"第三关", "They go home.", "They go", "block", "他们去", 3030, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 2, "translation": "去"},
		}},
		{"第三关", "They go home.", "home", "word", "家", 3040, []map[string]any{
			{"pos": "名词", "item": "home", "answer": true, "phonetic": map[string]any{"uk": "/həʊm/", "us": "/hoʊm/"}, "position": 1, "translation": "家"},
		}},
		{"第三关", "They go home.", "They go home.", "sentence", "他们回家。", 3050, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 2, "translation": "去"},
			{"pos": "名词", "item": "home", "answer": true, "phonetic": map[string]any{"uk": "/həʊm/", "us": "/hoʊm/"}, "position": 3, "translation": "家"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 4, "translation": ""},
		}},
	}
}

func (s *ContentItemSeeder) Run() error {
	query := facades.Orm().Query()
	items := contentItemDefs()

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
		// Build level name→ID map
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return fmt.Errorf("failed to query levels for game %s: %w", game.Name, err)
		}
		levelIDs := make(map[string]string)
		for _, l := range levels {
			levelIDs[l.Name] = l.ID
		}

		// Build meta (levelID:sourceData)→ID map
		levelIDList := make([]any, 0, len(levelIDs))
		for _, id := range levelIDs {
			levelIDList = append(levelIDList, id)
		}
		var metas []models.ContentMeta
		if err := query.WhereIn("game_level_id", levelIDList).Get(&metas); err != nil {
			return fmt.Errorf("failed to query metas for game %s: %w", game.Name, err)
		}
		metaIDs := make(map[string]string)
		for _, m := range metas {
			key := m.GameLevelID + ":" + m.SourceData
			metaIDs[key] = m.ID
		}

		for _, item := range items {
			levelID, ok := levelIDs[item.LevelName]
			if !ok {
				continue
			}

			metaKey := levelID + ":" + item.MetaSourceData
			metaID, ok := metaIDs[metaKey]
			if !ok {
				continue
			}

			// Serialize items to JSON string
			itemsJSON, err := json.Marshal(item.Items)
			if err != nil {
				return fmt.Errorf("failed to marshal items: %w", err)
			}
			itemsStr := string(itemsJSON)
			translation := item.Translation

			var existing models.ContentItem
			if err := query.Where("content", item.Content).Where("content_type", item.ContentType).Where("content_meta_id", metaID).First(&existing); err != nil || existing.ID == "" {
				if err := query.Create(&models.ContentItem{
					ID:            ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
					GameLevelID:   levelID,
					ContentMetaID: &metaID,
					Content:       item.Content,
					ContentType:   item.ContentType,
					Translation:   &translation,
					Items:         &itemsStr,
					Order:         item.Order,
					IsActive:      true,
				}); err != nil {
					return fmt.Errorf("failed to create content item '%s': %w", item.Content, err)
				}
			} else {
				if _, err := query.Where("content", item.Content).Where("content_type", item.ContentType).Where("content_meta_id", metaID).Update(&models.ContentItem{
					GameLevelID: levelID,
					Translation: &translation,
					Items:       &itemsStr,
					Order:       item.Order,
					IsActive:    true,
				}); err != nil {
					return fmt.Errorf("failed to update content item '%s': %w", item.Content, err)
				}
			}
			count++
		}
	}

	log.Printf("Seeded %d content items\n", count)
	return nil
}
```

- [ ] **Step 2: Commit content seeders**

```bash
cd dx-api
git add database/seeders/content_meta_seeder.go database/seeders/content_item_seeder.go
git commit -m "feat: add content meta and content item seeders (9 metas + 43 items × 50 games)"
```

---

### Task 14: Build Verification

- [ ] **Step 1: Verify compilation**

```bash
cd dx-api
go build ./...
```

Expected: clean build, no errors.

- [ ] **Step 2: Run `go vet`**

```bash
cd dx-api
go vet ./...
```

Expected: no issues.

- [ ] **Step 3: Fix any build errors**

If there are compilation errors, fix them. Common issues:
- Missing imports (check `pq`, `ulid/v2`, etc.)
- Model field type mismatches (pointer vs value)
- Goravel ORM method signatures

- [ ] **Step 4: Commit build fixes if any**

```bash
cd dx-api
git add -A
git commit -m "fix: resolve build errors in seeders"
```

---

### Task 15: Integration Test

- [ ] **Step 1: Run seeders against a test database**

Ensure PostgreSQL is running and the database exists. Then:

```bash
cd dx-api
go run . artisan db:seed
```

Expected output:
```
Seeded 30 admin users
Seeded 6 admin permits
Seeded 1 admin role with permit link
Seeded 32 admin menus
Seeded 15 game categories
Seeded 22 game presses
Seeded 100 users
Seeded 50 games
Seeded 150 game levels
Seeded 450 content metas
Seeded 2150 content items
```

- [ ] **Step 2: Verify idempotency — run again**

```bash
cd dx-api
go run . artisan db:seed
```

Expected: same output, no duplicate records, no errors.

- [ ] **Step 3: Verify record counts in database**

```bash
psql -d douxue -c "
  SELECT 'adm_users' AS t, COUNT(*) FROM adm_users
  UNION ALL SELECT 'adm_permits', COUNT(*) FROM adm_permits
  UNION ALL SELECT 'adm_roles', COUNT(*) FROM adm_roles
  UNION ALL SELECT 'adm_role_permits', COUNT(*) FROM adm_role_permits
  UNION ALL SELECT 'adm_menus', COUNT(*) FROM adm_menus
  UNION ALL SELECT 'game_categories', COUNT(*) FROM game_categories
  UNION ALL SELECT 'game_presses', COUNT(*) FROM game_presses
  UNION ALL SELECT 'users', COUNT(*) FROM users
  UNION ALL SELECT 'games', COUNT(*) FROM games
  UNION ALL SELECT 'game_levels', COUNT(*) FROM game_levels
  UNION ALL SELECT 'content_metas', COUNT(*) FROM content_metas
  UNION ALL SELECT 'content_items', COUNT(*) FROM content_items;
"
```

Expected counts (at minimum the seeded amounts — may be higher if pre-existing data):
- adm_users: >= 30
- adm_permits: >= 6
- adm_roles: >= 1
- adm_role_permits: >= 1
- adm_menus: >= 32
- game_categories: >= 15
- game_presses: >= 22
- users: >= 100
- games: >= 50
- game_levels: >= 150
- content_metas: >= 450
- content_items: >= 2150

- [ ] **Step 4: Run specific seeder to verify individual execution**

```bash
cd dx-api
go run . artisan db:seed --seeder=AdmUserSeeder
```

Expected: only admin users seeded.

- [ ] **Step 5: Final commit**

```bash
cd dx-api
git add -A
git commit -m "feat: complete Goravel database seeders — 12 seeders, ~3,100 records"
```
