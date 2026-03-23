# Controller-Model Alignment Refactoring

**Date:** 2026-03-23
**Status:** Draft
**Scope:** dx-api controllers, requests, services, routes + one dx-web frontend fix

## Problem

Controllers in `dx-api/app/http/controllers/` mix multiple models in single files, making them hard to maintain:

- `api/community_controller.go` handles 9 unrelated models (notice, feedback, game_report, user_redeem, content_seek, user_referral, leaderboard, dashboard, heatmap)
- `api/tracking_controller.go` handles 4 models (user_master, user_unknown, user_review, user_favorite)
- `api/game_controller.go` handles 5 models (game, game_category, game_press, game_stats_total, user_favorite)
- `api/admin_community_controller.go` is a byte-for-byte duplicate of `adm/community_controller.go`
- `controllers/user_controller.go` is dead "Hello Goravel" placeholder code
- Request files follow the same mixed pattern
- Two service files also mix models

## Principle

**Each controller corresponds to one model with the same name.** A controller may read related models for enrichment (e.g., game_controller loading category names), but the primary model determines the file name.

Admin controllers live exclusively in `adm/`, never in `api/`.

The same naming pattern applies to request and service files.

## Changes

### Phase 1: Delete Dead/Duplicate Code (3 files)

| File | Reason |
|------|--------|
| `controllers/user_controller.go` | Dead "Hello Goravel" scaffold |
| `api/admin_community_controller.go` | Exact duplicate of `adm/community_controller.go` |
| `api/community_controller.go` | Replaced by 8 new files |

### Phase 2: Split api/community_controller.go (8 new controllers)

| New file | Struct | Methods | Primary model |
|----------|--------|---------|---------------|
| `api/notice_controller.go` | NoticeController | GetNotices, MarkNoticesRead | notice |
| `api/feedback_controller.go` | FeedbackController | SubmitFeedback | feedback |
| `api/game_report_controller.go` | GameReportController | SubmitReport | game_report |
| `api/user_redeem_controller.go` | UserRedeemController | GetRedeems, RedeemCode | user_redeem |
| `api/content_seek_controller.go` | ContentSeekController | GetContentSeeks, SubmitContentSeek | content_seek |
| `api/user_referral_controller.go` | UserReferralController | GetInviteData, GetReferrals | user_referral |
| `api/leaderboard_controller.go` | LeaderboardController | GetLeaderboard | game_stats_total (aggregated) |
| `api/hall_controller.go` | HallController | GetDashboard, GetHeatmap, `currentYear()` helper | user (aggregated) |

### Phase 3: Split api/tracking_controller.go (4 new controllers)

| New file | Struct | Methods | Primary model |
|----------|--------|---------|---------------|
| `api/user_master_controller.go` | UserMasterController | MarkMastered, ListMastered, MasterStats, DeleteMastered, BulkDeleteMastered | user_master |
| `api/user_unknown_controller.go` | UserUnknownController | MarkUnknown, ListUnknown, UnknownStats, DeleteUnknown, BulkDeleteUnknown | user_unknown |
| `api/user_review_controller.go` | UserReviewController | MarkReview, ListReviews, ReviewStats, DeleteReview, BulkDeleteReviews | user_review |
| `api/user_favorite_controller.go` | UserFavoriteController | ToggleFavorite, ListFavorites, Favorited (from GameController) | user_favorite |

Delete `api/tracking_controller.go` after split.

**Note:** Phases 3 and 4 should be implemented together since `Favorited` moves from `game_controller.go` (Phase 4) to `user_favorite_controller.go` (Phase 3).

### Phase 4: Split api/game_controller.go (3 new controllers + trim existing)

| New file | Struct | Methods | Primary model |
|----------|--------|---------|---------------|
| `api/game_category_controller.go` | GameCategoryController | Categories | game_category |
| `api/game_press_controller.go` | GamePressController | Presses | game_press |
| `api/game_stats_controller.go` | GameStatsController | Stats | game_stats_total |

**Remaining in game_controller.go:** List, Search, Detail, Played (all primary game model operations).

**Remove from GameController:**
- **ActiveSession** — duplicate of `GameSessionController.CheckAnyActive()`, both calling `services.CheckAnyActiveSession()`
- **Favorited** — moved to `UserFavoriteController` in Phase 3
- **Categories** — moved to `GameCategoryController` above
- **Presses** — moved to `GamePressController` above
- **Stats** — moved to `GameStatsController` above

### Phase 5: Split adm/community_controller.go (2 new controllers)

| New file | Struct | Methods | Primary model |
|----------|--------|---------|---------------|
| `adm/notice_controller.go` | NoticeController | CreateNotice, UpdateNotice, DeleteNotice | notice |
| `adm/redeem_controller.go` | RedeemController | GenerateCodes, GetAllRedeems | user_redeem |

