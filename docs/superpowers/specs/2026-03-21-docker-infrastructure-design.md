# Docker Infrastructure Design

## Overview

Containerize the Douxue monorepo (dx-api + dx-web) with Docker Compose for both development and production environments. Nginx reverse-proxies all traffic. A single `deploy/` directory centralizes compose files, env files, and nginx configs.

## Architecture

```
                    ┌─────────────────────────────┐
                    │         nginx (:80)          │
                    │    reverse proxy (HTTP)      │
                    └──────┬──────────────┬────────┘
                           │              │
                    /api/* │       /* │
                    /adm/* │              │
                           ▼              ▼
                    ┌────────────┐  ┌────────────┐
                    │  dx-api    │  │  dx-web     │
                    │  (:3001)   │  │  (:3000)    │
                    └──┬─────┬──┘  └─────────────┘
                       │     │
                 ┌─────┘     └─────┐
                 ▼                 ▼
          ┌────────────┐    ┌────────────┐
          │  postgres   │    │   redis    │
          │  (:5432)    │    │  (:6379)   │
          └────────────┘    └────────────┘
```

Nginx routing:
- `/api/*` and `/adm/*` -> dx-api (port 3001)
- `/*` -> dx-web (port 3000)

## File Structure

```
dx-source/
├── dx-api/
│   ├── Dockerfile          # prod: multi-stage Go build (exists, minor update)
│   ├── Dockerfile.dev      # dev: air hot reload
│   └── .dockerignore
├── dx-web/
│   ├── Dockerfile          # prod: multi-stage standalone Next.js
│   ├── Dockerfile.dev      # dev: npm run dev with Turbopack
│   └── .dockerignore
└── deploy/
    ├── docker-compose.dev.yml
    ├── docker-compose.prod.yml
    ├── env/
    │   ├── .env.dev
    │   └── .env.prod
    └── nginx/
        ├── nginx.dev.conf
        └── nginx.prod.conf
```

## Dockerfiles

### dx-api/Dockerfile (prod) — update existing

Current multi-stage build is kept. Changes:
- Ensure `APP_HOST=0.0.0.0` in env so the server listens on all interfaces
- Remove `COPY --from=builder /build/storage/ /www/storage/` — storage dir created at runtime via volume
- Add `RUN mkdir -p /www/storage/app/uploads` in runtime stage

### dx-api/Dockerfile.dev

```dockerfile
FROM golang:1.24-alpine
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 3001
CMD ["air"]
```

Source code bind-mounted at runtime. Air watches `.go` files and rebuilds automatically.

### dx-web/Dockerfile (prod)

```dockerfile
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS builder
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

Requires `output: "standalone"` in `next.config.ts`.

`NEXT_PUBLIC_API_URL` is passed as a build arg because Next.js inlines `NEXT_PUBLIC_*` vars at build time into the client JS bundle. The compose file passes it via `build.args`.

### dx-web/Dockerfile.dev

```dockerfile
FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
EXPOSE 3000
CMD ["npm", "run", "dev"]
```

Source bind-mounted. Anonymous volumes preserve `node_modules` and `.next`.

## API URL Strategy

`NEXT_PUBLIC_API_URL` is baked into client JS at build time and used by the browser. `api-server.ts` (SSR) runs inside the Docker network and needs to reach dx-api directly.

| Context | Var | Dev Value | Prod Value |
|---------|-----|-----------|------------|
| Browser (api-client.ts) | `NEXT_PUBLIC_API_URL` | `http://localhost` | `https://your-domain.com` |
| SSR (api-server.ts) | `API_INTERNAL_URL` | `http://dx-api:3001` | `http://dx-api:3001` |

Code change: update `api-server.ts` to read `API_INTERNAL_URL` (falls back to `NEXT_PUBLIC_API_URL`):

```ts
const API_URL = process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
```

The `next.config.ts` image rewrite (`/api/uploads/images/:id`) becomes redundant behind nginx (nginx already routes `/api/*` to dx-api), but is harmless to keep for non-Docker local dev.

## Docker Compose

### docker-compose.dev.yml

Services:
- **nginx** — `nginx:alpine`, port 80:80, mounts `nginx.dev.conf`
- **dx-api** — builds `Dockerfile.dev`, bind-mounts `../dx-api:/app`, Go module + build cache in named volumes
- **dx-web** — builds `Dockerfile.dev`, bind-mounts `../dx-web:/app`, anonymous volumes for `node_modules` and `.next`
- **postgres** — `postgres:16-alpine`, port 5432 exposed to host, healthcheck via `pg_isready`
- **redis** — `redis:7-alpine`, port 6379 exposed to host, healthcheck via `redis-cli ping`

dx-api depends on postgres (healthy) + redis (healthy). All services use `env_file: ./env/.env.dev`. Restart policy: `unless-stopped`.

