# Goravel JWT Auth Migration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the custom dual-token JWT auth with Goravel's built-in single-JWT auth, stored in httpOnly cookies, with auto-refresh in middleware and single-device enforcement via login timestamp.

**Architecture:** Goravel's `LoginUsingID`/`Parse`/`Refresh`/`Logout` replaces all custom JWT code. A single JWT in an httpOnly cookie replaces the access-token-in-memory + opaque-refresh-token-in-cookie model. Middleware auto-refreshes expired tokens (Goravel's official pattern). Single-device enforcement compares JWT `iat` against a login timestamp in Redis.

**Tech Stack:** Go/Goravel (backend), Next.js 16 (frontend), Redis, PostgreSQL

**Spec:** `docs/superpowers/specs/2026-04-03-goravel-jwt-auth-migration-design.md`

---

## File Map

### Backend (dx-api) — files changing

| File | Action | Responsibility |
|---|---|---|
| `config/jwt.go` | Rewrite | JWT + cookie config (Goravel defaults) |
| `app/http/middleware/jwt_auth.go` | Rewrite | Cookie-based auth + auto-refresh + single-device |
| `app/http/middleware/adm_jwt_auth.go` | Rewrite | Same for admin guard |
| `app/services/api/auth_service.go` | Rewrite | User auth using Goravel's LoginUsingID/Logout |
| `app/services/adm/auth_service.go` | Rewrite | Admin auth using Goravel's LoginUsingID/Logout |
| `app/http/controllers/api/auth_controller.go` | Rewrite | SignUp/SignIn/Me/Logout with cookie helpers |
| `app/http/controllers/adm/auth_controller.go` | Rewrite | Login/Me/Logout with cookie helpers |
| `routes/api.go` | Edit | Remove refresh route, move SSE to protected |
| `routes/adm.go` | Edit | Remove refresh route |
| `app/consts/error_code.go` | Edit | Remove TokenExpired, InvalidRefreshToken |
| `app/services/api/errors.go` | Edit | Remove ErrInvalidRefreshToken |
| `app/services/adm/errors.go` | Edit | Remove ErrInvalidRefreshToken |
| `app/http/controllers/api/group_game_controller.go` | Edit | Remove manual token parsing from Events |
| `app/http/controllers/api/group_notify_controller.go` | Edit | Remove manual token parsing from Notify |
| `app/helpers/jwt.go` | Delete | Custom JWT issuance — replaced by Goravel |
| `app/helpers/refresh_token.go` | Delete | Opaque refresh system — eliminated |

### Frontend (dx-web) — files changing

| File | Action | Responsibility |
|---|---|---|
| `src/lib/token.ts` | Delete | In-memory token — no purpose with cookies |
| `src/components/in/auth-guard.tsx` | Delete | Session recovery — replaced by proxy.ts |
| `src/lib/api-client.ts` | Rewrite | Remove all token logic, use credentials: include |
| `src/proxy.ts` | Edit | Check dx_token instead of dx_refresh |
| `src/hooks/use-group-sse.ts` | Rewrite | Cookie-based SSE, no token injection |
| `src/hooks/use-group-notify.ts` | Rewrite | Cookie-based SSE, no token injection |
| `src/features/web/ai-custom/helpers/stream-progress.ts` | Edit | credentials: include instead of Bearer |
| `src/features/web/ai-custom/helpers/generate-api.ts` | Edit | credentials: include instead of Bearer |
| `src/features/web/ai-custom/helpers/format-api.ts` | Edit | credentials: include instead of Bearer |
| `src/features/web/play-single/components/game-play-shell.tsx` | Edit | credentials: include for beforeunload sync |
| `src/features/web/play-group/components/group-play-shell.tsx` | Edit | credentials: include for beforeunload sync |
| `src/features/com/images/hooks/use-image-uploader.ts` | Edit | withCredentials for Uppy XHR |
| `src/features/web/auth/hooks/use-signin.ts` | Edit | Remove setAccessToken |
| `src/features/web/auth/hooks/use-signup.ts` | Edit | Remove setAccessToken |
| `src/features/web/auth/components/user-profile-menu.tsx` | Edit | Remove removeToken |
| `src/components/in/landing-header.tsx` | Rewrite | Server-side login detection |
| `src/app/(web)/hall/layout.tsx` | Edit | Remove AuthGuard wrapper |

### Deploy

| File | Action |
|---|---|
| `deploy/env/.env.dev` | Edit — update JWT env vars |
| `deploy/env/.env.example` | Edit — update JWT env vars |

---

## Task 1: Backend Config + Deploy Env

**Files:**
- Modify: `dx-api/config/jwt.go`
- Modify: `deploy/env/.env.dev`
- Modify: `deploy/env/.env.example`

- [ ] **Step 1: Rewrite config/jwt.go**

Replace the entire file:

```go
package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("jwt", map[string]any{
		"secret":      config.Env("JWT_SECRET", ""),
		"ttl":         config.Env("JWT_TTL", 60),
		"refresh_ttl": config.Env("JWT_REFRESH_TTL", 20160),
	})
	config.Add("jwt_cookie", map[string]any{
		"secure": config.Env("JWT_COOKIE_SECURE", true),
	})
}
```

- [ ] **Step 2: Update deploy/env/.env.dev**

Replace lines 12-14:

```
JWT_TTL=10
REFRESH_TOKEN_TTL=10080
REFRESH_COOKIE_SECURE=false
```

With:

```
JWT_TTL=60
JWT_REFRESH_TTL=20160
JWT_COOKIE_SECURE=false
```

