# 学习 page progress card + label cleanup — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a paginated "我的学习进度" card to the dx-mini 学习 tab page above the existing stat-card row (sourced from a new `/api/tracking/sessions` endpoint that wraps the existing dashboard query), rename "不认识" → "生词本" (and "不认识词库" → "生词本词库"), switch the page to `navigationStyle: "custom"`, and lift content flush to the system status bar — without changing any existing data flow, route, or other client.

**Architecture:** The new endpoint exports `hall_service.go`'s existing private `getSessionProgress` function and reuses its SQL byte-for-byte; a new thin `UserSessionController` wraps it under the existing `/api/tracking` route group. The dx-mini 学习 page extends its current `Promise.all` of three stat fetches into a `Promise.allSettled` of four, pre-computes display fields (per the project's WXML-can't-call-methods constraint), and renders a self-contained card with a custom prev/next paginator (5 per page, hidden when ≤5 sessions). The card stays inline (single consumer), and `GAME_MODE_LABELS` is inlined in `learn.ts` (4 entries, single consumer).

**Tech Stack:** Goravel (Go) + GORM raw SQL on dx-api side; native WeChat Mini Program with TypeScript + WXML + WXSS + Vant Weapp 1.11 + the project's `<dx-icon>` Lucide SVG component on dx-mini side.

**Reference spec:** `dx-mini/docs/superpowers/specs/2026-05-01-learn-page-progress-card-design.md`

---

## File Structure

**dx-api (3 changed, 1 new):**
- `app/services/api/hall_service.go` — rename one private function to exported.
- `app/http/controllers/api/user_session_controller.go` — *(new)* one controller, one method.
- `routes/api.go` — register one route inside the existing `/tracking` group.

**dx-mini (5 changed, 1 regenerated):**
- `scripts/build-icons.mjs` — append two icon entries.
- `miniprogram/components/dx-icon/icons.ts` — *(regenerated, do not hand-edit)*.
- `miniprogram/pages/learn/learn.json` — flip to custom nav.
- `miniprogram/pages/learn/learn.ts` — add session loading, pagination state, handlers.
- `miniprogram/pages/learn/learn.wxml` — new card markup; rename two labels; new top shell.
- `miniprogram/pages/learn/learn.wxss` — new top padding rule + new card styles.

The plan ships dx-api first (so the endpoint exists before dx-mini calls it), then dx-mini. Each task ends with a commit so the repo never sits in a half-applied state.

---

## Task 1: Export `ListSessionProgress` in `hall_service.go`

**Files:**
- Modify: `dx-api/app/services/api/hall_service.go`

The existing private `getSessionProgress(userID string) ([]SessionProgress, error)` becomes exported as `ListSessionProgress`. The body is unchanged. The single internal caller inside `GetDashboard` is updated.

- [ ] **Step 1: Rename the function declaration**

In `dx-api/app/services/api/hall_service.go` (around line 207), change:

```go
// getSessionProgress returns recent game session progress entries.
func getSessionProgress(userID string) ([]SessionProgress, error) {
```

to:

```go
// ListSessionProgress returns recent game session progress entries (up to 20, ordered by last_played_at DESC).
func ListSessionProgress(userID string) ([]SessionProgress, error) {
```

The function body (the `SELECT ... FROM game_sessions s ...` raw SQL block, the row mapping loop, and the `return results, nil`) stays exactly the same.

- [ ] **Step 2: Update the internal call site**

In `dx-api/app/services/api/hall_service.go` (around line 124), change:

```go
	sessions, err := getSessionProgress(userID)
```

to:

```go
	sessions, err := ListSessionProgress(userID)
```

- [ ] **Step 3: Verify clean compile**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Verify clean static analysis**

Run: `cd dx-api && go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 5: Verify no regressions**

Run: `cd dx-api && go test -race ./...`
Expected: all existing tests pass (the rename has no behavioral change, but this guards against accidental ripple).

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/hall_service.go
git commit -m "refactor(api): export ListSessionProgress so other controllers can reuse it"
```

---

## Task 2: Create `UserSessionController`

**Files:**
- Create: `dx-api/app/http/controllers/api/user_session_controller.go`

