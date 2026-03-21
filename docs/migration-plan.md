# Douxue: Server-Side Migration Plan

> **Migrate all server-side logic from dx-web (Next.js) to dx-api (Goravel) as a standalone API provider.**

## Overview

| Item | Detail |
|------|--------|
| **Goal** | dx-api becomes the sole API backend; dx-web becomes a pure frontend client |
| **Strategy** | Gradual, domain-by-domain, improving structure along the way |
| **Schema ownership** | Prisma manages DB schema during migration; Goravel maps to existing tables. Transfer ownership to Goravel after full migration |
| **Auth** | Goravel JWT as single auth source — separate JWT guards for client users and admin users |
| **Frontend ↔ API** | Direct API calls via `NEXT_PUBLIC_API_URL` env var, no proxy |
| **Response format** | Envelope: `{"code": 0, "message": "ok", "data": {...}}` |
| **Pagination** | Cursor-based (public) + offset-based (admin) inside envelope |
| **dx-api port** | 3001 |
| **dx-web port** | 3000 (unchanged) |
| **API structure** | Two route groups: `/api/*` (client APIs for web/WeChat/Flutter) and `/adm/*` (admin APIs for admin portal) |

## Current Status

**Progress: 147/302 tasks completed (Phases 0-3 done)**

| Phase | Status |
|-------|--------|
| P0: Infrastructure & Config | COMPLETED |
| P1: Auth (Sign In / Sign Up / JWT) | COMPLETED |
| P2: User Profile & Settings | COMPLETED |
| P3: Games & Content (Read-Only) | COMPLETED |
| P4: Game Sessions & Gameplay | **NEXT** |
| P5-P11 | Not started |

**Next step:** Start Phase 4 — the most complex phase (game sessions, scoring, combos, stats, DB transactions).

## Progress Legend

- [ ] Not started
- [x] Completed
- [~] In progress

---

## Phase 0: Infrastructure & Config [COMPLETED]

> **Goal:** A running Go API server with DB, Redis, middleware, response/pagination helpers, all constants, all ORM models, and a health check. No business logic yet — just the skeleton every phase depends on.
>
> **Status:** All tasks completed. `go build` and `go vet` pass clean. 8 helpers, 5 middleware, 19 constants, 46 models, 2 route files created. Frontend API client created.

### Architecture

```
┌─────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  dx-web     │  │  WeChat Mini-App │  │  Flutter App     │
│  (Next.js)  │  │  (future)        │  │  (future)        │
└──────┬──────┘  └────────┬─────────┘  └────────┬─────────┘
       │                  │                     │
       └──────────────────┼─────────────────────┘
                          │  HTTP / JWT (user)
                   ┌──────▼──────┐
                   │   dx-api    │
                   │  (Goravel)  │
                   ├─────────────┤
                   │ /api/*      │ ← Client APIs (user JWT)
                   │ /adm/*      │ ← Admin APIs (admin JWT + RBAC)
                   ├─────────────┤
                   │  PostgreSQL │
                   │  Redis      │
                   └─────────────┘
```

- **`/api/*`** — Client APIs consumed by dx-web, WeChat, Flutter. Authenticated via user JWT (`users` table).
- **`/adm/*`** — Admin APIs consumed by the admin portal. Authenticated via admin JWT (`adm_users` table) with RBAC permission checks (`adm_permits`, `adm_roles`). Admin endpoints will be added incrementally after the client API migration is complete.

### Code Structure (api vs adm split)

Controllers, services, and requests are split into `api/` and `adm/` subdirectories. Models, constants, and helpers are shared.

```
dx-api/
├── routes/
│   ├── api.go                              # /api/* client routes
│   └── adm.go                              # /adm/* admin routes
├── app/
│   ├── http/
│   │   ├── controllers/
│   │   │   ├── api/                        # Client controllers
│   │   │   │   ├── auth_controller.go
│   │   │   │   ├── user_controller.go
│   │   │   │   ├── game_controller.go
│   │   │   │   ├── game_session_controller.go
│   │   │   │   ├── content_controller.go
│   │   │   │   ├── favorite_controller.go
│   │   │   │   ├── tracking_controller.go
│   │   │   │   ├── hall_controller.go
│   │   │   │   ├── leaderboard_controller.go
│   │   │   │   ├── notice_controller.go
│   │   │   │   ├── feedback_controller.go
│   │   │   │   ├── referral_controller.go
│   │   │   │   ├── redeem_controller.go
│   │   │   │   ├── upload_controller.go
│   │   │   │   ├── course_game_controller.go
│   │   │   │   └── ai_custom_controller.go
│   │   │   └── adm/                        # Admin controllers
│   │   │       ├── auth_controller.go
│   │   │       ├── notice_controller.go
│   │   │       ├── redeem_controller.go
│   │   │       └── ...                     # more added later
│   │   ├── middleware/
│   │   │   ├── jwt_auth.go                 # Client user JWT
│   │   │   ├── adm_jwt_auth.go             # Admin user JWT
│   │   │   ├── adm_rbac.go                 # Admin RBAC
│   │   │   ├── adm_operate_log.go          # Admin audit log
│   │   │   └── rate_limit.go               # Shared
│   │   └── requests/
│   │       ├── api/                        # Client validations
│   │       │   ├── auth_request.go
│   │       │   ├── user_request.go
│   │       │   ├── game_request.go
│   │       │   ├── session_request.go
│   │       │   ├── upload_request.go
│   │       │   ├── course_game_request.go
│   │       │   └── ai_custom_request.go
│   │       └── adm/                        # Admin validations
│   │           ├── auth_request.go
│   │           └── ...                     # more added later
│   ├── services/
│   │   ├── api/                            # Client business logic
│   │   │   ├── auth_service.go
│   │   │   ├── user_service.go
│   │   │   ├── game_service.go
│   │   │   ├── session_service.go
│   │   │   ├── content_service.go
│   │   │   ├── stats_service.go
│   │   │   ├── tracking_service.go
│   │   │   ├── favorite_service.go
│   │   │   ├── hall_service.go
│   │   │   ├── leaderboard_service.go
│   │   │   ├── notice_service.go
│   │   │   ├── feedback_service.go
│   │   │   ├── referral_service.go
│   │   │   ├── redeem_service.go
│   │   │   ├── upload_service.go
│   │   │   ├── course_game_service.go
│   │   │   ├── course_content_service.go
│   │   │   ├── ai_custom_service.go
│   │   │   ├── content_seek_service.go
│   │   │   └── bean_service.go
│   │   ├── adm/                            # Admin business logic
│   │   │   ├── auth_service.go
│   │   │   ├── notice_service.go
│   │   │   ├── redeem_service.go
│   │   │   └── ...                         # more added later
│   │   └── shared/                         # Shared across api & adm
│   │       └── email_service.go            # Email sending (used by both)
│   ├── models/                             # Shared — same DB tables
│   ├── constants/                          # Shared — same domain values
│   └── helpers/                            # Shared — same utilities
```

### 0.1 Environment & Config

- [x] **0.1.1** Configure `.env` with all required variables:
  ```env
  APP_NAME=Douxue
  APP_ENV=local
  APP_DEBUG=true
  APP_KEY=<generate>
  APP_URL=http://localhost
  APP_HOST=127.0.0.1
  APP_PORT=3001

  JWT_SECRET=<generate-strong-secret>

  DB_CONNECTION=postgres
  DB_HOST=127.0.0.1
  DB_PORT=5432
  DB_DATABASE=douxue
  DB_USERNAME=<your-user>
  DB_PASSWORD=<your-password>

  REDIS_HOST=127.0.0.1
  REDIS_PORT=6379
  REDIS_PASSWORD=
  REDIS_DB=0

  MAIL_HOST=<smtp-host>
  MAIL_PORT=465
  MAIL_USERNAME=<smtp-user>
  MAIL_PASSWORD=<smtp-pass>
  MAIL_FROM_ADDRESS=<from-email>
  MAIL_FROM_NAME=Douxue

  DEEPSEEK_API_KEY=<deepseek-key>
  DEEPSEEK_BASE_URL=https://api.deepseek.com

  STORAGE_PATH=storage/app
  ```

- [x] **0.1.2** Update `config/http.go`:
  - Port → `3001` (from env `APP_PORT`)
  - Request timeout → 30s (AI endpoints need longer)

- [x] **0.1.3** Update `config/database.go`:
  - Ensure PostgreSQL connection reads from `DB_*` env vars
  - Pool settings: `max_idle=10`, `max_open=100`, `idle_timeout=3600s`
  - Singular table names OFF (Prisma uses plural by convention — verify actual table names)

