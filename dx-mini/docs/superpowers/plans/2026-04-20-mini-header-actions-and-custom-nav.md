# dx-mini: Tab-Page Custom Nav + Header-Action Move Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the dark-mode toggle + notifications bell from the home page to the me page, drop the now-redundant 公告通知 van-cell, and remove the default native navigation bar on all four tab pages (首页, 课程, 排行榜, 我的) via a shared status-bar padding idiom.

**Architecture:** Each affected page receives three coordinated edits — a `.json` flip to `navigationStyle: "custom"`, a `.ts` hook (`onLoad`) that computes `statusBarHeight` from `wx.getSystemInfoSync()` and exposes it on page data, and a `.wxml` top-level wrapper that injects that value as the `--status-bar-height` CSS custom property. Each page's `.wxss` then reserves vertical space below the status bar sufficient to clear the WeChat capsule button (`calc(var(--status-bar-height, 20px) + 88rpx)`). The home → me button relocation is bundled into the me and home edits so every page file is touched at most once.

**Tech Stack:** WeChat Mini Program (glass-easel, native WXML/WXSS), TypeScript strict, Vant Weapp 1.11.x, `<dx-icon>` (SVG renderer).

**Spec:** [2026-04-20-mini-header-actions-and-custom-nav-design.md](../specs/2026-04-20-mini-header-actions-and-custom-nav-design.md)

**Branch:** `feat/mini-tab-page-chrome` (already created; spec already committed as `fba99bf`).

---

## Task 1: Update the me page

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.ts`
- Modify: `dx-mini/miniprogram/pages/me/me.wxml`
- Modify: `dx-mini/miniprogram/pages/me/me.wxss`
- Modify: `dx-mini/miniprogram/pages/me/me.json`

- [ ] **Step 1: Rewrite `me.ts`**

Replace the entire file contents with:

```typescript
import { api } from '../../utils/api'
import { formatDate, gradeLabel } from '../../utils/format'
import { clearToken } from '../../utils/auth'
import { ws } from '../../utils/ws'

interface ProfileData {
  id: string; grade: string; username: string; nickname: string | null
  avatarUrl: string | null; city: string | null; beans: number
  exp: number; level: number; inviteCode: string; currentPlayStreak: number
  vipDueAt: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    arrowColor: '#9ca3af',
    cellIconColor: '#6b7280',
    loading: true,
    profile: null as ProfileData | null,
    avatarChar: '',
    statusBarHeight: 20,
    formatDate,
    gradeLabel,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ statusBarHeight })
  },
  onShow() {
    const theme = app.globalData.theme
    this.setData({
      theme,
      primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
      arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',
      cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
    });
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 4, theme }) }
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      const avatarChar = (profile.nickname || profile.username).charAt(0)
      this.setData({ loading: false, profile, avatarChar })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleTheme() {
    const next: 'light' | 'dark' = this.data.theme === 'light' ? 'dark' : 'light'
    wx.setStorageSync('dx_theme', next)
    app.globalData.theme = next
    this.setData({
      theme: next,
      primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
      arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',
      cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
    })
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ theme: next }) }
  },
  goProfileEdit() { wx.navigateTo({ url: '/pages/me/profile-edit/profile-edit' }) },
  goNotices() { wx.navigateTo({ url: '/pages/me/notices/notices' }) },
  goGroups() { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  goInvite() { wx.navigateTo({ url: '/pages/me/invite/invite' }) },
  goRedeem() { wx.navigateTo({ url: '/pages/me/redeem/redeem' }) },
  goPurchase() { wx.navigateTo({ url: '/pages/me/purchase/purchase' }) },
  logout() {
    wx.showModal({
      title: '退出登录',
      content: '确定退出？',
      success: (res) => {
        if (!res.confirm) return
        clearToken()
        ws.disconnect()
        wx.reLaunch({ url: '/pages/login/login' })
      },
    })
  },
})
```

Changes vs. current file: added `statusBarHeight: 20` to `data`, added `onLoad`, added `toggleTheme` method. Everything else unchanged.

- [ ] **Step 2: Rewrite `me.wxml`**

Replace the entire file contents with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="top-bar">
      <dx-icon
        name="{{theme === 'dark' ? 'sun' : 'moon'}}"
        size="22px"
        color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        bind:click="toggleTheme"
      />
      <dx-icon
        name="bell"
        size="22px"
        color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}"
        bind:click="goNotices"
      />
    </view>

    <van-loading wx:if="{{loading}}" size="30px" color="{{primaryColor}}" class="center-loader" />
    <block wx:if="{{!loading && profile}}">
      <view class="profile-header" bind:tap="goProfileEdit">
        <view class="avatar-wrap">
          <van-image
            wx:if="{{profile.avatarUrl}}"
            src="{{profile.avatarUrl}}"
            width="64px"
            height="64px"
            radius="50%"
            fit="cover"
          />
          <view wx:else class="avatar-fallback">
            <text>{{avatarChar}}</text>
          </view>
        </view>
        <view class="profile-info">
          <text class="profile-name">{{profile.nickname || profile.username}}</text>
          <view class="profile-badges">
            <text class="grade-badge">{{gradeLabel(profile.grade)}}</text>
            <text class="exp-badge">Lv.{{profile.level}}</text>
          </view>
        </view>
        <dx-icon name="chevron-right" size="16px" color="{{arrowColor}}" />
      </view>

      <view class="stats-bar">
        <view class="stat-item">
          <text class="stat-v">{{profile.beans}}</text>
          <text class="stat-l">金豆</text>
        </view>
        <view class="stat-divider" />
        <view class="stat-item">
          <text class="stat-v">{{profile.exp}}</text>
          <text class="stat-l">经验值</text>
        </view>
        <view class="stat-divider" />
        <view class="stat-item">
          <text class="stat-v">{{profile.currentPlayStreak}}</text>
          <text class="stat-l">连续天数</text>
        </view>
      </view>

      <view wx:if="{{profile.vipDueAt}}" class="vip-bar">
        <dx-icon name="crown" size="16px" color="#d97706" />
        <text class="vip-text">会员有效期至 {{formatDate(profile.vipDueAt)}}</text>
      </view>

      <van-cell-group inset custom-style="margin:16px;">
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
      </van-cell-group>

      <van-cell-group inset custom-style="margin:0 16px 16px;">
        <van-cell title="退出登录" title-style="color:var(--van-danger-color);" bind:click="logout" />
      </van-cell-group>
    </block>
  </view>
</van-config-provider>
```

