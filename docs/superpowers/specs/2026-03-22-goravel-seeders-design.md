# Goravel Database Seeders Design

## Overview

Port all 12 Prisma seed files from the original dx-web to Goravel seeders in dx-api. The original seeds were deleted in Phase 11 of the migration (commit `685491d`) but never recreated on the Go side.

## Goals

- Reproduce all original seed data faithfully in Goravel format
- Idempotent upserts — safe to re-run without destroying existing data
- 50 games (expanded from original 1) with duplicated levels and content for development
- Single `DatabaseSeeder` orchestrator calling individual seeders in dependency order

## File Structure

```
dx-api/database/seeders/
├── database_seeder.go        # Orchestrator — calls all seeders in order
├── adm_user_seeder.go        # 30 admin users (upsert by username)
├── adm_permit_seeder.go      # 6 admin permissions (upsert by slug)
├── adm_role_seeder.go        # 1 admin role + role-permit link (upsert by slug)
├── adm_menu_seeder.go        # 6 parent + 26 child menus (upsert by name+parent)
├── game_category_seeder.go   # 4 parent + 11 child categories (upsert by name+parent)
├── game_press_seeder.go      # 22 presses (upsert by name)
├── user_seeder.go            # 100 users (upsert by username)
├── game_seeder.go            # 50 games (upsert by name)
├── game_level_seeder.go      # 150 levels, 3 per game (upsert by name+gameId)
├── content_meta_seeder.go    # 450 metas, 9 per game (upsert by sourceData+levelId)
└── content_item_seeder.go    # ~2,250 items, ~45 per game (upsert by content+type+metaId)
```

Plus `bootstrap/seeders.go` for registration.

## Dependency Chain

```
AdmUserSeeder
AdmPermitSeeder
AdmRoleSeeder       → depends on AdmPermitSeeder (role-permit links)
AdmMenuSeeder
GameCategorySeeder
GamePressSeeder
UserSeeder
GameSeeder          → depends on UserSeeder, GameCategorySeeder, GamePressSeeder
GameLevelSeeder     → depends on GameSeeder
ContentMetaSeeder   → depends on GameLevelSeeder
ContentItemSeeder   → depends on ContentMetaSeeder, GameLevelSeeder
```

The `DatabaseSeeder` calls them in this exact order via `facades.Seeder().Call()`.

## DatabaseSeeder

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

## Upsert Strategy

Every seeder uses idempotent upserts — find by unique key(s), update if exists, create if not.

**Important:** Goravel's `First()` returns a zero-value struct (empty `ID`) when no record is found, rather than always returning an error. The correct not-found check is:

```go
// Pattern used in each seeder (not a shared function — each seeder inlines its own)
var existing Model
if err := query.Where("unique_field", value).First(&existing); err != nil || existing.ID == "" {
    // Not found — create with new ULID
    return query.Create(&model)
}
// Found — update fields (preserve existing ID)
return query.Where("unique_field", value).Updates(&model)
```

**Null parent_id handling:** For `AdmMenuSeeder` and `GameCategorySeeder`, top-level items have `parent_id IS NULL`. Use `.WhereNull("parent_id")` instead of `.Where("parent_id", nil)` to generate correct SQL.

**Single-writer assumption:** These seeders are designed for single-writer execution (dev setup, CI). No unique indexes on upsert keys are required — the application-level find-then-create is sufficient.

Unique keys per seeder:

| Seeder | Upsert Key(s) |
|--------|---------------|
| AdmUserSeeder | `username` |
| AdmPermitSeeder | `slug` |
| AdmRoleSeeder | `slug` (role), `adm_role_id` + `adm_permit_id` (junction) |
| AdmMenuSeeder | `name` + `parent_id` (use `WhereNull` for top-level) |
| GameCategorySeeder | `name` + `parent_id` (use `WhereNull` for top-level) |
| GamePressSeeder | `name` |
| UserSeeder | `username` |
| GameSeeder | `name` |
| GameLevelSeeder | `name` + `game_id` |
| ContentMetaSeeder | `source_data` + `game_level_id` |
| ContentItemSeeder | `content` + `content_type` + `content_meta_id` (all seeded items have non-null `content_meta_id`) |

