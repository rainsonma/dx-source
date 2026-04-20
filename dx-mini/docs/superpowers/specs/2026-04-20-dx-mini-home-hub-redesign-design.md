# dx-mini: Home Top-Section Hub Redesign — Design

**Date:** 2026-04-20
**Scope:** `dx-mini` (mostly) + `dx-api` (small backend addition). No changes to `dx-web` or `deploy/`.
**Status:** Approved design; implementation plan pending.

## 1. Goals

Replace the current home-page skeleton (search-bar + stat cards + heatmap) with a teal-themed "hub" top section inspired by the McDonald's mini-program layout, translated to Douxue's teal palette:

1. **Teal header band** with greeting title on the capsule row, a subtitle line, and Lv/grade chips — the band finishes with an oval bottom so it visually encloses the search bar.
2. **Combined two-column card** below the teal band (one container, 1px vertical divider) advertising 升级 VIP and 奖励计划 — background is a pale teal → pink vertical gradient.
3. **Five gradient-teal circles** below the card: 学习 / 群组 / 打卡 / 留言 / 建议.
4. **Remove the old hero-section** (greeting/subtitle/stat row) and the heatmap — this page is now a hub, not a dashboard.
5. **Four new stub pages** under `/pages/me/` for destinations that don't exist yet: study, tasks, community, feedback.
6. **One small backend addition:** expose `level` in the dashboard profile payload so the mini doesn't have to port the level table.

## 2. Non-goals

- No change to other tab pages (games / leaderboard / me / learn / login) beyond adding routes to `app.json`.
- No redesign of the purchase, invite, groups, or notices pages — the hub links to them as-is.
- No real implementation inside the 4 new stub pages — each is a single "敬请期待" screen.
- No stat-card or heatmap replacement elsewhere in the app (yet). The old `/api/hall/heatmap` endpoint stays on the server; the mini simply stops calling it.
- No change to `dx-web` or `deploy/`. The backend change is a single new field on an existing response — no new endpoint, no migration.
- No change to the tab bar or the custom-nav status-bar padding idiom introduced in the prior PR.

## 3. Constraints

- **Custom nav style** is already in effect for the home page (from the previous tab-page-chrome work). The WeChat capsule stays in the top-right; the greeting title occupies the left of the same horizontal band.
- **`<dx-icon>` is the only sanctioned icon primitive** — all new icons go through `scripts/build-icons.mjs`, regenerated via `npm run build:icons`.
- **No `?.` / `??`** in dx-mini TS or WXML — use explicit null checks or `||`.
- **TypeScript strict mode** must still pass. No new error categories beyond the tolerated `this`-in-`Component` pattern.
- **WXSS supports ellipse border-radius** (`border-bottom-left-radius: 50% 28px`) and standard gradients — the oval curve and the pale teal→pink gradient are pure CSS, no SVG clip needed.
- **No `?.`/`??` in template interpolations either** — the existing home.wxml already uses `{{greeting ? greeting.title : '...'}}`, which we'll extend with the same pattern for nickname + grade.

## 4. Architecture

### 4.1 Page structure (top → bottom)

```
<page-container> (custom nav)
  <teal-wrap>                           ← teal gradient, oval bottom
    <status-bar-spacer>                  ← reserves statusBarHeight px
    <nav-row>                            ← 40px band, same height as capsule
      <nav-greet>{greeting.title} {nickname}</nav-greet>
      (capsule occupies right ~100rpx, rendered by WeChat)
    </nav-row>
    <greet-body>
      <greet-sub>{greeting.subtitle}</greet-sub>
      <badge-row>
        <badge class="lvl">Lv.{profile.level}</badge>
        <badge>{gradeLabel(profile.grade)}</badge>
      </badge-row>
    </greet-body>
    <search-row>
      <search-box bind:tap="goSearch">
        <dx-icon name="search" .../>
        <text>搜索课程</text>
      </search-box>
    </search-row>
  </teal-wrap>

  <combined-card>                        ← margin-top: -12px (rides into teal)
    <card-half bind:tap="goPurchase">
      <card-text>
        <h4>升级 VIP</h4>
        <p>选择适合您的会员方案，或通过兑换码升级会员，解锁更多学习功能</p>
      </card-text>
      <illo.vip>
        <dx-icon name="crown" color="#fff"/>
      </illo.vip>
    </card-half>
    <divider/>
    <card-half bind:tap="goInvite">
      <card-text>
        <h4>奖励计划</h4>
        <p>如果喜欢斗学就推荐给好朋友一起来快乐学习吧！</p>
      </card-text>
      <illo.gift>
        <dx-icon name="gift" color="#fff"/>
      </illo.gift>
    </card-half>
  </combined-card>

  <circle-row>                           ← 5 equal-width columns
    <circle-item bind:tap="goStudy">    <circle class="c1"><dx-icon name="chart-pie"/></circle>      <label>学习</label></circle-item>
    <circle-item bind:tap="goGroups">   <circle class="c2"><dx-icon name="users"/></circle>          <label>群组</label></circle-item>
    <circle-item bind:tap="goTasks">    <circle class="c3"><dx-icon name="calendar-check"/></circle> <label>打卡</label></circle-item>
    <circle-item bind:tap="goCommunity"><circle class="c4"><dx-icon name="sticker"/></circle>        <label>留言</label></circle-item>
    <circle-item bind:tap="goFeedback"> <circle class="c5"><dx-icon name="flag"/></circle>           <label>建议</label></circle-item>
  </circle-row>
</page-container>
```