Thin controller: auth check, call `services.ListSessionProgress`, return the success envelope. Mirrors the structure of the existing `HallController.GetDashboard` so the codebase's pattern stays consistent.

- [ ] **Step 1: Write the controller file**

Create `dx-api/app/http/controllers/api/user_session_controller.go` with the following content:

```go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type UserSessionController struct{}

func NewUserSessionController() *UserSessionController {
	return &UserSessionController{}
}

// ListSessions returns the user's recent game session progress (up to 20).
func (c *UserSessionController) ListSessions(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	rows, err := services.ListSessionProgress(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list sessions")
	}
	return helpers.Success(ctx, rows)
}
```

- [ ] **Step 2: Verify clean compile**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Verify clean static analysis**

Run: `cd dx-api && go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/user_session_controller.go
git commit -m "feat(api): add UserSessionController exposing recent game session progress"
```

---

## Task 3: Register the route under `/api/tracking/sessions`

**Files:**
- Modify: `dx-api/routes/api.go`

Add the route inside the existing `protected.Prefix("/tracking").Group(...)` block (between the master/unknown/review subgroups), so it inherits the same JWT auth middleware that the rest of `/tracking` uses.

- [ ] **Step 1: Add the controller construction and route line**

In `dx-api/routes/api.go`, locate the tracking group (around line 148):

```go
			// Tracking routes (mastered / unknown / review)
			userMasterController := apicontrollers.NewUserMasterController()
			userUnknownController := apicontrollers.NewUserUnknownController()
			userReviewController := apicontrollers.NewUserReviewController()
			protected.Prefix("/tracking").Group(func(tracking route.Router) {
```

Replace the comment and the three controller-construction lines with these four lines (adding the new controller construction):

```go
			// Tracking routes (mastered / unknown / review / sessions)
			userMasterController := apicontrollers.NewUserMasterController()
			userUnknownController := apicontrollers.NewUserUnknownController()
			userReviewController := apicontrollers.NewUserReviewController()
			userSessionController := apicontrollers.NewUserSessionController()
			protected.Prefix("/tracking").Group(func(tracking route.Router) {
```

Then, inside the group body, after the Review subgroup ends (after the line `tracking.Delete("/review", userReviewController.BulkDeleteReviews)` and before the closing `})` — around line 169), insert:

```go

				// Sessions
				tracking.Get("/sessions", userSessionController.ListSessions)
```

(Note the leading blank line for visual grouping consistency with the Mastered / Unknown / Review subgroups above.)

- [ ] **Step 2: Verify clean compile**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Verify clean static analysis**

Run: `cd dx-api && go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Smoke-test the route exists**

Run: `cd dx-api && go run . &` then in another shell:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3001/api/tracking/sessions
```

Expected: `401` (unauthorized — confirms the route is registered and the JWT middleware is enforcing auth, which is correct since we haven't sent a token). Then stop the server with `kill %1` (or however the user prefers to stop background processes).

If the user is running `air` for hot reload elsewhere, skip this step and rely on Steps 2-3 instead — the route registration will fail at compile time if it's wrong.

- [ ] **Step 5: Verify no regressions**

Run: `cd dx-api && go test -race ./...`
Expected: all existing tests pass.

- [ ] **Step 6: Commit**

```bash
git add dx-api/routes/api.go
git commit -m "feat(api): expose GET /api/tracking/sessions for the mini learn page"
```

---

## Task 4: Add `list-checks` and `gamepad-2` to the icon inventory

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Regenerate: `dx-mini/miniprogram/components/dx-icon/icons.ts`

The build-icons script's static WXML scan (per the project's `feedback_dx_mini_icon_strategy.md` memory) refuses to build if any `<dx-icon name="X">` reference isn't covered by the inventory — so add the two new entries first, then regenerate.

- [ ] **Step 1: Append two ICONS entries**

In `dx-mini/scripts/build-icons.mjs`, locate the `ICONS` array. Find the last entry just before the closing `]`:

```js
  ['check-circle-2',      'circle-check'],     // lucide-static dropped check-circle-2; circle-check is the closest equivalent
]
```

Insert two new entries right before the closing `]`:

```js
  ['check-circle-2',      'circle-check'],     // lucide-static dropped check-circle-2; circle-check is the closest equivalent
  ['list-checks',         'list-checks'],
  ['gamepad-2',           'gamepad-2'],
]
```

- [ ] **Step 2: Verify the source SVGs exist**

Run: `ls dx-mini/node_modules/lucide-static/icons/list-checks.svg dx-mini/node_modules/lucide-static/icons/gamepad-2.svg`
Expected: both files listed, exit code 0.

If either file is missing, the script will throw a clear `lucide-static is missing "X.svg"` error in the next step. In that case, run `cd dx-mini && npm install` first to ensure `lucide-static` is fully installed.

- [ ] **Step 3: Regenerate `icons.ts`**

Run: `cd dx-mini && npm run build:icons`
Expected: no errors. The script writes `miniprogram/components/dx-icon/icons.ts`. Inspect the file briefly to confirm `list-checks` and `gamepad-2` keys appear in the exported map.

- [ ] **Step 4: Commit (both files together)**

```bash
git add dx-mini/scripts/build-icons.mjs dx-mini/miniprogram/components/dx-icon/icons.ts
git commit -m "feat(mini): add list-checks and gamepad-2 to dx-icon inventory"
```

---

## Task 5: Switch `learn.json` to custom navigation

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.json`

Drop the WeChat default nav title; add `navigationStyle: "custom"` so the page can manage its own top spacing.

- [ ] **Step 1: Replace the file contents**

Overwrite `dx-mini/miniprogram/pages/learn/learn.json` with:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "dx-icon": "/components/dx-icon/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

(`navigationBarTitleText` is removed because the default WeChat nav bar is no longer rendered.)

- [ ] **Step 2: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.json
git commit -m "feat(mini): switch learn page to custom navigation style"
```

---

## Task 6: Rewrite `learn.ts` with sessions loading + pagination

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.ts`

Add session-loading, pre-computed display fields, custom-paginator state, and the new tap handler. Keep the existing stat-loading behavior, but switch from `Promise.all` to `Promise.allSettled` so a single slow/failed branch doesn't blank everything else.

- [ ] **Step 1: Overwrite the file with the new implementation**

Replace `dx-mini/miniprogram/pages/learn/learn.ts` entirely with:

```ts
import { api } from '../../utils/api'

interface Stats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

interface SessionProgress {
  gameId: string
  gameName: string
  gameMode: string
  completedLevels: number
  totalLevels: number
  score: number
  exp: number
  lastPlayedAt: string
}

interface ProgressItem {
  gameId: string
  title: string
  progressPct: number
  barColor: string
  barWidth: string
}

const PAGE_SIZE = 5

const PROGRESS_COLORS = [
  '#14b8a6',
  '#3b82f6',
  '#f59e0b',
  '#ec4899',
  '#8b5cf6',
  '#06b6d4',
]