- [x] **0.1.4** Update `config/cache.go`:
  - Default driver → `redis`
  - Redis connection from `REDIS_*` env vars

- [x] **0.1.5** Update `config/jwt.go`:
  - Secret from `JWT_SECRET` env var
  - TTL: 60 min
  - Refresh TTL: 20160 min (2 weeks)
  - Two JWT guards: `"user"` (for client APIs, `users` table) and `"admin"` (for admin APIs, `adm_users` table)

- [x] **0.1.6** Update `config/mail.go`:
  - SMTP settings from `MAIL_*` env vars

- [x] **0.1.7** Update `config/cors.go`:
  - Allow origins: `http://localhost:3000` (dev), production URL
  - Allow credentials: true
  - Allow headers: Authorization, Content-Type, X-Requested-With

- [x] **0.1.8** Update `config/hashing.go`:
  - Bcrypt rounds: 12 (matches current dx-web bcrypt config)

- [x] **0.1.9** Update `config/filesystems.go`:
  - Local disk root → `storage/app`

### 0.2 Helpers

- [x] **0.2.1** `app/helpers/response.go` — Envelope response builders:
  ```go
  // Success returns {"code": 0, "message": "ok", "data": data}
  func Success(ctx http.Context, data any) http.Response

  // Error returns {"code": code, "message": msg, "data": nil}
  func Error(ctx http.Context, httpStatus int, code int, message string) http.Response

  // Paginated returns {"code": 0, "message": "ok", "data": {"items": [], "nextCursor": "", "hasMore": bool}}
  func Paginated(ctx http.Context, items any, nextCursor string, hasMore bool) http.Response

  // PaginatedOffset returns {"code": 0, "message": "ok", "data": {"items": [], "total": n, "page": p, "pageSize": s}}
  func PaginatedOffset(ctx http.Context, items any, total int64, page int, pageSize int) http.Response
  ```

- [x] **0.2.2** `app/helpers/pagination.go` — Cursor & offset pagination utilities:
  ```go
  // ParseCursorParams extracts cursor and limit from query string
  func ParseCursorParams(ctx http.Context, defaultLimit int) (cursor string, limit int)

  // ParseOffsetParams extracts page and pageSize from query string
  func ParseOffsetParams(ctx http.Context, defaultPageSize int) (page int, pageSize int, offset int)
  ```

- [x] **0.2.3** `app/helpers/redis.go` — Redis client singleton:
  ```go
  // GetRedis returns the shared Redis client
  func GetRedis() *redis.Client

  // SetWithTTL sets a key with expiration
  func SetWithTTL(key string, value string, ttl time.Duration) error

  // Get retrieves a value by key
  func Get(key string) (string, error)

  // Del deletes a key
  func Del(key string) error
  ```

- [x] **0.2.4** `app/helpers/rate_limit.go` — Sliding window rate limiter:
  ```go
  // CheckRateLimit returns true if the request is within rate limit
  // Uses Redis sorted sets (same algorithm as dx-web)
  func CheckRateLimit(key string, limit int, windowSeconds int) (bool, error)
  ```

- [x] **0.2.5** `app/helpers/hash.go` — Bcrypt helpers:
  ```go
  func HashPassword(password string) (string, error)
  func CheckPassword(password string, hash string) bool
  ```

- [x] **0.2.6** `app/helpers/random.go` — Code generation:
  ```go
  // GenerateCode returns a random N-digit numeric string (e.g. "482916")
  func GenerateCode(length int) string

  // GenerateInviteCode returns a random alphanumeric invite code
  func GenerateInviteCode(length int) string
  ```

- [x] **0.2.7** `app/helpers/assert_fk.go` — Foreign key assertion:
  ```go
  // AssertFK performs a SELECT ... FOR UPDATE to lock a row and verify it exists
  // Used in transactional writes to ensure referenced records exist
  func AssertFK(tx orm.Transaction, table string, id string) error
  ```

- [x] **0.2.8** `app/helpers/sse.go` — Server-Sent Events:
  ```go
  // SSEWriter wraps http response for streaming
  type SSEWriter struct { ... }

  // NewSSEWriter creates writer and sets SSE headers
  func NewSSEWriter(ctx http.Context) *SSEWriter

  // Write sends a data event
  func (w *SSEWriter) Write(data any) error

  // Close sends done event and flushes
  func (w *SSEWriter) Close()
  ```

### 0.3 Middleware

- [x] **0.3.1** `app/http/middleware/jwt_auth.go`:
  - Extract JWT from `Authorization: Bearer <token>` header
  - Validate token via Goravel auth facade
  - Attach user ID to context
  - Return `{"code": 40100, "message": "unauthorized"}` on failure

- [x] **0.3.2** `app/http/middleware/adm_jwt_auth.go`:
  - Separate JWT guard for admin users (validates against `adm_users` table)
  - Extract JWT from `Authorization: Bearer <token>` header
  - Attach admin user ID to context
  - Return `{"code": 40100, "message": "unauthorized"}` on failure

- [x] **0.3.3** `app/http/middleware/adm_rbac.go`:
  - RBAC permission check for admin routes
  - Loads admin user's permissions (direct + role-based) from `adm_permits`
  - Matches request HTTP method + path against permitted `http_methods` + `http_paths`
  - Return `{"code": 40300, "message": "forbidden"}` on insufficient permissions

- [x] **0.3.4** `app/http/middleware/adm_operate_log.go`:
  - Logs admin operations to `adm_operates` table
  - Records: admin_user_id, path, method, ip, input (request body)

- [x] **0.3.5** `app/http/middleware/rate_limit.go`:
  - Configurable rate limit per route group
  - Uses `helpers.CheckRateLimit()` with user ID + route as key
  - Return `{"code": 42900, "message": "too many requests"}` on failure

### 0.4 Constants

Migrate all 18 files from `dx-web/src/consts/` → `dx-api/app/constants/`:

- [x] **0.4.1** `error_code.go` — Centralized API error codes:
  ```go
  // Success
  CodeSuccess = 0

  // 400xx: Validation
  CodeValidationError   = 40000
  CodeInvalidEmail      = 40001
  CodeInvalidPassword   = 40002
  CodeDuplicateEmail    = 40003
  CodeDuplicateUsername = 40004
  CodeInvalidCode       = 40005
  CodeCodeExpired       = 40006
  CodeInsufficientBeans = 40007

  // 401xx: Auth
  CodeUnauthorized      = 40100
  CodeTokenExpired      = 40101
  CodeInvalidToken      = 40102

  // 403xx: Permission
  CodeForbidden         = 40300

  // 404xx: Not Found
  CodeNotFound          = 40400
  CodeUserNotFound      = 40401
  CodeGameNotFound      = 40402
  CodeSessionNotFound   = 40403
  CodeLevelNotFound     = 40404
  CodeContentNotFound   = 40405

  // 429xx: Rate Limit
  CodeRateLimited       = 42900

  // 500xx: Server Error
  CodeInternalError     = 50000
  CodeAIServiceError    = 50001
  CodeEmailSendError    = 50002
  ```

- [x] **0.4.2** `game_degree.go` — practice, beginner, intermediate, advanced + content type mapping
- [x] **0.4.3** `game_mode.go` — lsrw, vocab-battle, vocab-match, vocab-elimination, listening-challenge
- [x] **0.4.4** `game_pattern.go` — listen, speak, read, write (default: write)
- [x] **0.4.5** `game_status.go` — draft, published, withdraw
- [x] **0.4.6** `content_type.go` — word, block, phrase, sentence
- [x] **0.4.7** `difficulty.go` — a1-a2, b1-b2, c1-c2
- [x] **0.4.8** `scoring.go` — combo bonuses, exp threshold (0.6), level complete exp (10)
- [x] **0.4.9** `score_rating.go` — rating thresholds and labels
- [x] **0.4.10** `review_interval.go` — spaced repetition: [1, 3, 7, 14, 30, 90] days + `GetNextReviewAt()`
- [x] **0.4.11** `user_grade.go` — free, month, season, year, lifetime + prices + month durations
- [x] **0.4.12** `user_level.go` — level progression: base=1000, multiplier=1.05, max=100, `GetLevel()`, `GetExpForLevel()`
- [x] **0.4.13** `bean_slug.go` — all 13 bean transaction type slugs
- [x] **0.4.14** `bean_reason.go` — human-readable reason labels (Chinese)
- [x] **0.4.15** `image_role.go` — 8 image roles
- [x] **0.4.16** `source_type.go` — sentence, vocab
- [x] **0.4.17** `source_from.go` — manual, ai
- [x] **0.4.18** `feedback_type.go` — feature, content, ux, bug, other
- [x] **0.4.19** `referral_status.go` — pending, paid, rewarded