- [ ] **Step 3: Update deploy/env/.env.example**

Replace lines 17-19:

```
JWT_TTL=10
REFRESH_TOKEN_TTL=10080
REFRESH_COOKIE_SECURE=true
```

With:

```
JWT_TTL=60
JWT_REFRESH_TTL=20160
JWT_COOKIE_SECURE=true
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: success (config change only, no callers broken yet)

- [ ] **Step 5: Commit**

```bash
git add dx-api/config/jwt.go deploy/env/.env.dev deploy/env/.env.example
git commit -m "refactor: update JWT config to Goravel defaults (60min TTL, 14d refresh)"
```

---

## Task 2: Rewrite User JWT Middleware

**Files:**
- Modify: `dx-api/app/http/middleware/jwt_auth.go`

- [ ] **Step 1: Rewrite jwt_auth.go**

Replace the entire file with Goravel's official pattern + cookie + single-device:

```go
package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/goravel/framework/auth"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

func JwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Cookie("dx_token")
		if token == "" {
			abortUnauthorized(ctx)
			return
		}

		payload, err := facades.Auth(ctx).Guard("user").Parse(token)
		if err != nil {
			if errors.Is(err, auth.ErrorTokenExpired) {
				newToken, refreshErr := facades.Auth(ctx).Guard("user").Refresh()
				if refreshErr != nil {
					clearTokenCookie(ctx, "dx_token")
					abortUnauthorized(ctx)
					return
				}
				setTokenCookie(ctx, "dx_token", newToken)
				// Re-parse to populate payload after refresh
				payload, _ = facades.Auth(ctx).Guard("user").Parse(newToken)
			} else {
				clearTokenCookie(ctx, "dx_token")
				abortUnauthorized(ctx)
				return
			}
		}

		if payload == nil {
			clearTokenCookie(ctx, "dx_token")
			abortUnauthorized(ctx)
			return
		}

		// Single-device check: token iat must be >= login timestamp
		userID, _ := facades.Auth(ctx).Guard("user").ID()
		loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":user")
		if redisErr != nil {
			clearTokenCookie(ctx, "dx_token")
			abortUnauthorized(ctx)
			return
		}

		loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
		if payload.IssuedAt.Unix() < loginTs {
			clearTokenCookie(ctx, "dx_token")
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

func abortUnauthorized(ctx contractshttp.Context) {
	_ = ctx.Response().Json(401, helpers.Response{
		Code:    consts.CodeUnauthorized,
		Message: "unauthorized",
	}).Abort()
}

func setTokenCookie(ctx contractshttp.Context, name, token string) {
	secure := facades.Config().GetBool("jwt_cookie.secure", true)
	ttl := facades.Config().GetInt("jwt.refresh_ttl", 20160)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func clearTokenCookie(ctx contractshttp.Context, name string) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("jwt_cookie.secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func SetTokenCookie(ctx contractshttp.Context, name, token string) {
	setTokenCookie(ctx, name, token)
}

func ClearTokenCookie(ctx contractshttp.Context, name string) {
	clearTokenCookie(ctx, name)
}
```

Note: `SetTokenCookie` and `ClearTokenCookie` are exported for use by auth controllers. The unexported versions are used internally by middleware.

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: may show errors from callers of old `ExtractAuthID` — that's expected, will fix in later tasks.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/http/middleware/jwt_auth.go
git commit -m "refactor: rewrite JwtAuth middleware for cookie + auto-refresh + iat check"
```

---

## Task 3: Rewrite Admin JWT Middleware

**Files:**
- Modify: `dx-api/app/http/middleware/adm_jwt_auth.go`

- [ ] **Step 1: Rewrite adm_jwt_auth.go**

Replace the entire file:

```go
package middleware

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/auth"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

func AdmJwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Cookie("dx_adm_token")
		if token == "" {
			abortUnauthorized(ctx)
			return
		}

		payload, err := facades.Auth(ctx).Guard("admin").Parse(token)
		if err != nil {
			if errors.Is(err, auth.ErrorTokenExpired) {
				newToken, refreshErr := facades.Auth(ctx).Guard("admin").Refresh()
				if refreshErr != nil {
					clearTokenCookie(ctx, "dx_adm_token")
					abortUnauthorized(ctx)
					return
				}
				setTokenCookie(ctx, "dx_adm_token", newToken)
				payload, _ = facades.Auth(ctx).Guard("admin").Parse(newToken)
			} else {
				clearTokenCookie(ctx, "dx_adm_token")
				abortUnauthorized(ctx)
				return
			}
		}

		if payload == nil {
			clearTokenCookie(ctx, "dx_adm_token")
			abortUnauthorized(ctx)
			return
		}

		userID, _ := facades.Auth(ctx).Guard("admin").ID()
		loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":admin")
		if redisErr != nil {
			clearTokenCookie(ctx, "dx_adm_token")
			abortUnauthorized(ctx)
			return
		}

		loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
		if payload.IssuedAt.Unix() < loginTs {
			clearTokenCookie(ctx, "dx_adm_token")
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

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/http/middleware/adm_jwt_auth.go
git commit -m "refactor: rewrite AdmJwtAuth middleware for cookie + auto-refresh + iat check"
```

---

## Task 4: Rewrite User Auth Service

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go`
- Modify: `dx-api/app/services/api/errors.go`

- [ ] **Step 1: Rewrite auth_service.go**

Replace the entire file:

```go
package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

// issueSession generates a JWT via Goravel and stores login timestamp in Redis.
func issueSession(ctx contractshttp.Context, userID string) (string, error) {
	token, err := facades.Auth(ctx).Guard("user").LoginUsingID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}

	// Store login timestamp for single-device enforcement
	loginTs := strconv.FormatInt(time.Now().Unix(), 10)
	ttl := time.Duration(facades.Config().GetInt("jwt.refresh_ttl", 20160)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:user", userID), loginTs, ttl); err != nil {
		return "", fmt.Errorf("failed to store login timestamp: %w", err)
	}

	return token, nil
}

// SignUp registers a new user with the given email, verification code, username, and password.
func SignUp(ctx contractshttp.Context, email, code, username, password string) (string, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signup_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return "", nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Check duplicate email
	var existing models.User
	err = facades.Orm().Query().Where("email", email).First(&existing)
	if err == nil && existing.ID != "" {
		return "", nil, ErrDuplicateEmail
	}

	// Derive username from email prefix if empty
	if username == "" {
		username = strings.Split(email, "@")[0]
	}

	// Check duplicate username
	err = facades.Orm().Query().Where("username", username).First(&existing)
	if err == nil && existing.ID != "" {
		return "", nil, ErrDuplicateUsername
	}

	// Auto-generate password if empty
	if password == "" {
		password = helpers.GenerateInviteCode(16)
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	emailStr := email
	user := models.User{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Username:   username,
		Email:      &emailStr,
		Password:   hashedPassword,
		IsActive:   true,
		InviteCode: helpers.GenerateInviteCode(8),
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// SignInByEmail authenticates a user via email and verification code.
// If the user does not exist, a new account is created automatically.
func SignInByEmail(ctx contractshttp.Context, email, code string) (string, *models.User, error) {
	// Verify code
	key := fmt.Sprintf("signin_code:%s", email)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return "", nil, ErrInvalidCode
	}
	_ = helpers.RedisDel(key)

	// Find user by email
	var user models.User
	err = facades.Orm().Query().Where("email", email).First(&user)
	if err != nil || user.ID == "" {
		// Auto-register
		username := strings.Split(email, "@")[0]

		var existingUser models.User
		if checkErr := facades.Orm().Query().Where("username", username).First(&existingUser); checkErr == nil && existingUser.ID != "" {
			username = fmt.Sprintf("%s_%s", username, helpers.GenerateCode(4))
		}

		pw := helpers.GenerateInviteCode(16)
		hashedPw, hashErr := helpers.HashPassword(pw)
		if hashErr != nil {
			return "", nil, fmt.Errorf("failed to hash password: %w", hashErr)
		}

		emailStr := email
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Username:   username,
			Email:      &emailStr,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
		}

		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// SignInByAccount authenticates a user via account (username, email, or phone) and password.
func SignInByAccount(ctx contractshttp.Context, account, password string) (string, *models.User, error) {
	var user models.User

	err := facades.Orm().Query().
		Where("username", account).
		OrWhere("email", account).
		OrWhere("phone", account).
		First(&user)
	if err != nil || user.ID == "" {
		return "", nil, ErrUserNotFound
	}

	if !helpers.CheckPassword(password, user.Password) {
		return "", nil, ErrInvalidPassword
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

// Logout blacklists the current JWT and deletes the login timestamp from Redis.
func Logout(ctx contractshttp.Context) {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err == nil && userID != "" {
		_ = facades.Auth(ctx).Guard("user").Logout()
		_ = helpers.RedisDel(fmt.Sprintf("user_auth:%s:user", userID))
	}
}

// GetCurrentUser retrieves the user profile by ID (password excluded via json tag).
func GetCurrentUser(userID string) (*models.User, error) {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// RecordLogin creates a UserLogin record for audit purposes.
func RecordLogin(userID, ip, userAgent string) {
	agent := userAgent
	login := models.UserLogin{
		ID:     uuid.Must(uuid.NewV7()).String(),
		UserID: userID,
		IP:     ip,
		Agent:  &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
```

- [ ] **Step 2: Remove ErrInvalidRefreshToken from errors.go**

In `dx-api/app/services/api/errors.go`, delete the line:

```go
ErrInvalidRefreshToken     = errors.New("invalid or expired refresh token")
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: errors from auth controller (still references old AuthResult) — fixed in Task 6.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/auth_service.go dx-api/app/services/api/errors.go
git commit -m "refactor: rewrite user auth service to use Goravel LoginUsingID"
```

---

## Task 5: Rewrite Admin Auth Service

**Files:**
- Modify: `dx-api/app/services/adm/auth_service.go`
- Modify: `dx-api/app/services/adm/errors.go`

- [ ] **Step 1: Rewrite auth_service.go**

Replace the entire file:

```go
package adm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

// issueAdminSession generates a JWT via Goravel and stores login timestamp in Redis.
func issueAdminSession(ctx contractshttp.Context, userID string) (string, error) {
	token, err := facades.Auth(ctx).Guard("admin").LoginUsingID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}

	loginTs := strconv.FormatInt(time.Now().Unix(), 10)
	ttl := time.Duration(facades.Config().GetInt("jwt.refresh_ttl", 20160)) * time.Minute
	if err := helpers.RedisSet(fmt.Sprintf("user_auth:%s:admin", userID), loginTs, ttl); err != nil {
		return "", fmt.Errorf("failed to store login timestamp: %w", err)
	}

	return token, nil
}

// AdminSignIn authenticates an admin user via username and password.
func AdminSignIn(ctx contractshttp.Context, username, password string) (string, *models.AdmUser, error) {
	var admUser models.AdmUser
	err := facades.Orm().Query().Where("username", username).First(&admUser)
	if err != nil || admUser.ID == "" {
		return "", nil, ErrAdminNotFound
	}

	if !admUser.IsActive {
		return "", nil, ErrAdminInactive
	}

	if !helpers.CheckPassword(password, admUser.Password) {
		return "", nil, ErrInvalidPassword
	}

	token, err := issueAdminSession(ctx, admUser.ID)
	if err != nil {
		return "", nil, err
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go RecordAdminLogin(admUser.ID, ip, userAgent)

	return token, &admUser, nil
}

// Logout blacklists the current JWT and deletes the login timestamp from Redis.
func Logout(ctx contractshttp.Context) {
	userID, err := facades.Auth(ctx).Guard("admin").ID()
	if err == nil && userID != "" {
		_ = facades.Auth(ctx).Guard("admin").Logout()
		_ = helpers.RedisDel(fmt.Sprintf("user_auth:%s:admin", userID))
	}
}

// GetAdminUser retrieves an admin user by ID.
func GetAdminUser(userID string) (*models.AdmUser, error) {
	var admUser models.AdmUser
	if err := facades.Orm().Query().Where("id", userID).First(&admUser); err != nil {
		return nil, fmt.Errorf("failed to find admin user: %w", err)
	}
	if admUser.ID == "" {
		return nil, ErrAdminNotFound
	}
	return &admUser, nil
}

// RecordAdminLogin creates an AdmLogin record for audit purposes.
func RecordAdminLogin(admUserID, ip, userAgent string) {
	agent := userAgent
	login := models.AdmLogin{
		ID:        uuid.Must(uuid.NewV7()).String(),
		AdmUserID: admUserID,
		Ip:        ip,
		Agent:     &agent,
	}
	_ = facades.Orm().Query().Create(&login)
}
```

- [ ] **Step 2: Remove ErrInvalidRefreshToken from errors.go**

In `dx-api/app/services/adm/errors.go`, delete the line:

```go
ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/adm/auth_service.go dx-api/app/services/adm/errors.go
git commit -m "refactor: rewrite admin auth service to use Goravel LoginUsingID"
```

---

## Task 6: Rewrite User Auth Controller

**Files:**
- Modify: `dx-api/app/http/controllers/api/auth_controller.go`

- [ ] **Step 1: Rewrite auth_controller.go**

Replace the entire file:

```go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/http/middleware"
	requests "dx-api/app/http/requests/api"
	"dx-api/app/models"
	services "dx-api/app/services/api"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// SignUp registers a new user.
func (c *AuthController) SignUp(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignUpRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	token, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "验证码无效或已过期")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "该邮箱已注册")
		case errors.Is(err, services.ErrDuplicateUsername):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateUsername, "用户名已被使用")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to sign up")
		}
	}

	middleware.SetTokenCookie(ctx, "dx_token", token)
	return helpers.Success(ctx, map[string]any{"user": user})
}

// SignIn authenticates a user via email+code or account+password.
func (c *AuthController) SignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignInRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	var (
		token string
		user  *models.User
		err   error
	)

	if req.Email != "" {
		token, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
	} else if req.Account != "" {
		token, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
	} else {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请输入邮箱或账号")
	}

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "验证码无效或已过期")
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		case errors.Is(err, services.ErrInvalidPassword):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidPassword, "密码错误")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to sign in")
		}
	}

	// Record login asynchronously
	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go services.RecordLogin(user.ID, ip, userAgent)

	middleware.SetTokenCookie(ctx, "dx_token", token)
	return helpers.Success(ctx, map[string]any{"user": user})
}

