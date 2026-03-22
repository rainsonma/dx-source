# Dual-Token Authentication Design

> Replace single JWT auth with access token + refresh token strategy for dx-api and dx-web.

## Overview

| Item | Detail |
|------|--------|
| **Goal** | Short-lived access token (JWT, 10 min) + long-lived refresh token (opaque, 7 days) |
| **Access token** | JWT, stored in browser memory (JS variable), sent via `Authorization: Bearer` header |
| **Refresh token** | Opaque 64-char hex string, stored in Redis, delivered via httpOnly cookie + response body |
| **Multi-client** | Browser uses cookie; WeChat/Flutter use response body token with their own secure storage |
| **Scope** | Both user auth (`/api/*`) and admin auth (`/adm/*`) |
| **Breaking change** | Yes — all active sessions invalidated on deploy, users must re-login |

### Multi-Client Token Storage

| Client | Access token | Refresh token | Refresh delivery |
|--------|-------------|---------------|-----------------|
| dx-web (browser) | JS memory variable | httpOnly cookie (auto-managed) | Cookie auto-sent |
| WeChat mini-program | `wx.setStorageSync()` | `wx.setStorageSync()` | Request body |
| Flutter app | `flutter_secure_storage` | `flutter_secure_storage` | Request body |

The dx-api always returns both tokens in the response body AND sets the httpOnly cookie. Browser clients use the cookie; non-browser clients use the body.

## Token Lifecycle

### Login / SignUp

```
User submits credentials
        ↓
dx-api validates → issues JWT access token (10 min) + generates opaque refresh token
        ↓
Response body: { "data": { "access_token": "eyJ...", "refresh_token": "a1b2c3...", "user": {...} } }
Set-Cookie: dx_refresh=<opaque>; HttpOnly; Secure; SameSite=Lax; Path=/api/auth; Max-Age=604800
        ↓
dx-web stores access token in memory (browser ignores refresh_token in body, uses cookie)
WeChat/Flutter stores both tokens in their own secure storage
```

- Refresh cookie `Path=/api/auth` — browser only sends it to auth endpoints
- `Secure` flag for HTTPS-only (configurable, disabled in dev)
- `SameSite=Lax` — safe because production uses nginx reverse proxy on the same domain (path routing: `/api/*` → Go, `/` → Next.js)
- Access token in response body only, never in a cookie
- Redis key: `refresh:{token}` → JSON `{"user_id":"...","guard":"user"}` with 7-day TTL

### Normal Request

```
Browser → GET /api/orders (Authorization: Bearer <access_token>)
dx-api → validates JWT → 200 OK
```

### Refresh (Access Token Expired)

```
Client → GET /api/orders → 401
Client → POST /api/auth/refresh
         Browser: cookie auto-sent, no body needed
         WeChat/Flutter: { "refresh_token": "a1b2c3..." } in request body
dx-api → reads refresh token (cookie first, fall back to body) → looks up Redis → issues new JWT
Response body: { "data": { "access_token": "eyJ...", "refresh_token": "d4e5f6..." } }
Set-Cookie: dx_refresh=<new_opaque>; ... (rotate refresh token)
Client → retries GET /api/orders with new access token
```

Refresh token rotation: each refresh issues a new refresh token and deletes the old one from Redis.

**Backend reads refresh token from:** `dx_refresh` cookie first → `refresh_token` in request body as fallback. This allows the same endpoint to serve browser and non-browser clients.

### Logout

```
Client → POST /api/auth/logout
         Browser: cookie auto-sent
         WeChat/Flutter: { "refresh_token": "a1b2c3..." } in request body
dx-api → reads refresh token (cookie first, fall back to body) → deletes from Redis
Set-Cookie: dx_refresh=; Max-Age=0 (clear cookie)
Browser → clears access token from memory + clears cookie client-side (safety fallback)
WeChat/Flutter → clears both tokens from secure storage
```

### Page Refresh

Access token is lost (memory-only). Next API call gets 401, triggers refresh via cookie. If refresh token is still valid, user stays logged in seamlessly.

