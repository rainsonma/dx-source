# Single-Device Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce one active session per user account — logging in on a new device instantly kicks out the previous device.

**Architecture:** Embed a unique `auth_id` in every JWT using `golang-jwt/jwt/v5` directly (replacing Goravel's `Login()`). Store the current valid `auth_id` per user in Redis. Middleware checks `auth_id` on every request — mismatch means the session was replaced. Frontend detects error code `40104` and shows a Chinese kick-out message.

**Tech Stack:** Go/Goravel, `golang-jwt/jwt/v5`, Redis, Next.js 16, TypeScript

**Spec:** `docs/superpowers/specs/2026-03-22-single-device-login-design.md`

---

## File Map

### New Files

| File | Responsibility |
|------|---------------|
| `dx-api/app/helpers/jwt.go` | Custom JWT generation with `auth_id` claim + `ExtractAuthID` parser |

### Modified Files

| File | Change |
|------|--------|
| `dx-api/app/helpers/refresh_token.go` | Add `AuthID` field to `RefreshTokenData`, update `StoreRefreshToken` signature |
| `dx-api/app/services/api/auth_service.go` | Generate `auth_id` on login, use `IssueAccessToken`, check `auth_id` on refresh, conditional logout |
| `dx-api/app/services/adm/auth_service.go` | Same changes for admin |
| `dx-api/app/services/api/errors.go` | Add `ErrSessionReplaced` |
| `dx-api/app/services/adm/errors.go` | Add `ErrSessionReplaced` |
| `dx-api/app/http/middleware/jwt_auth.go` | Add per-request `auth_id` check against Redis |
| `dx-api/app/http/middleware/adm_jwt_auth.go` | Same for admin |
| `dx-api/app/consts/error_code.go` | Add `CodeSessionReplaced = 40104` |
| `dx-web/src/lib/api-client.ts` | Detect `40104` in 401 handler, show Chinese message |

---

## Task 1: Error Codes & JWT Helper

**Files:**
- Create: `dx-api/app/helpers/jwt.go`
- Modify: `dx-api/app/consts/error_code.go`
- Modify: `dx-api/app/services/api/errors.go`
- Modify: `dx-api/app/services/adm/errors.go`

- [ ] **Step 1: Add error code**

In `dx-api/app/consts/error_code.go`, add after `CodeInvalidRefreshToken = 40103`:

```go
CodeSessionReplaced = 40104
```

- [ ] **Step 2: Add error sentinels**

In `dx-api/app/services/api/errors.go`, add:

```go
ErrSessionReplaced = errors.New("session replaced by another device")
```

In `dx-api/app/services/adm/errors.go`, add:

```go
ErrSessionReplaced = errors.New("session replaced by another device")
```

- [ ] **Step 3: Create JWT helper**

Create `dx-api/app/helpers/jwt.go`:

```go
package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goravel/framework/facades"
)

// Claims defines the JWT payload. The Key field is read by Goravel's
// Auth().Guard().ID() for backward compatibility.
type Claims struct {
	jwt.RegisteredClaims
	Key    string `json:"key"`     // user ID (Goravel compatible)
	AuthID string `json:"auth_id"` // session identifier for single-device enforcement
}

// IssueAccessToken creates a signed JWT with the user ID and auth_id claims.
func IssueAccessToken(userID, authID string) (string, error) {
	secret := facades.Config().GetString("jwt.secret", "")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not configured")
	}

	ttl := facades.Config().GetInt("jwt.ttl", 10)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Key:    userID,
		AuthID: authID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ExtractAuthID reads the auth_id claim from a Bearer token without re-verifying
// the signature. Only call AFTER Goravel's Parse() has validated the token.
func ExtractAuthID(bearerToken string) string {
	raw := strings.TrimPrefix(bearerToken, "Bearer ")
	parts := strings.SplitN(raw, ".", 3)
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

- [ ] **Step 4: Add `golang-jwt/jwt/v5` as direct dependency**

Run: `cd dx-api && go get github.com/golang-jwt/jwt/v5`

This promotes the existing indirect dependency to direct.

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/helpers/jwt.go dx-api/app/consts/error_code.go dx-api/app/services/api/errors.go dx-api/app/services/adm/errors.go dx-api/go.mod dx-api/go.sum
git commit -m "feat: add custom JWT helper with auth_id claim, add session-replaced error code"
```

---

## Task 2: Refresh Token — Add `AuthID` Field

**Files:**
- Modify: `dx-api/app/helpers/refresh_token.go`

- [ ] **Step 1: Add `AuthID` to `RefreshTokenData`**

Change the struct from:

```go
type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Guard  string `json:"guard"`
}
```

To:

```go
type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Guard  string `json:"guard"`
	AuthID string `json:"auth_id"`
}
```

- [ ] **Step 2: Update `StoreRefreshToken` signature**

Change from:

```go
func StoreRefreshToken(token, userID, guard string) error {
```

To:

```go
func StoreRefreshToken(token, userID, guard, authID string) error {
```

And update the marshal call:

```go
data, err := json.Marshal(RefreshTokenData{UserID: userID, Guard: guard, AuthID: authID})
```

- [ ] **Step 3: Do NOT commit yet** — callers of `StoreRefreshToken` need updating (Task 3). Build will fail.

---

## Task 3: User Auth Service — Single Device (commit with Task 2)

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go`

- [ ] **Step 1: Add `issueSession` helper function**

Add a private helper at the bottom of `auth_service.go` to consolidate the login flow (replaces duplicated code in SignUp, SignInByEmail, SignInByAccount):

```go
// issueSession generates auth_id, invalidates old sessions, issues tokens.
func issueSession(userID string) (*AuthResult, error) {
	authID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	// Set current auth_id in Redis (invalidates previous device instantly)
	ttl := time.Duration(facades.Config().GetInt("refresh_token.ttl", 10080)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:user", userID), authID, ttl); err != nil {
		return nil, fmt.Errorf("failed to store auth_id: %w", err)
	}

	// Delete all old refresh tokens
	_ = helpers.DeleteUserRefreshTokens(userID, "user")

	// Issue JWT with auth_id
	accessToken, err := helpers.IssueAccessToken(userID, authID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	// Generate and store refresh token with auth_id
	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, userID, "user", authID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
```

- [ ] **Step 2: Update SignUp**

Replace lines 100-113 (the `facades.Auth(ctx).Guard("user").Login(&user)` block through the return) with:

```go
result, err := issueSession(user.ID)
if err != nil {
	return nil, nil, err
}
return result, &user, nil
```

Remove `ctx contractshttp.Context` from SignUp's parameters since Goravel's Login is no longer used. **Wait** — check if `ctx` is used elsewhere in SignUp. It's not (only used for `facades.Auth(ctx)`). But changing the function signature would break the controller. Keep `ctx` in the signature for now — it's not harmful and keeps the interface stable.

Actually, simpler: just replace the token-issuing block, keep the signature unchanged.

- [ ] **Step 3: Update SignInByEmail**

Replace lines 185-198 (the Login + refresh token block) with:

```go
result, err := issueSession(user.ID)
if err != nil {
	return nil, nil, err
}
return result, &user, nil
```

- [ ] **Step 4: Update SignInByAccount**

Replace lines 218-231 (the Login + refresh token block) with:

```go
result, err := issueSession(user.ID)
if err != nil {
	return nil, nil, err
}
return result, &user, nil
```

- [ ] **Step 5: Update RefreshToken**

After looking up the refresh token data (line 237), add the `auth_id` check before issuing a new token:

```go
data, err := helpers.LookupRefreshToken(oldRefreshToken)
if err != nil {
	return nil, ErrInvalidRefreshToken
}
if data.Guard != "user" {
	return nil, ErrInvalidRefreshToken
}

// Check if this session is still the active one
currentAuthID, err := helpers.RedisGet(fmt.Sprintf("user_auth:%s:user", data.UserID))
if err != nil || currentAuthID != data.AuthID {
	return nil, ErrSessionReplaced
}

var user models.User
if err := facades.Orm().Query().Where("id", data.UserID).First(&user); err != nil || user.ID == "" {
	return nil, ErrUserNotFound
}

// Issue new JWT with same auth_id
accessToken, err := helpers.IssueAccessToken(data.UserID, data.AuthID)
if err != nil {
	return nil, fmt.Errorf("failed to issue access token: %w", err)
}

// Rotate refresh token (keep same auth_id)
_ = helpers.DeleteRefreshToken(oldRefreshToken, data.UserID, "user")

newRefreshToken, err := helpers.GenerateRefreshToken()
if err != nil {
	return nil, fmt.Errorf("failed to generate refresh token: %w", err)
}
if err := helpers.StoreRefreshToken(newRefreshToken, data.UserID, "user", data.AuthID); err != nil {
	return nil, fmt.Errorf("failed to store refresh token: %w", err)
}

return &AuthResult{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
```

- [ ] **Step 6: Update Logout**

Replace the current `Logout` function with:

```go
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		return nil
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

- [ ] **Step 7: Remove unused import**

The `facades` import may no longer be needed if all `facades.Auth(ctx).Guard("user").Login()` calls are removed and replaced with `helpers.IssueAccessToken()`. Check: `facades.Orm()` is still used in `RefreshToken` and `GetCurrentUser`, so `facades` stays. But `contractshttp` may no longer be needed by `issueSession` — check usage. Actually `ctx` is still in the function signatures of SignUp etc., so `contractshttp` stays too.

The `facades.Config()` is used by `issueSession` via helpers, not directly. Check if `facades` is still imported by other functions in the file. Yes — `facades.Orm()` in RefreshToken, GetCurrentUser. Keep it.

- [ ] **Step 8: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 9: Update user auth controller to handle `ErrSessionReplaced`**

In `dx-api/app/http/controllers/api/auth_controller.go`, in the `Refresh` method, add a new error branch after the existing `ErrInvalidRefreshToken` check:

```go
if errors.Is(err, services.ErrSessionReplaced) {
	clearRefreshCookie(ctx)
	return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeSessionReplaced, "您的账号已在其他设备登录")
}
```

- [ ] **Step 10: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 11: Commit (Tasks 2 + 3 together)**

```bash
git add dx-api/app/helpers/refresh_token.go dx-api/app/services/api/auth_service.go dx-api/app/http/controllers/api/auth_controller.go
git commit -m "feat: user auth enforces single-device login with auth_id"
```

---

## Task 4: Admin Auth Service — Single Device

**Files:**
- Modify: `dx-api/app/services/adm/auth_service.go`

- [ ] **Step 1: Add `issueAdminSession` helper**

Same pattern as user, using guard `"admin"`:

```go
func issueAdminSession(userID string) (*AuthResult, error) {
	authID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	ttl := time.Duration(facades.Config().GetInt("refresh_token.ttl", 10080)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:admin", userID), authID, ttl); err != nil {
		return nil, fmt.Errorf("failed to store auth_id: %w", err)
	}

	_ = helpers.DeleteUserRefreshTokens(userID, "admin")

	accessToken, err := helpers.IssueAccessToken(userID, authID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(refreshToken, userID, "admin", authID); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
```

- [ ] **Step 2: Update AdminSignIn**

Replace lines 39-56 (the `facades.Auth(ctx).Guard("admin").Login()` block through the return) with:

```go
result, err := issueAdminSession(admUser.ID)
if err != nil {
	return nil, nil, err
}

ip := ctx.Request().Ip()
userAgent := ctx.Request().Header("User-Agent", "")
go RecordAdminLogin(admUser.ID, ip, userAgent)

return result, &admUser, nil
```

- [ ] **Step 3: Update RefreshToken**

Add `auth_id` check (same pattern as user). After lookup:

```go
currentAuthID, err := helpers.RedisGet(fmt.Sprintf("user_auth:%s:admin", data.UserID))
if err != nil || currentAuthID != data.AuthID {
	return nil, ErrSessionReplaced
}
```

Replace `facades.Auth(ctx).Guard("admin").Login(&admUser)` with `helpers.IssueAccessToken(data.UserID, data.AuthID)`.

Update `StoreRefreshToken` call to include `data.AuthID`.

- [ ] **Step 4: Update Logout**

Same conditional delete pattern as user:

```go
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	if data.Guard != "admin" {
		return nil
	}
	key := fmt.Sprintf("user_auth:%s:%s", data.UserID, data.Guard)
	currentAuthID, _ := helpers.RedisGet(key)
	if currentAuthID == data.AuthID {
		_ = helpers.RedisDel(key)
	}
	return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
}
```

- [ ] **Step 5: Update admin auth controller to handle `ErrSessionReplaced`**

In `dx-api/app/http/controllers/adm/auth_controller.go`, in the `Refresh` method, add a new error branch after the existing `ErrInvalidRefreshToken` check:

```go
if errors.Is(err, admservice.ErrSessionReplaced) {
	clearAdmRefreshCookie(ctx)
	return helpers.Error(ctx, 401, consts.CodeSessionReplaced, "您的账号已在其他设备登录")
}
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/services/adm/auth_service.go dx-api/app/http/controllers/adm/auth_controller.go
git commit -m "feat: admin auth enforces single-device login with auth_id"
```

---

## Task 5: Middleware — Per-Request `auth_id` Check

**Files:**
- Modify: `dx-api/app/http/middleware/jwt_auth.go`
- Modify: `dx-api/app/http/middleware/adm_jwt_auth.go`

- [ ] **Step 1: Update JwtAuth middleware**

Replace the entire `jwt_auth.go` with:

```go
package middleware

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func JwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		payload, err := facades.Auth(ctx).Guard("user").Parse(token)
		if err != nil || payload == nil {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		// Single-device check: verify auth_id matches current session
		userID, _ := facades.Auth(ctx).Guard("user").ID()
		authID := helpers.ExtractAuthID(token)
		currentAuthID, redisErr := helpers.RedisGet(fmt.Sprintf("user_auth:%s:user", userID))
		if redisErr != nil || authID == "" || currentAuthID != authID {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}
```

- [ ] **Step 2: Update AdmJwtAuth middleware**

Replace the entire `adm_jwt_auth.go` with the same pattern but using `"admin"` guard:

```go
package middleware

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func AdmJwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		payload, err := facades.Auth(ctx).Guard("admin").Parse(token)
		if err != nil || payload == nil {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		// Single-device check
		userID, _ := facades.Auth(ctx).Guard("admin").ID()
		authID := helpers.ExtractAuthID(token)
		currentAuthID, redisErr := helpers.RedisGet(fmt.Sprintf("user_auth:%s:admin", userID))
		if redisErr != nil || authID == "" || currentAuthID != authID {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/middleware/jwt_auth.go dx-api/app/http/middleware/adm_jwt_auth.go
git commit -m "feat: middleware checks auth_id per request for single-device enforcement"
```

---

## Task 6: Frontend — Detect Kick-Out

**Files:**
- Modify: `dx-web/src/lib/api-client.ts`

- [ ] **Step 1: Update `apiFetch` 401 handler**

The current 401 handler (lines 83-112) immediately attempts refresh. Change it to read the response body first and check for `40104`:

```typescript
if (res.status === 401) {
    // Read response body to check error code
    const errorData: ApiResponse<null> = await res.clone().json().catch(() => ({ code: 0, message: "", data: null }));

    // Kicked out by another device — don't attempt refresh
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
        // ... existing retry logic
    }
}
```

Note: use `res.clone().json()` so the original response body is still available if needed. Wrap in `.catch()` in case the response body is not valid JSON.

- [ ] **Step 2: Update `refreshAccessToken` to detect `40104`**

Change the error handling in `refreshAccessToken` (lines 38-59). Replace:

```typescript
.then((res) => {
    if (!res.ok) throw new Error("refresh failed");
    return res.json();
})
```

With:

```typescript
.then(async (res) => {
    if (!res.ok) {
        const errorData = await res.json().catch(() => ({ code: 0 }));
        if (errorData.code === 40104) {
            clearAccessToken();
            if (typeof window !== "undefined") {
                alert("您的账号已在其他设备登录");
                window.location.href = "/auth/signin";
            }
        }
        throw new Error("refresh failed");
    }
    return res.json();
})
```

- [ ] **Step 3: Update `uploadApi.uploadImage` 401 handler**

Find the upload function's 401 handling and add the same `40104` check before attempting refresh. Read the response body first:

```typescript
if (res.status === 401) {
    const errorData = await res.clone().json().catch(() => ({ code: 0 }));
    if (errorData.code === 40104) {
        clearAccessToken();
        if (typeof window !== "undefined") {
            alert("您的账号已在其他设备登录");
            window.location.href = "/auth/signin";
        }
        throw new Error("Session replaced");
    }
    // ... existing refresh retry logic
}
```

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npm run build`
Expected: Build passes.

- [ ] **Step 5: Commit**

```bash
cd dx-web && git add src/lib/api-client.ts && git commit -m "feat: detect session-replaced error, show Chinese kick-out message"
```

---

## Task 7: Final Verification & Cleanup

- [ ] **Step 1: Full backend build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Clean pass.

- [ ] **Step 2: Full frontend build**

Run: `cd dx-web && npm run build`
Expected: Clean pass.

- [ ] **Step 3: Verify no remaining Goravel Login calls in auth services**

Run: `grep -n "Guard.*Login" dx-api/app/services/api/auth_service.go dx-api/app/services/adm/auth_service.go`
Expected: No matches — all replaced with `helpers.IssueAccessToken()`.

- [ ] **Step 4: Update parent repo submodule pointer**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web
git commit -m "feat: update dx-web submodule for single-device login"
```

- [ ] **Step 5: Update auth design doc**

Add the single-device login section to `docs/dx-auth-design.md`:
- New Redis key: `user_auth:{userId}:{guard}`
- Middleware per-request check
- Error code `40104`
- Kick-out UX

Commit:
```bash
git add docs/dx-auth-design.md
git commit -m "docs: add single-device login to auth design doc"
```
