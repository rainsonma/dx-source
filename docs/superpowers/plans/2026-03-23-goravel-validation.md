# Request Validation Improvement Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Strengthen all 44 request structs with proper validation rules (UUID, email, enum, numeric ranges, string lengths), custom messages (Chinese for api/, English for adm/), and input filters (trim).

**Architecture:** Create an `InEnum()` helper that generates `in:val1,val2,...` rule strings from `consts/` values. Create one custom rule (`strong_password`). Upgrade all existing FormRequest structs with richer rules/messages/filters. Convert 10 plain structs to full FormRequest implementations. Update 2 controllers to use `helpers.Validate()` instead of manual parsing.

**Tech Stack:** Goravel v1.17.2, Go 1.26

**Spec:** `docs/superpowers/specs/2026-03-23-request-validation-design.md`

---

### Task 1: Create InEnum helper

**Files:**
- Create: `dx-api/app/helpers/enum_rules.go`
- Test: `dx-api/app/helpers/enum_rules_test.go`

- [ ] **Step 1: Write test**

```go
// dx-api/app/helpers/enum_rules_test.go
package helpers

import (
	"strings"
	"testing"
)

func TestInEnum_KnownEnum(t *testing.T) {
	result := InEnum("degree")
	if !strings.HasPrefix(result, "in:") {
		t.Fatalf("expected 'in:' prefix, got %q", result)
	}
	if !strings.Contains(result, "intermediate") {
		t.Fatalf("expected 'intermediate' in result, got %q", result)
	}
}

func TestInEnum_UnknownEnum_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown enum")
		}
	}()
	InEnum("nonexistent")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dx-api && go test ./app/helpers/ -run TestInEnum -v`
Expected: FAIL — `InEnum` not defined

- [ ] **Step 3: Write implementation**

```go
// dx-api/app/helpers/enum_rules.go
package helpers

import (
	"strings"

	"dx-api/app/consts"
)

var enumValues = map[string][]string{
	"degree":        {consts.GameDegreePractice, consts.GameDegreeBeginner, consts.GameDegreeIntermediate, consts.GameDegreeAdvanced},
	"pattern":       {consts.GamePatternListen, consts.GamePatternSpeak, consts.GamePatternRead, consts.GamePatternWrite},
	"mode":          {consts.GameModeLSRW, consts.GameModeVocabBattle, consts.GameModeVocabMatch, consts.GameModeVocabElimination, consts.GameModeListeningChallenge},
	"feedback_type": {consts.FeedbackTypeFeature, consts.FeedbackTypeContent, consts.FeedbackTypeUX, consts.FeedbackTypeBug, consts.FeedbackTypeOther},
	"content_type":  {consts.ContentTypeWord, consts.ContentTypeBlock, consts.ContentTypePhrase, consts.ContentTypeSentence},
	"image_role":    {consts.ImageRoleAdmUserAvatar, consts.ImageRoleUserAvatar, consts.ImageRoleCategoryCover, consts.ImageRoleTemplateCover, consts.ImageRoleGameCover, consts.ImageRolePressCover, consts.ImageRoleGameGroupCover, consts.ImageRolePostImage},
	"source_from":   {consts.SourceFromManual, consts.SourceFromAI},
	"source_type":   {consts.SourceTypeSentence, consts.SourceTypeVocab},
}

// InEnum returns an "in:val1,val2,..." rule string for the named enum.
// Panics on unknown enum names to catch typos at startup.
func InEnum(name string) string {
	vals, ok := enumValues[name]
	if !ok {
		panic("unknown enum: " + name)
	}
	return "in:" + strings.Join(vals, ",")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dx-api && go test ./app/helpers/ -run TestInEnum -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/helpers/enum_rules.go app/helpers/enum_rules_test.go
git commit -m "feat: add InEnum helper for validation rule generation"
```

---

### Task 2: Create strong_password custom rule

**Files:**
- Create: `dx-api/app/rules/strong_password.go`
- Create: `dx-api/app/rules/strong_password_test.go`
- Create: `dx-api/bootstrap/rules.go`
- Modify: `dx-api/bootstrap/app.go`

- [ ] **Step 1: Write test**

```go
// dx-api/app/rules/strong_password_test.go
package rules

import (
	"context"
	"testing"
)

func TestStrongPassword_Passes(t *testing.T) {
	rule := &StrongPassword{}

	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"valid", "Abc123!@", true},
		{"missing uppercase", "abc123!@", false},
		{"missing lowercase", "ABC123!@", false},
		{"missing digit", "Abcdef!@", false},
		{"missing special", "Abc12345", false},
		{"empty", "", false},
		{"all types", "P@ssw0rd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Passes(context.Background(), nil, tt.val)
			if got != tt.want {
				t.Errorf("Passes(%q) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestStrongPassword_Signature(t *testing.T) {
	rule := &StrongPassword{}
	if rule.Signature() != "strong_password" {
		t.Errorf("Signature() = %q, want %q", rule.Signature(), "strong_password")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dx-api && go test ./app/rules/ -run TestStrongPassword -v`
Expected: FAIL — package/type not found

- [ ] **Step 3: Write strong_password rule**

```go
// dx-api/app/rules/strong_password.go
package rules

import (
	"context"
	"unicode"

	"github.com/goravel/framework/contracts/validation"
)

type StrongPassword struct{}

func (r *StrongPassword) Signature() string {
	return "strong_password"
}

func (r *StrongPassword) Passes(_ context.Context, _ validation.Data, val any, _ ...any) bool {
	s, ok := val.(string)
	if !ok || s == "" {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func (r *StrongPassword) Message(_ context.Context) string {
	return ":attribute 必须包含大写字母、小写字母、数字和特殊字符"
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dx-api && go test ./app/rules/ -run TestStrongPassword -v`
Expected: PASS