### 0.5 ORM Models

Define all Goravel model structs mapping to existing Prisma tables. Each model in its own file under `app/models/`. Must match Prisma table names and column names exactly.

**User domain:**
- [x] **0.5.1** `user.go` — fields: id, grade, username, nickname, email, phone, password, avatar_id, city, introduction, beans, granted_beans, exp, invite_code, current_play_streak, max_play_streak, last_played_at, vip_due_at, last_read_notice_at, is_active, created_at, updated_at
- [x] **0.5.2** `user_login.go` — id, user_id, ip, agent, country, province, city, isp
- [x] **0.5.3** `user_setting.go` — id, user_id, group, key, value, value_type, created_at, updated_at
- [x] **0.5.4** `user_bean.go` — id, user_id, beans, origin, result, slug, reason, data
- [x] **0.5.5** `user_favorite.go` — id, user_id, game_id
- [x] **0.5.6** `user_master.go` — id, user_id, content_item_id, game_id, game_level_id
- [x] **0.5.7** `user_unknown.go` — id, user_id, content_item_id, game_id, game_level_id
- [x] **0.5.8** `user_review.go` — id, user_id, content_item_id, game_id, game_level_id, last_review_at, next_review_at, review_count
- [x] **0.5.9** `user_referral.go` — id, referrer_id, invitee_id, status, reward_amount, rewarded_at
- [x] **0.5.10** `user_redeem.go` — id, code, grade, user_id, redeemed_at
- [x] **0.5.11** `user_follow.go` — id, follower_id, following_id

**Game domain:**
- [x] **0.5.12** `game.go` — id, name, description, user_id, mode, game_category_id, game_press_id, icon, cover_id, order, is_active, status
- [x] **0.5.13** `game_level.go` — id, game_id, name, description, order, passing_score, is_active
- [x] **0.5.14** `game_category.go` — id, parent_id, cover_id, name, alias, description, order, is_enabled (self-referencing for hierarchy)
- [x] **0.5.15** `game_press.go` — id, name, cover_id, order
- [x] **0.5.16** `game_group.go` — id, name, description, owner_id, cover_id, current_game_id, invite_code, is_active, created_at, updated_at
- [x] **0.5.17** `game_subgroup.go` — id, game_group_id, game_id, description

**Session & stats domain:**
- [x] **0.5.18** `game_session_total.go` — id, user_id, game_id, degree, pattern, current_level_id, current_content_item_id, started_at, ended_at, score, exp, max_combo, correct_count, wrong_count, skip_count, play_time, total_levels_count, played_levels_count
- [x] **0.5.19** `game_session_level.go` — id, game_session_total_id, game_level_id, degree, pattern, current_content_item_id, started_at, ended_at, score, exp, max_combo, correct_count, wrong_count, skip_count, play_time, total_items_count, played_items_count
- [x] **0.5.20** `game_record.go` — id, user_id, game_session_total_id, game_session_level_id, game_level_id, content_item_id, is_correct, source_answer, user_answer, base_score, combo_score, duration
- [x] **0.5.21** `game_stats_total.go` — id, user_id, game_id, total_sessions, total_exp, highest_score, total_scores, total_play_time, first_played_at, last_played_at, first_completed_at, last_completed_at, completion_count
- [x] **0.5.22** `game_stats_level.go` — id, user_id, game_level_id, total_sessions, best_score, total_play_time, first_played_at, last_played_at, first_completed_at, last_completed_at, completion_count

**Content domain:**
- [x] **0.5.23** `content_item.go` — id, game_level_id, content_meta_id, content, content_type, uk_audio_id, us_audio_id, definition, translation, explanation, items (JSON), structure (JSON), order, tags, is_active
- [x] **0.5.24** `content_meta.go` — id, game_level_id, source_type, source_from, source_data, translation, is_break_done, order, is_active, created_at, updated_at
- [x] **0.5.25** `content_seek.go` — id, user_id, course_name, description, disk_url, count, created_at, updated_at

**Community domain:**
- [x] **0.5.26** `post.go` — id, user_id, content, image_id, tags, like_count, comment_count, share_count, is_active
- [x] **0.5.27** `post_comment.go` — id, post_id, user_id, content, parent_comment_id
- [x] **0.5.28** `post_like.go` — id, user_id, post_id
- [x] **0.5.29** `post_bookmark.go` — id, user_id, post_id

**System domain:**
- [x] **0.5.30** `image.go` — id, adm_user_id, user_id, url, name, mime, size, role
- [x] **0.5.31** `audio.go` — id, adm_user_id, user_id, url, name, mime, size, duration, role
- [x] **0.5.32** `notice.go` — id, title, content, icon, is_active
- [x] **0.5.33** `feedback.go` — id, user_id, type, description
- [x] **0.5.34** `game_report.go` — id, user_id, game_id, game_level_id, content_item_id, reason, note
- [x] **0.5.35** `game_group_member.go` — id, game_group_id, user_id, role (owner/member), created_at, updated_at
- [x] **0.5.36** `game_subgroup_member.go` — id, game_subgroup_id, user_id, role (leader/member), created_at, updated_at
- [x] **0.5.37** `setting.go` — id, group, label, key, value, value_type, value_from, value_options (JSON), description, order, is_enabled, created_at, updated_at

**Admin domain (all 9 models — needed for admin JWT auth, RBAC, and future admin API endpoints):**
- [x] **0.5.38** `adm_user.go` — id, username, nickname, password, avatar_id, is_active, created_at, updated_at
- [x] **0.5.39** `adm_role.go` — id, slug, name, created_at, updated_at
- [x] **0.5.40** `adm_permit.go` — id, slug, name, http_methods (string[]), http_paths (string[]), created_at, updated_at
- [x] **0.5.41** `adm_user_role.go` — id, adm_user_id, adm_role_id (junction: user ↔ role)
- [x] **0.5.42** `adm_user_permit.go` — id, adm_user_id, adm_permit_id (junction: user ↔ direct permit)
- [x] **0.5.43** `adm_role_permit.go` — id, adm_role_id, adm_permit_id (junction: role ↔ permit)
- [x] **0.5.44** `adm_login.go` — id, adm_user_id, ip, agent, country, province, city, isp, created_at, updated_at
- [x] **0.5.45** `adm_operate.go` — id, adm_user_id, path, method, ip, input, created_at, updated_at
- [x] **0.5.46** `adm_menu.go` — id, parent_id, name, alias, icon, uri, order, created_at, updated_at (self-referencing hierarchy)

> **Important:** Column names must exactly match the database. Before writing models, inspect actual table/column names with `\d table_name` in psql, since Prisma may use `@map` annotations that rename columns.

### 0.6 Routes & Health Check

- [x] **0.6.1** Set up route groups in `routes/api.go` (client APIs):
  ```go
  // Client API routes under /api prefix
  api := route.Prefix("/api")

  // Public routes (no auth)
  api.Get("/health", ...)

  // Auth routes (no auth required)
  auth := api.Prefix("/auth")

  // Protected routes (user JWT middleware)
  protected := api.Middleware(middleware.JwtAuth())
  ```

- [x] **0.6.2** Set up route groups in `routes/adm.go` (admin APIs):
  ```go
  // Admin API routes under /adm prefix
  adm := route.Prefix("/adm")

  // Admin auth routes (no auth required)
  adm.Post("/auth/login", ...)

  // Protected admin routes (admin JWT + RBAC + operation log)
  admProtected := adm.Middleware(
      middleware.AdmJwtAuth(),
      middleware.AdmRbac(),
      middleware.AdmOperateLog(),
  )

  // Admin endpoints will be added here incrementally
  ```

- [x] **0.6.3** Register `routes/adm.go` in `bootstrap/app.go`

- [x] **0.6.4** Health check endpoint — `GET /api/health`:
  - Verify DB connection
  - Verify Redis connection
  - Return `{"code": 0, "message": "ok", "data": {"db": true, "redis": true}}`

### 0.7 Frontend API Client

- [x] **0.7.1** Add `NEXT_PUBLIC_API_URL=http://localhost:3001` to dx-web `.env`

- [x] **0.7.2** Create `dx-web/src/lib/api-client.ts`:
  ```typescript
  // Base API client for calling dx-api
  // - Reads NEXT_PUBLIC_API_URL
  // - Attaches JWT from storage to Authorization header
  // - Handles token refresh on 401
  // - Parses envelope response
  // - Typed generic: apiClient.get<GameCard[]>("/api/games")
  ```

### 0.8 Verification

