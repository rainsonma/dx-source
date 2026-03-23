# Email Controller Extraction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract all verification-code email logic from auth and user modules into a dedicated email controller, service, and request layer with new consolidated `/api/email/*` routes.

**Architecture:** Create 3 new files (email controller, service, request) in the existing dx-api structure. Move verification-code logic verbatim from auth_service/user_service into the new email service. Wire new `/api/email/*` routes. Update dx-web API client paths. Remove old code from auth/user modules.

**Tech Stack:** Go/Goravel (backend), Next.js/TypeScript (frontend), Redis (rate limiting + code storage)

**Spec:** `docs/superpowers/specs/2026-03-23-email-controller-extraction-design.md`

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `dx-api/app/services/api/email_service.go` | Verification code logic: rate limit, generate, store, send |
| Create | `dx-api/app/http/requests/api/email_request.go` | `SendCodeRequest` validation struct |
| Create | `dx-api/app/http/controllers/api/email_controller.go` | 3 thin handler methods |
| Modify | `dx-api/routes/api.go` | Add `/api/email/*` group, remove old send-code routes |
| Modify | `dx-api/app/services/api/auth_service.go` | Remove `SendSignUpCode`, `SendSignInCode`, `com` import |
| Modify | `dx-api/app/services/api/user_service.go` | Remove `SendChangeEmailCode`, `com` and `time` imports |
| Modify | `dx-api/app/http/controllers/api/auth_controller.go` | Remove `SendSignUpCode`, `SendSignInCode` methods |
| Modify | `dx-api/app/http/controllers/api/user_controller.go` | Remove `SendEmailCode` method |
| Modify | `dx-api/app/http/requests/api/auth_request.go` | Remove `SendCodeRequest` struct and methods |
| Modify | `dx-api/app/http/requests/api/user_request.go` | Remove `SendEmailCodeRequest` struct and methods |
| Modify | `dx-web/src/lib/api-client.ts` | Update 3 endpoint paths |
| Modify | `dx-web/src/features/web/me/actions/me.action.ts` | Update 1 hardcoded path |

---

### Task 1: Create email service

**Files:**
- Create: `dx-api/app/services/api/email_service.go`

- [ ] **Step 1: Create `email_service.go`**

```go
package api

import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/services/com"

	"github.com/goravel/framework/facades"
)

// SendSignUpCode generates and sends a signup verification code to the given email.
func SendSignUpCode(email string) error {
	key := fmt.Sprintf("signup_code:%s", email)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:signup_code:%s", email), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SendSignInCode generates and sends a signin verification code to the given email.
func SendSignInCode(email string) error {
	key := fmt.Sprintf("signin_code:%s", email)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:signin_code:%s", email), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SendChangeEmailCode sends a verification code for changing email.
func SendChangeEmailCode(userID, email string) error {
	key := fmt.Sprintf("change_email_code:%s", userID)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:change_email_code:%s", userID), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	// Check email not already taken by another user
	var existing models.User
	err = facades.Orm().Query().Where("email", email).Where("id != ?", userID).First(&existing)
	if err == nil && existing.ID != "" {
		return ErrDuplicateEmail
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
```

**Important:** This creates duplicate function names in `package api`. The duplicates are removed in Task 4. Tasks 1-4 are applied as one atomic commit — the Go build will not pass until all removals and route changes are complete.

- [ ] **Step 2: Verify file created**

Run: `ls dx-api/app/services/api/email_service.go`
Expected: file exists

---

### Task 2: Create email request

**Files:**
- Create: `dx-api/app/http/requests/api/email_request.go`

- [ ] **Step 1: Create `email_request.go`**

```go
package api

import "github.com/goravel/framework/contracts/http"

// SendCodeRequest validates email for sending verification codes.
type SendCodeRequest struct {
	Email string `form:"email" json:"email"`
}

func (r *SendCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required|email",
	}
}
func (r *SendCodeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "trim",
	}
}
func (r *SendCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required": "请输入邮箱地址",
		"email.email":    "邮箱地址格式不正确",
	}
}
```