const GAME_MODE_LABELS: Record<string, string> = {
  'word-sentence': '连词成句',
  'vocab-battle': '词汇对轰',
  'vocab-match': '词汇配对',
  'vocab-elimination': '词汇消消乐',
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    arrowColor: '#9ca3af',
    accentColors: { teal: '#10b981', amber: '#f59e0b', purple: '#6366f1' } as { teal: string; amber: string; purple: string },
    statusBarHeight: 20,
    loading: true,
    masterStats: null as Stats | null,
    unknownStats: null as Stats | null,
    reviewStats: null as ReviewStats | null,
    sessions: [] as SessionProgress[],
    progressItems: [] as ProgressItem[],
    pageItems: [] as ProgressItem[],
    progressPage: 1,
    progressTotalPages: 1,
    hasPagination: false,
    prevDisabled: true,
    nextDisabled: true,
    pageLabel: '第 1 / 1 页',
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({ statusBarHeight: sys.statusBarHeight || 20 })
  },
  onShow() {
    this.setData({
      theme: app.globalData.theme,
      arrowColor: app.globalData.theme === 'dark' ? '#6b7280' : '#9ca3af',
    })
    this.loadAll()
  },
  async loadAll() {
    this.setData({ loading: true })
    const results = await Promise.allSettled([
      api.get<Stats>('/api/tracking/master/stats'),
      api.get<Stats>('/api/tracking/unknown/stats'),
      api.get<ReviewStats>('/api/tracking/review/stats'),
      api.get<SessionProgress[]>('/api/tracking/sessions'),
    ])
    const masterStats = results[0].status === 'fulfilled' ? results[0].value : this.data.masterStats
    const unknownStats = results[1].status === 'fulfilled' ? results[1].value : this.data.unknownStats
    const reviewStats = results[2].status === 'fulfilled' ? results[2].value : this.data.reviewStats
    const sessions = results[3].status === 'fulfilled' ? results[3].value : []

    this.setData({ loading: false, masterStats, unknownStats, reviewStats, sessions })
    this.rebuildProgress(this.data.progressPage)

    if (results.some((r) => r.status === 'rejected')) {
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  rebuildProgress(targetPage: number) {
    const items: ProgressItem[] = this.data.sessions.map((s, i) => {
      const modeText = GAME_MODE_LABELS[s.gameMode] || s.gameMode
      const pct = s.totalLevels === 0 ? 0 : Math.round((s.completedLevels / s.totalLevels) * 100)
      return {
        gameId: s.gameId,
        title: s.gameName + ' · ' + modeText,
        progressPct: pct,
        barColor: PROGRESS_COLORS[i % PROGRESS_COLORS.length],
        barWidth: pct + '%',
      }
    })

    const totalPages = Math.max(1, Math.ceil(items.length / PAGE_SIZE))
    const page = Math.min(Math.max(1, targetPage), totalPages)
    const start = (page - 1) * PAGE_SIZE
    const pageItems = items.slice(start, start + PAGE_SIZE)
    const hasPagination = items.length > PAGE_SIZE

    this.setData({
      progressItems: items,
      pageItems,
      progressPage: page,
      progressTotalPages: totalPages,
      hasPagination,
      prevDisabled: page === 1,
      nextDisabled: page === totalPages,
      pageLabel: '第 ' + page + ' / ' + totalPages + ' 页',
    })
  },
  prevProgressPage() {
    if (this.data.prevDisabled) return
    this.rebuildProgress(this.data.progressPage - 1)
  },
  nextProgressPage() {
    if (this.data.nextDisabled) return
    this.rebuildProgress(this.data.progressPage + 1)
  },
  goGame(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (id) wx.navigateTo({ url: '/pages/games/detail/detail?id=' + id })
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
```

- [ ] **Step 2: Type-check the project**

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json`
Expected: zero NEW errors beyond the pre-existing `Component({methods})` typing pattern that the codebase already tolerates (per `feedback_dx_mini_no_optional_chaining.md`'s sibling note about `miniprogram-api-typings`). If a NEW error appears, fix it before committing — common pitfalls:
- Optional chaining (`?.`) and nullish coalescing (`??`) are NOT allowed (see `feedback_dx_mini_no_optional_chaining.md`); the code above deliberately uses `||` and explicit ternaries instead.
- String concatenation with `+` is used instead of template literals where the existing codebase shows that preference.

- [ ] **Step 3: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.ts
git commit -m "feat(mini): add session loading + pagination state to learn page"
```

---

## Task 7: Rewrite `learn.wxml` with the new card and renamed labels

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.wxml`

Wrap the page in the leaderboard-style shell, insert the new 学习进度 card above the stat-card row, and rename the two display labels (`不认识` → `生词本`, `不认识词库` → `生词本词库`). The handlers `goUnknown` and the navigation target `/pages/learn/unknown/unknown` remain unchanged — only the visible labels shift.

- [ ] **Step 1: Overwrite the file**

Replace `dx-mini/miniprogram/pages/learn/learn.wxml` entirely with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <view wx:if="{{!loading}}" class="content">

      <view class="progress-card">
        <view class="pc-header">
          <view class="pc-icon-wrap">
            <dx-icon name="list-checks" size="20px" color="{{accentColors.teal}}" />
          </view>
          <text class="pc-title">我的学习进度</text>
        </view>

        <view wx:if="{{progressItems.length === 0}}" class="pc-empty">
          <dx-icon name="gamepad-2" size="40px" color="{{arrowColor}}" />
          <text class="pc-empty-title">还没有学习进度</text>
          <text class="pc-empty-sub">去发现课程游戏吧</text>
        </view>

        <block wx:else>
          <view class="pc-list">
            <view
              wx:for="{{pageItems}}"
              wx:key="gameId"
              class="pc-row"
              data-id="{{item.gameId}}"
              bind:tap="goGame"
            >
              <view class="pc-row-top">
                <text class="pc-row-title">{{item.title}}</text>
                <text class="pc-row-pct">{{item.progressPct}}%</text>
              </view>
              <view class="pc-bar-track">
                <view class="pc-bar-fill" style="background:{{item.barColor}};width:{{item.barWidth}}"></view>
              </view>
            </view>
          </view>

          <view wx:if="{{hasPagination}}" class="pc-pager">
            <view class="pc-pager-btn {{prevDisabled ? 'disabled' : ''}}" bind:tap="prevProgressPage">‹ 上一页</view>
            <text class="pc-pager-label">{{pageLabel}}</text>
            <view class="pc-pager-btn {{nextDisabled ? 'disabled' : ''}}" bind:tap="nextProgressPage">下一页 ›</view>
          </view>
        </block>
      </view>

      <view class="stat-card-row">
        <view class="learn-stat-card teal" bind:tap="goMastered">
          <text class="lsc-value">{{masterStats.total || 0}}</text>
          <text class="lsc-label">已掌握</text>
          <text class="lsc-sub">本周 +{{masterStats.thisWeek || 0}}</text>
        </view>
        <view class="learn-stat-card amber" bind:tap="goUnknown">
          <text class="lsc-value">{{unknownStats.total || 0}}</text>
          <text class="lsc-label">生词本</text>
          <text class="lsc-sub">本周 +{{unknownStats.thisWeek || 0}}</text>
        </view>
        <view class="learn-stat-card purple" bind:tap="goReview">
          <text class="lsc-value">{{reviewStats.pending || 0}}</text>
          <text class="lsc-label">待复习</text>
          <text class="lsc-sub">逾期 {{reviewStats.overdue || 0}}</text>
        </view>
      </view>

      <view class="quick-links">
        <view class="quick-link" bind:tap="goMastered">
          <dx-icon name="check" size="20px" color="{{accentColors.teal}}" />
          <text class="ql-text">已掌握词库</text>
          <dx-icon name="chevron-right" size="14px" color="{{arrowColor}}" />
        </view>
        <view class="quick-link" bind:tap="goUnknown">
          <dx-icon name="help-circle" size="20px" color="{{accentColors.amber}}" />
          <text class="ql-text">生词本词库</text>
          <dx-icon name="chevron-right" size="14px" color="{{arrowColor}}" />
        </view>
        <view class="quick-link" bind:tap="goReview">
          <dx-icon name="clock" size="20px" color="{{accentColors.purple}}" />
          <text class="ql-text">待复习队列</text>
          <dx-icon name="chevron-right" size="14px" color="{{arrowColor}}" />
        </view>
      </view>

    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 2: Re-run the icon build to confirm static WXML scan still passes**

Run: `cd dx-mini && npm run build:icons`
Expected: no errors. The script's static scan walks all `*.wxml` files under `miniprogram/` and verifies every `<dx-icon name="X">` reference matches an entry in the `ICONS` inventory; `list-checks` and `gamepad-2` were added in Task 4, so this should be silent. If it complains, return to Task 4 and confirm both entries are present.

- [ ] **Step 3: Type-check (sanity)**

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json`
Expected: same baseline as after Task 6 (no NEW errors).

- [ ] **Step 4: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.wxml
git commit -m "feat(mini): add 学习进度 card + rename 不认识 → 生词本 on learn page"
```

---

## Task 8: Update `learn.wxss` — top padding switch + new card styles

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.wxss`

Replace the `.page-container` rule with the leaderboard-style top-padding so content lifts to the top, change `.content` to keep only horizontal padding, and append the styles for the new progress card. The existing stat-row and quick-link styles stay unchanged.

- [ ] **Step 1: Overwrite the file**

Replace `dx-mini/miniprogram/pages/learn/learn.wxss` entirely with:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
  padding-bottom: 100rpx;
}
.center-loader { display: flex; justify-content: center; padding: 40px; }
.content { padding: 0 16px; }

/* progress card */
.progress-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  padding: 16px;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.pc-header { display: flex; align-items: center; gap: 10px; }
.pc-icon-wrap {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(13, 148, 136, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
.dark .pc-icon-wrap { background: rgba(20, 184, 166, 0.16); }
.pc-title { font-size: 15px; font-weight: 700; color: var(--text-primary); }

.pc-empty {
  min-height: 144px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.pc-empty-title { font-size: 13px; color: var(--text-secondary); }
.pc-empty-sub { font-size: 12px; color: var(--text-secondary); opacity: 0.85; }

.pc-list { display: flex; flex-direction: column; gap: 12px; }
.pc-row { display: flex; flex-direction: column; gap: 6px; padding: 4px 0; }
.pc-row-top { display: flex; align-items: center; justify-content: space-between; }
.pc-row-title {
  font-size: 13px;
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-right: 8px;
}
.pc-row-pct { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.pc-bar-track { height: 6px; border-radius: 999px; background: var(--bg-page); overflow: hidden; }
.dark .pc-bar-track { background: rgba(255, 255, 255, 0.06); }
.pc-bar-fill { height: 100%; border-radius: 999px; }

.pc-pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 4px;
  border-top: 1px solid var(--border-color);
}
.pc-pager-btn {
  font-size: 12px;
  color: var(--primary);
  padding: 8px 10px;
  border-radius: 8px;
}
.pc-pager-btn.disabled { color: var(--text-secondary); opacity: 0.5; }
.pc-pager-label { font-size: 12px; color: var(--text-secondary); }

/* existing stat row */
.stat-card-row { display: flex; gap: 10px; margin-bottom: 20px; }
.learn-stat-card {
  flex: 1;
  border-radius: 12px;
  padding: 14px 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.learn-stat-card.teal { background: rgba(13, 148, 136, 0.12); }
.learn-stat-card.amber { background: rgba(245, 158, 11, 0.12); }
.learn-stat-card.purple { background: rgba(99, 102, 241, 0.12); }
.dark .learn-stat-card.teal { background: rgba(20, 184, 166, 0.12); }
.dark .learn-stat-card.amber { background: rgba(245, 158, 11, 0.10); }
.dark .learn-stat-card.purple { background: rgba(99, 102, 241, 0.10); }
.lsc-value { font-size: 24px; font-weight: 700; color: var(--text-primary); }
.lsc-label { font-size: 12px; color: var(--text-primary); font-weight: 500; }
.lsc-sub { font-size: 11px; color: var(--text-secondary); }

/* existing quick links */
.quick-links { display: flex; flex-direction: column; gap: 8px; }
.quick-link {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.ql-text { flex: 1; font-size: 15px; color: var(--text-primary); }
```

- [ ] **Step 2: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.wxss
git commit -m "style(mini): lift learn page to top + add progress-card + pager styles"
```

---

## Task 9: Manual smoke test in WeChat DevTools

**Files:** none (verification only — no commit unless a regression is found and fixed)

Per the project memory `project_wechat_devtools_realdevice_bug.md`, use **预览 (Preview)** with the WeChat 小程序助手 phone app, NOT 真机调试 (Real Device Debug), which is broken on the current DevTools version.

- [ ] **Step 1: Confirm dx-api is running**

The dx-api server must be reachable from the device that previews the mini program. From the dx-api root, ensure the server is up:

```bash
cd dx-api && go run .
```

(Or use `air` if hot reload is preferred.) Confirm port 3001 is reachable on the LAN IP that the WeChat DevTools `setDevApiBaseUrl(...)` console helper points at.

- [ ] **Step 2: Open the mini program in WeChat DevTools**

Open `dx-mini` in WeChat Developer Tools. Confirm 详情 → 本地设置 → 「不校验合法域名...」 is checked (dev only). In the DevTools console, if the dev API base URL is not yet set, run:

```js
require('./utils/config').setDevApiBaseUrl('http://<lan-ip>')
```

- [ ] **Step 3: Click 预览 and scan with the 小程序助手 phone app**

Navigate to the 学习 tab on the device. Verify:

- [ ] Default WeChat title bar (the "学习" title that used to sit at the top) is gone.
- [ ] The new 学习进度 card sits flush near the top of the page, just below the system status bar / WeChat capsule, with no overlap.
- [ ] The 学习进度 card header shows the `list-checks` icon (filled in teal) and the title 我的学习进度.

- [ ] **Step 4: Verify content states**

- [ ] **Empty state** — for an account that has never played a game, the card shows the `gamepad-2` icon, 还没有学习进度, and 去发现课程游戏吧. No pager.
- [ ] **1–5 sessions** — list renders, no pager visible.
- [ ] **6+ sessions** — pager visible. 上一页 disabled (gray, no tap) on page 1; 下一页 disabled on the last page; tapping the enabled side flips pages and updates the `第 X / Y 页` label.
- [ ] **Tap a row** — navigates to `/pages/games/detail/detail?id=<gameId>`.
- [ ] **Each row** shows `<gameName> · <modeLabel>`, the percent on the right, and a colored progress bar.

- [ ] **Step 5: Verify the rename**

- [ ] The middle stat card under the progress card shows label **生词本** (not 不认识).
- [ ] The middle quick-link shows label **生词本词库** (not 不认识词库). Tapping it still navigates to `/pages/learn/unknown/unknown` (handler unchanged).

- [ ] **Step 6: Verify the dark theme**

Toggle dark mode (e.g. from the 我的 page's 深色模式 cell, or whichever toggle the project uses). Return to the 学习 tab. Verify:

- [ ] Card background, text, bar track, accent fill, and pager are all readable.
- [ ] No white flash, no clipped text, no invisible icons.

- [ ] **Step 7: Verify no regressions on adjacent surfaces**

- [ ] Tap each tab in the bottom tab bar (首页 / 课程 / 排行 / 消息 / 我的) — every tab loads as before.
- [ ] From the 学习 tab, tap each of the three quick links and each of the three stat cards — all navigate to their existing pages (`/pages/learn/mastered/mastered`, `/pages/learn/unknown/unknown`, `/pages/learn/review/review`).
- [ ] Reload the home page and confirm the 我的学习进度 area on home (which uses the same SessionProgress shape via `/api/hall/dashboard`) still renders correctly — the rename of `getSessionProgress` → `ListSessionProgress` should have zero behavioral effect on home.

- [ ] **Step 8: Final cleanup**

If any defect surfaced, fix it inline (returning to the relevant earlier task) and commit the fix as `fix(mini): ...` or `fix(api): ...`. If everything passed, no commit needed for this task.

---

## Self-review notes (already applied)

- **Spec coverage:** Every spec section maps to a task — backend endpoint extraction (Task 1), new controller (Task 2), route registration (Task 3), icon inventory (Task 4), nav style switch (Task 5), JS state + handlers (Task 6), markup + label rename (Task 7), styles + top-pad lift (Task 8), verification (Task 9).
- **Placeholder scan:** No "TBD", "TODO", "fill in", or "similar to Task N". Each step has the actual code or command.
- **Type consistency:** `ListSessionProgress` (with capital L) in Task 1 matches the call site updated in Task 1 step 2 and the controller call in Task 2 step 1. `ProgressItem.title / progressPct / barColor / barWidth / gameId` in Task 6 step 1 matches the WXML field references `item.title / item.progressPct / item.barColor / item.barWidth / item.gameId` in Task 7 step 1. `prevDisabled / nextDisabled / hasPagination / pageLabel / progressPage / progressTotalPages` set in `rebuildProgress` (Task 6) all match the WXML bindings (Task 7) and the `prevProgressPage / nextProgressPage` early-returns (Task 6). `accentColors.teal` referenced in WXML (Task 7) is set in `data` (Task 6).
- **Convention checks:** No optional chaining or nullish coalescing in `.ts` (per `feedback_dx_mini_no_optional_chaining.md`); no method calls in WXML (per `feedback_dx_mini_wxml_no_method_calls.md` — all display strings pre-computed in `rebuildProgress`); icon names are added to the inventory before the WXML references them (per `feedback_dx_mini_icon_strategy.md`); manual smoke test uses 预览 not 真机调试 (per `project_wechat_devtools_realdevice_bug.md`).
