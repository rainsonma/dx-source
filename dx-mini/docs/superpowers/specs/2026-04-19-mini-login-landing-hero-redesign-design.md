# dx-mini login landing — hero gradient redesign

**Date:** 2026-04-19
**Scope:** `dx-mini/miniprogram/pages/login/` only
**Goal:** Bring the dx-mini login screen visually in line with the dx-web home hero by porting the same 5-stop pastel gradient onto the page, removing the WeChat-rendered title strip so the gradient flows edge-to-edge, and refreshing the copy.
**Non-goal:** Touching any other page, the global app shell, or the existing `wx.login` → `/api/auth/wechat-mini` flow.

---

## Background

`dx-mini/miniprogram/pages/login/login.{json,wxml,wxss,ts}` today renders the WeChat system nav bar (white, title `斗学`) above a flat `var(--bg-page)` body containing:

```
[ 斗学 ]                  ← 48px, brand teal text
[ 边玩边学，轻松记单词 ]    ← 14px secondary text

[ 使用微信登录 ]           ← Vant primary button
```

The dx-web home page (`dx-web/src/app/(web)/(home)/page.tsx`) paints a fixed 620 px top band with the Tailwind class `bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white` behind its hero. We want that same brand surface on the mini login screen, adapted to a phone canvas.

---

## Requirements

1. Page background uses the same five-stop top-to-bottom pastel gradient as dx-web's hero band, scaled to fill the entire phone screen including the area where the WeChat title bar used to be.
2. The WeChat-rendered title bar is removed; the gradient flows under the system status bar.
3. The body keeps the `斗学` brand mark (interpretation A from brainstorming).
4. Tagline copy changes to two lines mirroring dx-web:
   - Line 1: `多种学习模式 · AI 定制内容 · 和朋友一起闯关`
   - Line 2: `每天 10 分钟，英语悄悄流利了`
5. Login button copy changes from `使用微信登录` to `微信一键登录`.
6. Layout adapts to phone safe-areas: the WeChat capsule menu must not overlap content; the home indicator must not overlap the button.
7. Single light theme. The page no longer reacts to `app.globalData.theme`.
8. A subtle one-shot fade-in/slide-up reveal plays on first paint.
9. The existing login flow (`wx.login` → POST `/api/auth/wechat-mini` → store token+userId → connect WS → `wx.reLaunch('/pages/home/home')`) is preserved bit-identical.
10. No new lint errors. No `?.` or `??` in TS or WXML. No `console.log`. No new external assets.

---

## Design

### Page chrome

`login.json` becomes:

```json
{
  "navigationStyle": "custom",
  "navigationBarTextStyle": "black",
  "usingComponents": {
    "van-button": "@vant/weapp/button/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- `navigationStyle: "custom"` makes WeChat skip drawing the title bar; our content fills the screen including the status-bar region.
- `navigationBarTextStyle: "black"` tells WeChat to draw the right-edge capsule menu (which is mandatory in custom mode and cannot be hidden) with dark icons. The top of our gradient (`#ccfbf1`, teal-100) is bright enough that dark icons hit AA contrast comfortably.
- `navigationBarTitleText` is dropped — it's irrelevant in custom mode.
- `navigationBarBackgroundColor` is dropped — same reason.
- `van-config-provider` is dropped from `usingComponents` because the page becomes single-theme.

### Background layer

The gradient is painted as the page's first absolutely-positioned child so it covers everything from the very top edge of the screen to the bottom safe-area:

```css
.login-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: linear-gradient(
    to bottom,
    #ccfbf1 0%,
    #dbeafe 25%,
    #ede9fe 50%,
    #fce7f3 75%,
    #ffffff 100%
  );
}
```

Hex values match Tailwind v3 (and Tailwind v4 oklch defaults to within delta-E ~1) for `teal-100 / blue-100 / violet-100 / pink-100 / white`.

A single soft radial glow accent sits behind the brand mark to give the eye a focal point — same `rgba(94, 234, 212, 0.35)` accent the dx-web hero uses behind its game demo:

```css
.login-bg::after {
  content: '';
  position: absolute;
  top: 36%;
  left: 50%;
  width: 600rpx;
  height: 600rpx;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, rgba(94, 234, 212, 0.35), transparent 70%);
  pointer-events: none;
}
```

### Content layer

