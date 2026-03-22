# Douxue Authentication Design

> Complete reference for the dual-token authentication system used across dx-api (Go/Goravel) and dx-web (Next.js 16).

## Architecture Overview

```
┌──────────────┐  ┌──────────────────┐  ┌──────────────────┐
│   dx-web     │  │  WeChat Mini-App │  │  Flutter App     │
│  (Next.js)   │  │  (future)        │  │  (future)        │
└──────┬───────┘  └────────┬─────────┘  └────────┬─────────┘
       │                   │                      │
       └───────────────────┼──────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   dx-api    │
                    │  (Goravel)  │
                    ├─────────────┤
                    │ /api/*      │  ← User JWT ("user" guard)
                    │ /adm/*      │  ← Admin JWT ("admin" guard)
                    ├─────────────┤
                    │  PostgreSQL │  ← users, adm_users
                    │  Redis      │  ← refresh tokens, codes, rate limits
                    └─────────────┘
```

The system uses **two separate tokens**:

| Token | Format | Lifetime | Storage (Browser) | Storage (Mobile) | Purpose |
|-------|--------|----------|-------------------|------------------|---------|
| **Access token** | JWT (HS256) | 10 minutes | JS memory variable | App secure storage | Authenticate API requests |
| **Refresh token** | Opaque 64-char hex | 7 days | httpOnly cookie (`dx_refresh`) | App secure storage | Obtain new access tokens |

The access token is short-lived and never persisted to disk. The refresh token is long-lived and stored in Redis server-side, delivered to browsers via an httpOnly cookie and to mobile clients via the response body.

## Two Auth Systems

The project has two independent auth systems sharing the same infrastructure:

| System | Guard | Users Table | Cookie | Route Prefix | Purpose |
|--------|-------|-------------|--------|-------------|---------|
| **User auth** | `"user"` | `users` | `dx_refresh` | `/api/auth/*` | End users (students) |
| **Admin auth** | `"admin"` | `adm_users` | `dx_adm_refresh` | `/adm/auth/*` | Admin panel operators |

Both use identical token flows. The guard field in Redis ensures tokens cannot be used cross-system.

---

## Token Flow Diagrams

### Sign-In / Sign-Up

```
Client                          dx-api                        Redis
  │                               │                             │
  │  POST /api/auth/signin        │                             │
  │  { account, password }        │                             │
  │──────────────────────────────►│                             │
  │                               │  Validate credentials       │
  │                               │  Issue JWT (10 min)         │
  │                               │  Generate refresh token     │
  │                               │─────────────────────────────►│
  │                               │  SET refresh:{token}        │
  │                               │  SADD user_refresh:{id}:user│
  │                               │                             │
  │  200 { access_token,          │                             │
  │        refresh_token, user }  │                             │
  │  Set-Cookie: dx_refresh=...   │                             │
  │◄──────────────────────────────│                             │
  │                               │                             │
  │  Store access_token in memory │                             │
```

### Normal API Request

```
Client                          dx-api
  │                               │
  │  GET /api/user/profile        │
  │  Authorization: Bearer {JWT}  │
  │──────────────────────────────►│
  │                               │  JwtAuth middleware:
  │                               │  Parse JWT → extract user ID
  │                               │
  │  200 { user profile }         │
  │◄──────────────────────────────│
```

### Token Refresh (Access Token Expired)

```
Client                          dx-api                        Redis
  │                               │                             │
  │  GET /api/user/profile        │                             │
  │  Authorization: Bearer {JWT}  │                             │
  │──────────────────────────────►│                             │
  │  401 Unauthorized             │                             │
  │◄──────────────────────────────│                             │
  │                               │                             │
  │  POST /api/auth/refresh       │                             │
  │  Cookie: dx_refresh=abc123    │                             │
  │──────────────────────────────►│                             │
  │                               │  GET refresh:abc123 ────────►│
  │                               │  ◄─── { user_id, guard }    │
  │                               │                             │
  │                               │  Verify guard == "user"     │
  │                               │  Load user from DB          │
  │                               │  Issue new JWT              │
  │                               │                             │
  │                               │  DEL refresh:abc123 ────────►│  (rotate)
  │                               │  SET refresh:def456 ────────►│
  │                               │  SREM + SADD user index ───►│
  │                               │                             │
  │  200 { access_token,          │                             │
  │        refresh_token }        │                             │
  │  Set-Cookie: dx_refresh=def456│                             │
  │◄──────────────────────────────│                             │
  │                               │                             │
  │  Retry GET /api/user/profile  │                             │
  │  Authorization: Bearer {new}  │                             │
  │──────────────────────────────►│                             │
  │  200 { user profile }         │                             │
  │◄──────────────────────────────│                             │
```

