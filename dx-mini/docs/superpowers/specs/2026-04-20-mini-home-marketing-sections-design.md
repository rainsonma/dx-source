# dx-mini: Home Marketing Sections — Design

**Date:** 2026-04-20
**Scope:** `dx-mini` (new components + home wiring) + `dx-api` (one-field addition). No changes to `dx-web` or `deploy/`.
**Status:** Approved design; implementation plan pending.

## 1. Goals

Extend the freshly-redesigned `/pages/home/home` hub with six marketing-style sections ported from `dx-web`'s anonymous landing page, re-tuned for a **logged-in mobile audience**. The sections anchor on real user data where that's honest (streak, bean count, vocabulary counts, current session, membership tier), and stay static-copy where it's brand story (WhyDifferent).

The six sections, in order of appearance below the existing circle hub:

1. **WhyDifferent** — `为什么是斗学`, three before→after rows.
2. **Features** — `核心玩法`, four game-mode cards with static illustrations.
3. **AiFeatures** — `AI 驱动 · 专属于你的学习`, bullets + typewriter demo.
4. **LearningLoop** — `学习闭环 · 从陌生到掌握`, three vocabulary-book cards with live counts.
5. **Community** — `一起玩才好玩`, four cards (排行榜 / 斗学社 / 学习群 / 连续打卡 with live streak).
6. **Membership** — `会员计划`, four tiers with member-aware CTAs.

## 2. Non-goals

- No change to the existing teal hub (header band / combined VIP-Rewards card / 5 circles). The new sections render **below** the circle strip.
- No redesign of `/pages/me/purchase/purchase`, `/pages/me/groups/groups`, `/pages/leaderboard/leaderboard`, or any stub page they link to.
- No FAQ or FinalCTA section from dx-web's landing — out of scope.
- No referral banner at the bottom of Membership (explicit design decision D5). 30% 返佣 stays only as a LIFETIME tier feature bullet.
- No dx-web-style horizontal "left → arrow → right" layout for WhyDifferent rows — we stack vertically (top → ↓ → bottom).
- No level-progression visualization card in LearningLoop (the Lv.01 ✓ / Lv.02 ✓ / Lv.03 ○ connector from dx-web). The three book cards carry the message on their own.
- No dark-mode reworking of the AiFeatures demo — the typewriter box stays dark in both themes by design.
- No unit tests. Project convention.
- No `/api/web/*` vs `/api/mini/*` split. Shared API surface is right for this team size.

## 3. Constraints

- **Custom nav style** stays as-is; the new sections scroll under the existing nav band.
- **`<dx-icon>`** is the only sanctioned icon primitive — all 11 new glyphs go through `scripts/build-icons.mjs`, regenerated via `npm run build:icons`.
- **No `?.` / `??`** in dx-mini TS or WXML — use explicit null checks or `||`.
- **TypeScript strict mode** must still pass, with only the tolerated `this`-in-`Component` pattern.
- **No continuous CSS loops or infinite animations.** Motion fires **once** when the section enters the viewport, via `wx.createIntersectionObserver`.
- **No `wx.createAnimation` or WXS animation tricks.** Everything is CSS `@keyframes` keyed off an `.is-in-view` class on the section root.
- **Glass-easel + Skyline** rendering stays; nothing here forces the legacy engine.
- **Vant Weapp 1.11.x** is available but **not used** inside the new components — the existing `<van-skeleton>` on the home hub is untouched.

## 4. Decisions locked during brainstorming

| # | Decision | Chosen |
|---|---|---|
| D1 | Placement & audience | Append below hub; member-aware CTAs on Membership. |
| D2 | Motion fidelity | Selective — reveal once on enter, no continuous loops. |
| D3 | Visual treatment | Hybrid — flat page bg for WhyDifferent / Features / Community, teal/violet gradient bands for AiFeatures / LearningLoop / Membership. Accent dash above every section. |
| D4 | Game-mode preview fidelity | Static illustration + one-shot animate-in per card. |
| D5 | Membership layout | Vertical stack, **no** referral CTA banner at the bottom. |
| D6 | Member-aware plumbing | Add `vipDueAt` to existing `DashboardProfile`. `profile.grade` (tier) is already returned. |
| D7 | FREE tier CTA | Drop button; render info-only with `默认权益` hint. |
| D8 | Section architecture | One custom component per section under `components/home/`. |