### 4.2 Visual spec (matches approved mockup `layout-mockup-v10.html`)

| Element | Spec |
|---|---|
| Page background | `var(--bg-page)` (unchanged) |
| Teal band background | `linear-gradient(180deg, #0d9488 0%, #0f766e 100%)` |
| Teal band oval bottom | `border-bottom-left-radius: 50% 28px; border-bottom-right-radius: 50% 28px` |
| Status bar reserve | existing `padding-top: calc(var(--status-bar-height, 20px) + 0rpx)` — band starts at the very top so teal colors under the status bar |
| Nav row height | 40px, `align-items: flex-end`, right-padding 110px to clear capsule |
| Greeting title | `17px 700`, ellipsis clip if too long |
| Subtitle | `13px` white @ 92% opacity, `margin-top: 2px` |
| Badges | `11px` in pill shape, `bg: rgba(255,255,255,0.2)`; Lv badge at `rgba(255,255,255,0.32)` bold |
| Search box | white bg, 22px radius, 10×14 padding, `box-shadow: 0 6px 18px rgba(0,0,0,0.12)` |
| Combined card | `margin: -12px 14px 0`, `border-radius: 14px`, `background: linear-gradient(180deg, #f0fdfa 0%, #f0fdfa 55%, #fdf2f8 100%)`, `box-shadow: 0 6px 18px rgba(13,148,136,0.10)` |
| Card grid | `grid-template-columns: 1fr 1px 1fr` |
| Divider column | `background: rgba(148,163,184,0.25); margin: 12px 0` |
| Card half | `padding: 14px 12px; display: flex; align-items: center; gap: 10px` |
| Card title | `15px 700`, color `#0f766e`, trailing `▸` in pink `#db2777` |
| Card body text | `10px`, color `#4b5563`, `line-height: 1.45` |
| Card illustration square | 40×40, radius 12; VIP gradient `#0d9488 → #14b8a6`; 奖励 gradient `#ec4899 → #f472b6` |
| Circle | 48×48, `border-radius: 50%`, `box-shadow: 0 4px 10px rgba(13,148,136,0.22)` |
| Circle gradients (5 variants, teal family) | c1 `#0d9488→#14b8a6`, c2 `#0f766e→#2dd4bf`, c3 `#0891b2→#14b8a6`, c4 `#059669→#14b8a6`, c5 `#115e59→#0d9488` |
| Circle icon | `<dx-icon size="22px" color="#fff">`, stroke-width default (1.25) |
| Circle label | `12px`, `margin-top: 8px` |

### 4.3 Dark mode

Add a `.dark` variant branch to the new styles so the page doesn't look broken when the user toggles theme on the me page:

| Element | Light | Dark |
|---|---|---|
| Teal band gradient | `#0d9488 → #0f766e` | `#14b8a6 → #0d9488` (slightly brighter top per `--primary` override) |
| Search box bg | `#fff` | `var(--bg-card)` i.e. `#1c1c1e` |
| Search placeholder color | `#9ca3af` | `#6b7280` |
| Search icon color | `#9ca3af` | `#6b7280` |
| Combined card gradient | `#f0fdfa → #fdf2f8` | `rgba(20,184,166,0.08) → rgba(236,72,153,0.08)` |
| Card title color | `#0f766e` | `#2dd4bf` |
| Card body color | `#4b5563` | `#9ca3af` |
| Circle label color | `#1a1a1a` | `#f5f5f5` |
| Divider color | `rgba(148,163,184,0.25)` | `rgba(255,255,255,0.06)` |

