# WeChat Mini Auth — dx-api Backend Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend dx-api with WeChat mini program authentication and platform tracking.

**Architecture:** Add `openid`/`unionid` to the users table and `platform` to user_logins. Add a new public endpoint `POST /api/auth/wechat-mini` that exchanges a wx.login code with WeChat's server, finds or auto-registers the user by openid, and returns a JWT. Existing auth flows are untouched except `RecordLogin` gains a `platform` parameter.

**Tech Stack:** Go, Goravel framework, PostgreSQL, existing `uuid`/`helpers`/`models`/`services` patterns.

**Spec:** `docs/superpowers/specs/2026-04-18-wechat-mini-design.md`

**Prerequisite:** None — complete this plan before starting the frontend plan.

---

## File Map

| Action | Path |
|--------|------|
| Modify | `database/migrations/20260322000001_create_users_table.go` |
| Modify | `database/migrations/20260322000009_create_user_logins_table.go` |
| Modify | `app/models/user.go` |
| Modify | `app/models/user_login.go` |
| Create | `app/consts/platform.go` |
| Modify | `app/services/api/auth_service.go` — `RecordLogin` signature |
| Modify | `app/http/controllers/api/auth_controller.go` — update `RecordLogin` callers |
| Create | `config/wechat.go` |
| Create | `app/http/requests/api/wechat_auth_request.go` |
| Create | `app/services/api/wechat_auth_service.go` |
| Create | `app/services/api/wechat_auth_service_test.go` |
| Create | `app/http/controllers/api/wechat_auth_controller.go` |
| Modify | `routes/api.go` — register new route |
| Modify | `.env.example` — add wechat vars |

---

### Task 1: Update migrations and models

**Files:**
- Modify: `database/migrations/20260322000001_create_users_table.go`
- Modify: `database/migrations/20260322000009_create_user_logins_table.go`
- Modify: `app/models/user.go`
- Modify: `app/models/user_login.go`

- [ ] **Step 1: Add openid + unionid to users migration**

In `database/migrations/20260322000001_create_users_table.go`, add after `table.TimestampsTz()`:

```go
// existing lines above ...
table.TimestampsTz()
table.Text("openid").Nullable()
table.Text("unionid").Nullable()
table.Unique("username")
table.Unique("email")
table.Unique("phone")
table.Unique("invite_code")
table.Unique("openid")
table.Index("avatar_id")
table.Index("nickname")
table.Index("created_at")
table.Index("last_played_at")
```

The full `Up()` body becomes:

```go
func (r *M20260322000001CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("grade").Default("")
			table.Text("username").Default("")
			table.Text("nickname").Nullable()
			table.Text("email").Nullable()
			table.Text("phone").Nullable()
			table.Text("password").Default("")
			table.Uuid("avatar_id").Nullable()
			table.Text("city").Nullable()
			table.Text("introduction").Nullable()
			table.Boolean("is_active").Default(true)
			table.Boolean("is_mock").Default(false)
			table.Integer("beans").Default(0)
			table.Integer("granted_beans").Default(0)
			table.Integer("exp").Default(0)
			table.Text("invite_code").Default("")
			table.Integer("current_play_streak").Default(0)
			table.Integer("max_play_streak").Default(0)
			table.TimestampTz("last_played_at").Nullable()
			table.TimestampTz("vip_due_at").Nullable()
			table.TimestampTz("last_read_notice_at").Nullable()
			table.Text("openid").Nullable()
			table.Text("unionid").Nullable()
			table.TimestampsTz()
			table.Unique("username")
			table.Unique("email")
			table.Unique("phone")
			table.Unique("invite_code")
			table.Unique("openid")
			table.Index("avatar_id")
			table.Index("nickname")
			table.Index("created_at")
			table.Index("last_played_at")
		})
	}
	return nil
}
```

- [ ] **Step 2: Add platform to user_logins migration**

In `database/migrations/20260322000009_create_user_logins_table.go`, the full `Up()` body becomes:

```go
func (r *M20260322000009CreateUserLoginsTable) Up() error {
	if !facades.Schema().HasTable("user_logins") {
		return facades.Schema().Create("user_logins", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Text("ip").Default("")
			table.Text("agent").Nullable()
			table.Text("platform").Nullable()
			table.Text("country").Nullable()
			table.Text("province").Nullable()
			table.Text("city").Nullable()
			table.Text("isp").Nullable()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("ip")
			table.Index("created_at")
		})
	}
	return nil
}
```

- [ ] **Step 3: Update User model**

Replace the `User` struct in `app/models/user.go` with:

```go
type User struct {
	orm.Timestamps
	ID                string           `gorm:"column:id;primaryKey" json:"id"`
	Grade             string           `gorm:"column:grade" json:"grade"`
	Username          string           `gorm:"column:username" json:"username"`
	Nickname          *string          `gorm:"column:nickname" json:"nickname"`
	Email             *string          `gorm:"column:email" json:"email"`
	Phone             *string          `gorm:"column:phone" json:"phone"`
	Password          string           `gorm:"column:password" json:"-"`
	AvatarID          *string          `gorm:"column:avatar_id" json:"avatar_id"`
	City              *string          `gorm:"column:city" json:"city"`
	Introduction      *string          `gorm:"column:introduction" json:"introduction"`
	IsActive          bool             `gorm:"column:is_active" json:"is_active"`
	IsMock            bool             `gorm:"column:is_mock" json:"is_mock"`
	Beans             int              `gorm:"column:beans" json:"beans"`
	GrantedBeans      int              `gorm:"column:granted_beans" json:"granted_beans"`
	Exp               int              `gorm:"column:exp" json:"exp"`
	InviteCode        string           `gorm:"column:invite_code" json:"invite_code"`
	CurrentPlayStreak int              `gorm:"column:current_play_streak" json:"current_play_streak"`
	MaxPlayStreak     int              `gorm:"column:max_play_streak" json:"max_play_streak"`
	LastPlayedAt      *carbon.DateTime `gorm:"column:last_played_at" json:"last_played_at"`
	VipDueAt          *carbon.DateTime `gorm:"column:vip_due_at" json:"vip_due_at"`
	LastReadNoticeAt  *carbon.DateTime `gorm:"column:last_read_notice_at" json:"last_read_notice_at"`
	OpenID            *string          `gorm:"column:openid"  json:"-"`
	UnionID           *string          `gorm:"column:unionid" json:"-"`
}
```

- [ ] **Step 4: Update UserLogin model**

Replace `app/models/user_login.go` with:

```go
package models

import "github.com/goravel/framework/database/orm"

type UserLogin struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	UserID   string  `gorm:"column:user_id" json:"user_id"`
	IP       string  `gorm:"column:ip" json:"ip"`
	Agent    *string `gorm:"column:agent" json:"agent"`
	Platform *string `gorm:"column:platform" json:"platform"`
	Country  *string `gorm:"column:country" json:"country"`
	Province *string `gorm:"column:province" json:"province"`
	City     *string `gorm:"column:city" json:"city"`
	ISP      *string `gorm:"column:isp" json:"isp"`
}

func (u *UserLogin) TableName() string {
	return "user_logins"
}
```

- [ ] **Step 5: Verify compilation**

```bash
cd dx-api && go build ./...
```

Expected: no output (clean build).

- [ ] **Step 6: Commit**

```bash
git add database/migrations/20260322000001_create_users_table.go \
        database/migrations/20260322000009_create_user_logins_table.go \
        app/models/user.go \
        app/models/user_login.go
git commit -m "feat: add openid/unionid to users and platform to user_logins"
```

---

### Task 2: Platform constants + updated RecordLogin

**Files:**
- Create: `app/consts/platform.go`
- Modify: `app/services/api/auth_service.go`
- Modify: `app/http/controllers/api/auth_controller.go`

- [ ] **Step 1: Create platform constants**

Create `app/consts/platform.go`:

```go
package consts

const (
	PlatformWebsite = "website"
	PlatformMini    = "mini"
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
)
```

- [ ] **Step 2: Update RecordLogin signature in auth_service.go**

Find `RecordLogin` in `app/services/api/auth_service.go` and replace it:

```go
// RecordLogin creates a UserLogin record for audit purposes.
func RecordLogin(userID, ip, userAgent, platform string) {
	agent := userAgent
	p := platform
	login := models.UserLogin{
		ID:       uuid.Must(uuid.NewV7()).String(),
		UserID:   userID,
		IP:       ip,
		Agent:    &agent,
		Platform: &p,
	}
	_ = facades.Orm().Query().Create(&login)
}
```

- [ ] **Step 3: Update the RecordLogin caller in auth_controller.go**

In `app/http/controllers/api/auth_controller.go`, find the line:

```go
go services.RecordLogin(user.ID, ip, userAgent)
```

Replace with:

```go
go services.RecordLogin(user.ID, ip, userAgent, consts.PlatformWebsite)
```

- [ ] **Step 4: Verify compilation**

```bash
cd dx-api && go build ./...
```

Expected: no output (clean build).

- [ ] **Step 5: Commit**

```bash
git add app/consts/platform.go \
        app/services/api/auth_service.go \
        app/http/controllers/api/auth_controller.go
git commit -m "feat: add platform constants and track login platform"
```