```html
<view class="login-page">
  <view class="login-bg" />
  <view class="login-content" style="padding-top: {{statusBarHeight}}px;">
    <view class="capsule-spacer" />
    <view class="spacer-top" />
    <view class="brand">
      <text class="brand-logo">斗学</text>
      <view class="brand-tagline">
        <text class="tagline-line">多种学习模式 · AI 定制内容 · 和朋友一起闯关</text>
        <text class="tagline-line">每天 10 分钟，英语悄悄流利了</text>
      </view>
    </view>
    <view class="spacer-bottom" />
    <view class="cta-wrap">
      <van-button
        type="primary"
        block
        round
        loading="{{loading}}"
        bind:click="handleLogin"
        custom-style="height:96rpx;font-size:32rpx;font-weight:600;"
      >微信一键登录</van-button>
    </view>
  </view>
</view>
```

- The wrapping `view.login-page` is `position: relative; min-height: 100vh; overflow: hidden`.
- `view.login-content` is `position: relative; z-index: 1; min-height: 100vh; display: flex; flex-direction: column; align-items: center; padding-bottom: env(safe-area-inset-bottom)`.
- `view.capsule-spacer` is a fixed 88 rpx block reserving room below the status bar so the brand never collides with the WeChat capsule.
- `view.spacer-top { flex: 0.6 }` and `view.spacer-bottom { flex: 1 }` pull the brand slightly above center, giving the button visual gravity at the bottom — standard mobile landing rhythm.

### Typography & color

