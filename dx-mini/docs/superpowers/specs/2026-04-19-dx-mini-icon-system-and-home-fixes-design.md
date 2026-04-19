# dx-mini: Icon System Migration & Home Page Fixes — Design

**Date:** 2026-04-19
**Scope:** dx-mini (WeChat Mini Program) only. No changes to dx-api, dx-web, or deploy.
**Status:** Approved design; implementation plan pending.

## 1. Goals

Fix three user-reported defects and establish an outline-first icon system:

1. **Two "undefined" on home page.** Backend returns `Greeting { title, subtitle }` (emoji baked into `title`, e.g., `"早上好 👋"`); mini reads `Greeting { text, emoji }` and renders `{{greeting.emoji}} {{greeting.text}}` — both fields are absent, producing literal `"undefined undefined"`.
2. **Dark-mode toggle icon not rendering.** `sunny-o` / `moon-o` do not exist in Vant Weapp 1.11.x; the `<van-icon>` renders an empty glyph (tap target still works but nothing is shown).
3. **All icons must be outline.** Several pages use filled glyphs (`star`, `success`, and Vant's default-filled `bell` / `search` / `column`). Vant 1.11.x does not provide outline variants for many of these.

## 2. Non-goals

- Changing the backend `Greeting` shape — mini is aligned to backend, not the reverse.
- Modifying dx-web, dx-api, or deploy configuration.
- Refactoring, lint-config changes, or feature work beyond the files touched.
- Replacing WeChat platform toast icons (`wx.showToast({ icon: 'none' })`).

## 3. Constraints

- Vant Weapp 1.11.x is pinned (`@vant/weapp: ^1.11.0`); upgrading is out of scope.
- TypeScript strict mode is enforced (`strict`, `noImplicitAny`, `noUnusedLocals`, `strictNullChecks`).
- WeChat Mini Program constraints: no runtime font loading from untrusted origins; package size budget; Skyline rendering enabled.
- Every page must keep current icon position, size, color intent, and tap handler behavior.

## 4. Architecture

### 4.1 New component: `<dx-icon>`

Thin wrapper around `<van-icon>` that pins `class-prefix="dx"` and forwards `name`, `size`, `color`, `custom-style`, `custom-class`, and taps.

```
miniprogram/components/dx-icon/
├── index.ts        # Component({ properties: { name, size, color, customStyle, customClass } })
├── index.wxml      # <van-icon class-prefix="dx" name="{{name}}" size="{{size}}" color="{{color}}" ...>
├── index.wxss      # (empty — font lives in app.wxss)
└── index.json      # { "component": true, "usingComponents": { "van-icon": "@vant/weapp/icon/index" } }
```

The wrapper re-emits `bind:tap` as `bind:click` on its external API. Consumers use `<dx-icon name="moon" bind:click="toggleTheme" />`.

### 4.2 Iconfont: Lucide-based custom font

- **Source icons:** [Lucide](https://lucide.dev) outline SVGs, fetched once from `lucide-static` at build time (not a mini dependency).
- **Glyph set:** Exactly the 20 names enumerated in §5 — no more, no less.
- **Font generation:** Upload SVGs to iconfont.cn (or equivalent), assign codepoints `\e001`–`\e014`, download `woff2`.
- **Delivery:** Base64-encode the `woff2`, inline inside `app.wxss` via `@font-face`. No runtime font fetch. No `wx.loadFontFace` call.
- **Size budget:** ~4–8 KB base64 for 20 outline glyphs, well within WeChat's package budget.

```css
/* appended to app.wxss */
@font-face {
  font-family: 'dx-iconfont';
  src: url('data:font/woff2;base64,...') format('woff2');
  font-weight: normal; font-style: normal; font-display: block;
}
[class*='dx-icon-']::before {
  font-family: 'dx-iconfont';
  font-style: normal; font-weight: normal;
  speak: none; display: inline-block; text-decoration: none;
  -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
}
.dx-icon-moon::before          { content: '\e001'; }
.dx-icon-sun::before           { content: '\e002'; }
/* … 18 more lines, one per name in §5 … */
```

### 4.3 Why a wrapper around `van-icon` (vs. raw `<text>` or full replacement)

- Keeps ergonomic `size`/`color`/tap handling for free — Vant's icon component already translates `color` prop into inline `style` on the inner text and `size` into `font-size` with proper units.
- Single abstraction: future icon additions touch one component.
- Lucide-named API matches dx-web mental model (`Moon` / `Sun` / `Bell` in React ↔ `moon` / `sun` / `bell` in mini).
- Does not orphan Vant as a dependency — still used for cells, tabs, skeletons, dialogs, etc.

## 5. Icon Inventory (20 glyphs)

Canonical Lucide names, singular, kebab-case. No `-o` suffix — outline is implicit.

| Lucide name | Replaces (Vant) | Used on |
|---|---|---|
| `search` | `search` | home.wxml:6 |
| `moon` | `moon-o` (broken) | home.wxml:11 (light-mode state) |
| `sun` | `sunny-o` (broken) | home.wxml:11 (dark-mode state) |
| `bell` | `bell` (filled), cell `icon="bell"` | home.wxml:16, me.wxml:52 |
| `chevron-right` | `arrow` | detail.wxml:39, favorites.wxml:20, me.wxml:26, groups.wxml:17, learn.wxml:26/31/36 |
| `chevron-left` | `arrow-left` | play.wxml:9 |
| `star` | `star` (filled) + `star-o` | games.wxml:22, detail.wxml:12, favorites.wxml:23 |
| `book-open` | `column` | games.wxml:45, detail.wxml:9 |
| `check` | `success` | play.wxml:33, learn.wxml:24 |
| `help-circle` | `question-o` | play.wxml:37, learn.wxml:29 |
| `clock` | `clock-o` | play.wxml:41, learn.wxml:34 |
| `crown` | `vip-card-o`, cell `icon="vip-card-o"` | me.wxml:47, me.wxml:56 |
| `users` | cell `icon="friends-o"` | me.wxml:53 |
| `gift` | cell `icon="gift-o"` | me.wxml:54 |
| `ticket` | cell `icon="coupon-o"` | me.wxml:55 |
| `copy` | `copy-o` | groups-detail.wxml:8, invite.wxml:9 |
| `home` | `wap-home` / `wap-home-o` | custom-tab-bar (tab 0) |
| `trending-up` | `chart-trending-o` | custom-tab-bar (tab 2) |
| `notebook-text` | `records` | custom-tab-bar (tab 3) |
| `user` | `contact` | custom-tab-bar (tab 4) |

Note: `star` is a single glyph covering both former filled and outline uses — color expresses favorited state (amber) vs. default (gray).

## 6. Fix Specifications

### 6.1 Fix #1 — Home page undefineds

**`pages/home/home.ts`:**

Replace the Greeting interface and stop destructuring unknown fields:

```ts
interface Greeting { title: string; subtitle: string }
```

Data declaration remains `greeting: null as Greeting | null`. The assignment `greeting: dash.greeting` in `loadData()` is unchanged — shape now matches backend.

**`pages/home/home.wxml`:** replace line 23 with two stacked elements:

```wxml
<text class="greeting-title">{{greeting ? greeting.title : '你好！'}}</text>
<text wx:if="{{greeting}}" class="greeting-subtitle">{{greeting.subtitle}}</text>
```

**`pages/home/home.wxss`:** rename `.greeting` → `.greeting-title` (keep existing font-size 18px, weight 600, primary color, display block), remove its bottom margin; add:

```css
.greeting-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  display: block;
  margin-top: 4px;
  margin-bottom: 16px;
}
.greeting-title {
  margin-bottom: 4px; /* was 16px */
}
```

Result: title line ("早上好 👋") renders 18px bold in primary text color; subtitle ("继续你的学习之旅，今天也要加油！") renders 13px in secondary text color below. Matches dx-web hall behavior.

### 6.2 Fix #2 — Dark-mode toggle icon

**`pages/home/home.wxml` lines 10–15**, replace the broken van-icon:

```wxml
<dx-icon
  name="{{theme === 'dark' ? 'sun' : 'moon'}}"
  size="22px"
  color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
  bind:click="toggleTheme"
/>
```

### 6.3 Fix #3 — Outline-everywhere migration

For every WXML entry in §5's "Used on" column, replace `<van-icon name="{old}" ...>` with `<dx-icon name="{new}" ...>`. All other props (`size`, `color`, `bind:tap`) carry over verbatim, with `bind:tap` renamed to `bind:click` on the new element.

**Special cases:**

- **`detail.wxml:12` favorite toggle** — unify to a single outline star, color-toggled:
  ```wxml
  <dx-icon name="star" size="24px" color="{{favorited ? '#f59e0b' : '#9ca3af'}}" />
  ```
- **`me.wxml:52–56` cell icons** — replace each `<van-cell icon="..." ...>` with a slot:
  ```wxml
  <van-cell title="公告通知" is-link bind:click="goNotices">
    <dx-icon slot="icon" name="bell" size="20px" color="{{cellIconColor}}" />
  </van-cell>
  ```
  Add `cellIconColor` to `me.ts` data (follow existing `arrowColor` / `primaryColor` pattern: theme-aware). Sizes: 20px to match Vant cell line-height.
- **`custom-tab-bar/index.ts`** — drop `activeIcon` from `TabItem` and each tab entry. Active/inactive differentiation is color-only (teal active, gray inactive), as Vant's existing `color` binding already distinguishes.
- **`custom-tab-bar/index.wxml` line 9** — change `<van-icon>` to `<dx-icon>`; keep `{{item.icon}}` binding and existing color expression.

### 6.4 `.json` registration

Add `"dx-icon": "/components/dx-icon/index"` to `usingComponents` in every page/component that currently declares `van-icon`:

- `pages/home/home.json`
- `pages/games/games.json`
- `pages/games/detail/detail.json`
- `pages/games/play/play.json`
- `pages/games/favorites/favorites.json`
- `pages/learn/learn.json`
- `pages/me/me.json`
- `pages/me/groups/groups.json`
- `pages/me/groups-detail/groups-detail.json`
- `pages/me/invite/invite.json`
- `custom-tab-bar/index.json`

Remove the now-unused `"van-icon"` entry from any page whose WXML no longer references it directly (it remains a dependency of `<dx-icon>` internally; the page-level `usingComponents` entry is only needed when WXML uses it).

## 7. Data-flow & Behavior

- **Greeting fetch:** `GET /api/hall/dashboard` returns the current shape; no backend changes. `loadData()` assigns `greeting: dash.greeting` exactly as today.
- **Theme toggle:** unchanged handler (`toggleTheme`). Only the icon-source component changes.
- **Icon rendering:** `<dx-icon>` → `<van-icon class-prefix="dx">` → inner `<text class="dx-icon dx-icon-{name}">` whose glyph comes from the inlined font-face; `color`/`size` applied as before.
- **Tap propagation:** `bind:click` on `<dx-icon>` → `bind:tap` internally → Vant icon emits tap → wrapper re-emits as `click`. Handler signatures on pages are unchanged (still plain `()` handlers).

## 8. Error handling & edge cases

- **Backend returns `greeting: null`** — `greeting ? greeting.title : '你好！'` branch renders fallback; subtitle `wx:if="{{greeting}}"` stays hidden. No undefineds possible.
- **Font fails to load** — glyphs render as empty. Mitigation: `font-display: block` prevents FOIT flash; if the font truly fails to load (shouldn't, it's embedded), pages still function; tap targets unchanged.
- **Missing icon name** — `<dx-icon name="does-not-exist">` renders an empty glyph. Guarded by the fixed inventory in §5; any future additions require updating `app.wxss`. Add a comment in `app.wxss` documenting the inventory.
- **van-cell slot size** — Vant cell's default icon size is ~16–20px with line-height alignment. Explicit `size="20px"` on the slotted `<dx-icon>` matches visually.

## 9. Testing & verification

### Automated gates

1. `tsc --noEmit` in `miniprogram/` — zero errors.
2. `npm install && npm run build` in `miniprogram/` (Vant npm build) — success.
3. WeChat DevTools compile (GUI or `cli -o`) — no missing `usingComponents`, no malformed WXML.

### Manual smoke tests (user-driven — simulator not available from Claude's environment)

Per page, in both light and dark theme, against a backend returning a populated dashboard:

| Page | Acceptance |
|---|---|
| Home | No "undefined" anywhere. Title + subtitle render from backend. Theme-toggle icon visible in both modes (moon in light, sun in dark); tap swaps icon, theme, and tab bar. Search + bell outline. |
| Games | Outline star filter; `book-open` empty state. |
| Detail | `book-open` cover placeholder; outline star (amber favorited, gray not); outline chevron rows. |
| Play | `chevron-left` back; `check` / `help-circle` / `clock` stat chips. |
| Favorites | Outline chevron + outline star. |
| Learn | Three stat rows: `check` / `help-circle` / `clock`; outline chevrons. |
| Me | `chevron-right` / `crown`, five cell rows with slotted `bell` / `users` / `gift` / `ticket` / `crown`. |
| Groups / Groups-detail / Invite | `chevron-right`, `copy` render; copy handlers still fire toast. |
| Tab bar | Single outline glyph per tab; active = teal, inactive = gray; switching tabs behaves as today. |
| Login, leaderboard, learn sub-pages (mastered/unknown/review), notices, profile-edit, redeem, purchase | Build compiles; pages unaffected functionally. |

### Regression traps

- Inlined base64 font bloating `app.wxss` beyond sanity (expected delta: ~5–8 KB).
- `bind:tap` vs `bind:click` mismatch in wrapper — verify theme toggle + copy-code buttons.
- van-cell slot alignment — visually confirm 5 rows in me.wxml.
- Skyline rendering + embedded fonts — spot-check first launch.

### Go/no-go gate

- Zero TS errors on `tsc --noEmit`.
- No literal "undefined" visible on home page against populated backend.
- Dark-mode toggle icon visible and functional in both themes.
- Every former icon site still renders an icon with equivalent size / color / tap behavior.
- Tab bar active state readable in both themes.

## 10. Rollout

Single feature branch in the dx-mini repo. Per user workflow: merge locally to `main`, push `main` only (no feature-branch push).

Commit structure (suggested):

1. `feat(mini): add dx-icon component with Lucide iconfont`
2. `fix(mini): home page greeting shape and title/subtitle rendering`
3. `fix(mini): dark-mode toggle uses moon/sun via dx-icon`
4. `refactor(mini): migrate all van-icon usages to dx-icon outline set`
5. `refactor(mini): tab bar uses outline icons with color-only active state`

Keeps each concern reviewable in isolation; final `main` merge reads as a coherent set.

## 11. Open items

None at design-freeze. Implementation plan (separate document) will expand each §6 subsection into ordered TDD-friendly steps.
