# Controller-Model Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split mixed controllers/requests/services so each file corresponds to one model by name.

**Architecture:** Pure file reorganization — extract methods from grab-bag files into properly named files, delete duplicates, rewire routes. Zero business logic changes.

**Tech Stack:** Go/Goravel, Next.js (one frontend fix)

**Spec:** `docs/superpowers/specs/2026-03-23-controller-model-alignment-design.md`

**Build note:** Tasks 2-7 create new files and delete old ones. The build will be temporarily broken until Task 8 rewires routes to the new controllers. This is expected for a pure refactoring — the build is fully green after Task 8.

---

## Task 1: Delete Dead Code

**Files:**
- Delete: `dx-api/app/http/controllers/user_controller.go`

**Note:** `api/admin_community_controller.go` (the duplicate) is deleted in Task 8 alongside the route rewiring, so the build stays green after every commit.

- [ ] **Step 1: Delete root user_controller.go (dead "Hello Goravel" scaffold)**

```bash
rm dx-api/app/http/controllers/user_controller.go
```

- [ ] **Step 2: Verify no references to root controllers package**

```bash
cd dx-api && grep -r "controllers\"" routes/ app/ --include="*.go" | grep -v "api\|adm"
```

Expected: no output (it was unused).

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: clean build.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/user_controller.go
git commit -m "chore: delete dead root user_controller (Hello Goravel scaffold)"
```

---

## Task 2: Split api/community_controller.go into 8 Controllers

**Files:**
- Create: `dx-api/app/http/controllers/api/notice_controller.go`
- Create: `dx-api/app/http/controllers/api/feedback_controller.go`
- Create: `dx-api/app/http/controllers/api/game_report_controller.go`
- Create: `dx-api/app/http/controllers/api/user_redeem_controller.go`
- Create: `dx-api/app/http/controllers/api/content_seek_controller.go`
- Create: `dx-api/app/http/controllers/api/user_referral_controller.go`
- Create: `dx-api/app/http/controllers/api/leaderboard_controller.go`
- Create: `dx-api/app/http/controllers/api/hall_controller.go`
- Delete: `dx-api/app/http/controllers/api/community_controller.go`

**Source mapping** (read `community_controller.go` for exact code):

| Lines | Methods | Target file |
|-------|---------|-------------|
| 136-160 | GetNotices, MarkNoticesRead | notice_controller.go |
| 162-187 | SubmitFeedback | feedback_controller.go |
| 189-214 | SubmitReport | game_report_controller.go |
| 216-262 | GetRedeems, RedeemCode | user_redeem_controller.go |
| 264-307 | GetContentSeeks, SubmitContentSeek | content_seek_controller.go |
| 96-134 | GetInviteData, GetReferrals | user_referral_controller.go |
| 26-45 | GetLeaderboard | leaderboard_controller.go |
| 47-94 | GetDashboard, GetHeatmap | hall_controller.go |
| 309-312 | currentYear() | hall_controller.go (private helper) |

- [ ] **Step 1: Create notice_controller.go**

Extract `GetNotices` (lines 136-146) and `MarkNoticesRead` (lines 148-160) from `community_controller.go`. Create struct `NoticeController` with constructor `NewNoticeController()`. Import only what these methods need (`services "dx-api/app/services/api"`, helpers, consts, facades).

- [ ] **Step 2: Create feedback_controller.go**

Extract `SubmitFeedback` (lines 162-187). Create struct `FeedbackController` with constructor. Needs `requests "dx-api/app/http/requests/api"` import for `SubmitFeedbackRequest`.

- [ ] **Step 3: Create game_report_controller.go**

Extract `SubmitReport` (lines 189-214). Create struct `GameReportController` with constructor. Needs `requests` import for `SubmitReportRequest`.

- [ ] **Step 4: Create user_redeem_controller.go**

Extract `GetRedeems` (lines 216-231) and `RedeemCode` (lines 233-262). Create struct `UserRedeemController` with constructor. Needs `requests` import for `RedeemCodeRequest`.

- [ ] **Step 5: Create content_seek_controller.go**

Extract `GetContentSeeks` (lines 264-277) and `SubmitContentSeek` (lines 279-307). Create struct `ContentSeekController` with constructor. Needs `requests` import for `SubmitContentSeekRequest`.

- [ ] **Step 6: Create user_referral_controller.go**

Extract `GetInviteData` (lines 96-112) and `GetReferrals` (lines 114-134). Create struct `UserReferralController` with constructor.

- [ ] **Step 7: Create leaderboard_controller.go**

Extract `GetLeaderboard` (lines 26-45). Create struct `LeaderboardController` with constructor. Needs `facades` import.

- [ ] **Step 8: Create hall_controller.go**

Extract `GetDashboard` (lines 47-63), `GetHeatmap` (lines 65-94), and `currentYear()` (lines 309-312). Create struct `HallController` with constructor. Needs `strconv`, `time`, `facades` imports.

- [ ] **Step 9: Delete community_controller.go**

```bash
rm dx-api/app/http/controllers/api/community_controller.go
```

- [ ] **Step 10: Commit**

```bash
git add dx-api/app/http/controllers/api/
git commit -m "refactor: split api/community_controller into 8 model-aligned controllers"
```

---

## Task 3: Split api/tracking_controller.go into 4 Controllers

**Files:**
- Create: `dx-api/app/http/controllers/api/user_master_controller.go`
- Create: `dx-api/app/http/controllers/api/user_unknown_controller.go`
- Create: `dx-api/app/http/controllers/api/user_review_controller.go`
- Create: `dx-api/app/http/controllers/api/user_favorite_controller.go`
- Delete: `dx-api/app/http/controllers/api/tracking_controller.go`

**Source mapping** (read `tracking_controller.go` for exact code):

| Lines | Methods | Target file |
|-------|---------|-------------|
| 26-113 | MarkMastered, ListMastered, MasterStats, DeleteMastered, BulkDeleteMastered | user_master_controller.go |
| 117-205 | MarkUnknown, ListUnknown, UnknownStats, DeleteUnknown, BulkDeleteUnknown | user_unknown_controller.go |
| 209-297 | MarkReview, ListReviews, ReviewStats, DeleteReview, BulkDeleteReviews | user_review_controller.go |
| 301-337 | ToggleFavorite, ListFavorites | user_favorite_controller.go |

- [ ] **Step 1: Create user_master_controller.go**

Extract mastered methods (lines 26-113). Create struct `UserMasterController` with constructor. Import `requests "dx-api/app/http/requests/api"` for `MarkTrackingRequest` and `BulkDeleteRequest`.

- [ ] **Step 2: Create user_unknown_controller.go**

Extract unknown methods (lines 117-205). Create struct `UserUnknownController` with constructor. Same request imports.

- [ ] **Step 3: Create user_review_controller.go**

Extract review methods (lines 209-297). Create struct `UserReviewController` with constructor. Same request imports.

- [ ] **Step 4: Create user_favorite_controller.go**

Extract ToggleFavorite + ListFavorites (lines 301-337). Create struct `UserFavoriteController` with constructor. Import `requests` for `ToggleFavoriteRequest`.

Also add `Favorited` method from `game_controller.go` (lines 137-155) — read that file and copy the method. This method calls `services.IsGameFavorited(userID, gameID)`.

- [ ] **Step 5: Delete tracking_controller.go**

```bash
rm dx-api/app/http/controllers/api/tracking_controller.go
```

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/http/controllers/api/
git commit -m "refactor: split api/tracking_controller into 4 model-aligned controllers"
```