### User Info Restoration

After a page refresh, the auth guard component attempts a silent refresh. On success, it calls `GET /api/auth/me` (protected by JwtAuth) with the new access token to restore user info. The refresh response itself only contains `access_token` — user data is fetched separately.

## Backend Changes (dx-api)

### Config — `config/jwt.go`

- `JWT_TTL` default: 60 → 10 minutes
- Remove `refresh_ttl` (no longer using Goravel's built-in refresh)
- Add `REFRESH_TOKEN_TTL` env var (default: 10080 minutes = 7 days)
- Add `REFRESH_COOKIE_SECURE` env var (default: true, false for local dev)

### New Helper — `app/helpers/refresh_token.go`

| Function | Purpose |
|----------|---------|
| `GenerateRefreshToken()` | 64-char crypto-random hex string |
| `StoreRefreshToken(token, userId, guard)` | Redis SET with 7-day TTL + add to user index |
| `LookupRefreshToken(token)` | Returns userId + guard, or error |
| `DeleteRefreshToken(token, userId, guard)` | DEL from Redis + remove from user index |
| `DeleteUserRefreshTokens(userId, guard)` | Iterate user index SET, DEL each token, DEL index |

**Redis key structure:**
- `refresh:{token}` → JSON `{"user_id":"...","guard":"user"}` with 7-day TTL
- `user_refresh:{userId}:{guard}` → Redis SET of active token strings (for "logout everywhere")

On token creation: `SET refresh:{token}` + `SADD user_refresh:{userId}:{guard} {token}`.
On rotation/logout: `DEL refresh:{token}` + `SREM user_refresh:{userId}:{guard} {token}`.
On revoke all: iterate SET members, DEL each `refresh:{token}`, then DEL the SET.

### Auth Service — `app/services/api/auth_service.go`

- `SignIn` / `SignUp`: after issuing JWT access token, call `GenerateRefreshToken()` + `StoreRefreshToken()`. Return both.
- `RefreshToken(refreshToken)`: look up Redis → validate → issue new JWT → rotate refresh token (delete old, store new) → return new access token + new refresh token.
- `Logout(refreshToken)`: delete refresh token from Redis. Replaces Goravel's `Guard.Logout()`.

### Auth Controller — `app/http/controllers/api/auth_controller.go`

- `SignIn` / `SignUp`: return access token + refresh token in body, also set refresh token as httpOnly cookie via `ctx.Response().Cookie()`.
- `Refresh`: read refresh token from `dx_refresh` cookie first, fall back to `refresh_token` in request body. Call service, set new cookie, return new access token + new refresh token.
- `Logout`: read refresh token from cookie or body, call service, clear cookie with `Max-Age=0`.

### Routes — `routes/api.go`

- `POST /api/auth/refresh` — move from protected (JwtAuth) to **public** group (access token is expired when this is called). Rate-limited: 10 requests/min per IP.
- `POST /api/auth/logout` — move to public (auth comes from refresh token cookie)

### Admin Auth

Same pattern applied to:
- `app/services/adm/auth_service.go`
- `app/http/controllers/adm/auth_controller.go`
- `routes/adm.go`

With guard `"admin"` and cookie name `dx_adm_refresh`.

**Admin-specific details:**
- Cookie: `dx_adm_refresh`; `Path=/adm/auth`; same HttpOnly/Secure/SameSite flags
- `POST /adm/auth/refresh` — new public endpoint (rate-limited: 10 req/min per IP)
- `POST /adm/auth/logout` — move from protected (AdmJwtAuth) to public
- Redis keys use guard `"admin"`: `refresh:{token}` → `{"user_id":"...","guard":"admin"}`

## Frontend Changes (dx-web)

### New Token Module — `lib/token.ts`

```typescript
let accessToken: string | null = null

export function getAccessToken(): string | null { return accessToken }
export function setAccessToken(token: string): void { accessToken = token }
export function clearAccessToken(): void { accessToken = null }
```

No localStorage, no cookie. Token lives only in JS memory. Lost on page refresh — next API call triggers refresh via cookie.

### API Client Rewrite — `lib/api-client.ts`

- Read token from `getAccessToken()` instead of `localStorage`
- On 401: attempt refresh before redirecting
- Refresh lock — only one refresh request at a time, others queue:

```typescript
let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  if (refreshPromise) return refreshPromise

  refreshPromise = fetch(`${API_URL}/api/auth/refresh`, {
    method: "POST",
    credentials: "include",
  })
  .then(res => {
    if (!res.ok) throw new Error("refresh failed")
    return res.json()
  })
  .then(data => {
    setAccessToken(data.data.access_token)
    return data.data.access_token
  })
  .finally(() => { refreshPromise = null })

  return refreshPromise
}
```

- On 401: call `refreshAccessToken()`, retry original request once. If refresh fails → clear token, redirect to `/auth/signin`.
- `uploadApi.uploadImage` bypasses `apiFetch` — must also be updated to use `getAccessToken()` and the 401-retry-with-refresh logic.

### Auth Hooks — `features/web/auth/hooks/`

- `use-signin.ts`: on success, call `setAccessToken(data.access_token)`. Refresh token auto-set by browser from `Set-Cookie` header.
- `use-signup.ts`: on success, call `setAccessToken(data.access_token)` and redirect to `/hall`. **Behavior change:** currently redirects to `/auth/signin` — changing to auto-login after signup.
- Remove `setToken()` / `setAuthCookie()` calls.

### Protected Route Guard — `components/in/auth-guard.tsx`

Client-side wrapper component:
- On mount: check `getAccessToken()`
- If no token: attempt silent refresh (cookie might still be valid)
- If refresh fails: redirect to `/auth/signin`
- Protected layouts use this instead of server-side `auth()` checks

### Logout

- Call `POST /api/auth/logout` with `credentials: "include"`
- Call `clearAccessToken()`
- Clear cookie client-side: `document.cookie = "dx_refresh=; path=/api/auth; max-age=0"`
- Redirect to `/auth/signin`

## Migration & Cleanup

### Files to Remove

| File | Reason |
|------|--------|
| `dx-web/src/lib/auth.ts` | Server-side JWT decode, no longer needed |
| `dx-web/src/lib/api-server.ts` | Server-side authenticated fetch, no longer needed |

### Files to Refactor

34 files currently import `apiServerFetch`. All authenticated server-side fetches must be converted to client-side `apiClient` calls or unauthenticated server fetches where appropriate.

**Key categories:**

| Category | Files | Change |
|----------|-------|--------|
| Page components (SSR data fetch) | ~14 `page.tsx` files in `app/(web)/` | Convert to client-side data fetching with `apiClient` |
| Server actions | ~16 action files in `features/web/*/actions/` | Convert to client-side calls or thin wrappers around `apiClient` |
| Auth services | `features/web/auth/services/user.service.ts` | Convert to client-side |
| Helpers | `features/web/hall/helpers/has-unread-notices.ts` | Convert to client-side |
| Components | `features/web/ai-custom/components/course-detail-content.tsx` | Convert to client-side |
| Proxy | `proxy.ts` | Remove auth-dependent redirect logic |
| Legacy cleanup | Remove `dx_token` cookie from returning users (clear on first visit) |

### Environment Updates

| Variable | Dev | Prod |
|----------|-----|------|
| `JWT_TTL` | `10` | `10` |
| `REFRESH_TOKEN_TTL` | `10080` | `10080` |
| `REFRESH_COOKIE_SECURE` | `false` | `true` |

### No Database Changes

Everything uses Redis. No new tables needed.

### Goravel Compatibility Note

Goravel's `Auth().Guard().Parse()` validates JWT expiration internally. With `JWT_TTL=10`, tokens expire after 10 minutes. The `refresh_ttl` config is no longer used since we handle refresh externally via opaque tokens. Verify during implementation that `Parse()` does not have an internal grace period that extends token validity beyond the configured TTL.

### Deployment Note

Breaking change — all active sessions invalidated. Users must log in again after deployment.
