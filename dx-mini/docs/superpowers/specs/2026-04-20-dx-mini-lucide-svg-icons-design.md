# dx-mini: Lucide SVG Icon System — Design

**Date:** 2026-04-20
**Scope:** `dx-mini` only. No changes to `dx-api`, `dx-web`, or `deploy/`.
**Supersedes:** [2026-04-19 Icon System Migration](./2026-04-19-dx-mini-icon-system-and-home-fixes-design.md) — replaces the iconfont approach with direct SVG.
**Status:** Approved design; implementation plan pending.

## 1. Goal

Replace the current iconfont-based `<dx-icon>` implementation with a component that renders Lucide SVGs directly, while keeping the public API identical so no page, component, or `custom-tab-bar` call site needs to change.

## 2. Non-goals

- Changing the set of icons in use (the current 20 stay; more can be added via the inventory list).
- Modifying any page's icon names, sizes, colors, or click bindings.
- Removing `@vant/weapp` — other Vant components (`van-cell`, `van-button`, etc.) stay.
- Changes to `dx-web`, `dx-api`, or `deploy/`.

## 3. Constraints

- WXML cannot render raw `<svg>` elements; SVG must be delivered via `<image>` `src`.
- WeChat `<image>` accepts `data:image/svg+xml;utf8,…` and `data:image/svg+xml;base64,…`.
- TypeScript strict mode must continue to pass (`npx tsc --noEmit`); the existing `Component({methods})` `this`-typing exemption may remain but no new `tsc` errors are allowed.
- Public API of `<dx-icon>` is frozen — existing call sites must keep rendering with the same markup.
- No `?.`/`??` in dx-mini TS or WXML (per project memory — dev toolchain rejects them).
- Real-device testing goes through 预览 + 小程序助手, not 真机调试 (per project memory — DevTools 2.02.2604152 on macOS is broken).

## 4. Architecture

### 4.1 Component layout

```
miniprogram/components/dx-icon/
├── index.ts       properties + property observer computing src and hostStyle
├── index.wxml     <view> host with a single <image> child, no <van-icon>
├── index.wxss     inline-block host styling, line-height:1, vertical-align:middle
├── index.json     { "component": true } — no usingComponents
└── icons.ts       AUTO-GENERATED — { [iconName]: rawSvgString }
```

`icons.ts` is the only regenerated file; the other four are hand-maintained.

### 4.2 Public API (frozen — identical to current)

| Prop / Event   | Type   | Default  | Behavior |
|----------------|--------|----------|----------|
| `name`         | String | `''`     | Logical icon name (e.g. `"moon"`, `"chevron-right"`). |
| `size`         | String | `''`     | Dimension like `"22px"`. A bare numeric string gets `"px"` appended. |
| `color`        | String | `''`     | CSS color substituted for `currentColor` in the SVG. Empty → `#000`. |
| `strokeWidth`  | String | `'1.25'` | Numeric stroke width substituted for Lucide's default `stroke-width="2"`. |
| `customStyle`  | String | `''`     | Appended to host inline style. |
| `customClass`  | String | `''`     | Extra class on the host. |
| `bind:click`   | Event  | —        | Host `bind:tap` re-emitted via `triggerEvent('click', detail)`. |

### 4.3 Render

```wxml
<view class="dx-icon-host {{customClass}}" style="{{hostStyle}}" bind:tap="onClick">
  <image src="{{src}}" style="width:100%;height:100%" mode="aspectFit" />
</view>
```

```wxss
.dx-icon-host {
  display: inline-block;
  vertical-align: middle;
  line-height: 1;
}
```

### 4.4 Observer logic (in `index.ts`)

On any change to `name`, `size`, `color`, `strokeWidth`, or `customStyle`:

1. `normalizedSize = /^\d+(\.\d+)?$/.test(size) ? size + 'px' : size`
2. `hostStyle = 'width:' + normalizedSize + ';height:' + normalizedSize + ';' + customStyle`
3. `rawSvg = icons[name] || ''` — empty falls through to an empty `<image src>`, which renders nothing.
4. `svg = rawSvg.replace(/currentColor/g, color || '#000').replace(/stroke-width="2"/g, 'stroke-width="' + strokeWidth + '"')`
5. `src = 'data:image/svg+xml;utf8,' + encodeURIComponent(svg)`

