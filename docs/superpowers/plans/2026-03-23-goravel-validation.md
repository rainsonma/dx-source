# Goravel Form Request Validation Migration

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate all manual controller validation to Goravel Form Request pattern with unified error response.

**Architecture:** Add `Rules()` / `Authorize()` (+ optional `Messages()` / `PrepareForValidation()`) to request structs. Replace `Bind()` + manual checks with `ValidateRequest()` via a shared `Validate` helper. All validation errors use `consts.CodeValidationError` (40000). Domain-specific error codes (40001-40008) remain in service-layer error handling.

**Tech Stack:** Goravel v1.17.2, Go

**Goravel FormRequest interface** (from `contracts/http/request.go`):
```go
type FormRequest interface {
    Authorize(ctx Context) error
    Rules(ctx Context) map[string]string
}
// Optional: FormRequestWithMessages, FormRequestWithPrepareForValidation
```

**Controller pattern** — every controller method that binds a request body becomes:
```go
var req requests.SomeRequest
if resp := helpers.Validate(ctx, &req); resp != nil {
    return resp
}
// req fields are populated and validated — proceed to service call
```

**Important constraints:**
- Goravel returns `"rules can't be empty"` error when `Rules()` returns nil/empty map — structs with no validation rules must keep using `Bind()`, NOT `ValidateRequest()`
- Validation error codes change from domain-specific (40001 `CodeInvalidEmail`, 40005 `CodeInvalidCode`) to generic (40000 `CodeValidationError`) — verify frontend doesn't depend on these specific codes
- `required` on slice fields (`[]string`, `[]struct`) may not catch empty slices — use `min_len:1` as well
- `in` rule with int types needs runtime verification

**Scope:**
- 15 request files (13 api + 2 adm), ~35 structs with validation rules
- 14 controller files to update (12 api + 2 adm)
- Route/query param validation stays in controllers (not request body)
- Query-param-bound methods (`CheckActive`, `CheckActiveLevel`, `Restore`) stay with manual validation
- Structs with no validation rules (`CompleteLevelRequest`, `AdvanceLevelRequest`, `RestartLevelRequest`, `UpdateContentItemRequest`) keep using `Bind()`
- File upload file handling stays in controller; only `role` field migrates to FormRequest
- Business logic error codes stay in services

---

### Task 1: Add Validate Helper

**Files:**
- Modify: `dx-api/app/helpers/response.go`

- [ ] **Step 1: Add Validate helper function**

Add to `response.go`:

```go
import (
    nethttp "net/http"
    contractshttp "github.com/goravel/framework/contracts/http"
    "douxue/app/consts"
)

// Validate binds and validates a form request, returning an error response on failure.
func Validate(ctx contractshttp.Context, req contractshttp.FormRequest) contractshttp.Response {
    errors, err := ctx.Request().ValidateRequest(req)
    if err != nil {
        return Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
    }
    if errors != nil {
        return Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, errors.One())
    }
    return nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 3: Commit**

```
refactor: add Validate helper for Goravel form request validation
```

---

### Task 2: Auth Requests + Controller

**Files:**
- Modify: `dx-api/app/http/requests/api/auth_request.go`
- Modify: `dx-api/app/http/controllers/api/auth_controller.go`

- [ ] **Step 1: Add validation methods to auth request structs**

```go
import "github.com/goravel/framework/contracts/http"

// --- SendCodeRequest ---
func (r *SendCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendCodeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "email": "required",
    }
}

// --- SignUpRequest ---
func (r *SignUpRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignUpRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "email": "required",
        "code":  "required|len:6",
    }
}
func (r *SignUpRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "code.required": "a 6-digit verification code is required",
        "code.len":      "a 6-digit verification code is required",
    }
}

