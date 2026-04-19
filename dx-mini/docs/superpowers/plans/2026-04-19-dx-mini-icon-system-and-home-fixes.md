# dx-mini: Icon System Migration & Home Page Fixes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix two undefineds on home page, fix missing dark-mode toggle icon, and migrate every icon usage in dx-mini to an outline Lucide-based iconfont exposed through a new `<dx-icon>` wrapper component.

**Architecture:** New `<dx-icon>` component wraps `<van-icon class-prefix="dx">`. A curated 20-glyph Lucide woff2 is base64-inlined in `app.wxss` via `@font-face` and mapped to `.dx-icon-{name}::before` codepoints. Home page Greeting interface is realigned to backend shape `{title, subtitle}` and renders as two stacked lines (dx-web parity). Tab bar loses its `activeIcon` field (dead even today); active state is color-only.

**Tech Stack:** WeChat Mini Program (native + TypeScript strict), Vant Weapp 1.11.x, glass-easel, Skyline rendering, `lucide-static` + `fantasticon` for one-shot font generation.

**Reference spec:** `docs/superpowers/specs/2026-04-19-dx-mini-icon-system-and-home-fixes-design.md`

**Note on TDD in this repo:** dx-mini has no automated test framework and none is in scope. The closest compile-time gate is `tsc --noEmit` plus WeChat DevTools compile. Every task below ends with that gate. The user has explicitly elected to do simulator walkthroughs themselves — do not claim feature correctness without their confirmation, only compile correctness.

**Working directory for every step:** `/Users/rainsen/Programs/Projects/douxue/dx-mini` (separate git repo from dx-source). Run all `git`, `npm`, and `tsc` commands from that directory.

---

## Task 1: Create the `<dx-icon>` component scaffold

**Files:**
- Create: `miniprogram/components/dx-icon/index.ts`
- Create: `miniprogram/components/dx-icon/index.wxml`
- Create: `miniprogram/components/dx-icon/index.wxss`
- Create: `miniprogram/components/dx-icon/index.json`

- [ ] **Step 1.1: Create component JSON registration**

Write `miniprogram/components/dx-icon/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

- [ ] **Step 1.2: Create component WXML**

Write `miniprogram/components/dx-icon/index.wxml`:

```xml
<van-icon
  class-prefix="dx-icon"
  name="{{name}}"
  size="{{size}}"
  color="{{color}}"
  custom-style="{{customStyle}}"
  custom-class="{{customClass}}"
  bind:click="onClick"
/>
```

Note: `class-prefix="dx-icon"` — Vant renders `<text class="dx-icon dx-icon-{name}">`. The CSS we add in Task 2 targets `.dx-icon-{name}::before`, which matches.

- [ ] **Step 1.3: Create empty component WXSS**

Write `miniprogram/components/dx-icon/index.wxss`:

```css
/* Glyphs are defined globally in app.wxss via @font-face + .dx-icon-*::before rules. */
```

- [ ] **Step 1.4: Create component TypeScript**

Write `miniprogram/components/dx-icon/index.ts`:

```ts
Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    name: { type: String, value: '' },
    size: { type: String, value: '' },
    color: { type: String, value: '' },
    customStyle: { type: String, value: '' },
    customClass: { type: String, value: '' },
  },
  methods: {
    onClick(e: WechatMiniprogram.CustomEvent) {
      this.triggerEvent('click', e.detail)
    },
  },
})
```

`addGlobalClass: true` lets app-level `.dx-icon-*` CSS reach the slotted `<text>` that Vant renders inside the component.

- [ ] **Step 1.5: Verify TypeScript compiles**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram`

Expected: clean exit, no errors. If `tsc` complains about missing types for `WechatMiniprogram.CustomEvent`, this type comes from `miniprogram/typings/` which already ships with the project — no extra setup needed.

