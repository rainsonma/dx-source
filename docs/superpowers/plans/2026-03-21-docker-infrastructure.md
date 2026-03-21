# Docker Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Containerize the Douxue monorepo with Docker Compose for dev and prod, with nginx reverse proxy, so the entire stack runs via a single `docker compose` command.

**Architecture:** Nginx on port 80 routes `/api/*` and `/adm/*` to dx-api (Go/Goravel), everything else to dx-web (Next.js). Postgres and Redis run as companion services. Dev uses bind mounts + hot reload; prod uses optimized multi-stage builds.

**Tech Stack:** Docker, Docker Compose, nginx, Go 1.24, Node 22, PostgreSQL 16, Redis 7

**Spec:** `docs/superpowers/specs/2026-03-21-docker-infrastructure-design.md`

---

### Task 1: Code changes — dx-web

**Files:**
- Modify: `dx-web/next.config.ts`
- Modify: `dx-web/src/lib/api-server.ts`

- [ ] **Step 1: Update `next.config.ts` — add standalone output, remove rewrite**

Replace the entire file content with:

```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
};

export default nextConfig;
```

The `/api/uploads/images/:id` rewrite is removed — nginx handles all `/api/*` routing.

- [ ] **Step 2: Update `api-server.ts` — use `API_INTERNAL_URL` for SSR**

Change line 5 in `dx-web/src/lib/api-server.ts` from:

```ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
```

to:

```ts
const API_URL = process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
```

