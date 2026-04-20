# dx-mini: Lucide SVG Icons Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the iconfont-based `<dx-icon>` implementation with a component that renders Lucide SVGs directly via `<image>` data URIs, keeping the public API identical so no page code changes.

**Architecture:** A new Node build script reads 20 Lucide SVGs from `lucide-static`, asserts the `currentColor` and `stroke-width="2"` substitution tokens are present, statically validates that every literal `<dx-icon name="...">` in WXML is declared, and emits `icons.ts` — a typed `{ [name]: svgString }` map. The rewritten `dx-icon` component observes `name`/`size`/`color`/`strokeWidth`/`customStyle`, substitutes color + stroke-width into the raw SVG at runtime, URI-encodes it into a `data:image/svg+xml;utf8,…` URL, and renders a `<view>`-wrapped `<image>`.

**Tech Stack:** WeChat Mini Program (glass-easel), TypeScript (strict), Node 18+ for the build script, `lucide-static@^0.460`.

**Spec:** [2026-04-20-dx-mini-lucide-svg-icons-design.md](../specs/2026-04-20-dx-mini-lucide-svg-icons-design.md)

---

## Task 1: Write the new build script

**Files:**
- Create: `dx-mini/scripts/build-icons.mjs`

- [ ] **Step 1: Create `dx-mini/scripts/build-icons.mjs`**

```javascript
// Builds miniprogram/components/dx-icon/icons.ts from Lucide SVGs.
//
// Adaptation notes:
//   - lucide-static@0.460 renamed "home" -> "house" and "help-circle" -> "circle-help".
//     The logical icon names (used by <dx-icon name="...">) remain "home" and "help-circle".

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '..')

// Canonical inventory. Each entry: [logicalName, lucideFilename].
// Add a row here and re-run `npm run build:icons` to expose a new icon.
const ICONS = [
  ['moon',          'moon'],
  ['sun',           'sun'],
  ['search',        'search'],
  ['bell',          'bell'],
  ['chevron-right', 'chevron-right'],
  ['chevron-left',  'chevron-left'],
  ['star',          'star'],
  ['book-open',     'book-open'],
  ['check',         'check'],
  ['help-circle',   'circle-help'],   // lucide-static renamed help-circle -> circle-help
  ['clock',         'clock'],
  ['crown',         'crown'],
  ['users',         'users'],
  ['gift',          'gift'],
  ['ticket',        'ticket'],
  ['copy',          'copy'],
  ['home',          'house'],         // lucide-static renamed home -> house
  ['trending-up',   'trending-up'],
  ['notebook-text', 'notebook-text'],
  ['user',          'user'],
]

const lucideDir = path.join(repoRoot, 'node_modules', 'lucide-static', 'icons')
const wxmlRoot = path.join(repoRoot, 'miniprogram')
const outPath = path.join(repoRoot, 'miniprogram', 'components', 'dx-icon', 'icons.ts')

// Read + assert each SVG has the runtime-injection tokens we rely on.
const svgs = {}
for (const [logicalName, lucideFile] of ICONS) {
  const srcPath = path.join(lucideDir, `${lucideFile}.svg`)
  if (!fs.existsSync(srcPath)) {
    throw new Error(`lucide-static is missing "${lucideFile}.svg" (logical: "${logicalName}"). Check the lucide-static version in package.json.`)
  }
  const svg = fs.readFileSync(srcPath, 'utf8').trim()
  if (!svg.includes('currentColor')) {
    throw new Error(`Lucide SVG "${lucideFile}" has no "currentColor" — runtime color injection would no-op for "${logicalName}".`)
  }
  if (!svg.includes('stroke-width="2"')) {
    throw new Error(`Lucide SVG "${lucideFile}" has no stroke-width="2" — runtime stroke-width injection would no-op for "${logicalName}".`)
  }
  svgs[logicalName] = svg
}

// Static WXML scan: every literal <dx-icon name="foo"> must be declared in ICONS.
// Dynamic {{...}} bindings don't match the regex and are intentionally skipped —
// those cases (e.g. tabbar item.icon) are covered by the curated list itself.
const declared = new Set(ICONS.map(([name]) => name))
const wxmlFiles = []
const walk = (dir) => {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, entry.name)
    if (entry.isDirectory()) walk(p)
    else if (entry.name.endsWith('.wxml')) wxmlFiles.push(p)
  }
}
walk(wxmlRoot)

const pattern = /<dx-icon[^>]*\sname="([a-z0-9-]+)"/g
for (const file of wxmlFiles) {
  const lines = fs.readFileSync(file, 'utf8').split('\n')
  for (let i = 0; i < lines.length; i++) {
    for (const match of lines[i].matchAll(pattern)) {
      const name = match[1]
      if (!declared.has(name)) {
        const rel = path.relative(repoRoot, file)
        throw new Error(`${rel}:${i + 1}: <dx-icon name="${name}"/> not in ICONS. Add it to scripts/build-icons.mjs and re-run npm run build:icons.`)
      }
    }
  }
}

// Emit icons.ts
const body = ICONS
  .map(([name]) => `  ${JSON.stringify(name)}: ${JSON.stringify(svgs[name])},`)
  .join('\n')

const content =
  `// Auto-generated by scripts/build-icons.mjs — do not edit by hand.\n` +
  `// Run \`npm run build:icons\` to regenerate.\n\n` +
  `export const icons: Record<string, string> = {\n` +
  body + '\n' +
  `}\n`