---

## Task 4: Split api/game_controller.go (3 New + Trim Existing)

**Files:**
- Create: `dx-api/app/http/controllers/api/game_category_controller.go`
- Create: `dx-api/app/http/controllers/api/game_press_controller.go`
- Create: `dx-api/app/http/controllers/api/game_stats_controller.go`
- Modify: `dx-api/app/http/controllers/api/game_controller.go`

**Implement together with Task 3** (Favorited method crosses both tasks).

- [ ] **Step 1: Create game_category_controller.go**

Extract `Categories` (lines 157-165 of game_controller.go). Create struct `GameCategoryController` with constructor.

- [ ] **Step 2: Create game_press_controller.go**

Extract `Presses` (lines 167-175 of game_controller.go). Create struct `GamePressController` with constructor.

- [ ] **Step 3: Create game_stats_controller.go**

Extract `Stats` (lines 117-135 of game_controller.go). Create struct `GameStatsController` with constructor. Needs `facades` import.

- [ ] **Step 4: Trim game_controller.go**

Remove these methods from `game_controller.go`:
- `Categories` (lines 157-165)
- `Presses` (lines 167-175)
- `Stats` (lines 117-135)
- `Favorited` (lines 137-155) — already moved to user_favorite_controller in Task 3
- `ActiveSession` (lines 97-115) — duplicate of GameSessionController.CheckAnyActive