Changes vs. current file: added `style="--status-bar-height: …"` on `.page-container`, prepended new `.top-bar` view, removed the `<van-cell title="公告通知" is-link bind:click="goNotices">` row (now the bell icon handles that). All other markup unchanged.

- [ ] **Step 3: Edit `me.wxss` — add top-bar rule and extend `.page-container`**

Replace the existing first line:

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 100rpx; }
```

with:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
  padding-bottom: 100rpx;
}
.top-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 20rpx;
  padding: 12rpx 32rpx;
}
```

All other rules in `me.wxss` stay unchanged.

- [ ] **Step 4: Edit `me.json` — enable custom nav**

Replace file contents with:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-cell": "@vant/weapp/cell/index",
    "van-cell-group": "@vant/weapp/cell-group/index",
    "dx-icon": "/components/dx-icon/index",
    "van-image": "@vant/weapp/image/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

Changes vs. current file: replaced `"navigationBarTitleText": "我的"` with `"navigationStyle": "custom"`. `usingComponents` unchanged.

- [ ] **Step 5: Verify no van-cell for 公告通知 remains**

Run: `cd dx-mini && grep -n "公告通知" miniprogram/pages/me/me.wxml`
Expected: no matches.

- [ ] **Step 6: Verify toggleTheme is present on me.ts**

Run: `cd dx-mini && grep -n "toggleTheme" miniprogram/pages/me/me.ts`
Expected: exactly two matches (method definition + one usage inside — or one definition only; either way, `grep -c` ≥ 1).

---

## Task 2: Update the home page

**Files:**
- Modify: `dx-mini/miniprogram/pages/home/home.ts`
- Modify: `dx-mini/miniprogram/pages/home/home.wxml`
- Modify: `dx-mini/miniprogram/pages/home/home.wxss`
- Modify: `dx-mini/miniprogram/pages/home/home.json`

- [ ] **Step 1: Rewrite `home.ts`**

Replace the entire file contents with:

```typescript
import { api } from '../../utils/api'

interface DashboardProfile {
  id: string
  username: string
  nickname: string | null
  grade: string
  exp: number
  beans: number
  avatarUrl: string | null
  currentPlayStreak: number
  inviteCode: string
  lastReadNoticeAt: string | null
}

interface MasterStats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }
interface Greeting { title: string; subtitle: string }

interface DashboardData {
  profile: DashboardProfile
  masterStats: MasterStats
  reviewStats: ReviewStats
  todayAnswers: number
  greeting: Greeting
}

interface HeatmapDay { date: string; count: number }
interface HeatmapData { year: number; days: HeatmapDay[]; accountYear: number }

interface HeatmapCell { date: string; level: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    profile: null as DashboardProfile | null,
    masterStats: null as MasterStats | null,
    reviewStats: null as ReviewStats | null,
    todayAnswers: 0,
    greeting: null as Greeting | null,
    heatmapCells: [] as HeatmapCell[],
    unreadNotices: false,
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    const tabBar = this.getTabBar() as any
    if (tabBar) { tabBar.setData({ active: 0, theme: app.globalData.theme }) }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    try {
      const [dash, heatmap] = await Promise.all([
        api.get<DashboardData>('/api/hall/dashboard'),
        api.get<HeatmapData>('/api/hall/heatmap'),
      ])
      const cells = this.buildHeatmapCells(heatmap.days)
      this.setData({
        loading: false,
        profile: dash.profile,
        masterStats: dash.masterStats,
        reviewStats: dash.reviewStats,
        todayAnswers: dash.todayAnswers,
        greeting: dash.greeting,
        heatmapCells: cells,
        unreadNotices: false,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  buildHeatmapCells(days: HeatmapDay[]): HeatmapCell[] {
    const map = new Map(days.map(d => [d.date, d.count]))
    const cells: HeatmapCell[] = []
    const today = new Date()
    for (let i = 48; i >= 0; i--) {
      const d = new Date(today)
      d.setDate(d.getDate() - i)
      const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
      const count = map.get(key) || 0
      const level = count === 0 ? 0 : count < 3 ? 1 : count < 6 ? 2 : count < 10 ? 3 : 4
      cells.push({ date: key, level })
    }
    return cells
  },
  goSearch() {
    wx.navigateTo({ url: '/pages/games/games' })
  },
})
```

Changes vs. current file: removed `toggleTheme` and `goNotices` methods, added `statusBarHeight: 20` to `data`, extended `onLoad` to compute statusBarHeight. All other logic unchanged.

- [ ] **Step 2: Rewrite `home.wxml`**

Replace the entire file contents with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <!-- Top bar -->
    <view class="top-bar">
      <view class="search-box" bind:tap="goSearch">
        <dx-icon name="search" size="16px" color="#9ca3af" />
        <text class="search-placeholder">搜索课程</text>
      </view>
    </view>

    <van-skeleton title row="5" loading="{{loading}}">
      <!-- Greeting + stats -->
      <view class="hero-section">
        <text class="greeting-title">{{greeting ? greeting.title : '你好！'}}</text>
        <text wx:if="{{greeting}}" class="greeting-subtitle">{{greeting.subtitle}}</text>
        <view class="stat-row">
          <view class="stat-card">
            <text class="stat-value">{{profile.currentPlayStreak || 0}}</text>
            <text class="stat-label">连续天数</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">{{masterStats.total || 0}}</text>
            <text class="stat-label">已掌握</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">{{todayAnswers || 0}}</text>
            <text class="stat-label">今日答题</text>
          </view>
        </view>
      </view>

      <!-- Heatmap -->
      <view class="section">
        <text class="section-title">学习热力图</text>
        <view class="heatmap">
          <view
            wx:for="{{heatmapCells}}"
            wx:key="date"
            class="heatmap-cell level-{{item.level}}"
          ></view>
        </view>
      </view>
    </van-skeleton>
  </view>
</van-config-provider>
```

Changes vs. current file: added `style="--status-bar-height: …"` on `.page-container`, dropped the entire `<view class="top-actions">` block and its two `<dx-icon>` children. All other markup unchanged.

- [ ] **Step 3: Edit `home.wxss` — add page-padding and drop dead rule**

Make two edits to `miniprogram/pages/home/home.wxss`:

First, replace:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
```

with:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
  padding-bottom: 100rpx;
}
```

Second, delete the now-unused `.top-actions` block:

```css
.top-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
```

All other rules in `home.wxss` stay unchanged. `.search-box` already has `flex: 1` so the search bar will naturally fill the row once the `.top-actions` sibling is gone.

- [ ] **Step 4: Edit `home.json` — enable custom nav**

Replace file contents with:

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

Changes: replaced `"navigationBarTitleText": "斗学"` with `"navigationStyle": "custom"`. `usingComponents` unchanged.

- [ ] **Step 5: Verify the two removed handlers are gone**

Run: `cd dx-mini && grep -n "toggleTheme\|goNotices" miniprogram/pages/home/home.ts miniprogram/pages/home/home.wxml`
Expected: no matches.

- [ ] **Step 6: Verify no `.top-actions` rule remains**

Run: `cd dx-mini && grep -n "top-actions" miniprogram/pages/home/home.wxss miniprogram/pages/home/home.wxml`
Expected: no matches.

---

## Task 3: Update the games page

**Files:**
- Modify: `dx-mini/miniprogram/pages/games/games.ts`
- Modify: `dx-mini/miniprogram/pages/games/games.wxml`
- Modify: `dx-mini/miniprogram/pages/games/games.wxss`
- Modify: `dx-mini/miniprogram/pages/games/games.json`

- [ ] **Step 1: Edit `games.ts` — add statusBarHeight to data, extend onLoad**

In `miniprogram/pages/games/games.ts`:

Add `statusBarHeight: 20,` to the `data` block, immediately after the `hasMore: false,` line so the block looks like:

```typescript
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    categories: [{ id: '', name: '全部' }] as Category[],
    activeCategoryId: '',
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
    statusBarHeight: 20,
  },
```

Replace the existing `onLoad`:

```typescript
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadCategories()
    this.loadGames(true)
  },
```

with:

```typescript
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
    this.loadCategories()
    this.loadGames(true)
  },
```

All other methods unchanged.

- [ ] **Step 2: Edit `games.wxml` — inject CSS variable**

Replace the opening `.page-container` line:

```xml
  <view class="page-container {{theme}}">
```

with:

```xml
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
```

All other markup unchanged.

- [ ] **Step 3: Edit `games.wxss` — add top padding**

Open `miniprogram/pages/games/games.wxss`. Find the existing `.page-container` rule and add the `padding-top` line. The rule should end up looking like (merge with whatever other declarations it has):

```css
.page-container {
  /* existing declarations stay */
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
}
```

If the file has no `.page-container` rule yet, add the full rule at the top:

```css
.page-container {
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
}
```

- [ ] **Step 4: Edit `games.json` — enable custom nav**

Replace file contents with:

```json
{
  "navigationStyle": "custom",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "dx-icon": "/components/dx-icon/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-image": "@vant/weapp/image/index"
  }
}
```

Changes: replaced `"navigationBarTitleText": "课程"` with `"navigationStyle": "custom"`. `enablePullDownRefresh` and `usingComponents` unchanged.

- [ ] **Step 5: Verify the CSS var injection**

Run: `cd dx-mini && grep -n "status-bar-height" miniprogram/pages/games/games.wxml`
Expected: exactly one match on the `.page-container` wrapper.

---

## Task 4: Update the leaderboard page

**Files:**
- Modify: `dx-mini/miniprogram/pages/leaderboard/leaderboard.ts`
- Modify: `dx-mini/miniprogram/pages/leaderboard/leaderboard.wxml`
- Modify: `dx-mini/miniprogram/pages/leaderboard/leaderboard.wxss`
- Modify: `dx-mini/miniprogram/pages/leaderboard/leaderboard.json`

- [ ] **Step 1: Edit `leaderboard.ts` — add statusBarHeight, extend onLoad**

In `miniprogram/pages/leaderboard/leaderboard.ts`:

Add `statusBarHeight: 20,` to the `data` block after `myRank: null as LeaderboardEntry | null,`:

```typescript
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    period: 'month' as 'day' | 'week' | 'month',
    lbType: 'exp' as 'exp' | 'playtime',
    entries: [] as LeaderboardEntry[],
    entries4Plus: [] as LeaderboardEntry[],
    myRank: null as LeaderboardEntry | null,
    statusBarHeight: 20,
  },
```

Replace the existing `onLoad`:

```typescript
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadLeaderboard()
  },
```

with:

```typescript
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({ theme: app.globalData.theme, statusBarHeight })
    this.loadLeaderboard()
  },