- [ ] **Step 5: Create bootstrap/rules.go and update bootstrap/app.go**

```go
// dx-api/bootstrap/rules.go
package bootstrap

import (
	contractsvalidation "github.com/goravel/framework/contracts/validation"

	"dx-api/app/rules"
)

func Rules() []contractsvalidation.Rule {
	return []contractsvalidation.Rule{
		&rules.StrongPassword{},
	}
}
```

In `dx-api/bootstrap/app.go`, add `.WithRules(Rules)` to the chain — insert it before `.WithProviders(Providers)`:

```go
// Add this line:
		WithRules(Rules).
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 7: Commit**

```bash
cd dx-api && git add app/rules/ bootstrap/rules.go bootstrap/app.go
git commit -m "feat: add strong_password custom validation rule"
```

---

### Task 3: Auth requests

**Files:**
- Modify: `dx-api/app/http/requests/api/auth_request.go`

- [ ] **Step 1: Rewrite auth_request.go**

Replace the entire file content:

```go
package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

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

// SignUpRequest validates signup data.
type SignUpRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func (r *SignUpRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignUpRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "required|email",
		"code":     "required|len:6",
		"username": "required|alpha_dash|min_len:3|max_len:20",
		"password": "required|min_len:8|strong_password",
	}
}
func (r *SignUpRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "trim",
		"username": "trim",
	}
}
func (r *SignUpRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required":          "请输入邮箱地址",
		"email.email":             "邮箱地址格式不正确",
		"code.required":           "请输入6位验证码",
		"code.len":                "请输入6位验证码",
		"username.required":       "请输入用户名",
		"username.alpha_dash":     "用户名只能包含字母、数字、下划线和横线",
		"username.min_len":        "用户名至少需要3个字符",
		"username.max_len":        "用户名不能超过20个字符",
		"password.required":       "请输入密码",
		"password.min_len":        "密码至少需要8个字符",
		"password.strong_password": "密码必须包含大写字母、小写字母、数字和特殊字符",
	}
}

// SignInRequest for signin — supports email+code OR account+password.
// Controller still validates the pairing logic (exactly one auth method).
type SignInRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Account  string `form:"account" json:"account"`
	Password string `form:"password" json:"password"`
}

func (r *SignInRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignInRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "required_without:account|email",
		"code":     "required_without:password|len:6",
		"account":  "required_without:email|min_len:3",
		"password": "required_without:code|min_len:8",
	}
}
func (r *SignInRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email":   "trim",
		"account": "trim",
	}
}
func (r *SignInRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.email":       "邮箱地址格式不正确",
		"code.len":          "请输入6位验证码",
		"account.min_len":   "账号至少需要3个字符",
		"password.min_len":  "密码至少需要8个字符",
	}
}

```

Note: auth_request.go only needs `"github.com/goravel/framework/contracts/http"` — no `validation` import needed since it doesn't use `PrepareForValidation`.

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/http/requests/api/auth_request.go
git commit -m "feat: strengthen auth request validation rules"
```

---

### Task 4: User requests

**Files:**
- Modify: `dx-api/app/http/requests/api/user_request.go`

- [ ] **Step 1: Rewrite user_request.go**

Replace the entire file content:

```go
package api

import "github.com/goravel/framework/contracts/http"

// UpdateProfileRequest validates profile update data.
type UpdateProfileRequest struct {
	Nickname     string `form:"nickname" json:"nickname"`
	City         string `form:"city" json:"city"`
	Introduction string `form:"introduction" json:"introduction"`
}

func (r *UpdateProfileRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateProfileRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname":     "max_len:20",
		"city":         "max_len:50",
		"introduction": "max_len:200",
	}
}
func (r *UpdateProfileRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname":     "trim",
		"city":         "trim",
		"introduction": "trim",
	}
}
func (r *UpdateProfileRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.max_len":     "昵称不能超过20个字符",
		"city.max_len":         "城市不能超过50个字符",
		"introduction.max_len": "简介不能超过200个字符",
	}
}

// UpdateAvatarRequest validates avatar update data.
type UpdateAvatarRequest struct {
	ImageID string `form:"image_id" json:"image_id"`
}

func (r *UpdateAvatarRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateAvatarRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"image_id": "required|uuid",
	}
}
func (r *UpdateAvatarRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"image_id.required": "请选择头像",
		"image_id.uuid":     "无效的图片ID",
	}
}

// SendEmailCodeRequest validates email code sending data.
type SendEmailCodeRequest struct {
	Email string `form:"email" json:"email"`
}

func (r *SendEmailCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendEmailCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required|email",
	}
}
func (r *SendEmailCodeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "trim",
	}
}
func (r *SendEmailCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required": "请输入邮箱地址",
		"email.email":    "邮箱地址格式不正确",
	}
}

// ChangeEmailRequest validates email change data.
type ChangeEmailRequest struct {
	Email string `form:"email" json:"email"`
	Code  string `form:"code" json:"code"`
}

func (r *ChangeEmailRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangeEmailRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required|email",
		"code":  "required|len:6",
	}
}
func (r *ChangeEmailRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "trim",
	}
}
func (r *ChangeEmailRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required": "请输入邮箱地址",
		"email.email":    "邮箱地址格式不正确",
		"code.required":  "请输入6位验证码",
		"code.len":       "请输入6位验证码",
	}
}

// ChangePasswordRequest validates password change data.
type ChangePasswordRequest struct {
	CurrentPassword string `form:"current_password" json:"current_password"`
	NewPassword     string `form:"new_password" json:"new_password"`
}

func (r *ChangePasswordRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangePasswordRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"current_password": "required",
		"new_password":     "required|min_len:8|strong_password",
	}
}
func (r *ChangePasswordRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"current_password.required":      "请输入当前密码",
		"new_password.required":          "请输入新密码",
		"new_password.min_len":           "新密码至少需要8个字符",
		"new_password.strong_password":   "新密码必须包含大写字母、小写字母、数字和特殊字符",
	}
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/http/requests/api/user_request.go
git commit -m "feat: strengthen user request validation rules"
```