## ID Generation

Same pattern as existing dx-api code:

```go
ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
```

Only generated on create. Existing IDs preserved on update.

## Password Hashing

Uses existing `helpers.HashPassword()` from `dx-api/app/helpers/hash.go`. This function returns `(string, error)` — each seeder must handle the error at the top of `Run()` and return early on failure.

- Admin users: `helpers.HashPassword("password123")`
- Regular users: `helpers.HashPassword("Password123")`

Hashed once at seeder start, reused for all records in that seeder:

```go
func (s *UserSeeder) Run() error {
    hashedPw, err := helpers.HashPassword("Password123")
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    // ... use hashedPw for all records
}
```

## FK Resolution

Seeders resolve foreign keys by querying parent records by name/slug — same pattern as the original Prisma seeders:

- `GameSeeder` → queries `users.username = "rainson"`, `game_categories.name`, `game_presses.name`
- `GameLevelSeeder` → queries `games.name`
- `ContentMetaSeeder` → queries `game_levels.name` + `game_levels.game_id`
- `ContentItemSeeder` → queries `content_metas.source_data` + `game_levels.name`

If a dependency is missing, the seeder returns an error (fail-fast).

## Seed Data

### AdmUserSeeder (30 records)

30 admin users, all with `IsActive: true` and password hashed via `helpers.HashPassword("password123")`:
`admin`, `manager`, `editor`, `moderator`, `support`, `analyst`, `developer`, `tester`, `designer`, `marketing`, `sales`, `finance`, `hr`, `ops`, `content`, `reviewer`, `auditor`, `trainer`, `consultant`, `partner`, `vendor`, `inventory`, `logistics`, `quality`, `compliance`, `security`, `backup`, `network`, `database`, `sysadmin`
Note: `Nickname` is `*string` — assign via pointer (`&nickname`).

### AdmPermitSeeder (6 records)

| Slug | Name | HTTP Methods | HTTP Paths |
|------|------|-------------|------------|
| `*` | All permissions | `[]string{}` | `[]string{"*"}` |
| `adm.dashboard` | Admin dashboard | `[]string{"GET"}` | `[]string{}` |
| `auth.login` | Admin login | `[]string{}` | `[]string{"/login", "/logout"}` |
| `adm.users` | Admin users | `[]string{}` | `[]string{"/adm-users/*"}` |
| `adm.roles` | Admin roles | `[]string{}` | `[]string{"/adm-roles/*"}` |
| `adm.permits` | Admin permits | `[]string{}` | `[]string{"/adm-permits/*"}` |

Note: `HttpMethods` and `HttpPaths` are `pq.StringArray` (`[]string`). Use `[]string{}` for empty (not `nil`) to store `{}` in PostgreSQL, not NULL.

### AdmRoleSeeder (1 role + 1 link)

Role `admin` / "Admin" linked to permit `*` (all permissions).
- Role upserted by `slug`
- Junction record (`adm_role_permits`) upserted by composite key `adm_role_id` + `adm_permit_id`

### AdmMenuSeeder (32 records)

6 parent menus: Dashboard, System, Settings, Materials, Games, Users
26 child menus nested under their parents (same data as original seed).
Original data recoverable via: `git -C dx-web show 685491d^:prisma/seeds/adm-menus-seed.ts`

### GameCategorySeeder (15 records)

4 parents: 同步练习, 应试练习, 分级练习, 实用英语 (all `IsEnabled: true`)
11 children under 同步练习: 一年级 through 中职 (all `IsEnabled: true`)
Order values: parents 1000–4000, children 1000–11000 (matching original seed).

### GamePressSeeder (22 records)

22 publishers: 人教版, 沪教版, 冀教版, 外研社版, 译林版, 北京版, 北师大版, 川教版, 教科版, 接力版, 科普版, 辽师大版, 鲁科版, 闽教版, 湘鲁版, 陕旅版, 湘少版, 粤人版, 重大版, EEC 版, 牛津上海版, 清华版
Order values: 1000–22000 (increment by 1000, matching original seed).