All other circle / illo gradients stay the same in dark mode — gradient-on-white contrast looks fine against the darker page.

### 4.4 Data flow

- The page still calls `api.get<DashboardData>('/api/hall/dashboard')` on `onShow` via `loadData()`.
- `DashboardData.profile` now includes `level: number` from the new backend field.
- `greeting` (title + random subtitle) continues to come from the same endpoint — no client-side greeting logic.
- The `/api/hall/heatmap` call goes away — the mini no longer fetches it.
- `masterStats`, `reviewStats`, `todayAnswers`, and `heatmapCells` are dropped from `Page.data`; same for `buildHeatmapCells()`.

### 4.5 Navigation targets

| Tile | Action |
|---|---|
| 升级 VIP card | `wx.navigateTo({ url: '/pages/me/purchase/purchase' })` |
| 奖励计划 card | `wx.navigateTo({ url: '/pages/me/invite/invite' })` |
| 学习 (chart-pie) | `wx.navigateTo({ url: '/pages/me/study/study' })` **(new stub)** |
| 群组 (users) | `wx.navigateTo({ url: '/pages/me/groups/groups' })` |
| 打卡 (calendar-check) | `wx.navigateTo({ url: '/pages/me/tasks/tasks' })` **(new stub)** |
| 留言 (sticker) | `wx.navigateTo({ url: '/pages/me/community/community' })` **(new stub)** |
| 建议 (flag) | `wx.navigateTo({ url: '/pages/me/feedback/feedback' })` **(new stub)** |
| Search box | `wx.navigateTo({ url: '/pages/games/games' })` (unchanged from current `goSearch`) |

All home-tile taps use `wx.navigateTo`; none of the targets are tab pages, so no `wx.switchTab` needed.

### 4.6 Stub pages

Each of the four new stub pages has the same shape:

- **`*.json`** — `{ "navigationBarTitleText": "<中文 label>", "usingComponents": { "dx-icon": "/components/dx-icon/index" } }`. Uses the default nav bar (back-button visible) — unlike tab pages, these are pushed via navigateTo and need the back button.
- **`*.wxml`** — a centered placeholder:
  ```wxml
  <view class="stub">
    <dx-icon name="clock" size="48px" color="#9ca3af" />
    <text class="stub-title">敬请期待</text>
    <text class="stub-desc">此功能正在开发中</text>
  </view>
  ```
- **`*.ts`** — empty `Page({})` (no data, no handlers).
- **`*.wxss`** — flex-center the stub content block.

Shared stub CSS can live inline per page — 4 files is under the duplication threshold. A shared component would be overkill.

### 4.7 Backend: add `level` to `DashboardProfile`

`dx-api/app/services/api/hall_service.go`:

```go
type DashboardProfile struct {
    ID                string  `json:"id"`
    Username          string  `json:"username"`
    Nickname          *string `json:"nickname"`
    Grade             string  `json:"grade"`
    Level             int     `json:"level"`          // NEW
    Exp               int     `json:"exp"`
    Beans             int     `json:"beans"`
    AvatarURL         *string `json:"avatarUrl"`
    CurrentPlayStreak int     `json:"currentPlayStreak"`
    InviteCode        string  `json:"inviteCode"`
    LastReadNoticeAt  any     `json:"lastReadNoticeAt"`
    CreatedAt         any     `json:"createdAt"`
}
```

Populate in `GetDashboard`:

```go
level, err := consts.GetLevel(user.Exp)
if err != nil {
    return nil, fmt.Errorf("failed to compute user level: %w", err)
}

profile := DashboardProfile{
    ID:                user.ID,
    // … existing fields …
    Level:             level,
    Exp:               user.Exp,
    // …
}
```

The `user/profile` endpoint already exposes `level` (see `user_service.go:16`). This just brings the hall dashboard endpoint to parity — a single source of truth via `consts.GetLevel`.

**Why backend, not client-side computation:** the level table is non-trivial (101 entries derived from an exponential formula in `consts/user_level.go`). Duplicating it in the mini would create two sources of truth. A JSON field costs 1 extra int per request.

### 4.8 Icon inventory additions

`dx-mini/scripts/build-icons.mjs` ICONS array gets four new rows:

```js
['chart-pie',     'chart-pie'],
['calendar-check','calendar-check'],
['sticker',       'sticker'],
['flag',          'flag'],
```