// Me returns the current authenticated user's profile.
func (c *AuthController) Me(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	user, err := services.GetCurrentUser(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
	}

	return helpers.Success(ctx, user)
}

// Logout blacklists the JWT and clears the cookie.
func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	// Parse token from cookie so Goravel can blacklist it
	token := ctx.Request().Cookie("dx_token")
	if token != "" {
		_, _ = facades.Auth(ctx).Guard("user").Parse(token)
		services.Logout(ctx)
	}

	middleware.ClearTokenCookie(ctx, "dx_token")
	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/http/controllers/api/auth_controller.go
git commit -m "refactor: rewrite user auth controller for cookie-based JWT"
```

---

## Task 7: Rewrite Admin Auth Controller

**Files:**
- Modify: `dx-api/app/http/controllers/adm/auth_controller.go`

- [ ] **Step 1: Rewrite auth_controller.go**

Replace the entire file:

```go
package adm

import (
	"errors"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/http/middleware"
	admservice "dx-api/app/services/adm"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (c *AuthController) Login(ctx contractshttp.Context) contractshttp.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")

	if username == "" || password == "" {
		return helpers.Error(ctx, 400, consts.CodeValidationError, "username and password are required")
	}

	token, admUser, err := admservice.AdminSignIn(ctx, username, password)
	if err != nil {
		if errors.Is(err, admservice.ErrAdminNotFound) || errors.Is(err, admservice.ErrInvalidPassword) {
			return helpers.Error(ctx, 401, consts.CodeUnauthorized, "invalid username or password")
		}
		if errors.Is(err, admservice.ErrAdminInactive) {
			return helpers.Error(ctx, 403, consts.CodeForbidden, "admin account is inactive")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "internal server error")
	}

	middleware.SetTokenCookie(ctx, "dx_adm_token", token)
	return helpers.Success(ctx, map[string]any{"user": admUser})
}