### UserSeeder (100 records)

- `rainson` / "Rainson" / grade "lifetime" / email rainsonma@gmail.com / `IsActive: true`
- `june` / "June" / grade "lifetime" / `IsActive: true`
- `user003` through `user100` / "用户003"–"用户100" / grade default / `IsActive: true`

Note: `Nickname` and `Email` are `*string` — assign via pointer.

All with password "Password123" and invite codes generated via `helpers.GenerateInviteCode(8)`.

### GameSeeder (50 records)

Generated by combining child categories × presses × volumes (上册/下册):

```
一年级上册 (人教版)     — category: 一年级, press: 人教版
一年级下册 (人教版)     — category: 一年级, press: 人教版
一年级上册 (沪教版)     — category: 一年级, press: 沪教版
一年级下册 (沪教版)     — category: 一年级, press: 沪教版
二年级上册 (人教版)     — category: 二年级, press: 人教版
二年级下册 (人教版)     — category: 二年级, press: 人教版
... first 50 combinations
```

All games: `mode: "lsrw"`, `status: "published"`, `IsActive: true`, `UserID: &rainson.ID` (pointer).
`GameCategoryID` and `GamePressID` are also `*string` pointers — assign via `&id`.
Order increments by 1000.

**Name format:** `fmt.Sprintf("%s%s (%s)", categoryName, volume, pressName)` — e.g., `"一年级上册 (人教版)"`. This format guarantees uniqueness across the 50 games.

### GameLevelSeeder (150 records)

3 levels per game (same names for all games):

| Name | Order | Passing Score | IsActive |
|------|-------|--------------|----------|
| 第一关 | 1000 | 60 | true |
| 第二关 | 2000 | 60 | true |
| 第三关 | 3000 | 60 | true |

### ContentMetaSeeder (450 records)

9 metas per game (3 per level), same sentences for all games:

**第一关:** "The food is ready." / "I am very hungry." / "It is a good day."
**第二关:** "A car is on the road." / "It is a red car." / "The driver is happy."
**第三关:** "The children go to school." / "The bell rings." / "They go home."

All: `sourceFrom: "manual"`, `sourceType: "sentence"`, `isBreakDone: true`.

### ContentItemSeeder (~2,250 records)

~45 content items per game (duplicated from original seed data). Each sentence is broken into word/phrase/block/sentence content items with:
- `contentType`: word, phrase, block, sentence
- `translation`: Chinese translation
- `items`: JSONB array with phonetics (uk/us), part of speech, position, answer flag
- `order`: sequential within each meta

Full item data matches the original `content-items-seed.ts` exactly.
Original data recoverable via: `git -C dx-web show 685491d^:prisma/seeds/content-items-seed.ts`

## Bootstrap Registration

### New file: `bootstrap/seeders.go`

`WithSeeders` expects a **function** returning a slice (same pattern as `WithMigrations`):

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

### Modified: `bootstrap/app.go`

Add `.WithSeeders(Seeders)` to the setup chain (matches existing `.WithMigrations(Migrations)` pattern).

## Usage

```bash
# Run all seeders
go run . artisan db:seed

# Run a specific seeder
go run . artisan db:seed --seeder=UserSeeder

# Reset DB and seed
go run . artisan migrate:fresh --seed
```

## Data Volume Summary

| Seeder | Records | Upsert Key |
|--------|---------|------------|
| AdmUserSeeder | 30 | username |
| AdmPermitSeeder | 6 | slug |
| AdmRoleSeeder | 1 + 1 | slug |
| AdmMenuSeeder | 32 | name+parent_id |
| GameCategorySeeder | 15 | name+parent_id |
| GamePressSeeder | 22 | name |
| UserSeeder | 100 | username |
| GameSeeder | 50 | name |
| GameLevelSeeder | 150 | name+game_id |
| ContentMetaSeeder | 450 | source_data+game_level_id |
| ContentItemSeeder | ~2,250 | content+content_type+content_meta_id |
| **Total** | **~3,107** | |