- [ ] **Step 1.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/components/dx-icon/
git commit -m "feat(mini): add dx-icon component scaffold (no font yet)"
```

---

## Task 2: Generate the Lucide iconfont and wire it into `app.wxss`

This task produces the woff2 glyph file, base64-encodes it, and installs it in `app.wxss` with class definitions for all 20 icons.

**Files:**
- Create: `scripts/build-iconfont.mjs` (dev tool, committed for reproducibility)
- Create: `package.json` at repo root (if not present) with devDependencies
- Modify: `miniprogram/app.wxss` (append @font-face + glyph classes)
- Create (committed for reference): `miniprogram/assets/fonts/dx-iconfont.woff2`
- Create (committed for reference): `miniprogram/assets/fonts/dx-iconfont.codepoints.json`

- [ ] **Step 2.1: Verify / initialize repo-root `package.json`**

Check whether `/Users/rainsen/Programs/Projects/douxue/dx-mini/package.json` already exists. If the dx-mini root has no `package.json` (only `miniprogram/package.json`), create it:

```json
{
  "name": "dx-mini-tools",
  "private": true,
  "type": "module",
  "scripts": {
    "build:iconfont": "node scripts/build-iconfont.mjs"
  },
  "devDependencies": {
    "lucide-static": "^0.460.0",
    "fantasticon": "^3.0.0"
  }
}
```

If a `package.json` already exists, instead run:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
npm install --save-dev lucide-static@^0.460.0 fantasticon@^3.0.0
```

Then add the `build:iconfont` script entry to the existing `scripts` block.

- [ ] **Step 2.2: Install deps**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npm install
```
Expected: `node_modules/` populated at the dx-mini repo root (separate from `miniprogram/node_modules/`). This root-level `node_modules` is a dev tool and should be gitignored — add `node_modules/` to `.gitignore` at the dx-mini root if not already present.

- [ ] **Step 2.3: Write the font-build script**

Write `scripts/build-iconfont.mjs`:

```js
// Builds miniprogram/assets/fonts/dx-iconfont.woff2 from Lucide SVGs,
// then regenerates app.wxss's font-face block (base64 inline) and the
// .dx-icon-{name}::before rules. Idempotent.

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import os from 'node:os'
import { generateFonts } from 'fantasticon'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '..')

// Canonical inventory. Order = codepoint order (\e001..\e014).
const ICONS = [
  'moon',
  'sun',
  'search',
  'bell',
  'chevron-right',
  'chevron-left',
  'star',
  'book-open',
  'check',
  'help-circle',
  'clock',
  'crown',
  'users',
  'gift',
  'ticket',
  'copy',
  'home',
  'trending-up',
  'notebook-text',
  'user',
]

const lucideDir = path.join(repoRoot, 'node_modules', 'lucide-static', 'icons')
const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'dx-iconfont-'))
const fontOutDir = path.join(repoRoot, 'miniprogram', 'assets', 'fonts')
fs.mkdirSync(fontOutDir, { recursive: true })

// Stage: copy only the needed SVGs into tmp
for (const name of ICONS) {
  const src = path.join(lucideDir, `${name}.svg`)
  if (!fs.existsSync(src)) {
    throw new Error(`lucide-static is missing "${name}.svg". Check the Lucide version in package.json.`)
  }
  fs.copyFileSync(src, path.join(tmpDir, `${name}.svg`))
}

// Stage: assign deterministic codepoints \e001..\eNNN
const codepoints = {}
ICONS.forEach((name, i) => {
  codepoints[name] = 0xe001 + i
})

fs.writeFileSync(
  path.join(fontOutDir, 'dx-iconfont.codepoints.json'),
  JSON.stringify(codepoints, null, 2) + '\n',
)

// Stage: run fantasticon
await generateFonts({
  name: 'dx-iconfont',
  inputDir: tmpDir,
  outputDir: tmpDir,
  fontTypes: ['woff2'],
  assetTypes: [],
  prefix: 'dx-icon',
  codepoints,
  normalize: true,
  fontHeight: 1024,
  descent: 0,
})

// Copy woff2 to miniprogram assets
const woffSrc = path.join(tmpDir, 'dx-iconfont.woff2')
const woffDest = path.join(fontOutDir, 'dx-iconfont.woff2')
fs.copyFileSync(woffSrc, woffDest)

// Stage: regenerate app.wxss block between markers
const wxssPath = path.join(repoRoot, 'miniprogram', 'app.wxss')
const wxss = fs.readFileSync(wxssPath, 'utf8')

const base64 = fs.readFileSync(woffDest).toString('base64')

const glyphLines = ICONS
  .map((name) => {
    const hex = codepoints[name].toString(16)
    return `.dx-icon-${name}::before { content: '\\${hex}'; }`
  })
  .join('\n')

