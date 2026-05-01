# 学习 page — add 学习进度 card, rename labels, lift to top of page

**Date:** 2026-05-01
**Scope:** dx-mini (`pages/learn/learn.{json,ts,wxml,wxss}`, `scripts/build-icons.mjs`, regenerated `components/dx-icon/icons.ts`); dx-api (`routes/api.go`, `app/services/api/hall_service.go`, new `app/http/controllers/api/user_session_controller.go`)
**Stakes:** UI addition on the learn tab page + one new read-only API endpoint that wraps an existing query. No DB schema, no auth changes, no impact on other pages or other clients (dx-web).

## Goal

The 学习 page is currently a thin shell — a row of three stat cards (已掌握 / 不认识 / 待复习) and a quick-links list. The dx-web hall already shows users a "我的学习进度" card listing their recent game sessions with progress bars; we want the same surface inside the mini program, placed above the existing stat row, so users see what to resume the moment they hit the 学习 tab.

Alongside that, two label adjustments and a layout cleanup:

- Rename "不认识" → "生词本" (and "不认识词库" → "生词本词库") so the surface matches the friendlier copy the rest of the product is moving toward. The underlying data flow is untouched — only display strings change.
- Switch the page to the same `navigationStyle: "custom"` + status-bar-spacer pattern every other tab page already uses, so the new card sits flush near the top of the screen instead of below WeChat's default nav title.

## Current state

```
[default WeChat nav bar — title "学习"]   ← takes ~44+statusBar pts
[16px content padding-top]
[stat-card-row  已掌握 / 不认识 / 待复习]
[quick-links    已掌握词库 / 不认识词库 / 待复习队列]
```

`learn.ts` fires three parallel `/api/tracking/*/stats` calls in `onShow` and renders the row + list. There is no game-session listing.

## Target state

```
[--status-bar-height + 88rpx top spacer]   ← clears system status bar + WeChat capsule
[progress-card  我的学习进度
                ┌ row · row · row · row · row ┐  (5 per page, up to 20 total)
                └ « prev   第 N / M 页   next »┘  (hidden when ≤5 sessions)
                — OR —
                [empty state: gamepad-2 + 还没有学习进度 + 去发现课程游戏吧]
[stat-card-row   已掌握 / 生词本 / 待复习]      ← middle label renamed
[quick-links     已掌握词库 / 生词本词库 / 待复习队列]   ← middle label renamed
```

The default WeChat nav title bar is gone. The WeChat capsule (top-right `…/×` pill) sits over the empty right edge of the top spacer — content begins below it.

## Architecture decisions

### Data source — new endpoint, not the dashboard reuse

We add `GET /api/tracking/sessions` (under the existing `protected.Prefix("/tracking")` group). It returns the same `[]SessionProgress` slice that `/api/hall/dashboard` already returns inside its `sessions` field, by exporting and reusing the existing `getSessionProgress` SQL helper in `hall_service.go`. The dashboard contract used by the home page stays byte-for-byte identical.

**Why not reuse `/api/hall/dashboard`:** that endpoint also recomputes master and review stats; calling it from the learn page would duplicate work the learn page already does via the three tracking-stats endpoints.

**Why not extend an existing tracking-stats endpoint:** sessions is a different shape from stats, and overloading an unrelated endpoint hurts readability.

### Pagination — inline custom prev/next, no `<van-pagination>`

The card holds up to 20 sessions paginated 5-per-page. We render a simple `<view>`-based prev/next bar with disabled-state styling (gray text, no tap response via class + `bind:tap` guard). Hidden entirely when `sessions.length <= 5`. This matches the lightweight feel of the rest of the learn page and avoids pulling Vant's heavier paginator.

### "查看全部" — omitted

The dx-web reference links to `/hall/games/mine`. dx-mini has no equivalent route, and inline pagination over 20 sessions is enough for this surface. Adding a new "我的进度" page would be net-new scope without payoff.

