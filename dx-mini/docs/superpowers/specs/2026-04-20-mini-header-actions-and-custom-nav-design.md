# dx-mini: Header-Action Move + Custom Nav on Tab Pages — Design

**Date:** 2026-04-20
**Scope:** `dx-mini` only. No changes to `dx-api`, `dx-web`, or `deploy/`.
**Status:** Approved design; implementation plan pending.

## 1. Goals

Two coupled changes on the four WeChat tab pages:

1. **Relocate the home-page header actions.** The dark-mode toggle (moon/sun) and the notifications bell currently live in `pages/home/home.wxml` as the right-side actions of the top bar. Move them to `pages/me/me.wxml` as a top-right action row, mirroring home's previous pattern. Replace the existing `<van-cell title="公告通知">` entry on the me page — the bell icon is now the single entry point to notices.
2. **Remove the default native title bar** on the four tab pages — `首页` (home), `课程` (games), `排行榜` (leaderboard), `我的` (me) — by switching each to `navigationStyle: "custom"`, with a shared status-bar padding idiom so page content sits cleanly below the status bar and the WeChat capsule button.

## 2. Non-goals

- No other pages change (login already uses custom nav; the learn / games-detail / games-play / me-sub-pages / notices / groups / invite / redeem / purchase / profile-edit pages keep their default title bars).
- No change to bottom tab bar (already custom).
- No visual redesign of search, profile header, or category tabs beyond the mechanical shifts required by the header/nav changes.
- No new icons added — both `moon`/`sun` and `bell` already exist in the curated inventory.
- No change to `dx-api`, `dx-web`, or shared types.

## 3. Constraints

- The four pages are tab-bar destinations — there is no back button to lose when the native nav bar goes away.
- The WeChat capsule button (right-side `⋯ | home-icon` pair) remains visible even with custom nav style; we must not overlap it.
- No `?.` / `??` in dx-mini TS or WXML (project memory).
- TypeScript strict mode must still pass — no new error categories beyond the tolerated `this`-in-`Component` pattern.
- `<dx-icon>` is the only sanctioned icon primitive (project memory).

## 4. Architecture

### 4.1 Header-action move

**`pages/home/home.wxml`** — drop the `<view class="top-actions">` wrapper and both inner `<dx-icon>` buttons. The `top-bar` becomes a single-child row with the search box, which expands to fill the available width.

```wxml
<view class="top-bar">
  <view class="search-box" bind:tap="goSearch">
    <dx-icon name="search" size="16px" color="#9ca3af" />
    <text class="search-placeholder">搜索课程</text>
  </view>
</view>
```

**`pages/home/home.ts`** — delete the `toggleTheme()` and `goNotices()` methods. Nothing else references them.

**`pages/home/home.wxss`** — `.search-box` gets `flex: 1` (or equivalent) so it fills the row now that the right-side wrapper is gone.

**`pages/me/me.wxml`** — prepend a new top-bar block ABOVE the existing `.profile-header`:

```wxml
<view class="top-bar">
  <dx-icon
    name="{{theme === 'dark' ? 'sun' : 'moon'}}"
    size="22px"
    color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
    bind:click="toggleTheme"
  />
  <dx-icon
    name="bell"
    size="22px"
    color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}"
    bind:click="goNotices"
  />
</view>
```

Remove the `<van-cell title="公告通知" is-link bind:click="goNotices">` row (with its `dx-icon slot="icon" name="bell"`) from the `<van-cell-group>` — the bell icon now owns the notices entry point.

**`pages/me/me.ts`** — add a `toggleTheme()` method with identical semantics to the one deleted from home:

```ts
toggleTheme() {
  const next: 'light' | 'dark' = this.data.theme === 'light' ? 'dark' : 'light'
  wx.setStorageSync('dx_theme', next)
  app.globalData.theme = next
  this.setData({
    theme: next,
    primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
    arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',
    cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
  })
  const tabBar = this.getTabBar() as any
  if (tabBar) { tabBar.setData({ theme: next }) }
}
```