After trimming, `game_controller.go` should contain only: struct definition, `List`, `Search`, `Played`, `Detail`. Remove unused imports.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/controllers/api/
git commit -m "refactor: split game_controller — extract category, press, stats controllers"
```

---

## Task 5: Split adm/community_controller.go into 2 Controllers

**Files:**
- Create: `dx-api/app/http/controllers/adm/notice_controller.go`
- Create: `dx-api/app/http/controllers/adm/redeem_controller.go`
- Delete: `dx-api/app/http/controllers/adm/community_controller.go`

- [ ] **Step 1: Create adm/notice_controller.go**

Extract `CreateNotice` (lines 22-38), `UpdateNotice` (lines 40-65), `DeleteNotice` (lines 67-82) from `adm/community_controller.go`. Create struct `NoticeController` with constructor `NewNoticeController()`. Keep same imports.

- [ ] **Step 2: Create adm/redeem_controller.go**

Extract `GenerateCodes` (lines 84-112) and `GetAllRedeems` (lines 114-124). Create struct `RedeemController` with constructor `NewRedeemController()`.

- [ ] **Step 3: Delete adm/community_controller.go**

```bash
rm dx-api/app/http/controllers/adm/community_controller.go
```

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/adm/
git commit -m "refactor: split adm/community_controller into notice and redeem controllers"
```

---

## Task 6: Split Request Files

**Files:**
- Create: `dx-api/app/http/requests/api/feedback_request.go`
- Create: `dx-api/app/http/requests/api/game_report_request.go`
- Create: `dx-api/app/http/requests/api/user_redeem_request.go`
- Create: `dx-api/app/http/requests/api/content_seek_request.go`
- Create: `dx-api/app/http/requests/api/user_favorite_request.go`
- Delete: `dx-api/app/http/requests/api/community_request.go`
- Modify: `dx-api/app/http/requests/api/tracking_request.go`
- Create: `dx-api/app/http/requests/adm/notice_request.go`
- Create: `dx-api/app/http/requests/adm/redeem_request.go`
- Delete: `dx-api/app/http/requests/adm/community_request.go`

- [ ] **Step 1: Split api/community_request.go into 4 files**

Read `community_request.go` and create:
- `feedback_request.go` — `SubmitFeedbackRequest` struct (lines 4-7)
- `game_report_request.go` — `SubmitReportRequest` struct (lines 9-16)
- `user_redeem_request.go` — `RedeemCodeRequest` struct (lines 18-20)
- `content_seek_request.go` — `SubmitContentSeekRequest` struct (lines 22-28)

Each file: `package api` header, the struct, nothing else.

```bash
rm dx-api/app/http/requests/api/community_request.go
```

- [ ] **Step 2: Split api/tracking_request.go**

Read `tracking_request.go`. Remove `ToggleFavoriteRequest` (lines 14-16) and create `user_favorite_request.go` with that struct. Keep `MarkTrackingRequest` and `BulkDeleteRequest` in `tracking_request.go` (shared by mastered/unknown/review).

- [ ] **Step 3: Split adm/community_request.go into 2 files**

