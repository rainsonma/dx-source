# Dual-Token Authentication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace single JWT auth with access token (JWT, 10 min) + refresh token (opaque, 7 days, Redis) for both user and admin auth.

**Architecture:** The Go backend issues both tokens on login. Access token is short-lived JWT validated by Goravel's built-in guard. Refresh token is an opaque hex string stored in Redis, delivered via httpOnly cookie + response body (multi-client). The frontend stores the access token in memory only, uses a refresh lock to handle concurrent 401s, and retries failed requests after refresh.

**Tech Stack:** Go/Goravel, Redis, Next.js 16, TypeScript

**Spec:** `docs/superpowers/specs/2026-03-22-dual-token-auth-design.md`

---

## File Map

### Backend — New Files

| File | Responsibility |
|------|---------------|
| `dx-api/app/helpers/refresh_token.go` | Refresh token generation, Redis CRUD, user index management |

### Backend — Modified Files

| File | Change |
|------|--------|
| `dx-api/config/jwt.go` | TTL 60→10, remove refresh_ttl, add refresh token config |
| `dx-api/app/services/api/auth_service.go` | Return refresh token from login/signup, rewrite RefreshToken/Logout |
| `dx-api/app/services/api/errors.go` | Add `ErrInvalidRefreshToken` sentinel |
| `dx-api/app/http/controllers/api/auth_controller.go` | Cookie handling, dual-token responses |
| `dx-api/routes/api.go` | Move refresh/logout to public, add rate limit |
| `dx-api/app/services/adm/auth_service.go` | Same changes as user auth service |
| `dx-api/app/services/adm/errors.go` | Add `ErrInvalidRefreshToken` sentinel |
| `dx-api/app/http/controllers/adm/auth_controller.go` | Cookie handling, dual-token responses |
| `dx-api/routes/adm.go` | Move logout to public, add refresh endpoint |
| `dx-api/app/consts/error_code.go` | Add `CodeInvalidRefreshToken` |
| `deploy/env/.env.dev` | Add REFRESH_TOKEN_TTL, REFRESH_COOKIE_SECURE |
| `deploy/env/.env.example` | Add REFRESH_TOKEN_TTL, REFRESH_COOKIE_SECURE |

### Frontend — New Files

| File | Responsibility |
|------|---------------|
| `dx-web/src/lib/token.ts` | In-memory access token store |
| `dx-web/src/components/in/auth-guard.tsx` | Client-side auth guard with silent refresh |

### Frontend — Modified Files

| File | Change |
|------|--------|
| `dx-web/src/lib/api-client.ts` | Use memory token, add refresh-on-401 with lock |
| `dx-web/src/features/web/auth/hooks/use-signin.ts` | Use `setAccessToken`, add `credentials: "include"` |
| `dx-web/src/features/web/auth/hooks/use-signup.ts` | Auto-login, redirect to `/hall` |
| `dx-web/src/proxy.ts` | Replace `dx_token` cookie check with `dx_refresh` cookie check |

### Frontend — Files to Remove

| File | Reason |
|------|--------|
| `dx-web/src/lib/auth.ts` | Server-side JWT decode, replaced by client-side auth guard |
| `dx-web/src/lib/api-server.ts` | Server-side authenticated fetch, no longer needed |

### Frontend — Files to Migrate (34 files using `apiServerFetch`)

These files will be migrated in Task 7 (batch conversion). Each needs to switch from `apiServerFetch` to `apiClient` and convert server components to client components where needed.

---

## Task 1: JWT Config & Error Codes

**Files:**
- Modify: `dx-api/config/jwt.go:7-41`
- Modify: `dx-api/app/consts/error_code.go:19-22`
- Modify: `dx-api/app/services/api/errors.go`
- Modify: `deploy/env/.env.dev`
- Modify: `deploy/env/.env.example`

- [ ] **Step 1: Update JWT config**

Replace `dx-api/config/jwt.go` contents:

```go
package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("jwt", map[string]any{
		"secret": config.Env("JWT_SECRET", ""),
		"ttl":    config.Env("JWT_TTL", 10),
	})
	config.Add("refresh_token", map[string]any{
		"ttl":           config.Env("REFRESH_TOKEN_TTL", 10080),
		"cookie_secure": config.Env("REFRESH_COOKIE_SECURE", true),
	})
}
```

