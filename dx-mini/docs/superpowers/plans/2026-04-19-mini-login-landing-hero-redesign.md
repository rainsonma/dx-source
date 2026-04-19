# dx-mini: Login landing hero-gradient redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the dx-web hero's 5-stop pastel gradient onto the dx-mini login page, remove the WeChat title bar so the gradient flows edge-to-edge under the system status bar, and refresh the copy (tagline + CTA label) to match dx-web.

**Architecture:** Four-file edit entirely inside `dx-mini/miniprogram/pages/login/`. One JSON config flip (activate custom navigation), one TS refactor (drop the dark-mode read, capture `statusBarHeight` via `wx.getWindowInfo`), a full WXSS rewrite (gradient + layout + typography + first-paint motion), and a matching WXML rewrite. No other page, global file, or new asset is touched. Single batched commit at the end.

**Tech Stack:** WeChat Mini Program (native + TypeScript), glass-easel + Skyline render, Vant Weapp 1.11.x, WXSS.

**Spec:** `dx-mini/docs/superpowers/specs/2026-04-19-mini-login-landing-hero-redesign-design.md`

**Working directory for all commands:** `/Users/rainsen/Programs/Projects/douxue/dx-source`

---

## File Structure

| File | Change | Responsibility |
|---|---|---|
| `dx-mini/miniprogram/pages/login/login.json` | Modify (full rewrite) | Per-page config — activate `navigationStyle:"custom"`, set capsule icon color, drop dark-mode Vant config-provider |
| `dx-mini/miniprogram/pages/login/login.ts` | Modify (full rewrite) | Page state — capture `statusBarHeight` on load, drop the `theme` binding, keep `handleLogin` bit-identical |
| `dx-mini/miniprogram/pages/login/login.wxss` | Modify (full rewrite) | Visual style — linear + radial gradient layers, flex layout, typography, first-paint keyframes |
| `dx-mini/miniprogram/pages/login/login.wxml` | Modify (full rewrite) | DOM — new brand / tagline / CTA layout with dynamic `padding-top: {{statusBarHeight}}px` |

**Order of edits:** `login.json` → `login.ts` → `login.wxss` → `login.wxml`. The four changes ship together in Task 7's commit; any intermediate state (after Task N but before Task 7) is a temporarily broken page, so do not launch the simulator between tasks.

---

## Task 1: Rewrite `login.json` for custom navigation

**Files:**
- Modify: `dx-mini/miniprogram/pages/login/login.json`

**Why this task:** `navigationStyle:"custom"` is the single switch that removes the white WeChat title bar. `navigationBarTextStyle:"black"` tells WeChat to render the mandatory right-edge capsule menu with dark icons, which is what we need for AA contrast against the gradient's top stop (`#ccfbf1`, teal-100). We also drop `van-config-provider` from `usingComponents` because the page becomes single-theme (`theme` data field is removed in Task 2).

- [ ] **Step 1: Verify current state**

Run:
```bash
cat dx-mini/miniprogram/pages/login/login.json
```

