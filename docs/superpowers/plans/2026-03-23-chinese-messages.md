# Chinese Client Messages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace all English user-facing messages with Chinese in the `/api/*` backend layer.

**Architecture:** Inline string replacement across 3 layers — error sentinels, validation messages, and controller messages. No new files, no i18n infrastructure. Only user-visible business messages are translated; system/auth messages stay English.

**Tech Stack:** Go, Goravel framework

**Spec:** `docs/superpowers/specs/2026-03-23-chinese-messages-design.md`

---

### Task 1: Translate error sentinels

**Files:**
- Modify: `dx-api/app/services/api/errors.go`

- [ ] **Step 1: Replace all 28 business error sentinel messages**

Replace each `errors.New("...")` string with its Chinese translation. Keep the 4 technical sentinels (`ErrInvalidRefreshToken`, `ErrSessionReplaced`, `ErrRateLimited`, `ErrForbidden`) in English.

```go
var (
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrSessionReplaced      = errors.New("session replaced by another device")
	ErrRateLimited          = errors.New("rate limited")
	ErrInvalidCode          = errors.New("验证码无效或已过期")
	ErrDuplicateEmail       = errors.New("该邮箱已注册")
	ErrDuplicateUsername    = errors.New("用户名已被使用")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrInvalidPassword      = errors.New("密码错误")
	ErrNicknameTaken        = errors.New("昵称已被使用")
	ErrImageNotFound        = errors.New("图片不存在")
	ErrImageNotOwned        = errors.New("该图片不属于您")
	ErrGameNotFound         = errors.New("游戏不存在")
	ErrSessionNotFound      = errors.New("会话不存在")
	ErrLevelNotFound        = errors.New("关卡不存在")
	ErrSessionLevelNotFound = errors.New("关卡会话不存在")
	ErrNoGameLevels         = errors.New("游戏没有关卡")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidPlayTime      = errors.New("无效的游玩时间")
	ErrInsufficientBeans    = errors.New("能量豆不足")
	ErrRedeemNotFound       = errors.New("兑换码不存在")
	ErrRedeemAlreadyUsed    = errors.New("兑换码已使用")
	ErrContentSeekExists    = errors.New("内容征集已存在")
	ErrFileTooLarge         = errors.New("文件大小不能超过2MB")
	ErrInvalidFileType      = errors.New("仅支持JPEG和PNG格式")
	ErrInvalidImageRole     = errors.New("无效的图片类型")
	ErrGamePublished        = errors.New("已发布的游戏不可编辑")
	ErrGameAlreadyPublished = errors.New("游戏已经是发布状态")
	ErrGameNotPublished     = errors.New("游戏未发布")
	ErrMetaNotFound         = errors.New("内容元数据不存在")
	ErrContentItemNotFound  = errors.New("练习单元不存在")
	ErrCapacityExceeded     = errors.New("超出关卡内容上限")
	ErrItemLimitExceeded    = errors.New("每条元数据练习单元数量已达上限")
)
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/errors.go
git commit -m "feat: translate error sentinels to Chinese"
```

---

### Task 2: Translate validation messages

**Files:**
- Modify: `dx-api/app/http/requests/api/auth_request.go`
- Modify: `dx-api/app/http/requests/api/user_request.go`
- Modify: `dx-api/app/http/requests/api/user_redeem_request.go`
- Modify: `dx-api/app/http/requests/api/feedback_request.go`
- Modify: `dx-api/app/http/requests/api/content_seek_request.go`

- [ ] **Step 1: Translate auth_request.go**

In `SignUpRequest.Messages()`:
```go
"code.required": "请输入6位验证码",
"code.len":      "请输入6位验证码",
```

- [ ] **Step 2: Translate user_request.go**

In `UpdateProfileRequest.Messages()`:
```go
"nickname.max_len":     "昵称不能超过20个字符",
"city.max_len":         "城市不能超过50个字符",
"introduction.max_len": "简介不能超过200个字符",
```

In `ChangeEmailRequest.Messages()`:
```go
"code.required": "请输入6位验证码",
"code.len":      "请输入6位验证码",
```

In `ChangePasswordRequest.Messages()`:
```go
"new_password.min_len": "新密码至少需要8个字符",
```

- [ ] **Step 3: Translate user_redeem_request.go**

In `RedeemCodeRequest.Messages()`:
```go
"code.len": "兑换码格式不正确",
```

- [ ] **Step 4: Translate feedback_request.go**

In `SubmitFeedbackRequest.Messages()`:
```go
"description.max_len": "描述不能超过200个字符",
```

- [ ] **Step 5: Translate content_seek_request.go**

In `SubmitContentSeekRequest.Messages()`:
```go
"course_name.max_len": "课程名称不能超过30个字符",
"description.max_len": "描述不能超过30个字符",
"disk_url.max_len":    "网盘链接不能超过30个字符",
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/http/requests/api/auth_request.go \
        dx-api/app/http/requests/api/user_request.go \
        dx-api/app/http/requests/api/user_redeem_request.go \
        dx-api/app/http/requests/api/feedback_request.go \
        dx-api/app/http/requests/api/content_seek_request.go
git commit -m "feat: translate validation messages to Chinese"
```

---

### Task 3: Translate controller messages — auth & user

**Files:**
- Modify: `dx-api/app/http/controllers/api/auth_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_controller.go`

- [ ] **Step 1: Translate auth_controller.go**

Replace these specific strings (leave `"unauthorized"`, `"failed to ..."`, `"refresh token required"`, `"invalid or expired refresh token"` untouched):