Delete `adm/community_controller.go` after split.

### Phase 6: Split Request Files

**Delete `api/community_request.go`**, create:

| New file | Structs |
|----------|---------|
| `api/feedback_request.go` | SubmitFeedbackRequest |
| `api/game_report_request.go` | SubmitReportRequest |
| `api/user_redeem_request.go` | RedeemCodeRequest |
| `api/content_seek_request.go` | SubmitContentSeekRequest |

**Modify `api/tracking_request.go`** — remove ToggleFavoriteRequest (keep shared MarkTrackingRequest + BulkDeleteRequest used by mastered/unknown/review). Create:

| New file | Structs |
|----------|---------|
| `api/user_favorite_request.go` | ToggleFavoriteRequest |

**Delete `adm/community_request.go`**, create:

| New file | Structs |
|----------|---------|
| `adm/notice_request.go` | CreateNoticeRequest, UpdateNoticeRequest |
| `adm/redeem_request.go` | GenerateCodesRequest |

### Phase 7: Split Service Files

**Split `api/game_service.go`** — move functions out:

| Function | Destination |
|----------|------------|
| ListCategories, CategoryData DTO | new `api/game_category_service.go` |
| ListPresses, PressData DTO | new `api/game_press_service.go` |
| GetGameStats, GameStatsData DTO | new `api/game_stats_service.go` (read-only query, separate from the session-lifecycle stats in `stats_service.go`) |
| IsGameFavorited | existing `api/favorite_service.go` |

Remaining in `game_service.go`: ListPublishedGames, SearchGames, GetPlayedGames, GetGameDetail + their DTOs.

**Split `api/tracking_service.go`** into:

| New file | Functions |
|----------|-----------|
| `api/user_master_service.go` | MarkAsMastered, ListMastered, GetMasterStats, DeleteMastered, BulkDeleteMastered, MasterStatsData DTO |
| `api/user_unknown_service.go` | MarkAsUnknown, ListUnknown, GetUnknownStats, DeleteUnknown, BulkDeleteUnknown, UnknownStatsData DTO |
| `api/user_review_service.go` | MarkAsReview, ListReviews, GetReviewStats, DeleteReview, BulkDeleteReviews, ReviewStatsData, ReviewItemData DTOs |
| `api/tracking_helpers.go` | enrichTrackingItems, batchLoadContentItems, batchLoadGameNames, TrackingItemData, TrackingContentData DTOs, rate limit constants |

Delete `api/tracking_service.go` after split.

### Phase 8: Update Routes

**`routes/api.go`:**
- Replace single `communityController` with 8 individual controllers
- Replace single `trackingController` with 4 individual controllers
- Split `gameController` references to use new category/press/stats/favorite controllers
- Remove `/games/{id}/active-session` route
- Replace `api.AdminCommunityController` with `adm.NoticeController` and `adm.RedeemController` (add new import for `adm` controllers package alongside existing `apicontrollers` alias)

**`routes/adm.go`:**
- Replace `adm.CommunityController` with `adm.NoticeController` and `adm.RedeemController`

### Phase 9: Frontend Fix (dx-web)

**File:** `dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`

Replace:
```typescript
apiClient.get<any>(`/api/games/${mapped.id}/active-session`)
```

With the existing helper:
```typescript
sessionApi.checkAnyActive(mapped.id)
```

This calls `/api/sessions/any-active?game_id={id}` which uses the same backend service function.

## Files Changed Summary

| Category | Deleted | Created | Modified |
|----------|---------|---------|----------|
| Controllers (root) | 1 | 0 | 0 |
| Controllers (api/) | 2 | 15 | 1 (game_controller.go trimmed) |
| Controllers (adm/) | 1 | 2 | 0 |
| Requests (api/) | 1 | 5 | 1 (tracking_request.go trimmed) |
| Requests (adm/) | 1 | 2 | 0 |
| Services (api/) | 1 | 7 | 2 (game_service, favorite_service) |
| Routes | 0 | 0 | 2 (api.go, adm.go) |
| Frontend | 0 | 0 | 1 (games/[id]/page.tsx) |
| **Total** | **7** | **31** | **7** |

**Note:** `game_request.go` is intentionally unchanged — GameController does inline query parsing, not request struct binding.

## Risks

- **Zero API contract changes** except removing `/games/{id}/active-session` (replaced by existing `/sessions/any-active`)
- Pure file reorganization — no business logic changes
- All functions keep their exact signatures
- Shared DTOs/helpers extracted to `tracking_helpers.go` remain in the same package

## Verification

After refactoring:
```bash
cd dx-api && go build ./...    # Must compile
cd dx-api && go vet ./...      # No warnings
cd dx-web && npm run build     # Frontend must build
```