```

All other methods unchanged.

- [ ] **Step 2: Edit `leaderboard.wxml` — inject CSS variable**

Replace the opening `.page-container` line:

```xml
  <view class="page-container {{theme}}">
```

with:

```xml
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
```

All other markup unchanged.

- [ ] **Step 3: Edit `leaderboard.wxss` — add top padding**

Open `miniprogram/pages/leaderboard/leaderboard.wxss`. Find the existing `.page-container` rule and add the `padding-top` line. The rule should end up looking like:

```css
.page-container {
  /* existing declarations stay */
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
}
```

If the file has no `.page-container` rule yet, add the full rule at the top:

```css
.page-container {
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
}
```

- [ ] **Step 4: Edit `leaderboard.json` — enable custom nav**

Replace file contents with:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index"
  }
}
```

Changes: replaced `"navigationBarTitleText": "排行榜"` with `"navigationStyle": "custom"`. `usingComponents` unchanged.

- [ ] **Step 5: Verify the CSS var injection**

Run: `cd dx-mini && grep -n "status-bar-height" miniprogram/pages/leaderboard/leaderboard.wxml`
Expected: exactly one match on the `.page-container` wrapper.

---

## Task 5: Repo-wide static verification

**Files:** read-only

- [ ] **Step 1: No residual title-bar config on the four tab pages**

Run:
```bash
cd dx-mini && grep -n "navigationBarTitleText" \
  miniprogram/pages/home/home.json \
  miniprogram/pages/games/games.json \
  miniprogram/pages/me/me.json \
  miniprogram/pages/leaderboard/leaderboard.json
```
Expected: no matches.

- [ ] **Step 2: Custom nav is set on exactly those four pages**

Run:
```bash
cd dx-mini && grep -rn "navigationStyle.*custom" miniprogram/pages
```
Expected: five matches — the four tab pages plus the pre-existing `pages/login/login.json`.

- [ ] **Step 3: No residual 公告通知 cell**

Run: `cd dx-mini && grep -rn "公告通知" miniprogram/pages/me/`
Expected: no matches.

- [ ] **Step 4: Home no longer defines the moved handlers**

Run: `cd dx-mini && grep -n "toggleTheme\|goNotices" miniprogram/pages/home/`
Expected: no matches.

- [ ] **Step 5: Me page exposes toggleTheme**

Run: `cd dx-mini && grep -n "toggleTheme" miniprogram/pages/me/me.ts miniprogram/pages/me/me.wxml`
Expected: matches in both files (method definition in .ts, `bind:click` in .wxml).

- [ ] **Step 6: TypeScript still clean (no new error patterns)**

Run: `cd dx-mini && npx -y -p typescript@5.5 tsc --noEmit 2>&1 | grep -E "^miniprogram/pages/(home|games|me|leaderboard)/"`
Expected: no output, OR output that matches the tolerated `this`-in-`Component({methods})` TS2339 pattern only (same shape as the existing `custom-tab-bar/index.ts(22,12): error TS2339: Property 'setData' does not exist` error). No new error categories.

---

## Task 6: Manual DevTools smoke test (user)

This task is performed by the user. Hand off and wait for confirmation before committing.

- [ ] **Step 1: Compile in DevTools**

Open `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini` in WeChat Developer Tools. If `miniprogram_npm` is stale, run 构建 npm. Click 编译 / 预览. Expected: no console errors on boot.

- [ ] **Step 2: Verify each tab page**

Check each tab-page pass in both light and dark theme where applicable:

  - [ ] **首页** — default nav bar is GONE. Capsule button still visible top-right and tappable. Status bar clear. Search box fills the row horizontally (no right-side gap). No dark-mode toggle or bell on this page anymore. Tapping the search bar still navigates to /pages/games/games.
  - [ ] **课程** — default nav bar is GONE. Capsule visible. Category `<van-tabs>` are the first interactive element below the reserved top padding.
  - [ ] **排行榜** — default nav bar is GONE. Capsule visible. Period tabs (`今日/本周/本月`) and then type tabs (`经验值/游戏时长`) render below the padded top.
  - [ ] **我的** — default nav bar is GONE. Capsule visible. New top-bar shows two right-aligned icons (moon/sun on left, bell on right) below the capsule. Tapping the moon/sun icon swaps it (moon ↔ sun) AND recolors the whole app including the bottom tab bar. Tapping the bell navigates to /pages/me/notices/notices. The 公告通知 row in the cell group is gone; the remaining cells are: 我的团队, 推荐有礼, 兑换码, 购买会员, (and the separate 退出登录 group).