const block =
  `/* === dx-iconfont (auto-generated by scripts/build-iconfont.mjs — do not edit by hand) === */\n` +
  `@font-face {\n` +
  `  font-family: 'dx-iconfont';\n` +
  `  src: url('data:font/woff2;base64,${base64}') format('woff2');\n` +
  `  font-weight: normal; font-style: normal; font-display: block;\n` +
  `}\n` +
  `[class*='dx-icon-']::before {\n` +
  `  font-family: 'dx-iconfont' !important;\n` +
  `  font-style: normal;\n` +
  `  font-weight: normal;\n` +
  `  speak: none;\n` +
  `  display: inline-block;\n` +
  `  text-decoration: none;\n` +
  `  -webkit-font-smoothing: antialiased;\n` +
  `  -moz-osx-font-smoothing: grayscale;\n` +
  `}\n` +
  glyphLines +
  `\n/* === end dx-iconfont === */\n`

const startMarker = '/* === dx-iconfont'
const endMarker = '/* === end dx-iconfont === */'

let nextWxss
if (wxss.includes(startMarker)) {
  const before = wxss.slice(0, wxss.indexOf(startMarker))
  const afterEnd = wxss.indexOf(endMarker) + endMarker.length
  const after = wxss.slice(afterEnd)
  nextWxss = before.trimEnd() + '\n\n' + block + after.replace(/^\n+/, '')
} else {
  nextWxss = wxss.trimEnd() + '\n\n' + block
}

fs.writeFileSync(wxssPath, nextWxss)

fs.rmSync(tmpDir, { recursive: true, force: true })

console.log(`Wrote ${ICONS.length} glyphs to ${woffDest} and patched ${wxssPath}.`)
```

- [ ] **Step 2.4: Run the build script**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npm run build:iconfont
```

Expected stdout: `Wrote 20 glyphs to .../dx-iconfont.woff2 and patched .../app.wxss.`

- [ ] **Step 2.5: Verify the output**

Check the WXSS patch:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -c "dx-icon-" miniprogram/app.wxss
```
Expected: `22` or more (20 glyph rules + one in selector `[class*='dx-icon-']` + one in the font-family reference, plus possibly more).

Check the woff2 exists:
```bash
ls -lh /Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram/assets/fonts/dx-iconfont.woff2
```
Expected: ~2–8 KB file.

Check the codepoints JSON:
```bash
cat /Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram/assets/fonts/dx-iconfont.codepoints.json
```
Expected: JSON with 20 keys mapping each Lucide name to decimal `57345`..`57364`.

- [ ] **Step 2.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add .gitignore package.json package-lock.json scripts/build-iconfont.mjs miniprogram/assets/fonts/ miniprogram/app.wxss
git commit -m "feat(mini): add Lucide iconfont (20 glyphs) inline in app.wxss"
```

Note: `package-lock.json` may or may not exist depending on npm setup — include it in the add list only if it was generated. Verify with `git status` first.

---

## Task 3: Fix #1 — Home page greeting undefineds

**Files:**
- Modify: `miniprogram/pages/home/home.ts`
- Modify: `miniprogram/pages/home/home.wxml`
- Modify: `miniprogram/pages/home/home.wxss`

- [ ] **Step 3.1: Update the `Greeting` interface in `home.ts`**

Open `miniprogram/pages/home/home.ts`. Find line 18:

```ts
interface Greeting { text: string; emoji: string }
```

Replace with:

```ts
interface Greeting { title: string; subtitle: string }
```

No other lines in `home.ts` need editing — the `data.greeting` type stays `Greeting | null`, and `this.setData({ greeting: dash.greeting, ... })` in `loadData()` stays verbatim.

- [ ] **Step 3.2: Update the home page template**

Open `miniprogram/pages/home/home.wxml`. Find line 23:

```xml
<text class="greeting">{{greeting ? greeting.emoji + ' ' + greeting.text : '你好！'}}</text>
```

Replace with the two-line stack:

```xml
<text class="greeting-title">{{greeting ? greeting.title : '你好！'}}</text>
<text wx:if="{{greeting}}" class="greeting-subtitle">{{greeting.subtitle}}</text>
```

- [ ] **Step 3.3: Update home page styles**

Open `miniprogram/pages/home/home.wxss`. Find the `.greeting` rule (starts ~line 35):