Expected output (exactly):
```json
{
  "navigationBarTitleText": "斗学",
  "navigationBarBackgroundColor": "#ffffff",
  "usingComponents": {
    "van-button": "@vant/weapp/button/index",
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

If the output differs, STOP — another change has been made and this plan needs an update.

- [ ] **Step 2: Write the new file**

Use the Write tool to replace the entire contents of `dx-mini/miniprogram/pages/login/login.json` with:

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

Note: `navigationBarTitleText` and `navigationBarBackgroundColor` are dropped because they are no-ops in custom mode. `van-config-provider` is dropped because the WXML no longer uses it (Task 4). `van-loading` stays: Vant Weapp's `button` loading prop does not use a separate `<van-loading>` component internally, but keeping the entry is a zero-risk preservation of the pre-existing dependency surface.

- [ ] **Step 3: Verify the JSON is valid**

Run:
```bash
python3 -c "import json; json.load(open('dx-mini/miniprogram/pages/login/login.json')); print('OK')"
```
Expected output: `OK`

- [ ] **Step 4: Verify the key fields are present (and the dropped ones are absent)**

Run:
```bash
grep -c '"navigationStyle": "custom"' dx-mini/miniprogram/pages/login/login.json
```
Expected output: `1`

Run:
```bash
grep -c '"navigationBarTextStyle": "black"' dx-mini/miniprogram/pages/login/login.json
```
Expected output: `1`

Run:
```bash
grep -c 'van-config-provider' dx-mini/miniprogram/pages/login/login.json
```
Expected output: `0`

Run:
```bash
grep -c 'navigationBarTitleText' dx-mini/miniprogram/pages/login/login.json
```
Expected output: `0`

No commit yet — we batch all four file changes into one commit in Task 7.

---

## Task 2: Rewrite `login.ts` to capture `statusBarHeight` and drop `theme`

**Files:**
- Modify: `dx-mini/miniprogram/pages/login/login.ts`

**Why this task:** The new WXML binds `padding-top: {{statusBarHeight}}px` on the content column so the brand never sits under the system status bar. That value comes from `wx.getWindowInfo()` (typed locally in `dx-mini/typings/index.d.ts`). At the same time, the page is now single-theme by design, so we remove the `theme` data field and the `app.globalData.theme` read from `onLoad`. `handleLogin` stays bit-identical to preserve the auth flow.

- [ ] **Step 1: Verify current state**

Run:
```bash
cat dx-mini/miniprogram/pages/login/login.ts
```

Expected output (exactly):
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
    theme: 'light' as 'light' | 'dark',
  },
  onLoad() {
    const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
    this.setData({ theme: app.globalData.theme })
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

If the output differs, STOP — the file has been modified outside this plan.

- [ ] **Step 2: Write the new file**

Use the Write tool to replace the entire contents of `dx-mini/miniprogram/pages/login/login.ts` with:

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

Key diffs from the old file:
- `data.theme` → `data.statusBarHeight: 20` (a safe default for every iPhone since the 7).
- `onLoad` no longer reads `app.globalData.theme`. Instead it calls `wx.getWindowInfo()` once, guarded by `typeof === 'function'` for base libraries so ancient they predate the API (<2.20.1, effectively nonexistent in 2026 but free to guard against).
- `handleLogin` is **unchanged** character-for-character — including the `getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()` type cast, which still correctly describes the app-level globalData shape (other pages still use `theme`; this page just stops reading it).

- [ ] **Step 3: Verify no banned TS syntax crept in**

Run:
```bash
grep -nE '\?\?|\?\.' dx-mini/miniprogram/pages/login/login.ts || echo NONE
```
Expected output: `NONE`

Run:
```bash
grep -n 'console\.' dx-mini/miniprogram/pages/login/login.ts || echo NONE
```
Expected output: `NONE`

- [ ] **Step 4: Verify the handleLogin body is bit-identical to the original**

Run:
```bash
grep -c 'api.post<AuthResponse>' dx-mini/miniprogram/pages/login/login.ts
```
Expected output: `1`

Run:
```bash
grep -c "wx.reLaunch({ url: '/pages/home/home' })" dx-mini/miniprogram/pages/login/login.ts
```
Expected output: `1`

Run:
```bash
grep -c '获取登录凭证失败' dx-mini/miniprogram/pages/login/login.ts
```
Expected output: `1`

No commit yet.

---

## Task 3: Rewrite `login.wxss` with gradient, layout, and motion

**Files:**
- Modify: `dx-mini/miniprogram/pages/login/login.wxss`

**Why this task:** This is the bulk of the visual work. The old WXSS paints a flat `var(--bg-page)` background and a centered vertical stack. The new WXSS replaces that with: a full-bleed 5-stop linear gradient, a single soft radial glow behind the brand, a flex column that reserves status-bar + capsule space at the top and the iOS home indicator at the bottom, typography for the logo and two-line tagline, and three staggered first-paint keyframe animations.

- [ ] **Step 1: Verify current state**

Run:
```bash
cat dx-mini/miniprogram/pages/login/login.wxss
```

Expected output (exactly):
```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
}
.login-body {
  width: 100%;
  padding: 0 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 48px;
}
.logo-wrap {
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.logo-text {
  font-size: 48px;
  font-weight: 700;
  color: var(--primary);
  letter-spacing: 4px;
}
.logo-sub {
  font-size: 14px;
  color: var(--text-secondary);
}
```

If the output differs, STOP.

- [ ] **Step 2: Write the new file**

Use the Write tool to replace the entire contents of `dx-mini/miniprogram/pages/login/login.wxss` with:

```css
/* Override the app-level page background so our gradient layer is the only
   surface visible above the WeChat shell. */
page {
  background: transparent;
}

/* === Layout scaffolding === */

.login-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
}

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

