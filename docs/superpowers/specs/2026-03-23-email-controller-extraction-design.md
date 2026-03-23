# Email Controller Extraction Design

Extract all verification-code email logic from auth and user modules into a dedicated email controller, service, and request layer.

## Goal

Consolidate scattered email-sending code into a single, dedicated module without breaking existing functionality.

## New Files (dx-api)

| File | Purpose |
|------|---------|
| `app/http/controllers/api/email_controller.go` | 3 handler methods for send-code endpoints |
| `app/http/requests/api/email_request.go` | `SendCodeRequest` struct (`email: required\|email`) |
| `app/services/api/email_service.go` | Verification code logic (rate limit, Redis, send) |

**Note:** `SendCodeRequest` in `email_request.go` uses the same name as the existing struct in `auth_request.go`. Both live in `package api`. The old struct must be removed in the same commit to avoid a compile error.

## Route Changes

Old routes removed, new `/api/email/*` group added:

| Old path | New path | Auth |
|----------|----------|------|
| `POST /api/auth/signup/send-code` | `POST /api/email/send-signup-code` | None |
| `POST /api/auth/signin/send-code` | `POST /api/email/send-signin-code` | None |
| `POST /api/user/email/send-code` | `POST /api/email/send-change-code` | JWT |

The `/api/email/*` group requires two sub-groups in `routes/api.go`:
- Public (no auth): `send-signup-code`, `send-signin-code`
- Protected (JWT middleware): `send-change-code`

## Email Controller

`EmailController` with 3 thin methods:

- `SendSignUpCode(ctx)` — validate `SendCodeRequest`, call service, handle `ErrRateLimited` / generic error
- `SendSignInCode(ctx)` — validate `SendCodeRequest`, call service, handle `ErrRateLimited` / generic error
- `SendChangeCode(ctx)` — extract userID from JWT, validate `SendCodeRequest`, call service, handle `ErrRateLimited` / `ErrDuplicateEmail` / generic error

## Email Request

Single `SendCodeRequest` struct used by all 3 endpoints:

```go
type SendCodeRequest struct {
    Email string `form:"email" json:"email"`
}
// Rules: email -> required|email
// Filters: email -> trim
// Messages:
//   email.required -> "请输入邮箱地址"
//   email.email    -> "邮箱地址格式不正确"
```

## Email Service

3 functions moved from auth_service.go and user_service.go:

### SendSignUpCode(email string) error

- Rate limit: `rate:signup_code:{email}` (1 per 60s)
- Redis key: `signup_code:{email}` (300s TTL)
- Calls `com.SendVerificationEmail(email, code)`

### SendSignInCode(email string) error

- Rate limit: `rate:signin_code:{email}` (1 per 60s)
- Redis key: `signin_code:{email}` (300s TTL)
- Calls `com.SendVerificationEmail(email, code)`

### SendChangeEmailCode(userID, email string) error

- Rate limit: `rate:change_email_code:{userID}` (1 per 60s)
- Checks email uniqueness (not taken by another user)
- Redis key: `change_email_code:{userID}` (300s TTL)
- Calls `com.SendVerificationEmail(email, code)`

## Unchanged

- `app/services/com/email_service.go` — stays as the low-level mail sender
- `app/services/api/errors.go` — error sentinels (`ErrRateLimited`, `ErrDuplicateEmail`) reused by the new email service

## Cleanup (removed from old files)

| File | Removed |
|------|---------|
| `app/services/api/auth_service.go` | `SendSignUpCode()`, `SendSignInCode()`, `com` import |
| `app/services/api/user_service.go` | `SendChangeEmailCode()`, `com` import |
| `app/http/controllers/api/auth_controller.go` | `SendSignUpCode()`, `SendSignInCode()` methods |
| `app/http/controllers/api/user_controller.go` | `SendEmailCode()` method |
| `app/http/requests/api/auth_request.go` | `SendCodeRequest` struct and all its methods |
| `app/http/requests/api/user_request.go` | `SendEmailCodeRequest` struct and all its methods |
| `routes/api.go` | Old send-code route lines |

## Frontend Changes (dx-web)

Update 3 endpoint paths in `src/lib/api-client.ts`:

| API method | Old path | New path |
|------------|----------|----------|
| `authApi.sendSignUpCode` | `/api/auth/signup/send-code` | `/api/email/send-signup-code` |
| `authApi.sendSignInCode` | `/api/auth/signin/send-code` | `/api/email/send-signin-code` |
| `userApi.sendEmailCode` | `/api/user/email/send-code` | `/api/email/send-change-code` |

Additionally update 1 hardcoded path in `src/features/web/me/actions/me.action.ts`:

| Line | Old path | New path |
|------|----------|----------|
| 65 | `/api/user/email/send-code` | `/api/email/send-change-code` |

## Verification

- `go build ./...` passes
- `npm run build` passes in dx-web
- All 3 new endpoints return same responses as old ones
- Redis keys and rate-limit keys unchanged — existing codes in-flight still work
- Auth signup/signin/change-email flows work end-to-end