## 5. API change (dx-api)

One additive, nullable field on the existing `/api/hall/dashboard` response.

**File:** `dx-api/app/services/api/hall_service.go`

```diff
 type DashboardProfile struct {
     ID                string  `json:"id"`
     Username          string  `json:"username"`
     Nickname          *string `json:"nickname"`
     Grade             string  `json:"grade"`
     Level             int     `json:"level"`
     Exp               int     `json:"exp"`
     Beans             int     `json:"beans"`
     AvatarURL         *string `json:"avatarUrl"`
     CurrentPlayStreak int     `json:"currentPlayStreak"`
     InviteCode        string  `json:"inviteCode"`
     LastReadNoticeAt  any     `json:"lastReadNoticeAt"`
     CreatedAt         any     `json:"createdAt"`
+    VipDueAt          *carbon.DateTime `json:"vipDueAt"`
 }
```

And in `GetDashboard`:

```diff
     profile := DashboardProfile{
         ...existing fields...,
+        VipDueAt: user.VipDueAt,
     }
```

Non-breaking for dx-web (new optional JSON key; their TS type extends). `dx-web` can later drop its redundant `/api/user/profile` fetch-for-vipDueAt on the hall home, but that refactor is out of scope for this spec.

## 6. Component architecture

```
dx-mini/miniprogram/
├── components/home/
│   ├── why-different/
│   │   ├── index.ts         (observer + setData({inView}))
│   │   ├── index.wxml
│   │   ├── index.wxss
│   │   └── index.json
│   ├── features/            (same 4-file shape)
│   ├── ai-features/
│   ├── learning-loop/
│   ├── community/
│   └── membership/
├── utils/
│   └── in-view.ts           (NEW helper, ~20 LOC)
├── pages/home/
│   ├── home.ts              (edit: orchestrate data + pass props)
│   ├── home.wxml            (edit: append 6 section tags after circle-row)
│   ├── home.wxss            (edit: section-rhythm resets)
│   └── home.json            (edit: register 6 new component paths)
├── scripts/build-icons.mjs  (edit: append 11 glyphs)
└── components/dx-icon/icons.ts   (regenerated)
```

### 6.1 Component public API

| Component | Props | Notes |
|---|---|---|
| `<home-why-different>` | — | Static copy only. |
| `<home-features>` | `recent-session` | Object `{gameId, gameName, completedLevels}` or null. |
| `<home-ai-features>` | `beans` | Number, live bean count. |
| `<home-learning-loop>` | `unknown-total`, `review-pending`, `master-total` | Three numbers, all dashboard-derived (plus the new `/api/tracking/unknown/stats` call). |
| `<home-community>` | `streak` | Number, live play streak. |
| `<home-membership>` | `grade`, `vip-due-at` | String tier + ISO datetime or null. |

All components are self-contained otherwise: they own their copy, their icons, their styles, and their intersection observer.

### 6.2 `utils/in-view.ts`

```ts
// Trigger a callback once when the element matching `selector`
// intersects the viewport. Disconnect after firing.
export function observeOnce(
  component: WechatMiniprogram.Component.Instance<any, any, any>,
  selector: string,
  cb: () => void,
  threshold = 0.15,
) {
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

Fallback: each component also sets a 1000 ms `setTimeout` in `attached()` that flips `inView: true` unconditionally. This protects against the rare case where `createIntersectionObserver` behaves unexpectedly (older real-device clients, 真机调试 quirks). First-to-fire wins.

### 6.3 Motion CSS pattern

Every section root renders as:

```html
<view class="section-root {{inView ? 'is-in-view' : ''}}">
  ...
</view>
```

And its WXSS defines keyframes that only fire when `.is-in-view` is present:

```css
.section-root .reveal { opacity: 0; transform: translateY(14px); }
.section-root.is-in-view .reveal {
  animation: reveal 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}