### Logout

```
Client                          dx-api                        Redis
  │                               │                             │
  │  POST /api/auth/logout        │                             │
  │  Cookie: dx_refresh=def456    │                             │
  │──────────────────────────────►│                             │
  │                               │  DEL refresh:def456 ────────►│
  │                               │  SREM user index ──────────►│
  │                               │                             │
  │  200 { ok }                   │                             │
  │  Set-Cookie: dx_refresh=;     │                             │
  │              Max-Age=0        │                             │
  │◄──────────────────────────────│                             │
  │                               │                             │
  │  Clear access_token from      │                             │
  │  memory + delete cookie       │                             │
```

### Page Refresh (Browser Reload)

```
1. Page reloads → access token lost (was only in memory)
2. AuthGuard component mounts
3. Checks getAccessToken() → null
4. Calls POST /api/auth/refresh (cookie auto-sent)
5. Gets new access token → stores in memory
6. User stays logged in seamlessly

If refresh token is also expired → redirect to /auth/signin
```

---

## Sign-In Methods

The user auth system supports three sign-in methods:

### 1. Email + Verification Code

```
POST /api/auth/signin/send-code  →  { email }
POST /api/auth/signin            →  { email, code }
```

- 6-digit numeric code sent to email
- Code stored in Redis (`signin_code:{email}`, TTL 300s)
- Rate limited: 1 code per email per 60 seconds
- **Auto-registration:** if email not found, creates a new user automatically
  - Username derived from email prefix (e.g., `john` from `john@example.com`)
  - If username taken, appends 4-digit suffix (e.g., `john_4829`)
  - Random 16-char password generated

### 2. Account + Password

```
POST /api/auth/signin  →  { account, password }
```

- `account` can be username, email, or phone
- Password verified with bcrypt (cost 12)
- No auto-registration — user must exist

### 3. Sign-Up (Explicit Registration)

```
POST /api/auth/signup/send-code  →  { email }
POST /api/auth/signup            →  { email, code, username?, password? }
```

- Email verification code required
- Username optional (derived from email if empty)
- Password optional (random 16-char if empty)
- Duplicate email/username checks
- On success: auto-login (returns tokens, redirects to `/hall`)

---

## Admin Authentication

```
POST /adm/auth/login  →  { username, password }
```

- Separate `adm_users` table
- `is_active` flag checked (inactive admins blocked)
- Same dual-token pattern as user auth
- Cookie name: `dx_adm_refresh`
- Login audited to `adm_logins` table

### Admin Guard (User-Facing Admin)

Routes at `/api/admin/*` use user JWT + an additional middleware:

```
JwtAuth → AdminGuard → handler
```

`AdminGuard` checks if the authenticated user's username is `"rainson"`. This is a hardcoded superuser check for user-facing admin operations (notice management, redeem code generation).

---

## API Endpoints

### User Auth Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/auth/signup/send-code` | Public | Send signup verification code |
| `POST` | `/api/auth/signup` | Public | Register + auto-login |
| `POST` | `/api/auth/signin/send-code` | Public | Send signin verification code |
| `POST` | `/api/auth/signin` | Public | Sign in (email+code or account+password) |
| `POST` | `/api/auth/refresh` | Public | Get new access token (rate limited) |
| `POST` | `/api/auth/logout` | Public | Invalidate refresh token |
| `GET` | `/api/auth/me` | JWT | Get current user profile |

### Admin Auth Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/adm/auth/login` | Public | Admin login |
| `POST` | `/adm/auth/refresh` | Public | Admin token refresh (rate limited) |
| `POST` | `/adm/auth/logout` | Public | Admin logout |
| `GET` | `/adm/auth/me` | Admin JWT | Get admin user profile |

### Route Protection Summary

| Prefix | Middleware | Description |
|--------|-----------|-------------|
| `/api/*` (public) | None | Game listing, auth, image serving |
| `/api/*` (protected) | `JwtAuth` | Profile, sessions, tracking, etc. |
| `/api/admin/*` | `JwtAuth` + `AdminGuard` | User-facing admin operations |
| `/adm/*` (public) | None | Admin auth endpoints |
| `/adm/*` (protected) | `AdmJwtAuth` + `AdmRbac` + `AdmOperateLog` | Admin panel CRUD |

---

## Redis Key Structure

### Verification Codes

| Key | Value | TTL | Description |
|-----|-------|-----|-------------|
| `signup_code:{email}` | 6-digit string | 300s | Signup verification code |
| `signin_code:{email}` | 6-digit string | 300s | Signin verification code |

### Rate Limiting (Sliding Window)

