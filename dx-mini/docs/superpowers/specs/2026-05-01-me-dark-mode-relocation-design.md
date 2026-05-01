# 我的 page — relocate dark-mode toggle, lift content, tidy cell rows

**Date:** 2026-05-01
**Scope:** dx-mini only — `pages/me/me.{wxml,wxss,ts}`
**Stakes:** UI polish on a tab page; no data, API, or auth changes

## Goal

The 我的 page currently reserves a full top-bar row solely for a sun/moon dark-mode icon. Move that toggle into the menu cell-group as a labeled row, then lift the rest of the page flush to the system status bar so the page no longer looks visually empty at the top. While we're in the cells, fix two pre-existing visual nits: the icons sit flush against their titles with no breathing room, and they're not vertically centered with the title text.

## Current state

```
[--status-bar-height + 88rpx empty padding]   ← page-container padding-top
[top-bar  ↘ sun/moon icon, right-aligned]      ← ~88rpx tall, otherwise empty
[profile-header   avatar  name+badges  chevron-right] ← whole row taps → profile edit
[stats-bar  beans / exp / streak]
[vip-bar (conditional)]
[van-cell-group   我的团队 / 推荐有礼 / 兑换码 / 购买会员 / 收藏的课程]
[van-cell-group   退出登录]
```

Issues:
1. Top-bar is a wasted row holding a single icon.
2. Inside cell rows, `slot="icon"` is rendered with no left-icon wrapper margin (Vant only gives the 4px wrap to the built-in `icon` prop, not the slot), so dx-icon sits flush against the title.
3. Cells lack `center`, so the icon and title flex children align to the top of the cell rather than the same horizontal line.

## Target state

```
[--status-bar-height empty]                    ← only system status bar cleared
[profile-header   avatar  name+badges]         ← chevron removed; whole row still tappable
[stats-bar  beans / exp / streak]
[vip-bar (conditional)]
[van-cell-group   我的团队 / 推荐有礼 / 兑换码 / 购买会员 / 收藏的课程 / 深色模式]
[van-cell-group   退出登录]
```

WeChat capsule (the …/× pill in the top-right) sits over the empty right edge of `profile-header` — no content collision because the chevron is gone and the avatar is on the left.

## Changes

### `pages/me/me.wxml`

1. **Delete the top-bar block.** Remove the entire wrapper:

   ```xml
   <view class="top-bar">
     <dx-icon name="..." ... bind:click="toggleTheme" />
   </view>
   ```

2. **Drop the chevron from `profile-header`.** Remove the trailing `<dx-icon name="chevron-right" .../>`. The whole row is already `bind:tap="goProfileEdit"`, so tapping the avatar/name continues to navigate to profile edit.

3. **Update each existing `<van-cell>` in the first cell-group:**
   - Add `center` attribute (vertically centers icon + title via `align-items: center`).
   - Add `custom-class="cell-icon"` to the `<dx-icon slot="icon" ...>` so the shared margin-right rule applies.

   Final shape per existing row:

   ```xml
   <van-cell title="我的团队" is-link center bind:click="goGroups">
     <dx-icon slot="icon" custom-class="cell-icon" name="users" size="20px" color="{{cellIconColor}}" />
   </van-cell>
   ```

   Apply identically to: 我的团队 (`users`), 推荐有礼 (`gift`), 兑换码 (`ticket`), 购买会员 (`crown`), 收藏的课程 (`star`).

4. **Append the new dark-mode row** at the end of the same cell-group, right after 收藏的课程:

   ```xml
   <van-cell title="深色模式" center bind:click="toggleTheme">
     <dx-icon
       slot="icon"
       custom-class="cell-icon"
       name="{{theme === 'dark' ? 'sun' : 'moon'}}"
       size="20px"
       color="{{cellIconColor}}"
     />
   </van-cell>
   ```

   Notes:
   - No `is-link` → no chevron, since this is a toggle, not navigation.
   - Sun/moon swap mirrors the prior top-bar convention (icon shows the action, not the current state).
   - No right-side label and no `<van-switch>` — keeps cell visually consistent with siblings and avoids adding a new dependency to `usingComponents` (which would require running 构建 npm in WeChat DevTools).

5. **Logout cell-group is untouched** — no icon there, no alignment issue.

### `pages/me/me.wxss`

1. **Lift content to status bar:**

   ```diff
    .page-container {
      min-height: 100vh;
      background: var(--bg-page);
   -  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
   +  padding-top: var(--status-bar-height, 20px);
      padding-bottom: 100rpx;
    }
   ```