.section-root.is-in-view .reveal.delay-1 { animation-delay: 80ms; }
@keyframes reveal {
  from { opacity: 0; transform: translateY(14px); }
  to   { opacity: 1; transform: translateY(0); }
}
```

Section-specific animations (typewriter, tile-rise, bar-rise, heatmap-fill, lift-pulse) follow the same gating: defined at root, only fire under `.is-in-view`, use `animation-fill-mode: both` so the end-state sticks.

## 7. Data orchestration (`pages/home/home.ts`)

Current `onShow` fetches `/api/hall/dashboard` only. Extend to fetch the one new stat in parallel.

### 7.1 Initial data (safe-empty defaults)

```ts
data: {
  // ...existing (profile, greeting, gradeLabelText)
  masterTotal: null as number | null,     // null → render '—'
  reviewPending: null as number | null,
  unknownTotal: null as number | null,
  recentSession: null as RecentSession | null,
  vipDueAt: null as string | null,
}
```

Every section's WXML uses `{{x !== null ? x : '—'}}` (or `wx:if="{{x !== null}}"`) so nothing reads uninitialized numbers as `0` before fetch lands.

### 7.2 `onShow` flow

```ts
async onShow() {
  try {
    const [dash, unknownStats] = await Promise.all([
      api.get<DashboardData>('/api/hall/dashboard'),
      api.get<{ total: number }>('/api/tracking/unknown/stats'),
    ])

    const profile = dash.profile
    const recentSession = dash.sessions && dash.sessions.length > 0
      ? {
          gameId: dash.sessions[0].gameId,
          gameName: dash.sessions[0].gameName,
          completedLevels: dash.sessions[0].completedLevels,
        }
      : null

    this.setData({
      // existing
      profile,
      greeting: dash.greeting,
      gradeLabelText: gradeLabel(profile.grade),
      // new, for marketing sections
      masterTotal: dash.masterStats.total,
      reviewPending: dash.reviewStats.pending,
      unknownTotal: unknownStats.total,
      recentSession,
      // new, for membership section
      vipDueAt: profile.vipDueAt,
    })
  } catch (err) {
    console.warn('[home] dashboard/unknown-stats fetch failed', err)
    // hub still renders; marketing sections check their own props and
    // render in safe-empty mode (see edge cases in §11).
  }
}
```

**Error semantics.** If the dashboard call fails, the existing hub falls back to its loading skeleton + default "你好" greeting (already the current behavior). Marketing sections either render with 0/empty values or (for Features resume chip and Membership CTA) render an abstract fallback so nothing looks broken.

If only the `/api/tracking/unknown/stats` call fails, LearningLoop renders `—` in the 生词本 slot and the other two cards still show real counts.

## 8. `pages/home/home.wxml` edits

Below the existing `<view class="circle-row">…</view>`, append:

```xml
<home-why-different />
<home-features recent-session="{{recentSession}}" />
<home-ai-features beans="{{profile.beans || 0}}" />
<home-learning-loop
  unknown-total="{{unknownTotal}}"
  review-pending="{{reviewPending}}"
  master-total="{{masterTotal}}"
/>
<home-community streak="{{profile.currentPlayStreak || 0}}" />
<home-membership
  grade="{{profile.grade || 'free'}}"
  vip-due-at="{{vipDueAt}}"
/>
```

Explicit `|| 0` / `|| 'free'` fallbacks because of the no-`??` constraint.

### 8.1 `home.json` addition

```json
"usingComponents": {
  ...existing...,
  "home-why-different": "../../components/home/why-different/index",
  "home-features": "../../components/home/features/index",
  "home-ai-features": "../../components/home/ai-features/index",
  "home-learning-loop": "../../components/home/learning-loop/index",
  "home-community": "../../components/home/community/index",
  "home-membership": "../../components/home/membership/index"
}
```

### 8.2 `home.wxss` edits

- No changes to existing hub styles.
- Add `.page-container { padding-bottom: 120rpx; }` if the last section crowds the tab bar (to be verified during smoke tests; hub currently has no bottom padding).

## 9. Section specifications

### 9.1 `<home-why-different>`

**Layout (top → bottom).**

```
<section-root class="why-different">
  <accent-dash color="teal"/>                                   ← 24×3px teal bar
  <tag>为什么是斗学</tag>
  <headline>不再死记硬背,<hl>让大脑爱上英语</hl></headline>
  <row-card .reveal .delay-0>
    <before>背了就忘,靠意志力硬撑</before>
    <arrow>↓</arrow>
    <after>游戏化循环,大脑自发想再玩一局</after>
  </row-card>
  <row-card .reveal .delay-1> ... row 2 ... </row-card>
  <row-card .reveal .delay-2> ... row 3 ... </row-card>
