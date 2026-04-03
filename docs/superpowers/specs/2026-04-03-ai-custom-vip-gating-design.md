# AI Custom VIP Gating Design

## Overview

Gate AI Custom (AI 随心配) features behind VIP membership. The list page is viewable by all users. All other operations (create, edit, AI generation, content management) require active VIP.

## VIP Status Logic

Same as game level gating: `grade == "lifetime"` OR (`grade != "free"` AND `vipDueAt > now()`).

## Backend Guards

### `dx-api/app/services/api/ai_custom_service.go`

Add `requireVip(userID)` to all 4 functions:
- `GenerateMetadata` (returns `*GenerateMetadataResult, error`)
- `FormatMetadata` (returns `*FormatMetadataResult, error`)
- `BreakMetadata` (void, streams SSE — need early return pattern)
- `GenerateContentItems` (void, streams SSE — need early return pattern)

For SSE functions (`BreakMetadata`, `GenerateContentItems`), since they don't return errors but stream via `SSEWriter`, check VIP at the top and write an SSE error event + return early if not VIP.

### `dx-api/app/services/api/course_game_service.go`

Add `requireVip(userID)` to all functions EXCEPT `ListUserGames` and `GetUserGameCounts`:
- `CreateGame` (returns `string, error`)
- `UpdateGame` (returns `error`)
- `DeleteGame` (returns `error`)
- `PublishGame` (returns `error`)
- `WithdrawGame` (returns `error`)
- `CreateLevel` (returns `string, error`)
- `DeleteLevel` (returns `error`)
- `GetCourseGameDetail` (returns `*CourseGameDetailData, error`)

### `dx-api/app/services/api/course_content_service.go`

Add `requireVip(userID)` to all public functions:
- `SaveMetadataBatch` (returns `int, error`)
- `ReorderMetadata` (returns `error`)
- `GetContentItemsByMeta` (returns `[]LevelContentData, error`)
- `InsertContentItem` (returns `*CourseContentItemData, error`)
- `UpdateContentItemText` (returns `error`)
- `ReorderContentItems` (returns `error`)
- `DeleteContentItem` (returns `error`)
- `DeleteAllLevelContent` (returns `error`)

### Controllers

- `ai_custom_controller.go`: Add `ErrVipRequired` check to `GenerateMetadata` and `FormatMetadata` error handling. For SSE endpoints, the service handles it directly.
- `course_game_controller.go`: Add `ErrVipRequired` case to `mapCourseGameError`.

## Frontend Guards

### `dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx`

- Fetch profile, compute `isVip`
- "创建课程" button → opens `UpgradeDialog` if not VIP
- Game card clicks → opens `UpgradeDialog` instead of navigating if not VIP
- Dialog message: "升级会员即可使用 AI 随心配，创建专属学习课程"

## Files Summary

### Backend

| File | Change |
|------|--------|
| `app/services/api/ai_custom_service.go` | VIP guard on all 4 functions |
| `app/services/api/course_game_service.go` | VIP guard on 9 functions (not ListUserGames, GetUserGameCounts) |
| `app/services/api/course_content_service.go` | VIP guard on all 8 public functions |
| `app/http/controllers/api/ai_custom_controller.go` | ErrVipRequired mapping |
| `app/http/controllers/api/course_game_controller.go` | ErrVipRequired in mapCourseGameError |

### Frontend

| File | Change |
|------|--------|
| `dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx` | VIP gating on create + card navigation |
