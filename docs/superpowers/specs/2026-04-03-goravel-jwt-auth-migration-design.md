# Migrate JWT Auth to Goravel Default Practice

**Date:** 2026-04-03
**Status:** Approved
**Scope:** dx-api (backend) + dx-web (frontend) + deploy

## Goal

Replace the custom dual-token JWT auth system (10-min access JWT + opaque refresh token in Redis) with Goravel's built-in single-JWT auth (`LoginUsingID` / `Parse` / `Refresh` / `Logout`). Keep single-device enforcement. Eliminate all client-side token management.

## Current State

- Custom JWT issuance via `helpers.IssueAccessToken()` with custom `auth_id` claim
- Opaque 64-char hex refresh token stored in Redis (`refresh:{token}`)
- 10-min access token TTL, 7-day refresh token TTL
- Access token in-memory on frontend, refresh token in httpOnly cookie `dx_refresh`
- Single-device enforcement via `auth_id` UUID in Redis (`user_auth:{userID}:{guard}`)
- Frontend manages tokens: Bearer header injection, silent refresh, retry on 401, SSE token query params

## Target State

- Goravel's `LoginUsingID()` for token generation, `Refresh()` for refresh, `Logout()` for blacklisting
- Single JWT (60-min TTL, 14-day refresh window) stored in httpOnly cookie `dx_token`
- Auto-refresh in middleware (Goravel's official pattern) — client never sees token expiry
- Single-device enforcement via `iat >= login timestamp` check in Redis
- Frontend has zero token management — `credentials: "include"` on all requests, browser sends cookie automatically

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| TTL | 60 min | Goravel default, covers most gameplay sessions |
| Refresh window | 14 days | Goravel default, user stays logged in for 2 weeks of inactivity |
| Token storage | httpOnly cookie | XSS-safe, automatic on all requests, SSR-compatible |
| Refresh mechanism | Auto-refresh in middleware | Goravel's official example pattern, zero client-side logic |
| Single-device key | `iat >= login timestamp` | Store login time in Redis; any token with `iat` before that is from a previous session. Refreshed tokens always have `iat` after login time, so auto-refresh never triggers false SessionReplaced. |
| Token count | 1 (JWT only) | No opaque refresh token, no Redis storage for tokens |

## Backend Changes (dx-api)

### Configuration

**`config/jwt.go`** — replace custom config with Goravel defaults + cookie config:

```go
"secret":        config.Env("JWT_SECRET", ""),
"ttl":           config.Env("JWT_TTL", 60),
"refresh_ttl":   config.Env("JWT_REFRESH_TTL", 20160),
"cookie_name":   "dx_token",
"cookie_secure": config.Env("JWT_COOKIE_SECURE", true),
```

Remove the `refresh_token` nested map.

### Middleware (`jwt_auth.go`)

Rewrite to follow Goravel's official example + single-device check + cookie I/O:

```
1. Read JWT from cookie "dx_token"
   → empty? → 401 Unauthorized
2. Parse(token) via Goravel
   → expired? → Refresh()
     → success? → set new "dx_token" cookie on response, continue
     → fail (refresh window exceeded)? → clear cookie, 401
   → invalid? → clear cookie, 401
3. Get userID via facades.Auth(ctx).Guard("user").ID()
4. Get iat from Goravel's Payload (payload.IssuedAt)
5. Get login timestamp from Redis: user_auth:{userID}:user
   → iat < login timestamp? → clear cookie, 401 SessionReplaced
   (token was issued before the most recent login — another device logged in)
6. ctx.Request().Next()
```

Same pattern for `adm_jwt_auth.go` with `"admin"` guard and `"dx_adm_token"` cookie.

### Auth Services

**`services/api/auth_service.go`** — rewrite `issueSession`:

```
func issueSession(ctx, userID) → (token string, err error)
  1. facades.Auth(ctx).Guard("user").LoginUsingID(userID) → token
  2. Store time.Now() as login timestamp in Redis: user_auth:{userID}:user (TTL = 14 days)
     (any token with iat before this timestamp is from a previous session)
  3. Return token
```

**`AuthResult` struct** — remove `RefreshToken` field, just return token string.

**Remove:**
- `RefreshToken()` method
- All `auth_id` UUID generation
- All `DeleteUserRefreshTokens()` calls

**`Logout()`** changes to:
```
1. facades.Auth(ctx).Guard("user").Logout()  // blacklist token in cache
2. Delete Redis key user_auth:{userID}:user
3. Return success
```

Same changes for `services/adm/auth_service.go`.

### Auth Controllers

**`controllers/api/auth_controller.go`:**

SignUp / SignIn:
- Call `issueSession(ctx, userID)`
- Set httpOnly cookie `dx_token` with returned JWT
- Return `{ user }` only (no tokens in response body)

Logout:
- Call service Logout (blacklists token + deletes Redis key)
- Clear `dx_token` cookie
- Return success

**Remove:**
- `Refresh()` controller method (middleware handles it)
- `setRefreshCookie()` / `clearRefreshCookie()` / `getRefreshToken()` helpers
- Inline rate limiting for refresh endpoint

**New cookie helpers:**
```go
func setTokenCookie(ctx, token string)  // sets "dx_token", httpOnly, Secure, SameSite=Lax
func clearTokenCookie(ctx)              // clears "dx_token"
```

Same pattern for admin controller with `dx_adm_token`.

### Routes

Remove:
```
POST /api/auth/refresh
POST /adm/auth/refresh
```

Keep:
```
POST /api/auth/signup     (public)
POST /api/auth/signin     (public)
POST /api/auth/logout     (public)
GET  /api/auth/me         (protected)
POST /adm/auth/login      (public)
POST /adm/auth/logout     (public)
GET  /adm/auth/me         (protected)
```

### Files to Delete

| File | Reason |
|---|---|
| `helpers/jwt.go` | Custom JWT issuance, Claims struct, ExtractAuthID — replaced by Goravel |
| `helpers/refresh_token.go` | Opaque refresh token system — eliminated |

### Error Codes

| Code | Action |
|---|---|
| `40100` Unauthorized | Keep |
| `40101` TokenExpired | Remove — client never sees this |
| `40102` InvalidToken | Keep |
| `40103` InvalidRefreshToken | Remove — no refresh tokens |
| `40104` SessionReplaced | Keep |

### Error Sentinels

| Error | Action |
|---|---|
| `ErrInvalidRefreshToken` (api + adm) | Remove |
| `ErrSessionReplaced` (api + adm) | Keep |

### Redis Keys

| Key | Action |
|---|---|
| `user_auth:{userID}:{guard}` | Keep — stores login timestamp (Unix seconds) instead of UUID. Comparison: token's `iat >= stored value`. |
| `refresh:{token}` | Remove (auto-expires in 7 days) |
| `user_refresh:{userID}:{guard}` | Remove (auto-expires in 7 days) |

## Frontend Changes (dx-web)

### Core Principle

Zero client-side token management. All requests use `credentials: "include"`. The browser sends the `dx_token` cookie automatically. The server sets/clears/refreshes the cookie.

### Delete Files

| File | Reason |
|---|---|
| `lib/token.ts` | In-memory token storage — no purpose with cookie auth |
| `components/in/auth-guard.tsx` | Session recovery via refresh — replaced by proxy.ts route protection + middleware auto-refresh |

### Rewrite: `lib/api-client.ts`

Remove:
- All imports from `token.ts` (`getAccessToken`, `setAccessToken`, `clearAccessToken`) and aliases (`getToken`, `setToken`, `removeToken`)
- `refreshAccessToken()` function and `refreshPromise` deduplication variable
- Bearer header injection in `apiFetch()` (`Authorization: Bearer {token}`)
- 401 retry logic (refresh → retry → redirect)
- Error code `40104` handling inside `refreshAccessToken()`

Simplify `apiFetch()` to:
- `fetch()` with `credentials: "include"`, parse response, handle errors
- On 401: if code `40104` (SessionReplaced) → alert + redirect to sign-in. Otherwise → redirect to sign-in. No retry.

Change `authApi`:
- `signUp` / `signIn` — remove `setAccessToken()` call, just redirect on success (cookie set by server)
- `logout` — remove `removeToken()` call, just call API + redirect

### Edit: `proxy.ts`

- Check `dx_token` cookie instead of `dx_refresh`
- Remove legacy `dx_token` cleanup code (it's now the primary cookie)
- Remove `dx_refresh` references

### Edit: SSE Hooks

**`hooks/use-group-sse.ts` and `hooks/use-group-notify.ts`:**
- Remove `getToken`, `refreshAccessToken` imports
- Remove `?token=${encodeURIComponent(token)}` from URL
- Remove `refreshAndConnect()` error handler and exponential backoff
- Use `new EventSource(url, { withCredentials: true })` — cookie sent automatically, EventSource auto-reconnects on error

### Edit: AI Streaming Helpers

**`features/web/ai-custom/helpers/stream-progress.ts`, `generate-api.ts`, `format-api.ts`:**
- Remove `getToken` import
- Remove `Authorization: Bearer {token}` header
- Add `credentials: "include"` to fetch options

### Edit: Game Play Shells

**`features/web/play-single/components/game-play-shell.tsx` and `play-group/components/group-play-shell.tsx`:**
- Remove `getToken` import
- `beforeunload` sync: use `credentials: "include"` instead of Bearer header

### Edit: Image Uploader

**`features/com/images/hooks/use-image-uploader.ts`:**
- Remove `getToken` import
- Remove `Authorization: Bearer {token}` from Uppy XHRUpload headers
- Add `withCredentials: true` to Uppy XHRUpload config

### Edit: Auth Hooks

**`features/web/auth/hooks/use-signin.ts` and `use-signup.ts`:**
- Remove `setAccessToken(res.data.access_token)` calls
- Just redirect to `/hall` on success

### Edit: User Profile Menu

**`features/web/auth/components/user-profile-menu.tsx`:**
- Remove `removeToken()` call in `handleSignOut()`
- Just `await authApi.logout()` then `window.location.href = "/"`

### Edit: Landing Header

**`components/in/landing-header.tsx`:**
- Remove `getAccessToken()` check
- Detect login state server-side: parent page component reads `dx_token` cookie existence via `cookies()` and passes `isLoggedIn` prop

### Edit: Hall Layout

**`app/(web)/hall/layout.tsx`:**
- Remove `AuthGuard` wrapper — `proxy.ts` handles route protection

### Edit: Upload API in api-client.ts

**`uploadApi.uploadImage()`:**
- Remove `getToken()` and `Authorization: Bearer {token}` header
- Use `credentials: "include"`

## Deploy Changes

### Environment Variables

| Variable | Action |
|---|---|
| `JWT_TTL` | Change from `10` to `60` |
| `JWT_REFRESH_TTL` | New, set to `20160` |
| `REFRESH_TOKEN_TTL` | Remove |
| `REFRESH_COOKIE_SECURE` | Rename to `JWT_COOKIE_SECURE` |

Files: `deploy/env/.env.dev`, `deploy/env/.env.prod`, `deploy/env/.env.example`

## Migration

- All existing sessions invalidated on deploy — every user must sign in again
- Old `dx_refresh` / `dx_adm_refresh` cookies expire naturally (7 days)
- Old Redis keys (`refresh:*`, `user_refresh:*`) expire naturally (7 days)
- Optional post-deploy Redis cleanup: `SCAN + DEL refresh:*` and `SCAN + DEL user_refresh:*`

## File Change Summary

### Backend (dx-api) — 14 files

| Action | Files |
|---|---|
| Delete | `helpers/jwt.go`, `helpers/refresh_token.go` |
| Rewrite | `middleware/jwt_auth.go`, `middleware/adm_jwt_auth.go` |
| Rewrite | `services/api/auth_service.go`, `services/adm/auth_service.go` |
| Rewrite | `controllers/api/auth_controller.go`, `controllers/adm/auth_controller.go` |
| Edit | `config/jwt.go` |
| Edit | `routes/api.go`, `routes/adm.go` |
| Edit | `consts/error_code.go` |
| Edit | `services/api/errors.go`, `services/adm/errors.go` |

### Frontend (dx-web) — 17 files

| Action | Files |
|---|---|
| Delete | `lib/token.ts`, `components/in/auth-guard.tsx` |
| Rewrite | `lib/api-client.ts` |
| Edit | `proxy.ts` |
| Edit | `hooks/use-group-sse.ts`, `hooks/use-group-notify.ts` |
| Edit | `features/web/ai-custom/helpers/stream-progress.ts`, `generate-api.ts`, `format-api.ts` |
| Edit | `features/web/play-single/components/game-play-shell.tsx` |
| Edit | `features/web/play-group/components/group-play-shell.tsx` |
| Edit | `features/com/images/hooks/use-image-uploader.ts` |
| Edit | `features/web/auth/hooks/use-signin.ts`, `use-signup.ts` |
| Edit | `features/web/auth/components/user-profile-menu.tsx` |
| Edit | `components/in/landing-header.tsx` |
| Edit | `app/(web)/hall/layout.tsx` |

### Deploy — 3 files

| Action | Files |
|---|---|
| Edit | `deploy/env/.env.dev`, `deploy/env/.env.prod`, `deploy/env/.env.example` |
