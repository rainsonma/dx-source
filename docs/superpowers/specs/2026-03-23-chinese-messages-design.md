# Chinese Client Messages Design

Replace English user-facing messages with Chinese in the `/api/*` backend layer. No i18n, no structural changes — inline string replacement only.

## Scope

**In scope:** `/api/*` client-facing user-visible messages only.

**Out of scope:**
- `/adm/*` admin routes — stay English
- Middleware messages (`"unauthorized"`, `"forbidden"`, `"too many requests"`)
- `"ok"` success response
- `"failed to ..."` patterns — frontend shows its own Chinese fallback
- `"internal server error"` — system-level
- Auth token messages (`"refresh token required"`, `"invalid or expired refresh token"`)
- `"X is required"` validation errors from URL params — indicate frontend bugs, not user input

## Changes

### 1. Error Sentinels (`services/api/errors.go`)

| Variable | Current | Chinese |
|----------|---------|---------|
| `ErrInvalidCode` | `"invalid or expired verification code"` | `"验证码无效或已过期"` |
| `ErrDuplicateEmail` | `"email already registered"` | `"该邮箱已注册"` |
| `ErrDuplicateUsername` | `"username already taken"` | `"用户名已被使用"` |
| `ErrUserNotFound` | `"user not found"` | `"用户不存在"` |
| `ErrInvalidPassword` | `"invalid password"` | `"密码错误"` |
| `ErrNicknameTaken` | `"nickname already taken"` | `"昵称已被使用"` |
| `ErrImageNotFound` | `"image not found"` | `"图片不存在"` |
| `ErrImageNotOwned` | `"image not owned by user"` | `"该图片不属于您"` |
| `ErrGameNotFound` | `"game not found"` | `"游戏不存在"` |
| `ErrSessionNotFound` | `"session not found"` | `"会话不存在"` |
| `ErrLevelNotFound` | `"level not found"` | `"关卡不存在"` |
| `ErrSessionLevelNotFound` | `"session level not found"` | `"关卡会话不存在"` |
| `ErrNoGameLevels` | `"game has no levels"` | `"游戏没有关卡"` |
| `ErrInvalidPlayTime` | `"invalid play time"` | `"无效的游玩时间"` |
| `ErrInsufficientBeans` | `"insufficient beans"` | `"能量豆不足"` |
| `ErrRedeemNotFound` | `"redeem code not found"` | `"兑换码不存在"` |
| `ErrRedeemAlreadyUsed` | `"redeem code already used"` | `"兑换码已使用"` |
| `ErrContentSeekExists` | `"content seek already exists"` | `"内容征集已存在"` |
| `ErrFileTooLarge` | `"file size exceeds 2MB limit"` | `"文件大小不能超过2MB"` |
| `ErrInvalidFileType` | `"only JPEG and PNG files are allowed"` | `"仅支持JPEG和PNG格式"` |
| `ErrInvalidImageRole` | `"invalid image role"` | `"无效的图片类型"` |
| `ErrGamePublished` | `"published game cannot be edited"` | `"已发布的游戏不可编辑"` |
| `ErrGameAlreadyPublished` | `"game is already published"` | `"游戏已经是发布状态"` |
| `ErrGameNotPublished` | `"game is not published"` | `"游戏未发布"` |
| `ErrMetaNotFound` | `"content metadata not found"` | `"内容元数据不存在"` |
| `ErrContentItemNotFound` | `"content item not found"` | `"练习单元不存在"` |
| `ErrCapacityExceeded` | `"level content capacity exceeded"` | `"超出关卡内容上限"` |
| `ErrItemLimitExceeded` | `"content item limit per metadata exceeded"` | `"每条元数据练习单元数量已达上限"` |

**Keep English** (technical/auth — frontend intercepts):
- `ErrInvalidRefreshToken`, `ErrSessionReplaced`, `ErrRateLimited`, `ErrForbidden`

### 2. Validation Messages (`requests/api/*.go`)

**auth_request.go:**
| Rule | Current | Chinese |
|------|---------|---------|
| `code.required` | `"a 6-digit verification code is required"` | `"请输入6位验证码"` |
| `code.len` | `"a 6-digit verification code is required"` | `"请输入6位验证码"` |

**user_request.go:**
| Rule | Current | Chinese |
|------|---------|---------|
| `nickname.max_len` | `"nickname must be at most 20 characters"` | `"昵称不能超过20个字符"` |
| `city.max_len` | `"city must be at most 50 characters"` | `"城市不能超过50个字符"` |
| `introduction.max_len` | `"introduction must be at most 200 characters"` | `"简介不能超过200个字符"` |
| `new_password.min_len` | `"new password must be at least 8 characters"` | `"新密码至少需要8个字符"` |
| `code.required` (ChangeEmailRequest) | `"a 6-digit verification code is required"` | `"请输入6位验证码"` |
| `code.len` (ChangeEmailRequest) | `"a 6-digit verification code is required"` | `"请输入6位验证码"` |

**user_redeem_request.go:**
| Rule | Current | Chinese |
|------|---------|---------|
| `code.len` | `"invalid redeem code format"` | `"兑换码格式不正确"` |