### Icons — add `list-checks` and `gamepad-2` to the canonical inventory

Append both to `scripts/build-icons.mjs`'s `ICONS` array, run `npm run build:icons` to regenerate `components/dx-icon/icons.ts`. `list-checks` for the card header (matches dx-web), `gamepad-2` for the empty state (matches dx-web).

### Top spacing — leaderboard pattern

`navigationStyle: "custom"`, `padding-top: calc(var(--status-bar-height, 20px) + 88rpx)` on `.page-container`, drop the existing `.content { padding: 16px }` top component (keep horizontal). Mirrors `pages/leaderboard/leaderboard.wxss` exactly.

### Card stays inline (not a Component)

It has one consumer. Extracting prematurely doesn't pay off. If a second consumer ever appears, extract then.

### `GAME_MODE_LABELS` stays inline in `learn.ts`

Four entries, one consumer. No shared util yet.

## Changes

### dx-api

#### `app/services/api/hall_service.go`

Rename the private function and export it:

```go
// Before:
func getSessionProgress(userID string) ([]SessionProgress, error) { ... }

// After:
func ListSessionProgress(userID string) ([]SessionProgress, error) { ... }
```

Update the one call site inside `GetDashboard` to use the new name. Body of the function is unchanged — same SQL, same `LIMIT 20`, same ordering by `MAX(s.last_played_at) DESC`.

#### `app/http/controllers/api/user_session_controller.go` (new file)

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

func NewUserSessionController() *UserSessionController { return &UserSessionController{} }

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

#### `routes/api.go`

Inside the existing `protected.Prefix("/tracking").Group(...)` block (between the master/unknown/review subgroups), add:

```go
userSessionController := apicontrollers.NewUserSessionController()
tracking.Get("/sessions", userSessionController.ListSessions)
```

Response envelope is `{code:0, message:"ok", data: SessionProgress[]}` — matches the rest of the API surface.

### dx-mini

#### `scripts/build-icons.mjs`

Append two entries to the `ICONS` array (logical → lucide filename):

```js
['list-checks', 'list-checks'],
['gamepad-2',   'gamepad-2'],
```

Then run `npm run build:icons` once to regenerate `miniprogram/components/dx-icon/icons.ts`. Don't hand-edit `icons.ts`.

#### `miniprogram/pages/learn/learn.json`

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

`navigationBarTitleText` removed.

#### `miniprogram/pages/learn/learn.ts`

Add to `data`:
- `statusBarHeight: number` (default 20)
- `sessions: SessionProgress[]` (raw, full list — up to 20)
- `progressItems: ProgressItem[]` (pre-computed display fields)
- `pageItems: ProgressItem[]` (current visible 5)
- `progressPage: number` (1-indexed)
- `progressTotalPages: number` (≥ 1)
- `hasPagination: boolean` (true iff `sessions.length > 5`)
- `prevDisabled: boolean`, `nextDisabled: boolean`
- `pageLabel: string` (`第 X / Y 页`)

Local `const` (top of file):

```ts
const PAGE_SIZE = 5
const PROGRESS_COLORS = ['#14b8a6', '#3b82f6', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4']
const GAME_MODE_LABELS: Record<string, string> = {
  'word-sentence':       '连词成句',
  'vocab-battle':         '词汇对轰',
  'vocab-match':          '词汇配对',
  'vocab-elimination':    '词汇消消乐',
}
```

`onLoad` (new):
```ts
onLoad() {
  const sys = wx.getSystemInfoSync()
  this.setData({ statusBarHeight: sys.statusBarHeight || 20 })
}
```

`onShow` keeps the existing theme + arrowColor sync, then calls `loadAll()`.

`loadAll()` replaces `loadStats()`:

```ts
async loadAll() {
  this.setData({ loading: true })
  const results = await Promise.allSettled([
    api.get<Stats>('/api/tracking/master/stats'),
    api.get<Stats>('/api/tracking/unknown/stats'),
    api.get<ReviewStats>('/api/tracking/review/stats'),
    api.get<SessionProgress[]>('/api/tracking/sessions'),
  ])
  const masterStats  = results[0].status === 'fulfilled' ? results[0].value : this.data.masterStats
  const unknownStats = results[1].status === 'fulfilled' ? results[1].value : this.data.unknownStats
  const reviewStats  = results[2].status === 'fulfilled' ? results[2].value : this.data.reviewStats
  const sessions     = results[3].status === 'fulfilled' ? results[3].value : []

  this.setData({ loading: false, masterStats, unknownStats, reviewStats, sessions })
  this.rebuildProgress(1)

  if (results.some((r) => r.status === 'rejected')) {
    wx.showToast({ title: '加载失败', icon: 'none' })
  }
}
```

`rebuildProgress(targetPage: number)`:

1. Build `progressItems` from `sessions`:
   - `title = ${gameName} · ${GAME_MODE_LABELS[gameMode] || gameMode}`
   - `progressPct = totalLevels === 0 ? 0 : Math.round(completedLevels / totalLevels * 100)`
   - `barColor = PROGRESS_COLORS[i % PROGRESS_COLORS.length]`
   - `barWidth = ${progressPct}%`
2. `totalPages = Math.max(1, Math.ceil(progressItems.length / PAGE_SIZE))`
3. Clamp `page = Math.min(Math.max(1, targetPage), totalPages)`
4. Slice `pageItems` from `(page-1)*PAGE_SIZE` to `page*PAGE_SIZE`
5. Compute `hasPagination = progressItems.length > PAGE_SIZE`, `prevDisabled = page === 1`, `nextDisabled = page === totalPages`, `pageLabel = '第 ' + page + ' / ' + totalPages + ' 页'`
6. `setData` once with all of the above

Handlers:
```ts
prevProgressPage() { if (!this.data.prevDisabled) this.rebuildProgress(this.data.progressPage - 1) },
nextProgressPage() { if (!this.data.nextDisabled) this.rebuildProgress(this.data.progressPage + 1) },
goGame(e: WechatMiniprogram.TouchEvent) {
  const id = e.currentTarget.dataset['id']
  if (id) wx.navigateTo({ url: '/pages/games/detail/detail?id=' + id })
},
```

Existing `goMastered` / `goUnknown` / `goReview` handlers are unchanged.

#### `miniprogram/pages/learn/learn.wxml`

Top-level shell switches to leaderboard pattern:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <view wx:if="{{!loading}}" class="content">

      <!-- 学习进度 card -->
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

      <!-- existing stat row, label change only -->
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

      <!-- existing quick links, label change only -->
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

#### `miniprogram/pages/learn/learn.wxss`