func (c *AuthController) Me(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("admin").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, 401, consts.CodeUnauthorized, "unauthorized")
	}

	admUser, err := admservice.GetAdminUser(userID)
	if err != nil {
		if errors.Is(err, admservice.ErrAdminNotFound) {
			return helpers.Error(ctx, 404, consts.CodeUserNotFound, "admin user not found")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "internal server error")
	}

	return helpers.Success(ctx, admUser)
}

func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	token := ctx.Request().Cookie("dx_adm_token")
	if token != "" {
		_, _ = facades.Auth(ctx).Guard("admin").Parse(token)
		admservice.Logout(ctx)
	}

	middleware.ClearTokenCookie(ctx, "dx_adm_token")
	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/http/controllers/adm/auth_controller.go
git commit -m "refactor: rewrite admin auth controller for cookie-based JWT"
```

---

## Task 8: Update Routes + Error Codes + SSE Handlers

**Files:**
- Modify: `dx-api/routes/api.go:61-66`
- Modify: `dx-api/routes/adm.go:20-24`
- Modify: `dx-api/app/consts/error_code.go:27-29`
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go:164-173`
- Modify: `dx-api/app/http/controllers/api/group_notify_controller.go:21-29`

- [ ] **Step 1: Remove refresh routes from api.go**