</section-root>
```

**Copy (verbatim).**
- Row 1 — `背了就忘,靠意志力硬撑` → `游戏化循环,大脑自发想再玩一局`
- Row 2 — `学的和用的两张皮` → `连词成句、对话、对战都是真实语料`
- Row 3 — `一个人孤独地学` → `好友开黑 · 学习群 · 排行榜 · 每日连胜`

**Visuals.** Card background white (`var(--bg-card)`), border `var(--border-color)`, rounded 24rpx. The `<after>` block sits on a subtle teal→violet linear-gradient strip inside the card to emphasize the "after" feel. `<before>` text color `var(--text-secondary)` with strikethrough. `<after>` text color `var(--text-primary)` bold. Arrow is a `<dx-icon name="arrow-right">` rotated 90deg, color `#0d9488`.

**Motion.** Three cards fade+lift with `.delay-0/1/2` at 80 ms increments. Arrow scales from 0.4 → 1 with a short bounce (via cubic-bezier), starting 120 ms after each card resolves.

**Navigation.** None — the cards are non-tappable.

### 9.2 `<home-features>`

**Layout.**

```
<section-root class="features">
  <accent-dash color="violet"/>
  <tag>核心玩法 · 4 种模式,覆盖听说读写</tag>
  <headline>挑一个就能上手,<hl>全都玩转就起飞</hl></headline>
  <resume-chip wx:if="{{recentSession}}" bindtap="goResume">
    继续 · {{recentSession.gameName}} Lv.{{recentSession.completedLevels + 1}}
    <dx-icon name="chevron-right"/>
  </resume-chip>
  <game-card mode="sentence" .reveal .delay-0> ... </game-card>
  <game-card mode="match"    .reveal .delay-1> ... </game-card>
  <game-card mode="eliminate" .reveal .delay-2> ... </game-card>
  <game-card mode="battle"   .reveal .delay-3> ... </game-card>
</section-root>
```

**Copy (verbatim per dx-web).**
- `连词成句` · `看到中文秒拼出英文句子,越快越高分。真实语料替你练习语序和搭配。` · icon `keyboard` on violet-100 square.
- `词汇配对` · `英文与中文快速配对,限时给分。巩固词汇量和中译英反应速度。` · icon `swords` on blue-100 square.
- `词汇消消乐` · `记忆配对消除,越快消除越高分。玩着玩着就把生词牢牢记住。` · icon `shuffle` on pink-100 square.
- `词汇对轰` · `和对手拼炮弹。拼对拼快就发射,紧张刺激的词汇对战。` · icon `crosshair` on red-100 square.

**Illustrations (static + animate-in on section enter).**
- `sentence`: 5 word tiles `[I] [love] [learning] [English] [→]`, each slides up with 100 ms stagger.
- `match`: left tile `negotiate` (teal-filled) → connector line draws left-to-right → right tile `谈判`.
- `eliminate`: 3×3 grid of colored cells pop-scale from 0.4 → 1, 50 ms stagger per cell.
- `battle`: 5 vertical bars scale-Y from 0 → full, 70 ms stagger. Bars alternate red and teal.

**Motion timing.** Card fades in first (0 ms), then the illustration inside begins at +180 ms. Per-card elements stagger as above. Total illustration run-time ~450–550 ms, then static.

**Resume chip behavior.**
- Renders only if `recentSession` is truthy.
- Tap → `wx.navigateTo({ url: '/pages/games/detail/detail?id=' + recentSession.gameId })`.
- Chip is visually subtle: `var(--bg-card)` with teal border, 20rpx vertical padding.

### 9.3 `<home-ai-features>`

**Layout.** First gradient section (white → teal-50 → white in light, dark-teal band in dark).