// --- SignInRequest ---
// OR logic: either (email + code) or (account + password)
func (r *SignInRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignInRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "email":    "required_without:account",
        "code":     "required_with:email|len:6",
        "account":  "required_without:email",
        "password": "required_with:account",
    }
}
func (r *SignInRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "email.required_without":   "email or account is required",
        "code.required_with":       "verification code is required",
        "code.len":                 "a 6-digit verification code is required",
        "account.required_without": "email or account is required",
        "password.required_with":   "password is required",
    }
}
```

- [ ] **Step 2: Update auth controller**

Replace each `Bind()` + manual validation block with `helpers.Validate(ctx, &req)`.

Methods to update:
- `SendSignUpCode()` — replace Bind + email check
- `SignUp()` — replace Bind + email/code checks
- `SendSignInCode()` — replace Bind + email check (reuses SendCodeRequest)
- `SignIn()` — replace Bind + OR logic checks

Keep: all service-layer error handling (`CodeDuplicateEmail`, `CodeCodeExpired`, `CodeRateLimited`, etc.)

- [ ] **Step 3: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 4: Commit**

```
refactor: migrate auth requests to Goravel form validation
```

---

### Task 3: User Requests + Controller

**Files:**
- Modify: `dx-api/app/http/requests/api/user_request.go`
- Modify: `dx-api/app/http/controllers/api/user_controller.go`

- [ ] **Step 1: Add validation methods**

```go
import "github.com/goravel/framework/contracts/http"

// --- UpdateProfileRequest ---
// All fields optional, but length-limited when present
func (r *UpdateProfileRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateProfileRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "nickname":     "max_len:20",
        "city":         "max_len:50",
        "introduction": "max_len:200",
    }
}
func (r *UpdateProfileRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "nickname.max_len":     "nickname must be at most 20 characters",
        "city.max_len":         "city must be at most 50 characters",
        "introduction.max_len": "introduction must be at most 200 characters",
    }
}

// --- UpdateAvatarRequest ---
func (r *UpdateAvatarRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateAvatarRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "image_id": "required",
    }
}

// --- SendEmailCodeRequest ---
func (r *SendEmailCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendEmailCodeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "email": "required",
    }
}

// --- ChangeEmailRequest ---
func (r *ChangeEmailRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangeEmailRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "email": "required",
        "code":  "required|len:6",
    }
}
func (r *ChangeEmailRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "code.required": "a 6-digit verification code is required",
        "code.len":      "a 6-digit verification code is required",
    }
}

// --- ChangePasswordRequest ---
func (r *ChangePasswordRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangePasswordRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "current_password": "required",
        "new_password":     "required|min_len:8",
    }
}
func (r *ChangePasswordRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "new_password.min_len": "new password must be at least 8 characters",
    }
}
```

- [ ] **Step 2: Update user controller**

Methods to update: `UpdateProfile()`, `UpdateAvatar()`, `SendEmailCode()`, `ChangeEmail()`, `ChangePassword()`

Replace Bind + manual checks → `helpers.Validate(ctx, &req)`. Keep service error handling.

- [ ] **Step 3: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 4: Commit**

```
refactor: migrate user requests to Goravel form validation
```

---

### Task 4: Session Requests + Controller

**Files:**
- Modify: `dx-api/app/http/requests/api/session_request.go`
- Modify: `dx-api/app/http/controllers/api/game_session_controller.go`

- [ ] **Step 1: Add validation methods to session request structs**

```go
import (
    "github.com/goravel/framework/contracts/http"
    "github.com/goravel/framework/contracts/validation"
    "douxue/app/consts"
)

// --- StartSessionRequest ---
func (r *StartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartSessionRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_id": "required",
    }
}
func (r *StartSessionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
    degree, _ := data.Get("degree")
    if degree == nil || degree == "" {
        data.Set("degree", consts.GameDegreeIntermediate)
    }
    return nil
}

// --- CheckActiveSessionRequest --- SKIP: uses Query() not Bind()
// --- CheckActiveLevelSessionRequest --- SKIP: uses Query() not Bind()

// --- StartLevelRequest ---
func (r *StartLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartLevelRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_level_id": "required",
    }
}
func (r *StartLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
    degree, _ := data.Get("degree")
    if degree == nil || degree == "" {
        data.Set("degree", consts.GameDegreeIntermediate)
    }
    return nil
}

// --- CompleteLevelRequest --- SKIP: no validation, keep Bind()
// --- AdvanceLevelRequest --- SKIP: no validation + route param fallback, keep Bind()
// --- RestartLevelRequest --- SKIP: uses route params, no Bind()