.login-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-bottom: env(safe-area-inset-bottom);
  box-sizing: border-box;
}

.capsule-spacer {
  height: 88rpx;
  flex: 0 0 auto;
}

.spacer-top    { flex: 0.6 1 0; }
.spacer-bottom { flex: 1 1 0; }

/* === Brand === */

.brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32rpx;
}

.brand-logo {
  font-size: 110rpx;
  font-weight: 800;
  letter-spacing: 12rpx;
  color: #0d9488;
  animation: hero-rise 480ms ease-out both;
}

.brand-tagline {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  padding: 0 60rpx;
  animation: hero-rise 480ms ease-out 120ms both;
}

.tagline-line {
  font-size: 28rpx;
  line-height: 1.6;
  color: #475569;
  text-align: center;
}

/* === CTA === */

.cta-wrap {
  width: 100%;
  padding: 0 60rpx 80rpx;
  box-sizing: border-box;
  animation: hero-rise 480ms ease-out 240ms both;
}

/* === First-paint reveal === */

@keyframes hero-rise {
  from { opacity: 0; transform: translateY(16rpx); }
  to   { opacity: 1; transform: translateY(0);     }
}
```

Notes on the WXSS:
- `page { background: transparent; }` overrides the `background: var(--bg-page)` declaration from `app.wxss:22` for this page only. Without it, the app's `#f5f5f5` would peek through at the edges of `.login-page` on any screen where the gradient layer doesn't clip to the viewport perfectly (portrait vs. rotated, iOS bounce scroll, etc.).
- The `.capsule-spacer` is a fixed 88 rpx block that sits directly under the status-bar inset. WeChat's capsule menu measures roughly 64 rpx + padding; 88 rpx clears it with visible breathing room on every device model we care about.
- The three `animation` declarations use `both` fill mode so the elements never flicker at their "from" state once the animation completes, and so they're invisible until the keyframes begin.

- [ ] **Step 3: Verify no banned patterns**

Run:
```bash
grep -nE '\?\?|\?\.' dx-mini/miniprogram/pages/login/login.wxss || echo NONE
```
Expected output: `NONE`

- [ ] **Step 4: Verify the gradient stops and the brand color are present**

Run:
```bash
grep -c '#ccfbf1' dx-mini/miniprogram/pages/login/login.wxss
```
Expected output: `1`

Run:
```bash
grep -c '#0d9488' dx-mini/miniprogram/pages/login/login.wxss
```
Expected output: `1`

Run:
```bash
grep -c 'hero-rise' dx-mini/miniprogram/pages/login/login.wxss
```
Expected output: `4` (three `animation:` references + one `@keyframes` definition)

No commit yet.

---

## Task 4: Rewrite `login.wxml` to use the new layout

**Files:**
- Modify: `dx-mini/miniprogram/pages/login/login.wxml`

**Why this task:** The new DOM drops the `<van-config-provider>` theme wrapper, introduces the gradient layer, reserves the status-bar area via a dynamic inline `padding-top`, and splits the tagline into two `<text>` blocks inside a flex column so the line break renders identically on iOS, Android, and DevTools.

