# Request Validation Improvement Design

## Problem

The dx-api request validation layer is weak:
- ~17 of ~44 request structs have zero validation rules
- Only 5 distinct rules used (`required`, `max_len`, `min_len`, `len`, `in`)
- No email format, UUID, enum, or numeric range validation
- No `Filters()` usage (text inputs not trimmed)
- No `Messages()` on ~70% of structs
- Inconsistent error message language
- ~11 structs are plain structs without FormRequest methods (`Authorize()` / `Rules()`)

## Goals

- Every request struct implements `Authorize()`, `Rules()`, `Messages()`, and `Filters()` where applicable
- Validate ID fields as UUID, enums against `consts/` values, numeric fields with ranges
- Chinese messages for all api/ requests, English for all adm/ requests
- Create one custom rule (`strong_password`), use a helper function for enum validation
- Preserve the unified response format — `helpers.Validate()` and `{ "code", "message", "data" }` envelope unchanged

## Constraint: Unified Response Format

The existing `helpers.Validate()` function is NOT modified:

```go
// helpers/response.go — unchanged
func Validate(ctx http.Context, req http.FormRequest) http.Response {
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

**Controller changes:** Controllers that already use `helpers.Validate()` keep their calling pattern unchanged. Controllers that currently do manual query-param parsing (game_controller, parts of session_controller) will be updated to use `helpers.Validate()` — this is a necessary part of converting plain structs to FormRequest.

## Section 1: Custom Rules & Enum Helper Infrastructure

### Custom Rule (`app/rules/`)

**`strong_password.go`** — Password complexity policy
- Signature: `strong_password`
- Passes: value contains uppercase + lowercase + digit + special character
- Message (zh): `:attribute 必须包含大写字母、小写字母、数字和特殊字符`

### Registration

**`bootstrap/rules.go`** (new) — registers `strong_password` via `WithRules()`

### Enum Helper (`app/helpers/enum_rules.go`)

Instead of a custom `enum` rule (which would clash with Goravel's built-in `enum` alias for `in`), use a helper function that generates `in:val1,val2,...` strings from consts:

```go
package helpers