In `dx-api/routes/api.go`, change lines 61-66 from:

```go
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/signup", authController.SignUp)
			auth.Post("/signin", authController.SignIn)
			auth.Post("/refresh", authController.Refresh)
			auth.Post("/logout", authController.Logout)
		})
```

To:

```go
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/signup", authController.SignUp)
			auth.Post("/signin", authController.SignIn)
			auth.Post("/logout", authController.Logout)
		})
```

- [ ] **Step 2: Move SSE routes to protected group in api.go**

Remove lines 78-84 (the public SSE routes):

```go
		// Group SSE events (query-param auth, not JWT middleware)
		groupGameController := apicontrollers.NewGroupGameController()
		router.Get("/groups/{id}/events", groupGameController.Events)

		// Group detail notification SSE (query-param auth, not JWT middleware)
		groupNotifyController := apicontrollers.NewGroupNotifyController()
		router.Get("/groups/{id}/notify", groupNotifyController.Notify)
```

Add them inside the protected group (after line 97, inside `router.Middleware(middleware.JwtAuth()).Group(func(protected route.Router) {`):

```go
			// Group SSE events (cookie auth via JwtAuth middleware)
			groupGameController := apicontrollers.NewGroupGameController()
			protected.Get("/groups/{id}/events", groupGameController.Events)
			groupNotifyController := apicontrollers.NewGroupNotifyController()
			protected.Get("/groups/{id}/notify", groupNotifyController.Notify)
```

Note: the existing `groupGameController` variable inside the protected group (line 332) was only used for non-SSE routes. The new SSE declaration at the top of the protected block shadows this — rename the existing one inside the groups sub-router to avoid conflict, or better: remove the duplicate `groupGameController` from line 332 since it now exists at the top of the protected block.

- [ ] **Step 3: Remove refresh route from adm.go**

In `dx-api/routes/adm.go`, change lines 20-24 from:

```go
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/login", admAuthController.Login)
			auth.Post("/refresh", admAuthController.Refresh)
			auth.Post("/logout", admAuthController.Logout)
		})
```

To:

```go
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/login", admAuthController.Login)
			auth.Post("/logout", admAuthController.Logout)
		})
```

- [ ] **Step 4: Update SSE handler — group_game_controller.go Events**

Replace the token-reading block (lines 165-173):

```go
	token := ctx.Request().Query("token", "")
	if token == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "missing token")
	}

	userID, err := helpers.ParseJWTUserID(token)
	if err != nil {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "invalid token")
	}
```

With (middleware already validated; just get user ID):

```go
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "unauthorized")
	}
```

Add the import `"github.com/goravel/framework/facades"` if not already present. Remove unused `"dx-api/app/helpers"` import if no other helpers are used in this method (check — `helpers.GroupSSEHub` is still used, so keep it).

- [ ] **Step 5: Update SSE handler — group_notify_controller.go Notify**

Replace the token-reading block (lines 22-29):

```go
	token := ctx.Request().Query("token", "")
	if token == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "missing token")
	}

	userID, err := helpers.ParseJWTUserID(token)
	if err != nil {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "invalid token")
	}
```

With:

```go
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "unauthorized")
	}
```

Add `"github.com/goravel/framework/facades"` import.

- [ ] **Step 6: Remove unused error codes from error_code.go**

In `dx-api/app/consts/error_code.go`, delete these two lines:

```go
	CodeTokenExpired        = 40101
	CodeInvalidRefreshToken = 40103
```

- [ ] **Step 7: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: errors about `helpers.IssueAccessToken`, `helpers.ExtractAuthID`, etc. from files we haven't touched yet — these are the files being deleted in the next task.

- [ ] **Step 8: Commit**

```bash
git add dx-api/routes/api.go dx-api/routes/adm.go dx-api/app/consts/error_code.go \
  dx-api/app/http/controllers/api/group_game_controller.go \
  dx-api/app/http/controllers/api/group_notify_controller.go
git commit -m "refactor: remove refresh routes, move SSE to protected, clean error codes"
```

---

## Task 9: Delete Old Backend Helper Files

**Files:**
- Delete: `dx-api/app/helpers/jwt.go`
- Delete: `dx-api/app/helpers/refresh_token.go`

- [ ] **Step 1: Delete files**

```bash
rm dx-api/app/helpers/jwt.go dx-api/app/helpers/refresh_token.go
```

- [ ] **Step 2: Verify full compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build. All references to deleted helpers have been replaced in previous tasks.