- [ ] **Step 1: Verify current state**

Run:
```bash
cat dx-mini/miniprogram/pages/login/login.wxml
```

Expected output (exactly):
```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <view class="login-body">
      <view class="logo-wrap">
        <text class="logo-text">斗学</text>
        <text class="logo-sub">边玩边学，轻松记单词</text>
      </view>
      <van-button
        type="primary"
        block
        round
        loading="{{loading}}"
        bind:click="handleLogin"
        custom-style="height:48px;font-size:16px;"
      >使用微信登录</van-button>
    </view>
  </view>
</van-config-provider>
```

If the output differs, STOP.

- [ ] **Step 2: Write the new file**

Use the Write tool to replace the entire contents of `dx-mini/miniprogram/pages/login/login.wxml` with:

```xml
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

Notes on the WXML:
- The dynamic `padding-top: {{statusBarHeight}}px` on `.login-content` comes from the `setData` call in `login.ts` Task 2 Step 2. The default `20` keeps content off the top edge even before `onLoad` resolves.
- The `<view class="login-bg" />` sits at `z-index: 0` inside `.login-page`; `.login-content` above it at `z-index: 1`. All taps flow to the content because `.login-bg` has `pointer-events: none`.
- The `.brand-tagline` renders the tagline as **two** centered lines, matching the dx-web hero layout. The user's brainstorm text concatenated them without a separator — that was a copy-paste artifact; the canonical split is `...和朋友一起闯关` / `每天 10 分钟，英语悄悄流利了`.
- The button's `custom-style` prop switches from `px` (which WeChat accepts but doesn't scale by device) to `rpx` so the height and font track screen density consistently with the rest of the new layout.

- [ ] **Step 3: Verify no banned patterns**

Run:
```bash
grep -nE '\?\?|\?\.' dx-mini/miniprogram/pages/login/login.wxml || echo NONE
```
Expected output: `NONE`

- [ ] **Step 4: Verify the copy changes landed and the old copy is gone**

Run:
```bash
grep -c '微信一键登录' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `1`

Run:
```bash
grep -c '使用微信登录' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `0`

Run:
```bash
grep -c '多种学习模式 · AI 定制内容 · 和朋友一起闯关' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `1`

Run:
```bash
grep -c '每天 10 分钟，英语悄悄流利了' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `1`

Run:
```bash
grep -c '边玩边学，轻松记单词' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `0`

Run:
```bash
grep -c '{{statusBarHeight}}' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `1`

Run:
```bash
grep -c 'van-config-provider' dx-mini/miniprogram/pages/login/login.wxml
```
Expected output: `0`

No commit yet.

---

## Task 5: Cross-file consistency + lint sweep

**Files:** All four changed files (read-only checks).

**Why this task:** Before asking the user to fire up DevTools, verify the four files reference each other correctly — the WXML classes exist in the WXSS, the WXML data bindings exist in the TS, and no banned language constructs leaked in anywhere.

- [ ] **Step 1: Check every class referenced in the WXML is defined in the WXSS**

Run:
```bash
for cls in login-page login-bg login-content capsule-spacer spacer-top spacer-bottom brand brand-logo brand-tagline tagline-line cta-wrap; do
  if ! grep -q "\.$cls" dx-mini/miniprogram/pages/login/login.wxss; then
    echo "MISSING: .$cls"
  fi