All four exist in `lucide-static@0.460` under their canonical filenames (verified). Run `npm run build:icons` to regenerate `miniprogram/components/dx-icon/icons.ts`.

## 5. Files changed

**Modified** (3 files):

- `dx-mini/miniprogram/pages/home/home.wxml` — full rewrite of body (see §4.1). Container wrapper and `van-config-provider`/`van-skeleton` remain.
- `dx-mini/miniprogram/pages/home/home.ts` — remove `masterStats`, `reviewStats`, `todayAnswers`, `heatmapCells`, `HeatmapDay`, `HeatmapData`, `HeatmapCell` types, `buildHeatmapCells()`, and the `api.get<HeatmapData>('/api/hall/heatmap')` call. Add `level: number` to the `DashboardProfile` TS interface (matches the new JSON field). Add `goPurchase`, `goInvite`, `goStudy`, `goGroups`, `goTasks`, `goCommunity`, `goFeedback` handlers. Keep `goSearch`.
- `dx-mini/miniprogram/pages/home/home.wxss` — drop the old `.hero-section`, `.greeting-*`, `.stat-*`, `.section`, `.section-title`, `.heatmap`, `.heatmap-cell*` rules. Add the new teal-wrap / greet-body / search / combined-card / circle rules plus their `.dark` variants.
- `dx-mini/miniprogram/pages/home/home.json` — unchanged. (`dx-icon`, `van-config-provider`, `van-skeleton` all still needed; `navigationStyle: custom` stays.)

**Created** (4 new stub pages × 4 files = 16 files):

- `dx-mini/miniprogram/pages/me/study/study.{json,ts,wxml,wxss}`
- `dx-mini/miniprogram/pages/me/tasks/tasks.{json,ts,wxml,wxss}`
- `dx-mini/miniprogram/pages/me/community/community.{json,ts,wxml,wxss}`
- `dx-mini/miniprogram/pages/me/feedback/feedback.{json,ts,wxml,wxss}`

**Modified** (1 file):

- `dx-mini/miniprogram/app.json` — add the 4 new page paths to the `pages` array.

**Modified** (1 file):

- `dx-mini/scripts/build-icons.mjs` — 4 new ICONS rows.

**Regenerated** (1 file):

- `dx-mini/miniprogram/components/dx-icon/icons.ts` — product of `npm run build:icons`, not hand-edited.

**Modified** (1 file, backend):

- `dx-api/app/services/api/hall_service.go` — add `Level` field + populate via `consts.GetLevel`.

**Total:** 3 mini-page edits, 16 new stub files, 1 app.json edit, 1 build-script edit, 1 regenerated icon file, 1 backend edit.

**Deleted:** none.

## 6. Migration order

Ordered so the mini keeps compiling at every checkpoint:

1. **Backend first.** Add `Level` to `DashboardProfile` in `hall_service.go`, wire in `consts.GetLevel`. Run `go build ./...` and `go test -race ./app/services/api/...` — should pass.
2. **Icons.** Append 4 rows to `scripts/build-icons.mjs`, run `npm run build:icons` in `dx-mini/`. The build script's static WXML scan currently rejects undeclared names, so we do this step BEFORE touching wxml files that reference them — otherwise `build:icons` fails on the first run after the WXML edit.
3. **Stub pages.** Create the 4 new page directories with their 4 files each. Add their paths to `app.json` so WeChat DevTools registers them.
4. **Home page rewrite.** Edit `home.wxml`, `home.ts`, `home.wxss` in that order. Keep the outer `<van-config-provider>` + `<van-skeleton>` wrapper and `.page-container` + `--status-bar-height` plumbing; replace only the body.
5. **Smoke test.** Load the home page in DevTools on both light and dark themes. Tap every tile. Verify each destination opens (purchase + invite + groups are real; the 4 stubs show "敬请期待").
6. **Commit.** One `feat(mini): home hub redesign` commit covering the mini changes; a separate `feat(api): expose level on hall dashboard` commit for the backend — or one combined `feat: home hub redesign` commit. Final call in the implementation plan.

## 7. Verification

### 7.1 Static checks