- [x] **0.8.1** `go build` succeeds
- [x] **0.8.2** Server starts on port 3001
- [x] **0.8.3** DB connection works (health check)
- [x] **0.8.4** Redis connection works (health check)
- [x] **0.8.5** CORS allows requests from localhost:3000

---

## Phase 1: Auth (Sign In / Sign Up / JWT) [COMPLETED]

> **Goal:** Users authenticate via Go API. NextAuth removed from dx-web. JWT issued by Goravel is the single auth token for all clients.
>
> **Status:** Backend + frontend complete. NextAuth removed. JWT cookie-based compatibility layer in place — all 37 files with `auth()` calls continue working. Frontend uses `authApi` client for signin/signup.

### 1.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Rate Limit | Description |
|--------|------|------|------------|-------------|
| POST | `/api/auth/signup/send-code` | No | 1/60s per email | Send signup verification code |
| POST | `/api/auth/signup` | No | - | Register new user |
| POST | `/api/auth/signin/send-code` | No | 1/60s per email | Send signin verification code |
| POST | `/api/auth/signin` | No | - | Login (email+code or account+password) |
| POST | `/api/auth/refresh` | Yes | - | Refresh JWT token |
| GET | `/api/auth/me` | Yes | - | Get current user profile |
| POST | `/api/auth/logout` | Yes | - | Invalidate token |

**Tasks:**

- [x] **1.1.1** `app/services/shared/email_service.go`:
  - `SendVerificationEmail(to, code string)` — compose HTML email, dispatch via Goravel mail/queue
  - Email template: verification code with styling (match current dx-web template)

- [x] **1.1.2** `app/services/api/auth_service.go`:
  - `SendSignUpCode(email)` — rate limit 1/60s, generate 6-digit code, store in Redis (TTL 300s), send email
  - `SignUp(email, code, username, password)` — verify code, check duplicate email/username, hash password, create user, generate invite code, handle referral (from request param), issue JWT
  - `SendSignInCode(email)` — rate limit 1/60s, generate code, store in Redis, send email
  - `SignInByEmail(email, code)` — verify code, find or auto-register user, issue JWT
  - `SignInByAccount(account, password)` — find by username/email/phone, verify bcrypt, issue JWT
  - `RefreshToken(userId)` — issue new JWT
  - `GetCurrentUser(userId)` — return user profile
  - `Logout(token)` — invalidate JWT (optional: add to Redis blacklist)
  - `RecordLogin(userId, ip, userAgent)` — create user_login entry on successful signin (extract geo from IP if available)

- [x] **1.1.3** `app/http/requests/api/auth_request.go`:
  - `SignUpRequest` — email (required, email format), code (required, 6 digits), username (optional, min 3), password (optional, min 8 with complexity)
  - `SignInRequest` — email+code OR account+password
  - `SendCodeRequest` — email (required, email format)

- [x] **1.1.4** `app/http/controllers/api/auth_controller.go`:
  - Thin handlers: parse request → call service → return envelope response
  - One method per endpoint

- [x] **1.1.5** `app/services/adm/auth_service.go` (admin auth scaffold):
  - `AdminSignIn(username, password)` — validate against `adm_users` table, verify bcrypt, check `is_active`, issue admin JWT, record login in `adm_logins`
  - `GetAdminUser(admUserId)` — return admin user profile with roles
  - `GetAdminPermissions(admUserId)` — load all permissions (direct + role-based) for RBAC middleware
  > **Note:** This sets up admin auth infrastructure. Admin CRUD endpoints will be added later.

- [x] **1.1.6** `app/http/controllers/adm/auth_controller.go`:
  - `POST /adm/auth/login` — admin login endpoint
  - `GET /adm/auth/me` — get current admin user (protected)
  - `POST /adm/auth/logout` — admin logout

- [x] **1.1.7** Register client auth routes in `routes/api.go` — public group
- [x] **1.1.8** Register admin auth routes in `routes/adm.go`

- [x] **1.1.9** Tests:
  - Signup flow: send code → signup → receive JWT
  - Signin by email: send code → signin → receive JWT
  - Signin by account: account+password → receive JWT
  - Token refresh
  - Rate limit on code sending
  - Duplicate email/username rejection
  - Invalid code rejection
  - Referral tracking on signup
  - Admin login with username+password
  - Admin RBAC middleware blocks unauthorized routes
  - Admin operation logging

### 1.2 Frontend (dx-web)

- [x] **1.2.1** Update `src/lib/api-client.ts`:
  - JWT storage (localStorage or httpOnly cookie via API)
  - Auto-attach `Authorization: Bearer <token>` header
  - Handle 401 → redirect to signin

- [x] **1.2.2** Create auth API functions (replace server actions):
  - `api.auth.sendSignUpCode(email)`
  - `api.auth.signUp({email, code, username, password})`
  - `api.auth.sendSignInCode(email)`
  - `api.auth.signIn({email, code} | {account, password})`
  - `api.auth.refresh()`
  - `api.auth.me()`
  - `api.auth.logout()`

- [x] **1.2.3** Update signin/signup page components to call API client instead of server actions

- [x] **1.2.4** Replace all `auth()` session checks with API client's `me()` or stored JWT decode

- [x] **1.2.5** Update `src/proxy.ts` (Next.js middleware):
  - Rewrite auth checks to use JWT from cookie/localStorage instead of NextAuth session
  - Redirect unauthenticated users from `/hall/*` to signin
  - Redirect authenticated users away from `/auth/*`

- [x] **1.2.6** Remove NextAuth:
  - Delete `src/lib/auth.ts`
  - Delete `src/app/api/auth/[...nextauth]/route.ts`
  - Delete `src/app/api/auth/invalidate/route.ts`
  - Remove `next-auth` from `package.json`

- [x] **1.2.7** Remove auth-related server actions:
  - Delete `src/features/web/auth/actions/signin.action.ts`
  - Delete `src/features/web/auth/actions/signup.action.ts`
  - Delete `src/features/web/auth/services/signin.service.ts`
  - Delete `src/features/web/auth/services/signup.service.ts`
  - Delete `src/features/web/auth/services/user.service.ts`

### 1.3 Verification

- [ ] **1.3.1** New user can sign up with email code
- [ ] **1.3.2** Existing user can sign in with email code
- [ ] **1.3.3** Existing user can sign in with account + password
- [ ] **1.3.4** JWT is stored and sent with subsequent requests
- [ ] **1.3.5** Protected pages redirect to signin when no token
- [ ] **1.3.6** Token refresh works before expiry
- [ ] **1.3.7** All existing dx-web pages still load correctly

---

## Phase 2: User Profile & Settings [COMPLETED]

> **Goal:** Profile management via Go API.
>
> **Status:** Backend (6 endpoints) + frontend complete. Server actions now proxy to Go API via `apiServerFetch`. Components unchanged.

### 2.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Rate Limit | Description |
|--------|------|------------|-------------|
| GET | `/api/user/profile` | - | Get full profile with level, avatar |
| PUT | `/api/user/profile` | - | Update nickname, city, introduction |
| PUT | `/api/user/avatar` | - | Set avatar from uploaded image ID |
| POST | `/api/user/email/send-code` | 1/60s | Send email change verification code |
| PUT | `/api/user/email` | - | Change email with code verification |
| PUT | `/api/user/password` | - | Change password (verify current) |

**Tasks:**

- [x] **2.1.1** `app/services/api/user_service.go`:
  - `GetProfile(userId)` — return user with computed level (from exp), avatar URL
  - `UpdateProfile(userId, nickname, city, introduction)` — validate uniqueness of nickname if changed
  - `UpdateAvatar(userId, imageId)` — verify image exists and belongs to user
  - `SendChangeEmailCode(userId, email)` — rate limit, verify new email not taken, send code
  - `ChangeEmail(userId, email, code)` — verify code, update email
  - `ChangePassword(userId, currentPassword, newPassword)` — verify current, hash new, update

- [x] **2.1.2** `app/http/requests/api/user_request.go`:
  - `UpdateProfileRequest` — nickname (optional, max 20), city (optional, max 50), introduction (optional, max 200)
  - `UpdateAvatarRequest` — image_id (required, uuid)
  - `ChangeEmailRequest` — email (required, email format), code (required, 6 digits)
  - `ChangePasswordRequest` — current_password (required), new_password (required, min 8 with complexity)

- [x] **2.1.3** `app/http/controllers/api/user_controller.go`

- [x] **2.1.4** Register routes — protected group

- [x] **2.1.5** Tests

### 2.2 Frontend (dx-web)

- [x] **2.2.1** Create user API functions (replace `me.action.ts`)
- [x] **2.2.2** Update profile page components to call API
- [x] **2.2.3** Remove `src/features/web/me/actions/me.action.ts`
- [x] **2.2.4** Remove `src/features/web/me/services/me.service.ts`