done
echo DONE
```
Expected output: `DONE` with no `MISSING:` lines.

- [ ] **Step 2: Check every `{{ }}` binding in the WXML maps to a `data` field in the TS**

Run:
```bash
grep -oE '{{[a-zA-Z]+}}' dx-mini/miniprogram/pages/login/login.wxml | sort -u
```
Expected output (exactly):
```
{{loading}}
{{statusBarHeight}}
```

Run:
```bash
grep -n 'loading:\|statusBarHeight:' dx-mini/miniprogram/pages/login/login.ts
```
Expected: both fields appear inside the `data` block.

- [ ] **Step 3: Sweep all four files for banned TS/WXML operators**

Run:
```bash
grep -nE '\?\?|\?\.' dx-mini/miniprogram/pages/login/*.ts dx-mini/miniprogram/pages/login/*.wxml dx-mini/miniprogram/pages/login/*.wxss || echo NONE
```
Expected output: `NONE`

- [ ] **Step 4: Sweep for `console.*`**

Run:
```bash
grep -n 'console\.' dx-mini/miniprogram/pages/login/*.ts dx-mini/miniprogram/pages/login/*.wxml || echo NONE
```
Expected output: `NONE`

- [ ] **Step 5: Confirm no files outside `pages/login/` were touched**

Run:
```bash
git status --short
```
Expected output (exactly, in any order):
```
 M dx-mini/miniprogram/pages/login/login.json
 M dx-mini/miniprogram/pages/login/login.ts
 M dx-mini/miniprogram/pages/login/login.wxml
 M dx-mini/miniprogram/pages/login/login.wxss
```

(The plan and spec are already committed before Task 1 begins; the only modifications should be the four files above.)

If any other file appears, STOP and investigate before continuing.

- [ ] **Step 6: Optional — run the TypeScript compiler if it's available on this machine**

TypeScript is not in `dx-mini/package.json`; WeChat DevTools compiles TS internally at load time. If you have a global `tsc` (e.g. from the `dx-web` workspace) and want a belt-and-braces type check, run:

```bash
cd dx-mini && npx --no-install tsc --noEmit 2>&1 | tail -20
```

Expected output: no errors (empty output, or a terminal prompt). If `npx --no-install` reports that `tsc` isn't installed, skip — Task 6's DevTools check is authoritative for runtime correctness.

No commit yet.

---

## Task 6: Manual DevTools verification

**Files:** None (this task opens WeChat DevTools and walks the acceptance criteria from the spec).

**Why this task:** There is no headless test harness for WeChat Mini Program UI. The spec's 10 acceptance criteria can only be validated by eyeballing the simulator. Drive through all of them before committing.

- [ ] **Step 1: Open the project in WeChat Developer Tools**

Open `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini` in WeChat DevTools. Wait for the compile to finish. The bottom status bar should read "编译成功" with no red errors.

- [ ] **Step 2: Navigate to `pages/login/login`**

In the DevTools left sidebar, click the 模拟器 pane. If the app auto-launches into `pages/home/home`, open the DevTools console and run:
```js
wx.reLaunch({ url: '/pages/login/login' })
```

- [ ] **Step 3: Walk the acceptance checklist**

Verify each of these against the simulator (iPhone 14, default device):

  - [ ] No WeChat title bar is visible. The gradient flows under the system status bar.
  - [ ] The WeChat capsule menu (right-edge pill) is drawn with **dark** icons, and the `斗学` brand mark sits visibly below it with clearance.
  - [ ] The brand mark renders as `斗学` in flat teal `#0d9488`, very large (~55 px on a 1x device), bold, with letter-spacing.
  - [ ] The tagline renders on **exactly two** centered lines: `多种学习模式 · AI 定制内容 · 和朋友一起闯关` then `每天 10 分钟，英语悄悄流利了`.
  - [ ] The button at the bottom reads `微信一键登录`, is full-width with ~30 px horizontal margin, rounded, teal-primary.
  - [ ] On first paint, the logo rises+fades in first, the tagline 120 ms later, the button 240 ms later. Total reveal ~720 ms.
  - [ ] Tap the CTA with a valid network reachable: the button enters loading state, the network tab shows `POST /api/auth/wechat-mini`, and on success the simulator reLaunches to `/pages/home/home`.
  - [ ] Toggle the system theme in DevTools 设置 → 通用 → 外观 to 深色 and navigate back to the login page. The gradient and copy render identically to light mode (the page is single-theme by design).
  - [ ] Tap the CTA with the API killed (stop dx-api, or switch off Wi-Fi): a toast surfaces ("登录失败" or the error message), and the button re-enables.

- [ ] **Step 4: Sanity-check one smaller and one larger device**

Still in DevTools, switch device to **iPhone SE** (320×568) and re-verify:
  - [ ] The brand mark, tagline, and button all fit on screen without overflow. The `spacer-top 0.6 / spacer-bottom 1.0` flex ratio may squeeze tighter but nothing should clip.

Switch device to **Pixel 6** (412×915) and re-verify:
  - [ ] The bottom safe-area inset behaves: the CTA sits above the Android navigation gesture bar with visible gap.

- [ ] **Step 5: Confirm no other page regressed**

From DevTools console:
```js
wx.reLaunch({ url: '/pages/home/home' })
```
The home page (first tab) still has its system navigation bar with title "斗学". Navigate through the other tabs (games / leaderboard / learn / me) and confirm nothing changed visually — `navigationStyle:"custom"` is per-page and should not leak.

If any check fails, STOP and fix before Task 7. Record what failed — the most common issues and their fixes are:

| Symptom | Likely cause | Fix |
|---|---|---|
| Brand mark is hidden under the capsule | `capsule-spacer` too short OR `statusBarHeight` not applied | Verify Task 2 Step 3's `{{statusBarHeight}}` result is `1`; bump `.capsule-spacer { height }` to 108 rpx |
| Gradient has a hard seam at the top | `page { background }` override missing | Re-verify Task 3 Step 2's first rule block |
| Tagline renders as one wrapping line instead of two | `.brand-tagline` missing `flex-direction: column` | Re-verify the `.brand-tagline` block in Task 3 |
| Button text is black/grey instead of white | `van-button type="primary"` prop lost | Re-verify the button in Task 4's WXML |
| First-paint animation never plays | Animation targets missing the `both` fill mode | Re-verify the three `animation:` lines end with `both` |

No commit yet.

---

## Task 7: Commit all four files together

**Files:** None edited in this task — it stages and commits the changes from Tasks 1–4.

- [ ] **Step 1: Review the staged diff one last time**

Run:
```bash
git diff dx-mini/miniprogram/pages/login/
```

Read the diff. Confirm:
- Exactly four files changed (`login.json`, `login.ts`, `login.wxml`, `login.wxss`).
- No stray whitespace or BOM additions.
- No accidental edits outside the four files (re-run `git status --short` if unsure).

- [ ] **Step 2: Stage the four source files**

Run:
```bash
git add \
  dx-mini/miniprogram/pages/login/login.json \
  dx-mini/miniprogram/pages/login/login.ts \
  dx-mini/miniprogram/pages/login/login.wxml \
  dx-mini/miniprogram/pages/login/login.wxss
```

- [ ] **Step 3: Commit**

Run:
```bash
git commit -m "$(cat <<'EOF'
feat(mini): redesign login landing with hero gradient and custom nav

Ports the dx-web home hero's 5-stop pastel gradient onto the login
page, flips navigationStyle to custom so the gradient flows edge-
to-edge under the status bar, and refreshes the copy to match the
web tagline and a shorter CTA label. Single-theme by design — the
page no longer reads app.globalData.theme. handleLogin is bit-
identical to preserve the wx.login -> /api/auth/wechat-mini flow.
EOF
)"
```

- [ ] **Step 4: Verify the commit landed**

Run:
```bash
git log --oneline -4
```
Expected: the newest line is the commit above; the three prior lines are the pre-existing history (`docs(mini): plan for login landing hero-gradient redesign`, `docs(mini): spec for login landing hero-gradient redesign`, `docs(mini): spec + plan for 真机调试 investigation`).

Run:
```bash
git show --stat HEAD
```
Expected: exactly 4 files in the stat line — the 4 source files under `dx-mini/miniprogram/pages/login/`.

Done. The login landing redesign is live on `main` locally. Do **not** push any feature branch — per repo policy, only `main` ships to remote, and it ships with a plain `git push origin main` once the user explicitly asks.