If there are remaining compile errors, they indicate a missed reference — fix them before proceeding.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/helpers/jwt.go dx-api/app/helpers/refresh_token.go
git commit -m "refactor: delete custom JWT and refresh token helpers"
```

---

## Task 10: Rewrite Frontend api-client.ts

**Files:**
- Modify: `dx-web/src/lib/api-client.ts`

- [ ] **Step 1: Remove token imports and refresh logic (lines 25-69)**

Delete lines 25-69 entirely (token imports, aliases, refreshPromise, refreshAccessToken function).

- [ ] **Step 2: Rewrite apiFetch (lines 72-136)**

Replace with:

```typescript
// Base fetch wrapper
async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (res.status === 401) {
    const errorData: ApiResponse<null> = await res.clone().json().catch(() => ({ code: 0, message: "", data: null }));

    if (errorData.code === 40104 && typeof window !== "undefined") {
      alert("您的账号已在其他设备登录");
      window.location.href = "/auth/signin";
    } else if (typeof window !== "undefined") {
      window.location.href = "/auth/signin";
    }
    throw new Error("Unauthorized");
  }

  return res.json();
}
```

- [ ] **Step 3: Remove token-related exports**

Delete these lines (near line 28-34):

```typescript
export const removeToken = clearAccessToken;
export const getToken = getAccessToken;
export const setToken = setAccessToken;
```

And the refreshAccessToken export.

- [ ] **Step 4: Update authApi signUp/signIn return handling**

In `authApi.signUp` and `authApi.signIn`, the response no longer contains `access_token` / `refresh_token`. The response is now `{ user }`. No code change needed in api-client.ts itself — callers will be updated in Task 13.

- [ ] **Step 5: Verify no remaining token references**

Run: `cd dx-web && grep -rn "getToken\|setToken\|removeToken\|getAccessToken\|setAccessToken\|clearAccessToken\|refreshAccessToken\|Bearer.*token\|token.*Bearer" src/lib/api-client.ts`
Expected: no matches

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/lib/api-client.ts
git commit -m "refactor: remove all token management from api-client, use credentials include"
```

---

## Task 11: Delete token.ts + auth-guard.tsx, Update proxy.ts + hall layout

**Files:**
- Delete: `dx-web/src/lib/token.ts`
- Delete: `dx-web/src/components/in/auth-guard.tsx`
- Modify: `dx-web/src/proxy.ts`
- Modify: `dx-web/src/app/(web)/hall/layout.tsx`

- [ ] **Step 1: Delete token.ts**

```bash
rm dx-web/src/lib/token.ts
```

- [ ] **Step 2: Delete auth-guard.tsx**

```bash
rm dx-web/src/components/in/auth-guard.tsx
```

- [ ] **Step 3: Rewrite proxy.ts**

Replace the entire file:

```typescript
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const authRoutes = ["/auth/signin", "/auth/signup"];
const protectedRoutes = ["/hall"];

export default function middleware(request: NextRequest) {
  const hasToken = !!request.cookies.get("dx_token")?.value;
  const { pathname } = request.nextUrl;

  // Users with token should not see signin/signup pages
  if (hasToken && authRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/hall", request.url));
  }

  // Users without token cannot access protected routes
  if (!hasToken && protectedRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/auth/signin", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/hall/:path*", "/auth/signin", "/auth/signup"],
};
```

- [ ] **Step 4: Update hall layout — remove AuthGuard**

Replace the entire file `dx-web/src/app/(web)/hall/layout.tsx`:

```typescript
import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";
import { SWRProvider } from "@/lib/swr";

export default function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <HallThemeProvider>
      <SWRProvider>{children}</SWRProvider>
    </HallThemeProvider>
  );
}
```

Note: SWRProvider was previously inside AuthGuard. It still needs to wrap the hall children.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/lib/token.ts dx-web/src/components/in/auth-guard.tsx \
  dx-web/src/proxy.ts dx-web/src/app/\(web\)/hall/layout.tsx
git commit -m "refactor: delete token.ts and auth-guard, update proxy to check dx_token"
```

---

## Task 12: Update Auth Hooks + Profile Menu

**Files:**
- Modify: `dx-web/src/features/web/auth/hooks/use-signin.ts:124,161`
- Modify: `dx-web/src/features/web/auth/hooks/use-signup.ts:109`
- Modify: `dx-web/src/features/web/auth/components/user-profile-menu.tsx:68-76`

- [ ] **Step 1: Update use-signin.ts**

Remove the `setAccessToken` import (find the import line that imports from `@/lib/token` or `@/lib/api-client`).

In `handleEmailSignIn`, replace lines 123-125:

```typescript
        } else {
          setAccessToken(res.data.access_token);
          setEmailState({ success: true });
```

With:

```typescript
        } else {
          setEmailState({ success: true });
```

In `handleAccountSignIn`, replace lines 161-162:

```typescript
          setAccessToken(res.data.access_token);
          setAccountState({ success: true });
```

With:

```typescript
          setAccountState({ success: true });
```

- [ ] **Step 2: Update use-signup.ts**

Remove the `setAccessToken` import.

Replace lines 108-110:

```typescript
        } else {
          setAccessToken(res.data.access_token);
          setSignUpState({ success: true });
```

With:

```typescript
        } else {
          setSignUpState({ success: true });
```

- [ ] **Step 3: Update user-profile-menu.tsx**

Remove `removeToken` import from `@/lib/api-client`.

Replace `handleSignOut` (lines 68-76):

```typescript
  async function handleSignOut() {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout API errors — clear local state regardless
    }
    removeToken();
    window.location.href = "/";
  }