This lets SSR calls use the Docker-internal URL (`http://dx-api:3001`) while the browser uses the public URL baked in at build time.

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npm run build`
Expected: Build succeeds. Output shows "standalone" mode in build summary.

- [ ] **Step 4: Commit**

```bash
git add dx-web/next.config.ts dx-web/src/lib/api-server.ts
git commit -m "feat: configure standalone output and split API URLs for Docker"
```

---

### Task 2: Code changes — dx-api

**Files:**
- Modify: `dx-api/config/cors.go`

- [ ] **Step 1: Remove hardcoded localhost:3000 from CORS config**

In `dx-api/config/cors.go`, change line 19 from:

```go
"allowed_origins":      []string{"http://localhost:3000", config.Env("CORS_ALLOWED_ORIGINS", "").(string)},
```

to:

```go
"allowed_origins":      []string{config.Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000").(string)},
```

Now CORS origins are fully controlled by the `CORS_ALLOWED_ORIGINS` env var, with a sensible default for bare-metal local dev.

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Compiles without errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/config/cors.go
git commit -m "refactor: make CORS origins fully env-driven"
```

---

### Task 3: dx-api Dockerfiles and .dockerignore

**Files:**
- Modify: `dx-api/Dockerfile`
- Create: `dx-api/Dockerfile.dev`
- Create: `dx-api/.dockerignore`

- [ ] **Step 1: Update `dx-api/Dockerfile` (prod)**

Replace the full contents of `dx-api/Dockerfile` with:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o main .

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /www

COPY --from=builder /build/main /www/
RUN mkdir -p /www/storage/app/uploads

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3001/api/health || exit 1

ENTRYPOINT ["/www/main"]
```

Changes from original: removed `COPY --from=builder /build/storage/`, added `RUN mkdir -p` for storage dir.

- [ ] **Step 2: Create `dx-api/Dockerfile.dev`**

```dockerfile
FROM golang:1.24-alpine

RUN go install github.com/air-verse/air@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

EXPOSE 3001
CMD ["air"]
```

- [ ] **Step 3: Create `dx-api/.dockerignore`**

```
.git
.env
.env.*
tmp/
*.md
docker-compose.yml
```

- [ ] **Step 4: Verify prod Dockerfile builds**

Run: `cd dx-api && docker build -t dx-api-test .`
Expected: Multi-stage build completes successfully.

- [ ] **Step 5: Commit**

```bash
git add dx-api/Dockerfile dx-api/Dockerfile.dev dx-api/.dockerignore
git commit -m "feat: update dx-api Dockerfile and add dev Dockerfile with .dockerignore"
```

---

### Task 4: dx-web Dockerfiles and .dockerignore

**Files:**
- Create: `dx-web/Dockerfile`
- Create: `dx-web/Dockerfile.dev`
- Create: `dx-web/.dockerignore`

- [ ] **Step 1: Create `dx-web/Dockerfile` (prod)**

```dockerfile
# Stage 1: install dependencies
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Stage 2: build
FROM node:22-alpine AS builder
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

# Stage 3: runtime
FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000 || exit 1

CMD ["node", "server.js"]
```

- [ ] **Step 2: Create `dx-web/Dockerfile.dev`**

```dockerfile
FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

EXPOSE 3000
CMD ["npm", "run", "dev"]
```

- [ ] **Step 3: Create `dx-web/.dockerignore`**

```
.git
.env
.env.*
node_modules/
.next/
*.md
```

- [ ] **Step 4: Verify prod Dockerfile builds**

Run: `cd dx-web && docker build --build-arg NEXT_PUBLIC_API_URL=http://localhost -t dx-web-test .`
Expected: Multi-stage build completes. Standalone output created.

- [ ] **Step 5: Commit**

```bash
git add dx-web/Dockerfile dx-web/Dockerfile.dev dx-web/.dockerignore
git commit -m "feat: add Dockerfiles and .dockerignore for dx-web"
```

---

### Task 5: Deploy directory — env files and .gitignore

**Files:**
- Create: `deploy/.gitignore`
- Create: `deploy/env/.env.example`
- Create: `deploy/env/.env.dev`

- [ ] **Step 1: Create `deploy/.gitignore`**

```
env/.env.dev
env/.env.prod
```

- [ ] **Step 2: Create `deploy/env/.env.example`**

```env
# ===========================================
# Douxue Docker Environment Configuration
# Copy to .env.dev or .env.prod and fill in values
# ===========================================

# --- App (dx-api) ---
APP_NAME=Douxue
APP_ENV=local                        # local | production
APP_KEY=                             # random string
APP_DEBUG=true                       # true | false
APP_URL=http://localhost
APP_HOST=0.0.0.0                     # must be 0.0.0.0 for Docker
APP_PORT=3001
LOG_CHANNEL=stack
LOG_LEVEL=debug                      # debug | info | error
JWT_SECRET=                          # required — generate a strong secret

# --- Database (PostgreSQL) ---
DB_HOST=postgres                     # Docker service name
DB_PORT=5432
DB_DATABASE=dxdb
DB_USERNAME=postgres
DB_PASSWORD=                         # required
POSTGRES_DB=dxdb                     # read by postgres image
POSTGRES_USER=postgres               # read by postgres image
POSTGRES_PASSWORD=                   # must match DB_PASSWORD

# --- Redis ---
REDIS_HOST=redis                     # Docker service name
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# --- CORS ---
CORS_ALLOWED_ORIGINS=http://localhost # prod: https://your-domain.com

# --- Storage ---
STORAGE_PATH=storage/app

# --- Mail (SMTP) ---
MAIL_HOST=
MAIL_PORT=
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM_ADDRESS=
MAIL_FROM_NAME=Douxue

# --- AI (DeepSeek) ---
DEEPSEEK_API_KEY=

# --- dx-web ---
NEXT_PUBLIC_API_URL=http://localhost  # browser-facing URL (through nginx)
API_INTERNAL_URL=http://dx-api:3001  # SSR-only, Docker-internal
```

- [ ] **Step 3: Create `deploy/env/.env.dev`**

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

# --- Mail ---
MAIL_HOST=
MAIL_PORT=
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM_ADDRESS=
MAIL_FROM_NAME=Douxue

# --- AI ---
DEEPSEEK_API_KEY=

# --- dx-web ---
NEXT_PUBLIC_API_URL=http://localhost
API_INTERNAL_URL=http://dx-api:3001
```

- [ ] **Step 4: Commit** (only .gitignore and .env.example — .env.dev is gitignored)

```bash
git add deploy/.gitignore deploy/env/.env.example
git commit -m "feat: add deploy env template and gitignore"
```

---

### Task 6: Nginx configuration

**Files:**
- Create: `deploy/nginx/nginx.dev.conf`
- Create: `deploy/nginx/nginx.prod.conf`

- [ ] **Step 1: Create `deploy/nginx/nginx.dev.conf`**

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server dx-api:3001;
    }

    upstream web {
        server dx-web:3000;
    }

    server {
        listen 80;
        client_max_body_size 4m;

        # API & Admin routes -> Go backend
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
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
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Everything else -> Next.js
        location / {
            proxy_pass http://web;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            # WebSocket support for HMR
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }
}
```

- [ ] **Step 2: Create `deploy/nginx/nginx.prod.conf`**

```nginx
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    sendfile    on;
    tcp_nopush  on;
    gzip        on;
    gzip_types  text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1000;

    upstream api {
        server dx-api:3001;
    }

    upstream web {
        server dx-web:3000;
    }

    server {
        listen 80;
        server_name your-domain.com;    # TODO: replace with real domain
        client_max_body_size 4m;

        # Security headers
        add_header X-Frame-Options DENY always;
        add_header X-Content-Type-Options nosniff always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy strict-origin-when-cross-origin always;
        # TODO: add Strict-Transport-Security after SSL is configured

        # API & Admin routes -> Go backend
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
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
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Next.js static assets — long cache
        location /_next/static/ {
            proxy_pass http://web;
            proxy_set_header Host $host;
            expires 365d;
            add_header Cache-Control "public, immutable";
        }

        # Everything else -> Next.js
        location / {
            proxy_pass http://web;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add deploy/nginx/
git commit -m "feat: add nginx configs for dev and prod"
```

---

### Task 7: Docker Compose files

**Files:**
- Create: `deploy/docker-compose.dev.yml`
- Create: `deploy/docker-compose.prod.yml`

- [ ] **Step 1: Create `deploy/docker-compose.dev.yml`**

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.dev.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - dx-api
      - dx-web
    restart: unless-stopped

  dx-api:
    build:
      context: ../dx-api
      dockerfile: Dockerfile.dev
    volumes:
      - ../dx-api:/app
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
    env_file:
      - ./env/.env.dev
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  dx-web:
    build:
      context: ../dx-web
      dockerfile: Dockerfile.dev
    volumes:
      - ../dx-web:/app
      - /app/node_modules
      - /app/.next
    env_file:
      - ./env/.env.dev
    depends_on:
      - dx-api
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    env_file:
      - ./env/.env.dev
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
  redisdata:
  go-mod-cache:
  go-build-cache:
```

- [ ] **Step 2: Create `deploy/docker-compose.prod.yml`**

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.prod.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - dx-api
      - dx-web
    restart: always

  dx-api:
    build:
      context: ../dx-api
      dockerfile: Dockerfile
    env_file:
      - ./env/.env.prod
    volumes:
      - api-storage:/www/storage
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: always

  dx-web:
    build:
      context: ../dx-web
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-https://your-domain.com}
    env_file:
      - ./env/.env.prod
    depends_on:
      - dx-api
    restart: always

  postgres:
    image: postgres:16-alpine
    env_file:
      - ./env/.env.prod
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: always

  redis:
    image: redis:7-alpine
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    restart: always

volumes:
  pgdata:
  redisdata:
  api-storage:
```

- [ ] **Step 3: Verify compose config**

Run: `cd deploy && docker compose -f docker-compose.dev.yml config`
Expected: Valid YAML, all service definitions resolved, no errors.

- [ ] **Step 4: Commit**

```bash
git add deploy/docker-compose.dev.yml deploy/docker-compose.prod.yml
git commit -m "feat: add Docker Compose for dev and prod environments"
```

---

### Task 8: Remove old dx-api docker-compose.yml

**Files:**
- Delete: `dx-api/docker-compose.yml`

- [ ] **Step 1: Delete the old compose file**

The old `dx-api/docker-compose.yml` is superseded by `deploy/docker-compose.dev.yml` and `deploy/docker-compose.prod.yml`.

Run: `rm dx-api/docker-compose.yml`

- [ ] **Step 2: Commit**

```bash
git add dx-api/docker-compose.yml
git commit -m "chore: remove old dx-api docker-compose.yml (moved to deploy/)"
```

---

### Task 9: Smoke test — dev compose

> **Note:** The postgres container starts with an empty `dxdb` database. You'll need to restore a database dump or run migrations before API endpoints that query the DB will work. The health endpoint (`/api/health`) should respond regardless.

- [ ] **Step 1: Start the dev stack**

Run: `cd deploy && docker compose -f docker-compose.dev.yml up --build`

Expected: All 5 services start. Watch logs for:
- `postgres` — "database system is ready to accept connections"
- `redis` — "Ready to accept connections"
- `dx-api` — air starts, Go binary compiles and runs on `:3001`
- `dx-web` — Next.js dev server starts on `:3000`
- `nginx` — no errors

- [ ] **Step 2: Verify nginx routing**

Run (in a separate terminal):
```bash
# Health check via nginx -> dx-api
curl -s http://localhost/api/health

# Frontend via nginx -> dx-web
curl -s -o /dev/null -w "%{http_code}" http://localhost
```

Expected: `/api/health` returns JSON response, `/` returns 200.

- [ ] **Step 3: Verify hot reload — dx-api**

Make a trivial change to any `.go` file (add a comment). Watch the `dx-api` container logs.
Expected: Air detects the change, rebuilds, and restarts.

- [ ] **Step 4: Verify hot reload — dx-web**

Open `http://localhost` in a browser. Make a trivial change to any component.
Expected: Page updates automatically via HMR.

- [ ] **Step 5: Tear down**

Run: `cd deploy && docker compose -f docker-compose.dev.yml down`