```
<section-root class="ai-features">
  <accent-dash color="teal"/>
  <tag><dx-icon name="sparkles"/> AI 驱动 · 专属于你的学习</tag>
  <headline>AI 帮你定制课程,<hl>你想学什么都可以</hl></headline>
  <bullet .reveal .delay-0>输入任意主题或场景,AI 按你的水平生成课程</bullet>
  <bullet .reveal .delay-1>CEFR A1–C2 全覆盖,难度智能匹配</bullet>
  <bullet .reveal .delay-2>内容沉淀进你的词汇系统,复习自动推送</bullet>
  <demo-card .reveal .delay-3>
    <input-label>输入主题:<hl>职场面试高频词</hl></input-label>
    <stream>
      <line>› negotiate · 谈判 — Let's negotiate the salary.</line>
      <line>› résumé · 简历 — Please send me your résumé.</line>
      <line>› confident · 自信 — Stay confident during the interview.</line>
      <line>› leverage · 优势 — Leverage your strengths.</line>
      <line>› follow up · 跟进 — I'll follow up next week.</line>
    </stream>
    <footer>
      <dx-icon name="coins" color="#d97706"/>
      你当前有 <hl>{{beans}}</hl> 能量豆 · 每次消耗 5 颗,失败全额退还
    </footer>
  </demo-card>
</section-root>
```

**Typewriter motion.** Each `<line>` starts at `width: 0; overflow: hidden; white-space: nowrap`. Only the **currently-typing** line shows the caret (`border-right: 2px solid #0d9488`); finished lines drop the caret via an `animation-fill-mode: forwards` rule that clears `border-right-color` on the completing line. Stagger:

```css
.section-root.is-in-view .stream .line:nth-child(1) {
  animation: typer 1.1s steps(44) both; animation-delay: 240ms;
}
.section-root.is-in-view .stream .line:nth-child(2) {
  animation: typer 1.1s steps(42) both; animation-delay: 1500ms;
}
... (3) 2700ms ... (4) 3900ms ... (5) 5000ms
@keyframes typer {
  from { width: 0; border-right-color: #0d9488; }
  99%  { border-right-color: #0d9488; }
  to   { width: 100%; border-right-color: transparent; }
}
```

Total ~6.5 s run-time, then the demo sits static. No loop.

**Demo card styling.** In light mode: `var(--bg-card)` white with teal border and `box-shadow: 0 4rpx 16rpx rgba(13,148,136,0.08)`. In dark mode: `#0f172a` terminal-like background, `#67e8f9` prefix color, `#e2e8f0` body text — **same treatment in both themes** by design (the demo evokes a code window).

### 9.4 `<home-learning-loop>`

**Layout.** Second gradient section.

```
<section-root class="learning-loop">
  <accent-dash color="teal"/>
  <tag>学习闭环 · 从陌生到掌握</tag>
  <headline>一套系统,追踪你每一个学习单元的命运</headline>
  <book-card .reveal .delay-0>
    <dx-icon name="book-open" color="#ec4899"/>
    <title>生词本</title>
    <count>{{unknownTotal !== null ? unknownTotal : '—'}} 条</count>
    <desc>持续沉淀你不会的词汇</desc>
  </book-card>
  <book-card .reveal .delay-1>
    <dx-icon name="refresh-cw" color="#7c3aed"/>
    <title>复习本</title>
    <count>{{reviewPending}} 待复习</count>
    <desc>[1, 3, 7, 14, 30, 90] 天节奏智能推送</desc>
  </book-card>
  <book-card .reveal .delay-2>
    <dx-icon name="circle-check" color="#0d9488"/>
    <title>已掌握</title>
    <count>{{masterTotal}} 条</count>
    <desc>看得见的词汇量增长</desc>
  </book-card>
  <footnote>艾宾浩斯遗忘曲线智能推送复习,你只需要玩。</footnote>
</section-root>
```

**Navigation.**
- 生词本 → `/pages/learn/unknown/unknown`
- 复习本 → `/pages/learn/review/review`
- 已掌握 → `/pages/learn/mastered/mastered`

**Motion.** Cards fade+lift staggered 80 ms. Numbers do **not** tween-up — they render at final value. Icons static.

### 9.5 `<home-community>`

**Layout.** Flat bg.