### docker-compose.prod.yml

Same services but:
- dx-api and dx-web build from production Dockerfiles (no bind mounts)
- dx-web passes `NEXT_PUBLIC_API_URL` via `build.args`
- Postgres and Redis ports NOT exposed to host
- Nginx uses `nginx.prod.conf` (placeholder domain, future SSL)
- Named volume `api-storage` for uploaded images
- Restart policy: `always`

## Environment Files

### deploy/env/.env.dev

```env
# --- App ---
APP_NAME=Douxue
APP_ENV=local
APP_KEY=dev-key-change-me
APP_DEBUG=true
APP_URL=http://localhost
APP_HOST=0.0.0.0
APP_PORT=3001
LOG_CHANNEL=stack
LOG_LEVEL=debug
JWT_SECRET=dev-jwt-secret-change-me

# --- Database ---
DB_HOST=postgres
DB_PORT=5432
DB_DATABASE=dxdb
DB_USERNAME=postgres
DB_PASSWORD=postgres
POSTGRES_DB=dxdb
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# --- Redis ---
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# --- CORS ---
CORS_ALLOWED_ORIGINS=http://localhost

# --- Storage ---
STORAGE_PATH=storage/app

# --- Mail (optional for dev) ---
MAIL_HOST=
MAIL_PORT=
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM_ADDRESS=
MAIL_FROM_NAME=Douxue

# --- AI (optional for dev) ---
DEEPSEEK_API_KEY=

# --- dx-web ---
NEXT_PUBLIC_API_URL=http://localhost
API_INTERNAL_URL=http://dx-api:3001
```

Key details:
- `DB_HOST=postgres` and `REDIS_HOST=redis` — Docker service names
- `POSTGRES_*` vars consumed by the postgres image, `DB_*` consumed by dx-api
- `APP_HOST=0.0.0.0` — Go server listens on all interfaces
- `NEXT_PUBLIC_API_URL=http://localhost` — browser calls through nginx
- `API_INTERNAL_URL=http://dx-api:3001` — SSR calls dx-api directly via Docker network

### deploy/env/.env.prod

Same structure with:
- `APP_ENV=production`, `APP_DEBUG=false`, `LOG_LEVEL=error`
- Placeholder secrets marked `<CHANGE_ME>`
- `NEXT_PUBLIC_API_URL=https://your-domain.com`
- `API_INTERNAL_URL=http://dx-api:3001`
- `CORS_ALLOWED_ORIGINS=https://your-domain.com`

## Nginx Configuration

### nginx.dev.conf

```nginx
events { worker_connections 1024; }

http {
    upstream api { server dx-api:3001; }
    upstream web { server dx-web:3000; }

    server {
        listen 80;
        client_max_body_size 4m;

        # API & Admin routes -> Go backend
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            # SSE support for AI streaming
            proxy_buffering off;
            proxy_cache off;
            proxy_read_timeout 300s;
            proxy_send_timeout 300s;
        }

        location /adm/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        # Everything else -> Next.js
        location / {
            proxy_pass http://web;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            # WebSocket support for HMR
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }
}
```

### nginx.prod.conf

Same routing plus:
- Gzip compression
- Security headers (`X-Frame-Options`, `X-Content-Type-Options`, `Strict-Transport-Security`)
- `client_max_body_size 4m` (matches Go API limit)
- Placeholder `server_name` for future domain
- Static asset caching for `_next/static`

## .dockerignore Files

### dx-api/.dockerignore

```
.git
.env
.env.*
tmp/
*.md
docker-compose.yml
```

Note: `storage/` and `.air.toml` NOT excluded — storage dir needed for prod Dockerfile structure, `.air.toml` is harmless.

### dx-web/.dockerignore

```
.git
.env
.env.*
node_modules/
.next/
*.md
```

## Scheduled Tasks

Goravel's scheduler runs within the main application process — `app.Start()` starts a goroutine that evaluates the schedule every minute. No separate cron container or host-level cron is needed. The two daily tasks (`app:reset-energy-beans` at 01:00, `app:update-play-streaks` at 02:00) run automatically inside the dx-api container.

## Code Changes Required

1. **dx-web/next.config.ts** — add `output: "standalone"` for production Docker builds
2. **dx-web/src/lib/api-server.ts** — use `API_INTERNAL_URL` env var (falls back to `NEXT_PUBLIC_API_URL`)
3. **dx-api/config/cors.go** — remove hardcoded `http://localhost:3000` from allowed origins (redundant behind nginx, `CORS_ALLOWED_ORIGINS` env var handles it)

## Usage

```bash
# Development
cd deploy
docker compose -f docker-compose.dev.yml up --build

# Production
cd deploy
docker compose -f docker-compose.prod.yml up --build -d
```