### 2.3 Verification

- [x] **2.3.1** Profile page loads with correct data
- [x] **2.3.2** Nickname/city/intro update works
- [x] **2.3.3** Avatar change works
- [x] **2.3.4** Email change with code works
- [x] **2.3.5** Password change works (rejects wrong current password)

---

## Phase 3: Games & Content (Read-Only) [COMPLETED]

> **Goal:** Game listing, search, categories, content items served from Go API.

### 3.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/games` | No | List published games (cursor, filters: categoryIds, pressId, mode) |
| GET | `/api/games/search` | No | Search by name (limit param) |
| GET | `/api/games/recent` | Yes | User's recently played games |
| GET | `/api/games/:id` | No | Game detail with levels |
| GET | `/api/games/:id/levels/:levelId/content` | Yes | Content items for a level (filtered by degree) |
| GET | `/api/game-categories` | No | Hierarchical categories |
| GET | `/api/game-presses` | No | Publishers list |

**Tasks:**

- [x] **3.1.1** `app/services/api/game_service.go`:
  - `ListPublishedGames(filters, cursor, limit)` — cursor pagination (12 per page), join cover/category/author
  - `SearchGames(query, limit)` — search by name, published only
  - `GetRecentGames(userId)` — from game_stats_total ordered by last_played_at
  - `GetGameDetail(gameId)` — game with levels, cover, category
  - `ListCategories()` — hierarchical categories with parent-child
  - `ListPresses()` — publishers list

- [x] **3.1.2** `app/services/api/content_service.go`:
  - `GetLevelContent(gameLevelId, degree)` — content items filtered by degree's content types, ordered by `order`

- [x] **3.1.3** `app/http/requests/api/game_request.go`:
  - `ListGamesRequest` — cursor (optional), limit (optional, default 12), category_ids (optional, array), press_id (optional), mode (optional)
  - `SearchGamesRequest` — query (required, min 1), limit (optional, default 10)
  - `LevelContentRequest` — degree (required, one of game degrees)

- [x] **3.1.4** `app/http/controllers/api/game_controller.go`, `app/http/controllers/api/content_controller.go`

- [x] **3.1.5** Register routes — mix of public and protected

- [x] **3.1.6** Tests

### 3.2 Frontend (dx-web)

- [x] **3.2.1** Create game/content API functions
- [x] **3.2.2** Update game listing, search, hall pages to call API
- [x] **3.2.3** Update play page content loading to call API
- [x] **3.2.4** Remove `src/features/web/games/actions/game.action.ts`
- [x] **3.2.5** Remove `src/features/web/hall/actions/game-search.action.ts`
- [x] **3.2.6** Remove `src/features/web/play/actions/content.action.ts`
- [x] **3.2.7** Remove related services and model query files that are no longer called from dx-web

### 3.3 Verification

- [x] **3.3.1** Game listing with pagination works
- [x] **3.3.2** Category/press filters work
- [x] **3.3.3** Game search returns results
- [x] **3.3.4** Recent games show for authenticated users
- [x] **3.3.5** Content items load correctly for each degree

---

## Phase 4: Game Sessions & Gameplay

> **Goal:** The core gameplay loop — sessions, levels, answers, scoring, stats. This is the most complex phase.

### 4.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Rate Limit | Description |
|--------|------|------------|-------------|
| POST | `/api/sessions/start` | - | Start game session |
| POST | `/api/sessions/:id/end` | - | End session |
| POST | `/api/sessions/:id/force-complete` | - | Force complete |
| POST | `/api/sessions/:id/levels/start` | - | Start a level |
| POST | `/api/sessions/:id/levels/:levelId/complete` | - | Complete level |
| POST | `/api/sessions/:id/levels/:levelId/advance` | - | Advance to next level |
| POST | `/api/sessions/:id/levels/:levelId/restart` | - | Restart level |
| POST | `/api/sessions/:id/answers` | 30/60s | Record answer |
| POST | `/api/sessions/:id/skips` | 30/60s | Record skip |
| POST | `/api/sessions/:id/sync-playtime` | - | Sync playtime |
| GET | `/api/sessions/active` | - | Check active session for game (by degree+pattern) |
| GET | `/api/sessions/active-level` | - | Check active level-session within a session |
| GET | `/api/sessions/any-active` | - | Check any active session for a game |
| GET | `/api/sessions/:id/restore` | - | Restore session data |
| PUT | `/api/sessions/:id/content-item` | - | Update current content item |

**Tasks:**

- [x] **4.1.1** `app/services/api/session_service.go` (largest service — consider splitting into sub-files):
  - `StartSession(userId, gameId, degree, pattern)` — check no active session, create session + first level
  - `EndSession(userId, sessionId)` — calculate final scores, mark ended, update stats
  - `ForceCompleteSession(userId, sessionId)` — complete all remaining, end session
  - `StartLevel(userId, sessionId, gameLevelId)` — create session level
  - `CompleteLevel(userId, sessionId, gameLevelId)` — mark level done, grant EXP if accuracy >= 60%
  - `AdvanceLevel(userId, sessionId)` — move to next level
  - `RestartLevel(userId, sessionId, gameLevelId)` — reset level stats
  - `RecordAnswer(userId, sessionId, gameLevelId, contentItemId, isCorrect, sourceAnswer, userAnswer, duration)` — rate limited, upsert record, update combo/scores, DB transaction
  - `RecordSkip(userId, sessionId, gameLevelId)` — rate limited, increment skip count
  - `SyncPlayTime(userId, sessionId, gameLevelId, playTime)` — validate 0-86400
  - `CheckActiveSession(userId, gameId, degree, pattern)` — find ongoing session
  - `CheckActiveLevelSession(userId, sessionId, gameLevelId)` — find active level-session within a session (for mid-level resume)
  - `CheckAnyActiveSession(userId, gameId)` — find any ongoing session
  - `RestoreSessionData(sessionId, gameLevelId)` — accumulated stats for resume
  - `UpdateCurrentContentItem(userId, sessionId, contentItemId)` — save resume point

- [x] **4.1.2** `app/services/api/stats_service.go`:
  - `UpsertGameStats(userId, gameId, sessionData)` — update total stats after session events
  - `UpdateGameStatsAfterSession(userId, gameId, sessionData)` — after session end
  - `UpsertLevelStats(userId, gameLevelId, levelData)` — after level complete
  - `MarkGameFirstCompletion(userId, gameId)` — first-time completion tracking

- [x] **4.1.3** `app/services/api/bean_service.go`:
  - `ConsumeBeans(userId, amount, slug, reason)` — debit beans, create ledger entry
  - `RefundBeans(userId, amount, slug, reason)` — credit beans back
  - `GetBalance(userId)` — current bean count

- [x] **4.1.4** `app/http/requests/api/session_request.go`:
  - Validation for each session endpoint
  - `StartSessionRequest` — game_id (required), degree (required), pattern (required)
  - `RecordAnswerRequest` — game_session_level_id (the session-level record ID, not the game level ID), game_level_id, content_item_id, is_correct, source_answer, user_answer, duration
  > **Note:** `game_session_level_id` is the session-level record; `game_level_id` is the game's level. Both are needed — the former for the record's FK, the latter for stats.

- [x] **4.1.5** `app/http/controllers/api/game_session_controller.go`

- [x] **4.1.6** Register routes — protected, some rate-limited

- [x] **4.1.7** DB transactions for atomic updates (session stats + game stats + records)

- [ ] **4.1.8** Tests — critical, test all state transitions:
  - Full session lifecycle: start → answer → complete level → advance → end
  - Scoring: combo bonuses, EXP threshold
  - Edge cases: duplicate answers, concurrent sessions, force complete
  - Rate limiting on answers/skips

### 4.2 Frontend (dx-web)

- [x] **4.2.1** Create session/gameplay API functions
- [x] **4.2.2** Update play page components to call API (this is the most complex frontend change)
- [x] **4.2.3** Update playtime sync (beacon API → regular POST to Go API)
- [ ] **4.2.4** Remove `src/features/web/play/services/session.service.ts`
- [ ] **4.2.5** Remove `src/app/api/play/sync-playtime/route.ts`
- [ ] **4.2.6** Remove related model files no longer called from dx-web

### 4.3 Verification

- [ ] **4.3.1** Start a game session
- [ ] **4.3.2** Answer questions and see score update
- [ ] **4.3.3** Combo bonuses apply correctly
- [ ] **4.3.4** Complete a level and see EXP granted (if accuracy >= 60%)
- [ ] **4.3.5** Advance to next level
- [ ] **4.3.6** End session and see stats updated
- [ ] **4.3.7** Resume an interrupted session
- [ ] **4.3.8** Force complete works
- [ ] **4.3.9** Playtime syncs correctly