fs.writeFileSync(outPath, content)

console.log(`Wrote ${ICONS.length} icons to ${path.relative(repoRoot, outPath)}.`)
```

- [ ] **Step 2: Verify the script parses**

Run: `cd dx-mini && node --check scripts/build-icons.mjs`
Expected: No output, exit code 0.

---

## Task 2: Update dx-mini/package.json + install

**Files:**
- Modify: `dx-mini/package.json`

- [ ] **Step 1: Edit `dx-mini/package.json`**

Replace the entire file contents with:

```json
{
  "name": "miniprogram-ts-quickstart",
  "version": "1.0.0",
  "description": "",
  "type": "module",
  "scripts": {
    "build:icons": "node scripts/build-icons.mjs"
  },
  "keywords": [],
  "author": "",
  "license": "",
  "devDependencies": {
    "lucide-static": "^0.460.0",
    "miniprogram-api-typings": "^2.8.3-1"
  }
}
```

Changes:
- Removed `fantasticon` and `oslllo-svg-fixer` (no longer needed).
- Renamed `build:iconfont` → `build:icons`.

- [ ] **Step 2: Reinstall dependencies**

Run: `cd dx-mini && npm install`
Expected: Lockfile updated, `fantasticon` + `oslllo-svg-fixer` removed from `node_modules`, `lucide-static` retained.

- [ ] **Step 3: Verify only expected devDeps remain**

Run: `cd dx-mini && npm ls --depth=0 --dev 2>/dev/null | head -10`
Expected: Output lists `lucide-static@0.460.x` and `miniprogram-api-typings@2.8.x`; no `fantasticon`, no `oslllo-svg-fixer`.

---

## Task 3: Generate `icons.ts`

**Files:**
- Create (auto-generated): `dx-mini/miniprogram/components/dx-icon/icons.ts`

- [ ] **Step 1: Run the build script**

Run: `cd dx-mini && npm run build:icons`
Expected: Single log line `Wrote 20 icons to miniprogram/components/dx-icon/icons.ts.`

- [ ] **Step 2: Verify the output file exists and has 20 entries**

Run: `cd dx-mini && grep -c "^  \"" miniprogram/components/dx-icon/icons.ts`
Expected: `20`

- [ ] **Step 3: Verify each logical icon name is present**

Run:
```bash
cd dx-mini && for n in moon sun search bell chevron-right chevron-left star book-open check help-circle clock crown users gift ticket copy home trending-up notebook-text user; do
  grep -q "\"$n\":" miniprogram/components/dx-icon/icons.ts || echo "MISSING: $n"