- `grep -rn "'hero-section'\|greeting-title\|heatmap-cell" dx-mini/miniprogram/pages/home/` — zero matches (dead styles/ids removed).
- `grep -rn "heatmap\|buildHeatmapCells\|/api/hall/heatmap" dx-mini/miniprogram/pages/home/` — zero matches.
- `grep -rn "<dx-icon" dx-mini/miniprogram/pages/home/home.wxml | wc -l` — exactly 8 (search + crown + gift + 5 circle icons).
- `grep -rn "name=\"chart-pie\"\|name=\"calendar-check\"\|name=\"sticker\"\|name=\"flag\"" dx-mini/miniprogram/pages/home/home.wxml` — one match each.
- `grep -rn "chart-pie\|calendar-check\|sticker\|flag" dx-mini/scripts/build-icons.mjs` — all four present.
- `grep -rn "pages/me/study\|pages/me/tasks\|pages/me/community\|pages/me/feedback" dx-mini/miniprogram/app.json` — all four present.
- `grep -rn "Level " dx-api/app/services/api/hall_service.go` — matches both the struct field and the `GetLevel` call.
- `npx -p typescript@5 tsc --noEmit` in `dx-mini/` — no new errors beyond the tolerated `this`-in-`Component` pattern.
- `go build ./...` and `go test -race ./...` in `dx-api/` — pass.

### 7.2 DevTools smoke test

- **Home page light mode:** teal band with greeting title on capsule row, subtitle + Lv.X + grade chips stacked below, search bar inside teal, oval curve under search, combined card overlapping teal bottom by 12px, 5 circles below. No visual artifacts at teal/white boundary.
- **Home page dark mode:** teal band uses dark-variant teal, search box dark, card gradient uses the rgba version, all text colors match §4.3.
- **Tap every tile:**
  - 升级 VIP → `/pages/me/purchase/purchase` — existing flow renders
  - 奖励计划 → `/pages/me/invite/invite` — existing flow renders
  - 学习 → `/pages/me/study/study` — "敬请期待" stub
  - 群组 → `/pages/me/groups/groups` — existing list
  - 打卡 → `/pages/me/tasks/tasks` — stub
  - 留言 → `/pages/me/community/community` — stub
  - 建议 → `/pages/me/feedback/feedback` — stub
- **Search box tap** → `/pages/games/games` (unchanged from previous behavior).
- **Long nickname** (>10 Chinese chars) — greeting title ellipsis-clips, does not wrap into the capsule.
- **Skeleton loading state** — the existing `<van-skeleton>` still wraps the dynamic parts; verify it renders during the dashboard fetch.
- **Back from a stub** — WeChat default back arrow returns to home; home's `onShow` re-fetches the dashboard (existing behavior, unchanged).

### 7.3 Backend check

- `curl` the dashboard endpoint with a valid JWT and confirm the JSON contains `"level": <number>` under `profile`.
- Existing `user_service.go` integration test should still pass (unchanged). Add a minimal assertion in whichever hall test covers `GetDashboard` that the returned level equals `consts.GetLevel(user.Exp)`. If no such test exists, that's out of scope — the field is additive and trivially derived.

### 7.4 Real-device check

Preview via 预览 + 小程序助手 (真机调试 is broken on the current DevTools — project memory). Confirm the oval curve renders and the gradient colors look right on a physical device (iOS + Android both).

## 8. Risks

- **Oval border-radius fidelity on older Android WebViews.** WXSS supports ellipse border-radius on the glass-easel + Skyline renderer (enabled for this project), but fallback to a flat bottom would be visually acceptable if a device ever doesn't honor the ellipse syntax. No code fallback needed — the layout still functions.
- **Gradient rendering on low-end devices.** Pale `#f0fdfa → #fdf2f8` is subtle; on some low-color-depth screens it may look like a flat color. That's fine — the layout doesn't rely on the gradient for hierarchy, just polish.
- **Capsule width varies across devices.** We reserved 110px of right-padding on `.nav-greet`; WeChat capsule is ~87rpx wide + right margin, so this gives safe overlap margin. If a device renders a wider capsule, greeting ellipsis-clips earlier — acceptable fallback.
- **Level exposure.** `Level` is additive to the response, never nullable. No existing client breaks. dx-web's hall page already derives level from exp using a client-side copy — it can be updated in a later PR to consume the field directly (out of scope here).
- **Stub proliferation.** 4 new stub pages now register in `app.json` but do nothing useful. Each is ~30 lines; collectively ~120 lines of dead weight until filled in. Acceptable trade against the alternative (toast-only tiles that feel broken).
- **Old heatmap endpoint unused.** `/api/hall/heatmap` still exists on the backend. dx-web may still call it (not touched in this change). Leaving the endpoint is safe — no cleanup needed in this PR.