---

### Task 3: WeChat config

**Files:**
- Create: `config/wechat.go`
- Modify: `.env.example`

- [ ] **Step 1: Create config/wechat.go**

```go
package config

import "github.com/goravel/framework/facades"

func init() {
	config := facades.Config()
	config.Add("wechat", map[string]any{
		"mini_app_id":     config.Env("WECHAT_MINI_APP_ID", ""),
		"mini_app_secret": config.Env("WECHAT_MINI_APP_SECRET", ""),
	})
}
```

This `init()` runs automatically when the `config` package is imported in `bootstrap/app.go` — no other registration needed.

- [ ] **Step 2: Add env vars to .env.example**

Append to `.env.example`:

```
WECHAT_MINI_APP_ID=
WECHAT_MINI_APP_SECRET=
```

- [ ] **Step 3: Verify compilation**

```bash
cd dx-api && go build ./...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add config/wechat.go .env.example
git commit -m "feat: add wechat mini program config"
```

---

### Task 4: WeChat auth request + service

**Files:**
- Create: `app/http/requests/api/wechat_auth_request.go`
- Create: `app/services/api/wechat_auth_service.go`
- Create: `app/services/api/wechat_auth_service_test.go`

- [ ] **Step 1: Write the failing test**

Create `app/services/api/wechat_auth_service_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchWechatSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "test_openid_abc123",
			"session_key": "test_session_key",
			"unionid":     "test_unionid_xyz",
			"errcode":     0,
			"errmsg":      "ok",
		})
	}))
	defer srv.Close()

	resp, err := fetchWechatSession("appid", "secret", "code", srv.URL+"?%s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OpenID != "test_openid_abc123" {
		t.Errorf("OpenID = %q, want %q", resp.OpenID, "test_openid_abc123")
	}
	if resp.UnionID != "test_unionid_xyz" {
		t.Errorf("UnionID = %q, want %q", resp.UnionID, "test_unionid_xyz")
	}
	if resp.ErrCode != 0 {
		t.Errorf("ErrCode = %d, want 0", resp.ErrCode)
	}
}

func TestFetchWechatSession_WechatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 40029,
			"errmsg":  "invalid code",
		})
	}))
	defer srv.Close()

	resp, err := fetchWechatSession("appid", "secret", "badcode", srv.URL+"?%s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ErrCode != 40029 {
		t.Errorf("ErrCode = %d, want 40029", resp.ErrCode)
	}
}

func TestGenerateWxUsername(t *testing.T) {
	tests := []struct {
		openID string
		want   string
	}{
		{"abcdefghijklmno", "wx_abcdefgh"},
		{"12345678xyz", "wx_12345678"},
	}
	for _, tt := range tests {
		got := generateWxUsername(tt.openID)
		if got != tt.want {
			t.Errorf("generateWxUsername(%q) = %q, want %q", tt.openID, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd dx-api && go test -race ./app/services/api/... -run "TestFetchWechatSession|TestGenerateWx" -v
```

Expected: FAIL — `fetchWechatSession`, `generateWxUsername` undefined.

- [ ] **Step 3: Create the request file**

Create `app/http/requests/api/wechat_auth_request.go`:

```go
package api

import "github.com/goravel/framework/contracts/http"

type WechatMiniAuthRequest struct {
	Code string `form:"code" json:"code"`
}

func (r *WechatMiniAuthRequest) Authorize(ctx http.Context) error { return nil }
func (r *WechatMiniAuthRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "required",
	}
}
func (r *WechatMiniAuthRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{"code": "trim"}
}
func (r *WechatMiniAuthRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "缺少 wx.login code",
	}
}
```

- [ ] **Step 4: Create the service file**

Create `app/services/api/wechat_auth_service.go`:

```go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatMiniSignIn exchanges a wx.login code for openid and signs in or registers the user.
func WechatMiniSignIn(ctx contractshttp.Context, code string) (string, *models.User, error) {
	appID := facades.Config().GetString("wechat.mini_app_id")
	secret := facades.Config().GetString("wechat.mini_app_secret")

	const wxURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	wxResp, err := fetchWechatSession(appID, secret, code, wxURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch wechat session: %w", err)
	}
	if wxResp.ErrCode != 0 {
		return "", nil, fmt.Errorf("wechat error %d: %s", wxResp.ErrCode, wxResp.ErrMsg)
	}

	var user models.User
	err = facades.Orm().Query().Where("openid", wxResp.OpenID).First(&user)
	if err != nil || user.ID == "" {
		user, err = registerWxUser(ctx, wxResp.OpenID, wxResp.UnionID)
		if err != nil {
			return "", nil, err
		}
	}

	token, err := issueSession(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func registerWxUser(ctx contractshttp.Context, openID, unionID string) (models.User, error) {
	username := generateWxUsername(openID)

	var existing models.User
	if err := facades.Orm().Query().Where("username", username).First(&existing); err == nil && existing.ID != "" {
		username = fmt.Sprintf("%s_%s", username, helpers.GenerateCode(4))
	}

	hashedPw, err := helpers.HashPassword(helpers.GenerateInviteCode(16))
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	oid := openID
	user := models.User{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Grade:      consts.UserGradeFree,
		Username:   username,
		Password:   hashedPw,
		OpenID:     &oid,
		IsActive:   true,
		InviteCode: helpers.GenerateInviteCode(8),
	}
	if unionID != "" {
		uid := unionID
		user.UnionID = &uid
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	if refErr := RecordReferralIfPresent(ctx, user.ID); refErr != nil {
		facades.Log().Warningf("record referral failed: %v", refErr)
	}

	return user, nil
}

func generateWxUsername(openID string) string {
	if len(openID) >= 8 {
		return "wx_" + openID[:8]
	}
	return "wx_" + openID
}

// fetchWechatSession calls the WeChat code2session API.
// urlFmt is the URL format string with %s placeholders for appID, secret, code.
// Passing a custom urlFmt allows test servers to be injected.
func fetchWechatSession(appID, secret, code, urlFmt string) (*wechatSessionResponse, error) {
	url := fmt.Sprintf(urlFmt, appID, secret, code)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result wechatSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
```

- [ ] **Step 5: Run tests and confirm they pass**

```bash
cd dx-api && go test -race ./app/services/api/... -run "TestFetchWechatSession|TestGenerateWx" -v
```

Expected:
```
--- PASS: TestFetchWechatSession_Success
--- PASS: TestFetchWechatSession_WechatError
--- PASS: TestGenerateWxUsername
PASS
```

- [ ] **Step 6: Commit**

```bash
git add app/http/requests/api/wechat_auth_request.go \
        app/services/api/wechat_auth_service.go \
        app/services/api/wechat_auth_service_test.go
git commit -m "feat: add wechat mini sign-in service with tests"
```

---

### Task 5: WeChat auth controller + route

**Files:**
- Create: `app/http/controllers/api/wechat_auth_controller.go`
- Modify: `routes/api.go`

- [ ] **Step 1: Create the controller**

Create `app/http/controllers/api/wechat_auth_controller.go`:

```go
package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type WechatAuthController struct{}

func NewWechatAuthController() *WechatAuthController {
	return &WechatAuthController{}
}

// MiniSignIn exchanges a wx.login code for a JWT.
// Returns token in the response body (mini program stores it in wx.storage, not cookies).
func (c *WechatAuthController) MiniSignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.WechatMiniAuthRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	token, user, err := services.WechatMiniSignIn(ctx, req.Code)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "微信登录失败")
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go services.RecordLogin(user.ID, ip, userAgent, consts.PlatformMini)

	return helpers.Success(ctx, map[string]any{"token": token, "user": user})
}
```

- [ ] **Step 2: Register the route in routes/api.go**

In `routes/api.go`, add the controller instantiation near the top of `Api()`, alongside the other controller declarations:

```go
wechatAuthController := apicontrollers.NewWechatAuthController()
```

Then inside `r.Prefix("/api").Group(...)`, in the auth routes group, add the new route:

```go
router.Prefix("/auth").Group(func(auth route.Router) {
    auth.Post("/signup", authController.SignUp)
    auth.Post("/signin", authController.SignIn)
    auth.Post("/logout", authController.Logout)
    auth.Post("/wechat-mini", wechatAuthController.MiniSignIn)
})
```

- [ ] **Step 3: Verify compilation**

```bash
cd dx-api && go build ./...
```

Expected: no output.

- [ ] **Step 4: Run all tests**

```bash
cd dx-api && go test -race ./...
```

Expected: all existing tests pass, new tests pass.

- [ ] **Step 5: Manual smoke test (requires running server + valid WeChat code)**

Start the server: `cd dx-api && air`

Test with an invalid code (should return wechat error):
```bash
curl -X POST http://localhost:3001/api/auth/wechat-mini \
  -H "Content-Type: application/json" \
  -d '{"code":"invalid_code_for_test"}'
```

Expected response:
```json
{"code": 50000, "message": "微信登录失败", "data": null}
```

- [ ] **Step 6: Commit**

```bash
git add app/http/controllers/api/wechat_auth_controller.go routes/api.go
git commit -m "feat: add POST /api/auth/wechat-mini endpoint"
```