done
```
Expected: No output (every name found).

---

## Task 4: Rewrite `dx-icon/index.ts`

**Files:**
- Modify: `dx-mini/miniprogram/components/dx-icon/index.ts`

- [ ] **Step 1: Replace the file contents**

Replace `dx-mini/miniprogram/components/dx-icon/index.ts` with:

```typescript
import { icons } from './icons'

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    name: { type: String, value: '' },
    size: { type: String, value: '' },
    color: { type: String, value: '' },
    strokeWidth: { type: String, value: '1.25' },
    customStyle: { type: String, value: '' },
    customClass: { type: String, value: '' },
  },
  data: {
    src: '',
    hostStyle: '',
  },
  observers: {
    'name, size, color, strokeWidth, customStyle'(
      name: string,
      size: string,
      color: string,
      strokeWidth: string,
      customStyle: string,
    ) {
      const normalizedSize = /^\d+(\.\d+)?$/.test(size) ? `${size}px` : size
      const hostStyle = `width:${normalizedSize};height:${normalizedSize};${customStyle}`
      const raw = (icons as Record<string, string>)[name] || ''
      const svg = raw
        .replace(/currentColor/g, color || '#000')
        .replace(/stroke-width="2"/g, `stroke-width="${strokeWidth}"`)
      const src = svg ? `data:image/svg+xml;utf8,${encodeURIComponent(svg)}` : ''
      this.setData({ src, hostStyle })
    },
  },
  methods: {
    onClick(e: WechatMiniprogram.CustomEvent) {
      this.triggerEvent('click', e.detail)
    },
  },
})
```

Notes:
- The observer watches 5 props via the comma-list syntax and is called once per micro-task when any listed prop changes.
- `normalizedSize` appends `px` only when the incoming string is purely numeric; e.g. `"22px"`, `"2em"`, `"48rpx"` all pass through unchanged.
- When `name` resolves to a missing key (shouldn't happen because of Task 1 Step 1's WXML scan, but belt-and-suspenders), `src` stays empty and the `<image>` renders blank rather than crashing.

---

## Task 5: Rewrite `dx-icon/index.wxml`

**Files:**
- Modify: `dx-mini/miniprogram/components/dx-icon/index.wxml`

- [ ] **Step 1: Replace the file contents**

Replace `dx-mini/miniprogram/components/dx-icon/index.wxml` with:

```xml
<view class="dx-icon-host {{customClass}}" style="{{hostStyle}}" bind:tap="onClick">
  <image src="{{src}}" style="width:100%;height:100%" mode="aspectFit" />
</view>
```

---

## Task 6: Rewrite `dx-icon/index.wxss`

**Files:**
- Modify: `dx-mini/miniprogram/components/dx-icon/index.wxss`

- [ ] **Step 1: Replace the file contents**

Replace `dx-mini/miniprogram/components/dx-icon/index.wxss` with:

```css
.dx-icon-host {
  display: inline-block;
  vertical-align: middle;
  line-height: 1;
}
```

---

## Task 7: Rewrite `dx-icon/index.json`

**Files:**
- Modify: `dx-mini/miniprogram/components/dx-icon/index.json`

- [ ] **Step 1: Replace the file contents**

Replace `dx-mini/miniprogram/components/dx-icon/index.json` with:

```json
{
  "component": true
}
```

The previous `"usingComponents": { "van-icon": "@vant/weapp/icon/index" }` is removed — the component no longer wraps van-icon.

---

## Task 8: TypeScript + grep pre-commit verification

**Files:** read-only

- [ ] **Step 1: Run tsc in strict mode**

Run: `cd dx-mini && npx tsc --noEmit`
Expected: Zero errors. (If tsc prints errors about the observer params or `this`, stop and investigate — do not commit until clean.)

- [ ] **Step 2: Confirm no iconfont references remain in component code**

Run: `cd dx-mini && grep -rn "dx-iconfont\|van-icon\|@vant/weapp/icon" miniprogram/components/dx-icon/`
Expected: No matches.

- [ ] **Step 3: Confirm `icons.ts` is reachable from the component**

Run: `cd dx-mini && grep -n "from './icons'" miniprogram/components/dx-icon/index.ts`
Expected: Exactly one match (`import { icons } from './icons'`).

---

## Task 9: Commit the renderer swap

**Files:**
- Stage: `dx-mini/package.json`, `dx-mini/package-lock.json`, `dx-mini/scripts/build-icons.mjs`, `dx-mini/miniprogram/components/dx-icon/index.ts`, `dx-mini/miniprogram/components/dx-icon/index.wxml`, `dx-mini/miniprogram/components/dx-icon/index.wxss`, `dx-mini/miniprogram/components/dx-icon/index.json`, `dx-mini/miniprogram/components/dx-icon/icons.ts`

- [ ] **Step 1: Review staged changes**

Run: `git status --short`
Expected: Modified `package.json`, `package-lock.json`, and the four dx-icon component files; new `scripts/build-icons.mjs` and `icons.ts`.

- [ ] **Step 2: Stage exactly those files**

Run:
```bash
git add dx-mini/package.json dx-mini/package-lock.json \
  dx-mini/scripts/build-icons.mjs \
  dx-mini/miniprogram/components/dx-icon/index.ts \
  dx-mini/miniprogram/components/dx-icon/index.wxml \
  dx-mini/miniprogram/components/dx-icon/index.wxss \
  dx-mini/miniprogram/components/dx-icon/index.json \
  dx-mini/miniprogram/components/dx-icon/icons.ts