---

### Task 5: Game requests + controller conversion

**Files:**
- Modify: `dx-api/app/http/requests/api/game_request.go`
- Modify: `dx-api/app/http/controllers/api/game_controller.go`
- Modify: `dx-api/app/http/controllers/api/content_controller.go`

- [ ] **Step 1: Rewrite game_request.go**

Replace the entire file content:

```go
package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

// ListGamesRequest holds query parameters for listing published games.
type ListGamesRequest struct {
	Cursor      string   `form:"cursor" json:"cursor"`
	Limit       int      `form:"limit" json:"limit"`
	CategoryIDs []string `form:"categoryIds" json:"categoryIds"`
	PressID     string   `form:"pressId" json:"pressId"`
	Mode        string   `form:"mode" json:"mode"`
}

func (r *ListGamesRequest) Authorize(ctx http.Context) error { return nil }
func (r *ListGamesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"limit":   "min:1|max:50",
		"pressId": "uuid",
		"mode":    helpers.InEnum("mode"),
	}
}
func (r *ListGamesRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"cursor":  "trim",
		"pressId": "trim",
	}
}
func (r *ListGamesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"limit.min":   "每页数量不能小于1",
		"limit.max":   "每页数量不能超过50",
		"pressId.uuid": "无效的出版社ID",
		"mode.in":      "无效的游戏模式",
	}
}

// SearchGamesRequest holds query parameters for searching games.
type SearchGamesRequest struct {
	Query string `form:"q" json:"q"`
	Limit int    `form:"limit" json:"limit"`
}

func (r *SearchGamesRequest) Authorize(ctx http.Context) error { return nil }
func (r *SearchGamesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"q":     "required|min_len:1|max_len:50",
		"limit": "min:1|max:50",
	}
}
func (r *SearchGamesRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"q": "trim",
	}
}
func (r *SearchGamesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"q.required":  "请输入搜索关键词",
		"q.max_len":   "搜索关键词不能超过50个字符",
		"limit.min":   "每页数量不能小于1",
		"limit.max":   "每页数量不能超过50",
	}
}

// LevelContentRequest holds query parameters for fetching level content.
type LevelContentRequest struct {
	Degree string `form:"degree" json:"degree"`
}

func (r *LevelContentRequest) Authorize(ctx http.Context) error { return nil }
func (r *LevelContentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"degree": helpers.InEnum("degree"),
	}
}
func (r *LevelContentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"degree.in": "无效的难度级别",
	}
}
func (r *LevelContentRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreePractice)
	}
	return nil
}
```

Note: `LevelContentRequest` imports `validation` and `consts` — update game_request.go imports:

```go
import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)
```

- [ ] **Step 2: Update game_controller.go — List method**

Replace the `List` method body to use `helpers.Validate()`:

```go
func (c *GameController) List(ctx contractshttp.Context) contractshttp.Response {
	var req requests.ListGamesRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	limit := req.Limit
	if limit == 0 {
		limit = helpers.DefaultCursorLimit
	}

	games, nextCursor, hasMore, err := services.ListPublishedGames(req.Cursor, limit, req.CategoryIDs, req.PressID, req.Mode)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list games")
	}

	return helpers.Paginated(ctx, games, nextCursor, hasMore)
}
```

Add the requests import if not already present:

```go
requests "dx-api/app/http/requests/api"
```

Remove the now-unused `"strings"` import.

- [ ] **Step 3: Update game_controller.go — Search method**

Replace the `Search` method body:

```go
func (c *GameController) Search(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SearchGamesRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	games, err := services.SearchGames(req.Query, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to search games")
	}

	return helpers.Success(ctx, games)
}
```

Remove the now-unused `"strconv"` import.

- [ ] **Step 4: Update content_controller.go — LevelContent method**

The `LevelContent` method lives in `content_controller.go` (not `game_controller.go`). Replace the manual parsing:

```go
func (c *ContentController) LevelContent(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	levelID := ctx.Request().Route("levelId")
	if levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "level id is required")
	}

	var req requests.LevelContentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	items, err := services.GetLevelContent(levelID, req.Degree)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get level content")
	}

	return helpers.Success(ctx, items)
}
```

Add the requests import to content_controller.go:

```go
requests "dx-api/app/http/requests/api"
```

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 6: Commit**

```bash
cd dx-api && git add app/http/requests/api/game_request.go app/http/controllers/api/game_controller.go app/http/controllers/api/content_controller.go
git commit -m "feat: add validation to game requests, convert to FormRequest"
```

---

### Task 6: Session requests — existing FormRequests

**Files:**
- Modify: `dx-api/app/http/requests/api/session_request.go`

This task updates the 6 structs that already have `Authorize()` + `Rules()`: StartSessionRequest, StartLevelRequest, RecordAnswerRequest, RecordSkipRequest, SyncPlayTimeRequest, EndSessionRequest.

- [ ] **Step 1: Rewrite session_request.go**

Replace the entire file content:

```go
package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ---------- StartSessionRequest ----------

type StartSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
	LevelID *string `form:"level_id" json:"level_id"`
}

func (r *StartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":  "required|uuid",
		"degree":   helpers.InEnum("degree"),
		"pattern":  helpers.InEnum("pattern"),
		"level_id": "uuid",
	}
}
func (r *StartSessionRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":  "trim",
		"pattern": "trim",
	}
}
func (r *StartSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
		"degree.in":        "无效的难度级别",
		"pattern.in":       "无效的练习模式",
		"level_id.uuid":    "无效的关卡ID",
	}
}
func (r *StartSessionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// ---------- CheckActiveSessionRequest ----------

type CheckActiveSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
}

func (r *CheckActiveSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *CheckActiveSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "required|uuid",
		"degree":  helpers.InEnum("degree"),
		"pattern": helpers.InEnum("pattern"),
	}
}
func (r *CheckActiveSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
		"degree.in":        "无效的难度级别",
		"pattern.in":       "无效的练习模式",
	}
}

// ---------- CheckActiveLevelSessionRequest ----------

type CheckActiveLevelSessionRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
}

func (r *CheckActiveLevelSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *CheckActiveLevelSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
		"game_level_id": "required|uuid",
	}
}
func (r *CheckActiveLevelSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
	}
}

// ---------- StartLevelRequest ----------

type StartLevelRequest struct {
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
}

func (r *StartLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
	}
}
func (r *StartLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
	}
}
func (r *StartLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// ---------- CompleteLevelRequest ----------
// Controller reads game_level_id from route param — only body fields validated here.

type CompleteLevelRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	Score       int    `form:"score" json:"score"`
	MaxCombo    int    `form:"max_combo" json:"max_combo"`
	TotalItems  int    `form:"total_items" json:"total_items"`
}

func (r *CompleteLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CompleteLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"score":       "min:0",
		"max_combo":   "min:0",
		"total_items": "min:0",
	}
}
func (r *CompleteLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"score.min":       "分数不能为负数",
		"max_combo.min":   "连击数不能为负数",
		"total_items.min": "总数不能为负数",
	}
}

// ---------- AdvanceLevelRequest ----------
// Controller has fallback: if next_level_id is empty, uses route param levelId.

type AdvanceLevelRequest struct {
	NextLevelID string `form:"next_level_id" json:"next_level_id"`
}

func (r *AdvanceLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *AdvanceLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"next_level_id": "uuid",
	}
}
func (r *AdvanceLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"next_level_id.uuid": "无效的关卡ID",
	}
}

// ---------- RecordAnswerRequest ----------

type RecordAnswerRequest struct {
	GameSessionLevelID string  `form:"game_session_level_id" json:"game_session_level_id"`
	GameLevelID        string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID      string  `form:"content_item_id" json:"content_item_id"`
	IsCorrect          bool    `form:"is_correct" json:"is_correct"`
	UserAnswer         string  `form:"user_answer" json:"user_answer"`
	SourceAnswer       string  `form:"source_answer" json:"source_answer"`
	BaseScore          int     `form:"base_score" json:"base_score"`
	ComboScore         int     `form:"combo_score" json:"combo_score"`
	Score              int     `form:"score" json:"score"`
	MaxCombo           int     `form:"max_combo" json:"max_combo"`
	PlayTime           int     `form:"play_time" json:"play_time"`
	NextContentItemID  *string `form:"next_content_item_id" json:"next_content_item_id"`
	Duration           int     `form:"duration" json:"duration"`
}

func (r *RecordAnswerRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordAnswerRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_level_id": "required|uuid",
		"game_level_id":         "required|uuid",
		"content_item_id":       "required|uuid",
		"base_score":            "min:0",
		"combo_score":           "min:0",
		"score":                 "min:0",
		"max_combo":             "min:0",
		"play_time":             "min:0",
		"duration":              "min:0",
		"next_content_item_id":  "uuid",
	}
}
func (r *RecordAnswerRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_level_id.required": "请指定关卡会话",
		"game_session_level_id.uuid":     "无效的关卡会话ID",
		"game_level_id.required":         "请指定关卡",
		"game_level_id.uuid":             "无效的关卡ID",
		"content_item_id.required":       "请指定内容项",
		"content_item_id.uuid":           "无效的内容项ID",
		"base_score.min":                 "基础分数不能为负数",
		"combo_score.min":                "连击分数不能为负数",
		"score.min":                      "分数不能为负数",
		"max_combo.min":                  "最大连击不能为负数",
		"play_time.min":                  "游玩时长不能为负数",
		"duration.min":                   "持续时间不能为负数",
		"next_content_item_id.uuid":      "无效的内容项ID",
	}
}

// ---------- RecordSkipRequest ----------

type RecordSkipRequest struct {
	GameLevelID       string  `form:"game_level_id" json:"game_level_id"`
	PlayTime          int     `form:"play_time" json:"play_time"`
	NextContentItemID *string `form:"next_content_item_id" json:"next_content_item_id"`
}

func (r *RecordSkipRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordSkipRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id":        "required|uuid",
		"play_time":            "min:0",
		"next_content_item_id": "uuid",
	}
}
func (r *RecordSkipRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required":    "请指定关卡",
		"game_level_id.uuid":        "无效的关卡ID",
		"play_time.min":             "游玩时长不能为负数",
		"next_content_item_id.uuid": "无效的内容项ID",
	}
}

// ---------- SyncPlayTimeRequest ----------

type SyncPlayTimeRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	PlayTime    int    `form:"play_time" json:"play_time"`
}

func (r *SyncPlayTimeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SyncPlayTimeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
		"play_time":     "required|min:0",
	}
}
func (r *SyncPlayTimeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"play_time.required":     "请提供游玩时长",
		"play_time.min":          "游玩时长不能为负数",
	}
}

// ---------- UpdateContentItemRequest ----------

type UpdateContentItemRequest struct {
	ContentItemID *string `form:"content_item_id" json:"content_item_id"`
}

func (r *UpdateContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "uuid",
	}
}
func (r *UpdateContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.uuid": "无效的内容ID",
	}
}

// ---------- EndSessionRequest ----------

type EndSessionRequest struct {
	GameID             string `form:"game_id" json:"game_id"`
	Score              int    `form:"score" json:"score"`
	Exp                int    `form:"exp" json:"exp"`
	MaxCombo           int    `form:"max_combo" json:"max_combo"`
	CorrectCount       int    `form:"correct_count" json:"correct_count"`
	WrongCount         int    `form:"wrong_count" json:"wrong_count"`
	SkipCount          int    `form:"skip_count" json:"skip_count"`
	AllLevelsCompleted bool   `form:"all_levels_completed" json:"all_levels_completed"`
}

func (r *EndSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *EndSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"score":         "min:0",
		"exp":           "min:0",
		"max_combo":     "min:0",
		"correct_count": "min:0",
		"wrong_count":   "min:0",
		"skip_count":    "min:0",
	}
}
func (r *EndSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":  "请选择游戏",
		"game_id.uuid":      "无效的游戏ID",
		"score.min":         "分数不能为负数",
		"exp.min":           "经验值不能为负数",
		"max_combo.min":     "最大连击不能为负数",
		"correct_count.min": "正确数不能为负数",
		"wrong_count.min":   "错误数不能为负数",
		"skip_count.min":    "跳过数不能为负数",
	}
}

// ---------- RestoreSessionRequest ----------
// Controller currently reads from query params — converting to FormRequest.

type RestoreSessionRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
}

func (r *RestoreSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *RestoreSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
	}
}
func (r *RestoreSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
	}
}
```