---

## Phase 5: User Tracking (Master / Unknown / Review / Favorites)

> **Goal:** Word tracking and favorites via Go API.

### 5.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Rate Limit | Description |
|--------|------|------------|-------------|
| POST | `/api/tracking/master` | 30/60s | Mark content item as mastered |
| GET | `/api/tracking/master` | - | List mastered items (paginated) |
| GET | `/api/tracking/master/stats` | - | Mastered word count/stats |
| DELETE | `/api/tracking/master/:id` | - | Remove single mastered entry |
| DELETE | `/api/tracking/master` | - | Bulk delete mastered entries (body: ids[]) |
| POST | `/api/tracking/unknown` | 30/60s | Mark content item as unknown |
| GET | `/api/tracking/unknown` | - | List unknown items (paginated) |
| GET | `/api/tracking/unknown/stats` | - | Unknown word count/stats |
| DELETE | `/api/tracking/unknown/:id` | - | Remove single unknown entry |
| DELETE | `/api/tracking/unknown` | - | Bulk delete unknown entries (body: ids[]) |
| POST | `/api/tracking/review` | 30/60s | Mark content item for review |
| GET | `/api/tracking/review` | - | Review items list (paginated) |
| GET | `/api/tracking/review/stats` | - | Review stats (due/total) |
| DELETE | `/api/tracking/review/:id` | - | Remove single review entry |
| DELETE | `/api/tracking/review` | - | Bulk delete review entries (body: ids[]) |
| POST | `/api/favorites/toggle` | - | Toggle game favorite |
| GET | `/api/favorites` | - | List user's favorite games |

**Tasks:**

- [ ] **5.1.1** `app/services/api/tracking_service.go`:
  - `MarkAsMastered(userId, contentItemId, gameId, gameLevelId)` — upsert user_master, remove from user_unknown
  - `ListMastered(userId, cursor, limit)` — paginated mastered items with content details
  - `GetMasterStats(userId)` — count mastered items
  - `DeleteMastered(userId, id)` — remove single entry
  - `BulkDeleteMastered(userId, ids)` — remove multiple entries
  - `MarkAsUnknown(userId, contentItemId, gameId, gameLevelId)` — upsert user_unknown
  - `ListUnknown(userId, cursor, limit)` — paginated unknown items
  - `GetUnknownStats(userId)` — count unknown items
  - `DeleteUnknown(userId, id)` — remove single entry
  - `BulkDeleteUnknown(userId, ids)` — remove multiple entries
  - `MarkAsReview(userId, contentItemId, gameId, gameLevelId)` — upsert user_review with spaced repetition (`GetNextReviewAt`)
  - `ListReviews(userId, cursor, limit)` — paginated review items
  - `GetReviewStats(userId)` — count due/total reviews
  - `DeleteReview(userId, id)` — remove single entry
  - `BulkDeleteReviews(userId, ids)` — remove multiple entries

- [ ] **5.1.2** `app/services/api/favorite_service.go`:
  - `ToggleFavorite(userId, gameId)` — insert or delete
  - `ListFavorites(userId)` — user's favorite games with details

- [ ] **5.1.3** Controllers, requests, routes

- [ ] **5.1.4** Tests

### 5.2 Frontend (dx-web)

- [ ] **5.2.1** Create tracking/favorite API functions
- [ ] **5.2.2** Update play page tracking buttons → API calls
- [ ] **5.2.3** Update favorites page → API calls
- [ ] **5.2.4** Remove migrated server actions and services
- [ ] **5.2.5** Remove related model files

### 5.3 Verification

- [ ] **5.3.1** Mark word as mastered during gameplay
- [ ] **5.3.2** Mark word as unknown
- [ ] **5.3.3** Mark word for review, verify next review date
- [ ] **5.3.4** Toggle favorite on/off
- [ ] **5.3.5** Favorites list shows correct games

---

## Phase 6: Community & Social

> **Goal:** Leaderboard, hall dashboard, referrals, notices, feedback, reports via Go API.

### 6.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Rate Limit | Description |
|--------|------|------|------------|-------------|
| GET | `/api/leaderboard` | Yes | - | Leaderboard by type (exp/playtime) & period (all/weekly/monthly) |
| GET | `/api/hall/dashboard` | Yes | - | Dashboard stats |
| GET | `/api/hall/heatmap` | Yes | - | Activity heatmap for year |
| GET | `/api/invite` | Yes | - | Invite page data (invite URL, stats, first page of referrals) |
| GET | `/api/referrals` | Yes | - | Referral records (paginated) |
| GET | `/api/notices` | Yes | - | System notices (cursor paginated, 20/page) |
| POST | `/api/notices/mark-read` | Yes | - | Mark notices as read (update user.last_read_notice_at) |
| POST | `/adm/notices` | Admin | - | Create notice |
| PUT | `/adm/notices/:id` | Admin | - | Update notice |
| DELETE | `/adm/notices/:id` | Admin | - | Soft delete notice |
| POST | `/api/feedback` | Yes | - | Submit feedback |
| POST | `/api/reports` | Yes | 10/60s | Submit content report |
| GET | `/api/redeems` | Yes | - | User's redemptions (paginated) |
| POST | `/api/redeems` | Yes | - | Redeem a code (transactional: mark used, update VIP, calc due date, grant beans) |
| POST | `/adm/redeems/generate` | Admin | - | Generate batch redeem codes |
| GET | `/adm/redeems` | Admin | - | List all redeems (admin, paginated) |
| GET | `/api/content-seek` | Yes | - | Content seek records |
| POST | `/api/content-seek` | Yes | - | Submit content seek (upsert by course_name, increment count) |

**Tasks:**

- [ ] **6.1.1** `app/services/api/leaderboard_service.go`:
  - `GetLeaderboard(type, period, userId)` — ranking query with user's position
  - 4 query variants: all-time exp, all-time playtime, windowed exp, windowed playtime
  - Uses raw SQL with CTEs and `RANK() OVER` window functions
  - Filters: `user.is_active = true`

- [ ] **6.1.2** `app/services/api/hall_service.go`:
  - `GetDashboard(userId)` — aggregate: profile, master stats, review stats, recent sessions, today's activity
  - `GetHeatmap(userId, year)` — raw SQL: daily answer counts grouped by date for the year

- [ ] **6.1.3** `app/services/api/referral_service.go`:
  - `GetInviteData(userId)` — combined endpoint: invite URL (from user.invite_code), referral stats, first page of referrals
  - `GetReferrals(userId, cursor, limit)` — paginated referral records

- [ ] **6.1.4** `app/services/api/notice_service.go`:
  - `GetNotices(cursor, limit)` — cursor pagination, 20 per page, active only
  - `MarkNoticesRead(userId)` — update user.last_read_notice_at to now

- [ ] **6.1.4b** `app/services/adm/notice_service.go`:
  - `CreateNotice(title, content, icon)` — via `/adm/notices`
  - `UpdateNotice(id, title, content, icon)` — via `/adm/notices/:id`
  - `DeleteNotice(id)` — soft delete, via `/adm/notices/:id`

- [ ] **6.1.5** `app/services/api/feedback_service.go`:
  - `SubmitFeedback(userId, type, description)` — create feedback record
  - `SubmitReport(userId, gameId, gameLevelId, contentItemId, reason, note)` — rate limited

- [ ] **6.1.6** `app/services/api/redeem_service.go`:
  - `GetRedeems(userId, cursor, limit)` — user's own redemptions
  - `RedeemCode(userId, code)` — transactional: verify code unused, mark redeemed, update user grade/VIP due date, grant beans (amount depends on grade)

- [ ] **6.1.6b** `app/services/adm/redeem_service.go`:
  - `GenerateCodes(grade, count)` — via `/adm/redeems/generate`
  - `GetAllRedeems(cursor, limit)` — via `/adm/redeems`

- [ ] **6.1.7** `app/services/api/content_seek_service.go`:
  - `GetContentSeeks(userId)` — user's content seek records
  - `SubmitContentSeek(userId, courseName, description, diskUrl)` — upsert by course_name, increment count

- [ ] **6.1.8** Controllers, requests, routes

- [ ] **6.1.9** Tests

### 6.2 Frontend (dx-web)

- [ ] **6.2.1** Create community API functions
- [ ] **6.2.2** Update hall, leaderboard, referral, notice pages → API calls
- [ ] **6.2.3** Remove migrated server actions and services
- [ ] **6.2.4** Remove related model files

### 6.3 Verification