```css
.greeting {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  display: block;
  margin-bottom: 16px;
}
```

Replace it with:

```css
.greeting-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  display: block;
  margin-bottom: 4px;
}
.greeting-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  display: block;
  margin-bottom: 16px;
}
```

- [ ] **Step 3.4: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit, no errors.

- [ ] **Step 3.5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/home/home.ts miniprogram/pages/home/home.wxml miniprogram/pages/home/home.wxss
git commit -m "fix(mini): align home greeting shape to backend {title, subtitle}"
```

---

## Task 4: Fix #2 + migrate home page icons to `<dx-icon>`

**Files:**
- Modify: `miniprogram/pages/home/home.json`
- Modify: `miniprogram/pages/home/home.wxml`

- [ ] **Step 4.1: Register `dx-icon` in home.json**

Open `miniprogram/pages/home/home.json`. Current contents:

```json
{
  "navigationBarTitleText": "斗学",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-skeleton": "@vant/weapp/skeleton/index"
  }
}
```

Replace with (remove `van-icon`, add `dx-icon`):

```json
{
  "navigationBarTitleText": "斗学",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-skeleton": "@vant/weapp/skeleton/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 4.2: Migrate the search icon**

In `miniprogram/pages/home/home.wxml`, find line 6:

```xml
<van-icon name="search" size="16px" color="#9ca3af" />
```

Replace with:

```xml
<dx-icon name="search" size="16px" color="#9ca3af" />
```

- [ ] **Step 4.3: Migrate the theme-toggle icon (dark-mode fix)**

In `miniprogram/pages/home/home.wxml`, find lines 10–15:

```xml
<van-icon
  name="{{theme === 'dark' ? 'sunny-o' : 'moon-o'}}"
  size="22px"
  color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
  bind:tap="toggleTheme"
/>
```

Replace with:

```xml
<dx-icon
  name="{{theme === 'dark' ? 'sun' : 'moon'}}"
  size="22px"
  color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
  bind:click="toggleTheme"
/>
```

Key changes: `sunny-o`/`moon-o` → `sun`/`moon` (Lucide names); `bind:tap` → `bind:click` (the dx-icon wrapper re-emits as `click`).

- [ ] **Step 4.4: Migrate the bell icon**

In `miniprogram/pages/home/home.wxml`, find line 16:

```xml
<van-icon name="bell" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" bind:tap="goNotices" />
```

Replace with:

```xml
<dx-icon name="bell" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" bind:click="goNotices" />
```

- [ ] **Step 4.5: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit.

- [ ] **Step 4.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/home/home.json miniprogram/pages/home/home.wxml
git commit -m "fix(mini): dark-mode toggle uses sun/moon via dx-icon; home migrated"
```

---

## Task 5: Migrate games, detail, play, favorites pages

Each sub-step is self-contained. After all four pages are migrated, run the type-check gate once and commit.

**Files:**
- Modify: `miniprogram/pages/games/games.json` + `games.wxml`
- Modify: `miniprogram/pages/games/detail/detail.json` + `detail.wxml`
- Modify: `miniprogram/pages/games/play/play.json` + `play.wxml`
- Modify: `miniprogram/pages/games/favorites/favorites.json` + `favorites.wxml`

- [ ] **Step 5.1: Migrate `games` page**

**`miniprogram/pages/games/games.json`** — swap `van-icon` for `dx-icon` in `usingComponents`. Current file:

```json
{
  "navigationBarTitleText": "课程",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-search": "@vant/weapp/search/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-image": "@vant/weapp/image/index"
  }
}
```

Replace `"van-icon": "@vant/weapp/icon/index",` with `"dx-icon": "/components/dx-icon/index",` — preserve the surrounding entries exactly as they are in the repo (the full list above may have drifted; keep whatever other components the file already registers, only swap the icon line).

**`miniprogram/pages/games/games.wxml`** line 22:
```xml
<van-icon name="star-o" size="16px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
```
Replace with:
```xml
<dx-icon name="star" size="16px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
```

Line 45:
```xml
<van-icon name="column" size="28px" color="#9ca3af" />
```
Replace with:
```xml
<dx-icon name="book-open" size="28px" color="#9ca3af" />
```

- [ ] **Step 5.2: Migrate `detail` page**

**`miniprogram/pages/games/detail/detail.json`** — swap `van-icon` → `dx-icon` in `usingComponents` (same pattern as Step 5.1).

**`miniprogram/pages/games/detail/detail.wxml`** line 9:
```xml
<van-icon name="column" size="48px" color="#9ca3af" />
```
Replace with:
```xml
<dx-icon name="book-open" size="48px" color="#9ca3af" />
```

Line 12 (favorite toggle — unify to single outline, color-toggled):
```xml
<van-icon name="{{favorited ? 'star' : 'star-o'}}" size="24px" color="{{favorited ? '#f59e0b' : '#9ca3af'}}" />
```
Replace with:
```xml
<dx-icon name="star" size="24px" color="{{favorited ? '#f59e0b' : '#9ca3af'}}" />
```

Line 39:
```xml
<van-icon name="arrow" size="14px" color="#9ca3af" />
```
Replace with:
```xml
<dx-icon name="chevron-right" size="14px" color="#9ca3af" />
```

- [ ] **Step 5.3: Migrate `play` page**

**`miniprogram/pages/games/play/play.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/games/play/play.wxml`** line 9:
```xml
<view bind:tap="goBack"><van-icon name="arrow-left" size="20px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" /></view>
```
Replace with:
```xml
<view bind:tap="goBack"><dx-icon name="chevron-left" size="20px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" /></view>
```

Line 33:
```xml
<van-icon name="success" size="16px" color="#10b981" />
```
Replace with:
```xml
<dx-icon name="check" size="16px" color="#10b981" />
```

Line 37:
```xml
<van-icon name="question-o" size="16px" color="#f59e0b" />
```
Replace with:
```xml
<dx-icon name="help-circle" size="16px" color="#f59e0b" />
```

Line 41:
```xml
<van-icon name="clock-o" size="16px" color="#6366f1" />
```
Replace with:
```xml
<dx-icon name="clock" size="16px" color="#6366f1" />
```

- [ ] **Step 5.4: Migrate `favorites` page**

**`miniprogram/pages/games/favorites/favorites.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/games/favorites/favorites.wxml`** line 20:
```xml
<van-icon name="arrow" size="14px" color="#9ca3af" />
```
Replace with:
```xml
<dx-icon name="chevron-right" size="14px" color="#9ca3af" />
```

Line 23 (decorative star on favorite card):
```xml
<van-icon name="star" size="20px" color="#ffffff" />
```
Replace with:
```xml
<dx-icon name="star" size="20px" color="#ffffff" />
```

- [ ] **Step 5.5: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit.

- [ ] **Step 5.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/games/
git commit -m "refactor(mini): migrate games/detail/play/favorites icons to dx-icon"
```

