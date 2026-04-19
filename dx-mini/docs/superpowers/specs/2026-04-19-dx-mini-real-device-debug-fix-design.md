# dx-mini: Fix `真机调试` "Cannot convert undefined or null to object" — Design

**Date:** 2026-04-19
**Scope:** dx-mini only. No changes to dx-api, dx-web, or deploy.
**Status:** Approved approach (#2 Config cleanup); implementation plan pending.

## 1. Problem

Clicking `真机调试` in WeChat DevTools (IDE 2.02.2604152) fails **instantly** with:

```
message：真机调试 Error: Cannot convert undefined or null to object
appid: wx6ffd2fe38aaf0c96
osType: darwin-arm64
```

No QR code is shown, so the build never reaches the phone. This rules out any code
running inside the Mini Program (login page, `app.ts` onLaunch, `config.ts` module-init,
`ws.ts` handshake, WXML bindings). The failure is in DevTools' Node-side project
pre-processing.

## 2. Root cause hypothesis

Three project-shape landmines are present; DevTools is likely dying on the first one
it hits. We fix all three because the error text carries no stack and diagnosing one
at a time would cost several debug-cycle round-trips.

| # | Landmine | Evidence |
|---|---|---|
| **A** | `"lazyCodeLoading": "requiredComponents"` in `miniprogram/app.json:40` combined with `"componentFramework": "glass-easel"` on line 39 | [Documented interaction](https://developers.weixin.qq.com/community/develop/doc/000e862d7946284ac76e8dfe45b800) that breaks real-device builds when `usingComponents` entries lack `placeholder:` declarations (all our Vant entries do). Commonly surfaces as white screens or config-lookup TypeErrors. |
| **B** | `"skylineRenderEnable": true` in `dx-mini/project.config.json:27` (and `project.private.config.json:9`) — but **no page.json in `miniprogram/pages/**` declares `"renderer": "skyline"`** | DevTools' Skyline packager walks the page list, expects per-page renderer configs, and `Object.keys()` on the missing entry throws exactly "Cannot convert undefined or null to object". Known pattern in [glass-easel #79](https://github.com/wechat-miniprogram/glass-easel/issues/79). |
| **C** | No `miniprogram/sitemap.json` exists, and `app.json` has no `sitemapLocation` override | Older DevTools (≤2.04) reported missing sitemap as a warning. Recent 2.02.2604x builds have been observed to hard-fail here on certain real-device paths. |

## 3. Non-goals

- **Not** changing `componentFramework: "glass-easel"` — glass-easel itself is fine and
  is needed for other features we've planned.
- **Not** touching runtime code (`.ts`, `.wxml`, `.wxss`) in this change. If the error
  persists after config cleanup, that's a separate diagnosis.
- **Not** upgrading Vant Weapp, WeChat DevTools, or `libVersion`.
- **Not** fixing the latent null-access in `learn.wxml` / `home.wxml` templates — those
  run at page-render time and are not the cause of a pre-QR DevTools error. Tracked as
  a separate follow-up.
- **Not** adding `placeholder:` declarations to every `usingComponents` entry to *keep*
  `lazyCodeLoading` working. Removing the opt-in is simpler and less error-prone.

## 4. Constraints

- Must not change any runtime behavior — all three edits are compile-time config.
- Must not introduce TypeScript, ESLint, or WXML lint errors.
- Reversible: every change is a single-file diff that can be reverted independently.
- Works offline in WeChat DevTools (no new network fetches at compile or boot).

## 5. Changes

### 5.1 Remove `lazyCodeLoading` from `miniprogram/app.json`

```diff
   "window": { ... },
   "style": "v2",
-  "componentFramework": "glass-easel",
-  "lazyCodeLoading": "requiredComponents"
+  "componentFramework": "glass-easel"
 }
```

Note: both lines must change in one edit — JSON doesn't tolerate a trailing
comma after `"glass-easel"`, so the comma must go when its neighbor does.

Default (no `lazyCodeLoading`) loads every component in a page on page-init — the
original WeChat behavior. Package size impact: negligible for dx-mini (single bundle,
~20 pages, Vant already tree-shaken by miniprogram-npm build step).

### 5.2 Align Skyline flag with reality in `project.config.json` and `project.private.config.json`

No page is actually Skyline-rendered (`grep -r '"renderer"' miniprogram/` → 0 matches),
so the flag is a config lie. Set both files to `false`:

```diff
 // project.config.json
-    "skylineRenderEnable": true,
+    "skylineRenderEnable": false,
```

```diff
 // project.private.config.json
-    "skylineRenderEnable": true,
+    "skylineRenderEnable": false,
```

No runtime effect — the flag only drives DevTools' Skyline preview button.

### 5.3 Add `miniprogram/sitemap.json`

Create with the WeChat-default permissive rule:

```json
{
  "desc": "关于本文件的更多信息，请参考文档 https://developers.weixin.qq.com/miniprogram/dev/reference/configuration/sitemap.html",
  "rules": [{
    "action": "allow",
    "page": "*"
  }]
}
```

No `sitemapLocation` override needed in `app.json` — DevTools auto-picks
`miniprogram/sitemap.json`. Allow-all is the intended state (the mini is public and
all pages are indexable).

## 6. Verification

After the three edits:

1. **Static checks** (must all pass unchanged):
   - `tsc --noEmit` from `dx-mini/` — no new TS errors.
   - DevTools → `详情 → 项目配置 → 勾选「ES6 转 ES5」` stays enabled; no new JSON-parse warnings.
   - `grep -r '"renderer"' dx-mini/miniprogram/` still returns 0 matches (sanity).
2. **DevTools simulator boot** (regression check):
   - Compile button → app opens in simulator → login page renders → login button tappable.
   - Theme toggle works. If logged in, can navigate to home/games/learn/me tabs.
3. **真机调试 primary check** (the fix):
   - Click `真机调试`. Expected: QR code appears within ~2 seconds, no error banner.
   - Scan with phone. App opens on phone; login page shows.
4. **Release/trial preview (secondary)**:
   - Click `预览` → QR code appears → scan → app loads. No error banner.

## 7. Risk & rollback

- **Risk of regressions:**
  - Removing `lazyCodeLoading` means every page's full component bundle loads on page-init
    instead of on-demand. On dx-mini this is sub-20 KB per page — imperceptible on any
    modern phone.
  - `skylineRenderEnable: false` hides the Skyline preview button in DevTools — only
    affects developer workflow, not built artifacts.
  - `sitemap.json` allow-all is WeChat's documented default; zero behavior change.
- **Rollback:** `git checkout -- dx-mini/miniprogram/app.json dx-mini/project.config.json dx-mini/project.private.config.json && rm dx-mini/miniprogram/sitemap.json`.
- **If 真机调试 still fails after this change:** the bug is outside the three landmines
  above. Next investigation steps are (a) check DevTools version for a known regression,
  (b) check `libVersion` compatibility between DevTools and the phone's WeChat client,
  (c) try clearing DevTools cache (`微信开发者工具 → 设置 → 通用设置 → 清除缓存`).

## 8. Out of scope (follow-ups)

- **latent WXML null-access** in `pages/learn/learn.wxml:7,9,12,14,17,19` and
  `pages/home/home.wxml:27,31` — render-time fragility, separate fix.
- **Logged-in user lands on `login` page** — `app.ts` onLaunch doesn't redirect
  authenticated users to `/pages/home/home`, so they see the login button even when
  a valid token is in storage. Separate UX bug.