| Find | Replace |
|------|---------|
| `"please wait before requesting another code"` | `"请稍后再请求验证码"` |
| `"invalid or expired verification code"` | `"验证码无效或已过期"` |
| `"email already registered"` | `"该邮箱已注册"` |
| `"username already taken"` | `"用户名已被使用"` |
| `"email or account is required"` | `"请输入邮箱或账号"` |
| `"user not found"` | `"用户不存在"` |
| `"invalid password"` | `"密码错误"` |
| `"invalid request"` | `"无效的请求"` |
| `"too many refresh requests"` | `"刷新请求过于频繁"` |

- [ ] **Step 2: Translate user_controller.go**

| Find | Replace |
|------|---------|
| `"nickname already taken"` | `"昵称已被使用"` |
| `"image not found"` | `"图片不存在"` |
| `"image does not belong to you"` | `"该图片不属于您"` |
| `"please wait before requesting another code"` | `"请稍后再请求验证码"` |
| `"email already registered"` | `"该邮箱已注册"` |
| `"invalid or expired verification code"` | `"验证码无效或已过期"` |
| `"current password is incorrect"` | `"当前密码错误"` |

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/auth_controller.go \
        dx-api/app/http/controllers/api/user_controller.go
git commit -m "feat: translate auth and user controller messages to Chinese"
```

---

### Task 4: Translate controller messages — game, session & upload

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_controller.go`
- Modify: `dx-api/app/http/controllers/api/game_session_controller.go`
- Modify: `dx-api/app/http/controllers/api/game_report_controller.go`
- Modify: `dx-api/app/http/controllers/api/upload_controller.go`

- [ ] **Step 1: Translate game_controller.go**

| Find | Replace |
|------|---------|
| `"game not found"` | `"游戏不存在"` |

- [ ] **Step 2: Translate game_session_controller.go**

Leave `"unauthorized"`, `"failed to ..."`, `"X is required"` untouched. Translate only:

| Find | Replace |
|------|---------|
| `"game has no levels"` | `"游戏没有关卡"` |
| `"session level not found"` | `"关卡会话不存在"` |
| `"session not found"` | `"会话不存在"` |
| `"invalid request"` | `"无效的请求"` |
| `"play_time must be between 0 and 86400"` | `"游玩时长必须在0到86400秒之间"` |

- [ ] **Step 3: Translate game_report_controller.go**

| Find | Replace |
|------|---------|
| `"too many reports, please try again later"` | `"举报过于频繁，请稍后再试"` |

- [ ] **Step 4: Translate upload_controller.go**

| Find | Replace |
|------|---------|
| `"file size exceeds 2MB limit"` | `"文件大小不能超过2MB"` |
| `"only JPEG and PNG files are allowed"` | `"仅支持JPEG和PNG格式"` |
| `"invalid image role"` | `"无效的图片类型"` |
| `"image not found"` | `"图片不存在"` |
| `"image file not found"` | `"图片文件不存在"` |

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/http/controllers/api/game_controller.go \
        dx-api/app/http/controllers/api/game_session_controller.go \
        dx-api/app/http/controllers/api/game_report_controller.go \
        dx-api/app/http/controllers/api/upload_controller.go
git commit -m "feat: translate game, session, and upload controller messages to Chinese"
```

---

### Task 5: Translate controller messages — remaining controllers

**Files:**
- Modify: `dx-api/app/http/controllers/api/user_redeem_controller.go`
- Modify: `dx-api/app/http/controllers/api/leaderboard_controller.go`
- Modify: `dx-api/app/http/controllers/api/hall_controller.go`
- Modify: `dx-api/app/http/controllers/api/user_referral_controller.go`
- Modify: `dx-api/app/http/controllers/api/ai_custom_controller.go`

- [ ] **Step 1: Translate user_redeem_controller.go**

| Find | Replace |
|------|---------|
| `"redeem code not found"` | `"兑换码不存在"` |
| `"redeem code already used"` | `"兑换码已使用"` |

- [ ] **Step 2: Translate leaderboard_controller.go**

| Find | Replace |
|------|---------|
| `"type must be exp or playtime"` | `"类型必须是经验值或游玩时长"` |
| `"period must be all, day, week, or month"` | `"时间范围必须是全部、日、周或月"` |

- [ ] **Step 3: Translate hall_controller.go**

| Find | Replace |
|------|---------|
| `"user not found"` | `"用户不存在"` |
| `"invalid year"` | `"无效的年份"` |

- [ ] **Step 4: Translate user_referral_controller.go**

| Find | Replace |
|------|---------|
| `"user not found"` | `"用户不存在"` |

- [ ] **Step 5: Translate ai_custom_controller.go**

Only 2 messages remain English (rest already Chinese):

| Find | Replace |
|------|---------|
| `"content is required"` | `"请输入内容"` |
| `"formatType must be sentence or vocab"` | `"格式类型必须是句子或词汇"` |

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/http/controllers/api/user_redeem_controller.go \
        dx-api/app/http/controllers/api/leaderboard_controller.go \
        dx-api/app/http/controllers/api/hall_controller.go \
        dx-api/app/http/controllers/api/user_referral_controller.go \
        dx-api/app/http/controllers/api/ai_custom_controller.go
git commit -m "feat: translate remaining controller messages to Chinese"
```

---

### Task 6: Final verification

- [ ] **Step 1: Full build check**

Run: `cd dx-api && go build ./...`

- [ ] **Step 2: Grep for remaining English business messages**

Run: `grep -rn '".*not found"' dx-api/app/http/controllers/api/ | grep -v unauthorized | grep -v "failed to"` to check for stragglers.

Run: `grep -rn '".*already"' dx-api/app/http/controllers/api/ | grep -v "failed to"` to verify all "already" messages translated.

- [ ] **Step 3: Verify no adm/ files were touched**

Run: `git diff --name-only | grep adm` — should produce no output.