Note: `RestartLevelRequest` is removed — it is dead code (controller reads from route param).

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success (if RestartLevelRequest is referenced elsewhere, check and remove those references)

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/http/requests/api/session_request.go
git commit -m "feat: strengthen session request validation, remove dead RestartLevelRequest"
```

---

### Task 7: Session controller updates

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_session_controller.go`

- [ ] **Step 1: Update CompleteLevel to use helpers.Validate**

Replace `ctx.Request().Bind(&req)` with `helpers.Validate(ctx, &req)`:

```go
func (c *GameSessionController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	gameLevelID := ctx.Request().Route("levelId")

	var req requests.CompleteLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CompleteLevel(userID, sessionID, gameLevelID, req.Score, req.MaxCombo, req.TotalItems)
	if err != nil {
		if errors.Is(err, services.ErrSessionLevelNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡会话不存在")
		}
		if errors.Is(err, services.ErrSessionNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, "会话不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to complete level")
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 2: Update AdvanceLevel to use helpers.Validate**

Keep the fallback logic but use `helpers.Validate` for format checking:

```go
func (c *GameSessionController) AdvanceLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	gameLevelID := ctx.Request().Route("levelId")

	var req requests.AdvanceLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	// Fallback to route param if body field is empty
	nextLevelID := req.NextLevelID
	if nextLevelID == "" {
		nextLevelID = gameLevelID
	}
	if nextLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "next_level_id is required")
	}

	if err := services.AdvanceLevel(userID, sessionID, nextLevelID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 3: Update CheckActive to use helpers.Validate**

```go
func (c *GameSessionController) CheckActive(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CheckActiveSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CheckActiveSession(userID, req.GameID, req.Degree, req.Pattern)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active session")
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 4: Update CheckActiveLevel to use helpers.Validate**

```go
func (c *GameSessionController) CheckActiveLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CheckActiveLevelSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CheckActiveLevelSession(userID, req.GameID, req.Degree, req.Pattern, req.GameLevelID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active level session")
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 5: Update Restore to use helpers.Validate**

```go
func (c *GameSessionController) Restore(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.RestoreSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.RestoreSessionData(userID, sessionID, req.GameLevelID)
	if err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 6: Update UpdateContentItem to use helpers.Validate**

```go
func (c *GameSessionController) UpdateContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.UpdateContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.UpdateCurrentContentItem(userID, sessionID, req.ContentItemID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 7: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 8: Commit**

```bash
cd dx-api && git add app/http/controllers/api/game_session_controller.go
git commit -m "feat: convert session controllers to use FormRequest validation"
```

---

### Task 8: Remaining API requests (tracking, favorites, feedback, report, upload, redeem, content_seek)

**Files:**
- Modify: `dx-api/app/http/requests/api/user_master_request.go`
- Modify: `dx-api/app/http/requests/api/user_review_request.go`
- Modify: `dx-api/app/http/requests/api/user_unknown_request.go`
- Modify: `dx-api/app/http/requests/api/user_favorite_request.go`
- Modify: `dx-api/app/http/requests/api/feedback_request.go`
- Modify: `dx-api/app/http/requests/api/game_report_request.go`
- Modify: `dx-api/app/http/requests/api/upload_request.go`
- Modify: `dx-api/app/http/requests/api/user_redeem_request.go`
- Modify: `dx-api/app/http/requests/api/content_seek_request.go`

- [ ] **Step 1: Update user_master_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type MarkMasteredRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkMasteredRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkMasteredRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required|uuid",
		"game_id":         "required|uuid",
		"game_level_id":   "required|uuid",
	}
}
func (r *MarkMasteredRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.required": "请指定内容",
		"content_item_id.uuid":     "无效的内容ID",
		"game_id.required":         "请选择游戏",
		"game_id.uuid":             "无效的游戏ID",
		"game_level_id.required":   "请指定关卡",
		"game_level_id.uuid":       "无效的关卡ID",
	}
}

