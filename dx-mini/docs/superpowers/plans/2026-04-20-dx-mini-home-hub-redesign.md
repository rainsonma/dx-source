# dx-mini: Home Top-Section Hub Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dx-mini home page's top section with a teal-themed "hub" layout (greeting band + combined VIP/奖励 card + five gradient-teal icon circles), drop the old stat row and heatmap, add a small `level` field to the backend dashboard response, and scaffold four new "敬请期待" stub pages for destinations that don't exist yet.

**Architecture:** Four concentric layers of change, applied bottom-up. (1) dx-api exposes `level` on `DashboardProfile`; (2) `scripts/build-icons.mjs` grows by four Lucide icons and regenerates `icons.ts`; (3) four stub-page trees land under `pages/me/` and register in `app.json`; (4) the home page's wxml/ts/wxss get a coordinated rewrite. The order matters: icons must be in the inventory before WXML references them (the build script's static scan enforces this).

**Tech Stack:** Go 1.21+ / Goravel (dx-api); WeChat Mini Program native (glass-easel + Skyline), TypeScript strict, Vant Weapp 1.11.x, `<dx-icon>` (Lucide SVG renderer).

**Spec:** [2026-04-20-dx-mini-home-hub-redesign-design.md](../specs/2026-04-20-dx-mini-home-hub-redesign-design.md)

**Branch:** `feat/mini-home-hub-redesign` (create on Task 1 start from current `main`).

---

## Task 1: Backend — add `level` to `DashboardProfile`

**Files:**
- Modify: `dx-api/app/services/api/hall_service.go`

- [ ] **Step 1: Create the feature branch**

From repo root:

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source checkout -b feat/mini-home-hub-redesign
```

Expected: `Switched to a new branch 'feat/mini-home-hub-redesign'`.

- [ ] **Step 2: Add `Level` field to `DashboardProfile` struct**

Edit `dx-api/app/services/api/hall_service.go`. Locate the `DashboardProfile` struct (currently lines 24–36). Insert a new `Level` field between `Grade` and `Exp`, mirroring the pattern already used in `user_service.go`:

```go
// DashboardProfile is the user profile subset shown on the dashboard.
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
}
```

- [ ] **Step 3: Populate `Level` in `GetDashboard`**

In the same file, inside `GetDashboard`, after the `user.ID == ""` check and before the `profile := DashboardProfile{...}` literal, add:

```go
	level, err := consts.GetLevel(user.Exp)
	if err != nil {
		return nil, fmt.Errorf("failed to compute user level: %w", err)
	}
```

Then in the struct literal, add the `Level: level,` line between `Grade` and `Exp`:

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
	}
```

- [ ] **Step 4: Verify the package compiles**

Run:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: no output (success). If it fails with `fmt` or `consts` import errors, confirm those imports already exist at the top of `hall_service.go` (they do — `fmt`, `dx-api/app/consts`).

- [ ] **Step 5: Run the existing test suite with race detector**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```

Expected: all packages PASS. `user_level_test.go` still passes (unchanged), and no other test touches `DashboardProfile`'s field list directly.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/hall_service.go
git commit -m "feat(api): expose user level on hall dashboard"
```

---

## Task 2: Add four Lucide icons to the inventory

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Regenerate: `dx-mini/miniprogram/components/dx-icon/icons.ts`

- [ ] **Step 1: Append four rows to the ICONS array**

Edit `dx-mini/scripts/build-icons.mjs`. Locate the `ICONS` array (around lines 16–38). Append these four rows just before the closing `]`:

```js
  ['chart-pie',      'chart-pie'],
  ['calendar-check', 'calendar-check'],
  ['sticker',        'sticker'],
  ['flag',           'flag'],
```

Context — the final array should read:

```js
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
  ['notebook-text', 'notebook-text'],
  ['user',          'user'],
  ['book-text',     'book-text'],
  ['trophy',        'trophy'],
  ['chart-pie',      'chart-pie'],
  ['calendar-check', 'calendar-check'],
  ['sticker',        'sticker'],
  ['flag',           'flag'],
]
```

- [ ] **Step 2: Regenerate icons.ts**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected output: `Wrote 25 icons to miniprogram/components/dx-icon/icons.ts.`

If the script throws `lucide-static is missing "<name>.svg"`, `node_modules/lucide-static/icons/` doesn't have the SVG — run `npm install` in `dx-mini/` to refresh.

- [ ] **Step 3: Verify icons.ts contains the four new entries**