- [ ] **Step 2: Add error code for invalid refresh token**

In `dx-api/app/consts/error_code.go`, add after `CodeInvalidToken = 40102`:

```go
CodeInvalidRefreshToken = 40103
```

- [ ] **Step 3: Add error sentinel for invalid refresh token**

In `dx-api/app/services/api/errors.go`, add:

```go
ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
```

- [ ] **Step 4: Add admin error sentinel**

Create `dx-api/app/services/adm/errors.go` if it doesn't exist, or add to it:

```go
ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
```

- [ ] **Step 5: Update env files**

In `deploy/env/.env.dev`, add after `JWT_SECRET=...` line:

```
JWT_TTL=10
REFRESH_TOKEN_TTL=10080
REFRESH_COOKIE_SECURE=false
```

In `deploy/env/.env.example`, add the same vars with production defaults:

```
JWT_TTL=10
REFRESH_TOKEN_TTL=10080
REFRESH_COOKIE_SECURE=true
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 7: Commit**

```bash
git add dx-api/config/jwt.go dx-api/app/consts/error_code.go dx-api/app/services/api/errors.go dx-api/app/services/adm/errors.go deploy/env/.env.dev deploy/env/.env.example
git commit -m "chore: update JWT TTL to 10 min, add refresh token config and error codes"
```

---

## Task 2: Refresh Token Helper

**Files:**
- Create: `dx-api/app/helpers/refresh_token.go`

- [ ] **Step 1: Create refresh token helper**

Create `dx-api/app/helpers/refresh_token.go`:

```go
package helpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
)

type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Guard  string `json:"guard"`
}