```

With:

```typescript
  async function handleSignOut() {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout API errors
    }
    window.location.href = "/";
  }
```

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/auth/hooks/use-signin.ts \
  dx-web/src/features/web/auth/hooks/use-signup.ts \
  dx-web/src/features/web/auth/components/user-profile-menu.tsx
git commit -m "refactor: remove setAccessToken/removeToken from auth hooks and profile menu"
```

---

## Task 13: Rewrite SSE Hooks

**Files:**
- Modify: `dx-web/src/hooks/use-group-sse.ts`
- Modify: `dx-web/src/hooks/use-group-notify.ts`

- [ ] **Step 1: Rewrite use-group-sse.ts**

Replace the entire file:

```typescript
"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useGroupSSE(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  listenersRef.current = listeners;

  useEffect(() => {
    if (!groupId) return;

    const url = `${API_URL}/api/groups/${groupId}/events`;
    const eventSource = new EventSource(url, { withCredentials: true });

    for (const eventName of Object.keys(listenersRef.current)) {
      eventSource.addEventListener(eventName, (e: MessageEvent) => {
        try {
          const data: unknown = JSON.parse(e.data);
          listenersRef.current[eventName]?.(data);
        } catch {
          // Discard malformed SSE messages
        }
      });
    }

    return () => {
      eventSource.close();
    };
  }, [groupId]);
}
```

- [ ] **Step 2: Rewrite use-group-notify.ts**

Replace the entire file:

```typescript
"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useGroupNotify(
  groupId: string | null,
  onUpdate: (scope: string) => void
): void {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  useEffect(() => {
    if (!groupId) return;

    const url = `${API_URL}/api/groups/${groupId}/notify`;
    const eventSource = new EventSource(url, { withCredentials: true });

    eventSource.addEventListener("group_updated", (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data) as { scope: string };
        callbackRef.current(data.scope);
      } catch {
        // Discard malformed SSE messages
      }
    });

    return () => {
      eventSource.close();
    };
  }, [groupId]);
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/hooks/use-group-sse.ts dx-web/src/hooks/use-group-notify.ts
git commit -m "refactor: simplify SSE hooks to use cookie auth, remove token injection"
```

---

## Task 14: Update AI Streaming + Game Shells + Image Uploader

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/helpers/stream-progress.ts`
- Modify: `dx-web/src/features/web/ai-custom/helpers/generate-api.ts`
- Modify: `dx-web/src/features/web/ai-custom/helpers/format-api.ts`
- Modify: `dx-web/src/features/web/play-single/components/game-play-shell.tsx`
- Modify: `dx-web/src/features/web/play-group/components/group-play-shell.tsx`
- Modify: `dx-web/src/features/com/images/hooks/use-image-uploader.ts`

- [ ] **Step 1: Update stream-progress.ts**

Remove `import { getToken } from "@/lib/api-client";` (line 1).

Replace lines 26-32:

```typescript
    const token = getToken();
    const res = await fetch(`${API_URL}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(body),
      signal,
    });
```

With:

```typescript
    const res = await fetch(`${API_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(body),
      signal,
    });
```

- [ ] **Step 2: Update generate-api.ts**

Remove `import { getToken } from "@/lib/api-client";` (line 2).

Replace lines 16-24:

```typescript
    const token = getToken();
    const res = await fetch(`${API_URL}/api/ai-custom/generate-metadata`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ difficulty, keywords }),
    });
```

With:

```typescript
    const res = await fetch(`${API_URL}/api/ai-custom/generate-metadata`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ difficulty, keywords }),
    });
```

- [ ] **Step 3: Update format-api.ts**

Same pattern. Remove `import { getToken } from "@/lib/api-client";` (line 2).

Replace lines 16-24:

```typescript
    const token = getToken();
    const res = await fetch(`${API_URL}/api/ai-custom/format-metadata`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ content, formatType }),
    });
```

With:

```typescript
    const res = await fetch(`${API_URL}/api/ai-custom/format-metadata`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ content, formatType }),
    });
```

- [ ] **Step 4: Update game-play-shell.tsx beforeunload**

Remove `import { getToken } from "@/lib/api-client";` (line 30).

Replace lines 103-116:

```typescript
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
      const token = getToken();
      const headers: Record<string, string> = { "Content-Type": "application/json" };
      if (token) headers["Authorization"] = `Bearer ${token}`;

      fetch(`${apiUrl}/api/play-single/${sid}/sync-playtime`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          game_level_id: lid,
          play_time: getElapsedSeconds(),
        }),
        keepalive: true,
      });
```

With:

```typescript
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

      fetch(`${apiUrl}/api/play-single/${sid}/sync-playtime`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          game_level_id: lid,
          play_time: getElapsedSeconds(),
        }),
        keepalive: true,
      });
```

- [ ] **Step 5: Update group-play-shell.tsx beforeunload**

Remove `import { getToken } from "@/lib/api-client";` (line 33).

Replace lines 195-209:

```typescript
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
      const token = getToken();
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (token) headers["Authorization"] = `Bearer ${token}`;

      fetch(`${apiUrl}/api/play-group/${sid}/sync-playtime`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          play_time: useGroupPlayStore.getState().playTime,
        }),
        keepalive: true,
      });
```

With:

```typescript
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

      fetch(`${apiUrl}/api/play-group/${sid}/sync-playtime`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          play_time: useGroupPlayStore.getState().playTime,
        }),
        keepalive: true,
      });