Top-padding switch + new card styles:

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
  background: rgba(13,148,136,0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
.dark .pc-icon-wrap { background: rgba(20,184,166,0.16); }
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
.pc-row-title { font-size: 13px; color: var(--text-primary); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding-right: 8px; }
.pc-row-pct { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.pc-bar-track { height: 6px; border-radius: 999px; background: var(--bg-page); overflow: hidden; }
.dark .pc-bar-track { background: rgba(255,255,255,0.06); }
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

/* existing styles below — unchanged */
.stat-card-row { display: flex; gap: 10px; margin-bottom: 20px; }
/* ... rest same as before ... */
```

The `.content { padding: 16px }` rule's top-pad is removed because the page container now handles it; horizontal stays at 16px. The `min-height: 100vh; background; padding-bottom` block on `.page-container` replaces the existing rule entirely.

## Edge cases

| Case | Behavior |
|------|----------|
| `sessions.length === 0` | `pc-empty` block renders; `hasPagination = false`; pager hidden. |
| `sessions.length === 5` | One page; `hasPagination = false`; pager hidden (no point). |
| `sessions.length > 5` (up to 20) | Pager visible; prev disabled on page 1; next disabled on last page. |
| `gameMode` not in `GAME_MODE_LABELS` | Title falls back to raw mode string (mirrors dx-web `game-progress-card.tsx`). |
| `totalLevels === 0` | `progressPct = 0`; bar renders empty (mirrors dx-web `calcProgress`). |
| One of the four parallel fetches fails | The other three still update; failed slot retains previous value (or `[]` for sessions); single toast surfaces "加载失败". |
| Page reopened with fewer sessions than before | `rebuildProgress` clamps `page` into `[1, totalPages]` so a stale `progressPage` never points past the array. |
| Theme switch while page is open | CSS variables propagate via `<van-config-provider theme>`; no explicit re-render needed for the new card. |
| Tap a row mid-pagination | `goGame` reads `data-id`; navigation proceeds; no race with pager state. |

## Verification

Must all pass before declaring done:

1. `cd dx-api && go build ./...` — clean compile.
2. `cd dx-api && go vet ./...` — clean.
3. `cd dx-api && go test -race ./...` — existing tests still pass; if existing tracking endpoints have controller-level tests, mirror that pattern for `ListSessions`.
4. `cd dx-mini && npx tsc --noEmit -p tsconfig.json` — zero NEW errors beyond the pre-existing `Component({methods})` typing pattern noted in `feedback_dx_mini_no_optional_chaining.md`'s sibling memory.
5. Re-run `cd dx-mini && npm run build:icons` once to regenerate `icons.ts`; the script's static WXML scan will fail loudly if `<dx-icon name="list-checks">` or `<dx-icon name="gamepad-2">` references aren't covered by the inventory entries.
6. WeChat DevTools manual smoke test (use 预览, not 真机调试 — see `project_wechat_devtools_realdevice_bug.md` memory):
   - Empty state renders for a brand-new account / one with no sessions
   - 6+ sessions: pagination works in both directions; disabled states render gray and don't tap
   - Tap a row → navigates to `/pages/games/detail/detail?id=<gameId>`
   - Dark theme: card frame, text, bar track, accent fill all readable
   - WeChat capsule does NOT overlap any tap target on the new card or anywhere in either theme
   - "不认识" → "生词本" label visible in stat row
   - "不认识词库" → "生词本词库" label visible in quick links list
   - Default WeChat title bar gone; first content (the new card) sits flush below the system status bar / capsule
   - Other tab pages still load and switch correctly (no regression)

## Out of scope

- No new `pages/learn/progress` full-page list, no "查看全部" link.
- No shared `game-mode-labels.ts` util in dx-mini — inline map only.
- No reusable progress-card Component — single-consumer markup stays inline.
- No `<van-pagination>` — custom prev/next bar.
- No DB schema changes; the new endpoint reuses the existing SQL byte-for-byte.
- No rename of `/api/tracking/unknown/*` routes, the `pages/learn/unknown` page name, or the `goUnknown` handler — only display labels change. (A full "unknown → vocab" rename across the codebase is a separate clean-up, not part of this spec.)
- No change to `/api/hall/dashboard` contract or to the home page that consumes it.
- No theme tweaks beyond the existing CSS variables in `app.wxss`.

## File list

**dx-api** (3 changed, 1 new):
- `routes/api.go`
- `app/services/api/hall_service.go`
- `app/http/controllers/api/user_session_controller.go` *(new)*
- (optional) `tests/...` matching existing tracking-test pattern

**dx-mini** (5 changed, 1 regenerated):
- `scripts/build-icons.mjs`
- `miniprogram/components/dx-icon/icons.ts` *(regenerated by `npm run build:icons`, do not hand-edit)*
- `miniprogram/pages/learn/learn.json`
- `miniprogram/pages/learn/learn.ts`
- `miniprogram/pages/learn/learn.wxml`
- `miniprogram/pages/learn/learn.wxss`