```bash
grep -E '"(chart-pie|calendar-check|sticker|flag)":' /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/components/dx-icon/icons.ts | wc -l
```

Expected: `4`.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/scripts/build-icons.mjs dx-mini/miniprogram/components/dx-icon/icons.ts
git commit -m "chore(mini): add chart-pie/calendar-check/sticker/flag icons"
```

---

## Task 3: Create the four stub pages

**Files (created):**
- `dx-mini/miniprogram/pages/me/study/study.json`
- `dx-mini/miniprogram/pages/me/study/study.ts`
- `dx-mini/miniprogram/pages/me/study/study.wxml`
- `dx-mini/miniprogram/pages/me/study/study.wxss`
- `dx-mini/miniprogram/pages/me/tasks/tasks.json`
- `dx-mini/miniprogram/pages/me/tasks/tasks.ts`
- `dx-mini/miniprogram/pages/me/tasks/tasks.wxml`
- `dx-mini/miniprogram/pages/me/tasks/tasks.wxss`
- `dx-mini/miniprogram/pages/me/community/community.json`
- `dx-mini/miniprogram/pages/me/community/community.ts`
- `dx-mini/miniprogram/pages/me/community/community.wxml`
- `dx-mini/miniprogram/pages/me/community/community.wxss`
- `dx-mini/miniprogram/pages/me/feedback/feedback.json`
- `dx-mini/miniprogram/pages/me/feedback/feedback.ts`
- `dx-mini/miniprogram/pages/me/feedback/feedback.wxml`
- `dx-mini/miniprogram/pages/me/feedback/feedback.wxss`

**Files (modified):**
- `dx-mini/miniprogram/app.json`

- [ ] **Step 1: Write `study.json`**

Create `dx-mini/miniprogram/pages/me/study/study.json` with:

```json
{
  "navigationBarTitleText": "学习",
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Write `study.ts`**

Create `dx-mini/miniprogram/pages/me/study/study.ts` with:

```typescript
Page({})
```

- [ ] **Step 3: Write `study.wxml`**

Create `dx-mini/miniprogram/pages/me/study/study.wxml` with:

```wxml
<view class="stub">
  <dx-icon name="clock" size="48px" color="#9ca3af" />
  <text class="stub-title">敬请期待</text>
  <text class="stub-desc">此功能正在开发中</text>
</view>
```

- [ ] **Step 4: Write `study.wxss`**

Create `dx-mini/miniprogram/pages/me/study/study.wxss` with:

```css
.stub {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  gap: 16rpx;
  color: #9ca3af;
}
.stub-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin-top: 8rpx;
}
.stub-desc {
  font-size: 13px;
  color: #9ca3af;
}
```

- [ ] **Step 5: Create `tasks/`, `community/`, `feedback/` with the same four files each**

Repeat steps 1–4 for each of the remaining three stubs. The only differences per stub: the directory name, the filenames, and `navigationBarTitleText`. Everything else is byte-identical to the `study` version.

| Directory | Filename prefix | `navigationBarTitleText` |
|---|---|---|
| `dx-mini/miniprogram/pages/me/tasks/` | `tasks` | `"打卡"` |
| `dx-mini/miniprogram/pages/me/community/` | `community` | `"留言"` |
| `dx-mini/miniprogram/pages/me/feedback/` | `feedback` | `"建议"` |

For `tasks.json`:

```json
{
  "navigationBarTitleText": "打卡",
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

For `tasks.ts`:

```typescript
Page({})
```

For `tasks.wxml`:

```wxml
<view class="stub">
  <dx-icon name="clock" size="48px" color="#9ca3af" />
  <text class="stub-title">敬请期待</text>
  <text class="stub-desc">此功能正在开发中</text>
</view>
```

For `tasks.wxss`:

```css
.stub {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  gap: 16rpx;
  color: #9ca3af;
}
.stub-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin-top: 8rpx;
}
.stub-desc {
  font-size: 13px;
  color: #9ca3af;
}
```

Do the same for `community/community.{json,ts,wxml,wxss}` (title `"留言"`) and `feedback/feedback.{json,ts,wxml,wxss}` (title `"建议"`). The `.ts`, `.wxml`, and `.wxss` files are identical across all four stubs.

- [ ] **Step 6: Register the four pages in `app.json`**

Edit `dx-mini/miniprogram/app.json`. Locate the `"pages"` array. After the existing `"pages/me/purchase/purchase"` entry (the last entry today), append four new strings:

```json
  "pages": [
    "pages/login/login",
    "pages/home/home",
    "pages/games/games",
    "pages/games/detail/detail",
    "pages/games/play/play",
    "pages/games/favorites/favorites",
    "pages/leaderboard/leaderboard",
    "pages/learn/learn",
    "pages/learn/mastered/mastered",
    "pages/learn/unknown/unknown",
    "pages/learn/review/review",
    "pages/me/me",
    "pages/me/profile-edit/profile-edit",
    "pages/me/notices/notices",
    "pages/me/groups/groups",
    "pages/me/groups-detail/groups-detail",
    "pages/me/invite/invite",
    "pages/me/redeem/redeem",
    "pages/me/purchase/purchase",
    "pages/me/study/study",
    "pages/me/tasks/tasks",
    "pages/me/community/community",
    "pages/me/feedback/feedback"
  ],