| Key | Type | Window | Description |
|-----|------|--------|-------------|
| `rate:signup_code:{email}` | Sorted set | 60s, max 1 | Signup code send limit |
| `rate:signin_code:{email}` | Sorted set | 60s, max 1 | Signin code send limit |
| `rate:refresh:{ip}` | Sorted set | 60s, max 10 | User token refresh limit |
| `rate:adm_refresh:{ip}` | Sorted set | 60s, max 10 | Admin token refresh limit |

### Refresh Tokens

| Key | Type | TTL | Description |
|-----|------|-----|-------------|
| `refresh:{token}` | String (JSON) | 7 days | Token → `{"user_id":"...","guard":"user"}` |
| `user_refresh:{userId}:{guard}` | Set | 7 days | User's active refresh tokens (for bulk revoke) |

---

## Cookie Configuration

### `dx_refresh` (User)

| Attribute | Value |
|-----------|-------|
| Name | `dx_refresh` |
| HttpOnly | `true` (not accessible by JavaScript) |
| Secure | Configurable (`REFRESH_COOKIE_SECURE`, default `true`) |
| SameSite | `Lax` |
| Path | `/` |
| Max-Age | 604,800 seconds (7 days) |

### `dx_adm_refresh` (Admin)

Same attributes as above, different cookie name.

### Why `Path=/`?

The cookie must be sent with **all** requests so the Next.js proxy middleware (`proxy.ts`) can check authentication status during page navigation. A scoped path like `/api/auth` would prevent the cookie from being sent with navigation requests to `/hall`.

---

## Frontend Architecture

### Token Storage

```
┌─────────────────────────────────────────────────────┐
│                    Browser                          │
│                                                     │
│  ┌──────────────────┐   ┌────────────────────────┐ │
│  │  JS Memory       │   │  httpOnly Cookie        │ │
│  │  ──────────────   │   │  ─────────────────────  │ │
│  │  access_token     │   │  dx_refresh = abc123   │ │
│  │  (lost on reload) │   │  (survives reload)     │ │
│  └──────────────────┘   └────────────────────────┘ │
│                                                     │
│  lib/token.ts            Set by dx-api response     │
│  getAccessToken()        Cannot be read by JS       │
│  setAccessToken()        Auto-sent via credentials  │
│  clearAccessToken()      Cleared by server or JS    │
└─────────────────────────────────────────────────────┘
```

### Refresh Lock (Concurrent 401 Handling)

When multiple API calls receive 401 simultaneously, only one refresh request is sent:

```typescript
let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  if (refreshPromise) return refreshPromise;  // queue behind existing refresh

  refreshPromise = fetch("/api/auth/refresh", {
    method: "POST",
    credentials: "include",
  })
    .then(/* parse, store new token */)
    .finally(() => { refreshPromise = null; });

  return refreshPromise;
}
```

All concurrent callers share the same promise. Once it resolves, they all retry with the new token.

### Route Protection

Two layers of route protection on the frontend:

**1. `proxy.ts` (Server-Side Middleware)**

Runs on every navigation. Checks the `dx_refresh` cookie:
- Has cookie + visiting `/auth/*` → redirect to `/hall`
- No cookie + visiting `/hall/*` → redirect to `/auth/signin`
- Cleans up legacy `dx_token` cookie if present

**2. `AuthGuard` Component (Client-Side)**

Wraps protected page content. On mount:
- Checks `getAccessToken()` — if present, render children
- If null, attempts silent refresh via cookie
- If refresh succeeds, render children
- If refresh fails, redirect to `/auth/signin`

---

## Validation

### Frontend (Zod Schemas)

**Sign-In:**
- Email: valid email format, required
- Code: exactly 6 digits
- Account: non-empty string
- Password: non-empty string

**Sign-Up:**
- Email: valid email format, required
- Code: exactly 6 digits
- Username: optional, max 30 chars, alphanumeric + underscore + hyphen
- Password: optional, min 8 chars, must include lowercase + uppercase + digit
- Agreed: must be `true` (terms acceptance)

### Backend (Service Layer)

- Verification codes: 6-digit match against Redis, auto-deleted after use
- Duplicate checks: email and username uniqueness queries
- Password: bcrypt comparison (cost 12)
- Refresh token: Redis lookup + guard validation

---

## Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `0` | 200 | Success |
| `40000` | 400 | Validation error |
| `40001` | 400 | Invalid email format |
| `40002` | 400 | Invalid password |
| `40003` | 409 | Duplicate email |
| `40004` | 409 | Duplicate username |
| `40005` | 400 | Invalid/expired verification code |
| `40100` | 401 | Unauthorized (missing/invalid JWT) |
| `40103` | 401 | Invalid/expired refresh token |
| `40300` | 403 | Forbidden (not admin) |
| `42900` | 429 | Rate limited |
| `50000` | 500 | Internal server error |
| `50002` | 500 | Email send error |