---

## Task 6: Migrate learn, groups, groups-detail, invite pages

**Files:**
- Modify: `miniprogram/pages/learn/learn.json` + `learn.wxml`
- Modify: `miniprogram/pages/me/groups/groups.json` + `groups.wxml`
- Modify: `miniprogram/pages/me/groups-detail/groups-detail.json` + `groups-detail.wxml`
- Modify: `miniprogram/pages/me/invite/invite.json` + `invite.wxml`

- [ ] **Step 6.1: Migrate `learn` page**

**`miniprogram/pages/learn/learn.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/learn/learn.wxml`** line 24:
```xml
<van-icon name="success" size="20px" color="{{accentColors.teal}}" />
```
Replace with:
```xml
<dx-icon name="check" size="20px" color="{{accentColors.teal}}" />
```

Lines 26, 31, 36 (three chevron instances on sub-row links):
```xml
<van-icon name="arrow" size="14px" color="{{arrowColor}}" />
```
Replace each with:
```xml
<dx-icon name="chevron-right" size="14px" color="{{arrowColor}}" />
```

Line 29:
```xml
<van-icon name="question-o" size="20px" color="{{accentColors.amber}}" />
```
Replace with:
```xml
<dx-icon name="help-circle" size="20px" color="{{accentColors.amber}}" />
```

Line 34:
```xml
<van-icon name="clock-o" size="20px" color="{{accentColors.purple}}" />
```
Replace with:
```xml
<dx-icon name="clock" size="20px" color="{{accentColors.purple}}" />
```

- [ ] **Step 6.2: Migrate `groups` page**