type BulkDeleteRequest struct {
	IDs []string `form:"ids" json:"ids"`
}

func (r *BulkDeleteRequest) Authorize(ctx http.Context) error { return nil }
func (r *BulkDeleteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"ids":   "required|min_len:1|max_len:100",
		"ids.*": "uuid",
	}
}
func (r *BulkDeleteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"ids.required": "请选择要删除的项目",
		"ids.min_len":  "请至少选择一项",
		"ids.max_len":  "单次最多删除100条",
		"ids.*.uuid":   "包含无效的ID",
	}
}
```

- [ ] **Step 2: Update user_review_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type MarkReviewRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkReviewRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkReviewRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required|uuid",
		"game_id":         "required|uuid",
		"game_level_id":   "required|uuid",
	}
}
func (r *MarkReviewRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.required": "请指定内容",
		"content_item_id.uuid":     "无效的内容ID",
		"game_id.required":         "请选择游戏",
		"game_id.uuid":             "无效的游戏ID",
		"game_level_id.required":   "请指定关卡",
		"game_level_id.uuid":       "无效的关卡ID",
	}
}
```

- [ ] **Step 3: Update user_unknown_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type MarkUnknownRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkUnknownRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkUnknownRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required|uuid",
		"game_id":         "required|uuid",
		"game_level_id":   "required|uuid",
	}
}
func (r *MarkUnknownRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.required": "请指定内容",
		"content_item_id.uuid":     "无效的内容ID",
		"game_id.required":         "请选择游戏",
		"game_id.uuid":             "无效的游戏ID",
		"game_level_id.required":   "请指定关卡",
		"game_level_id.uuid":       "无效的关卡ID",
	}
}
```

- [ ] **Step 4: Update user_favorite_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type ToggleFavoriteRequest struct {
	GameID string `form:"game_id" json:"game_id"`
}

func (r *ToggleFavoriteRequest) Authorize(ctx http.Context) error { return nil }
func (r *ToggleFavoriteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "required|uuid",
	}
}
func (r *ToggleFavoriteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
	}
}
```

- [ ] **Step 5: Update feedback_request.go**

```go
package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

type SubmitFeedbackRequest struct {
	Type        string `form:"type" json:"type"`
	Description string `form:"description" json:"description"`
}

func (r *SubmitFeedbackRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitFeedbackRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        "required|" + helpers.InEnum("feedback_type"),
		"description": "required|min_len:2|max_len:200",
	}
}
func (r *SubmitFeedbackRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"description": "trim",
	}
}
func (r *SubmitFeedbackRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required":        "请选择反馈类型",
		"type.in":              "无效的反馈类型",
		"description.required": "请输入反馈内容",
		"description.min_len":  "反馈内容不能少于2个字符",
		"description.max_len":  "描述不能超过200个字符",
	}
}
```

- [ ] **Step 6: Update game_report_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type SubmitReportRequest struct {
	GameID        string  `form:"game_id" json:"game_id"`
	GameLevelID   string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID string  `form:"content_item_id" json:"content_item_id"`
	Reason        string  `form:"reason" json:"reason"`
	Note          *string `form:"note" json:"note"`
}

func (r *SubmitReportRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitReportRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":         "required|uuid",
		"game_level_id":   "required|uuid",
		"content_item_id": "required|uuid",
		"reason":          "required|max_len:200",
		"note":            "max_len:500",
	}
}
func (r *SubmitReportRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"reason": "trim",
		"note":   "trim",
	}
}
func (r *SubmitReportRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":         "请指定游戏",
		"game_id.uuid":             "无效的游戏ID",
		"game_level_id.required":   "请指定关卡",
		"game_level_id.uuid":       "无效的关卡ID",
		"content_item_id.required": "请指定内容项",
		"content_item_id.uuid":     "无效的内容项ID",
		"reason.required":          "请选择举报原因",
		"reason.max_len":           "举报原因不能超过200个字符",
		"note.max_len":             "备注不能超过500个字符",
	}
}
```

- [ ] **Step 7: Update upload_request.go**

```go
package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

type UploadImageRequest struct {
	Role string `form:"role" json:"role"`
}

func (r *UploadImageRequest) Authorize(ctx http.Context) error { return nil }
func (r *UploadImageRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"role": "required|" + helpers.InEnum("image_role"),
	}
}
func (r *UploadImageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"role.required": "请指定图片用途",
		"role.in":       "无效的图片用途",
	}
}
```

- [ ] **Step 8: Update user_redeem_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type RedeemCodeRequest struct {
	Code string `form:"code" json:"code"`
}

func (r *RedeemCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *RedeemCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "required|len:19",
	}
}
func (r *RedeemCodeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "trim|upper",
	}
}
func (r *RedeemCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "请输入兑换码",
		"code.len":      "兑换码格式不正确",
	}
}
```

- [ ] **Step 9: Update content_seek_request.go**

```go
package api

import "github.com/goravel/framework/contracts/http"

type SubmitContentSeekRequest struct {
	CourseName  string `form:"course_name" json:"course_name"`
	Description string `form:"description" json:"description"`
	DiskUrl     string `form:"disk_url" json:"disk_url"`
}

func (r *SubmitContentSeekRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitContentSeekRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name": "required|min_len:2|max_len:30",
		"description": "required|min_len:2|max_len:200",
		"disk_url":    "required|full_url|max_len:500",
	}
}
func (r *SubmitContentSeekRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name": "trim",
		"description": "trim",
		"disk_url":    "trim",
	}
}
func (r *SubmitContentSeekRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name.required": "请输入课程名称",
		"course_name.min_len":  "课程名称至少需要2个字符",
		"course_name.max_len":  "课程名称不能超过30个字符",
		"description.required": "请输入描述",
		"description.min_len":  "描述至少需要2个字符",
		"description.max_len":  "描述不能超过200个字符",
		"disk_url.required":    "请输入网盘链接",
		"disk_url.full_url":    "请输入有效的网盘链接",
		"disk_url.max_len":     "网盘链接不能超过500个字符",
	}
}
```

