# CLAUDE.md

## Project Overview

Douxue — English learning through games. Monorepo with three projects:

- **dx-web** (Next.js 16) — Pure frontend client, no server-side DB access
- **dx-api** (Go/Goravel) — Sole API backend, all business logic and DB operations
- **dx-mini** (WeChat Mini Program, native + TypeScript) — Mobile client, consumes dx-api over HTTP/JWT + WebSocket

dx-web and dx-api share one PostgreSQL database and Redis instance and deploy
together via `deploy/docker-compose.*.yml`. dx-mini talks to the same dx-api
but **does NOT deploy with the container stack** — it ships via WeChat
Developer Tools (manual upload) or `miniprogram-ci` (automated) to Tencent's
WeChat servers, and is only pinned to dx-api by URL.

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

### dx-mini (WeChat Mini Program)

```bash
cd dx-mini
# npm install         # Only needed for dev tools (lucide-static) at repo root
# npm run build:icons  # Regenerate Lucide SVG map into miniprogram/components/dx-icon/icons.ts (one-shot)
# Actual development happens in WeChat Developer Tools:
#   - Open project root: /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini
#   - miniprogram/ is the source tree (miniprogramRoot in project.config.json)
#   - 详情 → 本地设置 → 勾选「不校验合法域名...」(dev only)
#   - In DevTools console: require('./utils/config').setDevApiBaseUrl('http://<lan-ip>')
#     to point at a LAN dx-api instance; clearDevApiBaseUrl() to reset.
```

## Architecture

```
dx-web (Next.js) ──┐
  port 3000        │
                   ├─── HTTP/JWT + WebSocket ───►  dx-api (Goravel)
dx-mini (WeChat)   │                                 port 3001
  WeChat DevTools  │                                 /api/*   client APIs (user JWT)
                   │                                 /adm/*   admin APIs (admin JWT)
                   │                                 /api/admin/*  user-facing admin (user JWT + admin guard)
                   │                                 /api/ws  WebSocket (auth via cookie for dx-web,
                   │                                          first-frame {op:"auth",token} for dx-mini)
                   │                                 PostgreSQL + Redis
                   ▼
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

## dx-mini Directory Structure

```
dx-mini/
├── miniprogram/                # miniprogramRoot (the source tree WeChat DevTools sees)
│   ├── app.ts / app.json / app.wxss   # App shell
│   ├── pages/                  # Native pages: home, games, learn, leaderboard, me, login, etc.
│   ├── components/dx-icon/     # Lucide SVG renderer; icons.ts is auto-generated
│   ├── custom-tab-bar/         # Color-only active state, outline icons
│   ├── utils/                  # api.ts, auth.ts, config.ts, format.ts, ws.ts
│   └── typings/                # WeChat API type shims
├── scripts/build-icons.mjs     # Regenerates Lucide SVG map into components/dx-icon/icons.ts (one-shot)
├── docs/superpowers/           # Design specs + implementation plans
├── package.json                # Dev-only deps (lucide-static)
├── project.config.json         # WeChat Developer Tools config
└── tsconfig.json               # Strict TS across miniprogram/
```

### dx-mini Conventions

- **UI lib** — Vant Weapp 1.11.x, pinned. Glass-easel + Skyline rendering enabled.
- **Icons** — Lucide SVG via the `<dx-icon>` component. Add glyphs to `scripts/build-icons.mjs` ICONS array and re-run `npm run build:icons`. Color via the `color` prop, default stroke width 1.25.
- **API** — every call goes through `utils/api.ts` → dx-api. Base URL resolved at read time from `wx.getStorageSync('dx_dev_api_base_url')` in dev, hardcoded prod domain in release/trial.
- **WebSocket** — auth via first-frame `{op:"auth",token}` envelope (not URL, not cookie). `ws.ts` buffers subscribes until `auth_success` lands.
- **Request shape** — mirrors dx-web/dx-api: `{code, message, data}` envelope; cursor pagination via `{items, nextCursor, hasMore}`.
- **Typing** — strict mode enabled but `miniprogram-api-typings` v2.8.3 can't type `this` inside `Component({methods})` — existing codebase already ignores these specific errors; don't introduce NEW tsc errors beyond that pattern.

### Deployment

dx-mini does NOT ship with `deploy/docker-compose.*.yml`. Two paths:

1. **Manual** — WeChat Developer Tools "上传" → 体验版 → promote to 正式版 in the 微信公众平台 admin console.
2. **CLI (future)** — `miniprogram-ci` npm package automates upload from `dx-mini/scripts/deploy.mjs`. Not wired yet.

Prod requires: filed public domain (ICP 备案), SSL cert, `wss://` for WebSocket, and the domain added to `socket合法域名` in the WeChat 公众平台 (5 edits/month cap).

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

### dx-mini (no .env file — URL is runtime-configurable)

- Dev API URL — stored in WeChat storage under key `dx_dev_api_base_url`; set once per DevTools install via the console helper `require('./utils/config').setDevApiBaseUrl('http://<lan>')`. Falls back to `http://localhost`.
- Release/trial API URL — hardcoded at `https://api.douxue.com` in `miniprogram/utils/config.ts` (change here, not via env).
- WeChat AppID — in `project.private.config.json` (git-ignored by dx-mini's own `.gitignore`).

## Code Style

- **Neat, short, and clean** — no bloat, no unnecessary verbosity
- **Many small files** over few large files
- **Composite logic from small snippets** — break large files into focused pieces
- **Always encapsulate** logic in functions or helpers when possible

## Code Guidelines

- **No console.log** in production code
- **Security** — validate input, check ownership, use error sentinels
- **Never modify** `dx-web/src/components/ui/` (shadcn/ui managed)
- **Never modify** `dx-mini/miniprogram/miniprogram_npm/` (generated by WeChat DevTools `构建 npm`)
- **Commit style** — `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`; tag mini-only changes with `(mini)`, api-only with `(api)`, web-only with `(web)`