**`miniprogram/pages/me/groups/groups.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/me/groups/groups.wxml`** line 17:
```xml
<van-icon name="arrow" size="14px" color="#9ca3af" />
```
Replace with:
```xml
<dx-icon name="chevron-right" size="14px" color="#9ca3af" />
```

- [ ] **Step 6.3: Migrate `groups-detail` page**

**`miniprogram/pages/me/groups-detail/groups-detail.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/me/groups-detail/groups-detail.wxml`** line 8:
```xml
<van-icon name="copy-o" size="14px" color="{{primaryColor}}" />
```
Replace with:
```xml
<dx-icon name="copy" size="14px" color="{{primaryColor}}" />
```

- [ ] **Step 6.4: Migrate `invite` page**

**`miniprogram/pages/me/invite/invite.json`** — swap `van-icon` → `dx-icon`.

**`miniprogram/pages/me/invite/invite.wxml`** line 9:
```xml
<van-icon name="copy-o" size="18px" color="{{primaryColor}}" />
```
Replace with:
```xml
<dx-icon name="copy" size="18px" color="{{primaryColor}}" />
```

- [ ] **Step 6.5: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit.

- [ ] **Step 6.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/learn/ miniprogram/pages/me/groups/ miniprogram/pages/me/groups-detail/ miniprogram/pages/me/invite/
git commit -m "refactor(mini): migrate learn/groups/groups-detail/invite icons to dx-icon"
```

---

## Task 7: Migrate the `me` page (includes van-cell slot pattern)

This is the only page where we need to replace `<van-cell icon="...">` attributes with a `<dx-icon slot="icon" ...>` child.

**Files:**
- Modify: `miniprogram/pages/me/me.json`
- Modify: `miniprogram/pages/me/me.ts`
- Modify: `miniprogram/pages/me/me.wxml`

- [ ] **Step 7.1: Add `cellIconColor` to `me.ts`**

Open `miniprogram/pages/me/me.ts`. Find the block where existing theme-aware colors like `arrowColor` and `primaryColor` are set (search for `arrowColor`). Alongside those, introduce `cellIconColor`.

Exact edit: wherever the page computes `arrowColor` (e.g., in `onShow` or a `refreshColors()` helper, depending on current file shape), add a parallel assignment:

```ts
// in the same setData block that sets arrowColor, primaryColor, etc.
this.setData({
  // ...existing keys,
  cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
})
```

Also add `cellIconColor: ''` to the `data: { ... }` initializer so the property exists from the first render.

If the file currently has no `refreshColors`-style helper and just hardcodes theme-aware colors once in `onShow`, follow the exact same pattern: compute `cellIconColor` next to the existing `arrowColor` computation.

- [ ] **Step 7.2: Swap `van-icon` → `dx-icon` in me.json**

Open `miniprogram/pages/me/me.json`. In `usingComponents`, replace `"van-icon": "@vant/weapp/icon/index"` with `"dx-icon": "/components/dx-icon/index"`.

Leave `van-cell`, `van-cell-group`, `van-image`, and any other non-icon Vant registrations untouched.

- [ ] **Step 7.3: Migrate standalone van-icon usages**

In `miniprogram/pages/me/me.wxml`, find line 26:
```xml
<van-icon name="arrow" size="16px" color="{{arrowColor}}" />
```
Replace with:
```xml
<dx-icon name="chevron-right" size="16px" color="{{arrowColor}}" />
```

Find line 47:
```xml
<van-icon name="vip-card-o" size="16px" color="#d97706" />
```
Replace with:
```xml
<dx-icon name="crown" size="16px" color="#d97706" />
```

- [ ] **Step 7.4: Migrate the five `van-cell icon=""` attributes to slot pattern**

In `miniprogram/pages/me/me.wxml`, find lines 52–56:

```xml
<van-cell title="公告通知" icon="bell" is-link bind:click="goNotices" />
<van-cell title="我的团队" icon="friends-o" is-link bind:click="goGroups" />
<van-cell title="推荐有礼" icon="gift-o" is-link bind:click="goInvite" />
<van-cell title="兑换码" icon="coupon-o" is-link bind:click="goRedeem" />
<van-cell title="购买会员" icon="vip-card-o" is-link bind:click="goPurchase" />
```

Replace with:

```xml
<van-cell title="公告通知" is-link bind:click="goNotices">
  <dx-icon slot="icon" name="bell" size="20px" color="{{cellIconColor}}" />