Note: me's `toggleTheme` also updates the three derived theme colors (`primaryColor`, `arrowColor`, `cellIconColor`) that the existing me page binds to, which home's version didn't need to do. `goNotices()` already exists on `me.ts` — no change.

**`pages/me/me.wxss`** — add:

```css
.top-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 20rpx;
  padding: 12rpx 32rpx;
}
```

### 4.2 Custom nav bar on tab pages

Apply the same three-part idiom to `home`, `games`, `leaderboard`, and `me`:

**JSON flip** — set `"navigationStyle": "custom"` in each page's `.json`, alongside existing `usingComponents` (do not touch the component list).

**`.ts` `onLoad`** — compute `statusBarHeight` via `wx.getSystemInfoSync()` and store it on page data. For pages that already define `onLoad` (home has one), extend it; for pages without one (games, me, leaderboard), add it.

```ts
onLoad() {
  const sys = wx.getSystemInfoSync()
  const statusBarHeight = sys.statusBarHeight || 20
  this.setData({ statusBarHeight })
  // … existing onLoad body, if any …
}
```

**`.wxml` wrapper** — inject the status-bar value as a CSS custom property on the top-level `.page-container`:

```wxml
<view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
```

**`.wxss` rule** — reserve vertical space below the status bar sufficient to clear the capsule button:

```css
.page-container {
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
}
```

88rpx (≈ 44px) is the WeChat capsule height plus its top/bottom margins. Content that sits in the first row below this padding — home's search bar, me's top-bar, games/leaderboard's `<van-tabs>` — is fully clear of the capsule vertically. Right-aligned content in that row still sits left of the capsule because the capsule occupies only the upper band; once we pass capsule height, the right edge is free.

### 4.3 Why this shape

- **Preserves tab-page ergonomics.** Tab pages lose a redundant title (the bottom tab already labels them); removing the native bar recovers ~44rpx of vertical space per page.
- **No per-device capsule math.** We reserve a fixed-height band the capsule already fits into, rather than calling `wx.getMenuButtonBoundingClientRect()` on every page and reasoning about offsets. Simpler, predictable across devices.
- **Single entry point to notices from the me page.** An icon + a cell reading "公告通知" would be two doors to the same room.
- **Theme toggle stays easy to reach.** It's still in a fixed position, just on the settings-adjacent page instead of the dashboard.

## 5. Files changed

**Modified** (16 files):

- `pages/home/home.wxml` — strip top-actions, keep search; inject `--status-bar-height` on page-container.
- `pages/home/home.ts` — remove `toggleTheme` + `goNotices`; extend `onLoad` with statusBarHeight.
- `pages/home/home.wxss` — `.search-box { flex: 1 }`; `.page-container` top-padding rule.
- `pages/home/home.json` — add `navigationStyle: "custom"`.
- `pages/me/me.wxml` — prepend new `.top-bar`; drop 公告通知 cell; inject `--status-bar-height`.
- `pages/me/me.ts` — add `toggleTheme`; extend `onShow`/`onLoad` with statusBarHeight (see §7 for placement).
- `pages/me/me.wxss` — add `.top-bar` rule; `.page-container` top-padding rule.
- `pages/me/me.json` — add `navigationStyle: "custom"`.
- `pages/games/games.wxml` — inject `--status-bar-height`.
- `pages/games/games.ts` — add `onLoad` with statusBarHeight.
- `pages/games/games.wxss` — `.page-container` top-padding rule.
- `pages/games/games.json` — add `navigationStyle: "custom"`.
- `pages/leaderboard/leaderboard.wxml` — inject `--status-bar-height`.
- `pages/leaderboard/leaderboard.ts` — add `onLoad` with statusBarHeight.
- `pages/leaderboard/leaderboard.wxss` — `.page-container` top-padding rule.
- `pages/leaderboard/leaderboard.json` — add `navigationStyle: "custom"`.

**Created:** none. **Deleted:** none.

## 6. Migration order

Ordered so the app keeps building at every checkpoint:

1. Edit `me.wxml` + `me.ts` + `me.wxss` — add top-bar + `toggleTheme`, remove 公告通知 cell. Me page works in isolation, still under default nav.
2. Edit `home.wxml` + `home.ts` + `home.wxss` — remove right-side icons and their handlers; let search expand.
3. Extend each tab page's `.ts` `onLoad` to compute and expose `statusBarHeight`.
4. Add `.page-container` top-padding rule to each tab page's `.wxss`.
5. Inject `style="--status-bar-height: {{statusBarHeight}}px"` on each `.page-container` wrapper in WXML.
6. Flip each tab page's `.json` to `navigationStyle: "custom"`.
7. Smoke-test in DevTools (see §8).
8. Commit as one `feat(mini)` commit.

## 7. `me.ts` onLoad integration

`me.ts` currently has no `onLoad`; initialization happens in `onShow` (theme/tabBar setup + `loadProfile`). To minimize disturbance, ADD an `onLoad` that computes `statusBarHeight` once and leave `onShow` intact:

```ts
onLoad() {
  const sys = wx.getSystemInfoSync()
  const statusBarHeight = sys.statusBarHeight || 20
  this.setData({ statusBarHeight })
}
```

Apply the same pattern to `games.ts` and `leaderboard.ts` (both also initialize in `onShow` today).

`home.ts` already has `onLoad` — extend rather than replace.

## 8. Verification

### 8.1 Static checks

- `grep -rn "toggleTheme\|goNotices" dx-mini/miniprogram/pages/home/` — zero matches.
- `grep -rn "navigationBarTitleText" dx-mini/miniprogram/pages/home dx-mini/miniprogram/pages/games dx-mini/miniprogram/pages/me/me.json dx-mini/miniprogram/pages/leaderboard` — zero matches. (Scope restricted to the four tab pages' top-level .json — other sub-pages keep their titles.)
- `grep -rn "navigationStyle.*custom" dx-mini/miniprogram/pages/home/home.json dx-mini/miniprogram/pages/games/games.json dx-mini/miniprogram/pages/me/me.json dx-mini/miniprogram/pages/leaderboard/leaderboard.json` — exactly four matches.
- `grep -rn "公告通知" dx-mini/miniprogram/pages/me/` — zero matches.
- `npx -p typescript@5 tsc --noEmit` in `dx-mini/` — no new errors beyond the tolerated `this`-in-`Component` pattern.

### 8.2 DevTools smoke test

Per page, confirm:

- **首页** — no default nav bar; search box fills the row with no right-side gap; content padded below capsule; capsule tappable; theme toggle and bell GONE.
- **课程** — no default nav bar; van-tabs sit under the padded top; capsule does not overlap.
- **排行榜** — no default nav bar; period tabs + type tabs sit under the padded top.
- **我的** — no default nav bar; new top-bar with two right-aligned icons sits fully below capsule; theme toggle (tap) swaps moon ↔ sun, recolors the whole app including the tab-bar; bell navigates to `/pages/me/notices/notices`; 公告通知 cell is gone from the cell group.
- Dark theme toggle still works and propagates to the other three pages (via `app.globalData.theme` + tab-bar refresh).

### 8.3 Real-device check

Preview via 预览 + 小程序助手 (真机调试 is broken — project memory).

## 9. Risks

- **Status bar padding miscalibration on notched devices.** If `statusBarHeight` reports a different value on a specific device, the 88rpx reserve for the capsule may feel slightly off. Mitigation: the `calc(var(--status-bar-height, 20px) + 88rpx)` formula uses the device-reported status bar plus a fixed capsule reserve — this is the WeChat-community-standard pattern. If needed, switch to `getMenuButtonBoundingClientRect()` for exact capsule bottom in a follow-up.
- **me page theme toggle interactions.** The new `toggleTheme` on `me.ts` also updates `primaryColor`/`arrowColor`/`cellIconColor` so the page recolors immediately. Home's `toggleTheme` didn't need this because home re-derives its colors inline in WXML. Keep both behaviors; no cross-page coupling.
- **Commit granularity.** Single commit keeps the per-page edit atomic (each me/home file is touched for both purposes). If bisect becomes a concern later, the per-page diff is still small and readable.