- [ ] **Step 10: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 11: Commit**

```bash
cd dx-api && git add app/http/requests/api/user_master_request.go app/http/requests/api/user_review_request.go app/http/requests/api/user_unknown_request.go app/http/requests/api/user_favorite_request.go app/http/requests/api/feedback_request.go app/http/requests/api/game_report_request.go app/http/requests/api/upload_request.go app/http/requests/api/user_redeem_request.go app/http/requests/api/content_seek_request.go
git commit -m "feat: strengthen remaining API request validation rules"
```

---

### Task 9: Course game requests

**Files:**
- Modify: `dx-api/app/http/requests/api/course_game_request.go`

- [ ] **Step 1: Rewrite course_game_request.go**

```go
package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

// ---------- CreateGameRequest ----------

type CreateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *CreateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":           "required|min_len:2|max_len:100",
		"description":    "max_len:500",
		"gameMode":       "required|" + helpers.InEnum("mode"),
		"gameCategoryId": "required|uuid",
		"gamePressId":    "required|uuid",
		"coverId":        "uuid",
	}
}
func (r *CreateGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *CreateGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":          "请输入游戏名称",
		"name.min_len":           "游戏名称至少需要2个字符",
		"name.max_len":           "游戏名称不能超过100个字符",
		"description.max_len":    "游戏描述不能超过500个字符",
		"gameMode.required":      "请选择游戏模式",
		"gameMode.in":            "无效的游戏模式",
		"gameCategoryId.required":"请选择游戏分类",
		"gameCategoryId.uuid":    "无效的游戏分类",
		"gamePressId.required":   "请选择出版社",
		"gamePressId.uuid":       "无效的出版社",
		"coverId.uuid":           "无效的封面图片",
	}
}

// ---------- UpdateGameRequest ----------

type UpdateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *UpdateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":           "required|min_len:2|max_len:100",
		"description":    "max_len:500",
		"gameMode":       "required|" + helpers.InEnum("mode"),
		"gameCategoryId": "required|uuid",
		"gamePressId":    "required|uuid",
		"coverId":        "uuid",
	}
}
func (r *UpdateGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *UpdateGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":          "请输入游戏名称",
		"name.min_len":           "游戏名称至少需要2个字符",
		"name.max_len":           "游戏名称不能超过100个字符",
		"description.max_len":    "游戏描述不能超过500个字符",
		"gameMode.required":      "请选择游戏模式",
		"gameMode.in":            "无效的游戏模式",
		"gameCategoryId.required":"请选择游戏分类",
		"gameCategoryId.uuid":    "无效的游戏分类",
		"gamePressId.required":   "请选择出版社",
		"gamePressId.uuid":       "无效的出版社",
		"coverId.uuid":           "无效的封面图片",
	}
}

// ---------- CreateLevelRequest ----------

type CreateLevelRequest struct {
	Name        string  `form:"name" json:"name"`
	Description *string `form:"description" json:"description"`
}

func (r *CreateLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "required|min_len:1|max_len:100",
		"description": "max_len:500",
	}
}
func (r *CreateLevelRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *CreateLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":       "请输入关卡名称",
		"name.max_len":        "关卡名称不能超过100个字符",
		"description.max_len": "关卡描述不能超过500个字符",
	}
}

// ---------- SaveMetadataBatchRequest ----------

type SaveMetadataBatchRequest struct {
	GameLevelID string              `json:"gameLevelId"`
	SourceFrom  string              `json:"sourceFrom"`
	Entries     []MetadataEntryJSON `json:"entries"`
}

func (r *SaveMetadataBatchRequest) Authorize(ctx http.Context) error { return nil }
func (r *SaveMetadataBatchRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entries":              "required|min_len:1|max_len:200",
		"gameLevelId":          "required|uuid",
		"sourceFrom":           "required|" + helpers.InEnum("source_from"),
		"entries.*.sourceData": "required",
		"entries.*.sourceType": "required|" + helpers.InEnum("source_type"),
	}
}
func (r *SaveMetadataBatchRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"entries.*.sourceData":  "trim",
		"entries.*.translation": "trim",
	}
}
func (r *SaveMetadataBatchRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entries.required":              "请提供内容数据",
		"entries.min_len":               "请至少提供一条数据",
		"entries.max_len":               "单次最多提交200条",
		"gameLevelId.required":          "请指定关卡",
		"gameLevelId.uuid":              "无效的关卡ID",
		"sourceFrom.required":           "请指定来源",
		"sourceFrom.in":                 "无效的来源类型",
		"entries.*.sourceData.required": "每条数据的内容不能为空",
		"entries.*.sourceType.required": "每条数据的类型不能为空",
		"entries.*.sourceType.in":       "无效的内容类型",
	}
}

type MetadataEntryJSON struct {
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	SourceType  string  `json:"sourceType"`
}

// ---------- ReorderMetadataRequest ----------

type ReorderMetadataRequest struct {
	GameLevelID string  `json:"gameLevelId"`
	MetaID      string  `json:"metaId"`
	NewOrder    float64 `json:"newOrder"`
}

func (r *ReorderMetadataRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderMetadataRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"metaId":      "required|uuid",
		"gameLevelId": "required|uuid",
		"newOrder":    "required|min:0",
	}
}
func (r *ReorderMetadataRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"metaId.required":      "请指定元数据",
		"metaId.uuid":          "无效的元数据ID",
		"gameLevelId.required": "请指定关卡",
		"gameLevelId.uuid":     "无效的关卡ID",
		"newOrder.required":    "请指定排序位置",
		"newOrder.min":         "排序位置不能为负数",
	}
}

// ---------- InsertContentItemRequest ----------

type InsertContentItemRequest struct {
	GameLevelID     string  `json:"gameLevelId"`
	ContentMetaID   string  `json:"contentMetaId"`
	Content         string  `json:"content"`
	ContentType     string  `json:"contentType"`
	Translation     *string `json:"translation"`
	ReferenceItemID string  `json:"referenceItemId"`
	Direction       string  `json:"direction"`
}

func (r *InsertContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *InsertContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"contentMetaId":   "required|uuid",
		"gameLevelId":     "required|uuid",
		"contentType":     helpers.InEnum("content_type"),
		"direction":       "in:before,after",
		"referenceItemId": "uuid",
	}
}
func (r *InsertContentItemRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content":     "trim",
		"translation": "trim",
	}
}
func (r *InsertContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"contentMetaId.required": "请指定元数据",
		"contentMetaId.uuid":     "无效的元数据ID",
		"gameLevelId.required":   "请指定关卡",
		"gameLevelId.uuid":       "无效的关卡ID",
		"contentType.in":         "无效的内容类型",
		"direction.in":           "插入方向只能为前或后",
		"referenceItemId.uuid":   "无效的参考项ID",
	}
}

// ---------- UpdateContentItemTextRequest ----------

type UpdateContentItemTextRequest struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
}

func (r *UpdateContentItemTextRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateContentItemTextRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content":     "max_len:2000",
		"translation": "max_len:2000",
	}
}
func (r *UpdateContentItemTextRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content":     "trim",
		"translation": "trim",
	}
}
func (r *UpdateContentItemTextRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.max_len":     "内容不能超过2000个字符",
		"translation.max_len": "翻译不能超过2000个字符",
	}
}

// ---------- ReorderContentItemRequest ----------

type ReorderContentItemRequest struct {
	ItemID   string  `json:"itemId"`
	NewOrder float64 `json:"newOrder"`
}

func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"itemId":   "required|uuid",
		"newOrder": "required|min:0",
	}
}
func (r *ReorderContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"itemId.required":   "请指定内容项",
		"itemId.uuid":       "无效的内容项ID",
		"newOrder.required": "请指定排序位置",
		"newOrder.min":      "排序位置不能为负数",
	}
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/http/requests/api/course_game_request.go
git commit -m "feat: strengthen course game request validation rules"
```