// --- RecordAnswerRequest ---
func (r *RecordAnswerRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordAnswerRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_session_level_id": "required",
        "game_level_id":         "required",
        "content_item_id":       "required",
    }
}

// --- RecordSkipRequest ---
func (r *RecordSkipRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordSkipRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_level_id": "required",
    }
}

// --- SyncPlayTimeRequest ---
func (r *SyncPlayTimeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SyncPlayTimeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_level_id": "required",
    }
}

// --- UpdateContentItemRequest --- SKIP: no validation, keep Bind()

// --- EndSessionRequest ---
func (r *EndSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *EndSessionRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_id": "required",
    }
}

// --- RestoreSessionRequest --- SKIP: uses Query() not Bind()
```

- [ ] **Step 2: Update game_session_controller**

Methods to migrate to `helpers.Validate(ctx, &req)`:
- `Start()`, `End()`, `StartLevel()`, `RecordAnswer()`, `RecordSkip()`, `SyncPlayTime()`

Methods to keep as-is (use `Bind()` or `Query()`):
- `CompleteLevel()`, `AdvanceLevel()`, `RestartLevel()`, `UpdateContentItem()` — no validation rules (nil Rules would crash)
- `CheckActive()`, `CheckActiveLevel()`, `CheckAnyActive()`, `Restore()` — use `ctx.Request().Query()`, not body binding

- [ ] **Step 3: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 4: Commit**

```
refactor: migrate session requests to Goravel form validation
```

---

### Task 5: Entity Requests + Controllers (upload, favorite, collections, report)

**Files:**
- Modify: `dx-api/app/http/requests/api/upload_request.go`
- Modify: `dx-api/app/http/requests/api/user_favorite_request.go`
- Modify: `dx-api/app/http/requests/api/user_master_request.go`
- Modify: `dx-api/app/http/requests/api/user_review_request.go`
- Modify: `dx-api/app/http/requests/api/user_unknown_request.go`
- Modify: `dx-api/app/http/requests/api/game_report_request.go`
- Modify: `dx-api/app/http/controllers/api/upload_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_favorite_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_master_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_review_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_unknown_controller.go`
- Modify: `dx-api/app/http/controllers/api/game_report_controller.go`

- [ ] **Step 1: Add validation to upload_request.go**

```go
func (r *UploadImageRequest) Authorize(ctx http.Context) error { return nil }
func (r *UploadImageRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "role": "required",
    }
}
```

- [ ] **Step 2: Add validation to user_favorite_request.go**

```go
func (r *ToggleFavoriteRequest) Authorize(ctx http.Context) error { return nil }
func (r *ToggleFavoriteRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_id": "required",
    }
}
```

- [ ] **Step 3: Add validation to user_master_request.go**

```go
func (r *MarkMasteredRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkMasteredRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "content_item_id": "required",
        "game_id":         "required",
        "game_level_id":   "required",
    }
}