- [ ] **Step 3: Theme persistence check**

Toggle dark mode on the 我的 page. Navigate to 首页, 课程, 排行榜 via the tab bar. Each page should render in dark mode. Restart the mini program in DevTools (点击刷新预览). The app should come up in dark mode (from `dx_theme` in storage via `app.globalData.theme`).

- [ ] **Step 4: Real-device smoke test**

Scan the 预览 QR with 小程序助手 (per project memory — 真机调试 is broken on current DevTools). Repeat Step 2's visual checks on a real phone, paying attention to capsule-icon collision on any iPhone with a notch.

- [ ] **Step 5: Confirm completion**

If every page renders correctly, every tap handler works, and theme toggle still propagates everywhere, proceed to Task 7. If anything looks off, stop and report the specific page + issue — do not commit broken work.

---

## Task 7: Commit and merge

**Files:** stage + commit only

- [ ] **Step 1: Review staged state**

Run: `cd dx-mini && git status --short`
Expected output: 16 modified files across `pages/{home,games,me,leaderboard}/{*.ts,*.wxml,*.wxss,*.json}`.

- [ ] **Step 2: Stage exactly those files**

Run:
```bash
cd dx-mini && git add \
  miniprogram/pages/home/home.ts miniprogram/pages/home/home.wxml miniprogram/pages/home/home.wxss miniprogram/pages/home/home.json \
  miniprogram/pages/games/games.ts miniprogram/pages/games/games.wxml miniprogram/pages/games/games.wxss miniprogram/pages/games/games.json \
  miniprogram/pages/me/me.ts miniprogram/pages/me/me.wxml miniprogram/pages/me/me.wxss miniprogram/pages/me/me.json \
  miniprogram/pages/leaderboard/leaderboard.ts miniprogram/pages/leaderboard/leaderboard.wxml miniprogram/pages/leaderboard/leaderboard.wxss miniprogram/pages/leaderboard/leaderboard.json
```

- [ ] **Step 3: Commit**

Run:
```bash
cd dx-mini && git commit -m "$(cat <<'EOF'
feat(mini): move header actions to me page + custom nav on tab pages

Relocates the dark-mode toggle and notifications bell from home's top
bar to a new top-right action row on the me page, and drops the now-
redundant 公告通知 van-cell entry. Home's search bar fills the row.

Removes the default WeChat navigation bar on 首页, 课程, 排行榜, and
我的 by switching each to navigationStyle: custom. Each tab page
computes statusBarHeight via wx.getSystemInfoSync() in onLoad, exposes
it as a CSS custom property on .page-container, and reserves
calc(var(--status-bar-height) + 88rpx) of top padding so content clears
the WeChat capsule button without any device-specific math.
EOF
)"
```

Expected: commit created, working tree clean.

- [ ] **Step 4: Merge to main locally**

Run:
```bash
git checkout main && git merge --no-ff feat/mini-tab-page-chrome -m "$(cat <<'EOF'
Merge branch 'feat/mini-tab-page-chrome'

- Moves theme toggle + notifications bell from home to me page; drops
  the redundant 公告通知 van-cell.
- Removes default nav bar on all four tab pages (首页/课程/排行榜/我的)
  via navigationStyle: custom + shared status-bar padding idiom.
EOF
)" && git branch -d feat/mini-tab-page-chrome && git log --oneline -5
```

Expected: merge commit created, feature branch deleted, `git log` shows the merge commit at HEAD.

---

## Rollback plan

The work lands as one feature commit on `feat/mini-tab-page-chrome` plus a merge commit on main. To undo:

- Before merge: `git checkout main && git branch -D feat/mini-tab-page-chrome` discards the whole branch.
- After merge: `git revert -m 1 <merge-commit>` reverts the merge as a single commit on main. The individual feature commit and the spec commit can be inspected but don't need to be touched.

If a specific tab page renders poorly after merge (e.g., capsule collision on an unusual device), the narrowest fix is to revert only that page's `.json` back to `"navigationBarTitleText": "…"` — the `.ts`/`.wxml`/`.wxss` edits are harmless on their own and leaving them keeps the diff small.