2. **Delete the `.top-bar` rule** (no longer used).

3. **Add the shared cell-icon spacing rule:**

   ```css
   .cell-icon {
     margin-right: 12px;
   }
   ```

   Why this works: `dx-icon` declares `options: { addGlobalClass: true }`, and its WXML applies the `customClass` prop value as a class on the host view (`<view class="dx-icon-host {{customClass}}">`). Page-level styles in `me.wxss` therefore reach the host element. 12px sits between Vant's default `--padding-base` (4px, visually too tight against a 20px icon and 14px Chinese text) and overly wide; consistent with iOS Settings-style row spacing.

### `pages/me/me.ts`

Remove the now-dead `arrowColor` data field — the chevron was the only consumer.

```diff
   data: {
     theme: 'light' as 'light' | 'dark',
     primaryColor: '#0d9488',
-    arrowColor: '#9ca3af',
     cellIconColor: '#6b7280',
     ...
   },
   onShow() {
     ...
     this.setData({
       theme,
       primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
-      arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',
       cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
     });
     ...
   },
   toggleTheme() {
     ...
     this.setData({
       theme: next,
       primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
-      arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',
       cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
     })
     ...
   },
```

`toggleTheme`, `goProfileEdit`, `loadProfile`, and the rest of the methods are unchanged. Storage key `dx_theme`, `app.globalData.theme` propagation, and tab-bar theme propagation all remain.

### `pages/me/me.json`

Unchanged. `usingComponents` already lists `van-cell`, `van-cell-group`, `dx-icon` — no new dependency.

## Edge cases

- **Long nicknames + WeChat capsule:** With the lift to status bar and the chevron gone, the right edge of `profile-header` sits beneath the capsule (top-right `~87×32px`). The avatar (left) and stats (below the capsule height) are clear. A very long `profile-name` (`flex: 1`, no max-width) could in theory extend rightward and pass under the capsule. Realistically uncommon (typical 微信昵称 fits well within the safe width of ~187px). **Not addressed in this change.** If it becomes an issue we'd add `padding-right: 100px` to `.profile-header` or cap `profile-info` width.
- **Theme persistence:** Existing `wx.setStorageSync('dx_theme', next)` flow is preserved verbatim.
- **Tab-bar theme propagation:** The existing `tabBar.setData({ theme: next })` call inside `toggleTheme` is preserved.
- **Loading / error state:** No change — `<van-loading>` still shows while `loading` and a toast still appears on profile fetch failure.

## Out of scope

- Other pages (`home`, `leaderboard`, `notices`, `community`, `games`) keep their current top-bar / status-bar-spacer / capsule clearance patterns.
- No changes to `app.wxss` (theme variables remain global).
- No changes to dx-web or dx-api or `deploy/`.

## Verification

After implementation:

1. **TypeScript** — `npx tsc -p tsconfig.json --noEmit`. The known baseline noise from `Component({methods})` typing in v2.8.3 must remain unchanged in count and identity. The me page is a `Page({...})` (not affected by that bug); `this.setData` calls inside `onShow` and `toggleTheme` should typecheck cleanly. Goal: **zero new errors**.
2. **WXML structure** — eyeball the rendered file: tags balanced, no orphan `top-bar`, no orphan chevron `dx-icon`.
3. **Manual smoke test in WeChat Developer Tools** (user-driven; agent cannot drive DevTools):
   - Light mode: profile-header sits flush below status bar; capsule overlays empty right edge cleanly; cell-group rows show icon + ~12px gap + title on the same line.
   - Tap `深色模式` — theme flips, icon swaps from moon → sun, all other rows recolor.
   - Tap `深色模式` again — flips back.
   - Reload page — theme persists (storage roundtrip).
   - Tap any of the 5 navigation cells — each still routes to its target page.
   - Tap `profile-header` (avatar / name / badges) — opens profile edit.
   - Tap `退出登录` — confirm dialog appears; confirm logs out and reLaunches login.
   - Switch tabs and return — bottom tab-bar reflects current theme.

## Files

| File | Change |
|---|---|
| `dx-mini/miniprogram/pages/me/me.wxml` | delete top-bar; delete chevron; add `center` + `custom-class="cell-icon"` to 5 existing cells; append new 深色模式 cell |
| `dx-mini/miniprogram/pages/me/me.wxss` | reduce `padding-top`; delete `.top-bar`; add `.cell-icon` |
| `dx-mini/miniprogram/pages/me/me.ts` | drop `arrowColor` field and its 2 setData lines |
| `dx-mini/miniprogram/pages/me/me.json` | unchanged |