import (
    "strings"
    "github.com/example/dx-api/app/consts"
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

func InEnum(name string) string {
    vals, ok := enumValues[name]
    if !ok {
        panic("unknown enum: " + name) // fail fast on typos
    }
    return "in:" + strings.Join(vals, ",")
}
```

Usage in Rules():
```go
"degree": helpers.InEnum("degree")
// produces: "in:practice,beginner,intermediate,advanced"
```

This leverages Goravel's built-in `in:` rule — no custom rule registration needed. Adding a new enum value in `consts/` automatically updates validation everywhere.

### Built-in Rules to Leverage

| Rule | Purpose | Fields |
|------|---------|--------|
| `uuid` | Validate UUID format | All `*_id` / `*Id` fields (~25) |
| `email` | Validate email format | Email fields in auth/user |
| `min:N` / `max:N` | Numeric bounds | Scores, combos, play times, limits |
| `in:values` | Fixed enums (via `InEnum()`) | Degree, pattern, mode, feedback type, etc. |
| `full_url` | URL format | disk_url in content_seek |
| `alpha_dash` | Username format | Username in signup |
| `min_len:N` / `max_len:N` | String length | Names, descriptions, notes |
| `required_without` | Conditional required | SignIn dual-mode auth |

### Filters Strategy

Add `Filters()` to all request structs with text input:
- `trim` on: nickname, city, introduction, email, course_name, description, title, content, note, reason
- `trim|upper` on: redeem code

### FormRequest Conversion

10 structs are currently plain structs (no `Authorize()` / `Rules()` methods). These will be converted to full FormRequest implementations:
- `CheckActiveSessionRequest`, `CheckActiveLevelSessionRequest`, `CompleteLevelRequest`
- `AdvanceLevelRequest`, `UpdateContentItemRequest`, `RestoreSessionRequest`
- `ListGamesRequest`, `SearchGamesRequest`, `LevelContentRequest`
- `UpdateContentItemTextRequest`

`RestartLevelRequest` is dead code (controller reads from route param, never binds the struct) — remove it.

Each converted struct gets `Authorize() error { return nil }` plus `Rules()`, `Messages()`, and `Filters()` as specified below. Controllers that currently do manual parsing for these structs will be updated to use `helpers.Validate()`.

**Route param vs body distinction:** Some session controllers read IDs from route params (`ctx.Request().Route("levelId")`) rather than the request body. These fields are NOT validated by the request struct — the route provides them. This is noted per-struct in Section 5.

## Section 2: Auth Requests (`api/auth_request.go`)

### `SendCodeRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| email | `required` | `required\|email` |
- Filters: `"email": "trim"`
- Messages: `"email.required": "请输入邮箱地址"`, `"email.email": "邮箱地址格式不正确"`

### `SignUpRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| email | `required` | `required\|email` |
| code | `required\|len:6` | `required\|len:6` |
| username | (none) | `required\|alpha_dash\|min_len:3\|max_len:20` |
| password | (none) | `required\|min_len:8\|strong_password` |
- Filters: `"email": "trim"`, `"username": "trim"`
- Messages: Chinese for all — email format, code length, username constraints, password strength

### `SignInRequest`
Currently a plain struct with no FormRequest methods. Adding `Authorize()` + `Rules()`.
| Field | Current | Proposed |
|-------|---------|----------|
| email | (none) | `required_without:account\|email` |
| code | (none) | `required_without:password\|len:6` |
| account | (none) | `required_without:email\|min_len:3` |
| password | (none) | `required_without:code\|min_len:8` |
- Filters: `"email": "trim"`, `"account": "trim"`
- Messages: Chinese for all
- Note: Controller still validates the pairing logic (email+code OR account+password). The request-level rules validate individual field formats; the controller validates that exactly one auth method is provided.

## Section 3: User Requests (`api/user_request.go`)

### `UpdateProfileRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| nickname | `max_len:20` | `max_len:20` (unchanged) |
| city | `max_len:50` | `max_len:50` (unchanged) |
| introduction | `max_len:200` | `max_len:200` (unchanged) |
- Filters: `"nickname": "trim"`, `"city": "trim"`, `"introduction": "trim"`
- Messages: keep existing Chinese messages

### `UpdateAvatarRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| image_id | `required` | `required\|uuid` |
- Messages: `"image_id.required": "请选择头像"`, `"image_id.uuid": "无效的图片ID"`

### `SendEmailCodeRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| email | `required` | `required\|email` |
- Filters: `"email": "trim"`
- Messages: `"email.required": "请输入邮箱地址"`, `"email.email": "邮箱地址格式不正确"`

### `ChangeEmailRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| email | `required` | `required\|email` |
| code | `required\|len:6` | `required\|len:6` (unchanged) |
- Filters: `"email": "trim"`
- Messages: keep existing Chinese code messages, add email format message

### `ChangePasswordRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| current_password | `required` | `required` (unchanged) |
| new_password | `required\|min_len:8` | `required\|min_len:8\|strong_password` |
- Messages: keep existing min_len message, add `"new_password.strong_password": "新密码必须包含大写字母、小写字母、数字和特殊字符"`

## Section 4: Game Requests (`api/game_request.go`)

All three structs are plain structs — converting to FormRequest. Controllers will switch from manual parsing to `helpers.Validate()`.

**Note on form tags:** These structs use camelCase form tags: `pressId` (not `press_id`), `categoryIds`, `q` (not `query`). Rule keys must match the form/json tag names.

### `ListGamesRequest`
| Field (form tag) | Current | Proposed |
|-------|---------|----------|
| cursor | (manual) | (optional, no rule) |
| limit | (manual) | `min:1\|max:50` |
| categoryIds | (manual) | (optional, split in service) |
| pressId | (manual) | `uuid` |
| mode | (manual) | `helpers.InEnum("mode")` |
- Filters: `"cursor": "trim"`, `"pressId": "trim"`
- Messages: `"limit.min": "每页数量不能小于1"`, `"limit.max": "每页数量不能超过50"`, `"pressId.uuid": "无效的出版社ID"`, `"mode.in": "无效的游戏模式"`

### `SearchGamesRequest`
| Field (form tag) | Current | Proposed |
|-------|---------|----------|
| q | (manual) | `required\|min_len:1\|max_len:50` |
| limit | (manual) | `min:1\|max:50` |
- Filters: `"q": "trim"`
- Messages: `"q.required": "请输入搜索关键词"`, `"q.max_len": "搜索关键词不能超过50个字符"`, `"limit.min": "每页数量不能小于1"`, `"limit.max": "每页数量不能超过50"`

### `LevelContentRequest`
| Field (form tag) | Current | Proposed |
|-------|---------|----------|
| degree | (none) | `helpers.InEnum("degree")` |
- Messages: `"degree.in": "无效的难度级别"`

## Section 5: Session Requests (`api/session_request.go`)

**Note on pointer fields:** `Pattern *string` and `LevelID *string` are optional pointer fields. Rules without `required` only fire when the field is present.

### `StartSessionRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | `required` | `required\|uuid` |
| degree | (none) | `helpers.InEnum("degree")` |
| pattern | (none) | `helpers.InEnum("pattern")` |
| level_id | (none) | `uuid` |
- Filters: `"degree": "trim"`, `"pattern": "trim"`
- PrepareForValidation: keep existing degree default logic
- Messages: Chinese for all

### `CheckActiveSessionRequest` (plain struct → FormRequest)
**Controller note:** Controller currently reads from `ctx.Request().Query()` manually, including nil-pointer construction for optional `pattern`. Convert to FormRequest and update controller to use `helpers.Validate()`. The `Pattern *string` pointer field handles optionality naturally — Goravel leaves it nil when absent.
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | (none) | `required\|uuid` |
| degree | (none) | `helpers.InEnum("degree")` |
| pattern | (none) | `helpers.InEnum("pattern")` |
- Messages: Chinese

### `CheckActiveLevelSessionRequest` (plain struct → FormRequest)
**Controller note:** Same as `CheckActiveSessionRequest` — controller reads from query params manually. Convert to FormRequest.
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | (none) | `required\|uuid` |
| degree | (none) | `helpers.InEnum("degree")` |
| pattern | (none) | `helpers.InEnum("pattern")` |
| game_level_id | (none) | `required\|uuid` |
- Messages: Chinese

### `StartLevelRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| game_level_id | `required` | `required\|uuid` |
| degree | (none) | `helpers.InEnum("degree")` |
| pattern | (none) | `helpers.InEnum("pattern")` |
- PrepareForValidation: keep existing degree default logic
- Messages: Chinese

### `CompleteLevelRequest` (plain struct → FormRequest)
Actual struct fields: `GameLevelID` (string), `Score` (int), `MaxCombo` (int), `TotalItems` (int).
**Controller note:** `game_level_id` comes from the route param `ctx.Request().Route("levelId")`, not from the request body. Do NOT add a `required` rule for it — the route provides it.
| Field | Current | Proposed |
|-------|---------|----------|
| score | (none) | `min:0` |
| max_combo | (none) | `min:0` |
| total_items | (none) | `min:0` |
- Messages: `"score.min": "分数不能为负数"`, `"max_combo.min": "连击数不能为负数"`, `"total_items.min": "总数不能为负数"`

### `AdvanceLevelRequest` (plain struct → FormRequest)
**Controller note:** The controller has fallback logic — if `next_level_id` is empty in the body, it falls back to the route param `levelId`. Do NOT mark as `required`; keep it optional with format-only validation.
| Field | Current | Proposed |
|-------|---------|----------|
| next_level_id | (none) | `uuid` |
- Messages: `"next_level_id.uuid": "无效的关卡ID"`

### `RestartLevelRequest` (dead code — remove struct)
**Controller note:** The controller reads `game_level_id` from the route param `ctx.Request().Route("levelId")` and never binds this struct. This is dead code. Remove the struct during implementation.

### `RecordAnswerRequest` (already has FormRequest methods)
Actual struct fields: `GameSessionLevelID`, `GameLevelID`, `ContentItemID`, `IsCorrect` (bool), `UserAnswer` (string), `SourceAnswer` (string), `BaseScore` (int), `ComboScore` (int), `Score` (int), `MaxCombo` (int), `PlayTime` (int), `NextContentItemID` (*string), `Duration` (int).
| Field | Current | Proposed |
|-------|---------|----------|
| game_session_level_id | `required` | `required\|uuid` |
| game_level_id | `required` | `required\|uuid` |
| content_item_id | `required` | `required\|uuid` |
| base_score | (none) | `min:0` |
| combo_score | (none) | `min:0` |
| score | (none) | `min:0` |
| max_combo | (none) | `min:0` |
| play_time | (none) | `min:0` |
| duration | (none) | `min:0` |
| next_content_item_id | (none) | `uuid` |
- Messages: Chinese for required + uuid + min

### `RecordSkipRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| game_level_id | `required` | `required\|uuid` |
| play_time | (none) | `min:0` |
| next_content_item_id | (none) | `uuid` |
- Messages: Chinese

### `SyncPlayTimeRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| game_level_id | `required` | `required\|uuid` |
| play_time | (none) | `required\|min:0` |
- Messages: Chinese

### `UpdateContentItemRequest` (plain struct → FormRequest)
| Field | Current | Proposed |
|-------|---------|----------|
| content_item_id | (none) | `uuid` |
- Messages: `"content_item_id.uuid": "无效的内容ID"`

### `EndSessionRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | `required` | `required\|uuid` |
| score | (none) | `min:0` |
| exp | (none) | `min:0` |
| max_combo | (none) | `min:0` |
| correct_count | (none) | `min:0` |
| wrong_count | (none) | `min:0` |
| skip_count | (none) | `min:0` |
- Messages: Chinese

### `RestoreSessionRequest` (plain struct → FormRequest)
**Controller note:** The controller currently reads `game_level_id` from `ctx.Request().Query()`. Convert to FormRequest and update controller to use `helpers.Validate()`.
| Field | Current | Proposed |
|-------|---------|----------|
| game_level_id | (none) | `required\|uuid` |
- Messages: `"game_level_id.required": "请指定关卡"`, `"game_level_id.uuid": "无效的关卡ID"`

## Section 6: Other API Requests

### `MarkMasteredRequest`, `MarkReviewRequest`, `MarkUnknownRequest`
All three identical:
| Field | Current | Proposed |
|-------|---------|----------|
| content_item_id | `required` | `required\|uuid` |
| game_id | `required` | `required\|uuid` |
| game_level_id | `required` | `required\|uuid` |
- Messages: Chinese for required + uuid

### `BulkDeleteRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| ids | `required\|min_len:1` | `required\|min_len:1\|max_len:100` |
| ids.* | (none) | `uuid` |
- Messages: `"ids.required": "请选择要删除的项目"`, `"ids.max_len": "单次最多删除100条"`, `"ids.*.uuid": "包含无效的ID"`

### `ToggleFavoriteRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | `required` | `required\|uuid` |
- Messages: `"game_id.required": "请选择游戏"`, `"game_id.uuid": "无效的游戏ID"`

### `SubmitFeedbackRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| type | `required` | `required\|` + `helpers.InEnum("feedback_type")` |
| description | `required\|max_len:200` | `required\|min_len:2\|max_len:200` |
- Filters: `"description": "trim"`
- Messages: keep existing max_len message, add `"type.in": "无效的反馈类型"`, `"description.min_len": "反馈内容不能少于2个字符"`

### `SubmitReportRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| game_id | `required` | `required\|uuid` |
| game_level_id | `required` | `required\|uuid` |
| content_item_id | `required` | `required\|uuid` |
| reason | `required` | `required\|max_len:200` |
| note | (none) | `max_len:500` |
- Filters: `"reason": "trim"`, `"note": "trim"`
- Messages: Chinese for all

### `UploadImageRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| role | `required` | `required\|` + `helpers.InEnum("image_role")` |
- Messages: `"role.required": "请指定图片用途"`, `"role.in": "无效的图片用途"`

### `RedeemCodeRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| code | `required\|len:19` | `required\|len:19` (unchanged) |
- Filters: `"code": "trim\|upper"`
- Messages: keep existing Chinese message

### `SubmitContentSeekRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| course_name | `required\|max_len:30` | `required\|min_len:2\|max_len:30` |
| description | `required\|max_len:30` | `required\|min_len:2\|max_len:200` |
| disk_url | `required\|max_len:30` | `required\|full_url\|max_len:500` |
- Filters: `"course_name": "trim"`, `"description": "trim"`, `"disk_url": "trim"`
- Messages: Chinese, including `"disk_url.full_url": "请输入有效的网盘链接"`
- Note: `description` max_len intentionally relaxed from 30→200 (30 is too tight for a description). `disk_url` max_len raised to 500 and validated as URL.

## Section 7: Course Game Requests (`api/course_game_request.go`)

**Note on form/json tags:** These structs use camelCase json tags: `gameLevelId`, `gameCategoryId`, `gamePressId`, `gameMode`, `coverId`, `contentMetaId`, `contentType`, `referenceItemId`, `sourceFrom`, `sourceData`, `sourceType`, `metaId`, `itemId`, `newOrder`. Rule keys must match these json tags exactly.

**Note on float64 fields:** `NewOrder` is `float64` in `ReorderMetadataRequest` and `ReorderContentItemRequest`. The `min:0` rule works correctly on float64 values.

### `CreateGameRequest` (already has FormRequest methods)
| Field (json tag) | Current | Proposed |
|-------|---------|----------|
| name | `required` | `required\|min_len:2\|max_len:100` |
| description | (none) | `max_len:500` |
| gameMode | `required` | `required\|` + `helpers.InEnum("mode")` |
| gameCategoryId | `required` | `required\|uuid` |
| gamePressId | `required` | `required\|uuid` |
| coverId | (none) | `uuid` |
- Filters: `"name": "trim"`, `"description": "trim"`
- Messages: Chinese for all

### `UpdateGameRequest` (already has FormRequest methods)
Same rules as `CreateGameRequest`. This is intentional — the endpoint is PUT (full replacement), not PATCH. The struct has the same fields as Create, so all fields should be required on update.

### `CreateLevelRequest` (already has FormRequest methods)
| Field | Current | Proposed |
|-------|---------|----------|
| name | `required` | `required\|min_len:1\|max_len:100` |
| description | (none) | `max_len:500` |
- Filters: `"name": "trim"`, `"description": "trim"`
- Messages: Chinese

### `SaveMetadataBatchRequest` (already has FormRequest methods)
| Field (json tag) | Current | Proposed |
|-------|---------|----------|
| entries | `required\|min_len:1` | `required\|min_len:1\|max_len:200` |
| gameLevelId | (none) | `required\|uuid` |
| sourceFrom | (none) | `required\|` + `helpers.InEnum("source_from")` |
| entries.*.sourceData | (none) | `required` |
| entries.*.sourceType | (none) | `required\|` + `helpers.InEnum("source_type")` |
- Filters: `"entries.*.sourceData": "trim"`, `"entries.*.translation": "trim"`
- Messages: Chinese for all including nested: `"entries.max_len": "单次最多提交200条"`, `"entries.*.sourceData.required": "每条数据的内容不能为空"`, `"entries.*.sourceType.in": "无效的内容类型"`

### `ReorderMetadataRequest` (already has FormRequest methods)
| Field (json tag) | Current | Proposed |
|-------|---------|----------|
| metaId | `required` | `required\|uuid` |
| gameLevelId | (none) | `required\|uuid` |
| newOrder | (none) | `required\|min:0` |
- Messages: `"metaId.required": "请指定元数据"`, `"newOrder.required": "请指定排序位置"`, `"newOrder.min": "排序位置不能为负数"`

### `InsertContentItemRequest` (already has FormRequest methods)
| Field (json tag) | Current | Proposed |
|-------|---------|----------|
| contentMetaId | `required` | `required\|uuid` |
| gameLevelId | (none) | `required\|uuid` |
| contentType | (none) | `helpers.InEnum("content_type")` |
| direction | (none) | `in:before,after` |
| referenceItemId | (none) | `uuid` |
- Filters: `"content": "trim"`, `"translation": "trim"`
- Messages: `"contentMetaId.required": "请指定元数据"`, `"contentType.in": "无效的内容类型"`, `"direction.in": "插入方向只能为前或后"`

### `UpdateContentItemTextRequest` (plain struct → FormRequest)
| Field | Current | Proposed |
|-------|---------|----------|
| content | (none) | `max_len:2000` |
| translation | (none) | `max_len:2000` |
- Filters: `"content": "trim"`, `"translation": "trim"`
- Messages: `"content.max_len": "内容不能超过2000个字符"`, `"translation.max_len": "翻译不能超过2000个字符"`

### `ReorderContentItemRequest` (already has FormRequest methods)
| Field (json tag) | Current | Proposed |
|-------|---------|----------|
| itemId | `required` | `required\|uuid` |
| newOrder | (none) | `required\|min:0` |
- Messages: `"itemId.required": "请指定内容项"`, `"newOrder.required": "请指定排序位置"`, `"newOrder.min": "排序位置不能为负数"`

## Section 8: Admin Requests (`adm/`)

### `CreateNoticeRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| title | `required\|max_len:200` | `required\|min_len:2\|max_len:200` |
| content | (none) | `required\|max_len:5000` |
| icon | (none) | `max_len:50` |
- Filters: `"title": "trim"`, `"content": "trim"`, `"icon": "trim"`
- Messages: English — `"title.required": "Title is required"`, `"title.min_len": "Title must be at least 2 characters"`, `"title.max_len": "Title must not exceed 200 characters"`, `"content.required": "Content is required"`, `"content.max_len": "Content must not exceed 5000 characters"`, `"icon.max_len": "Icon must not exceed 50 characters"`

### `UpdateNoticeRequest`
Same rules and messages as `CreateNoticeRequest`.

### `GenerateCodesRequest`
| Field | Current | Proposed |
|-------|---------|----------|
| grade | `required\|in:month,season,year,lifetime` | `required\|in:month,season,year,lifetime` (unchanged) |
| count | `required\|in:10,50,100,500` | `required\|in:10,50,100,500` (unchanged) |
- Messages: English — `"grade.required": "Grade is required"`, `"grade.in": "Grade must be one of: month, season, year, lifetime"`, `"count.required": "Count is required"`, `"count.in": "Count must be one of: 10, 50, 100, 500"`
- Note: Count remains string type (Goravel float64 workaround). Hardcoded `in:` values used here because the valid grades for redeem generation are a subset of all grades (excludes "free").

## Summary

| Metric | Before | After |
|--------|--------|-------|
| Structs with Authorize() + Rules() | ~27/44 | 44/44 (100%) |
| Structs with Messages() | ~9/44 | 44/44 (100%) |
| Structs with Filters() | 0/44 | ~22/44 (50%) |
| Distinct rules used | 5 | ~15 |
| UUID-validated ID fields | 0 | ~25 |
| Enum-validated fields (via InEnum) | 0 | ~12 |
| Numeric range checks | 0 | ~20 |
| Custom rules | 0 | 1 (strong_password) |

## Files to Create

- `app/rules/strong_password.go` — custom password complexity rule
- `app/helpers/enum_rules.go` — `InEnum()` helper function
- `bootstrap/rules.go` — registers strong_password rule

## Files to Modify

### Request files
- `app/http/requests/api/auth_request.go`
- `app/http/requests/api/user_request.go`
- `app/http/requests/api/game_request.go`
- `app/http/requests/api/session_request.go`
- `app/http/requests/api/game_report_request.go`
- `app/http/requests/api/feedback_request.go`
- `app/http/requests/api/upload_request.go`
- `app/http/requests/api/user_favorite_request.go`
- `app/http/requests/api/user_redeem_request.go`
- `app/http/requests/api/user_master_request.go`
- `app/http/requests/api/user_review_request.go`
- `app/http/requests/api/user_unknown_request.go`
- `app/http/requests/api/course_game_request.go`
- `app/http/requests/api/content_seek_request.go`
- `app/http/requests/adm/notice_request.go`
- `app/http/requests/adm/redeem_request.go`

### Bootstrap
- `bootstrap/app.go` — add `WithRules()` call

### Controllers (FormRequest conversion only)
- `app/http/controllers/api/game_controller.go` — ListGames, SearchGames, LevelContent: switch from manual query parsing to `helpers.Validate()`
- `app/http/controllers/api/game_session_controller.go` — CheckActiveSession, CheckActiveLevelSession, RestoreSession: switch from manual query parsing to `helpers.Validate()`. CompleteLevelRequest, AdvanceLevelRequest: add `helpers.Validate()` for body fields (keep route param reads as-is). Remove RestartLevelRequest usage (dead code).

## Files NOT Modified

- `app/helpers/response.go` — `Validate()` helper unchanged
- `app/consts/*` — read-only, enum values consumed by `InEnum()` helper