All computed fields live on component `data` and are set via a single `this.setData({ src, hostStyle })` per change batch.

### 4.5 Why this shape

- `<image>` is the only portable way to deliver arbitrary SVG in WeChat Mini Program WXML.
- Runtime color injection (rather than CSS mask) means the real SVG, with its real stroke, is what the device draws — matches the user's intent of "use Lucide SVG icons directly."
- Keeping the wrapper and API identical avoids touching the 32 call sites.
- `strokeWidth` as a prop preserves per-instance override while pinning a project default of `1.25`.

## 5. Build script — `scripts/build-icons.mjs`

Replaces `scripts/build-iconfont.mjs`. Run manually: `npm run build:icons`.

### 5.1 Inputs

- `dx-mini/node_modules/lucide-static/icons/{lucideFilename}.svg` — source SVGs, pinned via `lucide-static` in `devDependencies`.
- Canonical `ICONS` array inside the script — same `[logicalName, lucideFilename]` shape as the current script, preserving logical-name aliases like `['help-circle', 'circle-help']` and `['home', 'house']`.

### 5.2 Process

1. For each `[logicalName, lucideFilename]` entry, read `node_modules/lucide-static/icons/${lucideFilename}.svg`. Fail with a clear message if missing (points at `lucide-static` version).
2. Keep the SVG as raw UTF-8. No stroke-to-fill transform, no font generation.
3. **Assert the runtime substitution tokens exist** in each SVG: the literal string `currentColor` must appear at least once, and `stroke-width="2"` must appear exactly once. If either assertion fails, abort with a message naming the logical icon — this catches upstream Lucide format drift before it reaches a device.
4. Statically scan `miniprogram/**/*.wxml` for literal name attributes matching the regex `<dx-icon[^>]*\sname="([a-z0-9-]+)"` (dynamic `{{…}}` bindings don't match and are intentionally skipped — those are covered by the curated `ICONS` list). Fail the build if any matched literal name is absent from `ICONS`, citing file + line.
5. Write `miniprogram/components/dx-icon/icons.ts`:

   ```ts
   // Auto-generated by scripts/build-icons.mjs — do not edit.
   export const icons: Record<string, string> = {
     moon: '<svg xmlns="…" …>…</svg>',
     // … one entry per ICONS row, in declaration order …
   }
   ```

6. Emit a one-line summary: `Wrote N icons to miniprogram/components/dx-icon/icons.ts.`

### 5.3 Outputs

Only `miniprogram/components/dx-icon/icons.ts`. No `app.wxss` patching; no font files; no `codepoints.json`.

### 5.4 Initial inventory (20 icons — same as today)

```
moon, sun, search, bell, chevron-right, chevron-left, star, book-open,
check, help-circle (→ circle-help), clock, crown, users, gift, ticket,
copy, home (→ house), trending-up, notebook-text, user
```

## 6. Dependency changes

| Location | Remove | Keep / add |
|---|---|---|
| `dx-mini/package.json` devDeps | `fantasticon`, `oslllo-svg-fixer` | `lucide-static`, `miniprogram-api-typings` |
| `dx-mini/package.json` scripts | `"build:iconfont"` | `"build:icons": "node scripts/build-icons.mjs"` |
| `dx-mini/miniprogram/package.json` | — | `@vant/weapp` unchanged |
| `components/dx-icon/index.json` | `usingComponents.van-icon` | `{ "component": true }` only |

## 7. Files deleted

- `miniprogram/assets/fonts/dx-iconfont.woff2`
- `miniprogram/assets/fonts/dx-iconfont.codepoints.json`
- `miniprogram/assets/fonts/` directory (after above, if empty)
- `scripts/build-iconfont.mjs`
- `/* === dx-iconfont … === */` block in `miniprogram/app.wxss` (the 37-line block only — surrounding CSS variables and `page` rules stay)

## 8. Migration sequence

Ordered so the app stays buildable at every checkpoint:

1. Write `scripts/build-icons.mjs`.
2. Update `dx-mini/package.json` — remove `fantasticon` + `oslllo-svg-fixer`, rename script, run `npm install`.
3. Run `npm run build:icons` → produces `icons.ts`.
4. Rewrite `dx-icon/index.ts`, `index.wxml`, `index.wxss`, `index.json`. New renderer is live after this step.
5. Delete the iconfont block from `app.wxss`.
6. Delete `miniprogram/assets/fonts/*`.
7. Delete `scripts/build-iconfont.mjs`.
8. DevTools smoke test (see §9).
9. Commit using `(mini)` scope — one or two commits (swap + cleanup).

## 9. Verification

### 9.1 Static checks (must all be green)

- `cd dx-mini && npx tsc --noEmit` — zero new errors.
- `grep -rn "dx-iconfont" dx-mini/` — no matches.
- `grep -rn "@vant/weapp/icon" dx-mini/miniprogram/components/dx-icon/` — no matches.
- `grep -rn "<van-icon" dx-mini/miniprogram/pages dx-mini/miniprogram/custom-tab-bar` — no matches.
- `grep -rn "build-iconfont" dx-mini/` — no matches.

### 9.2 Visual / runtime checks in WeChat DevTools

For every page below, confirm icons render, colors match the previous build, and tap handlers still fire:

- `pages/home` — search, theme toggle (moon ↔ sun), bell → `goNotices`.
- `pages/learn` — check, chevron-right, help-circle, clock.
- `pages/me` — chevron-right, crown, plus `<van-cell slot="icon">`: bell, users, gift, ticket, crown.
- `pages/me/groups`, `pages/me/invite`, `pages/me/groups-detail` — chevron-right, copy.
- `pages/games/favorites` — chevron-right, star.
- `pages/games/detail` — book-open, star toggle, chevron-right.
- `pages/games` — star, book-open.
- `pages/games/play` — chevron-left, check, help-circle, clock.
- `custom-tab-bar` — all 5 tabs (home, book-open, trending-up, notebook-text, user), active vs inactive color.

### 9.3 Real-device check

Preview + 小程序助手 smoke test on a phone — 真机调试 is broken per project memory.

## 10. Rollout

Single branch, two logical commits:

1. `feat(mini): replace iconfont with inline Lucide SVG in <dx-icon>` — new build script, new component internals, regenerated `icons.ts`.
2. `chore(mini): remove dx-iconfont assets and tooling` — delete fonts, old script, app.wxss block, old devDeps.

No API, WS, or shared-type changes — dx-api and dx-web are unaffected.

## 11. Memory & CLAUDE.md updates (after merge)

- New feedback memory `feedback_dx_mini_icon_strategy.md`: dx-mini always uses Lucide SVG icons via `<dx-icon>`; color injected at runtime via `currentColor` replacement; default stroke width 1.25; inventory curated in `scripts/build-icons.mjs` ICONS array.
- Edit the dx-mini Conventions → Icons line in `CLAUDE.md` from "Icons — Lucide via the `<dx-icon>` wrapper. NEVER import `<van-icon>` directly in pages; add glyphs to `scripts/build-iconfont.mjs` ICONS array and re-run `npm run build:iconfont`." to: "Icons — Lucide SVG via `<dx-icon>`; add glyphs to `scripts/build-icons.mjs` ICONS array and re-run `npm run build:icons`. Color via the `color` prop, default stroke width 1.25."

## 12. Risks

- **Layout shift at icon sites.** Old `<van-icon>` baseline-aligned as text; new host is inline-block. The `vertical-align: middle; line-height: 1` default covers most cases, but a 1–2 px shift in rare spots is possible. Caught in §9.2.
- **`<image>` SVG data URI support.** WeChat base library 3.15.1 (project's pinned baseline) supports both `utf8` and `base64` SVG data URIs; the `utf8` form is used for simpler runtime substitution. If an older device surfaces an issue, swap to base64 at the observer level — isolated change.
- **SVG pattern assumptions.** The observer assumes Lucide SVGs contain literal `currentColor` and `stroke-width="2"`. Both assumptions hold for `lucide-static@0.460`; §5.2 step 3 asserts these tokens in `build-icons.mjs` so upstream drift fails at build time instead of on a device.