```

- [ ] **Step 3: Commit**

Run:
```bash
git commit -m "$(cat <<'EOF'
feat(mini): replace iconfont with inline Lucide SVG in <dx-icon>

Rewrites the dx-icon component to render Lucide SVGs directly via an
<image> data URI. Color is substituted into currentColor at runtime and
stroke-width defaults to 1.25. Public props (name/size/color/customStyle/
customClass) and the click event are unchanged, so the 32 call sites
across pages and custom-tab-bar keep working without edits.

The new build script scripts/build-icons.mjs reads Lucide SVGs from
lucide-static, asserts the substitution tokens exist, statically checks
that every literal <dx-icon name="..."> in WXML is declared, and emits
miniprogram/components/dx-icon/icons.ts.

The old iconfont assets and app.wxss block are removed in a follow-up
commit.
EOF
)"
```

Expected: Commit created.

---

## Task 10: Remove the `dx-iconfont` block from `app.wxss`

**Files:**
- Modify: `dx-mini/miniprogram/app.wxss`

- [ ] **Step 1: Delete the auto-generated block**

Open `dx-mini/miniprogram/app.wxss` and delete lines 56 through 92 inclusive — everything from `/* === dx-iconfont (auto-generated by scripts/build-iconfont.mjs — do not edit by hand) === */` through `/* === end dx-iconfont === */`. Keep the trailing blank line so the file doesn't collapse the surrounding rules.

After the edit, the file should end at the `.page-container.dark { background: #0f0f0f; }` / `.card {...}` rules (around line 55) with no iconfont-related content afterward.

- [ ] **Step 2: Verify no iconfont rule or font-face remains**

Run: `cd dx-mini && grep -c "dx-iconfont\|dx-icon-[a-z]" miniprogram/app.wxss`
Expected: `0`

---

## Task 11: Delete font assets

**Files:**
- Delete: `dx-mini/miniprogram/assets/fonts/dx-iconfont.woff2`
- Delete: `dx-mini/miniprogram/assets/fonts/dx-iconfont.codepoints.json`
- Delete (if empty): `dx-mini/miniprogram/assets/fonts/`

- [ ] **Step 1: Remove both font artifacts**

Run:
```bash
rm dx-mini/miniprogram/assets/fonts/dx-iconfont.woff2 \
   dx-mini/miniprogram/assets/fonts/dx-iconfont.codepoints.json
```

- [ ] **Step 2: Remove the fonts directory if empty**

Run: `rmdir dx-mini/miniprogram/assets/fonts/ 2>/dev/null || echo "directory not empty — leaving it"`
Expected: Either silent (removed) or the "not empty" message (in which case leave it — something else landed there).

- [ ] **Step 3: Confirm no stray references to the deleted files**

Run: `cd dx-mini && grep -rn "dx-iconfont.woff2\|dx-iconfont.codepoints" miniprogram/ scripts/ 2>/dev/null`
Expected: No matches.

---

## Task 12: Delete the old build script

**Files:**
- Delete: `dx-mini/scripts/build-iconfont.mjs`

- [ ] **Step 1: Remove the obsolete script**

Run: `rm dx-mini/scripts/build-iconfont.mjs`

- [ ] **Step 2: Verify no references remain anywhere**

Run: `cd dx-mini && grep -rn "build-iconfont\|build:iconfont" . --include=*.{mjs,js,ts,json,md} 2>/dev/null`
Expected: Only matches inside `dx-mini/docs/superpowers/specs/2026-04-19-*.md` or `dx-mini/docs/superpowers/plans/2026-04-19-*.md` (historical docs — leave untouched). Nothing in live code or in `package.json`.

---

## Task 13: Update the project `CLAUDE.md` icon convention

**Files:**
- Modify: `CLAUDE.md` (repo root)

- [ ] **Step 1: Replace the icons convention line**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/CLAUDE.md`, find this bullet in the `dx-mini Conventions` section:

```
- **Icons** — Lucide via the `<dx-icon>` wrapper. NEVER import `<van-icon>` directly in pages; add glyphs to `scripts/build-iconfont.mjs` ICONS array and re-run `npm run build:iconfont`.
```

Replace it with:

```
- **Icons** — Lucide SVG via the `<dx-icon>` component. Add glyphs to `scripts/build-icons.mjs` ICONS array and re-run `npm run build:icons`. Color via the `color` prop, default stroke width 1.25.
```

- [ ] **Step 2: Verify the replacement**

Run: `grep -n "build:icons\|build:iconfont" CLAUDE.md`
Expected: Exactly one match for `build:icons`, zero matches for `build:iconfont`.

---

## Task 14: Final repo-wide verification

**Files:** read-only

- [ ] **Step 1: No `dx-iconfont` reference anywhere in dx-mini**

Run: `cd dx-mini && grep -rn "dx-iconfont" --include=*.{ts,js,mjs,json,wxml,wxss,md} . 2>/dev/null`
Expected: Only historical doc matches under `docs/superpowers/specs/2026-04-19-*.md` and `docs/superpowers/plans/2026-04-19-*.md`. No matches in live code.

- [ ] **Step 2: No `<van-icon>` in pages or tabbar**

Run: `cd dx-mini && grep -rn "<van-icon" miniprogram/pages miniprogram/custom-tab-bar 2>/dev/null`
Expected: Zero matches.

- [ ] **Step 3: `tsc --noEmit` still clean**

Run: `cd dx-mini && npx tsc --noEmit`
Expected: Zero errors.

- [ ] **Step 4: Dependencies pruned**

Run: `cd dx-mini && grep -E "fantasticon|oslllo-svg-fixer" package.json`
Expected: Zero matches.

---

## Task 15: Manual smoke test in WeChat DevTools

This task is performed by the user — the agent must STOP and hand off.

**Prerequisites:** WeChat Developer Tools open, project loaded at `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini`. If `miniprogram_npm` is stale, run 构建 npm.

- [ ] **Step 1: Compile the project in DevTools**

Use DevTools 编译 / 预览. Expected: no console errors on boot; login page renders.

- [ ] **Step 2: Icon checklist — verify each page**

Visit each page and confirm every icon renders with the expected color, size, and tap behavior. Both light and dark themes where applicable.

  - [ ] `pages/home` — search (top bar), theme toggle (moon ↔ sun on tap), bell (tap triggers `goNotices`).
  - [ ] `pages/learn` — check, chevron-right, help-circle, clock.
  - [ ] `pages/me` — chevron-right, crown, and `<van-cell slot="icon">` projections: bell, users, gift, ticket, crown.
  - [ ] `pages/me/groups` — chevron-right.
  - [ ] `pages/me/invite` — copy.
  - [ ] `pages/me/groups-detail` — copy.
  - [ ] `pages/games/favorites` — chevron-right, star.
  - [ ] `pages/games/detail` — book-open, star toggle (favorited vs not), chevron-right.
  - [ ] `pages/games` — star, book-open.
  - [ ] `pages/games/play` — chevron-left, check, help-circle, clock.
  - [ ] `custom-tab-bar` — 5 tabs (home, book-open, trending-up, notebook-text, user) render in both active (primary color) and inactive (#9ca3af) states; switching tabs still works.

- [ ] **Step 3: Real-device check via 预览 + 小程序助手**

Scan the DevTools 预览 QR code with 小程序助手 (per project memory — 真机调试 is broken on the current DevTools build). Repeat the icon checklist on a real phone.

- [ ] **Step 4: Confirm completion**

If every icon renders correctly and no tap handler regressed, proceed to Task 16. If a specific icon looks wrong, capture the page and icon name and stop — iteration needed.

---

## Task 16: Commit the cleanup + docs

**Files:**
- Stage: `dx-mini/miniprogram/app.wxss`, `CLAUDE.md`, deleted `dx-mini/miniprogram/assets/fonts/*`, deleted `dx-mini/scripts/build-iconfont.mjs`

- [ ] **Step 1: Review staged changes**

Run: `git status --short`
Expected: Modified `dx-mini/miniprogram/app.wxss` and `CLAUDE.md`; deleted `dx-mini/miniprogram/assets/fonts/dx-iconfont.woff2`, `dx-mini/miniprogram/assets/fonts/dx-iconfont.codepoints.json`, and `dx-mini/scripts/build-iconfont.mjs`.

- [ ] **Step 2: Stage the changes**

Run:
```bash
git add dx-mini/miniprogram/app.wxss CLAUDE.md
git add -u dx-mini/miniprogram/assets/fonts/ dx-mini/scripts/build-iconfont.mjs
```

(`git add -u` stages the deletions; do not use `git add .`.)

- [ ] **Step 3: Commit**

Run:
```bash
git commit -m "$(cat <<'EOF'
chore(mini): remove dx-iconfont assets and tooling

Drops the base64 font-face block from app.wxss, the generated woff2
and codepoints.json under miniprogram/assets/fonts/, and the old
build-iconfont.mjs script. Updates CLAUDE.md's dx-mini icon convention
to describe the new SVG-based workflow.
EOF
)"
```

Expected: Commit created, working tree clean.

- [ ] **Step 4: Final status check**

Run: `git status`
Expected: `nothing to commit, working tree clean`.

---

## Task 17: Persist the new convention in memory

**Files:**
- Create: `/Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/feedback_dx_mini_icon_strategy.md`
- Modify: `/Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/MEMORY.md`

- [ ] **Step 1: Create the feedback memory file**

Write this exact content to `/Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/feedback_dx_mini_icon_strategy.md`:

```markdown
---
name: dx-mini Icon Strategy
description: dx-mini icons are always Lucide SVGs rendered via the <dx-icon> component — never iconfont, never <van-icon>, never emoji.
type: feedback
---

dx-mini uses Lucide SVG icons only, rendered through `<dx-icon>`.

**Why:** The project migrated off iconfont to get true SVG fidelity, runtime color control, and per-instance stroke width. Lucide is the canonical visual language shared with dx-web.

**How to apply:**
- Add new icons by appending `[logicalName, lucideFilename]` to the `ICONS` array in `dx-mini/scripts/build-icons.mjs`, then running `npm run build:icons` to regenerate `dx-mini/miniprogram/components/dx-icon/icons.ts`. Commit the regenerated file.
- Never introduce `<van-icon>`, iconfont `.dx-icon-*` classes, emoji-as-icon, or raw `<image src="...svg">` tags in pages.
- `<dx-icon>` props: `name` (required), `size` (e.g. `"22px"`), `color` (CSS color, required for visibility since `<image>` doesn't inherit text color), optional `strokeWidth` (default `"1.25"`), `customStyle`, `customClass`. Tap via `bind:click`.
- If a Lucide SVG has been renamed upstream (e.g. `home` → `house`), keep the logical name stable and record the rename as a comment on the ICONS row — call sites never learn about the upstream filename.
```

- [ ] **Step 2: Add the index entry to MEMORY.md**

Open `/Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/MEMORY.md` and append this line at the end of the index list (after the existing `project_wechat_devtools_realdevice_bug.md` entry):

```
- [feedback_dx_mini_icon_strategy.md](feedback_dx_mini_icon_strategy.md) — dx-mini icons are always Lucide SVG via <dx-icon>; iconfont and van-icon are retired
```

- [ ] **Step 3: Verify the memory files**

Run: `ls /Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/feedback_dx_mini_icon_strategy.md`
Expected: File listed (no "No such file").

Run: `grep -c "feedback_dx_mini_icon_strategy" /Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/MEMORY.md`
Expected: `1`

Memory files are outside the git repo — no commit needed. This task is the final step.

---

## Rollback plan

Each commit is self-contained and revertable.

- To undo cleanup only (keep the new renderer but bring the font assets back): `git revert <cleanup-commit>`.
- To undo both (restore iconfont rendering end-to-end): `git revert <cleanup-commit> && git revert <feature-commit>`.

If a regression shows up after cleanup (e.g. DevTools complains about a specific Lucide SVG on an old device), the fastest mitigation is swapping `data:image/svg+xml;utf8,` for `data:image/svg+xml;base64,` in `index.ts`'s observer — one-line change, no rebuild needed.