**feedback_request.go:**
| Rule | Current | Chinese |
|------|---------|---------|
| `description.max_len` | `"description must be at most 200 characters"` | `"描述不能超过200个字符"` |

**content_seek_request.go:**
| Rule | Current | Chinese |
|------|---------|---------|
| `course_name.max_len` | `"course name must be at most 30 characters"` | `"课程名称不能超过30个字符"` |
| `description.max_len` | `"description must be at most 30 characters"` | `"描述不能超过30个字符"` |
| `disk_url.max_len` | `"disk url must be at most 30 characters"` | `"网盘链接不能超过30个字符"` |

### 3. Controller Messages (`controllers/api/*.go`)

Only translate user-visible business messages. Keep `"unauthorized"`, `"forbidden"`, `"failed to ..."`, `"internal server error"`, and `"X is required"` in English.

**auth_controller.go:**
| Current | Chinese |
|---------|---------|
| `"please wait before requesting another code"` | `"请稍后再请求验证码"` |
| `"invalid or expired verification code"` | `"验证码无效或已过期"` |
| `"email already registered"` | `"该邮箱已注册"` |
| `"username already taken"` | `"用户名已被使用"` |
| `"email or account is required"` | `"请输入邮箱或账号"` |
| `"user not found"` | `"用户不存在"` |
| `"invalid password"` | `"密码错误"` |
| `"invalid request"` | `"无效的请求"` |
| `"too many refresh requests"` | `"刷新请求过于频繁"` |
| `"user not found"` (Me endpoint) | `"用户不存在"` |

**user_controller.go:**
| Current | Chinese |
|---------|---------|
| `"nickname already taken"` | `"昵称已被使用"` |
| `"image not found"` | `"图片不存在"` |
| `"image does not belong to you"` | `"该图片不属于您"` |
| `"please wait before requesting another code"` | `"请稍后再请求验证码"` |
| `"email already registered"` | `"该邮箱已注册"` |
| `"invalid or expired verification code"` | `"验证码无效或已过期"` |
| `"current password is incorrect"` | `"当前密码错误"` |

**upload_controller.go:**
| Current | Chinese |
|---------|---------|
| `"file size exceeds 2MB limit"` | `"文件大小不能超过2MB"` |
| `"only JPEG and PNG files are allowed"` | `"仅支持JPEG和PNG格式"` |
| `"invalid image role"` | `"无效的图片类型"` |
| `"image not found"` | `"图片不存在"` |
| `"image file not found"` | `"图片文件不存在"` |

**game_session_controller.go:**
| Current | Chinese |
|---------|---------|
| `"game has no levels"` | `"游戏没有关卡"` |
| `"session level not found"` | `"关卡会话不存在"` |
| `"session not found"` | `"会话不存在"` |
| `"invalid request"` | `"无效的请求"` |
| `"play_time must be between 0 and 86400"` | `"游玩时长必须在0到86400秒之间"` |

**game_controller.go:**
| Current | Chinese |
|---------|---------|
| `"game not found"` | `"游戏不存在"` |

**game_report_controller.go:**
| Current | Chinese |
|---------|---------|
| `"too many reports, please try again later"` | `"举报过于频繁，请稍后再试"` |

**user_redeem_controller.go:**
| Current | Chinese |
|---------|---------|
| `"redeem code not found"` | `"兑换码不存在"` |
| `"redeem code already used"` | `"兑换码已使用"` |

**leaderboard_controller.go:**
| Current | Chinese |
|---------|---------|
| `"type must be exp or playtime"` | `"类型必须是经验值或游玩时长"` |
| `"period must be all, day, week, or month"` | `"时间范围必须是全部、日、周或月"` |

**hall_controller.go:**
| Current | Chinese |
|---------|---------|
| `"user not found"` | `"用户不存在"` |
| `"invalid year"` | `"无效的年份"` |

**user_referral_controller.go:**
| Current | Chinese |
|---------|---------|
| `"user not found"` | `"用户不存在"` |

**ai_custom_controller.go:**
| Current | Chinese |
|---------|---------|
| `"content is required"` | `"请输入内容"` |
| `"formatType must be sentence or vocab"` | `"格式类型必须是句子或词汇"` |

### 4. Already Chinese (no change needed)

- `course_game_controller.go` — `mapCourseGameError` already maps all sentinels to Chinese
- `ai_custom_controller.go` — most messages already Chinese; only `"content is required"` and `"formatType must be sentence or vocab"` remain English (listed above)
- Rate-limit messages in `user_unknown_controller.go`, `user_master_controller.go`, `user_review_controller.go`, `game_session_controller.go` — already `"操作过于频繁，请稍后再试"`

### 5. Files NOT Changed

- `middleware/*.go` — all stay English
- `controllers/adm/*.go` — all stay English
- `requests/adm/*.go` — all stay English
- `helpers/response.go` — `"ok"` stays
- Frontend (`dx-web/`) — already Chinese

## Testing

- `go build ./...` — verify compilation
- Manual spot-check of key API responses