</van-cell>
<van-cell title="我的团队" is-link bind:click="goGroups">
  <dx-icon slot="icon" name="users" size="20px" color="{{cellIconColor}}" />
</van-cell>
<van-cell title="推荐有礼" is-link bind:click="goInvite">
  <dx-icon slot="icon" name="gift" size="20px" color="{{cellIconColor}}" />
</van-cell>
<van-cell title="兑换码" is-link bind:click="goRedeem">
  <dx-icon slot="icon" name="ticket" size="20px" color="{{cellIconColor}}" />
</van-cell>
<van-cell title="购买会员" is-link bind:click="goPurchase">
  <dx-icon slot="icon" name="crown" size="20px" color="{{cellIconColor}}" />
</van-cell>
```

- [ ] **Step 7.5: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit. If `cellIconColor` is flagged as unused, recheck Step 7.1 — the property must be both declared in `data` and updated in `onShow` (or wherever `arrowColor` is set).

- [ ] **Step 7.6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/me/me.json miniprogram/pages/me/me.ts miniprogram/pages/me/me.wxml
git commit -m "refactor(mini): migrate me page icons + van-cell slot icons to dx-icon"
```

---

## Task 8: Migrate the custom tab bar (and drop dead `activeIcon` field)

**Files:**
- Modify: `miniprogram/custom-tab-bar/index.json`
- Modify: `miniprogram/custom-tab-bar/index.ts`
- Modify: `miniprogram/custom-tab-bar/index.wxml`

Context: `custom-tab-bar/index.wxml` line 10 currently references only `item.icon` (not `item.activeIcon`). The active-state differentiation is color-driven at line 12. So the `activeIcon` field is already dead code and we remove it as part of this cleanup.

- [ ] **Step 8.1: Swap `van-icon` → `dx-icon` in `index.json`**

Open `miniprogram/custom-tab-bar/index.json`. Current:
```json
{
  "component": true,
  "usingComponents": {
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

Replace with:
```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 8.2: Update `index.ts` — new Lucide names, drop `activeIcon`**

Open `miniprogram/custom-tab-bar/index.ts`. Replace the entire file contents with:

```ts
interface TabItem {
  icon: string
  text: string
  path: string
}

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    tabs: [
      { icon: 'home',          text: '首页',   path: '/pages/home/home' },
      { icon: 'book-open',     text: '课程',   path: '/pages/games/games' },
      { icon: 'trending-up',   text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'notebook-text', text: '学习',   path: '/pages/learn/learn' },
      { icon: 'user',          text: '我的',   path: '/pages/me/me' },
    ] as TabItem[],
  },
  lifetimes: {
    attached() {
      const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
      this.setData({ theme: app.globalData.theme })
    },
  },
  methods: {
    switchTab(e: WechatMiniprogram.TouchEvent) {
      const path = e.currentTarget.dataset['path'] as string
      wx.switchTab({ url: path })
    },
  },
})
```

- [ ] **Step 8.3: Update `index.wxml`**

Open `miniprogram/custom-tab-bar/index.wxml`. Current lines 9–13:
```xml
<van-icon
  name="{{item.icon}}"
  size="22px"
  color="{{active === index ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : '#9ca3af'}}"
/>
```

Replace with:
```xml
<dx-icon
  name="{{item.icon}}"
  size="22px"
  color="{{active === index ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : '#9ca3af'}}"
/>
```

- [ ] **Step 8.4: Run the type-check gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit.

- [ ] **Step 8.5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/custom-tab-bar/
git commit -m "refactor(mini): tab bar uses outline dx-icons with color-only active state"
```

---

## Task 9: Final verification and handoff

**Files:** none (verification only)

- [ ] **Step 9.1: Audit that no `van-icon` reference remains outside the wrapper**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -rn "van-icon" miniprogram/ --exclude-dir=node_modules --exclude-dir=miniprogram_npm
```

Expected matches: ONLY
- `miniprogram/components/dx-icon/index.json` (the wrapper registers van-icon internally)
- `miniprogram/components/dx-icon/index.wxml` (the wrapper renders van-icon)
- Any page `.json` file that still needs `van-icon` for some reason (there should be none)

If anything else matches, revisit the relevant task and finish the migration before proceeding.