```

- [ ] **Step 6: Update use-image-uploader.ts**

Remove `import { getToken } from "@/lib/api-client";` (find the import line, likely line 7).

Replace lines 48-55:

```typescript
    uppy.use(XHRUpload, {
      endpoint: `${API_URL}/api/uploads/images`,
      fieldName: "file",
      formData: true,
      headers: (() => {
        const token = getToken();
        return token ? { Authorization: `Bearer ${token}` } : {};
      }) as unknown as Record<string, string>,
```

With:

```typescript
    uppy.use(XHRUpload, {
      endpoint: `${API_URL}/api/uploads/images`,
      fieldName: "file",
      formData: true,
      withCredentials: true,
```

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/ai-custom/helpers/stream-progress.ts \
  dx-web/src/features/web/ai-custom/helpers/generate-api.ts \
  dx-web/src/features/web/ai-custom/helpers/format-api.ts \
  dx-web/src/features/web/play-single/components/game-play-shell.tsx \
  dx-web/src/features/web/play-group/components/group-play-shell.tsx \
  dx-web/src/features/com/images/hooks/use-image-uploader.ts
git commit -m "refactor: replace Bearer token with credentials include in streaming and uploads"
```

---

## Task 15: Update Landing Header

**Files:**
- Modify: `dx-web/src/components/in/landing-header.tsx`

- [ ] **Step 1: Rewrite landing-header.tsx**

The landing page is a server-rendered page. Since `dx_token` is httpOnly (JS can't read it), we need to pass login state from the parent. But `LandingHeader` is used as `"use client"`. The simplest approach: make it accept an `isLoggedIn` prop.

Replace the entire file:

```typescript
import Link from "next/link";
import { GraduationCap, SquareArrowRightEnter } from "lucide-react";
import { MobileNav } from "@/components/in/mobile-nav";

interface LandingHeaderProps {
  isLoggedIn?: boolean;
}

export function LandingHeader({ isLoggedIn = false }: LandingHeaderProps) {
  return (
    <header className="flex h-20 w-full items-center justify-between px-5 lg:px-20">
      <Link href="/" className="flex items-center gap-2.5">
        <GraduationCap className="h-9 w-9 text-teal-600" />
        <span className="text-[22px] font-semibold text-slate-900">斗学</span>
      </Link>
      <nav className="hidden items-center gap-9 lg:flex">
        <Link
          href="/docs"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          文档
        </Link>
        <Link
          href="/features"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          功能
        </Link>
        <Link
          href="/changelog"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          更新日志
        </Link>
        <Link
          href="#faq"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          常见问题
        </Link>
        <Link
          href="#contact"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          联系我们
        </Link>
      </nav>
      <div className="flex items-center gap-3">
        {isLoggedIn ? (
          <Link
            href="/hall"
            className="hidden items-center gap-2 rounded-lg bg-teal-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-teal-700 lg:inline-flex"
          >
            进入学习大厅
            <SquareArrowRightEnter className="h-4 w-4" />
          </Link>
        ) : (
          <>
            <Link
              href="/auth/signin"
              className="hidden rounded-lg border border-slate-300 px-6 py-2.5 text-sm font-medium text-slate-900 hover:bg-slate-50 lg:inline-flex"
            >
              登录
            </Link>
            <Link
              href="/auth/signup"
              className="hidden rounded-lg bg-teal-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-teal-700 lg:inline-flex"
            >
              注册
            </Link>
          </>
        )}
        <MobileNav isLoggedIn={isLoggedIn} />
      </div>
    </header>
  );
}
```

Note: This component is no longer `"use client"` — it's now a server component (no hooks, no state). The `"use client"` directive is removed.

- [ ] **Step 2: Update parent page that uses LandingHeader**

Find the page that renders `<LandingHeader />` and pass the `isLoggedIn` prop:

```typescript
import { cookies } from "next/headers";

// In the page component:
const cookieStore = await cookies();
const isLoggedIn = !!cookieStore.get("dx_token")?.value;

<LandingHeader isLoggedIn={isLoggedIn} />
```

Search for `<LandingHeader` to find the parent page and update it accordingly.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/components/in/landing-header.tsx
git commit -m "refactor: make LandingHeader server component with isLoggedIn prop"
```

---

## Task 16: Final Verification

- [ ] **Step 1: Backend — full build**

Run: `cd dx-api && go build ./...`
Expected: clean

- [ ] **Step 2: Frontend — lint + build**

Run: `cd dx-web && npm run lint && npm run build`
Expected: clean (no references to deleted token.ts or auth-guard.tsx)

- [ ] **Step 3: Search for remaining old references**

```bash
# Backend: no references to deleted helpers
grep -rn "helpers.IssueAccessToken\|helpers.ExtractAuthID\|helpers.GenerateRefreshToken\|helpers.StoreRefreshToken\|helpers.LookupRefreshToken\|helpers.DeleteRefreshToken\|helpers.DeleteUserRefreshTokens\|helpers.ParseJWTUserID\|refresh_token\.\|dx_refresh\|dx_adm_refresh" dx-api/ --include="*.go"

# Frontend: no references to deleted modules
grep -rn "getAccessToken\|setAccessToken\|clearAccessToken\|refreshAccessToken\|getToken\|setToken\|removeToken\|from.*token\|auth-guard\|dx_refresh" dx-web/src/ --include="*.ts" --include="*.tsx"
```

Expected: no matches (or only in comments/docs, which are fine).

- [ ] **Step 4: Commit any remaining fixes**

If grep found issues, fix and commit.

- [ ] **Step 5: Final commit message**

```bash
git log --oneline -15
```

Verify all task commits are present and tell a clear story.