- [ ] **6.3.1** Dashboard loads with correct stats
- [ ] **6.3.2** Heatmap shows daily activity
- [ ] **6.3.3** Leaderboard displays rankings
- [ ] **6.3.4** Notices paginate correctly
- [ ] **6.3.5** Feedback submission works
- [ ] **6.3.6** Content report works with rate limit

---

## Phase 7: File Uploads & Media

> **Goal:** Image uploads via Go API. Static file serving from Go. This phase comes before AI Custom and Course Game Management because both depend on uploaded images (game covers, content images).

### 7.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/uploads/images` | Yes | Upload image (2MB, JPEG/PNG) |
| GET | `/api/uploads/images/:id` | No | Serve uploaded image |

**Tasks:**

- [ ] **7.1.1** `app/services/api/upload_service.go`:
  - `UploadImage(userId, file, role)` — validate size/MIME, generate path (ULID), save to disk, create DB record
  - `GetImageURL(imageId)` — return serve URL

- [ ] **7.1.2** `app/http/requests/api/upload_request.go`:
  - File size <= 2MB
  - MIME: image/jpeg, image/png
  - Role: must be valid image role

- [ ] **7.1.3** `app/http/controllers/api/upload_controller.go`

- [ ] **7.1.4** Static file serving route for uploaded images

- [ ] **7.1.5** Routes

- [ ] **7.1.6** Tests

### 7.2 Frontend (dx-web)

- [ ] **7.2.1** Update Uppy upload target URL → Go API
- [ ] **7.2.2** Update image URLs to point to Go API
- [ ] **7.2.3** Remove `src/app/api/uploads/` route
- [ ] **7.2.4** Remove `src/features/com/images/services/upload-image.service.ts`

### 7.3 Verification

- [ ] **7.3.1** Image upload works via Uppy
- [ ] **7.3.2** Uploaded images serve correctly
- [ ] **7.3.3** Avatar update uses uploaded image

---

## Phase 8: Course Game Management (User-Created Games)

> **Goal:** Full CRUD for user-created games, levels, content metadata, and content items. This is a complete domain with ownership checks, status-based guards, and content limits.

### 8.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/course-games` | Yes | List user's own games (paginated, filterable by status) |
| POST | `/api/course-games` | Yes | Create a new game |
| PUT | `/api/course-games/:id` | Yes | Update game properties |
| DELETE | `/api/course-games/:id` | Yes | Delete game (owner only) |
| POST | `/api/course-games/:id/publish` | Yes | Publish game (validates readiness) |
| POST | `/api/course-games/:id/withdraw` | Yes | Withdraw published game |
| POST | `/api/course-games/:id/levels` | Yes | Add a level to a game |
| DELETE | `/api/course-games/:id/levels/:levelId` | Yes | Remove a level |
| POST | `/api/course-games/:id/levels/:levelId/metadata` | Yes | Save content metadata (batch, with limit validation) |
| PUT | `/api/course-games/:id/levels/:levelId/metadata/reorder` | Yes | Reorder content metadata |
| GET | `/api/course-games/:id/levels/:levelId/content-items` | Yes | Get content items by metadata |
| POST | `/api/course-games/:id/levels/:levelId/content-items` | Yes | Insert content item at position |
| PUT | `/api/course-games/:id/content-items/:itemId` | Yes | Update content item text/translation |
| PUT | `/api/course-games/:id/content-items/reorder` | Yes | Reorder content items |
| DELETE | `/api/course-games/:id/content-items/:itemId` | Yes | Delete single content item |
| DELETE | `/api/course-games/:id/levels/:levelId/content-items` | Yes | Delete all content from a level |

**Tasks:**

- [ ] **8.1.1** `app/services/api/course_game_service.go`:
  - `ListUserGames(userId, status, cursor, limit)` — user's own games, paginated
  - `CreateGame(userId, name, description, mode, categoryId, pressId, coverId)` — create with draft status
  - `UpdateGame(userId, gameId, ...)` — owner check, reject if published
  - `DeleteGame(userId, gameId)` — owner check, cascade delete levels/content
  - `PublishGame(userId, gameId)` — validate: has levels, levels have content, set status=published
  - `WithdrawGame(userId, gameId)` — set status=withdraw
  - `CreateLevel(userId, gameId, name, description)` — add level with auto-incremented order
  - `DeleteLevel(userId, gameId, gameLevelId)` — owner check, cascade delete content

- [ ] **8.1.2** `app/services/api/course_content_service.go`:
  - `SaveMetadataBatch(userId, gameId, gameLevelId, metadatas)` — bulk create content_meta with limit validation (MAX_SENTENCES, MAX_VOCAB, MAX_ITEMS_PER_META)
  - `ReorderMetadata(userId, gameId, gameLevelId, orderedIds)` — reorder content_meta
  - `GetContentItemsByMeta(userId, gameId, gameLevelId)` — content items grouped by metadata
  - `InsertContentItem(userId, gameId, gameLevelId, contentMetaId, position, data)` — insert at position, shift others
  - `UpdateContentItemText(userId, gameId, itemId, content, translation)` — update text fields
  - `ReorderContentItems(userId, gameId, orderedIds)` — reorder items
  - `DeleteContentItem(userId, gameId, itemId)` — single delete
  - `DeleteAllLevelContent(userId, gameId, gameLevelId)` — delete all content from level

- [ ] **8.1.3** `app/http/requests/api/course_game_request.go`:
  - Validation for each endpoint
  - Ownership verification helper (game belongs to user)
  - Status guard helper (reject edits to published games)

- [ ] **8.1.4** `app/http/controllers/api/course_game_controller.go`

- [ ] **8.1.5** Routes — all protected

- [ ] **8.1.6** Tests

### 8.2 Frontend (dx-web)

- [ ] **8.2.1** Create course game API functions
- [ ] **8.2.2** Update AI custom game creation pages → API calls
- [ ] **8.2.3** Remove `src/features/web/ai-custom/actions/course-game.action.ts`
- [ ] **8.2.4** Remove `src/features/web/ai-custom/services/course-game.service.ts`

### 8.3 Verification

- [ ] **8.3.1** Create a game with levels
- [ ] **8.3.2** Add content metadata and items to levels
- [ ] **8.3.3** Publish game (validates readiness)
- [ ] **8.3.4** Cannot edit published game
- [ ] **8.3.5** Withdraw and re-edit works
- [ ] **8.3.6** Delete game cascades correctly
- [ ] **8.3.7** Content limits enforced (MAX_SENTENCES, MAX_VOCAB)

---

## Phase 9: AI Custom Content (DeepSeek + SSE)

> **Goal:** AI content generation served from Go API with SSE streaming. Depends on Phase 8 (course game management) for saving AI-generated content to games.

### 9.1 Backend (dx-api)