Read `adm/community_request.go` and create:
- `notice_request.go` — `CreateNoticeRequest` + `UpdateNoticeRequest` (lines 3-15)
- `redeem_request.go` — `GenerateCodesRequest` (lines 17-20)

```bash
rm dx-api/app/http/requests/adm/community_request.go
```

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/requests/
git commit -m "refactor: split request files to match model-aligned controllers"
```

---

## Task 7: Split Service Files

**Files:**
- Create: `dx-api/app/services/api/game_category_service.go`
- Create: `dx-api/app/services/api/game_press_service.go`
- Create: `dx-api/app/services/api/game_stats_service.go`
- Modify: `dx-api/app/services/api/game_service.go`
- Modify: `dx-api/app/services/api/favorite_service.go`
- Create: `dx-api/app/services/api/user_master_service.go`
- Create: `dx-api/app/services/api/user_unknown_service.go`
- Create: `dx-api/app/services/api/user_review_service.go`
- Create: `dx-api/app/services/api/tracking_helpers.go`
- Delete: `dx-api/app/services/api/tracking_service.go`

- [ ] **Step 1: Create game_category_service.go**

Move from `game_service.go`: `ListCategories` function (lines 471-519) + `CategoryData` DTO (lines 63-69). Add required imports (`fmt`, models, facades).

- [ ] **Step 2: Create game_press_service.go**

Move from `game_service.go`: `ListPresses` function (lines 521-539) + `PressData` DTO (lines 71-75). Add required imports.

- [ ] **Step 3: Create game_stats_service.go**

Move from `game_service.go`: `GetGameStats` function (lines 440-458) + `GameStatsData` DTO (lines 429-438). Add required imports. This is the read-only query — separate from the session-lifecycle stats in `stats_service.go`.

- [ ] **Step 4: Move IsGameFavorited to favorite_service.go**

Move `IsGameFavorited` function (lines 460-469 of game_service.go) to the existing `favorite_service.go`. It already has the right imports.

- [ ] **Step 5: Trim game_service.go**

Remove all moved functions + DTOs: `ListCategories`, `ListPresses`, `GetGameStats`, `IsGameFavorited`, `CategoryData`, `PressData`, `GameStatsData`. Clean up unused imports. Remaining: `ListPublishedGames`, `SearchGames`, `GetPlayedGames`, `GetGameDetail` + their DTOs.

- [ ] **Step 6: Create tracking_helpers.go**

Move from `tracking_service.go`: shared DTOs (`TrackingItemData`, `TrackingContentData`, lines 27-40), helper functions (`enrichTrackingItems` lines 386-444, `batchLoadContentItems` lines 446-462, `batchLoadGameNames` lines 464-476), and rate limit constants (lines 16-21).

- [ ] **Step 7: Create user_master_service.go**

Move from `tracking_service.go`: `MasterStatsData` DTO (lines 55-59), `MarkAsMastered` (lines 78-107), `ListMastered` (lines 109-136), `GetMasterStats` (lines 138-151), `DeleteMastered` (lines 153-159), `BulkDeleteMastered` (lines 161-170).

- [ ] **Step 8: Create user_unknown_service.go**

Move from `tracking_service.go`: `UnknownStatsData` DTO (lines 61-66), `MarkAsUnknown` (lines 174-196), `ListUnknown` (lines 198-225), `GetUnknownStats` (lines 227-239), `DeleteUnknown` (lines 241-247), `BulkDeleteUnknown` (lines 249-258).

- [ ] **Step 9: Create user_review_service.go**

Move from `tracking_service.go`: `ReviewStatsData` DTO (lines 68-73), `ReviewItemData` DTO (lines 42-53), `MarkAsReview` (lines 262-287), `ListReviews` (lines 289-350), `GetReviewStats` (lines 352-363), `DeleteReview` (lines 365-371), `BulkDeleteReviews` (lines 373-382).

**Import note:** This file needs `github.com/goravel/framework/support/carbon` (used in `MarkAsReview`) and `dx-api/app/consts` (for `consts.GetNextReviewAt`). Also uses `newID()` which is defined in `session_service.go` — accessible because all files are in the same `api` package.

- [ ] **Step 10: Delete tracking_service.go**

```bash
rm dx-api/app/services/api/tracking_service.go
```

- [ ] **Step 11: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: build may fail due to route references to deleted controller types — this is expected and will be fixed in Task 8.

- [ ] **Step 12: Commit**

```bash
git add dx-api/app/services/api/
git commit -m "refactor: split service files to match model-aligned controllers"
```

---

## Task 8: Delete Duplicate Controller + Update Routes

**Files:**
- Delete: `dx-api/app/http/controllers/api/admin_community_controller.go`
- Modify: `dx-api/routes/api.go`
- Modify: `dx-api/routes/adm.go`

**Note:** The duplicate `admin_community_controller.go` is deleted here (not in Task 1) so the build stays green — the route rewiring happens in the same commit.

- [ ] **Step 1: Delete api/admin_community_controller.go**

```bash
rm dx-api/app/http/controllers/api/admin_community_controller.go
```

- [ ] **Step 2: Update routes/api.go imports**

Add a new import for adm controllers (this file currently only imports `apicontrollers`):
```go
admcontrollers "dx-api/app/http/controllers/adm"
```

- [ ] **Step 3: Replace community controller references in api.go**

Replace the single `communityController := apicontrollers.NewCommunityController()` and all its method references with 8 individual controllers:

```go
noticeController := apicontrollers.NewNoticeController()
feedbackController := apicontrollers.NewFeedbackController()
gameReportController := apicontrollers.NewGameReportController()
userRedeemController := apicontrollers.NewUserRedeemController()
contentSeekController := apicontrollers.NewContentSeekController()
userReferralController := apicontrollers.NewUserReferralController()
leaderboardController := apicontrollers.NewLeaderboardController()
hallController := apicontrollers.NewHallController()
```

Update route registrations:
- `communityController.GetLeaderboard` → `leaderboardController.GetLeaderboard`
- `communityController.GetDashboard` → `hallController.GetDashboard`
- `communityController.GetHeatmap` → `hallController.GetHeatmap`
- `communityController.GetInviteData` → `userReferralController.GetInviteData`
- `communityController.GetReferrals` → `userReferralController.GetReferrals`
- `communityController.GetNotices` → `noticeController.GetNotices`
- `communityController.MarkNoticesRead` → `noticeController.MarkNoticesRead`
- `communityController.SubmitFeedback` → `feedbackController.SubmitFeedback`
- `communityController.SubmitReport` → `gameReportController.SubmitReport`
- `communityController.GetRedeems` → `userRedeemController.GetRedeems`
- `communityController.RedeemCode` → `userRedeemController.RedeemCode`
- `communityController.GetContentSeeks` → `contentSeekController.GetContentSeeks`
- `communityController.SubmitContentSeek` → `contentSeekController.SubmitContentSeek`

- [ ] **Step 4: Replace tracking controller references in api.go**

Replace the single `trackingController := apicontrollers.NewTrackingController()` with 4 controllers:

```go
userMasterController := apicontrollers.NewUserMasterController()
userUnknownController := apicontrollers.NewUserUnknownController()
userReviewController := apicontrollers.NewUserReviewController()
userFavoriteController := apicontrollers.NewUserFavoriteController()
```

Update route registrations for mastered/unknown/review tracking groups. Update favorites:
- `trackingController.ToggleFavorite` → `userFavoriteController.ToggleFavorite`
- `trackingController.ListFavorites` → `userFavoriteController.ListFavorites`

- [ ] **Step 5: Split game controller references in api.go**

Create new controller instances:
```go
gameCategoryController := apicontrollers.NewGameCategoryController()
gamePressController := apicontrollers.NewGamePressController()
gameStatsController := apicontrollers.NewGameStatsController()
```

Update routes:
- `gameController.Categories` → `gameCategoryController.Categories`
- `gameController.Presses` → `gamePressController.Presses`
- `gameController.Stats` → `gameStatsController.Stats`
- `gameController.Favorited` → `userFavoriteController.Favorited`
- Remove the `/games/{id}/active-session` route entirely

- [ ] **Step 6: Replace admin community controller in api.go**

Replace:
```go
admCommunityController := apicontrollers.NewAdminCommunityController()
```

With:
```go
admNoticeController := admcontrollers.NewNoticeController()
admRedeemController := admcontrollers.NewRedeemController()
```

Update route registrations:
- `admCommunityController.CreateNotice` → `admNoticeController.CreateNotice`
- `admCommunityController.UpdateNotice` → `admNoticeController.UpdateNotice`
- `admCommunityController.DeleteNotice` → `admNoticeController.DeleteNotice`
- `admCommunityController.GenerateCodes` → `admRedeemController.GenerateCodes`
- `admCommunityController.GetAllRedeems` → `admRedeemController.GetAllRedeems`

- [ ] **Step 7: Update routes/adm.go**

Replace:
```go
communityController := adm.NewCommunityController()
```

With:
```go
noticeController := adm.NewNoticeController()
redeemController := adm.NewRedeemController()
```

Update route registrations:
- `communityController.CreateNotice` → `noticeController.CreateNotice`
- `communityController.UpdateNotice` → `noticeController.UpdateNotice`
- `communityController.DeleteNotice` → `noticeController.DeleteNotice`
- `communityController.GenerateCodes` → `redeemController.GenerateCodes`
- `communityController.GetAllRedeems` → `redeemController.GetAllRedeems`

- [ ] **Step 8: Verify full build**

```bash
cd dx-api && go build ./...
cd dx-api && go vet ./...
```

Expected: clean build, no warnings.

- [ ] **Step 9: Commit**

```bash
git add dx-api/app/http/controllers/api/admin_community_controller.go dx-api/routes/
git commit -m "refactor: delete duplicate admin controller, rewire all routes to model-aligned controllers"
```

---

## Task 9: Frontend Fix

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`