```
<section-root class="community">
  <accent-dash color="amber"/>
  <tag>一起玩才好玩</tag>
  <headline>和朋友开黑,<hl>排行榜上见</hl></headline>
  <feature-card .reveal .delay-0 bindtap="goLeaderboard">
    <dx-icon name="trophy" color="#f59e0b"/>
    <title>排行榜</title>
    <desc>经验值与在线时长,按日、周、月六种榜单,登上领奖台。</desc>
  </feature-card>
  <feature-card .reveal .delay-1 bindtap="goCommunity">
    <dx-icon name="message-square" color="#ea580c"/>
    <title>斗学社</title>
    <desc>发帖、评论、点赞、关注,把学习心得变成社交动态。</desc>
  </feature-card>
  <feature-card .reveal .delay-2 bindtap="goGroups">
    <dx-icon name="users" color="#3b82f6"/>
    <title>学习群</title>
    <desc>组建小组一起闯关,群内可直接发起课程对战。</desc>
  </feature-card>
  <streak-card .reveal .delay-3>
    <dx-icon name="flame" color="#0d9488"/>
    <title>连续打卡</title>
    <streak-number>你已连胜 <hl>{{streak}}</hl> 天</streak-number>
    <heatmap>
      <cell wx:for="{{7}}" ... class="{{index < heatCells ? 'filled' : ''}}"/>
    </heatmap>
    <desc>每天玩至少一次,保持连胜;错过一天从 1 重来。</desc>
  </streak-card>
</section-root>
```

**Computed data.** `heatCells = Math.min(streak, 7)`, computed in `index.ts`'s `properties`/`observers`.

**Navigation targets.**
- 排行榜 → `/pages/leaderboard/leaderboard`
- 斗学社 → `/pages/me/community/community` (stub; kept intentionally)
- 学习群 → `/pages/me/groups/groups`
- 连续打卡 → non-tappable.

**Motion.** Cards stagger in. Heatmap fill: **only cells with `.filled` class** animate `background-color: var(--border-color) → #0d9488`, left-to-right, 80 ms per cell (via nth-child `animation-delay`), starting 400 ms after the streak card resolves. Unfilled cells stay at `var(--border-color)`. Final frame sticks via `animation-fill-mode: forwards`.

### 9.6 `<home-membership>`

**Layout.** Third gradient section (violet-600 → teal-600, vivid in both themes).

```
<section-root class="membership">
  <accent-dash color="white"/>
  <tag>会员计划</tag>
  <headline>选一个你最舒服的节奏,<hl>越早开始越划算</hl></headline>
  <desc>还有季度会员等更多选项,在会员页查看完整对比。</desc>
  <tier-card tier="free"     .reveal .delay-0>...</tier-card>
  <tier-card tier="month"    .reveal .delay-1>...</tier-card>
  <tier-card tier="year"     .reveal .delay-2 class="recommended">...</tier-card>
  <tier-card tier="lifetime" .reveal .delay-3 class="lifetime">...</tier-card>
</section-root>
```

**Prices & features (verbatim from `app/consts/user_grade.go`).**

| Tier | Title | Price | Features |
|---|---|---|---|
| `free` | 免费会员 | ¥0 | `部分关卡`, `基础游戏` |
| `month` | 月度会员 | ¥39 / 月 | `全部关卡畅玩`, `AI 随心学`, `PK + 群组` |
| `year` | 年度会员 (推荐) | ¥309 / 年 | `超值优惠套餐`, `包含月度全部权益`, `优先客服支持` |
| `lifetime` | 终身会员 (最超值) | ¥1999 一次性 | `一次付费,终身生效`, `最高能量豆赠送`, `邀请好友首充享 30% 返佣` |

Each feature bullet prefixed with `<dx-icon name="circle-check">` (teal on white cards, white on the gradient LIFETIME card).

**Member-aware CTA matrix.** Driven by `grade` + `vipDueAt`. Computed in `index.ts` as derived `buttonStates`.

| `grade` \ Tier | FREE | MONTH | YEAR | LIFETIME |
|---|---|---|---|---|
| `free` | `默认权益` info | `立即开通` primary | `立即开通` primary | `立即开通` primary |
| `month` | `默认权益` info | `续费 · 还剩 N 天` primary | `立即开通` primary | `立即开通` primary |
| `season` | `默认权益` info | `已包含` disabled | `升级到年度` primary | `升级到终身` primary |
| `year` | `默认权益` info | `已包含` disabled | `续费 · 还剩 N 天` primary | `升级到终身` primary |
| `lifetime` | `默认权益` info | `已包含` disabled | `已包含` disabled | `✨ 已开通` disabled, celebratory |