```

Do not touch any other key (`tabBar`, `window`, `style`, `componentFramework` stay as-is).

- [ ] **Step 7: Verify the page tree with TypeScript**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx -p typescript@5 tsc --noEmit 2>&1 | grep -v "miniprogram_npm" | head -30
```

Expected: no new TS errors beyond the existing tolerated `this`-in-`Component` pattern (which doesn't appear in `Page({})` anyway).

- [ ] **Step 8: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/study dx-mini/miniprogram/pages/me/tasks dx-mini/miniprogram/pages/me/community dx-mini/miniprogram/pages/me/feedback dx-mini/miniprogram/app.json
git commit -m "feat(mini): scaffold study/tasks/community/feedback stub pages"
```

---

## Task 4: Rewrite the home page top section

**Files:**
- Modify: `dx-mini/miniprogram/pages/home/home.ts`
- Modify: `dx-mini/miniprogram/pages/home/home.wxml`
- Modify: `dx-mini/miniprogram/pages/home/home.wxss`
- Unchanged: `dx-mini/miniprogram/pages/home/home.json`

- [ ] **Step 1: Rewrite `home.ts`**

Replace the entire contents of `dx-mini/miniprogram/pages/home/home.ts` with:

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
}

interface Greeting { title: string; subtitle: string }

interface DashboardData {
  profile: DashboardProfile
  greeting: Greeting
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
      const dash = await api.get<DashboardData>('/api/hall/dashboard')
      this.setData({
        loading: false,
        profile: dash.profile,
        greeting: dash.greeting,
        gradeLabelText: gradeLabel(dash.profile.grade),
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

Differences from the current file, for review:
- Removed `MasterStats`, `ReviewStats`, `HeatmapDay`, `HeatmapData`, `HeatmapCell` interfaces.
- Removed `masterStats`, `reviewStats`, `todayAnswers`, `heatmapCells`, `unreadNotices` from `Page.data`.
- Removed `buildHeatmapCells()`.
- Removed the `api.get<HeatmapData>('/api/hall/heatmap')` call; `loadData` now fetches only the dashboard.
- Added `level: number` to `DashboardProfile`.
- Added `gradeLabelText` computed at load time (so WXML can bind `{{gradeLabelText}}` without a method call).
- Added 7 navigation handlers (`goPurchase`, `goInvite`, `goStudy`, `goGroups`, `goTasks`, `goCommunity`, `goFeedback`); kept `goSearch`.
- Imported `gradeLabel` from `utils/format`.

- [ ] **Step 2: Rewrite `home.wxml`**

Replace the entire contents of `dx-mini/miniprogram/pages/home/home.wxml` with:

```wxml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <!-- Teal header band -->
    <view class="teal-wrap">
      <view class="status-bar-spacer"></view>

      <view class="nav-row">
        <text class="nav-greet" wx:if="{{profile}}">{{greeting ? greeting.title : '你好'}} {{profile.nickname || profile.username}}</text>
        <text class="nav-greet" wx:else>你好</text>
      </view>

      <van-skeleton title row="2" loading="{{loading}}">
        <view class="greet-body">
          <text class="greet-sub">{{greeting ? greeting.subtitle : ''}}</text>
          <view class="badge-row" wx:if="{{profile}}">
            <text class="badge lvl">Lv.{{profile.level}}</text>
            <text class="badge">{{gradeLabelText}}</text>
          </view>
        </view>
      </van-skeleton>

      <view class="search-row">
        <view class="search-box" bind:tap="goSearch">
          <dx-icon name="search" size="16px" color="#9ca3af" />
          <text class="search-placeholder">搜索课程</text>
        </view>
      </view>
    </view>

    <!-- Combined VIP / 奖励 card -->
    <view class="combined-card">
      <view class="card-half" bind:tap="goPurchase">
        <view class="card-text">
          <text class="card-title">升级 VIP</text>
          <text class="card-desc">选择适合您的会员方案，或通过兑换码升级会员，解锁更多学习功能</text>
        </view>
        <view class="illo illo-vip">
          <dx-icon name="crown" size="20px" color="#ffffff" />
        </view>
      </view>
      <view class="divider"></view>
      <view class="card-half" bind:tap="goInvite">
        <view class="card-text">
          <text class="card-title">奖励计划</text>
          <text class="card-desc">如果喜欢斗学就推荐给好朋友一起来快乐学习吧！</text>
        </view>
        <view class="illo illo-gift">
          <dx-icon name="gift" size="20px" color="#ffffff" />
        </view>
      </view>
    </view>

    <!-- Five gradient circles -->
    <view class="circle-row">
      <view class="circle-item" bind:tap="goStudy">
        <view class="circle c1"><dx-icon name="chart-pie" size="22px" color="#ffffff" /></view>
        <text class="circle-label">学习</text>
      </view>
      <view class="circle-item" bind:tap="goGroups">
        <view class="circle c2"><dx-icon name="users" size="22px" color="#ffffff" /></view>
        <text class="circle-label">群组</text>
      </view>
      <view class="circle-item" bind:tap="goTasks">
        <view class="circle c3"><dx-icon name="calendar-check" size="22px" color="#ffffff" /></view>
        <text class="circle-label">打卡</text>
      </view>
      <view class="circle-item" bind:tap="goCommunity">
        <view class="circle c4"><dx-icon name="sticker" size="22px" color="#ffffff" /></view>
        <text class="circle-label">留言</text>
      </view>
      <view class="circle-item" bind:tap="goFeedback">
        <view class="circle c5"><dx-icon name="flag" size="22px" color="#ffffff" /></view>
        <text class="circle-label">建议</text>
      </view>
    </view>
  </view>
</van-config-provider>
```

Differences from the current file, for review:
- Removed `.top-bar` (the old search-box row at the top — the new design embeds search inside the teal band).
- Removed `.hero-section` (greeting title, subtitle, 3 stat cards).
- Removed `.section` with the heatmap.
- Added `.teal-wrap`, `.nav-row`, `.greet-body`, `.search-row`, `.combined-card`, `.circle-row` structure.
- Added 7 new `bind:tap` handlers on tiles.
- Kept `<van-config-provider>` + `--status-bar-height` CSS var plumbing.
- Kept `<van-skeleton>` around the dynamic portion so the loading shimmer still appears.

- [ ] **Step 3: Rewrite `home.wxss`**

Replace the entire contents of `dx-mini/miniprogram/pages/home/home.wxss` with:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}

/* ----- Teal header band ----- */

.teal-wrap {
  background: linear-gradient(180deg, #0d9488 0%, #0f766e 100%);
  color: #ffffff;
  padding-bottom: 18px;
  border-bottom-left-radius: 50% 28px;
  border-bottom-right-radius: 50% 28px;
}

.page-container.dark .teal-wrap {
  background: linear-gradient(180deg, #14b8a6 0%, #0d9488 100%);
}

.status-bar-spacer {
  height: var(--status-bar-height, 20px);
}

.nav-row {
  position: relative;
  height: 40px;
  display: flex;
  align-items: flex-end;
  padding: 0 18px 2px;
  color: #ffffff;
}

.nav-greet {
  font-size: 17px;
  font-weight: 700;
  line-height: 1.1;
  padding-right: 220rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.greet-body {
  padding: 0 20px;
}

.greet-sub {
  font-size: 13px;
  opacity: 0.92;
  line-height: 1.35;
  margin-top: 2px;
  display: block;
}

.badge-row {
  margin-top: 6px;
  display: flex;
  gap: 8px;
}

.badge {
  font-size: 11px;
  padding: 3px 9px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.2);
  color: #ffffff;
}

.badge.lvl {
  background: rgba(255, 255, 255, 0.32);
  font-weight: 600;
}

.search-row {
  padding: 14px 16px 0;
}

.search-box {
  background: #ffffff;
  border-radius: 22px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.12);
}

.search-placeholder {
  font-size: 13px;
  color: #9ca3af;
}

.page-container.dark .search-box {
  background: var(--bg-card);
}

.page-container.dark .search-placeholder {
  color: #6b7280;
}

/* ----- Combined VIP / 奖励 card ----- */

.combined-card {
  margin: -12px 14px 0;
  background: linear-gradient(180deg, #f0fdfa 0%, #f0fdfa 55%, #fdf2f8 100%);
  border-radius: 14px;
  box-shadow: 0 6px 18px rgba(13, 148, 136, 0.10);
  display: grid;
  grid-template-columns: 1fr 1px 1fr;
  overflow: hidden;
  position: relative;
  z-index: 2;
}

.page-container.dark .combined-card {
  background: linear-gradient(180deg, rgba(20, 184, 166, 0.08) 0%, rgba(20, 184, 166, 0.08) 55%, rgba(236, 72, 153, 0.08) 100%);
}

.card-half {
  padding: 14px 12px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f766e;
  display: block;
}

.page-container.dark .card-title {
  color: #2dd4bf;
}

.card-desc {
  font-size: 10px;
  color: #4b5563;
  line-height: 1.45;
  display: block;
}

.page-container.dark .card-desc {
  color: #9ca3af;
}

.illo {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.illo-vip {
  background: linear-gradient(135deg, #0d9488, #14b8a6);
}

.illo-gift {
  background: linear-gradient(135deg, #ec4899, #f472b6);
}

.divider {
  background: rgba(148, 163, 184, 0.25);
  margin: 12px 0;
}

.page-container.dark .divider {
  background: rgba(255, 255, 255, 0.08);
}

/* ----- Five gradient-teal circles ----- */

.circle-row {
  padding: 22px 14px 18px;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.circle-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(13, 148, 136, 0.22);
}

.circle.c1 { background: linear-gradient(135deg, #0d9488, #14b8a6); }
.circle.c2 { background: linear-gradient(135deg, #0f766e, #2dd4bf); }
.circle.c3 { background: linear-gradient(135deg, #0891b2, #14b8a6); }
.circle.c4 { background: linear-gradient(135deg, #059669, #14b8a6); }
.circle.c5 { background: linear-gradient(135deg, #115e59, #0d9488); }

.circle-label {
  font-size: 12px;
  color: #1a1a1a;
}

.page-container.dark .circle-label {
  color: #f5f5f5;
}
```

Differences from the current file, for review:
- Dropped `.top-bar`, `.search-box` (old), `.search-placeholder` (old), `.hero-section`, `.greeting-title`, `.greeting-subtitle`, `.stat-row`, `.stat-card`, `.stat-value`, `.stat-label`, `.section`, `.section-title`, `.heatmap`, `.heatmap-cell`, `.heatmap-cell.level-*`, and `.dark .heatmap-cell.level-4` rules.
- Dropped the `padding-top: calc(var(--status-bar-height, 20px) + 88rpx)` rule — the new layout uses an explicit `.status-bar-spacer` inside the teal band instead of top padding on the container, so the teal color starts at the very top edge.
- Added all the new rules listed above.

- [ ] **Step 4: Sanity-check that `home.json` still declares the needed components**

Run:

```bash
cat /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/home/home.json
```

Expected output (unchanged from current):

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-skeleton": "@vant/weapp/skeleton/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

If anything is missing or different, do NOT edit in this step — flag for review. The new WXML uses `van-config-provider`, `van-skeleton`, and `dx-icon`, all already declared.

- [ ] **Step 5: Run TypeScript check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx -p typescript@5 tsc --noEmit 2>&1 | grep -v "miniprogram_npm" | head -40
```

Expected: no new TS errors beyond the existing tolerated `this`-in-`Component` pattern. The new `home.ts` uses `Page({...})` (not `Component`), so that specific tolerated pattern does not appear in it.

- [ ] **Step 6: Run the icon build scanner to confirm WXML references are valid**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected output: `Wrote 25 icons to miniprogram/components/dx-icon/icons.ts.` — and NO `<dx-icon name="..."/> not in ICONS` errors. If it throws for any of `chart-pie`, `calendar-check`, `sticker`, `flag`, Task 2 was not applied correctly — go back.

- [ ] **Step 7: Verify the static-check grep list from the spec**

Run each expected-zero grep:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
grep -rn "hero-section\|greeting-title\|heatmap-cell" dx-mini/miniprogram/pages/home/
grep -rn "heatmap\|buildHeatmapCells\|/api/hall/heatmap" dx-mini/miniprogram/pages/home/
```

Expected: no matches for either.

Then confirm the dx-icon count:

```bash
grep -c "<dx-icon" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/home/home.wxml
```

Expected: `8` (search + crown + gift + 5 circles).

- [ ] **Step 8: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/home/home.ts dx-mini/miniprogram/pages/home/home.wxml dx-mini/miniprogram/pages/home/home.wxss
git commit -m "feat(mini): home top-section hub redesign"
```

---

## Task 5: End-to-end smoke test + merge to main

**Files:** none modified; this task is pure verification.

- [ ] **Step 1: Launch dx-api locally**

In one terminal:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .
```

Expected: server starts on `http://localhost:3001` with no panics.

- [ ] **Step 2: Confirm the dashboard response includes `level`**

In another terminal, obtain a valid user JWT (from a prior login response or DevTools localStorage), then:

```bash
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:3001/api/hall/dashboard | python3 -m json.tool | head -20
```

Expected: the `profile` object inside `data` contains a `"level": <int>` key with a non-negative number.

- [ ] **Step 3: Open WeChat DevTools on the dx-mini project**

Project root: `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini`.

In DevTools console (after login completes and home loads):

```js
require('./utils/config').setDevApiBaseUrl('http://localhost')
```

Refresh the simulator. The home page should render the new top section.

- [ ] **Step 4: Light-mode visual check**

Confirm on the home tab:

- Teal band with greeting title (e.g. `下午好 ☕ 小明`) on the same row as the capsule, ellipsis-clipping if long.
- Subtitle and Lv/grade chips directly below, tight spacing.
- Search bar inside the teal, oval curve finishes just below it.
- Combined card riding up 12px into the teal; pale teal at top fading to pale pink at bottom.
- VIP half shows crown icon; 奖励 half shows gift icon; 1px divider between.
- Five gradient-teal circles (chart-pie, users, calendar-check, sticker, flag) with labels 学习 / 群组 / 打卡 / 留言 / 建议.
- Nothing below the circle row.

- [ ] **Step 5: Dark-mode visual check**

Switch to the me tab, tap the moon icon to toggle dark mode, return to the home tab. Confirm:

- Teal band still renders (slightly brighter top stop per dark-mode variant).
- Search box is dark-surfaced (not white).
- Card gradient is the rgba-dimmed teal/pink.
- Card title color changes to teal-mint (`#2dd4bf`).
- Circle labels flip to light.

Toggle back to light mode via the moon icon before proceeding.

- [ ] **Step 6: Tap-through check — every tile navigates correctly**

Tap each in order, back out to home between taps:

- 升级 VIP card → `/pages/me/purchase/purchase` — existing purchase screen.
- 奖励计划 card → `/pages/me/invite/invite` — existing invite screen.
- 学习 circle → `/pages/me/study/study` — "敬请期待" stub.
- 群组 circle → `/pages/me/groups/groups` — existing groups list.
- 打卡 circle → `/pages/me/tasks/tasks` — stub.
- 留言 circle → `/pages/me/community/community` — stub.
- 建议 circle → `/pages/me/feedback/feedback` — stub.
- Search box → `/pages/games/games` — existing course list.

- [ ] **Step 7: Long-nickname stress test**

In DevTools console, temporarily patch the profile render to verify ellipsis handling:

```js
getCurrentPages()[0].setData({ profile: { ...getCurrentPages()[0].data.profile, nickname: '非常非常非常非常非常非常长的昵称测试' } })
```

Expected: greeting title ellipsis-clips; no overlap with the capsule; no line wrap.

Refresh the simulator to discard the patch.

- [ ] **Step 8: Real-device preview**

In DevTools, click 预览 (preview). Scan the QR with WeChat → open the 小程序助手 → open the page. Do NOT use 真机调试 — broken on current DevTools version per project memory.

Confirm on the real device:

- Oval bottom on teal band renders (not a flat edge).
- Gradient colors reproduce on iOS and Android (if both available).
- All taps navigate correctly.

- [ ] **Step 9: Merge to main and push**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git checkout main
git merge --no-ff feat/mini-home-hub-redesign -m "Merge branch 'feat/mini-home-hub-redesign'"
git push origin main
```

Do NOT push the feature branch (project memory: merge feature branches to main locally and push only main; never push feature branches to remote).

- [ ] **Step 10: Clean up the feature branch locally (optional)**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git branch -d feat/mini-home-hub-redesign
```

Expected: `Deleted branch feat/mini-home-hub-redesign (was <sha>).` If git complains about unmerged commits, double-check the merge succeeded.
