# dx-mini: Home Marketing Sections Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add six scroll-revealed marketing sections below the existing hub on `/pages/home/home` (WhyDifferent / Features / AiFeatures / LearningLoop / Community / Membership), anchored on live user data where honest, with member-aware membership CTAs.

**Architecture:** Three concentric layers, bottom-up. (1) dx-api exposes one new field (`vipDueAt`) on `DashboardProfile`. (2) dx-mini grows 11 new Lucide icons plus two small utilities (`utils/in-view.ts` for scroll-triggered one-shot reveals, `daysUntil()` in `utils/format.ts` for the 续费 countdown). (3) Six new self-contained components under `components/home/` — one per section — with one shared motion pattern (CSS `@keyframes` gated by `.is-in-view` set via `wx.createIntersectionObserver` with a 1000 ms fallback). `home.ts` / `home.wxml` / `home.json` / `home.wxss` are edited last to wire it all in and add the second dashboard-parallel stat fetch.

**Tech Stack:** Go 1.21+ / Goravel (dx-api); WeChat Mini Program native (glass-easel + Skyline), TypeScript strict, Vant Weapp 1.11.x (not used inside new components), `<dx-icon>` (Lucide SVG renderer).

**Spec:** [2026-04-20-mini-home-marketing-sections-design.md](../specs/2026-04-20-mini-home-marketing-sections-design.md)

**Branch:** `feat/mini-home-marketing-sections` (create on Task 1 start from current `main`).

**Commit convention (per user's git-workflow rule):**
- Tag mini-only commits with `(mini)`, api-only with `(api)`.
- No attribution footer — disabled globally via `~/.claude/settings.json`.
- Never push the feature branch. Merge locally to `main` at the end, then push `main` only.

**Constraint reminders:**
- **No `?.` or `??`** anywhere in TS or WXML. Use explicit null checks (`x !== null`) or `||`.
- **No continuous CSS loops or `animation-iteration-count: infinite`.** Every animation fires once on `.is-in-view`.
- **No `wx.createAnimation`**. CSS `@keyframes` only.
- **`<dx-icon>` is the only icon primitive.** Every new icon goes through `scripts/build-icons.mjs` first.

---

## Task 1: dx-api — expose `vipDueAt` on the dashboard

**Files:**
- Modify: `dx-api/app/services/api/hall_service.go`

- [ ] **Step 1: Create the feature branch**

From repo root:

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source checkout -b feat/mini-home-marketing-sections
```

Expected: `Switched to a new branch 'feat/mini-home-marketing-sections'`.

- [ ] **Step 2: Add `VipDueAt` field to `DashboardProfile` struct**

Edit `dx-api/app/services/api/hall_service.go`. Locate the `DashboardProfile` struct (currently lines 24–37). Append the new field after `CreatedAt`:

```go
// DashboardProfile is the user profile subset shown on the dashboard.
type DashboardProfile struct {
	ID                string           `json:"id"`
	Username          string           `json:"username"`
	Nickname          *string          `json:"nickname"`
	Grade             string           `json:"grade"`
	Level             int              `json:"level"`
	Exp               int              `json:"exp"`
	Beans             int              `json:"beans"`
	AvatarURL         *string          `json:"avatarUrl"`
	CurrentPlayStreak int              `json:"currentPlayStreak"`
	InviteCode        string           `json:"inviteCode"`
	LastReadNoticeAt  any              `json:"lastReadNoticeAt"`
	CreatedAt         any              `json:"createdAt"`
	VipDueAt          *carbon.DateTime `json:"vipDueAt"`
}
```

- [ ] **Step 3: Add the `carbon` import**

At the top of `dx-api/app/services/api/hall_service.go`, the existing imports are:

```go
import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)
```

Add `"github.com/goravel/framework/support/carbon"` to the second group (third-party), so the import block becomes:

```go
import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)
```

- [ ] **Step 4: Populate `VipDueAt` in `GetDashboard`**

In the same file, inside `GetDashboard` (around lines 93–106), append the `VipDueAt` line to the struct literal:

```go
	profile := DashboardProfile{
		ID:                user.ID,
		Username:          user.Username,
		Nickname:          user.Nickname,
		Grade:             user.Grade,
		Level:             level,
		Exp:               user.Exp,
		Beans:             user.Beans,
		AvatarURL:         user.AvatarURL,
		CurrentPlayStreak: user.CurrentPlayStreak,
		InviteCode:        user.InviteCode,
		LastReadNoticeAt:  user.LastReadNoticeAt,
		CreatedAt:         user.CreatedAt,
		VipDueAt:          user.VipDueAt,
	}
```

Note: `carbon` must end up referenced; if Go complains about an unused import, verify the `VipDueAt *carbon.DateTime` type reference is present in the struct.

- [ ] **Step 5: Verify the package compiles**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: no output (success).

- [ ] **Step 6: Run the test suite with race detector**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```

Expected: all packages PASS.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/hall_service.go
git commit -m "feat(api): expose vipDueAt on hall dashboard"
```

---

## Task 2: dx-mini — add 11 new Lucide icons

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Regenerate: `dx-mini/miniprogram/components/dx-icon/icons.ts`

Prerequisite: all 11 Lucide filenames (`keyboard.svg`, `swords.svg`, `shuffle.svg`, `crosshair.svg`, `sparkles.svg`, `coins.svg`, `refresh-cw.svg`, `circle-check.svg`, `message-square.svg`, `flame.svg`, `arrow-right.svg`) exist in `dx-mini/node_modules/lucide-static/icons/`. Verified: `ls dx-mini/node_modules/lucide-static/icons/ | grep …` returns 11 matches.

- [ ] **Step 1: Append 11 rows to the ICONS array**

Edit `dx-mini/scripts/build-icons.mjs`. The current `ICONS` array ends at `['flag', 'flag'],`. Append 11 rows before the closing `]`:

```js
  ['keyboard',       'keyboard'],
  ['swords',         'swords'],
  ['shuffle',        'shuffle'],
  ['crosshair',      'crosshair'],
  ['sparkles',       'sparkles'],
  ['coins',          'coins'],
  ['refresh-cw',     'refresh-cw'],
  ['circle-check',   'circle-check'],
  ['message-square', 'message-square'],
  ['flame',          'flame'],
  ['arrow-right',    'arrow-right'],
```

The final array should end with these 11 rows immediately before `]`.

- [ ] **Step 2: Regenerate icons.ts**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected output: `Wrote 36 icons to miniprogram/components/dx-icon/icons.ts.`

If the script throws `lucide-static is missing "<name>.svg"`, a version mismatch exists — inspect `dx-mini/node_modules/lucide-static/icons/` directly and adjust the lucide filename column (second column) to whatever the SVG file is actually called. Lucide has renamed some glyphs across versions (e.g. `home` → `house` in 0.460); a similar rename on a new icon would need the same two-column treatment.

- [ ] **Step 3: Verify icons.ts contains the 11 new entries**

```bash
grep -cE '"(keyboard|swords|shuffle|crosshair|sparkles|coins|refresh-cw|circle-check|message-square|flame|arrow-right)":' /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/components/dx-icon/icons.ts
```

Expected: `11`.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/scripts/build-icons.mjs dx-mini/miniprogram/components/dx-icon/icons.ts
git commit -m "chore(mini): add 11 marketing-home Lucide icons"
```

---

## Task 3: dx-mini — scroll-observer helper + `daysUntil` formatter

**Files:**
- Create: `dx-mini/miniprogram/utils/in-view.ts`
- Modify: `dx-mini/miniprogram/utils/format.ts`

- [ ] **Step 1: Create `utils/in-view.ts`**

Write `dx-mini/miniprogram/utils/in-view.ts`:

```typescript
// Fire a callback once when the first element matching `selector` enters
// the viewport. The observer auto-disconnects after firing. Intended for
// one-shot "on reveal" animations triggered by scroll.
//
// Usage from a Component's `attached` lifetime:
//   observeOnce(this, '.section-root', () => this.setData({ inView: true }))
//
// Every caller should also install a setTimeout fallback (~1000ms) that
// flips `inView: true` unconditionally, in case createIntersectionObserver
// misbehaves on older clients or in 真机调试.

export function observeOnce(
  component: WechatMiniprogram.Component.TrivialInstance,
  selector: string,
  cb: () => void,
  threshold = 0.15,
): WechatMiniprogram.IntersectionObserver {
  const io = component.createIntersectionObserver({ thresholds: [threshold] })
  io.relativeToViewport().observe(selector, (res) => {
    if (res.intersectionRatio > 0) {
      cb()
      io.disconnect()
    }
  })
  return io
}
```

- [ ] **Step 2: Add `daysUntil` to `utils/format.ts` and fix `gradeLabel` map**

The existing `gradeLabel` map uses `monthly`/`quarterly`/`yearly` but the backend returns `month`/`season`/`year` (per `dx-api/app/consts/user_grade.go`). This is a pre-existing bug — the current hub home falls through the `|| grade` fallback, rendering the raw string. We fix it while we're adding `daysUntil`.

Edit `dx-mini/miniprogram/utils/format.ts`. Replace the file entirely with:

```typescript
export function formatDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function formatRelativeDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  if (days < 30) return `${Math.floor(days / 7)}周前`
  if (days < 365) return `${Math.floor(days / 30)}个月前`
  return `${Math.floor(days / 365)}年前`
}

export function formatNumber(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  return String(n)
}

export function gradeLabel(grade: string): string {
  const map: Record<string, string> = {
    free: '免费',
    month: '月度会员',
    season: '季度会员',
    year: '年度会员',
    lifetime: '终身会员',
  }
  return map[grade] || grade
}

// Days between `now` and the ISO date string. Returns 0 for null/expired.
// Used by the membership section to render "续费 · 还剩 N 天" on the
// user's current tier button.
export function daysUntil(isoDate: string | null | undefined): number {
  if (!isoDate) return 0
  const target = new Date(isoDate).getTime()
  const now = Date.now()
  return Math.max(0, Math.ceil((target - now) / 86400000))
}
```

- [ ] **Step 3: Type-check the miniprogram**

WeChat Mini Program projects don't usually run a separate `tsc` step — the DevTools build does it. But during development you can sanity-check with:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors beyond the pre-existing tolerated `this`-in-`Component` warnings. If a fresh error cites `utils/in-view.ts` or `utils/format.ts`, something in the two files above is wrong.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/utils/in-view.ts dx-mini/miniprogram/utils/format.ts
git commit -m "feat(mini): add in-view helper and daysUntil formatter"
```

---

## Task 4: dx-mini — `<home-why-different>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/why-different/index.json`
- `dx-mini/miniprogram/components/home/why-different/index.ts`
- `dx-mini/miniprogram/components/home/why-different/index.wxml`
- `dx-mini/miniprogram/components/home/why-different/index.wxss`

This task establishes the **shared component pattern** used by Tasks 5–9. Each later component repeats the same four-file shape.

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/why-different/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/why-different/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'

Component({
  options: {
    addGlobalClass: true,
  },
  data: {
    inView: false,
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      // Fallback: if the observer hasn't fired after 1s (older clients,
      // 真机调试 quirks), reveal unconditionally so the section isn't blank.
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
})
```

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/why-different/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <text class="tag">为什么是斗学</text>
  <text class="headline">不再死记硬背,<text class="hl">让大脑爱上英语</text></text>

  <view class="row-card reveal delay-0">
    <text class="before">背了就忘,靠意志力硬撑</text>
    <view class="arrow">
      <dx-icon name="arrow-right" size="14px" color="#0d9488" customStyle="transform: rotate(90deg);" />
    </view>
    <text class="after">游戏化循环,大脑自发想再玩一局</text>
  </view>

  <view class="row-card reveal delay-1">
    <text class="before">学的和用的两张皮</text>
    <view class="arrow">
      <dx-icon name="arrow-right" size="14px" color="#0d9488" customStyle="transform: rotate(90deg);" />
    </view>
    <text class="after">连词成句、对话、对战都是真实语料</text>
  </view>

  <view class="row-card reveal delay-2">
    <text class="before">一个人孤独地学</text>
    <view class="arrow">
      <dx-icon name="arrow-right" size="14px" color="#0d9488" customStyle="transform: rotate(90deg);" />
    </view>
    <text class="after">好友开黑 · 学习群 · 排行榜 · 每日连胜</text>
  </view>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/why-different/index.wxss`:

```css
.section-root {
  padding: 40rpx 32rpx 32rpx;
  background: var(--bg-page);
}

.accent-dash {
  width: 48rpx;
  height: 6rpx;
  border-radius: 3rpx;
  background: #0d9488;
  margin-bottom: 20rpx;
}

.tag {
  display: block;
  font-size: 22rpx;
  color: #0d9488;
  font-weight: 700;
  letter-spacing: 2rpx;
  text-transform: uppercase;
  margin-bottom: 12rpx;
}

.headline {
  display: block;
  font-size: 34rpx;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1.35;
  margin-bottom: 24rpx;
}

.headline .hl {
  color: #0d9488;
}

.row-card {
  background: var(--bg-card);
  border: 1rpx solid var(--border-color);
  border-radius: 20rpx;
  padding: 24rpx 28rpx;
  margin-bottom: 16rpx;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10rpx;
}

.before {
  font-size: 26rpx;
  color: var(--text-secondary);
  text-decoration: line-through;
  text-decoration-color: rgba(100, 116, 139, 0.4);
}

.arrow {
  display: flex;
  justify-content: center;
  margin: 4rpx 0;
}

.after {
  font-size: 28rpx;
  font-weight: 600;
  color: var(--text-primary);
  background: linear-gradient(135deg, rgba(13, 148, 136, 0.08), rgba(124, 58, 237, 0.08));
  padding: 12rpx 16rpx;
  border-radius: 12rpx;
}

/* ----- Reveal motion (one-shot) ----- */

.reveal {
  opacity: 0;
  transform: translateY(14rpx);
}

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/why-different/
git commit -m "feat(mini): add why-different home section"
```

---

## Task 5: dx-mini — `<home-features>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/features/index.json`
- `dx-mini/miniprogram/components/home/features/index.ts`
- `dx-mini/miniprogram/components/home/features/index.wxml`
- `dx-mini/miniprogram/components/home/features/index.wxss`

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/features/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/features/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'

interface RecentSession {
  gameId: string
  gameName: string
  completedLevels: number
}

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    recentSession: {
      type: null,
      value: null,
    },
  },
  data: {
    inView: false,
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
  methods: {
    goResume() {
      const session = (this.data as any).recentSession as RecentSession | null
      if (!session) return
      wx.navigateTo({ url: `/pages/games/detail/detail?id=${session.gameId}` })
    },
  },
})
```