func (r *BulkDeleteRequest) Authorize(ctx http.Context) error { return nil }
func (r *BulkDeleteRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "ids": "required|min_len:1",
    }
}
```

- [ ] **Step 4: Add validation to user_review_request.go**

```go
func (r *MarkReviewRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkReviewRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "content_item_id": "required",
        "game_id":         "required",
        "game_level_id":   "required",
    }
}
```

- [ ] **Step 5: Add validation to user_unknown_request.go**

```go
func (r *MarkUnknownRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkUnknownRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "content_item_id": "required",
        "game_id":         "required",
        "game_level_id":   "required",
    }
}
```

- [ ] **Step 6: Add validation to game_report_request.go**

```go
func (r *SubmitReportRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitReportRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "game_id":         "required",
        "game_level_id":   "required",
        "content_item_id": "required",
        "reason":          "required",
    }
}
```

- [ ] **Step 7: Update all 6 controllers**

Replace `Bind()` + manual checks → `helpers.Validate(ctx, &req)` in:
- `upload_controller.go`: `UploadImage()` — switch from `ctx.Request().Input("role")` to `req.Role` after validation; keep file handling (`ctx.Request().File()`) in controller
- `user_favorite_controller.go`: `ToggleFavorite()`
- `user_master_controller.go`: `MarkMastered()`, `BulkDeleteMastered()`
- `user_review_controller.go`: `MarkReview()`, `BulkDeleteReviews()`
- `user_unknown_controller.go`: `MarkUnknown()`, `BulkDeleteUnknown()`
- `game_report_controller.go`: `SubmitReport()`

**Note:** `BulkDeleteReviews()` and `BulkDeleteUnknown()` reuse `BulkDeleteRequest` from `user_master_request.go`. Verify they import from the correct package.

- [ ] **Step 8: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 9: Commit**

```
refactor: migrate entity requests to Goravel form validation
```

---

### Task 6: Misc Requests + Controllers (feedback, redeem, content_seek)

**Files:**
- Modify: `dx-api/app/http/requests/api/feedback_request.go`
- Modify: `dx-api/app/http/requests/api/user_redeem_request.go`
- Modify: `dx-api/app/http/requests/api/content_seek_request.go`
- Modify: `dx-api/app/http/controllers/api/feedback_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_redeem_controller.go`
- Modify: `dx-api/app/http/controllers/api/content_seek_controller.go`

- [ ] **Step 1: Add validation to feedback_request.go**

```go
func (r *SubmitFeedbackRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitFeedbackRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "type":        "required",
        "description": "required|max_len:200",
    }
}
func (r *SubmitFeedbackRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "description.max_len": "description must be at most 200 characters",
    }
}
```

- [ ] **Step 2: Add validation to user_redeem_request.go**

```go
func (r *RedeemCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *RedeemCodeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "code": "required|len:19",
    }
}
func (r *RedeemCodeRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "code.len": "invalid redeem code format",
    }
}
```

- [ ] **Step 3: Add validation to content_seek_request.go**

```go
func (r *SubmitContentSeekRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitContentSeekRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "course_name":  "required|max_len:30",
        "description":  "required|max_len:30",
        "disk_url":     "required|max_len:30",
    }
}
func (r *SubmitContentSeekRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "course_name.max_len":  "course name must be at most 30 characters",
        "description.max_len":  "description must be at most 30 characters",
        "disk_url.max_len":     "disk url must be at most 30 characters",
    }
}
```

- [ ] **Step 4: Update controllers**

Replace Bind + manual checks → `helpers.Validate(ctx, &req)` in:
- `feedback_controller.go`: `SubmitFeedback()`
- `user_redeem_controller.go`: `RedeemCode()`
- `content_seek_controller.go`: `SubmitContentSeek()`

- [ ] **Step 5: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 6: Commit**

```
refactor: migrate misc requests to Goravel form validation
```

---

### Task 7: Course Game Requests + Controllers

**Files:**
- Modify: `dx-api/app/http/requests/api/course_game_request.go`
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go`

- [ ] **Step 1: Add validation to course_game_request.go**

```go
import "github.com/goravel/framework/contracts/http"

// --- CreateGameRequest ---
func (r *CreateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateGameRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "name":           "required",
        "gameMode":       "required",
        "gameCategoryId": "required",
        "gamePressId":    "required",
    }
}

// --- UpdateGameRequest ---
func (r *UpdateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateGameRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "name": "required",
    }
}

// --- CreateLevelRequest ---
func (r *CreateLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateLevelRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "name": "required",
    }
}

// --- SaveMetadataBatchRequest ---
func (r *SaveMetadataBatchRequest) Authorize(ctx http.Context) error { return nil }
func (r *SaveMetadataBatchRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "entries": "required|min_len:1",
    }
}

// --- ReorderMetadataRequest ---
func (r *ReorderMetadataRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderMetadataRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "metaId": "required",
    }
}

// --- InsertContentItemRequest ---
func (r *InsertContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *InsertContentItemRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "contentMetaId": "required",
    }
}

// --- UpdateContentItemTextRequest --- SKIP: no validation, keep Bind()

// --- ReorderContentItemRequest ---
func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "itemId": "required",
    }
}
```

- [ ] **Step 2: Update course_game_controller**