- [ ] **Step 9.2: Audit that no `*-o` outline-suffix icon names remain**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -rn "name=\"[a-z-]*-o\"" miniprogram/ --exclude-dir=node_modules --exclude-dir=miniprogram_npm
```

Expected: no matches.

- [ ] **Step 9.3: Audit that no filled-variant names slip through**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -rE "name=\"(sunny-o|moon-o|star-o|wap-home-o|wap-home|chart-trending-o|records|contact|column|success|arrow|arrow-left|clock-o|question-o|vip-card-o|copy-o|friends-o|gift-o|coupon-o)\"" miniprogram/ --exclude-dir=node_modules --exclude-dir=miniprogram_npm
```

Expected: no matches (all Vant names should have been replaced with Lucide names).

- [ ] **Step 9.4: Audit that no `van-cell icon="..."` attributes remain**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -rn 'van-cell[^>]* icon=' miniprogram/ --exclude-dir=node_modules --exclude-dir=miniprogram_npm
```

Expected: no matches. If any remain, the me page Task 7 migration is incomplete.

- [ ] **Step 9.5: Final TypeScript gate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && npx tsc --noEmit --project miniprogram
```
Expected: clean exit.

- [ ] **Step 9.6: Print the manual-test checklist for the user**

Output the following literal text to the user (do not paraphrase):

```
Implementation complete. Automated gates pass (tsc --noEmit clean, no legacy icon names detected).

Please run the manual smoke test in WeChat DevTools. Checklist (both light and dark theme):

Home:           no "undefined" text anywhere; greeting title + subtitle render; theme-toggle icon visible (moon in light, sun in dark); tap swaps icon + theme + tab bar; search and bell outline.
Games:          outline star on filter chip; book-open in empty state.
Detail:         book-open cover placeholder; outline star (amber when favorited, gray when not); chevron-right rows.
Play:           chevron-left back; check/help-circle/clock stat chips.
Favorites:      chevron-right rows; outline star decoration.
Learn:          three stat rows render check/help-circle/clock; chevron-right on sub-rows.
Me:             chevron-right, crown render; five cells show slotted bell/users/gift/ticket/crown; cell taps still navigate.
Groups / Groups-detail / Invite: chevron-right and copy render; copy handlers still fire toast.
Tab bar:        single outline glyph per tab; active tab teal, inactive gray; switching tabs works.
Unaffected:     login, leaderboard, learn sub-pages, notices, profile-edit, redeem, purchase — these pages should compile and render as before.

Report back any visual regressions or missing icons so I can fix before merge.
```

- [ ] **Step 9.7: Do NOT merge to main or push**

Per the user's git-workflow preference ("Merge feature branches to main locally and push only main; never push feature branches to remote"), the final `main` merge is the user's call. The feature work is complete; leave `main` untouched and wait for the user's manual-test sign-off before suggesting the merge.

---

## Self-review checklist

(For the plan author — not for the executor.)

- [x] Every spec section (§1–§11) has at least one task:
  - §1 Goals — Tasks 3, 4, 5–8
  - §3 Constraints (TS strict, no runtime font fetch) — explicit gates in every task
  - §4.1 `<dx-icon>` component — Task 1
  - §4.2 Iconfont generation + inlining — Task 2
  - §4.3 Architectural justification — no task needed (design rationale)
  - §5 Icon inventory — encoded in the `ICONS` array in Task 2.3 and used verbatim in every migration task
  - §6.1 Greeting fix — Task 3
  - §6.2 Dark-mode fix — Task 4 (Step 4.3)
  - §6.3 Outline migration — Tasks 4, 5, 6, 7, 8
  - §6.4 `.json` registration — covered in every page task's first sub-step
  - §7 Data-flow — no task needed (behavior doesn't change)
  - §8 Error handling — `greeting: null` fallback lives in Task 3.2 template
  - §9 Testing — Task 9
  - §10 Rollout — commit structure already mirrors spec's suggested sequence
  - §11 Open items — none
- [x] No placeholders, TBDs, or "similar to previous" references in task bodies.
- [x] Types and names consistent across tasks (e.g., `cellIconColor` in Task 7 is declared both in `data` and setData; `TabItem` interface in Task 8 drops `activeIcon` exactly where the old interface declared it).
- [x] Every code block shows the full text to write, not a sketch.
- [x] Every task ends with either a commit step or (for Task 9) a verification-only handoff.
