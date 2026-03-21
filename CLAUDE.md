# CLAUDE.md

## Project Overview

Douxue — English learning through games. Monorepo with two projects:

- **dx-web** (Next.js 16) — Pure frontend client, no server-side DB access
- **dx-api** (Go/Goravel) — Sole API backend, all business logic and DB operations

Both share one PostgreSQL database and Redis instance.

## Commands

### dx-web (frontend)

```bash
cd dx-web
npm run dev       # Development server with Turbopack (http://localhost:3000)
npm run build     # Production build
npm run start     # Run production server
npm run lint      # ESLint
```

### dx-api (backend)

```bash
cd dx-api
go run .          # Run server (http://localhost:3001)
air               # Hot reload development
go test -race ./...   # Run all tests with race detector
go build ./...    # Verify compilation
```

## Architecture

```
dx-web (Next.js)  ──── HTTP/JWT ────►  dx-api (Goravel)
  port 3000                              port 3001
  pure frontend                          /api/*  client APIs (user JWT)
  apiServerFetch (SSR)                   /adm/*  admin APIs (admin JWT)
  apiClient (CSR)                        /api/admin/*  user-facing admin (user JWT + admin guard)
                                         PostgreSQL + Redis
```

**Response envelope:** `{ "code": 0, "message": "ok", "data": {...} }`
**Pagination:** Cursor-based (public) or offset-based (admin) inside `data`
**Auth:** JWT in `Authorization: Bearer {token}` header; also stored in `dx_token` httpOnly cookie for SSR

## dx-web Directory Structure

```
dx-web/src/
├── app/
│   ├── (web)/                  # Public frontend routes
│   │   ├── auth/               # Sign in / sign up
│   │   └── hall/               # Games hall (protected)
│   └── adm/                    # Admin backend
├── components/
│   ├── ui/                     # shadcn/ui (do NOT modify)
│   └── in/                     # Custom shared components
├── features/
│   ├── web/                    # Web portal features
│   │   └── {feature}/
│   │       ├── components/     # Feature UI
│   │       ├── hooks/          # React hooks
│   │       ├── actions/        # Server actions (use apiServerFetch)
│   │       ├── schemas/        # Zod validation
│   │       ├── helpers/        # Utility functions
│   │       └── types/          # TypeScript types
│   ├── adm/                    # Admin features
│   └── com/                    # Shared features
├── consts/                     # Domain constants
├── hooks/                      # Shared React hooks
└── lib/
    ├── api-client.ts           # Client-side API fetch (Bearer token from localStorage)
    ├── api-server.ts           # Server-side API fetch (JWT from dx_token cookie)
    ├── auth.ts                 # JWT auth — reads dx_token cookie, returns { user: { id, name } }
    ├── avatar.ts               # Deterministic avatar colors
    ├── format.ts               # Date/time formatters
    └── utils.ts                # cn() Tailwind helper
```

### dx-web Conventions

- **Thin route pages** — `page.tsx` files delegate to features, no inline business logic
- **Server actions** call Go API via `apiServerFetch()`, never access DB directly
- **Client code** calls Go API via `apiClient` from `api-client.ts`
- **Types** defined locally in action files and exported for component use
- **Imports** use `@/` alias, direct paths (no barrel exports)
- **Styling** — TailwindCSS v4, oklch theme, shadcn/ui New York style
- **Icons** — Lucide React
- **Validation** — Zod schemas

## dx-api Directory Structure

```
dx-api/
├── app/
│   ├── http/
│   │   ├── controllers/
│   │   │   ├── api/            # Client controllers (user JWT)
│   │   │   └── adm/            # Admin controllers (admin JWT + RBAC)
│   │   ├── middleware/         # JWT auth, RBAC, rate limit, admin guard
│   │   └── requests/
│   │       ├── api/            # Client request validation
│   │       └── adm/            # Admin request validation
│   ├── services/
│   │   ├── api/                # Client business logic
│   │   └── adm/                # Admin business logic
│   ├── models/                 # GORM models (48 files, map to existing DB tables)
│   ├── helpers/                # Response envelope, Redis, DeepSeek client, SSE, etc.
│   ├── constants/              # Error codes, game modes, bean slugs, etc.
│   ├── facades/                # Goravel facades (ORM, Auth, Redis, Mail, Queue)
│   ├── jobs/                   # Background queue jobs (SendEmailJob)
│   └── console/
│       └── commands/           # Scheduled CLI commands
├── routes/
│   ├── api.go                  # /api/* client routes
│   └── adm.go                  # /adm/* admin routes
├── config/                     # App, database, JWT, CORS, cache, mail config
├── bootstrap/app.go            # App initialization (commands, schedules, jobs)
└── main.go                     # Entry point
```

### dx-api Conventions

- **Controllers** are thin — validate input, call service, return response
- **Services** contain all business logic and DB queries
- **Models** are plain GORM structs, no methods — all logic in services
- **Helpers** for cross-cutting concerns (response formatting, rate limiting, ID generation)
- **Error sentinels** in `services/api/errors.go` — use `errors.Is()` for matching
- **ID generation** — ULID via `ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)`
- **Transactions** — `facades.Orm().Transaction(func(tx orm.Query) error { ... })`
- **Raw SQL** for complex queries, GORM query builder for simple ones
- **Admin guard** — `middleware.AdminGuard()` checks username == "rainson"

### API Route Groups

| Prefix | Auth | Purpose |
|--------|------|---------|
| `/api/*` (public) | None | Game listing, auth endpoints, image serving |
| `/api/*` (protected) | User JWT | Profile, sessions, tracking, favorites, AI, etc. |
| `/api/admin/*` | User JWT + AdminGuard | Notice CRUD, redeem generation (user-facing admin) |
| `/adm/*` | Admin JWT + RBAC | Full admin panel APIs |

### Scheduled Tasks

- `app:update-play-streaks` — daily 02:00, continues/resets play streaks
- `app:reset-energy-beans` — daily 01:00, monthly bean grants for paid members

## Environment Variables

### dx-web (.env)

```
NEXT_PUBLIC_API_URL=http://localhost:3001    # Go API endpoint
```

### dx-api (.env)

```
APP_PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=douxue
DB_USERNAME=postgres
DB_PASSWORD=
JWT_SECRET=
CORS_ALLOWED_ORIGINS=http://localhost:3000
REDIS_HOST=localhost
REDIS_PORT=6379
MAIL_HOST=    MAIL_PORT=    MAIL_USERNAME=    MAIL_PASSWORD=
DEEPSEEK_API_KEY=
```

## Code Guidelines

- **No console.log** in production code
- **Security** — validate input, check ownership, use error sentinels
- **Never modify** `dx-web/src/components/ui/` (shadcn/ui managed)
- **Commit style** — `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- **Always ask** before git commit