**Important:** This creates a duplicate `SendCodeRequest` type — the old one in `auth_request.go` is removed in Task 4.

- [ ] **Step 2: Verify file created**

Run: `ls dx-api/app/http/requests/api/email_request.go`
Expected: file exists

---

### Task 3: Create email controller

**Files:**
- Create: `dx-api/app/http/controllers/api/email_controller.go`

- [ ] **Step 1: Create `email_controller.go`**

```go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type EmailController struct{}

func NewEmailController() *EmailController {
	return &EmailController{}
}

// SendSignUpCode sends a verification code for signup.
func (c *EmailController) SendSignUpCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendSignUpCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SendSignInCode sends a verification code for signin.
func (c *EmailController) SendSignInCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendSignInCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SendChangeCode sends a verification code for changing email.
func (c *EmailController) SendChangeCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendChangeEmailCode(userID, req.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrRateLimited):
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "该邮箱已注册")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
		}
	}

	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 2: Verify file created**

Run: `ls dx-api/app/http/controllers/api/email_controller.go`
Expected: file exists

---

### Task 4: Remove old code and update routes

This task removes the old email code from auth/user modules and updates routes. All steps must be completed together — the build cannot pass at intermediate steps.

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go`
- Modify: `dx-api/app/services/api/user_service.go`
- Modify: `dx-api/app/http/controllers/api/auth_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_controller.go`
- Modify: `dx-api/app/http/requests/api/auth_request.go`
- Modify: `dx-api/app/http/requests/api/user_request.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Remove `SendSignUpCode` and `SendSignInCode` from `auth_service.go`**

Remove the `SendSignUpCode` function (lines 18-40) and `SendSignInCode` function (lines 138-160). Remove the `com` import (`"dx-api/app/services/com"`). The `time` import stays — `issueSession` uses `time.Duration` on line 53. The `helpers` import stays (used by other functions).

- [ ] **Step 2: Remove `SendChangeEmailCode` from `user_service.go`**

Remove the `SendChangeEmailCode` function (lines 159-188). Remove the `com` import (`"dx-api/app/services/com"`). Also remove `time` import — no longer used anywhere in user_service.go after this removal.

- [ ] **Step 3: Remove `SendSignUpCode` and `SendSignInCode` methods from `auth_controller.go`**

Remove the `SendSignUpCode` handler method (lines 25-40) and `SendSignInCode` handler method (lines 71-86). All imports stay — remaining methods still use them.

- [ ] **Step 4: Remove `SendEmailCode` method from `user_controller.go`**

Remove the `SendEmailCode` handler method (lines 89-113). All imports stay — remaining methods still use them.

- [ ] **Step 5: Remove `SendCodeRequest` from `auth_request.go`**

Remove the `SendCodeRequest` doc comment and struct with its 4 methods: `Authorize`, `Rules`, `Filters`, `Messages` (lines 5-26). The remaining `SignUpRequest` and `SignInRequest` structs still use the `http` import.

- [ ] **Step 6: Remove `SendEmailCodeRequest` from `user_request.go`**

Remove the `SendEmailCodeRequest` doc comment and struct with its 4 methods: `Authorize`, `Rules`, `Filters`, `Messages` (lines 52-74). The remaining structs still use the `http` import.

- [ ] **Step 7: Add email controller instantiation in `routes/api.go`**

Add after the `authController` instantiation (around line 18):

```go
emailController := apicontrollers.NewEmailController()
```

- [ ] **Step 8: Add `/api/email/*` public routes in `routes/api.go`**

Inside the `/api` prefix group, add public email routes after the auth routes block (around line 67):

```go
// Email verification code routes (public)
router.Prefix("/email").Group(func(email route.Router) {
	email.Post("/send-signup-code", emailController.SendSignUpCode)
	email.Post("/send-signin-code", emailController.SendSignInCode)
})
```

- [ ] **Step 9: Add `/api/email/*` protected route in `routes/api.go`**

Inside the protected routes group (after `middleware.JwtAuth()`), add the change-email route:

```go
// Email verification code route (protected)
protected.Post("/email/send-change-code", emailController.SendChangeCode)
```

- [ ] **Step 10: Remove old send-code routes from `routes/api.go`**

Remove these 2 lines from the auth routes block:
```go
auth.Post("/signup/send-code", authController.SendSignUpCode)
auth.Post("/signin/send-code", authController.SendSignInCode)
```

Remove this line from the user profile routes block:
```go
user.Post("/email/send-code", userController.SendEmailCode)
```

- [ ] **Step 11: Verify Go build passes**

Run: `cd dx-api && go build ./...`
Expected: no errors

- [ ] **Step 12: Commit backend changes**

```bash
git add dx-api/app/services/api/email_service.go \
       dx-api/app/http/requests/api/email_request.go \
       dx-api/app/http/controllers/api/email_controller.go \
       dx-api/app/services/api/auth_service.go \
       dx-api/app/services/api/user_service.go \
       dx-api/app/http/controllers/api/auth_controller.go \
       dx-api/app/http/controllers/api/user_controller.go \
       dx-api/app/http/requests/api/auth_request.go \
       dx-api/app/http/requests/api/user_request.go \
       dx-api/routes/api.go
git commit -m "refactor: extract email verification into dedicated controller, service, and request"
```

---

### Task 5: Update frontend API paths

**Files:**
- Modify: `dx-web/src/lib/api-client.ts`
- Modify: `dx-web/src/features/web/me/actions/me.action.ts`

- [ ] **Step 1: Update `api-client.ts` — signup code path**

In `authApi.sendSignUpCode` (line 170), change:
```typescript
// old
return apiClient.post<null>("/api/auth/signup/send-code", { email });
// new
return apiClient.post<null>("/api/email/send-signup-code", { email });
```

- [ ] **Step 2: Update `api-client.ts` — signin code path**

In `authApi.sendSignInCode` (line 186), change:
```typescript
// old
return apiClient.post<null>("/api/auth/signin/send-code", { email });
// new
return apiClient.post<null>("/api/email/send-signin-code", { email });
```

- [ ] **Step 3: Update `api-client.ts` — change email code path**

In `userApi.sendEmailCode` (line 234), change:
```typescript
// old
return apiClient.post<null>("/api/user/email/send-code", { email });
// new
return apiClient.post<null>("/api/email/send-change-code", { email });
```

- [ ] **Step 4: Update `me.action.ts` — hardcoded path**

In `sendEmailCodeAction` (line 65), change:
```typescript
// old
const res = await apiClient.post("/api/user/email/send-code", { email: parsed.data.email });
// new
const res = await apiClient.post("/api/email/send-change-code", { email: parsed.data.email });
```

- [ ] **Step 5: Verify frontend build passes**

Run: `cd dx-web && npm run build`
Expected: no errors

- [ ] **Step 6: Commit frontend changes**

```bash
git add dx-web/src/lib/api-client.ts dx-web/src/features/web/me/actions/me.action.ts
git commit -m "refactor: update email API paths to new /api/email/* endpoints"
```

---

### Task 6: Final verification

- [ ] **Step 1: Verify Go build**

Run: `cd dx-api && go build ./...`
Expected: no errors

- [ ] **Step 2: Verify no leftover references to old paths**

Run in dx-api: `grep -r "signup/send-code\|signin/send-code\|email/send-code" routes/ app/`
Expected: no matches

Run in dx-web: `grep -r "auth/signup/send-code\|auth/signin/send-code\|user/email/send-code" src/`
Expected: no matches

- [ ] **Step 3: Verify new service functions exist only once**

Run: `grep -r "func SendSignUpCode\|func SendSignInCode\|func SendChangeEmailCode" dx-api/app/services/`
Expected: each function appears exactly once, all in `email_service.go`

- [ ] **Step 4: Verify no orphaned `com` imports**

Run: `grep -l '"dx-api/app/services/com"' dx-api/app/services/api/`
Expected: only `email_service.go`

- [ ] **Step 5: Run Go tests**

Run: `cd dx-api && go test -race ./...`
Expected: all tests pass