All responses follow the envelope format: `{ "code": N, "message": "...", "data": ... }`

---

## Security Measures

| Measure | Implementation |
|---------|---------------|
| Password hashing | bcrypt, cost factor 12 |
| JWT signing | HS256 with `JWT_SECRET` from environment |
| Token rotation | Each refresh deletes old token, issues new one |
| XSS protection | Refresh token in httpOnly cookie (JS can't access) |
| CSRF mitigation | `SameSite=Lax` cookie attribute |
| Rate limiting | Redis sliding window on code sends (1/60s) and refreshes (10/60s) |
| Guard isolation | Refresh tokens tagged with guard (`"user"` or `"admin"`), validated on use |
| Audit logging | Login records with IP + User-Agent stored in `user_logins` / `adm_logins` |
| Short-lived access | JWT expires in 10 minutes, limiting exposure window |
| Bulk revocation | `DeleteUserRefreshTokens()` can invalidate all sessions for a user |

---

## Multi-Client Support

The API returns both tokens in the response body AND sets the httpOnly cookie. Each client type uses the appropriate mechanism:

| Client | Access Token Storage | Refresh Token Storage | Refresh Delivery |
|--------|---------------------|----------------------|-----------------|
| dx-web (Browser) | JS memory variable | httpOnly cookie (automatic) | Cookie auto-sent |
| WeChat Mini-Program | `wx.setStorageSync()` | `wx.setStorageSync()` | Request body |
| Flutter App | `flutter_secure_storage` | `flutter_secure_storage` | Request body |

The refresh and logout endpoints accept the refresh token from **either** the `dx_refresh` cookie or the `refresh_token` field in the request body. Cookie is checked first.

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | (required) | HMAC secret for signing JWTs |
| `JWT_TTL` | `10` | Access token lifetime in minutes |
| `REFRESH_TOKEN_TTL` | `10080` | Refresh token lifetime in minutes (7 days) |
| `REFRESH_COOKIE_SECURE` | `true` | Set `false` for local dev (HTTP) |

### Goravel Auth Guards

```go
// config/auth.go
"user"  → JWT driver, users table (User model)
"admin" → JWT driver, adm_users table (AdmUser model)
```

---

## File Reference

### Backend (dx-api)

| File | Responsibility |
|------|---------------|
| `config/jwt.go` | JWT + refresh token configuration |
| `config/auth.go` | Guard definitions (user, admin) |
| `app/helpers/refresh_token.go` | Refresh token generation, Redis CRUD, user index |
| `app/helpers/hash.go` | bcrypt password hashing |
| `app/helpers/rate_limit.go` | Redis sliding window rate limiter |
| `app/helpers/redis.go` | Redis client singleton |
| `app/helpers/random.go` | Verification code + invite code generation |
| `app/services/api/auth_service.go` | User auth logic (signup, signin, refresh, logout) |
| `app/services/adm/auth_service.go` | Admin auth logic |
| `app/http/controllers/api/auth_controller.go` | User auth HTTP handlers + cookie helpers |
| `app/http/controllers/adm/auth_controller.go` | Admin auth HTTP handlers + cookie helpers |
| `app/http/middleware/jwt_auth.go` | User JWT validation middleware |
| `app/http/middleware/adm_jwt_auth.go` | Admin JWT validation middleware |
| `app/http/middleware/admin_guard.go` | User-facing admin check (username == "rainson") |
| `app/consts/error_code.go` | Numeric error code constants |
| `app/services/api/errors.go` | User auth error sentinels |
| `app/services/adm/errors.go` | Admin auth error sentinels |
| `routes/api.go` | User API route registration |
| `routes/adm.go` | Admin API route registration |

### Frontend (dx-web)

| File | Responsibility |
|------|---------------|
| `src/lib/token.ts` | In-memory access token store |
| `src/lib/api-client.ts` | HTTP client with auto-refresh on 401 |
| `src/proxy.ts` | Next.js middleware — route protection via cookie check |
| `src/components/in/auth-guard.tsx` | Client-side auth guard with silent refresh |
| `src/features/web/auth/hooks/use-signin.ts` | Sign-in form state + handlers |
| `src/features/web/auth/hooks/use-signup.ts` | Sign-up form state + handlers |
| `src/features/web/auth/schemas/signin.schema.ts` | Zod validation for sign-in |
| `src/features/web/auth/schemas/signup.schema.ts` | Zod validation for sign-up |
| `src/features/web/auth/services/user.service.ts` | User profile fetch service |