// GenerateRefreshToken returns a cryptographically random 64-char hex string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// StoreRefreshToken stores a refresh token in Redis with TTL and adds it to the user index.
func StoreRefreshToken(token, userID, guard string) error {
	ctx := context.Background()
	rdb := GetRedis()
	ttl := time.Duration(facades.Config().GetInt("refresh_token.ttl", 10080)) * time.Minute

	data, err := json.Marshal(RefreshTokenData{UserID: userID, Guard: guard})
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	pipe := rdb.Pipeline()
	pipe.Set(ctx, "refresh:"+token, string(data), ttl)
	pipe.SAdd(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), token)
	pipe.Expire(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

// LookupRefreshToken retrieves and validates a refresh token from Redis.
func LookupRefreshToken(token string) (*RefreshTokenData, error) {
	ctx := context.Background()
	val, err := GetRedis().Get(ctx, "refresh:"+token).Result()
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse refresh token data: %w", err)
	}
	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis and the user index.
func DeleteRefreshToken(token, userID, guard string) error {
	ctx := context.Background()
	rdb := GetRedis()

	pipe := rdb.Pipeline()
	pipe.Del(ctx, "refresh:"+token)
	pipe.SRem(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), token)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a user+guard combo.
func DeleteUserRefreshTokens(userID, guard string) error {
	ctx := context.Background()
	rdb := GetRedis()
	indexKey := fmt.Sprintf("user_refresh:%s:%s", userID, guard)

	tokens, err := rdb.SMembers(ctx, indexKey).Result()
	if err != nil {
		return fmt.Errorf("failed to list user refresh tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := rdb.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, "refresh:"+t)
	}
	pipe.Del(ctx, indexKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/helpers/refresh_token.go
git commit -m "feat: add refresh token helper with Redis storage and user index"
```

---

## Task 3: User Auth Service — Dual Token

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go:42-100` (SignUp)
- Modify: `dx-api/app/services/api/auth_service.go:126-177` (SignInByEmail)
- Modify: `dx-api/app/services/api/auth_service.go:179-202` (SignInByAccount)
- Modify: `dx-api/app/services/api/auth_service.go:204-211` (RefreshToken)

- [ ] **Step 1: Add AuthResult struct**

At the top of `dx-api/app/services/api/auth_service.go`, after imports, add:

```go
// AuthResult holds the tokens returned after login/signup/refresh.
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
```

- [ ] **Step 2: Update SignUp to return AuthResult**

Change `SignUp` signature from returning `(string, *models.User, error)` to `(*AuthResult, *models.User, error)`.

After `facades.Auth(ctx).Guard("user").Login(&user)` returns the access token, add:

```go
refreshToken, err := helpers.GenerateRefreshToken()
if err != nil {
	return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
}
if err := helpers.StoreRefreshToken(refreshToken, user.ID, "user"); err != nil {
	return nil, nil, fmt.Errorf("failed to store refresh token: %w", err)
}

return &AuthResult{AccessToken: token, RefreshToken: refreshToken}, &user, nil
```

- [ ] **Step 3: Update SignInByEmail to return AuthResult**

Change `SignInByEmail` signature from `(string, *models.User, error)` to `(*AuthResult, *models.User, error)`.

Same pattern as SignUp — after `Login(&user)`, generate and store refresh token, return `&AuthResult{...}`.

- [ ] **Step 4: Update SignInByAccount to return AuthResult**

Change `SignInByAccount` signature from `(string, *models.User, error)` to `(*AuthResult, *models.User, error)`.

Same pattern — generate and store refresh token after Login.

- [ ] **Step 5: Rewrite RefreshToken**

Replace the current `RefreshToken` function with:

```go
// RefreshToken validates an opaque refresh token, issues a new JWT access token,
// and rotates the refresh token.
func RefreshToken(ctx contractshttp.Context, oldRefreshToken string) (*AuthResult, error) {
	data, err := helpers.LookupRefreshToken(oldRefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if data.Guard != "user" {
		return nil, ErrInvalidRefreshToken
	}

	// Load user to issue new JWT
	var user models.User
	if err := facades.Orm().Query().Where("id", data.UserID).First(&user); err != nil || user.ID == "" {
		return nil, ErrUserNotFound
	}

	accessToken, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	// Rotate refresh token
	_ = helpers.DeleteRefreshToken(oldRefreshToken, data.UserID, "user")

	newRefreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(newRefreshToken, data.UserID, "user"); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}
```

- [ ] **Step 6: Rewrite Logout**

Replace the unused `Logout` (currently in controller) with a service-level function:

```go
// Logout deletes the given refresh token from Redis.
// Validates the guard to prevent cross-guard token deletion.
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		// Token already gone — treat as success
		return nil
	}
	if data.Guard != "user" {
		return nil // Not a user token — ignore silently
	}
	return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
}
```

- [ ] **Step 7: Do NOT commit yet** — the controller still references old signatures. Continue to Task 4.

---

## Task 4: User Auth Controller — Cookie Handling (commit together with Task 3)

**Files:**
- Modify: `dx-api/app/http/controllers/api/auth_controller.go`

- [ ] **Step 1: Add cookie helper to controller**

Add a helper method at the bottom of `auth_controller.go`:

```go
// setRefreshCookie sets the refresh token as an httpOnly cookie.
func setRefreshCookie(ctx contractshttp.Context, token string) {
	secure := facades.Config().GetBool("refresh_token.cookie_secure", true)
	ttl := facades.Config().GetInt("refresh_token.ttl", 10080)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_refresh",
		Value:    token,
		Path:     "/api/auth",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

// clearRefreshCookie clears the refresh token cookie.
func clearRefreshCookie(ctx contractshttp.Context) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_refresh",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("refresh_token.cookie_secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

// getRefreshToken reads refresh token from cookie first, then request body.
func getRefreshToken(ctx contractshttp.Context) string {
	if token := ctx.Request().Cookie("dx_refresh"); token != "" {
		return token
	}
	return ctx.Request().Input("refresh_token")
}

// Import note: add "fmt" to the import block for rate limiting in Refresh handler.
```

- [ ] **Step 2: Update SignUp handler**

Update the `SignUp` method to use `AuthResult`:

```go
func (c *AuthController) SignUp(ctx contractshttp.Context) contractshttp.Response {
	// ... validation unchanged ...

	result, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
	if err != nil {
		// ... error handling unchanged ...
	}

	setRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          user,
	})
}
```

- [ ] **Step 3: Update SignIn handler**

Update the `SignIn` method similarly — both `SignInByEmail` and `SignInByAccount` now return `(*AuthResult, *models.User, error)`:

```go
var (
	result *services.AuthResult
	user   *models.User
	err    error
)

if req.Email != "" && req.Code != "" {
	result, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
} else if req.Account != "" && req.Password != "" {
	result, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
} else {
	return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "provide email+code or account+password")
}

// ... error handling unchanged ...

go services.RecordLogin(user.ID, ip, userAgent)

setRefreshCookie(ctx, result.RefreshToken)
return helpers.Success(ctx, map[string]any{
	"access_token":  result.AccessToken,
	"refresh_token": result.RefreshToken,
	"user":          user,
})
```

- [ ] **Step 4: Rewrite Refresh handler**

```go
func (c *AuthController) Refresh(ctx contractshttp.Context) contractshttp.Response {
	oldToken := getRefreshToken(ctx)
	if oldToken == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeInvalidRefreshToken, "refresh token required")
	}

	result, err := services.RefreshToken(ctx, oldToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			clearRefreshCookie(ctx)
			return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeInvalidRefreshToken, "invalid or expired refresh token")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to refresh token")
	}

	setRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}
```

- [ ] **Step 5: Rewrite Logout handler**

```go
func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	token := getRefreshToken(ctx)
	if token != "" {
		_ = services.Logout(token)
	}

	clearRefreshCookie(ctx)
	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 7: Commit (Tasks 3 + 4 together)**

```bash
git add dx-api/app/services/api/auth_service.go dx-api/app/http/controllers/api/auth_controller.go
git commit -m "feat: user auth returns dual tokens with httpOnly cookie and refresh rotation"
```

---

## Task 5: User Routes — Public Refresh & Logout with Rate Limit

**Files:**
- Modify: `dx-api/routes/api.go:55-68`

- [ ] **Step 1: Move refresh and logout to public auth group, add rate limit**

In `dx-api/routes/api.go`, change the auth route groups from:

```go
// Auth routes (public, no JWT required)
router.Prefix("/auth").Group(func(auth route.Router) {
	auth.Post("/signup/send-code", authController.SendSignUpCode)
	auth.Post("/signup", authController.SignUp)
	auth.Post("/signin/send-code", authController.SendSignInCode)
	auth.Post("/signin", authController.SignIn)
})

// Auth routes (protected, JWT required)
router.Prefix("/auth").Middleware(middleware.JwtAuth()).Group(func(auth route.Router) {
	auth.Post("/refresh", authController.Refresh)
	auth.Get("/me", authController.Me)
	auth.Post("/logout", authController.Logout)
})
```

To:

```go
// Auth routes (public, no JWT required)
router.Prefix("/auth").Group(func(auth route.Router) {
	auth.Post("/signup/send-code", authController.SendSignUpCode)
	auth.Post("/signup", authController.SignUp)
	auth.Post("/signin/send-code", authController.SendSignInCode)
	auth.Post("/signin", authController.SignIn)
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/logout", authController.Logout)
})

// Auth routes (protected, JWT required)
router.Prefix("/auth").Middleware(middleware.JwtAuth()).Group(func(auth route.Router) {
	auth.Get("/me", authController.Me)
})
```

Note: Rate limiting for `/api/auth/refresh` is handled by the existing `CheckRateLimit` helper — add it inside the `Refresh` controller handler at the start:

- [ ] **Step 2: Add rate limit to Refresh handler**

In `auth_controller.go`, at the beginning of the `Refresh` method (before `getRefreshToken`), add:

```go
ip := ctx.Request().Ip()
allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:refresh:%s", ip), 10, 60)
if err != nil || !allowed {
	return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "too many refresh requests")
}
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 4: Commit**

```bash
git add dx-api/routes/api.go dx-api/app/http/controllers/api/auth_controller.go
git commit -m "feat: move refresh/logout to public routes, add rate limit on refresh"
```

---

## Task 6: Admin Auth — Same Pattern

**Files:**
- Modify: `dx-api/app/services/adm/auth_service.go`
- Modify: `dx-api/app/http/controllers/adm/auth_controller.go`
- Modify: `dx-api/routes/adm.go`

- [ ] **Step 1: Update admin auth service**

In `dx-api/app/services/adm/auth_service.go`:

Add `AuthResult` struct:

```go
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
```

Change `AdminSignIn` return from `(string, *models.AdmUser, error)` to `(*AuthResult, *models.AdmUser, error)`. After `Login(&admUser)`:

```go
refreshToken, err := helpers.GenerateRefreshToken()
if err != nil {
	return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
}
if err := helpers.StoreRefreshToken(refreshToken, admUser.ID, "admin"); err != nil {
	return nil, nil, fmt.Errorf("failed to store refresh token: %w", err)
}

return &AuthResult{AccessToken: token, RefreshToken: refreshToken}, &admUser, nil
```

Add `RefreshToken`:

```go
func RefreshToken(ctx contractshttp.Context, oldRefreshToken string) (*AuthResult, error) {
	data, err := helpers.LookupRefreshToken(oldRefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if data.Guard != "admin" {
		return nil, ErrInvalidRefreshToken
	}

	var admUser models.AdmUser
	if err := facades.Orm().Query().Where("id", data.UserID).First(&admUser); err != nil || admUser.ID == "" {
		return nil, ErrAdminNotFound
	}

	accessToken, err := facades.Auth(ctx).Guard("admin").Login(&admUser)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	_ = helpers.DeleteRefreshToken(oldRefreshToken, data.UserID, "admin")

	newRefreshToken, err := helpers.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	if err := helpers.StoreRefreshToken(newRefreshToken, data.UserID, "admin"); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}
```

Add `Logout`:

```go
func Logout(refreshToken string) error {
	data, err := helpers.LookupRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	if data.Guard != "admin" {
		return nil
	}
	return helpers.DeleteRefreshToken(refreshToken, data.UserID, data.Guard)
}
```

- [ ] **Step 2: Update admin auth controller**

In `dx-api/app/http/controllers/adm/auth_controller.go`, add admin-specific cookie helpers:

```go
func setAdmRefreshCookie(ctx contractshttp.Context, token string) {
	secure := facades.Config().GetBool("refresh_token.cookie_secure", true)
	ttl := facades.Config().GetInt("refresh_token.ttl", 10080)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_adm_refresh",
		Value:    token,
		Path:     "/adm/auth",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func clearAdmRefreshCookie(ctx contractshttp.Context) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_adm_refresh",
		Value:    "",
		Path:     "/adm/auth",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("refresh_token.cookie_secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func getAdmRefreshToken(ctx contractshttp.Context) string {
	if token := ctx.Request().Cookie("dx_adm_refresh"); token != "" {
		return token
	}
	return ctx.Request().Input("refresh_token")
}
```

Update `Login` handler to set cookie and return dual tokens.

Add `Refresh` handler with rate limiting:

```go
func (c *AuthController) Refresh(ctx contractshttp.Context) contractshttp.Response {
	ip := ctx.Request().Ip()
	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:adm_refresh:%s", ip), 10, 60)
	if err != nil || !allowed {
		return helpers.Error(ctx, 429, consts.CodeRateLimited, "too many refresh requests")
	}

	oldToken := getAdmRefreshToken(ctx)
	if oldToken == "" {
		return helpers.Error(ctx, 401, consts.CodeInvalidRefreshToken, "refresh token required")
	}

	result, err := admservice.RefreshToken(ctx, oldToken)
	if err != nil {
		if errors.Is(err, admservice.ErrInvalidRefreshToken) {
			clearAdmRefreshCookie(ctx)
			return helpers.Error(ctx, 401, consts.CodeInvalidRefreshToken, "invalid or expired refresh token")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "failed to refresh token")
	}

	setAdmRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}
```

Update `Logout` handler to read token from cookie/body, call service, clear cookie.

- [ ] **Step 3: Update admin routes**

In `dx-api/routes/adm.go`, change from:

```go
// Admin auth routes (public)
router.Prefix("/auth").Group(func(auth route.Router) {
	auth.Post("/login", admAuthController.Login)
})

// Admin auth routes (protected)
router.Prefix("/auth").Middleware(middleware.AdmJwtAuth()).Group(func(auth route.Router) {
	auth.Get("/me", admAuthController.Me)
	auth.Post("/logout", admAuthController.Logout)
})
```

To:

```go
// Admin auth routes (public)
router.Prefix("/auth").Group(func(auth route.Router) {
	auth.Post("/login", admAuthController.Login)
	auth.Post("/refresh", admAuthController.Refresh)
	auth.Post("/logout", admAuthController.Logout)
})

// Admin auth routes (protected)
router.Prefix("/auth").Middleware(middleware.AdmJwtAuth()).Group(func(auth route.Router) {
	auth.Get("/me", admAuthController.Me)
})
```

- [ ] **Step 4: Verify build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Build passes.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/services/adm/auth_service.go dx-api/app/http/controllers/adm/auth_controller.go dx-api/routes/adm.go dx-api/app/services/adm/errors.go
git commit -m "feat: admin auth uses dual tokens with httpOnly cookies"
```

---

## Task 7: Frontend — Token Module & API Client Rewrite

**Files:**
- Create: `dx-web/src/lib/token.ts`
- Modify: `dx-web/src/lib/api-client.ts`

- [ ] **Step 1: Create in-memory token module**

Create `dx-web/src/lib/token.ts`:

```typescript
let accessToken: string | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string): void {
  accessToken = token;
}

export function clearAccessToken(): void {
  accessToken = null;
}
```

- [ ] **Step 2: Rewrite api-client.ts token management**

In `dx-web/src/lib/api-client.ts`:

Remove: `TOKEN_KEY`, `REFRESH_TOKEN_KEY` constants, `getToken()`, `setToken()`, `removeToken()`, `setAuthCookie()`, `removeAuthCookie()` functions.

Add imports at top:

```typescript
import { getAccessToken, setAccessToken, clearAccessToken } from "@/lib/token";
```

- [ ] **Step 3: Add refresh lock mechanism**

Add after imports:

```typescript
let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = fetch(`${API_URL}/api/auth/refresh`, {
    method: "POST",
    credentials: "include",
  })
    .then((res) => {
      if (!res.ok) throw new Error("refresh failed");
      return res.json();
    })
    .then((data: ApiResponse<{ access_token: string }>) => {
      if (data.code !== 0) throw new Error("refresh failed");
      setAccessToken(data.data.access_token);
      return data.data.access_token;
    })
    .finally(() => {
      refreshPromise = null;
    });

  return refreshPromise;
}
```

- [ ] **Step 4: Rewrite apiFetch with 401 retry**

```typescript
async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const token = getAccessToken();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  if (token) {
    (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  // On 401, attempt refresh and retry once
  if (res.status === 401) {
    try {
      const newToken = await refreshAccessToken();
      const retryHeaders = { ...headers } as Record<string, string>;
      retryHeaders["Authorization"] = `Bearer ${newToken}`;

      const retryRes = await fetch(`${API_URL}${path}`, {
        ...options,
        headers: retryHeaders,
        credentials: "include",
      });

      if (retryRes.status === 401) {
        clearAccessToken();
        if (typeof window !== "undefined") {
          window.location.href = "/auth/signin";
        }
        throw new Error("Unauthorized");
      }

      return retryRes.json();
    } catch {
      clearAccessToken();
      if (typeof window !== "undefined") {
        window.location.href = "/auth/signin";
      }
      throw new Error("Unauthorized");
    }
  }

  return res.json();
}
```

- [ ] **Step 5: Update uploadApi.uploadImage**

Update to use `getAccessToken()` and add 401 retry:

```typescript
async uploadImage(file: File, role: string) {
  const token = getAccessToken();
  const formData = new FormData();
  formData.append("file", file);
  formData.append("role", role);

  const doUpload = async (authToken: string | null) => {
    const res = await fetch(`${API_URL}/api/uploads/images`, {
      method: "POST",
      headers: authToken ? { Authorization: `Bearer ${authToken}` } : {},
      body: formData,
      credentials: "include",
    });
    return res;
  };

  let res = await doUpload(token);

  if (res.status === 401) {
    try {
      const newToken = await refreshAccessToken();
      res = await doUpload(newToken);
    } catch {
      clearAccessToken();
      if (typeof window !== "undefined") {
        window.location.href = "/auth/signin";
      }
      throw new Error("Unauthorized");
    }
  }

  if (res.status === 401) {
    clearAccessToken();
    if (typeof window !== "undefined") {
      window.location.href = "/auth/signin";
    }
    throw new Error("Unauthorized");
  }

  const data: ApiResponse<{ id: string; url: string; name: string }> = await res.json();
  return data;
},
```

- [ ] **Step 6: Update authApi response types**

Update `signIn` and `signUp` return types from `{ token: string; user: ... }` to `{ access_token: string; refresh_token: string; user: ... }`.

Update `refresh` return type to `{ access_token: string; refresh_token: string }`.

Update `logout` to bypass `apiFetch` (since logout is now a public endpoint and does not need Bearer token):

```typescript
async logout() {
  try {
    await fetch(`${API_URL}/api/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
  } catch {
    // Ignore network errors on logout
  }
  clearAccessToken();
  if (typeof document !== "undefined") {
    document.cookie = "dx_refresh=; path=/api/auth; max-age=0";
  }
},
```

Remove `authApi.refresh()` — refresh is now handled automatically by the 401 interceptor in `apiFetch`. Keeping it as a public API risks misuse.

- [ ] **Step 7: Clean up exports**

Remove from exports: `getToken`, `setToken`, `removeToken`, `setAuthCookie`, `removeAuthCookie`, `TOKEN_KEY`, `REFRESH_TOKEN_KEY`.

Add to exports: `refreshAccessToken` (needed by auth-guard).

- [ ] **Step 8: Do NOT commit yet** — hooks still reference old `setToken`. Continue to Task 8.

---

## Task 8: Frontend — Auth Hooks & Auth Guard (commit together with Task 7)

**Files:**
- Modify: `dx-web/src/features/web/auth/hooks/use-signin.ts:6,121,158`
- Modify: `dx-web/src/features/web/auth/hooks/use-signup.ts:6,43,107`
- Create: `dx-web/src/components/in/auth-guard.tsx`

- [ ] **Step 1: Update use-signin.ts**

Change import from `import { authApi, setToken } from "@/lib/api-client"` to `import { authApi } from "@/lib/api-client"` and add `import { setAccessToken } from "@/lib/token"`.

Replace `setToken(res.data.token)` (line 121) with `setAccessToken(res.data.access_token)`.

Replace `setToken(res.data.token)` (line 158) with `setAccessToken(res.data.access_token)`.

- [ ] **Step 2: Update use-signup.ts**

Change import from `import { authApi } from "@/lib/api-client"` — add `import { setAccessToken } from "@/lib/token"`.

After `setSignUpState({ success: true })` (line 108), add `setAccessToken(res.data.access_token)`.

Change redirect from `router.push("/auth/signin")` (line 43) to `router.push("/hall")`.

- [ ] **Step 3: Create auth-guard component**

Create `dx-web/src/components/in/auth-guard.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getAccessToken } from "@/lib/token";
import { refreshAccessToken } from "@/lib/api-client";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [status, setStatus] = useState<"loading" | "authenticated" | "unauthenticated">("loading");

  useEffect(() => {
    async function checkAuth() {
      if (getAccessToken()) {
        setStatus("authenticated");
        return;
      }

      // No token in memory — try silent refresh via cookie
      try {
        await refreshAccessToken();
        setStatus("authenticated");
      } catch {
        setStatus("unauthenticated");
        router.replace("/auth/signin");
      }
    }

    checkAuth();
  }, [router]);

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return null;
  }

  return <>{children}</>;
}
```

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npm run build`
Expected: May fail due to remaining `apiServerFetch` imports. Those are fixed in Task 9.

- [ ] **Step 5: Commit (Tasks 7 + 8 together)**

```bash
git add dx-web/src/lib/token.ts dx-web/src/lib/api-client.ts dx-web/src/features/web/auth/hooks/use-signin.ts dx-web/src/features/web/auth/hooks/use-signup.ts dx-web/src/components/in/auth-guard.tsx
git commit -m "feat: in-memory token, refresh-on-401, auth guard, updated auth hooks"
```

---

## Task 9: Frontend — Proxy & SSR Migration

**Files:**
- Modify: `dx-web/src/proxy.ts`
- Delete: `dx-web/src/lib/auth.ts`
- Delete: `dx-web/src/lib/api-server.ts`
- Modify: 34 files importing `apiServerFetch`

This is the largest task. It converts all server-side authenticated fetches to client-side.

- [ ] **Step 1: Update proxy.ts**

Replace `dx_token` cookie check with `dx_refresh` cookie check:

```typescript
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const authRoutes = ["/auth/signin", "/auth/signup"];
const protectedRoutes = ["/hall"];

export default function middleware(request: NextRequest) {
  const hasRefreshToken = !!request.cookies.get("dx_refresh")?.value;
  const { pathname } = request.nextUrl;

  // Users with refresh token should not see signin/signup pages
  if (hasRefreshToken && authRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/hall", request.url));
  }

  // Users without refresh token cannot access protected routes
  if (!hasRefreshToken && protectedRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/auth/signin", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/hall/:path*", "/auth/signin", "/auth/signup"],
};
```

- [ ] **Step 2: Delete auth.ts and api-server.ts**

Delete `dx-web/src/lib/auth.ts` and `dx-web/src/lib/api-server.ts`.

- [ ] **Step 3: Migrate server actions and page components**

For each of the 34 files importing `apiServerFetch`:

**Pattern for server actions** (files in `features/web/*/actions/`):

These files currently export async functions that call `apiServerFetch`. Convert them to simply re-export the relevant `apiClient` methods, or if the action does data transformation, keep the function but call `apiClient` instead.

Example — a file like `features/web/hall/actions/heatmap.action.ts`:

Before:
```typescript
"use server";
import { apiServerFetch } from "@/lib/api-server";
export async function getHeatmap(year: number) {
  return apiServerFetch(`/api/hall/heatmap?year=${year}`);
}
```

After — delete the file if the component can call `apiClient` directly, or convert:
```typescript
import { apiClient } from "@/lib/api-client";
export async function getHeatmap(year: number) {
  return apiClient.get(`/api/hall/heatmap?year=${year}`);
}
```

**Pattern for page components** (files in `app/(web)/`):

These are server components that fetch data at render time. Convert to client components that fetch on mount:

Before:
```tsx
import { apiServerFetch } from "@/lib/api-server";
export default async function Page() {
  const data = await apiServerFetch("/api/some-endpoint");
  return <SomeComponent data={data.data} />;
}
```

After:
```tsx
"use client";
import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { AuthGuard } from "@/components/in/auth-guard";

export default function Page() {
  const [data, setData] = useState(null);
  useEffect(() => {
    apiClient.get("/api/some-endpoint").then((res) => {
      if (res.code === 0) setData(res.data);
    });
  }, []);

  return (
    <AuthGuard>
      <SomeComponent data={data} />
    </AuthGuard>
  );
}
```

**Note:** Each page/action file is different. The agent implementing this task should read each file and apply the appropriate conversion. The key rules are:
- Remove `"use server"` directives from action files that switch to client-side
- Add `"use client"` to page components that become client components
- Replace `apiServerFetch` with the matching `apiClient` method
- Wrap protected page content in `<AuthGuard>`
- Handle loading states as needed

- [ ] **Step 4: Clear legacy dx_token cookie**

In `dx-web/src/proxy.ts`, add cleanup for old cookie at the top of the middleware function:

```typescript
// Clear legacy dx_token cookie if present
if (request.cookies.get("dx_token")?.value) {
  const response = NextResponse.next();
  response.cookies.delete("dx_token");
  // Continue with the rest of the middleware logic using this response
  // ... (integrate into the existing flow)
}
```

- [ ] **Step 5: Verify build**

Run: `cd dx-web && npm run build`
Expected: Build passes with no `apiServerFetch` references.

- [ ] **Step 6: Verify no remaining references**

Run: `grep -r "apiServerFetch\|api-server\|lib/auth" dx-web/src/ --include="*.ts" --include="*.tsx"`
Expected: No matches (except possibly in unused type imports — fix those too).

- [ ] **Step 6.1: Clean up dead env variable**

Remove `API_INTERNAL_URL` from `deploy/env/.env.dev` and `deploy/env/.env.example` — it was only used by `api-server.ts` which is now deleted.

- [ ] **Step 7: Commit**

```bash
git add -A dx-web/
git commit -m "refactor: migrate all SSR auth fetches to client-side, remove api-server and auth.ts"
```

---

## Task 10: Final Verification

- [ ] **Step 1: Full backend build**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: Clean pass.

- [ ] **Step 2: Full frontend build**

Run: `cd dx-web && npm run build`
Expected: Clean pass.

- [ ] **Step 3: Lint**

Run: `cd dx-web && npm run lint`
Expected: Clean pass.

- [ ] **Step 4: Verify no old token references**

Run: `grep -r "dx_token\|setToken\|getToken\|removeToken\|setAuthCookie\|removeAuthCookie\|TOKEN_KEY\|REFRESH_TOKEN_KEY" dx-web/src/ --include="*.ts" --include="*.tsx"`
Expected: No matches.

Run: `grep -r "Guard.*Refresh\|Guard.*Logout\|refresh_ttl" dx-api/ --include="*.go" | grep -v "_test.go"`
Expected: No matches (old Goravel refresh/logout patterns removed).

- [ ] **Step 5: Commit any fixes**

If any issues found, fix and commit.

```bash
git commit -m "chore: final cleanup for dual-token auth migration"
```