**Endpoints:**

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/ai-custom/generate-metadata` | Yes | Generate story from keywords (5 beans) |
| POST | `/api/ai-custom/format-metadata` | Yes | Format text into learning units |
| POST | `/api/ai-custom/break-metadata` | Yes | Break sentences into units (SSE stream) |
| POST | `/api/ai-custom/generate-content-items` | Yes | Generate word phonetics (SSE stream) |

**Tasks:**

- [ ] **9.1.1** `app/helpers/deepseek.go`:
  - DeepSeek API client (HTTP)
  - Chat completion request/response types
  - Streaming support for SSE endpoints

- [ ] **9.1.2** `app/services/api/ai_custom_service.go`:
  - `GenerateMetadata(userId, keywords)` — consume 5 beans, call DeepSeek, moderation check, refund on error
  - `FormatMetadata(userId, rawText)` — structure into sentence/vocab units, moderation, bean consumption
  - `BreakMetadata(userId, contentMetas, writer)` — SSE: break sentences into word/block/phrase/sentence units, concurrent (limit 20, use semaphore pattern)
  - `GenerateContentItems(userId, contentMetas, writer)` — SSE: generate IPA phonetics, POS, translations, concurrent (limit 50, use semaphore pattern)

- [ ] **9.1.3** Moderation logic — detect inappropriate content, return warning

- [ ] **9.1.4** `app/http/controllers/api/ai_custom_controller.go`

- [ ] **9.1.5** Routes — protected

- [ ] **9.1.6** Tests (mock DeepSeek API)

### 9.2 Frontend (dx-web)

- [ ] **9.2.1** Create AI API functions with SSE consumption (EventSource or fetch + ReadableStream)
- [ ] **9.2.2** Update AI custom pages → API calls
- [ ] **9.2.3** Remove `src/app/api/ai-custom/` routes (4 files)
- [ ] **9.2.4** Remove `src/features/web/ai-custom/services/` (4 AI service files, keep course-game if not already removed)
- [ ] **9.2.5** Remove `src/lib/sse-stream.ts`

### 9.3 Verification

- [ ] **9.3.1** Generate metadata from keywords (beans deducted)
- [ ] **9.3.2** Format raw text into structured content
- [ ] **9.3.3** SSE streaming works for break-metadata
- [ ] **9.3.4** SSE streaming works for generate-content-items
- [ ] **9.3.5** Moderation blocks inappropriate content
- [ ] **9.3.6** Bean refund on AI error
- [ ] **9.3.7** Concurrent SSE streams handle partial failures gracefully

---

## Phase 10: Background Jobs & Cron Scripts

> **Goal:** Email queue and cron jobs running in Go.

### 10.1 Backend (dx-api)

**Tasks:**

- [ ] **10.1.1** `app/jobs/send_email_job.go`:
  - Goravel queue job for sending emails
  - Receives: to, subject, html
  - Uses Goravel mail facade
  - Handles failures gracefully

- [ ] **10.1.2** `app/console/commands/update_play_streaks.go`:
  - Artisan command: `app:update-play-streaks`
  - Logic: users with `last_played_at = yesterday` → increment `current_play_streak`, update `max_play_streak`; users with `last_played_at < yesterday` → reset `current_play_streak` to **1** (not 0); users with `last_played_at = today` → skip
  - Schedule: daily at 2:00 AM

- [ ] **10.1.3** `app/console/commands/reset_energy_beans.go`:
  - Artisan command: `app:reset-energy-beans`
  - Logic: find users on membership anniversary → debit unused beans → credit new grant
  - Handle day-of-month edge cases (29/30/31 in shorter months)
  - VIP: 10,000 beans, Lifetime: 15,000 beans
  - Bean slugs: `monthly-reset-debit` (debit), `monthly-reset-credit` (credit)
  - Schedule: daily at 1:00 AM
  - DB transactions for atomic debit/credit

- [ ] **10.1.4** Register commands in Goravel kernel

- [ ] **10.1.5** Register schedules in `app/console/kernel.go`

- [ ] **10.1.6** Tests

### 10.2 Frontend (dx-web)

- [ ] **10.2.1** Remove `src/workers/email.worker.ts`
- [ ] **10.2.2** Remove `scripts/cron/update-play-streaks.ts`
- [ ] **10.2.3** Remove `scripts/cron/reset-energy-beans.ts`
- [ ] **10.2.4** Remove `scripts/lib/db.ts`
- [ ] **10.2.5** Remove BullMQ and related dependencies from `package.json`

### 10.3 Verification

- [ ] **10.3.1** Email sending works via queue
- [ ] **10.3.2** Play streak cron updates correctly (reset to 1, not 0)
- [ ] **10.3.3** Bean reset cron runs correctly (test with mock data)

---

## Phase 11: Cleanup & Finalization

> **Goal:** Remove all remaining server-only code from dx-web, clean up dependencies, finalize deployment.

### 11.1 dx-web Cleanup

- [ ] **11.1.1** Audit remaining server-only files — remove any leftover services, models, server actions
- [ ] **11.1.2** Remove `src/models/` directory entirely (all DB ops now in Go)
- [ ] **11.1.3** Remove `src/lib/db.ts` (Prisma client)
- [ ] **11.1.4** Remove `src/lib/redis.ts` (Redis client)
- [ ] **11.1.5** Remove `src/lib/rate-limit.ts`
- [ ] **11.1.6** Remove `src/lib/email.ts` (nodemailer transporter)
- [ ] **11.1.7** Remove Prisma config:
  - Delete `prisma/` directory (schema, migrations, seeds)
  - Delete `src/generated/prisma/`
  - Remove `@prisma/client`, `@prisma/adapter-pg`, `prisma` from `package.json`
- [ ] **11.1.8** Remove server-only dependencies from `package.json`:
  - `bcrypt` / `bcryptjs`
  - `nodemailer`
  - `bullmq`
  - `ioredis`
  - `pg` (PostgreSQL driver)
  - Any other server-only packages
- [ ] **11.1.9** Verify `npm run build` succeeds with reduced dependencies
- [ ] **11.1.10** Verify all pages work with API-only data flow
- [ ] **11.1.11** Keep client-side utilities that remain in dx-web:
  - `src/lib/avatar.ts` (deterministic avatar colors — client-side only)
  - `src/lib/format.ts` (date/time formatters — client-side only)
  - `src/lib/utils.ts` (cn() helper — client-side only)

### 11.2 dx-api Finalization

- [ ] **11.2.1** Write `dx-api/CLAUDE.md` — project guidelines, conventions, directory structure
- [ ] **11.2.2** Transfer schema ownership: create Goravel migrations from existing DB state
  - Use `goravel artisan make:migration` for each table
  - Or generate from `pg_dump --schema-only`
  - Prisma is no longer the schema source
- [ ] **11.2.3** Update `dx-api/Dockerfile` for production:
  - Multi-stage build
  - Include storage directory
  - Health check
- [ ] **11.2.4** Update `docker-compose.yml`:
  - dx-api service (port 3001)
  - dx-web service (port 3000)
  - PostgreSQL service
  - Redis service
  - Shared network
- [ ] **11.2.5** Update `.env.example` for both projects
- [ ] **11.2.6** Write API documentation (optional: Swagger/OpenAPI)

### 11.3 Final Verification

- [ ] **11.3.1** Full signup → signin → play game → view stats flow
- [ ] **11.3.2** AI content generation flow (SSE streaming)
- [ ] **11.3.3** Course game creation → publish → play flow
- [ ] **11.3.4** Image upload and display
- [ ] **11.3.5** All community features (leaderboard, notices, referrals)
- [ ] **11.3.6** Cron jobs execute correctly
- [ ] **11.3.7** `npm run build` on dx-web — no server-side code remaining
- [ ] **11.3.8** `go build` on dx-api — all features compile
- [ ] **11.3.9** Docker compose up — both services run and communicate

---

## Quick Reference: What Goes Where

| dx-web (Next.js) — Frontend Only | dx-api `/api/*` — Client APIs | dx-api `/adm/*` — Admin APIs |
|---|---|---|
| React components, pages, layouts | Controllers, services, models | Admin controllers, services |
| Client-side hooks, state (Zustand) | Business logic, validation | Admin CRUD operations |
| API client (`src/lib/api-client.ts`) | User JWT auth | Admin JWT auth + RBAC |
| Tailwind styles, shadcn/ui | Database queries (Goravel ORM) | Operation audit logging |
| Client-side form validation (Zod) | Server-side validation (requests) | Permission-based access control |
| SSE consumption (EventSource) | SSE production (streaming) | — |
| Static assets (public/) | File uploads, media serving | Content management |
| — | Redis, rate limiting, queue | — |
| — | Cron jobs, email sending | — |
| — | DeepSeek AI integration | — |

## Notes

- **Always verify dx-web still works** after each phase before moving to the next
- **Run `go test -race ./...`** after each backend change
- **Check existing Prisma table/column names** before writing Goravel models (`\d table_name` in psql)
- **Keep dx-web's `"use server"` files** until the corresponding Go API is verified and the frontend is updated
- **Commit after each sub-task** with descriptive messages following conventional commits
- During migration, both dx-web and dx-api connect to the **same PostgreSQL database**
- Constants with Chinese labels (e.g., bean reasons) should use the same Chinese strings in Go
- **Admin API structure** (`/adm/*` routes) is scaffolded in Phase 0-1: models, middleware (JWT + RBAC + operation log), admin auth endpoints. Admin CRUD endpoints for content management, user management, etc. will be added incrementally after the client API migration is complete.

## Risk Areas

| Risk | Phase | Mitigation |
|------|-------|------------|
| **Session service transactions** | P4 | Goravel ORM transactions must handle the same atomic patterns as Prisma's `$transaction`. Verify transaction isolation level and deadlock handling. Test heavily. |
| **Leaderboard raw SQL** | P6 | Uses CTEs and `RANK() OVER` window functions — 4 distinct query variants. Must replicate as raw SQL in Goravel, not ORM. |
| **SSE concurrent streams** | P9 | Go goroutines for concurrency. Use semaphore pattern (buffered channel) to limit concurrent DeepSeek calls. Handle partial failures without corrupting the stream. |
| **NextAuth removal** | P1 | Highest-risk frontend change. All auth state shifts from server sessions to client-side JWT. Test every protected page. Update `src/proxy.ts` middleware. |
| **Bcrypt compatibility** | P1 | Go's `bcrypt` must verify hashes created by Node.js `bcryptjs`. Both use standard bcrypt — compatible, but verify with real password hashes before deploying. |