Note on typing: properties are accessible at runtime via `this.data.<prop>`, but the TypeScript typings for `Component({properties: ...})` don't auto-merge property names into `this.data`'s type. The `(this.data as any).recentSession` pattern is the established pragmatic workaround in this codebase's mini TS — no need to duplicate the prop into `data:` just to appease the type checker.

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/features/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <text class="tag">核心玩法 · 4 种模式,覆盖听说读写</text>
  <text class="headline">挑一个就能上手,<text class="hl">全都玩转就起飞</text></text>

  <view class="resume-chip" wx:if="{{recentSession}}" bind:tap="goResume">
    <text class="resume-label">继续 · {{recentSession.gameName}} Lv.{{recentSession.completedLevels + 1}}</text>
    <dx-icon name="chevron-right" size="14px" color="#0d9488" />
  </view>

  <!-- Card 1: 连词成句 -->
  <view class="game-card reveal delay-0">
    <view class="card-head">
      <view class="ico ico-violet"><dx-icon name="keyboard" size="18px" color="#7c3aed" /></view>
      <view class="card-text">
        <text class="name">连词成句</text>
        <text class="desc">看到中文秒拼出英文句子,越快越高分。真实语料替你练习语序和搭配。</text>
      </view>
    </view>
    <view class="illus illus-tiles">
      <text class="tile">I</text>
      <text class="tile">love</text>
      <text class="tile">learning</text>
      <text class="tile">English</text>
      <text class="tile tile-accent">→</text>
    </view>
  </view>

  <!-- Card 2: 词汇配对 -->
  <view class="game-card reveal delay-1">
    <view class="card-head">
      <view class="ico ico-blue"><dx-icon name="swords" size="18px" color="#0d9488" /></view>
      <view class="card-text">
        <text class="name">词汇配对</text>
        <text class="desc">英文与中文快速配对,限时给分。巩固词汇量和中译英反应速度。</text>
      </view>
    </view>
    <view class="illus illus-match">
      <text class="tile tile-fill">negotiate</text>
      <view class="line"></view>
      <text class="tile">谈判</text>
    </view>
  </view>

  <!-- Card 3: 词汇消消乐 -->
  <view class="game-card reveal delay-2">
    <view class="card-head">
      <view class="ico ico-pink"><dx-icon name="shuffle" size="18px" color="#059669" /></view>
      <view class="card-text">
        <text class="name">词汇消消乐</text>
        <text class="desc">记忆配对消除,越快消除越高分。玩着玩着就把生词牢牢记住。</text>
      </view>
    </view>
    <view class="illus illus-grid">
      <view class="cell c-pink"></view>
      <view class="cell c-teal-light"></view>
      <view class="cell c-violet"></view>
      <view class="cell c-teal-light"></view>
      <view class="cell c-teal"></view>
      <view class="cell c-pink"></view>
      <view class="cell c-violet"></view>
      <view class="cell c-pink"></view>
      <view class="cell c-teal-light"></view>
    </view>
  </view>

  <!-- Card 4: 词汇对轰 -->
  <view class="game-card reveal delay-3">
    <view class="card-head">
      <view class="ico ico-red"><dx-icon name="crosshair" size="18px" color="#dc2626" /></view>
      <view class="card-text">
        <text class="name">词汇对轰</text>
        <text class="desc">和对手拼炮弹。拼对拼快就发射,紧张刺激的词汇对战。</text>
      </view>
    </view>
    <view class="illus illus-bars">
      <view class="bar" style="height: 60%;"></view>
      <view class="bar bar-teal" style="height: 40%;"></view>
      <view class="bar" style="height: 80%;"></view>
      <view class="bar bar-teal" style="height: 50%;"></view>
      <view class="bar" style="height: 70%;"></view>
    </view>
  </view>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/features/index.wxss`:

```css
.section-root {
  padding: 40rpx 32rpx 32rpx;
  background: var(--bg-page);
}

.accent-dash {
  width: 48rpx; height: 6rpx; border-radius: 3rpx;
  background: #7c3aed; margin-bottom: 20rpx;
}

.tag {
  display: block; font-size: 22rpx; color: #7c3aed;
  font-weight: 700; letter-spacing: 2rpx;
  margin-bottom: 12rpx;
}

.headline {
  display: block; font-size: 34rpx; font-weight: 800;
  color: var(--text-primary); line-height: 1.35; margin-bottom: 20rpx;
}