- `.brand-logo`: 110 rpx, `font-weight: 800`, `letter-spacing: 12rpx`, flat `color: #0d9488` (the project's `--primary` brand teal). No gradient text.
- `.tagline-line`: 28 rpx, `color: #475569` (slate-600), `line-height: 1.6`, `text-align: center`. Each line is its own `<text>` block stacked in a vertical flex container with 8 rpx gap, ensuring the line break renders consistently across iOS / Android / DevTools.
- Button: Vant primary, full-width via `block`, rounded via `round`, `96 rpx` tall, `32 rpx / 600` text. The horizontal pad lives on `.cta-wrap` (`padding: 0 60rpx 80rpx`).

### Motion

One CSS-only first-paint reveal — no external library, no Skyline-specific APIs:

```css
@keyframes hero-rise {
  from { opacity: 0; transform: translateY(16rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
.brand-logo    { animation: hero-rise 480ms ease-out both; }
.brand-tagline { animation: hero-rise 480ms ease-out 120ms both; }
.cta-wrap      { animation: hero-rise 480ms ease-out 240ms both; }
```

Three elements, staggered 120 ms each, total runtime 720 ms. `both` keeps the final state pinned. No infinite loops, no scroll triggers, no JS.

### State / TS contract — `login.ts`

The file keeps its current structure. Only changes:

- `data` becomes `{ loading: false, statusBarHeight: 20 }`. The `theme` field is removed.
- `onLoad` reads `wx.getWindowInfo()` once and `setData({ statusBarHeight })`. The function is guarded with a typeof check so a missing API on an ancient base library degrades gracefully to the `20` default rather than throwing.
- `handleLogin` is bit-identical to today: `wx.login` → `api.post<AuthResponse>('/api/auth/wechat-mini', { code })` → `setToken` / `setUserId` → write `app.globalData.userId` → `ws.connect` → `ws.subscribe('user::<id>')` → `wx.reLaunch('/pages/home/home')`. Failure paths still surface via `wx.showToast`.

```ts
import { api } from '../../utils/api'
import { setToken, setUserId } from '../../utils/auth'
import { ws } from '../../utils/ws'

interface AuthResponse {
  token: string
  user: { id: string }
}

Page({
  data: {
    loading: false,
    statusBarHeight: 20,
  },
  onLoad() {
    if (typeof wx.getWindowInfo === 'function') {
      const info = wx.getWindowInfo()
      this.setData({ statusBarHeight: info.statusBarHeight })
    }
  },
  handleLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })
    wx.login({
      success: (res) => {
        api.post<AuthResponse>('/api/auth/wechat-mini', { code: res.code })
          .then((data) => {
            setToken(data.token)
            setUserId(data.user.id)
            const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()
            app.globalData.userId = data.user.id
            ws.connect(data.token)
            ws.subscribe(`user::${data.user.id}`)
            wx.reLaunch({ url: '/pages/home/home' })
          })
          .catch((err: Error) => {
            wx.showToast({ title: err.message || '登录失败', icon: 'none' })
            this.setData({ loading: false })
          })
      },
      fail: () => {
        wx.showToast({ title: '获取登录凭证失败', icon: 'none' })
        this.setData({ loading: false })
      },
    })
  },
})
```

---

## Files touched

| File | Change |
|---|---|
| `dx-mini/miniprogram/pages/login/login.json` | Add `navigationStyle:"custom"`, set `navigationBarTextStyle:"black"`, drop `navigationBarTitleText` and `navigationBarBackgroundColor`, drop `van-config-provider` from `usingComponents`. |
| `dx-mini/miniprogram/pages/login/login.wxml` | Restructure: drop `van-config-provider` wrapper, add `.login-bg` layer, brand block (logo + two-line tagline), CTA block. Remove the `{{theme}}` class binding. Bind dynamic `padding-top` to `statusBarHeight`. |
| `dx-mini/miniprogram/pages/login/login.wxss` | Replace existing rules with the new gradient + radial glow + flex layout + typography + keyframes. |
| `dx-mini/miniprogram/pages/login/login.ts` | Replace `theme` data with `statusBarHeight`; replace `onLoad` body with the `wx.getWindowInfo` read; `handleLogin` unchanged. |

**Not touched:** `app.json`, `app.wxss`, `app.ts`, any other page, `project.config.json`, `tsconfig.json`, the iconfont build script, the `typings/` shims, or any file outside `pages/login/`. No new image, font, or CSS asset is shipped.

---

## Risks & mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| `background-clip: text` would have failed on older Skyline builds — moot now that the brand mark uses flat `color: #0d9488`. | n/a | Decision already taken. |
| `wx.getWindowInfo()` missing on very old base libraries (< 2.20.1, vanishingly rare in 2026). | very low | `typeof` guard falls back to `statusBarHeight: 20`. |
| Capsule menu icon contrast. | low | Status-bar text color set to `"black"` via `navigationBarTextStyle`; verified against gradient top stop `#ccfbf1` for AA contrast against dark icons. |
| `navigationStyle:"custom"` is per-page in WeChat, but a global `app.json window.navigationStyle` could override. | nil | We checked: `app.json window` has no `navigationStyle`, so the per-page setting is authoritative. |
| Cross-page regression. | nil | All changes are scoped to `pages/login/`. The post-login `wx.reLaunch('/pages/home/home')` lands on the existing home page which keeps its system nav bar. |
| Loss of dark-mode rendering on this page. | accepted | Brand decision — a bright pastel gradient is the brand identity; the web has no dark variant either. The `theme` data field on the home page and elsewhere is unaffected. |

---

## Verification (acceptance criteria)

The redesign is correct when, in WeChat DevTools (iPhone 14 + Android Pixel 6 simulators):

1. Opening `pages/login/login` shows no WeChat title bar; the gradient flows under the system status bar.
2. The capsule menu in the top right is visible with dark icons and does not overlap the `斗学` brand mark.
3. The brand mark renders as `斗学` in flat `#0d9488` brand teal at ~110 rpx.
4. The tagline renders on exactly two centered lines with the canonical text.
5. The CTA reads `微信一键登录` and is full-width with a small horizontal margin.
6. On first paint, the brand mark, tagline, and button rise+fade in over ~720 ms total (staggered 120 ms each).
7. Tapping the CTA still produces the same network call (`POST /api/auth/wechat-mini`) and on success reLaunches to `/pages/home/home`.
8. Loading state still spins inside the button and re-enables on error.
9. `tsc --noEmit` (the dx-mini tsconfig pass) reports no new errors.
10. No `?.` / `??` operators are introduced in TS or WXML. No `console.log`. No new files outside the four listed above.

Manual test matrix:
- iPhone 14 (390×844) — base layout target.
- iPhone SE (320×568) — verify the spacers compress without overlapping content.
- Pixel 6 (412×915) — verify the bottom safe-area inset behaves on Android.
- Light system theme + dark system theme — verify the page renders identically in both (page is single-theme).
- Loading state — tap button, verify spinner shows in-button and toast surfaces on simulated error.
- Network failure — kill dx-api, tap button, verify toast appears and button re-enables.