---

### Task 10: Admin requests

**Files:**
- Modify: `dx-api/app/http/requests/adm/notice_request.go`
- Modify: `dx-api/app/http/requests/adm/redeem_request.go`

- [ ] **Step 1: Update notice_request.go**

```go
package adm

import "github.com/goravel/framework/contracts/http"

type CreateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *CreateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "required|min_len:2|max_len:200",
		"content": "required|max_len:5000",
		"icon":    "max_len:50",
	}
}
func (r *CreateNoticeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "trim",
		"content": "trim",
		"icon":    "trim",
	}
}
func (r *CreateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.required":   "Title is required",
		"title.min_len":    "Title must be at least 2 characters",
		"title.max_len":    "Title must not exceed 200 characters",
		"content.required": "Content is required",
		"content.max_len":  "Content must not exceed 5000 characters",
		"icon.max_len":     "Icon must not exceed 50 characters",
	}
}

type UpdateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *UpdateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "required|min_len:2|max_len:200",
		"content": "required|max_len:5000",
		"icon":    "max_len:50",
	}
}
func (r *UpdateNoticeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "trim",
		"content": "trim",
		"icon":    "trim",
	}
}
func (r *UpdateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.required":   "Title is required",
		"title.min_len":    "Title must be at least 2 characters",
		"title.max_len":    "Title must not exceed 200 characters",
		"content.required": "Content is required",
		"content.max_len":  "Content must not exceed 5000 characters",
		"icon.max_len":     "Icon must not exceed 50 characters",
	}
}
```

- [ ] **Step 2: Update redeem_request.go**

```go
package adm

import "github.com/goravel/framework/contracts/http"

// Count is string because Goravel validation operates on raw parsed data
// where JSON numbers become float64 — the in rule cannot compare float64.
type GenerateCodesRequest struct {
	Grade string `form:"grade" json:"grade"`
	Count string `form:"count" json:"count"`
}

func (r *GenerateCodesRequest) Authorize(ctx http.Context) error { return nil }
func (r *GenerateCodesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"grade": "required|in:month,season,year,lifetime",
		"count": "required|in:10,50,100,500",
	}
}
func (r *GenerateCodesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"grade.required": "Grade is required",
		"grade.in":       "Grade must be one of: month, season, year, lifetime",
		"count.required": "Count is required",
		"count.in":       "Count must be one of: 10, 50, 100, 500",
	}
}
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add app/http/requests/adm/notice_request.go app/http/requests/adm/redeem_request.go
git commit -m "feat: strengthen admin request validation, consistent English messages"
```

---

### Task 11: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Full build**

Run: `cd dx-api && go build ./...`
Expected: success

- [ ] **Step 2: Run all tests**

Run: `cd dx-api && go test -race ./...`
Expected: all pass

- [ ] **Step 3: Verify no unused imports**

Run: `cd dx-api && go vet ./...`
Expected: no errors