.headline .hl { color: #7c3aed; }

/* ----- Resume chip ----- */

.resume-chip {
  display: flex; align-items: center; justify-content: space-between;
  background: var(--bg-card); border: 1rpx solid #0d9488;
  border-radius: 999rpx; padding: 14rpx 22rpx; margin-bottom: 20rpx;
}

.resume-label {
  font-size: 24rpx; color: #0d9488; font-weight: 600;
  flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* ----- Game card ----- */

.game-card {
  background: var(--bg-card); border: 1rpx solid var(--border-color);
  border-radius: 20rpx; padding: 24rpx; margin-bottom: 16rpx;
}

.card-head { display: flex; gap: 16rpx; align-items: flex-start; }

.ico {
  width: 56rpx; height: 56rpx; border-radius: 14rpx;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.ico-violet { background: #ede9fe; }
.ico-blue   { background: #dbeafe; }
.ico-pink   { background: #fce7f3; }
.ico-red    { background: #fee2e2; }

.card-text { flex: 1; min-width: 0; }

.name {
  display: block; font-size: 28rpx; font-weight: 700;
  color: var(--text-primary); margin-bottom: 4rpx;
}

.desc {
  display: block; font-size: 22rpx; color: var(--text-secondary);
  line-height: 1.5;
}

/* ----- Illustration ----- */

.illus {
  margin-top: 18rpx; height: 96rpx; background: rgba(241, 245, 249, 0.6);
  border-radius: 12rpx; padding: 0 16rpx; display: flex; align-items: center;
  gap: 8rpx; overflow: hidden;
}

.page-container.dark .illus { background: rgba(255, 255, 255, 0.04); }

/* Tiles (sentence + match) */

.tile {
  background: var(--bg-card); border: 1rpx solid rgba(148, 163, 184, 0.35);
  border-radius: 8rpx; padding: 6rpx 12rpx;
  font-size: 20rpx; font-weight: 600; color: var(--text-primary);
}

.tile-accent { background: rgba(13, 148, 136, 0.15); border-color: #0d9488; color: #0d9488; }
.tile-fill   { background: rgba(13, 148, 136, 0.15); border-color: #0d9488; color: #0d9488; }

/* Match line */

.illus-match .line {
  width: 48rpx; height: 4rpx; background: #0d9488; border-radius: 2rpx;
  transform: scaleX(0); transform-origin: left center;
}

.section-root.is-in-view .illus-match .line {
  animation: line-draw 0.5s ease-out 0.35s both;
}

@keyframes line-draw { to { transform: scaleX(1); } }

/* Grid (eliminate) */

.illus-grid { display: grid; grid-template-columns: repeat(9, 1fr); gap: 6rpx; padding: 8rpx 16rpx; }

.cell { width: 100%; height: 20rpx; border-radius: 4rpx; transform: scale(0.4); opacity: 0; }

.c-pink       { background: #ec4899; }
.c-teal-light { background: rgba(13, 148, 136, 0.25); }
.c-violet     { background: #7c3aed; }
.c-teal       { background: #0d9488; }

.section-root.is-in-view .illus-grid .cell {
  animation: cell-pop 0.4s ease-out both;
}

.section-root.is-in-view .illus-grid .cell:nth-child(1) { animation-delay: 350ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(2) { animation-delay: 400ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(3) { animation-delay: 450ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(4) { animation-delay: 500ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(5) { animation-delay: 550ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(6) { animation-delay: 600ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(7) { animation-delay: 650ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(8) { animation-delay: 700ms; }
.section-root.is-in-view .illus-grid .cell:nth-child(9) { animation-delay: 750ms; }

@keyframes cell-pop { to { transform: scale(1); opacity: 1; } }

/* Bars (battle) */

.illus-bars { height: 96rpx; align-items: flex-end; padding: 8rpx 0; gap: 10rpx; justify-content: center; }

.bar { width: 10rpx; background: #ef4444; border-radius: 3rpx; transform: scaleY(0); transform-origin: bottom center; }

.bar.bar-teal { background: #0d9488; }

.section-root.is-in-view .illus-bars .bar {
  animation: bar-rise 0.45s ease-out both;
}

.section-root.is-in-view .illus-bars .bar:nth-child(1) { animation-delay: 350ms; }
.section-root.is-in-view .illus-bars .bar:nth-child(2) { animation-delay: 420ms; }
.section-root.is-in-view .illus-bars .bar:nth-child(3) { animation-delay: 490ms; }
.section-root.is-in-view .illus-bars .bar:nth-child(4) { animation-delay: 560ms; }
.section-root.is-in-view .illus-bars .bar:nth-child(5) { animation-delay: 630ms; }

@keyframes bar-rise { to { transform: scaleY(1); } }

/* Sentence tiles reveal */

.illus-tiles .tile { opacity: 0; transform: translateY(8rpx); }

.section-root.is-in-view .illus-tiles .tile {
  animation: tile-up 0.35s ease-out both;
}

.section-root.is-in-view .illus-tiles .tile:nth-child(1) { animation-delay: 350ms; }
.section-root.is-in-view .illus-tiles .tile:nth-child(2) { animation-delay: 430ms; }
.section-root.is-in-view .illus-tiles .tile:nth-child(3) { animation-delay: 510ms; }
.section-root.is-in-view .illus-tiles .tile:nth-child(4) { animation-delay: 590ms; }
.section-root.is-in-view .illus-tiles .tile:nth-child(5) { animation-delay: 670ms; }

@keyframes tile-up { to { opacity: 1; transform: translateY(0); } }

/* ----- Section reveal (cards) ----- */

.reveal { opacity: 0; transform: translateY(14rpx); }

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }
.section-root.is-in-view .reveal.delay-3 { animation-delay: 240ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/features/
git commit -m "feat(mini): add features home section with static illustrations"
```

---

## Task 6: dx-mini — `<home-ai-features>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/ai-features/index.json`
- `dx-mini/miniprogram/components/home/ai-features/index.ts`
- `dx-mini/miniprogram/components/home/ai-features/index.wxml`
- `dx-mini/miniprogram/components/home/ai-features/index.wxss`

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/ai-features/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/ai-features/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    beans: {
      type: Number,
      value: 0,
    },
  },
  data: {
    inView: false,
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
})
```

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/ai-features/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <view class="tag-row">
    <dx-icon name="sparkles" size="14px" color="#0d9488" />
    <text class="tag">AI 驱动 · 专属于你的学习</text>
  </view>
  <text class="headline"><text class="hl">AI</text> 帮你定制课程,你想学什么都可以</text>

  <view class="bullet reveal delay-0">输入任意主题或场景,AI 按你的水平生成课程</view>
  <view class="bullet reveal delay-1">CEFR A1–C2 全覆盖,难度智能匹配</view>
  <view class="bullet reveal delay-2">内容沉淀进你的词汇系统,复习自动推送</view>

  <view class="demo-card reveal delay-3">
    <text class="input-label">输入主题:<text class="hl">职场面试高频词</text></text>
    <view class="stream">
      <text class="line">› negotiate · 谈判 — Let's negotiate the salary.</text>
      <text class="line">› résumé · 简历 — Please send me your résumé.</text>
      <text class="line">› confident · 自信 — Stay confident during the interview.</text>
      <text class="line">› leverage · 优势 — Leverage your strengths.</text>
      <text class="line">› follow up · 跟进 — I'll follow up next week.</text>
    </view>
    <view class="footer">
      <dx-icon name="coins" size="13px" color="#d97706" />
      <text class="footer-text">你当前有 <text class="hl">{{beans}}</text> 能量豆 · 每次消耗 5 颗,失败全额退还</text>
    </view>
  </view>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/ai-features/index.wxss`:

```css
.section-root {
  padding: 56rpx 32rpx 48rpx;
  background: linear-gradient(180deg, var(--bg-page), rgba(13, 148, 136, 0.08), var(--bg-page));
}

.page-container.dark .section-root {
  background: linear-gradient(180deg, var(--bg-page), rgba(13, 148, 136, 0.12), var(--bg-page));
}

.accent-dash {
  width: 48rpx; height: 6rpx; border-radius: 3rpx;
  background: #0d9488; margin-bottom: 20rpx;
}

.tag-row {
  display: flex; align-items: center; gap: 8rpx; margin-bottom: 12rpx;
}

.tag {
  font-size: 22rpx; color: #0d9488; font-weight: 700; letter-spacing: 2rpx;
}

.headline {
  display: block; font-size: 34rpx; font-weight: 800;
  color: var(--text-primary); line-height: 1.35; margin-bottom: 24rpx;
}

.headline .hl { color: #0d9488; }

.bullet {
  font-size: 26rpx; color: var(--text-primary); line-height: 1.6;
  padding-left: 20rpx; position: relative; margin-bottom: 10rpx;
}

.bullet::before {
  content: ''; position: absolute; left: 0; top: 20rpx;
  width: 8rpx; height: 8rpx; border-radius: 50%; background: #0d9488;
}

.demo-card {
  margin-top: 24rpx; background: #0f172a; border-radius: 16rpx;
  padding: 24rpx; font-family: 'SF Mono', Menlo, monospace;
  box-shadow: 0 8rpx 24rpx rgba(13, 148, 136, 0.15);
}

.input-label {
  display: block; font-size: 22rpx; color: #94a3b8;
  padding-bottom: 14rpx; border-bottom: 1rpx solid rgba(255, 255, 255, 0.08);
  margin-bottom: 16rpx;
}

.input-label .hl { color: #67e8f9; }

.stream { display: flex; flex-direction: column; gap: 8rpx; min-height: 200rpx; }

.line {
  display: block; font-size: 22rpx; color: #e2e8f0;
  white-space: nowrap; overflow: hidden; width: 0;
  border-right: 3rpx solid #0d9488;
}

.section-root.is-in-view .stream .line:nth-child(1) {
  animation: typer 1100ms steps(44) 240ms both;
}
.section-root.is-in-view .stream .line:nth-child(2) {
  animation: typer 1100ms steps(42) 1500ms both;
}
.section-root.is-in-view .stream .line:nth-child(3) {
  animation: typer 1100ms steps(50) 2700ms both;
}
.section-root.is-in-view .stream .line:nth-child(4) {
  animation: typer 1100ms steps(38) 3900ms both;
}
.section-root.is-in-view .stream .line:nth-child(5) {
  animation: typer 1100ms steps(42) 5000ms both;
}

@keyframes typer {
  from { width: 0; border-right-color: #0d9488; }
  99%  { border-right-color: #0d9488; }
  to   { width: 100%; border-right-color: transparent; }
}

.footer {
  margin-top: 18rpx; padding-top: 14rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.08);
  display: flex; align-items: center; gap: 8rpx;
}

.footer-text {
  font-size: 20rpx; color: #94a3b8; font-family: -apple-system, sans-serif;
}

.footer-text .hl { color: #fbbf24; font-weight: 700; }

/* ----- Reveal ----- */

.reveal { opacity: 0; transform: translateY(14rpx); }

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }
.section-root.is-in-view .reveal.delay-3 { animation-delay: 240ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/ai-features/
git commit -m "feat(mini): add ai-features home section with live bean count"
```

---

## Task 7: dx-mini — `<home-learning-loop>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/learning-loop/index.json`
- `dx-mini/miniprogram/components/home/learning-loop/index.ts`
- `dx-mini/miniprogram/components/home/learning-loop/index.wxml`
- `dx-mini/miniprogram/components/home/learning-loop/index.wxss`

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/learning-loop/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/learning-loop/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    unknownTotal: {
      type: null,
      value: null,
    },
    reviewPending: {
      type: null,
      value: null,
    },
    masterTotal: {
      type: null,
      value: null,
    },
  },
  data: {
    inView: false,
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
  methods: {
    goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
    goReview()  { wx.navigateTo({ url: '/pages/learn/review/review' }) },
    goMastered(){ wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  },
})
```

Note: `type: null` is the mini-program way to accept any type for a property. We need this because the values can be `number | null`.

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/learning-loop/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <text class="tag">学习闭环 · 从陌生到掌握</text>
  <text class="headline">一套系统,<text class="hl">追踪你每一个学习单元的命运</text></text>

  <view class="book-card reveal delay-0" bind:tap="goUnknown">
    <view class="card-head">
      <view class="ico ico-pink"><dx-icon name="book-open" size="18px" color="#ec4899" /></view>
      <view class="card-text">
        <text class="name">生词本</text>
        <text class="count">{{unknownTotal !== null ? unknownTotal : '—'}} 条</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
    <text class="desc">持续沉淀你不会的词汇</text>
  </view>

  <view class="book-card reveal delay-1" bind:tap="goReview">
    <view class="card-head">
      <view class="ico ico-violet"><dx-icon name="refresh-cw" size="18px" color="#7c3aed" /></view>
      <view class="card-text">
        <text class="name">复习本</text>
        <text class="count">{{reviewPending !== null ? reviewPending : '—'}} 待复习</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
    <text class="desc">[1, 3, 7, 14, 30, 90] 天节奏智能推送</text>
  </view>

  <view class="book-card reveal delay-2" bind:tap="goMastered">
    <view class="card-head">
      <view class="ico ico-teal"><dx-icon name="circle-check" size="18px" color="#0d9488" /></view>
      <view class="card-text">
        <text class="name">已掌握</text>
        <text class="count">{{masterTotal !== null ? masterTotal : '—'}} 条</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
    <text class="desc">看得见的词汇量增长</text>
  </view>

  <text class="footnote">艾宾浩斯遗忘曲线智能推送复习,你只需要玩。</text>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/learning-loop/index.wxss`:

```css
.section-root {
  padding: 56rpx 32rpx 48rpx;
  background: linear-gradient(180deg, rgba(13, 148, 136, 0.08), rgba(124, 58, 237, 0.08));
}

.page-container.dark .section-root {
  background: linear-gradient(180deg, rgba(13, 148, 136, 0.14), rgba(124, 58, 237, 0.14));
}

.accent-dash {
  width: 48rpx; height: 6rpx; border-radius: 3rpx;
  background: #0d9488; margin-bottom: 20rpx;
}

.tag {
  display: block; font-size: 22rpx; color: #0d9488;
  font-weight: 700; letter-spacing: 2rpx; margin-bottom: 12rpx;
}

.headline {
  display: block; font-size: 34rpx; font-weight: 800;
  color: var(--text-primary); line-height: 1.35; margin-bottom: 24rpx;
}

.headline .hl { color: #7c3aed; }

.book-card {
  background: var(--bg-card); border: 1rpx solid var(--border-color);
  border-radius: 20rpx; padding: 24rpx; margin-bottom: 16rpx;
}

.card-head { display: flex; align-items: center; gap: 16rpx; }

.ico {
  width: 56rpx; height: 56rpx; border-radius: 14rpx;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.ico-pink   { background: rgba(236, 72, 153, 0.12); }
.ico-violet { background: rgba(124, 58, 237, 0.12); }
.ico-teal   { background: rgba(13, 148, 136, 0.12); }

.card-text { flex: 1; min-width: 0; }

.name {
  display: block; font-size: 28rpx; font-weight: 700;
  color: var(--text-primary);
}

.count {
  display: block; font-size: 24rpx; color: #0d9488; font-weight: 600; margin-top: 2rpx;
}

.desc {
  display: block; font-size: 22rpx; color: var(--text-secondary);
  line-height: 1.5; margin-top: 10rpx;
}

.footnote {
  display: block; font-size: 22rpx; color: var(--text-secondary);
  text-align: center; margin-top: 20rpx; line-height: 1.6;
}

/* Reveal */

.reveal { opacity: 0; transform: translateY(14rpx); }

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/learning-loop/
git commit -m "feat(mini): add learning-loop home section with live counts"
```

---

## Task 8: dx-mini — `<home-community>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/community/index.json`
- `dx-mini/miniprogram/components/home/community/index.ts`
- `dx-mini/miniprogram/components/home/community/index.wxml`
- `dx-mini/miniprogram/components/home/community/index.wxss`

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/community/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/community/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    streak: {
      type: Number,
      value: 0,
    },
  },
  data: {
    inView: false,
    heatCells: [false, false, false, false, false, false, false],
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
  observers: {
    streak(streak: number) {
      const filled = Math.max(0, Math.min(streak || 0, 7))
      const cells = [false, false, false, false, false, false, false]
      for (let i = 0; i < filled; i++) cells[i] = true
      this.setData({ heatCells: cells })
    },
  },
  methods: {
    goLeaderboard() { wx.navigateTo({ url: '/pages/leaderboard/leaderboard' }) },
    goCommunity()   { wx.navigateTo({ url: '/pages/me/community/community' }) },
    goGroups()      { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  },
})
```

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/community/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <text class="tag">一起玩才好玩</text>
  <text class="headline">和朋友开黑,<text class="hl">排行榜上见</text></text>

  <view class="feat-card reveal delay-0" bind:tap="goLeaderboard">
    <view class="card-head">
      <view class="ico ico-amber"><dx-icon name="trophy" size="18px" color="#f59e0b" /></view>
      <view class="card-text">
        <text class="name">排行榜</text>
        <text class="desc">经验值与在线时长,按日、周、月六种榜单,登上领奖台。</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
  </view>

  <view class="feat-card reveal delay-1" bind:tap="goCommunity">
    <view class="card-head">
      <view class="ico ico-orange"><dx-icon name="message-square" size="18px" color="#ea580c" /></view>
      <view class="card-text">
        <text class="name">斗学社</text>
        <text class="desc">发帖、评论、点赞、关注,把学习心得变成社交动态。</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
  </view>

  <view class="feat-card reveal delay-2" bind:tap="goGroups">
    <view class="card-head">
      <view class="ico ico-blue"><dx-icon name="users" size="18px" color="#3b82f6" /></view>
      <view class="card-text">
        <text class="name">学习群</text>
        <text class="desc">组建小组一起闯关,群内可直接发起课程对战。</text>
      </view>
      <dx-icon name="chevron-right" size="14px" color="#94a3b8" />
    </view>
  </view>

  <view class="streak-card reveal delay-3">
    <view class="card-head">
      <view class="ico ico-teal"><dx-icon name="flame" size="18px" color="#0d9488" /></view>
      <text class="name">连续打卡</text>
    </view>
    <text class="streak-number">你已连胜 <text class="hl">{{streak}}</text> 天</text>
    <view class="heatmap">
      <view class="cell {{item ? 'filled' : ''}}" wx:for="{{heatCells}}" wx:key="index"></view>
    </view>
    <text class="desc">每天玩至少一次,保持连胜;错过一天从 1 重来。</text>
  </view>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/community/index.wxss`:

```css
.section-root {
  padding: 40rpx 32rpx 32rpx;
  background: var(--bg-page);
}

.accent-dash {
  width: 48rpx; height: 6rpx; border-radius: 3rpx;
  background: #f59e0b; margin-bottom: 20rpx;
}

.tag {
  display: block; font-size: 22rpx; color: #f59e0b;
  font-weight: 700; letter-spacing: 2rpx; margin-bottom: 12rpx;
}

.headline {
  display: block; font-size: 34rpx; font-weight: 800;
  color: var(--text-primary); line-height: 1.35; margin-bottom: 24rpx;
}

.headline .hl { color: #f59e0b; }

.feat-card, .streak-card {
  background: var(--bg-card); border: 1rpx solid var(--border-color);
  border-radius: 20rpx; padding: 24rpx; margin-bottom: 16rpx;
}

.card-head { display: flex; align-items: center; gap: 16rpx; }

.ico {
  width: 56rpx; height: 56rpx; border-radius: 14rpx;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.ico-amber  { background: rgba(245, 158, 11, 0.12); }
.ico-orange { background: rgba(234, 88, 12, 0.12); }
.ico-blue   { background: rgba(59, 130, 246, 0.12); }
.ico-teal   { background: rgba(13, 148, 136, 0.12); }

.card-text { flex: 1; min-width: 0; }

.name {
  display: block; font-size: 28rpx; font-weight: 700;
  color: var(--text-primary); margin-bottom: 4rpx;
}

.desc {
  display: block; font-size: 22rpx; color: var(--text-secondary);
  line-height: 1.5;
}

/* Streak card extras */

.streak-number {
  display: block; font-size: 26rpx; color: var(--text-primary);
  margin: 16rpx 0 14rpx; font-weight: 600;
}

.streak-number .hl {
  font-size: 48rpx; color: #0d9488; font-weight: 800;
  padding: 0 6rpx;
}

.heatmap {
  display: flex; gap: 8rpx; margin-bottom: 16rpx;
}

.cell {
  flex: 1; height: 32rpx; border-radius: 6rpx;
  background: var(--border-color);
}

.section-root.is-in-view .cell.filled {
  animation: fill 0.4s ease-out both;
}

.section-root.is-in-view .cell.filled:nth-child(1) { animation-delay: 400ms; }
.section-root.is-in-view .cell.filled:nth-child(2) { animation-delay: 480ms; }
.section-root.is-in-view .cell.filled:nth-child(3) { animation-delay: 560ms; }
.section-root.is-in-view .cell.filled:nth-child(4) { animation-delay: 640ms; }
.section-root.is-in-view .cell.filled:nth-child(5) { animation-delay: 720ms; }
.section-root.is-in-view .cell.filled:nth-child(6) { animation-delay: 800ms; }
.section-root.is-in-view .cell.filled:nth-child(7) { animation-delay: 880ms; }

@keyframes fill {
  from { background: var(--border-color); }
  to   { background: #0d9488; }
}

/* Reveal */

.reveal { opacity: 0; transform: translateY(14rpx); }

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }
.section-root.is-in-view .reveal.delay-3 { animation-delay: 240ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/community/
git commit -m "feat(mini): add community home section with live streak"
```

---

## Task 9: dx-mini — `<home-membership>` section

**Files (created):**
- `dx-mini/miniprogram/components/home/membership/index.json`
- `dx-mini/miniprogram/components/home/membership/index.ts`
- `dx-mini/miniprogram/components/home/membership/index.wxml`
- `dx-mini/miniprogram/components/home/membership/index.wxss`

This is the most complex section because of the member-aware CTA matrix. Copy the exact CTA matrix from §9.6 of the spec — it's the source of truth.

- [ ] **Step 1: Write `index.json`**

Create `dx-mini/miniprogram/components/home/membership/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `index.ts`**

Create `dx-mini/miniprogram/components/home/membership/index.ts`:

```typescript
import { observeOnce } from '../../../utils/in-view'
import { daysUntil } from '../../../utils/format'

type Grade = 'free' | 'month' | 'season' | 'year' | 'lifetime'

interface ButtonState {
  label: string
  disabled: boolean
  celebratory: boolean
}

interface TierStates {
  free: ButtonState
  month: ButtonState
  year: ButtonState
  lifetime: ButtonState
  currentExpiresIn: number  // 0 if not on a time-bounded tier
  showExpiryLine: boolean    // true for season/year, false otherwise
}

function computeStates(grade: Grade, vipDueAt: string | null): TierStates {
  const info: ButtonState = { label: '默认权益', disabled: true, celebratory: false }
  const open = (l: string): ButtonState => ({ label: l, disabled: false, celebratory: false })
  const off  = (l: string): ButtonState => ({ label: l, disabled: true,  celebratory: false })
  const done = (l: string): ButtonState => ({ label: l, disabled: true,  celebratory: true })

  const n = daysUntil(vipDueAt)

  if (grade === 'free') {
    return {
      free: info,
      month: open('立即开通'),
      year: open('立即开通'),
      lifetime: open('立即开通'),
      currentExpiresIn: 0,
      showExpiryLine: false,
    }
  }
  if (grade === 'month') {
    return {
      free: info,
      month: open(n > 0 ? `续费 · 还剩 ${n} 天` : '立即开通'),
      year: open('立即开通'),
      lifetime: open('立即开通'),
      currentExpiresIn: n,
      showExpiryLine: false,
    }
  }
  if (grade === 'season') {
    return {
      free: info,
      month: off('已包含'),
      year: open('升级到年度'),
      lifetime: open('升级到终身'),
      currentExpiresIn: n,
      showExpiryLine: true,
    }
  }
  if (grade === 'year') {
    return {
      free: info,
      month: off('已包含'),
      year: open(n > 0 ? `续费 · 还剩 ${n} 天` : '立即开通'),
      lifetime: open('升级到终身'),
      currentExpiresIn: n,
      showExpiryLine: true,
    }
  }
  if (grade === 'lifetime') {
    return {
      free: info,
      month: off('已包含'),
      year: off('已包含'),
      lifetime: done('✨ 已开通'),
      currentExpiresIn: 0,
      showExpiryLine: false,
    }
  }
  // Unknown grade: treat as free.
  return {
    free: info,
    month: open('立即开通'),
    year: open('立即开通'),
    lifetime: open('立即开通'),
    currentExpiresIn: 0,
    showExpiryLine: false,
  }
}

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    grade: {
      type: String,
      value: 'free',
    },
    vipDueAt: {
      type: String,
      value: '',
    },
  },
  data: {
    inView: false,
    states: computeStates('free', null),
  },
  lifetimes: {
    attached() {
      observeOnce(this, '.section-root', () => this.setData({ inView: true }))
      setTimeout(() => {
        if (!this.data.inView) this.setData({ inView: true })
      }, 1000)
    },
  },
  observers: {
    'grade, vipDueAt'(grade: string, vipDueAt: string) {
      const g = (grade || 'free') as Grade
      const vip = vipDueAt ? vipDueAt : null
      this.setData({ states: computeStates(g, vip) })
    },
  },
  methods: {
    goPurchase(e: WechatMiniprogram.CustomEvent) {
      const disabled = e.currentTarget.dataset.disabled
      if (disabled) return
      wx.navigateTo({ url: '/pages/me/purchase/purchase' })
    },
  },
})
```

- [ ] **Step 3: Write `index.wxml`**

Create `dx-mini/miniprogram/components/home/membership/index.wxml`:

```xml
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  <view class="accent-dash"></view>
  <text class="tag">会员计划</text>
  <text class="headline">选一个你最舒服的节奏,<text class="hl">越早开始越划算</text></text>
  <text class="desc">还有季度会员等更多选项,在会员页查看完整对比。</text>
  <text class="expiry-line" wx:if="{{states.showExpiryLine}}">你的会员还剩 {{states.currentExpiresIn}} 天</text>

  <!-- FREE -->
  <view class="tier-card tier-free reveal delay-0">
    <view class="tier-head">
      <text class="tier-name">免费会员</text>
      <text class="tier-price">¥0</text>
    </view>
    <view class="features">
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>部分关卡</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>基础游戏</text></view>
    </view>
    <view class="cta cta-info">{{states.free.label}}</view>
  </view>

  <!-- MONTH -->
  <view class="tier-card reveal delay-1">
    <view class="tier-head">
      <text class="tier-name">月度会员</text>
      <text class="tier-price">¥39<text class="unit">/月</text></text>
    </view>
    <view class="features">
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>全部关卡畅玩</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>AI 随心学</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>PK + 群组</text></view>
    </view>
    <view class="cta {{states.month.disabled ? 'cta-off' : 'cta-primary'}}"
          data-disabled="{{states.month.disabled}}"
          bind:tap="goPurchase">
      {{states.month.label}}
    </view>
  </view>

  <!-- YEAR (recommended) -->
  <view class="tier-card tier-year reveal delay-2">
    <view class="badge-rec">推荐</view>
    <view class="tier-head">
      <text class="tier-name">年度会员</text>
      <text class="tier-price">¥309<text class="unit">/年</text></text>
    </view>
    <view class="features">
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>超值优惠套餐</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>包含月度全部权益</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#0d9488" /><text>优先客服支持</text></view>
    </view>
    <view class="cta {{states.year.disabled ? 'cta-off' : 'cta-primary'}}"
          data-disabled="{{states.year.disabled}}"
          bind:tap="goPurchase">
      {{states.year.label}}
    </view>
  </view>

  <!-- LIFETIME -->
  <view class="tier-card tier-lifetime reveal delay-3">
    <view class="badge-life">最超值</view>
    <view class="tier-head">
      <text class="tier-name">终身会员</text>
      <text class="tier-price">¥1999<text class="unit">一次性</text></text>
    </view>
    <view class="features">
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#ffffff" /><text>一次付费,终身生效</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#ffffff" /><text>最高能量豆赠送</text></view>
      <view class="feat"><dx-icon name="circle-check" size="12px" color="#ffffff" /><text>邀请好友首充享 30% 返佣</text></view>
    </view>
    <view class="cta {{states.lifetime.disabled ? (states.lifetime.celebratory ? 'cta-done' : 'cta-off-dark') : 'cta-on-dark'}}"
          data-disabled="{{states.lifetime.disabled}}"
          bind:tap="goPurchase">
      {{states.lifetime.label}}
    </view>
  </view>
</view>
```

- [ ] **Step 4: Write `index.wxss`**

Create `dx-mini/miniprogram/components/home/membership/index.wxss`:

```css
.section-root {
  padding: 64rpx 32rpx 56rpx;
  background: linear-gradient(135deg, #7c3aed 0%, #0d9488 100%);
}

.accent-dash {
  width: 48rpx; height: 6rpx; border-radius: 3rpx;
  background: #ffffff; margin-bottom: 20rpx;
}

.tag {
  display: block; font-size: 22rpx; color: rgba(255, 255, 255, 0.85);
  font-weight: 700; letter-spacing: 2rpx; margin-bottom: 12rpx;
}

.headline {
  display: block; font-size: 34rpx; font-weight: 800;
  color: #ffffff; line-height: 1.35; margin-bottom: 10rpx;
}

.headline .hl { color: #a7f3d0; }

.desc {
  display: block; font-size: 22rpx; color: rgba(255, 255, 255, 0.8);
  line-height: 1.5; margin-bottom: 12rpx;
}

.expiry-line {
  display: block; font-size: 22rpx; color: #fef3c7;
  font-weight: 600; margin-bottom: 20rpx;
}

/* Tier card */

.tier-card {
  background: var(--bg-card); border: 1rpx solid var(--border-color);
  border-radius: 20rpx; padding: 28rpx 24rpx;
  margin-bottom: 20rpx; position: relative;
}

.tier-year {
  border: 3rpx solid #0d9488;
  box-shadow: 0 12rpx 32rpx rgba(13, 148, 136, 0.25);
}

.tier-lifetime {
  background: linear-gradient(135deg, #7c3aed 0%, #0d9488 100%);
  border-color: rgba(255, 255, 255, 0.3);
  color: #ffffff;
}

.badge-rec, .badge-life {
  position: absolute; top: -16rpx; left: 28rpx;
  font-size: 20rpx; font-weight: 700; color: #ffffff;
  padding: 4rpx 14rpx; border-radius: 8rpx;
}

.badge-rec  { background: #0d9488; }
.badge-life { background: #7c3aed; }

.tier-head {
  display: flex; align-items: baseline; justify-content: space-between;
  margin-bottom: 16rpx;
}

.tier-name {
  font-size: 30rpx; font-weight: 700; color: var(--text-primary);
}

.tier-lifetime .tier-name { color: #ffffff; }

.tier-price {
  font-size: 36rpx; font-weight: 800; color: var(--text-primary);
}

.tier-lifetime .tier-price { color: #ffffff; }

.tier-price .unit {
  font-size: 22rpx; color: var(--text-secondary); font-weight: 500;
  margin-left: 4rpx;
}

.tier-lifetime .tier-price .unit { color: rgba(255, 255, 255, 0.7); }

.features {
  display: flex; flex-direction: column; gap: 10rpx; margin-bottom: 18rpx;
}

.feat {
  display: flex; align-items: center; gap: 10rpx;
  font-size: 22rpx; color: var(--text-secondary);
}

.tier-lifetime .feat { color: rgba(255, 255, 255, 0.9); }

/* CTA */

.cta {
  text-align: center; padding: 18rpx; border-radius: 14rpx;
  font-size: 26rpx; font-weight: 600;
}

.cta-primary {
  background: #0d9488; color: #ffffff;
}

.cta-on-dark { background: #ffffff; color: #0d9488; }

.cta-off {
  background: var(--border-color); color: var(--text-secondary);
  pointer-events: none;
}

.cta-off-dark {
  background: rgba(255, 255, 255, 0.15); color: rgba(255, 255, 255, 0.6);
  pointer-events: none;
}

.cta-info {
  background: transparent; color: var(--text-secondary);
  border: 1rpx dashed var(--border-color); pointer-events: none;
}

.cta-done {
  background: rgba(255, 255, 255, 0.2); color: #ffffff;
  pointer-events: none; border: 1rpx solid rgba(255, 255, 255, 0.4);
}

/* Reveal + YEAR lift */

.reveal { opacity: 0; transform: translateY(14rpx); }

.section-root.is-in-view .reveal {
  animation: rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.section-root.is-in-view .reveal.delay-0 { animation-delay: 0ms; }
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
.section-root.is-in-view .reveal.delay-2 { animation-delay: 160ms; }
.section-root.is-in-view .reveal.delay-3 { animation-delay: 240ms; }

@keyframes rev {
  from { opacity: 0; transform: translateY(14rpx); }
  to   { opacity: 1; transform: translateY(0); }
}

.section-root.is-in-view .tier-year {
  animation:
    rev 0.45s cubic-bezier(0.22, 1, 0.36, 1) 160ms both,
    year-lift 0.64s ease-in-out 720ms 1 both;
}

@keyframes year-lift {
  0%, 100% { transform: translateY(0); }
  50%      { transform: translateY(-6rpx); }
}
```

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/components/home/membership/
git commit -m "feat(mini): add membership home section with member-aware CTAs"
```

---

## Task 10: dx-mini — wire sections into `pages/home/home`

**Files (modified):**
- `dx-mini/miniprogram/pages/home/home.ts`
- `dx-mini/miniprogram/pages/home/home.wxml`
- `dx-mini/miniprogram/pages/home/home.wxss`
- `dx-mini/miniprogram/pages/home/home.json`

- [ ] **Step 1: Update `home.json` — register the 6 new components**

Replace the contents of `dx-mini/miniprogram/pages/home/home.json` with:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-skeleton": "@vant/weapp/skeleton/index",
    "dx-icon": "/components/dx-icon/index",
    "home-why-different": "/components/home/why-different/index",
    "home-features": "/components/home/features/index",
    "home-ai-features": "/components/home/ai-features/index",
    "home-learning-loop": "/components/home/learning-loop/index",
    "home-community": "/components/home/community/index",
    "home-membership": "/components/home/membership/index"
  }
}
```

- [ ] **Step 2: Rewrite `home.ts` — parallel fetch + extra state**

Replace the contents of `dx-mini/miniprogram/pages/home/home.ts` with:

```typescript
import { api } from '../../utils/api'
import { gradeLabel } from '../../utils/format'

interface DashboardProfile {
  id: string
  username: string
  nickname: string | null
  grade: string
  level: number
  exp: number
  beans: number
  avatarUrl: string | null
  currentPlayStreak: number
  inviteCode: string
  lastReadNoticeAt: string | null
  vipDueAt: string | null
}

interface Greeting { title: string; subtitle: string }

interface SessionProgress {
  gameId: string
  gameName: string
  gameMode: string
  completedLevels: number
  totalLevels: number
  score: number
  exp: number
  lastPlayedAt: string
}

interface MasterStats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

interface DashboardData {
  profile: DashboardProfile
  masterStats: MasterStats
  reviewStats: ReviewStats
  sessions: SessionProgress[]
  todayAnswers: number
  greeting: Greeting
}

interface UnknownStats { total: number; thisWeek: number; thisMonth: number }

interface RecentSession {
  gameId: string
  gameName: string
  completedLevels: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    profile: null as DashboardProfile | null,
    greeting: null as Greeting | null,
    gradeLabelText: '',
    statusBarHeight: 20,
    // marketing sections
    masterTotal: null as number | null,
    reviewPending: null as number | null,
    unknownTotal: null as number | null,
    recentSession: null as RecentSession | null,
    vipDueAt: '' as string,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 0, theme: app.globalData.theme }) }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    try {
      const [dash, unknownStats] = await Promise.all([
        api.get<DashboardData>('/api/hall/dashboard'),
        api.get<UnknownStats>('/api/tracking/unknown/stats'),
      ])

      const sessions = dash.sessions || []
      const recentSession: RecentSession | null = sessions.length > 0
        ? {
            gameId: sessions[0].gameId,
            gameName: sessions[0].gameName,
            completedLevels: sessions[0].completedLevels,
          }
        : null

      this.setData({
        loading: false,
        profile: dash.profile,
        greeting: dash.greeting,
        gradeLabelText: gradeLabel(dash.profile.grade),
        masterTotal: dash.masterStats.total,
        reviewPending: dash.reviewStats.pending,
        unknownTotal: unknownStats.total,
        recentSession,
        vipDueAt: dash.profile.vipDueAt || '',
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goSearch() { wx.navigateTo({ url: '/pages/games/games' }) },
  goPurchase() { wx.navigateTo({ url: '/pages/me/purchase/purchase' }) },
  goInvite() { wx.navigateTo({ url: '/pages/me/invite/invite' }) },
  goStudy() { wx.navigateTo({ url: '/pages/me/study/study' }) },
  goGroups() { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  goTasks() { wx.navigateTo({ url: '/pages/me/tasks/tasks' }) },
  goCommunity() { wx.navigateTo({ url: '/pages/me/community/community' }) },
  goFeedback() { wx.navigateTo({ url: '/pages/me/feedback/feedback' }) },
})
```

- [ ] **Step 3: Append 6 section tags to `home.wxml`**

Edit `dx-mini/miniprogram/pages/home/home.wxml`. Locate the closing `</view>` of `.circle-row` (currently line 81). Immediately after the closing `</view>` of the `.circle-row` block and before the closing `</view>` of `.page-container`, insert:

```xml
    <!-- ===== Marketing sections (scroll-revealed below the hub) ===== -->
    <home-why-different />
    <home-features recent-session="{{recentSession}}" />
    <home-ai-features beans="{{profile ? profile.beans : 0}}" />
    <home-learning-loop
      unknown-total="{{unknownTotal}}"
      review-pending="{{reviewPending}}"
      master-total="{{masterTotal}}"
    />
    <home-community streak="{{profile ? profile.currentPlayStreak : 0}}" />
    <home-membership
      grade="{{profile ? profile.grade : 'free'}}"
      vip-due-at="{{vipDueAt}}"
    />
```

Use the `profile ? profile.X : Y` pattern (not `||`) to avoid reading properties of a null profile. Once profile loads, the `profile ?` branch selects the real value; until then, safe defaults render.

The final WXML ends as:

```xml
    </view> <!-- .circle-row -->

    <!-- ===== Marketing sections (scroll-revealed below the hub) ===== -->
    <home-why-different />
    <home-features recent-session="{{recentSession}}" />
    <home-ai-features beans="{{profile ? profile.beans : 0}}" />
    <home-learning-loop
      unknown-total="{{unknownTotal}}"
      review-pending="{{reviewPending}}"
      master-total="{{masterTotal}}"
    />
    <home-community streak="{{profile ? profile.currentPlayStreak : 0}}" />
    <home-membership
      grade="{{profile ? profile.grade : 'free'}}"
      vip-due-at="{{vipDueAt}}"
    />
  </view> <!-- .page-container -->
</van-config-provider>
```

- [ ] **Step 4: Ensure `home.wxss` has enough bottom breathing room**

Edit `dx-mini/miniprogram/pages/home/home.wxss`. Locate `.page-container` (currently line 1–5):

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
```

Increase bottom padding from `100rpx` to `140rpx` so the last section doesn't crowd the custom tab bar:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 140rpx;
}
```

No other WXSS changes needed — each section component carries its own layout.

- [ ] **Step 5: Run the icon build script to validate WXML icon references**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: `Wrote 36 icons to miniprogram/components/dx-icon/icons.ts.` with no errors. The static scan walks all WXML files and will surface any `<dx-icon name="...">` literal that isn't declared in the ICONS array.

- [ ] **Step 6: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/home/
git commit -m "feat(mini): wire marketing sections into home page"
```

---

## Task 11: Smoke tests + merge to main

**Files:** none modified (verification + merge).

Per the user's git-workflow rule: merge feature branch to main locally, then push only main. Feature branches are not pushed to remote.

- [ ] **Step 1: Open WeChat Developer Tools**

Open the project at `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini`. Confirm the build succeeds with no red errors in the console.

- [ ] **Step 2: Point at a working dx-api**

In the DevTools console tab, run:

```javascript
require('./utils/config').setDevApiBaseUrl('http://<your-lan-ip>:3001')
```

Or if dx-api is already running on `localhost:3001` and the "不校验合法域名" checkbox is on, leave it on the default.

Make sure dx-api is running (`cd dx-api && air` or `go run .`).

- [ ] **Step 3: Smoke test — light mode cold load**

1. Kill the mini program and re-open it.
2. Log in.
3. Land on the home page.
4. Scroll slowly top → bottom.

Expected:
- The hub (teal band / combined card / 5 circles) renders first, as before.
- Below the circles, 6 sections reveal as they enter the viewport, with subtle fade + translateY.
- AI typewriter lines type in sequentially when the AI section comes into view.
- 7-cell heatmap on the Community section fills left-to-right (only cells under the current streak count light up).
- Year membership card does a subtle one-shot lift after revealing.
- No continuous looping motion anywhere after the initial reveal.

- [ ] **Step 4: Smoke test — dark mode**

1. Tap the 我 tab → theme toggle → switch to dark.
2. Tap the 首页 tab → observe.

Expected:
- All sections render with dark backgrounds (`var(--bg-page)` / gradient variants for AI / LearningLoop).
- Membership section stays in its vivid violet→teal gradient (intentional).
- AI typewriter demo stays terminal-dark in both themes.
- Accent dashes, headlines, card texts are legible.

- [ ] **Step 5: Smoke test — membership CTA states**

1. As a `free` user (default if your test account has no subscription), observe all paid tier buttons say `立即开通`, FREE shows `默认权益`.
2. In DevTools Storage, open the network panel → mock a `/api/hall/dashboard` response with `profile.grade = 'year'` and `profile.vipDueAt = '<a date 180 days ahead>'`. Reload home.
3. Expect:
   - FREE → `默认权益`
   - MONTH → `已包含` (disabled)
   - YEAR → `续费 · 还剩 180 天` (primary)
   - LIFETIME → `升级到终身` (primary)
   - A "你的会员还剩 180 天" line above the tier cards.
4. Repeat for `lifetime`: MONTH/YEAR disabled (`已包含`), LIFETIME shows `✨ 已开通` with celebratory styling.

If mocking the response is hard, at minimum verify the `free` path looks right and trust the CTA matrix logic unit test (there is none, but the `computeStates` function is deterministic and reviewable).

- [ ] **Step 6: Smoke test — navigation**

1. Tap the resume chip (if `sessions` has entries) → arrives at `/pages/games/detail/detail`.
2. Tap each LearningLoop card → arrives at respective `/pages/learn/...`.
3. Tap Community cards → `/pages/leaderboard/leaderboard`, `/pages/me/community/community`, `/pages/me/groups/groups`.
4. Tap any non-disabled Membership CTA → `/pages/me/purchase/purchase`.
5. Tap disabled CTAs → no navigation.

- [ ] **Step 7: Smoke test — network failure**

1. In DevTools network panel, block `/api/hall/dashboard`.
2. Reload home.

Expected:
- Hub shows skeleton (existing behavior).
- Below the hub, marketing sections render in safe-empty state: no resume chip, `—` in the learning counts, streak `0`, `立即开通` CTAs across tiers.
- One toast: `加载失败`. No error spam.

- [ ] **Step 8: Real-device preview**

In DevTools click "预览" → scan the QR code with WeChat → open via 小程序助手 (per memory: 真机调试 is currently broken on macOS DevTools 2.02.2604152 due to a simulator-screenshot race, so preview is the right path).

Smoke the same light-mode cold load scenario on a real phone. Confirm:
- Scroll is smooth (≥55 fps via the in-DevTools perf panel while on simulator).
- No visible jank during typewriter or heatmap fills.

- [ ] **Step 9: Merge feature branch to main**

From repo root, with everything committed:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git checkout main
git merge --no-ff feat/mini-home-marketing-sections
```

Expected: a merge commit referencing the 10 feature commits. Resolve any conflicts (shouldn't be any since this is additive).

- [ ] **Step 10: Push main only**

```bash
git push origin main
```

Do not push the feature branch (per user's git-workflow memory). The feature branch can be deleted locally once main has absorbed it:

```bash
git branch -d feat/mini-home-marketing-sections
```