Only the `续费` button — which targets the user's **current** tier — shows "还剩 N 天". Upgrade buttons are bare, because "还剩 N 天" refers to the source tier's expiry and putting it inside an upgrade label conflates two tiers. For `season` and `year` users, the days-remaining information is surfaced via a subtle line above the tier cards: `你的会员还剩 N 天`.

N = `daysUntil(vipDueAt)`. If `vipDueAt` is null and `grade` is a time-bounded tier, the 续费 button falls back to `立即开通` (defensive; shouldn't happen with healthy data).

**Helper.** Add to `utils/format.ts`:

```ts
export function daysUntil(isoDate: string | null | undefined): number {
  if (!isoDate) return 0
  const target = new Date(isoDate).getTime()
  const now = Date.now()
  return Math.max(0, Math.ceil((target - now) / 86400000))
}
```

**Active CTA behavior.** All tappable tier buttons route to `/pages/me/purchase/purchase`. Disabled ones have `pointer-events: none` and muted opacity.

**Motion.**
- Tiers stagger in at 80 ms each (0 ms / 80 ms / 160 ms / 240 ms).
- YEAR (recommended) card does a **one-shot lift** *after* its reveal resolves:

  ```css
  .section-root.is-in-view .tier-year {
    animation:
      reveal 0.45s cubic-bezier(0.22,1,0.36,1) 160ms both,
      year-lift 0.64s ease-in-out 720ms 1 both;
  }
  @keyframes year-lift {
    0%, 100% { transform: translateY(0); }
    50%      { transform: translateY(-6rpx); }
  }
  ```

  The lift animation starts 720 ms after the section enters view — ~100 ms after the YEAR card's own reveal finishes — so the pulse doesn't overlap the fade-in.

**Visuals.**
- Section background: `linear-gradient(135deg, #7c3aed, #0d9488)` in both themes (intentional brand crescendo).
- Tag, headline, desc: white text with alpha variants for hierarchy.
- FREE / MONTH / YEAR cards: `var(--bg-card)` (white in light, `#1c1c1e` in dark) with standard border. YEAR card gets `border: 2rpx solid #0d9488` + teal shadow glow.
- LIFETIME card: inner `linear-gradient(135deg, #7c3aed, #0d9488)`, white text, white ring.
- Badges: `推荐` (teal solid), `最超值` (violet solid), both positioned `top: -16rpx; left: 24rpx`.

## 10. Dark-mode matrix

| Section | Light bg | Dark bg |
|---|---|---|
| WhyDifferent | `var(--bg-page)` flat | same |
| Features | flat | flat |
| AiFeatures | `linear-gradient(180deg, #fff, #f0fdfa, #fff)` | `linear-gradient(180deg, var(--bg-page), #0f2524, var(--bg-page))` |
| LearningLoop | `linear-gradient(180deg, #f0fdfa, #ede9fe)` | `linear-gradient(180deg, #0f2524, #1a1230)` |
| Community | flat | flat |
| Membership | `linear-gradient(135deg, #7c3aed, #0d9488)` | same (stays vivid — intentional) |

Illustration palettes use `opacity: 0.7` in dark mode via `.dark .illus *` for a softer feel. All text / icons / cards read from CSS vars already set in `app.wxss`.

## 11. Edge cases

- **Dashboard API fails.** Hub still renders (existing skeleton behavior). All six marketing sections render with default props (`streak=0`, `beans=0`, `masterTotal=0`, `recentSession=null`, `grade='free'`, `vipDueAt=null`). Experience: "neutral new user" copy with `立即开通` everywhere.
- **`/api/tracking/unknown/stats` fails only.** Dashboard's other data displays. LearningLoop renders `—` in 生词本 slot. Other two cards unchanged.
- **`recentSession` null.** Features resume chip hidden, card grid renders full-width. No empty space.
- **`vipDueAt` null on a time-bounded grade.** Button falls back to `立即开通` for that tier (defensive — shouldn't happen with healthy data).
- **`profile.grade` unrecognized.** Treated as `free` in the CTA matrix. No UI breakage.
- **Older WeChat client without `createIntersectionObserver`.** Every component has a 1000 ms `setTimeout` in `attached()` that sets `inView: true`. Sections are fully visible; they just skip reveal animation.
- **User toggles light/dark via `/pages/me`.** On return to home, all sections pick up new CSS-var values. No explicit re-render needed.
- **Very narrow screen (< 320 px).** Illustrations use `flex-wrap`, tier card features stack tighter but don't clip. Accepted.
- **Very wide screen (tablet via WeChat).** Section max-width capped at 720rpx, centered. Existing hub already does this; sections inherit.
- **Resume chip navigation when game is offline / withdrawn.** dx-api returns `404` from `/api/games/{id}` → detail page already handles via existing error flow.

## 12. Smoke tests (manual, WeChat DevTools + 预览)

1. **Light mode cold load.** Scroll home top→bottom. Each section reveals once as it enters viewport. No continuous motion after reveal completes.
2. **Dark mode.** Toggle via `/pages/me` → return to home → verify all six sections render dark. Typewriter demo stays dark-themed in both modes.
3. **Logged-in `free` user.** All paid-tier buttons show `立即开通`. FREE tier shows `默认权益` info chip.
4. **Simulate each non-free grade** by editing `profile.grade` in devtools storage or proxying response → verify CTA matrix cells.
5. **Kill network after initial load.** Hub still renders (cached), marketing sections show safe-empty defaults. No error toast spam.
6. **Tap every nav target.**
   - Features resume chip → game detail.
   - LearningLoop cards → unknown / review / mastered lists.
   - Community cards → leaderboard / community stub / groups list.
   - Membership tier buttons → purchase page.
7. **Scroll performance** on a low-end Android via DevTools simulator. FPS stays ≥55 during reveal transitions.
8. **Real device preview (安卓 + iOS)** via WeChat DevTools 预览 → 微信"斗学"小程序助手. Visually compare reveal timing and illustration quality.
9. **Type-check + build.** `tsc --noEmit` inside `dx-mini/miniprogram/` passes with no new errors. WeChat DevTools "Upload" dry-run succeeds.

## 13. Rollout plan

- **Feature branch** `feat/mini-home-marketing-sections` (merged to `main` locally, only `main` pushed per the user's git-workflow rule).
- **Commit sequence** (each small, each green on its own):
  1. `feat(api): expose vipDueAt on hall dashboard`
  2. `chore(mini): add 11 marketing-home Lucide icons`
  3. `feat(mini): add in-view helper and section component shells`
  4. `feat(mini): add why-different section`
  5. `feat(mini): add features section with static illustrations`
  6. `feat(mini): add ai-features section with live bean count`
  7. `feat(mini): add learning-loop section with live counts`
  8. `feat(mini): add community section with live streak`
  9. `feat(mini): add membership section with member-aware CTAs`
  10. `feat(mini): wire sections into home and fetch unknown stats`
- After step 10, run manual smoke tests (§12). Merge to `main`.

## 14. File inventory

**dx-api — 1 file edited**
- `dx-api/app/services/api/hall_service.go` — add `VipDueAt` field + population.

**dx-mini — 1 helper + 24 component files + 1 generated icon file**
- `dx-mini/miniprogram/utils/in-view.ts` (new)
- `dx-mini/miniprogram/components/home/why-different/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/home/features/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/home/ai-features/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/home/learning-loop/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/home/community/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/home/membership/{index.ts,wxml,wxss,json}`
- `dx-mini/miniprogram/components/dx-icon/icons.ts` (regenerated by build script)

**dx-mini — 6 files edited**
- `dx-mini/miniprogram/pages/home/home.ts` — add parallel stat fetch + derived-prop setData.
- `dx-mini/miniprogram/pages/home/home.wxml` — append 6 section tags.
- `dx-mini/miniprogram/pages/home/home.wxss` — minor padding tweak.
- `dx-mini/miniprogram/pages/home/home.json` — register 6 components.
- `dx-mini/miniprogram/utils/format.ts` — add `daysUntil(isoDate)` helper.
- `dx-mini/miniprogram/scripts/build-icons.mjs` — append 11 icon names.

**Total:** 1 go file edited, 1 utils file new, 24 new mini component files, 6 edited mini files, 1 regenerated icon file. **33 file touches.**

## 15. Icons to add to `scripts/build-icons.mjs`

```
keyboard, swords, shuffle, crosshair, sparkles,
coins, refresh-cw, circle-check, message-square,
flame, arrow-right
```

Run `npm run build:icons` once; commits the regenerated `components/dx-icon/icons.ts`.
