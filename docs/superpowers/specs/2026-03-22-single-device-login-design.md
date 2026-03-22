# Single-Device Login Design

> Enforce one active session per user account. Logging in on a new device instantly kicks out the previous device.

## Overview

| Item | Detail |
|------|--------|
| **Goal** | Each user/admin account can only be signed in on one device at a time |
| **Mechanism** | `auth_id` claim in JWT + Redis lookup per request for instant kick-out |
| **JWT generation** | Custom JWT via `golang-jwt/jwt/v5` (replaces Goravel's `Login()`) |
| **Goravel compatibility** | JWT includes `key` claim so `facades.Auth().Guard().ID()` still works |
| **Scope** | Both user auth (`/api/*`) and admin auth (`/adm/*`) |
| **Kick-out UX** | Chinese message: "您的账号已在其他设备登录" |

### Goravel Compatibility Note

The custom JWT includes an `auth_id` claim that Goravel's `Claims` struct does not define. Go's `encoding/json` silently ignores unknown fields during unmarshaling, so Goravel's `Parse()` and `ID()` continue to work. This is a stable Go language behavior, but should be regression-tested if Goravel's JWT implementation changes.

### Simultaneous Login Edge Case

If the same user triggers two logins simultaneously (e.g., Device B and Device C at the same millisecond), the last write to `user_auth:{userId}:{guard}` wins. The earlier device may receive a valid-looking response but will be kicked on the next request. This is acceptable — the behavior is correct (single device wins) even if the UX is briefly inconsistent.

## How It Works

### Login (New Device)

```
User signs in on Device B
        ↓
dx-api generates new auth_id (ULID)
        ↓
Redis: SET user_auth:{userId}:{guard} → new_auth_id (TTL = 7 days)
        ↓
Delete all old refresh tokens for this user
        ↓
Issue JWT with claims: { key: userId, auth_id: new_auth_id, exp: ... }
Store new refresh token with auth_id in its data
        ↓
Device A's session is now invalid (auth_id mismatch)
```

### Normal Request (Per-Request Check)

```
Device A sends request with JWT
        ↓
Middleware: parse JWT → extract userId + auth_id
        ↓
Redis: GET user_auth:{userId}:{guard}
        ↓
Compare JWT auth_id vs Redis auth_id
        ↓
Match → proceed normally
Mismatch → 401 with code 40104 ("您的账号已在其他设备登录")
```

### Refresh (Also Checked)

```
Device A tries to refresh
        ↓
Look up refresh token → get auth_id from token data
        ↓
Redis: GET user_auth:{userId}:{guard}
        ↓
Compare refresh token auth_id vs Redis auth_id
        ↓
Match → issue new JWT (same auth_id), rotate refresh token
Mismatch → 401 with code 40104
```

## Backend Changes

### New File: `app/helpers/jwt.go`

Custom JWT generation using `golang-jwt/jwt/v5`:

```go
type Claims struct {
    jwt.RegisteredClaims
    Key    string `json:"key"`     // user ID (Goravel compatible)
    AuthID string `json:"auth_id"` // session identifier
}

func IssueAccessToken(userID, authID string) (string, error)

// ExtractAuthID parses auth_id from a Bearer token without re-verifying signature.
// Only call AFTER Goravel's Parse() has validated the token.
func ExtractAuthID(bearerToken string) string
```

**`IssueAccessToken`:**
- Signs with `JWT_SECRET` (same secret Goravel uses)
- Sets `exp` to `now + JWT_TTL` minutes
- Sets `key` = user ID (Goravel's `Auth().Guard().ID()` reads this)
- Sets `auth_id` = unique session ID

Replaces all `facades.Auth(ctx).Guard("...").Login(&user)` calls in auth services.

**`ExtractAuthID`:**
- Strips `Bearer ` prefix
- Base64url-decodes the payload segment (middle part of JWT)
- JSON-unmarshals to extract `auth_id` field
- Returns empty string on any error

```go
func ExtractAuthID(bearerToken string) string {
    token := strings.TrimPrefix(bearerToken, "Bearer ")
    parts := strings.SplitN(token, ".", 3)
    if len(parts) != 3 {
        return ""
    }
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return ""
    }
    var claims struct {
        AuthID string `json:"auth_id"`
    }
    if err := json.Unmarshal(payload, &claims); err != nil {
        return ""
    }
    return claims.AuthID
}
```

This is safe because Goravel's `Parse()` has already verified the JWT signature and expiration before `ExtractAuthID` is called.

### Modified: `app/helpers/refresh_token.go`

Add `AuthID` field to `RefreshTokenData`:

```go
type RefreshTokenData struct {
    UserID string `json:"user_id"`
    Guard  string `json:"guard"`
    AuthID string `json:"auth_id"`
}
```

Update `StoreRefreshToken` signature to accept `authID`.

### New Redis Key

| Key | Value | TTL | Description |
|-----|-------|-----|-------------|
| `user_auth:{userId}:{guard}` | `auth_id` string | 7 days (same as refresh token TTL) | Tracks the current valid session |

TTL matches refresh token lifetime. When the refresh token expires and the user must re-login, this key also expires automatically.

### Modified: `app/services/api/auth_service.go`

**SignUp / SignInByEmail / SignInByAccount:**
1. Generate `auth_id` = `ulid.MustNew(...)`
2. `SET user_auth:{userId}:user` → `auth_id` (TTL = refresh token TTL)
3. `DeleteUserRefreshTokens(userId, "user")` — invalidate all old sessions
4. `helpers.IssueAccessToken(user.ID, authID)` — replaces `facades.Auth(ctx).Guard("user").Login(&user)`
5. `StoreRefreshToken(token, userId, "user", authID)` — include auth_id

**RefreshToken:**
1. Look up refresh token → get `auth_id`
2. `GET user_auth:{userId}:user` from Redis
3. If mismatch → return `ErrSessionReplaced`
4. Issue new JWT with same `auth_id`
5. Rotate refresh token (with same `auth_id`)

**Logout:**
- Look up refresh token → get `auth_id`
- Compare with `GET user_auth:{userId}:user`
- Only delete the `user_auth` key if `auth_id` matches (prevents a stale session's logout from invalidating the current active session)
- Delete refresh token from Redis

```go
func Logout(refreshToken string) error {
    data, err := helpers.LookupRefreshToken(refreshToken)
    if err != nil {
        return nil // token already gone
    }
    if data.Guard != "user" {
        return nil
    }
    // Only clear session key if this is the active session
    key := fmt.Sprintf("user_auth:%s:%s", data.UserID, data.Guard)
    currentAuthID, _ := helpers.RedisGet(key)
    if currentAuthID == data.AuthID {
        _ = helpers.RedisDel(key)
    }
    return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
}
```

### Modified: `app/services/adm/auth_service.go`

Same changes as user auth, using guard `"admin"`.

### Modified: `app/http/middleware/jwt_auth.go`

After Goravel's `Parse()` validates the JWT:
1. Get user ID via `facades.Auth(ctx).Guard("user").ID()`
2. Extract `auth_id` via `helpers.ExtractAuthID(token)`
3. `GET user_auth:{userId}:user` from Redis
4. If `auth_id` != Redis value → return 401 with `CodeSessionReplaced`

```go
// After Parse() succeeds and user ID is obtained:
token := ctx.Request().Header("Authorization", "")
authID := helpers.ExtractAuthID(token)
currentAuthID, err := helpers.RedisGet(fmt.Sprintf("user_auth:%s:user", userID))
if err != nil || authID == "" || currentAuthID != authID {
    _ = ctx.Response().Json(401, helpers.Response{
        Code:    consts.CodeSessionReplaced,
        Message: "您的账号已在其他设备登录",
    }).Abort()
    return
}
```

### Modified: `app/http/middleware/adm_jwt_auth.go`

Same check using guard `"admin"`.

### Modified: `app/consts/error_code.go`

```go
CodeSessionReplaced = 40104
```

### Modified: `app/services/api/errors.go` and `app/services/adm/errors.go`

```go
ErrSessionReplaced = errors.New("session replaced by another device")
```

## Frontend Changes

### Modified: `dx-web/src/lib/api-client.ts`

**In `apiFetch` 401 handler** — read response body first, check for kick-out before attempting refresh:

```typescript
if (res.status === 401) {
    const errorData: ApiResponse<null> = await res.json();

    // Kicked out by another device — show message, don't attempt refresh
    if (errorData.code === 40104) {
        clearAccessToken();
        if (typeof window !== "undefined") {
            alert("您的账号已在其他设备登录");
            window.location.href = "/auth/signin";
        }
        throw new Error("Session replaced");
    }

    // Normal 401 — attempt refresh and retry
    try {
        const newToken = await refreshAccessToken();
        // ... retry with new token
    } catch {
        clearAccessToken();
        if (typeof window !== "undefined") {
            window.location.href = "/auth/signin";
        }
        throw new Error("Unauthorized");
    }
}
```

**In `refreshAccessToken`** — detect `40104` during refresh:

```typescript
.then((res) => {
    if (!res.ok) {
        return res.json().then((data: ApiResponse<null>) => {
            if (data.code === 40104) {
                clearAccessToken();
                if (typeof window !== "undefined") {
                    alert("您的账号已在其他设备登录");
                    window.location.href = "/auth/signin";
                }
            }
            throw new Error("refresh failed");
        });
    }
    return res.json();
})
```

**In `uploadApi.uploadImage`** — same `40104` check in its independent 401 handler.

### No SSR Changes Needed

All authenticated requests are client-side only (`apiClient`). The deleted `api-server.ts` is no longer in the codebase. `proxy.ts` only checks cookie presence, not token validity.

## Cost Analysis

**Per-request overhead:** One Redis `GET` (~0.5ms) in the middleware for every authenticated request. Negligible compared to typical DB query times (5-50ms).

**Memory overhead:** One Redis key per active user (`user_auth:{userId}:{guard}`). Auto-expires with 7-day TTL. Minimal.