- [ ] **Step 1: Read the file and find the active-session call**

Read `dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`. Find the `Promise.all` around line 98-106 that calls `/api/games/${mapped.id}/active-session`.

- [ ] **Step 2: Replace with existing sessionApi helper**

Replace:
```typescript
apiClient.get<any>(`/api/games/${mapped.id}/active-session`)
```

With:
```typescript
sessionApi.checkAnyActive(mapped.id)
```

Update the import — currently only `apiClient` is imported:
```typescript
// Before:
import { apiClient } from "@/lib/api-client";
// After:
import { apiClient, sessionApi } from "@/lib/api-client";
```

- [ ] **Step 3: Verify frontend build**

```bash
cd dx-web && npm run build
```

Expected: clean build.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/\(main\)/games/\[id\]/page.tsx
git commit -m "refactor: replace removed /games/:id/active-session with sessionApi.checkAnyActive"
```

---

## Task 10: Final Verification

- [ ] **Step 1: Run full Go build**

```bash
cd dx-api && go build ./...
```

- [ ] **Step 2: Run Go vet**

```bash
cd dx-api && go vet ./...
```

- [ ] **Step 3: Run tests**

```bash
cd dx-api && go test -race ./...
```

- [ ] **Step 4: Run frontend build**

```bash
cd dx-web && npm run build
```

- [ ] **Step 5: Verify file counts match spec**

```bash
# Should have these new controller files in api/
ls dx-api/app/http/controllers/api/ | wc -l
# Expected: ~22 files (was ~10, added 15, deleted 2)

# Should have no community_controller or tracking_controller or admin_community_controller
ls dx-api/app/http/controllers/api/ | grep -E "community|tracking|admin_community"
# Expected: no output

# adm should have notice + redeem + auth (no community)
ls dx-api/app/http/controllers/adm/
# Expected: auth_controller.go notice_controller.go redeem_controller.go

# Root controllers dir should be empty or gone
ls dx-api/app/http/controllers/*.go 2>/dev/null
# Expected: no output
```

- [ ] **Step 6: Final commit (if any cleanup needed)**

```bash
git status
```