Replace Bind + manual checks → `helpers.Validate(ctx, &req)` in:
- `Create()`, `Update()`, `CreateLevel()`, `SaveMetadata()`, `ReorderMetadata()`, `InsertContentItem()`, `ReorderContentItems()`

Methods to keep as-is: `UpdateContentItemText()` — no validation rules (nil Rules would crash), keep `Bind()`

**Keep in controller:** route param validation (`gameID == ""`, `levelID == ""`, `itemID == ""`) — these are path params, not body fields.

- [ ] **Step 3: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 4: Commit**

```
refactor: migrate course game requests to Goravel form validation
```

---

### Task 8: Admin Requests + Controllers

**Files:**
- Modify: `dx-api/app/http/requests/adm/notice_request.go`
- Modify: `dx-api/app/http/requests/adm/redeem_request.go`
- Modify: `dx-api/app/http/controllers/adm/notice_controller.go`
- Modify: `dx-api/app/http/controllers/adm/redeem_controller.go`

- [ ] **Step 1: Add validation to adm/notice_request.go**

```go
import "github.com/goravel/framework/contracts/http"

// --- CreateNoticeRequest ---
func (r *CreateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateNoticeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "title": "required|max_len:200",
    }
}
func (r *CreateNoticeRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "title.max_len": "title must be at most 200 characters",
    }
}

// --- UpdateNoticeRequest ---
func (r *UpdateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateNoticeRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "title": "required|max_len:200",
    }
}
func (r *UpdateNoticeRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "title.max_len": "title must be at most 200 characters",
    }
}
```

- [ ] **Step 2: Add validation to adm/redeem_request.go**

```go
import "github.com/goravel/framework/contracts/http"

// --- GenerateCodesRequest ---
func (r *GenerateCodesRequest) Authorize(ctx http.Context) error { return nil }
func (r *GenerateCodesRequest) Rules(ctx http.Context) map[string]string {
    return map[string]string{
        "grade": "required|in:month,season,year,lifetime",
        "count": "required|in:10,50,100,500",
    }
}
func (r *GenerateCodesRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "grade.in": "invalid grade",
        "count.in": "count must be 10, 50, 100, or 500",
    }
}
```

**MANDATORY verification:** Test that `in:10,50,100,500` works with int type for `count`. JSON-decoded ints become `float64` internally — if the `in` rule fails, add `PrepareForValidation` to cast `count` to int, or validate `count` manually in the controller.

- [ ] **Step 3: Update admin controllers**

Replace Bind + manual checks → `helpers.Validate(ctx, &req)` in:
- `adm/notice_controller.go`: `CreateNotice()`, `UpdateNotice()`
- `adm/redeem_controller.go`: `GenerateCodes()`

**Keep in controller:** route param validation (`id == ""`)

- [ ] **Step 4: Verify build + test**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 5: Commit**

```
refactor: migrate admin requests to Goravel form validation
```

---

### Task 9: Final Verification

- [ ] **Step 1: Full build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 2: Full test suite**

Run: `cd dx-api && go test -race ./...`

- [ ] **Step 3: Go vet**

Run: `cd dx-api && go vet ./...`

- [ ] **Step 4: Verify unused request structs**

Check if `game_request.go` structs (ListGamesRequest, SearchGamesRequest, LevelContentRequest) are used anywhere. If not, they can remain as plain structs (no validation needed since they're for query params or unused).

Run: `grep -r "ListGamesRequest\|SearchGamesRequest\|LevelContentRequest" dx-api/`

- [ ] **Step 5: Final commit (if any remaining fixes)**

```
chore: final cleanup after validation migration
```

---

## Out of Scope (follow-up)

These items were intentionally excluded from this migration:

1. **Admin auth controller** (`adm/auth_controller.go` Login) — no existing request struct; create `adm/auth_request.go` as a separate task
2. **AI custom controller** (`ai_custom_controller.go`) — has inline validation but no request struct file; create `ai_custom_request.go` as a separate task
3. **GET endpoint query param validation** — leaderboard type/period, hall year, game search query; these use `ctx.Request().Query()` not body binding
4. **Adding new validation rules** — only existing manual validation was migrated; consider adding rules for currently unvalidated fields (e.g., SignUp username/password) as a separate task
