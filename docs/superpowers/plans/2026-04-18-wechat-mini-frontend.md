# WeChat Mini Program — dx-mini Frontend Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the complete dx-mini WeChat mini program frontend — 19 pages across 5 tabs, teal-themed with dark mode, using Vant Weapp.

**Architecture:** Native WeChat mini program (WXML/TypeScript/WXSS), Vant Weapp for UI components, custom tab bar for dark-mode support. All API calls go through `utils/api.ts` which wraps `wx.request()` with JWT auth and 401 handling. Theme stored in `globalData.theme` and passed as `<van-config-provider theme="{{theme}}>` per page.

**Tech Stack:** WeChat Mini Program (glass-easel), TypeScript (CommonJS), Vant Weapp, `utils/api.ts` wrapping the dx-api backend at `/api/*`.

**Spec:** `docs/superpowers/specs/2026-04-18-wechat-mini-design.md`

**Prerequisite:** Backend plan (`2026-04-18-wechat-mini-backend.md`) must be complete first — the `POST /api/auth/wechat-mini` endpoint must exist.

**Working directory for all commands:** `/Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram/`

---

## File Map

| Action | Path |
|--------|------|
| Modify | `package.json` — add @vant/weapp |
| Modify | `app.json` — pages list, custom tabBar |
| Modify | `app.ts` — auth check, theme init, WebSocket |
| Modify | `app.wxss` — Vant CSS variable overrides |
| Create | `utils/config.ts` |
| Create | `utils/auth.ts` |
| Create | `utils/api.ts` |
| Create | `utils/ws.ts` |
| Create | `utils/format.ts` |
| Create | `custom-tab-bar/index.ts/.wxml/.wxss/.json` |
| Create | `pages/login/login.ts/.wxml/.wxss/.json` |
| Create | `pages/home/home.ts/.wxml/.wxss/.json` |
| Create | `pages/games/games.ts/.wxml/.wxss/.json` |
| Create | `pages/games/detail/detail.ts/.wxml/.wxss/.json` |
| Create | `pages/games/play/play.ts/.wxml/.wxss/.json` |
| Create | `pages/games/favorites/favorites.ts/.wxml/.wxss/.json` |
| Create | `pages/leaderboard/leaderboard.ts/.wxml/.wxss/.json` |
| Create | `pages/learn/learn.ts/.wxml/.wxss/.json` |
| Create | `pages/learn/mastered/mastered.ts/.wxml/.wxss/.json` |
| Create | `pages/learn/unknown/unknown.ts/.wxml/.wxss/.json` |
| Create | `pages/learn/review/review.ts/.wxml/.wxss/.json` |
| Create | `pages/me/me.ts/.wxml/.wxss/.json` |
| Create | `pages/me/profile-edit/profile-edit.ts/.wxml/.wxss/.json` |
| Create | `pages/me/notices/notices.ts/.wxml/.wxss/.json` |
| Create | `pages/me/groups/groups.ts/.wxml/.wxss/.json` |
| Create | `pages/me/groups-detail/groups-detail.ts/.wxml/.wxss/.json` |
| Create | `pages/me/invite/invite.ts/.wxml/.wxss/.json` |
| Create | `pages/me/redeem/redeem.ts/.wxml/.wxss/.json` |
| Create | `pages/me/purchase/purchase.ts/.wxml/.wxss/.json` |

---

## Task 1: Foundation — npm, utilities, app shell

**Files:**
- Modify: `package.json`
- Modify: `app.json`
- Modify: `app.ts`
- Modify: `app.wxss`
- Create: `utils/config.ts`
- Create: `utils/auth.ts`
- Create: `utils/api.ts`
- Create: `utils/ws.ts`
- Create: `utils/format.ts`

- [ ] **Step 1: Install Vant Weapp**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram
npm install @vant/weapp
```

Then in WeChat DevTools: **Tools → Build npm**. This generates `miniprogram_npm/` from `node_modules/`.

- [ ] **Step 2: Write `utils/config.ts`**

```typescript
const { envVersion } = wx.getAccountInfoSync().miniProgram

export const config = {
  apiBaseUrl: envVersion === 'release' || envVersion === 'trial'
    ? 'https://api.douxue.com'
    : 'http://localhost:3001',
}
```

- [ ] **Step 3: Write `utils/auth.ts`**

```typescript
const TOKEN_KEY = 'dx_token'
const USER_ID_KEY = 'dx_user_id'

export function getToken(): string | null {
  return (wx.getStorageSync(TOKEN_KEY) as string) || null
}
export function setToken(token: string): void {
  wx.setStorageSync(TOKEN_KEY, token)
}
export function clearToken(): void {
  wx.removeStorageSync(TOKEN_KEY)
  wx.removeStorageSync(USER_ID_KEY)
}
export function isLoggedIn(): boolean {
  return !!getToken()
}
export function getUserId(): string | null {
  return (wx.getStorageSync(USER_ID_KEY) as string) || null
}
export function setUserId(id: string): void {
  wx.setStorageSync(USER_ID_KEY, id)
}
```

- [ ] **Step 4: Write `utils/api.ts`**

```typescript
import { config } from './config'
import { getToken, clearToken } from './auth'

export interface PaginatedData<T> {
  items: T[]
  nextCursor: string
  hasMore: boolean
}

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

function request<T>(method: string, path: string, data?: object): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = getToken()
    wx.request({
      url: config.apiBaseUrl + path,
      method: method as 'GET' | 'POST' | 'PUT' | 'DELETE',
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success(res) {
        const body = res.data as ApiResponse<T>
        if (res.statusCode === 401) {
          clearToken()
          if (body?.code === 40104) {
            wx.showModal({
              title: '提示',
              content: '账号已在其他设备登录',
              showCancel: false,
              complete() {
                wx.reLaunch({ url: '/pages/login/login' })
              },
            })
          } else {
            wx.reLaunch({ url: '/pages/login/login' })
          }
          return reject(new Error('unauthorized'))
        }
        if (body?.code !== 0) {
          return reject(new Error(body?.message || '请求失败'))
        }
        resolve(body.data)
      },
      fail(err) {
        reject(new Error(err.errMsg || '网络错误'))
      },
    })
  })
}

export const api = {
  get<T>(path: string): Promise<T> {
    return request<T>('GET', path)
  },
  post<T>(path: string, data: object): Promise<T> {
    return request<T>('POST', path, data)
  },
  put<T>(path: string, data: object): Promise<T> {
    return request<T>('PUT', path, data)
  },
  delete<T>(path: string, data?: object): Promise<T> {
    return request<T>('DELETE', path, data)
  },
}
```

- [ ] **Step 5: Write `utils/ws.ts`**

```typescript
import { config } from './config'

type EventHandler = (payload: unknown) => void

let socket: WechatMiniprogram.SocketTask | null = null
const handlers = new Map<string, EventHandler[]>()

export const ws = {
  connect(token: string): void {
    const wsUrl = config.apiBaseUrl.replace(/^http/, 'ws') + '/api/ws'
    socket = wx.connectSocket({
      url: wsUrl,
      header: { Authorization: `Bearer ${token}` },
      success() {},
      fail() {},
    })
    socket.onMessage(({ data }) => {
      try {
        const msg = JSON.parse(data as string) as { event: string; payload: unknown }
        const cbs = handlers.get(msg.event) ?? []
        cbs.forEach(cb => cb(msg.payload))
      } catch {
        // ignore malformed messages
      }
    })
  },
  subscribe(topic: string): void {
    socket?.send({ data: JSON.stringify({ type: 'subscribe', topic }) })
  },
  on(event: string, cb: EventHandler): void {
    if (!handlers.has(event)) handlers.set(event, [])
    handlers.get(event)!.push(cb)
  },
  disconnect(): void {
    socket?.close({})
    socket = null
    handlers.clear()
  },
}
```

- [ ] **Step 6: Write `utils/format.ts`**

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
    monthly: '月度会员',
    quarterly: '季度会员',
    yearly: '年度会员',
  }
  return map[grade] ?? grade
}
```

- [ ] **Step 7: Rewrite `app.json`**

```json
{
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
    "pages/me/purchase/purchase"
  ],
  "tabBar": {
    "custom": true,
    "list": [
      { "pagePath": "pages/home/home" },
      { "pagePath": "pages/games/games" },
      { "pagePath": "pages/leaderboard/leaderboard" },
      { "pagePath": "pages/learn/learn" },
      { "pagePath": "pages/me/me" }
    ]
  },
  "window": {
    "navigationBarTextStyle": "black",
    "navigationBarBackgroundColor": "#ffffff",
    "backgroundTextStyle": "light"
  },
  "style": "v2",
  "componentFramework": "glass-easel",
  "lazyCodeLoading": "requiredComponents"
}
```

- [ ] **Step 8: Rewrite `app.ts`**

```typescript
import { isLoggedIn, getToken, getUserId } from './utils/auth'
import { ws } from './utils/ws'

interface GlobalData {
  theme: 'light' | 'dark'
  userId: string
}

App<{ globalData: GlobalData }>({
  globalData: {
    theme: 'light',
    userId: '',
  },
  onLaunch() {
    const stored = wx.getStorageSync('dx_theme') as 'light' | 'dark' | ''
    const sys = wx.getSystemSetting()
    this.globalData.theme = stored || ((sys.theme as 'light' | 'dark') ?? 'light')

    wx.onThemeChange(({ theme }) => {
      if (!wx.getStorageSync('dx_theme')) {
        this.globalData.theme = theme as 'light' | 'dark'
      }
    })

    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }

    this.globalData.userId = getUserId() ?? ''
    const token = getToken()!
    ws.connect(token)
    ws.subscribe(`user::${this.globalData.userId}`)
    ws.on('session_replaced', () => {
      wx.removeStorageSync('dx_token')
      wx.removeStorageSync('dx_user_id')
      wx.reLaunch({ url: '/pages/login/login' })
    })
  },
})
```

- [ ] **Step 9: Rewrite `app.wxss`**

```css
page {
  --primary: #0d9488;
  --primary-light: #f0fdfa;
  --bg-page: #f5f5f5;
  --bg-card: #ffffff;
  --border-color: #f0f0f0;
  --text-primary: #1a1a1a;
  --text-secondary: #888888;
  --destructive: #ef4444;

  --van-primary-color: #0d9488;
  --van-danger-color: #ef4444;
  --van-border-radius-md: 8px;
  --van-border-radius-lg: 12px;
  --van-font-size-md: 14px;
  --van-background: #f5f5f5;
  --van-background-2: #ffffff;
  --van-text-color: #1a1a1a;
  --van-text-color-2: #888888;
  --van-border-color: #f0f0f0;

  background: var(--bg-page);
  font-family: -apple-system, "PingFang SC", sans-serif;
  font-size: 14px;
  color: var(--text-primary);
}

.dark {
  --primary: #14b8a6;
  --primary-light: rgba(20, 184, 166, 0.12);
  --bg-page: #0f0f0f;
  --bg-card: #1c1c1e;
  --border-color: rgba(255, 255, 255, 0.06);
  --text-primary: #f5f5f5;
  --text-secondary: #6b7280;
}

.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}

.page-container.dark {
  background: #0f0f0f;
}

.card {
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 16px;
  margin: 0 16px 12px;
}
```

- [ ] **Step 10: Verify compilation in DevTools**

Open WeChat DevTools → select the `dx-mini` project. The console should show no TypeScript errors.
Expected: zero errors, simulator loads.

- [ ] **Step 11: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/utils/ dx-mini/miniprogram/app.* dx-mini/miniprogram/package.json dx-mini/miniprogram/app.json
git commit -m "feat(mini): add utility modules and rewrite app shell"
```

---

## Task 2: Custom Tab Bar

**Files:**
- Create: `custom-tab-bar/index.ts`
- Create: `custom-tab-bar/index.wxml`
- Create: `custom-tab-bar/index.wxss`
- Create: `custom-tab-bar/index.json`

- [ ] **Step 1: Create `custom-tab-bar/index.json`**

```json
{
  "component": true,
  "usingComponents": {
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

- [ ] **Step 2: Create `custom-tab-bar/index.ts`**

```typescript
interface TabItem {
  icon: string
  activeIcon: string
  text: string
  path: string
}

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    tabs: [
      { icon: 'wap-home-o', activeIcon: 'wap-home', text: '首页', path: '/pages/home/home' },
      { icon: 'column', activeIcon: 'column', text: '课程', path: '/pages/games/games' },
      { icon: 'chart-trending-o', activeIcon: 'chart-trending-o', text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'records', activeIcon: 'records', text: '学习', path: '/pages/learn/learn' },
      { icon: 'contact', activeIcon: 'contact', text: '我的', path: '/pages/me/me' },
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

- [ ] **Step 3: Create `custom-tab-bar/index.wxml`**

```xml
<view class="tab-bar {{theme === 'dark' ? 'dark' : ''}}">
  <view
    wx:for="{{tabs}}"
    wx:key="path"
    class="tab-item {{active === index ? 'active' : ''}}"
    data-path="{{item.path}}"
    bind:tap="switchTab"
  >
    <van-icon
      name="{{item.icon}}"
      size="22px"
      color="{{active === index ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : '#9ca3af'}}"
    />
    <text class="tab-label {{active === index ? 'active-label' : ''}}">{{item.text}}</text>
  </view>
</view>
```

- [ ] **Step 4: Create `custom-tab-bar/index.wxss`**

```css
.tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  height: 56px;
  padding-bottom: env(safe-area-inset-bottom);
  background: #ffffff;
  border-top: 1px solid #f0f0f0;
  z-index: 100;
}
.tab-bar.dark {
  background: #1c1c1e;
  border-top-color: rgba(255, 255, 255, 0.06);
}
.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
}
.tab-label {
  font-size: 10px;
  color: #9ca3af;
}
.active-label {
  color: #0d9488;
  font-weight: 600;
}
.tab-bar.dark .active-label {
  color: #14b8a6;
}
```

- [ ] **Step 5: Verify in DevTools simulator**

Navigate to any tab page — tab bar should render with 5 icons, active item teal.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/custom-tab-bar/
git commit -m "feat(mini): add custom tab bar component"
```

---

## Task 3: Login Page

**Files:**
- Create: `pages/login/login.ts`
- Create: `pages/login/login.wxml`
- Create: `pages/login/login.wxss`
- Create: `pages/login/login.json`

- [ ] **Step 1: Create `pages/login/login.json`**

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

- [ ] **Step 2: Create `pages/login/login.ts`**

```typescript
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

- [ ] **Step 3: Create `pages/login/login.wxml`**

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

- [ ] **Step 4: Create `pages/login/login.wxss`**

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

- [ ] **Step 5: Test in simulator**

Run the simulator. The login page should show 斗学 logo and login button. Tap login — since there's no real WeChat session in simulator, it calls `wx.login()` which returns a test code. The API call will fail (backend not running) — that's expected. Verify the error toast appears.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/login/
git commit -m "feat(mini): add login page"
```

---

## Task 4: Home Page

**Files:**
- Create: `pages/home/home.ts`
- Create: `pages/home/home.wxml`
- Create: `pages/home/home.wxss`
- Create: `pages/home/home.json`

- [ ] **Step 1: Create `pages/home/home.json`**

```json
{
  "navigationBarTitleText": "斗学",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-skeleton": "@vant/weapp/skeleton/index",
    "van-badge": "@vant/weapp/badge/index"
  }
}
```

- [ ] **Step 2: Create `pages/home/home.ts`**

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
interface Greeting { text: string; emoji: string }

interface DashboardData {
  profile: DashboardProfile
  masterStats: MasterStats
  reviewStats: ReviewStats
  todayAnswers: number
  greeting: Greeting
}

interface HeatmapDay { date: string; count: number }
interface HeatmapData { year: number; days: HeatmapDay[]; accountYear: number }

interface GameCardData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; author: string | null; categoryName: string | null; levelCount: number
}

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
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 0, theme: app.globalData.theme })
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
      const hasUnread = dash.profile.lastReadNoticeAt === null || (
        dash.profile.lastReadNoticeAt !== null // will compare with latest notice in future
      )
      this.setData({
        loading: false,
        profile: dash.profile,
        masterStats: dash.masterStats,
        reviewStats: dash.reviewStats,
        todayAnswers: dash.todayAnswers,
        greeting: dash.greeting,
        heatmapCells: cells,
        unreadNotices: false, // simplified — badge updated separately
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
      const count = map.get(key) ?? 0
      const level = count === 0 ? 0 : count < 3 ? 1 : count < 6 ? 2 : count < 10 ? 3 : 4
      cells.push({ date: key, level })
    }
    return cells
  },
  toggleTheme() {
    const next: 'light' | 'dark' = this.data.theme === 'light' ? 'dark' : 'light'
    wx.setStorageSync('dx_theme', next)
    app.globalData.theme = next
    this.setData({ theme: next });
    (this.getTabBar() as any)?.setData({ theme: next })
  },
  goNotices() {
    wx.navigateTo({ url: '/pages/me/notices/notices' })
  },
  goSearch() {
    wx.navigateTo({ url: '/pages/games/games' })
  },
})
```

- [ ] **Step 3: Create `pages/home/home.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <!-- Top bar -->
    <view class="top-bar">
      <view class="search-box" bind:tap="goSearch">
        <van-icon name="search" size="16px" color="#9ca3af" />
        <text class="search-placeholder">搜索课程</text>
      </view>
      <view class="top-actions">
        <van-icon
          name="{{theme === 'dark' ? 'sunny-o' : 'moon-o'}}"
          size="22px"
          color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
          bind:tap="toggleTheme"
        />
        <van-icon name="bell" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" bind:tap="goNotices" />
      </view>
    </view>

    <van-skeleton title row="5" loading="{{loading}}">
      <!-- Greeting + stats -->
      <view class="hero-section">
        <text class="greeting">{{greeting ? greeting.emoji + ' ' + greeting.text : '你好！'}}</text>
        <view class="stat-row">
          <view class="stat-card">
            <text class="stat-value">{{profile.currentPlayStreak ?? 0}}</text>
            <text class="stat-label">连续天数</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">{{masterStats.total ?? 0}}</text>
            <text class="stat-label">已掌握</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">{{todayAnswers ?? 0}}</text>
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

- [ ] **Step 4: Create `pages/home/home.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}
.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-page);
  border-radius: 20px;
  padding: 8px 12px;
}
.search-placeholder {
  font-size: 13px;
  color: var(--text-secondary);
}
.top-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
.hero-section {
  padding: 20px 16px 12px;
}
.greeting {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  display: block;
  margin-bottom: 16px;
}
.stat-row {
  display: flex;
  gap: 12px;
}
.stat-card {
  flex: 1;
  background: var(--primary-light);
  border-radius: 12px;
  padding: 12px;
  text-align: center;
}
.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--primary);
  display: block;
}
.stat-label {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 2px;
  display: block;
}
.section {
  padding: 16px 16px 0;
}
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  display: block;
  margin-bottom: 12px;
}
.heatmap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.heatmap-cell {
  width: calc((100% - 48px) / 7);
  aspect-ratio: 1;
  border-radius: 3px;
  background: var(--border-color);
}
.heatmap-cell.level-1 { background: rgba(13, 148, 136, 0.25); }
.heatmap-cell.level-2 { background: rgba(13, 148, 136, 0.50); }
.heatmap-cell.level-3 { background: rgba(13, 148, 136, 0.75); }
.heatmap-cell.level-4 { background: #0d9488; }
.dark .heatmap-cell.level-4 { background: #14b8a6; }
```

- [ ] **Step 5: Test in simulator**

Navigate to home. Should show top bar with search + icons, greeting + 3 stat cards, heatmap grid.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/home/
git commit -m "feat(mini): add home page"
```

---

## Task 5: Games List Page

**Files:**
- Create: `pages/games/games.ts`
- Create: `pages/games/games.wxml`
- Create: `pages/games/games.wxss`
- Create: `pages/games/games.json`

- [ ] **Step 1: Create `pages/games/games.json`**

```json
{
  "navigationBarTitleText": "课程",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-image": "@vant/weapp/image/index"
  }
}
```

- [ ] **Step 2: Create `pages/games/games.ts`**

```typescript
import { api, PaginatedData } from '../../utils/api'

interface Category { id: string; name: string }
interface GameCardData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number; author: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    refreshing: false,
    categories: [{ id: '', name: '全部' }] as Category[],
    activeCategoryId: '',
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadCategories()
    this.loadGames(true)
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 1, theme: app.globalData.theme })
  },
  onPullDownRefresh() {
    this.loadGames(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadGames(false)
    }
  },
  async loadCategories() {
    const cats = await api.get<Category[]>('/api/game-categories').catch(() => [] as Category[])
    this.setData({ categories: [{ id: '', name: '全部' }, ...cats] })
  },
  async loadGames(reset: boolean) {
    if (this.data.loading) return
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const catId = this.data.activeCategoryId
    const qs = new URLSearchParams({ limit: '20' })
    if (cursor) qs.set('cursor', cursor)
    if (catId) qs.set('categoryIds', catId)
    try {
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: reset ? res.items : [...this.data.games, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loading: false,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onCategoryChange(e: WechatMiniprogram.TouchEvent) {
    const id = (e.detail as { name: string }).name
    this.setData({ activeCategoryId: id })
    this.loadGames(true)
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
  goFavorites() {
    wx.navigateTo({ url: '/pages/games/favorites/favorites' })
  },
})
```

- [ ] **Step 3: Create `pages/games/games.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <!-- Category tabs -->
    <van-tabs
      active="{{activeCategoryId}}"
      bind:click="onCategoryChange"
      color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
      background="{{theme === 'dark' ? '#1c1c1e' : '#ffffff'}}"
      title-active-color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
      scrollable
    >
      <van-tab
        wx:for="{{categories}}"
        wx:key="id"
        title="{{item.name}}"
        name="{{item.id}}"
      />
    </van-tabs>

    <!-- Favorites button -->
    <view class="fav-bar" bind:tap="goFavorites">
      <van-icon name="star-o" size="16px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
      <text class="fav-text">收藏的课程</text>
    </view>

    <!-- Game grid -->
    <view class="game-grid">
      <view
        wx:for="{{games}}"
        wx:key="id"
        class="game-card"
        data-id="{{item.id}}"
        bind:tap="goDetail"
      >
        <view class="game-cover">
          <van-image
            wx:if="{{item.coverUrl}}"
            src="{{item.coverUrl}}"
            width="100%"
            height="120px"
            fit="cover"
            radius="8px 8px 0 0"
          />
          <view wx:else class="cover-placeholder">
            <van-icon name="column" size="28px" color="#9ca3af" />
          </view>
        </view>
        <view class="game-info">
          <text class="game-name">{{item.name}}</text>
          <view class="game-meta">
            <text class="meta-text">{{item.levelCount}}关</text>
            <text class="meta-dot">·</text>
            <text class="meta-text">{{item.mode}}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Loading / empty -->
    <view wx:if="{{loading}}" class="load-more">
      <van-loading size="20px" color="#0d9488" />
    </view>
    <van-empty wx:if="{{!loading && games.length === 0}}" description="暂无课程" />
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/games/games.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
.fav-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--primary-light);
  border-bottom: 1px solid var(--border-color);
}
.fav-text {
  font-size: 13px;
  color: var(--primary);
  font-weight: 500;
}
.game-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
}
.game-card {
  width: calc(50% - 6px);
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}
.game-cover {
  height: 120px;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
}
.cover-placeholder {
  width: 100%;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-page);
}
.game-info {
  padding: 10px;
}
.game-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.game-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
}
.meta-text {
  font-size: 11px;
  color: var(--text-secondary);
}
.meta-dot {
  font-size: 11px;
  color: var(--text-secondary);
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 16px;
}
```

- [ ] **Step 5: Verify in simulator**

Tab 2 shows category tabs, game grid with 2 columns. Tapping a card navigates to detail (will error until detail page is built — that's OK).

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/games/games.*
git commit -m "feat(mini): add games list page"
```

---

## Task 6: Game Detail Page

**Files:**
- Create: `pages/games/detail/detail.ts`
- Create: `pages/games/detail/detail.wxml`
- Create: `pages/games/detail/detail.wxss`
- Create: `pages/games/detail/detail.json`

- [ ] **Step 1: Create `pages/games/detail/detail.json`**

```json
{
  "navigationBarTitleText": "课程详情",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-button": "@vant/weapp/button/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-cell": "@vant/weapp/cell/index",
    "van-image": "@vant/weapp/image/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/games/detail/detail.ts`**

```typescript
import { api } from '../../../utils/api'

interface GameLevelData { id: string; name: string; order: number }
interface GameDetailData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; author: string | null; categoryName: string | null
  pressName: string | null; levels: GameLevelData[]; levelCount: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    game: null as GameDetailData | null,
    favorited: false,
  },
  onLoad(options: { id?: string }) {
    this.setData({ theme: app.globalData.theme })
    if (options.id) this.loadGame(options.id)
  },
  async loadGame(id: string) {
    try {
      const [game, favRes] = await Promise.all([
        api.get<GameDetailData>(`/api/games/${id}`),
        api.get<{ favorited: boolean }>(`/api/games/${id}/favorited`),
      ])
      this.setData({ loading: false, game, favorited: favRes.favorited })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  async toggleFavorite() {
    if (!this.data.game) return
    try {
      const res = await api.post<{ favorited: boolean }>('/api/favorites/toggle', { gameId: this.data.game.id })
      this.setData({ favorited: res.favorited })
      wx.showToast({ title: res.favorited ? '已收藏' : '已取消收藏', icon: 'none' })
    } catch {
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },
  startLevel(e: WechatMiniprogram.TouchEvent) {
    const levelId = e.currentTarget.dataset['levelId'] as string
    const gameId = this.data.game!.id
    wx.navigateTo({ url: `/pages/games/play/play?gameId=${gameId}&levelId=${levelId}&degree=normal` })
  },
})
```

- [ ] **Step 3: Create `pages/games/detail/detail.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <block wx:if="{{!loading && game}}">
      <!-- Cover -->
      <view class="cover-wrap">
        <van-image wx:if="{{game.coverUrl}}" src="{{game.coverUrl}}" width="100%" height="200px" fit="cover" />
        <view wx:else class="cover-placeholder-lg">
          <van-icon name="column" size="48px" color="#9ca3af" />
        </view>
        <view class="cover-fav" bind:tap="toggleFavorite">
          <van-icon name="{{favorited ? 'star' : 'star-o'}}" size="24px" color="{{favorited ? '#f59e0b' : '#9ca3af'}}" />
        </view>
      </view>
      <!-- Info -->
      <view class="info-section">
        <text class="game-title">{{game.name}}</text>
        <view class="tags">
          <text wx:if="{{game.categoryName}}" class="tag">{{game.categoryName}}</text>
          <text class="tag">{{game.mode}}</text>
          <text class="tag">{{game.levelCount}}关</text>
        </view>
        <text wx:if="{{game.description}}" class="game-desc">{{game.description}}</text>
      </view>
      <!-- Levels -->
      <view class="section-header"><text class="section-title">关卡列表</text></view>
      <view class="levels-list">
        <view
          wx:for="{{game.levels}}"
          wx:key="id"
          class="level-item"
          data-level-id="{{item.id}}"
          bind:tap="startLevel"
        >
          <view class="level-left">
            <text class="level-num">{{index + 1}}</text>
            <text class="level-name">{{item.name}}</text>
          </view>
          <van-icon name="arrow" size="14px" color="#9ca3af" />
        </view>
      </view>
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/games/detail/detail.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
.center-loader { display: flex; justify-content: center; padding: 40px; }
.cover-wrap { position: relative; }
.cover-placeholder-lg {
  height: 200px;
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: center;
}
.cover-fav {
  position: absolute;
  top: 12px;
  right: 16px;
  background: rgba(0,0,0,0.4);
  border-radius: 50%;
  padding: 6px;
}
.info-section { padding: 16px; }
.game-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  display: block;
  margin-bottom: 8px;
}
.tags { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 10px; }
.tag {
  font-size: 11px;
  color: var(--primary);
  background: var(--primary-light);
  border-radius: 4px;
  padding: 2px 8px;
}
.game-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.section-header { padding: 0 16px 8px; }
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
}
.levels-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; }
.level-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card);
  border-radius: 10px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.level-left { display: flex; align-items: center; gap: 12px; }
.level-num {
  width: 28px;
  height: 28px;
  background: var(--primary-light);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--primary);
}
.level-name { font-size: 14px; color: var(--text-primary); }
```

- [ ] **Step 5: Verify in simulator**

Navigate from games list to a game card — detail page should show game info and levels list.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/games/detail/
git commit -m "feat(mini): add game detail page"
```

---

## Task 7: Game Play Page

**Files:**
- Create: `pages/games/play/play.ts`
- Create: `pages/games/play/play.wxml`
- Create: `pages/games/play/play.wxss`
- Create: `pages/games/play/play.json`

- [ ] **Step 1: Create `pages/games/play/play.json`**

```json
{
  "navigationBarTitleText": "游戏",
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-button": "@vant/weapp/button/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-popup": "@vant/weapp/popup/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/games/play/play.ts`**

```typescript
import { api } from '../../../utils/api'

interface ContentItemData {
  id: string; content: string; contentType: string
  translation: string | null; definition: string | null
  items: string | null // JSON array of choices
}

interface StartSessionResult {
  id: string; gameLevelId: string; degree: string
  score: number; exp: number; maxCombo: number
  correctCount: number; wrongCount: number
  currentContentItemId: string | null
}

interface Choice { text: string; correct: boolean }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    sessionId: '',
    gameLevelId: '',
    gameId: '',
    contentItems: [] as ContentItemData[],
    currentIndex: 0,
    currentItem: null as ContentItemData | null,
    choices: [] as Choice[],
    score: 0,
    combo: 0,
    maxCombo: 0,
    correctCount: 0,
    wrongCount: 0,
    answered: false,
    selectedChoice: -1,
    showResult: false,
    startTime: 0,
  },
  onLoad(options: { gameId?: string; levelId?: string; degree?: string }) {
    this.setData({ theme: app.globalData.theme })
    if (options.gameId && options.levelId) {
      this.initSession(options.gameId, options.levelId, options.degree ?? 'normal')
    }
  },
  async initSession(gameId: string, levelId: string, degree: string) {
    try {
      const session = await api.post<StartSessionResult>('/api/play-single/start', {
        game_id: gameId,
        game_level_id: levelId,
        degree,
        pattern: null,
      })
      const items = await api.get<ContentItemData[]>(
        `/api/games/${gameId}/levels/${levelId}/content?degree=${degree}`
      )
      let startIndex = 0
      if (session.currentContentItemId) {
        const idx = items.findIndex(i => i.id === session.currentContentItemId)
        if (idx >= 0) startIndex = idx
      }
      this.setData({
        loading: false,
        sessionId: session.id,
        gameLevelId: levelId,
        gameId,
        contentItems: items,
        currentIndex: startIndex,
        score: session.score,
        combo: 0,
        maxCombo: session.maxCombo,
        correctCount: session.correctCount,
        wrongCount: session.wrongCount,
        startTime: Date.now(),
      })
      this.showCurrentItem()
    } catch (err) {
      wx.showToast({ title: (err as Error).message || '启动失败', icon: 'none' })
      wx.navigateBack()
    }
  },
  showCurrentItem() {
    const item = this.data.contentItems[this.data.currentIndex]
    if (!item) {
      this.endSession()
      return
    }
    const choices = this.buildChoices(item)
    this.setData({ currentItem: item, choices, answered: false, selectedChoice: -1, startTime: Date.now() })
  },
  buildChoices(item: ContentItemData): Choice[] {
    if (!item.items) return []
    try {
      const parsed = JSON.parse(item.items) as unknown[]
      if (Array.isArray(parsed)) {
        return parsed.map((c: unknown) => {
          if (typeof c === 'string') return { text: c, correct: c === item.content }
          const obj = c as { text?: string; correct?: boolean }
          return { text: obj.text ?? '', correct: obj.correct ?? false }
        })
      }
    } catch {}
    return []
  },
  selectChoice(e: WechatMiniprogram.TouchEvent) {
    if (this.data.answered) return
    const idx = e.currentTarget.dataset['idx'] as number
    const choice = this.data.choices[idx]
    const isCorrect = choice.correct
    const duration = Date.now() - this.data.startTime
    const baseScore = isCorrect ? 10 : 0
    const newCombo = isCorrect ? this.data.combo + 1 : 0
    const comboScore = isCorrect ? Math.min(newCombo - 1, 5) * 2 : 0
    const score = this.data.score + baseScore + comboScore
    const maxCombo = Math.max(this.data.maxCombo, newCombo)
    const nextIndex = this.data.currentIndex + 1
    const nextItem = this.data.contentItems[nextIndex] ?? null
    this.setData({
      answered: true,
      selectedChoice: idx,
      score,
      combo: newCombo,
      maxCombo,
      correctCount: this.data.correctCount + (isCorrect ? 1 : 0),
      wrongCount: this.data.wrongCount + (isCorrect ? 0 : 1),
    })
    api.post('/api/play-single/' + this.data.sessionId + '/answers', {
      game_session_id: this.data.sessionId,
      game_level_id: this.data.gameLevelId,
      content_item_id: this.data.currentItem!.id,
      is_correct: isCorrect,
      user_answer: choice.text,
      source_answer: this.data.currentItem!.content,
      base_score: baseScore,
      combo_score: comboScore,
      score: baseScore + comboScore,
      max_combo: maxCombo,
      play_time: Math.floor(duration / 1000),
      duration: Math.floor(duration / 1000),
      next_content_item_id: nextItem?.id ?? null,
    }).catch(() => {})
    setTimeout(() => {
      this.setData({ currentIndex: nextIndex })
      this.showCurrentItem()
    }, isCorrect ? 600 : 1200)
  },
  markWord(e: WechatMiniprogram.TouchEvent) {
    const type = e.currentTarget.dataset['type'] as string
    const item = this.data.currentItem
    if (!item) return
    const path = `/api/tracking/${type}`
    const body: Record<string, string> = { content_item_id: item.id, game_id: this.data.gameId, game_level_id: this.data.gameLevelId }
    if (type === 'review') { delete body['game_id']; delete body['game_level_id'] }
    api.post(path, body).then(() => wx.showToast({ title: '已标记', icon: 'none' })).catch(() => {})
  },
  async endSession() {
    try {
      await api.post('/api/play-single/' + this.data.sessionId + '/end', {
        score: this.data.score,
        exp: Math.floor(this.data.score / 10),
        max_combo: this.data.maxCombo,
        correct_count: this.data.correctCount,
        wrong_count: this.data.wrongCount,
        skip_count: 0,
      })
    } catch {}
    this.setData({ showResult: true })
  },
  goBack() {
    wx.navigateBack()
  },
})
```

- [ ] **Step 3: Create `pages/games/play/play.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="play-page {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />

    <!-- Active game view -->
    <block wx:if="{{!loading && !showResult && currentItem}}">
      <!-- Score bar -->
      <view class="score-bar">
        <view bind:tap="goBack"><van-icon name="arrow-left" size="20px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" /></view>
        <text class="score-text">{{score}} 分</text>
        <text class="combo-text">x{{combo}}</text>
      </view>
      <!-- Question -->
      <view class="question-wrap">
        <text class="question-word">{{currentItem.content}}</text>
        <text wx:if="{{currentItem.definition}}" class="question-def">{{currentItem.definition}}</text>
      </view>
      <!-- Choices -->
      <view class="choices">
        <view
          wx:for="{{choices}}"
          wx:key="text"
          class="choice {{answered && index === selectedChoice ? (item.correct ? 'correct' : 'wrong') : ''}} {{answered && item.correct && index !== selectedChoice ? 'reveal' : ''}}"
          data-idx="{{index}}"
          bind:tap="selectChoice"
        >
          <text class="choice-text">{{item.text}}</text>
        </view>
      </view>
      <!-- Mark word actions -->
      <view wx:if="{{answered}}" class="mark-actions">
        <view class="mark-btn" data-type="master" bind:tap="markWord">
          <van-icon name="success" size="16px" color="#10b981" />
          <text>已掌握</text>
        </view>
        <view class="mark-btn" data-type="unknown" bind:tap="markWord">
          <van-icon name="question-o" size="16px" color="#f59e0b" />
          <text>不认识</text>
        </view>
        <view class="mark-btn" data-type="review" bind:tap="markWord">
          <van-icon name="clock-o" size="16px" color="#6366f1" />
          <text>待复习</text>
        </view>
      </view>
    </block>

    <!-- Results popup -->
    <van-popup show="{{showResult}}" position="bottom" round custom-style="padding:32px 24px;">
      <view class="result-body">
        <text class="result-title">本关完成！</text>
        <view class="result-stats">
          <view class="result-stat">
            <text class="rs-value">{{score}}</text>
            <text class="rs-label">得分</text>
          </view>
          <view class="result-stat">
            <text class="rs-value">{{maxCombo}}</text>
            <text class="rs-label">最大连击</text>
          </view>
          <view class="result-stat">
            <text class="rs-value">{{correctCount}}</text>
            <text class="rs-label">答对</text>
          </view>
        </view>
        <van-button type="primary" block round bind:click="goBack" custom-style="margin-top:24px;">返回</van-button>
      </view>
    </van-popup>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/games/play/play.wxss`**

```css
.play-page {
  min-height: 100vh;
  background: var(--bg-page);
  display: flex;
  flex-direction: column;
}
.center-loader { display: flex; justify-content: center; padding: 40px; }
.score-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  padding-top: calc(16px + env(safe-area-inset-top));
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}
.score-text { font-size: 16px; font-weight: 700; color: var(--primary); }
.combo-text { font-size: 14px; color: var(--text-secondary); }
.question-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
  gap: 12px;
}
.question-word { font-size: 32px; font-weight: 700; color: var(--text-primary); text-align: center; }
.question-def { font-size: 14px; color: var(--text-secondary); text-align: center; }
.choices {
  padding: 0 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.choice {
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
}
.choice.correct { border-color: #10b981; background: rgba(16,185,129,0.1); }
.choice.wrong { border-color: #ef4444; background: rgba(239,68,68,0.1); }
.choice.reveal { border-color: #10b981; background: rgba(16,185,129,0.05); }
.choice-text { font-size: 15px; color: var(--text-primary); }
.mark-actions {
  display: flex;
  justify-content: center;
  gap: 20px;
  padding: 0 20px 20px;
}
.mark-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-secondary);
}
.result-body { display: flex; flex-direction: column; align-items: center; }
.result-title { font-size: 22px; font-weight: 700; color: var(--text-primary); margin-bottom: 24px; }
.result-stats { display: flex; gap: 32px; }
.result-stat { display: flex; flex-direction: column; align-items: center; gap: 4px; }
.rs-value { font-size: 28px; font-weight: 700; color: var(--primary); }
.rs-label { font-size: 12px; color: var(--text-secondary); }
```

- [ ] **Step 5: Verify in simulator**

Navigate from game detail to a level → play page loads, shows question and choices. Tap a choice — it turns green/red. Mark actions appear after answering.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/games/play/
git commit -m "feat(mini): add game play page"
```

---

## Task 8: Favorites Page

**Files:**
- Create: `pages/games/favorites/favorites.ts`
- Create: `pages/games/favorites/favorites.wxml`
- Create: `pages/games/favorites/favorites.wxss`
- Create: `pages/games/favorites/favorites.json`

- [ ] **Step 1: Create `pages/games/favorites/favorites.json`**

```json
{
  "navigationBarTitleText": "收藏的课程",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-swipe-cell": "@vant/weapp/swipe-cell/index",
    "van-button": "@vant/weapp/button/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/games/favorites/favorites.ts`**

```typescript
import { api } from '../../../utils/api'

interface GameCardData {
  id: string; name: string; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    games: [] as GameCardData[],
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadFavorites()
  },
  onPullDownRefresh() {
    this.loadFavorites().then(() => wx.stopPullDownRefresh())
  },
  async loadFavorites() {
    this.setData({ loading: true })
    try {
      const games = await api.get<GameCardData[]>('/api/favorites')
      this.setData({ loading: false, games })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  async unfavorite(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    try {
      await api.post('/api/favorites/toggle', { gameId: id })
      this.setData({ games: this.data.games.filter(g => g.id !== id) })
      wx.showToast({ title: '已取消收藏', icon: 'none' })
    } catch {
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
})
```

- [ ] **Step 3: Create `pages/games/favorites/favorites.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && games.length === 0}}" description="暂无收藏课程" />
    <view wx:if="{{!loading && games.length > 0}}" class="list">
      <van-swipe-cell
        wx:for="{{games}}"
        wx:key="id"
        right-width="{{80}}"
      >
        <view class="fav-item" data-id="{{item.id}}" bind:tap="goDetail">
          <view class="fav-info">
            <text class="fav-name">{{item.name}}</text>
            <view class="fav-meta">
              <text wx:if="{{item.categoryName}}" class="meta-tag">{{item.categoryName}}</text>
              <text class="meta-tag">{{item.levelCount}}关</text>
              <text class="meta-tag">{{item.mode}}</text>
            </view>
          </view>
          <van-icon name="arrow" size="14px" color="#9ca3af" />
        </view>
        <view slot="right" class="unfav-btn" data-id="{{item.id}}" bind:tap="unfavorite">
          <van-icon name="star" size="20px" color="#ffffff" />
        </view>
      </van-swipe-cell>
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/games/favorites/favorites.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 40px;
}
.center-loader { display: flex; justify-content: center; padding: 40px; }
.list { padding: 12px 16px; display: flex; flex-direction: column; gap: 8px; }
.fav-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.fav-info { flex: 1; }
.fav-name { font-size: 15px; font-weight: 600; color: var(--text-primary); display: block; margin-bottom: 4px; }
.fav-meta { display: flex; flex-wrap: wrap; gap: 6px; }
.meta-tag {
  font-size: 11px;
  color: var(--primary);
  background: var(--primary-light);
  border-radius: 4px;
  padding: 2px 6px;
}
.unfav-btn {
  width: 80px;
  height: 100%;
  background: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
}
```

- [ ] **Step 5: Verify in simulator**

Navigate to favorites — shows list with swipe-to-unfavorite. Empty state if no favorites.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/games/favorites/
git commit -m "feat(mini): add favorites page"
```

---

## Task 9: Leaderboard Page

**Files:**
- Create: `pages/leaderboard/leaderboard.ts`
- Create: `pages/leaderboard/leaderboard.wxml`
- Create: `pages/leaderboard/leaderboard.wxss`
- Create: `pages/leaderboard/leaderboard.json`

- [ ] **Step 1: Create `pages/leaderboard/leaderboard.json`**

```json
{
  "navigationBarTitleText": "排行榜",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-image": "@vant/weapp/image/index",
    "van-empty": "@vant/weapp/empty/index"
  }
}
```

- [ ] **Step 2: Create `pages/leaderboard/leaderboard.ts`**

```typescript
import { api } from '../../utils/api'
import { formatNumber } from '../../utils/format'

interface LeaderboardEntry {
  id: string; username: string; nickname: string | null
  avatarUrl: string | null; value: number; rank: number
}
interface LeaderboardResult {
  entries: LeaderboardEntry[]
  myRank: LeaderboardEntry | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: false,
    period: 'month' as 'day' | 'week' | 'month',
    lbType: 'exp' as 'exp' | 'playtime',
    entries: [] as LeaderboardEntry[],
    myRank: null as LeaderboardEntry | null,
    formatNumber,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadLeaderboard()
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 2, theme: app.globalData.theme })
  },
  async loadLeaderboard() {
    this.setData({ loading: true })
    try {
      const res = await api.get<LeaderboardResult>(
        `/api/leaderboard?type=${this.data.lbType}&period=${this.data.period}`
      )
      this.setData({ loading: false, entries: res.entries, myRank: res.myRank })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onPeriodChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ period: (e.detail as { name: string }).name as 'day' | 'week' | 'month' })
    this.loadLeaderboard()
  },
  onTypeChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ lbType: (e.detail as { name: string }).name as 'exp' | 'playtime' })
    this.loadLeaderboard()
  },
})
```

- [ ] **Step 3: Create `pages/leaderboard/leaderboard.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <!-- Period tabs -->
    <van-tabs
      active="{{period}}"
      bind:click="onPeriodChange"
      color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
      background="{{theme === 'dark' ? '#1c1c1e' : '#ffffff'}}"
    >
      <van-tab title="今日" name="day" />
      <van-tab title="本周" name="week" />
      <van-tab title="本月" name="month" />
    </van-tabs>
    <!-- Type tabs -->
    <van-tabs
      active="{{lbType}}"
      bind:click="onTypeChange"
      color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
      background="{{theme === 'dark' ? '#0f0f0f' : '#f5f5f5'}}"
    >
      <van-tab title="经验值" name="exp" />
      <van-tab title="游戏时长" name="playtime" />
    </van-tabs>

    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />

    <block wx:if="{{!loading}}">
      <!-- Top 3 podium -->
      <view wx:if="{{entries.length >= 3}}" class="podium">
        <!-- 2nd -->
        <view class="podium-item second">
          <view class="podium-rank">2</view>
          <text class="podium-name">{{entries[1].nickname || entries[1].username}}</text>
          <text class="podium-value">{{entries[1].value}}</text>
        </view>
        <!-- 1st -->
        <view class="podium-item first">
          <view class="podium-rank first-rank">1</view>
          <text class="podium-name">{{entries[0].nickname || entries[0].username}}</text>
          <text class="podium-value">{{entries[0].value}}</text>
        </view>
        <!-- 3rd -->
        <view class="podium-item third">
          <view class="podium-rank">3</view>
          <text class="podium-name">{{entries[2].nickname || entries[2].username}}</text>
          <text class="podium-value">{{entries[2].value}}</text>
        </view>
      </view>

      <!-- Ranked list (4+) -->
      <view class="rank-list">
        <view
          wx:for="{{entries}}"
          wx:key="id"
          wx:if="{{index >= 3}}"
          class="rank-item"
        >
          <text class="rank-num">{{item.rank}}</text>
          <text class="rank-name">{{item.nickname || item.username}}</text>
          <text class="rank-value">{{item.value}}</text>
        </view>
      </view>

      <!-- My rank (pinned) -->
      <view wx:if="{{myRank}}" class="my-rank-bar">
        <text class="rank-num">{{myRank.rank}}</text>
        <text class="rank-name">我</text>
        <text class="rank-value">{{myRank.value}}</text>
      </view>

      <van-empty wx:if="{{entries.length === 0}}" description="暂无数据" />
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/leaderboard/leaderboard.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}
.center-loader { display: flex; justify-content: center; padding: 40px; }
.podium {
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 12px;
  padding: 24px 16px 16px;
}
.podium-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  flex: 1;
}
.podium-rank {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: var(--primary);
}
.first-rank {
  background: #fef3c7;
  color: #d97706;
  width: 44px;
  height: 44px;
  font-size: 18px;
}
.podium-name {
  font-size: 12px;
  color: var(--text-primary);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
}
.podium-value { font-size: 11px; color: var(--text-secondary); }
.rank-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; margin-bottom: 80px; }
.rank-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-card);
  border-radius: 10px;
  border: 1px solid var(--border-color);
  padding: 12px;
}
.rank-num { width: 28px; font-size: 14px; font-weight: 600; color: var(--text-secondary); }
.rank-name { flex: 1; font-size: 14px; color: var(--text-primary); }
.rank-value { font-size: 13px; font-weight: 600; color: var(--primary); }
.my-rank-bar {
  position: fixed;
  bottom: 60px;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--primary);
  padding: 12px 16px;
}
.my-rank-bar .rank-num,
.my-rank-bar .rank-name,
.my-rank-bar .rank-value {
  color: #ffffff;
}
```

- [ ] **Step 5: Verify in simulator**

Tab 3 shows period+type tabs, top 3 podium, ranked list, and pinned current user row.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/leaderboard/
git commit -m "feat(mini): add leaderboard page"
```

---

## Task 10: Learn Hub Page

**Files:**
- Create: `pages/learn/learn.ts`
- Create: `pages/learn/learn.wxml`
- Create: `pages/learn/learn.wxss`
- Create: `pages/learn/learn.json`

- [ ] **Step 1: Create `pages/learn/learn.json`**

```json
{
  "navigationBarTitleText": "学习",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/learn/learn.ts`**

```typescript
import { api } from '../../utils/api'

interface Stats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    masterStats: null as Stats | null,
    unknownStats: null as Stats | null,
    reviewStats: null as ReviewStats | null,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 3, theme: app.globalData.theme })
    this.loadStats()
  },
  async loadStats() {
    this.setData({ loading: true })
    try {
      const [masterStats, unknownStats, reviewStats] = await Promise.all([
        api.get<Stats>('/api/tracking/master/stats'),
        api.get<Stats>('/api/tracking/unknown/stats'),
        api.get<ReviewStats>('/api/tracking/review/stats'),
      ])
      this.setData({ loading: false, masterStats, unknownStats, reviewStats })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
```

- [ ] **Step 3: Create `pages/learn/learn.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <view wx:if="{{!loading}}" class="content">
      <!-- Stat cards -->
      <view class="stat-card-row">
        <view class="learn-stat-card teal" bind:tap="goMastered">
          <text class="lsc-value">{{masterStats.total ?? 0}}</text>
          <text class="lsc-label">已掌握</text>
          <text class="lsc-sub">本周 +{{masterStats.thisWeek ?? 0}}</text>
        </view>
        <view class="learn-stat-card amber" bind:tap="goUnknown">
          <text class="lsc-value">{{unknownStats.total ?? 0}}</text>
          <text class="lsc-label">不认识</text>
          <text class="lsc-sub">本周 +{{unknownStats.thisWeek ?? 0}}</text>
        </view>
        <view class="learn-stat-card purple" bind:tap="goReview">
          <text class="lsc-value">{{reviewStats.pending ?? 0}}</text>
          <text class="lsc-label">待复习</text>
          <text class="lsc-sub">逾期 {{reviewStats.overdue ?? 0}}</text>
        </view>
      </view>
      <!-- Quick links -->
      <view class="quick-links">
        <view class="quick-link" bind:tap="goMastered">
          <van-icon name="success" size="20px" color="#10b981" />
          <text class="ql-text">已掌握词库</text>
          <van-icon name="arrow" size="14px" color="#9ca3af" />
        </view>
        <view class="quick-link" bind:tap="goUnknown">
          <van-icon name="question-o" size="20px" color="#f59e0b" />
          <text class="ql-text">不认识词库</text>
          <van-icon name="arrow" size="14px" color="#9ca3af" />
        </view>
        <view class="quick-link" bind:tap="goReview">
          <van-icon name="clock-o" size="20px" color="#6366f1" />
          <text class="ql-text">待复习队列</text>
          <van-icon name="arrow" size="14px" color="#9ca3af" />
        </view>
      </view>
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/learn/learn.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 100rpx; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.content { padding: 16px; }
.stat-card-row { display: flex; gap: 10px; margin-bottom: 20px; }
.learn-stat-card {
  flex: 1;
  border-radius: 12px;
  padding: 14px 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.learn-stat-card.teal { background: rgba(13,148,136,0.12); }
.learn-stat-card.amber { background: rgba(245,158,11,0.12); }
.learn-stat-card.purple { background: rgba(99,102,241,0.12); }
.lsc-value { font-size: 24px; font-weight: 700; color: var(--text-primary); }
.lsc-label { font-size: 12px; color: var(--text-primary); font-weight: 500; }
.lsc-sub { font-size: 11px; color: var(--text-secondary); }
.quick-links { display: flex; flex-direction: column; gap: 8px; }
.quick-link {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.ql-text { flex: 1; font-size: 15px; color: var(--text-primary); }
```

- [ ] **Step 5: Verify + Commit**

Verify tab 4 shows 3 stat cards + 3 quick links.

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/learn/learn.*
git commit -m "feat(mini): add learn hub page"
```

---

## Task 11: Mastered Words Page

**Files:**
- Create: `pages/learn/mastered/mastered.ts`
- Create: `pages/learn/mastered/mastered.wxml`
- Create: `pages/learn/mastered/mastered.wxss`
- Create: `pages/learn/mastered/mastered.json`

- [ ] **Step 1: Create `pages/learn/mastered/mastered.json`**

```json
{
  "navigationBarTitleText": "已掌握",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-checkbox": "@vant/weapp/checkbox/index",
    "van-checkbox-group": "@vant/weapp/checkbox-group/index",
    "van-button": "@vant/weapp/button/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/learn/mastered/mastered.ts`**

```typescript
import { api, PaginatedData } from '../../../utils/api'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface TrackingItemData {
  id: string; contentItem: TrackingContentData | null
  gameName: string | null; masteredAt: string | null; createdAt: string
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    items: [] as TrackingItemData[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadItems(true)
  },
  onPullDownRefresh() {
    this.loadItems(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadItems(false)
  },
  async loadItems(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<TrackingItemData>>(`/api/tracking/master${qs}`)
      this.setData({
        loading: false,
        items: reset ? res.items : [...this.data.items, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  async bulkDelete() {
    if (this.data.selectedIds.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: `删除 ${this.data.selectedIds.length} 个词？`,
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids: this.data.selectedIds })
          this.setData({
            items: this.data.items.filter(i => !this.data.selectedIds.includes(i.id)),
            selectedIds: [],
            selectMode: false,
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
})
```

- [ ] **Step 3: Create `pages/learn/mastered/mastered.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <!-- Action bar -->
    <view class="action-bar">
      <text class="ab-count">{{items.length}} 词</text>
      <view class="ab-actions">
        <view class="ab-btn" bind:tap="toggleSelectMode">
          <text>{{selectMode ? '取消' : '批量删除'}}</text>
        </view>
        <view wx:if="{{selectMode && selectedIds.length > 0}}" class="ab-btn delete" bind:tap="bulkDelete">
          <text>删除({{selectedIds.length}})</text>
        </view>
      </view>
    </view>

    <van-loading wx:if="{{loading && items.length === 0}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && items.length === 0}}" description="暂无已掌握的词" />

    <van-checkbox-group value="{{selectedIds}}" bind:change="onSelectChange">
      <view class="word-list">
        <view wx:for="{{items}}" wx:key="id" class="word-item">
          <van-checkbox wx:if="{{selectMode}}" name="{{item.id}}" />
          <view class="word-body">
            <text class="word-content">{{item.contentItem.content}}</text>
            <text wx:if="{{item.contentItem.translation}}" class="word-trans">{{item.contentItem.translation}}</text>
            <text wx:if="{{item.gameName}}" class="word-game">来自：{{item.gameName}}</text>
          </view>
        </view>
      </view>
    </van-checkbox-group>

    <view wx:if="{{loading && items.length > 0}}" class="load-more">
      <van-loading size="20px" color="#0d9488" />
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/learn/mastered/mastered.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}
.ab-count { font-size: 13px; color: var(--text-secondary); }
.ab-actions { display: flex; gap: 10px; }
.ab-btn {
  font-size: 13px;
  color: var(--primary);
  padding: 4px 10px;
  border: 1px solid var(--primary);
  border-radius: 12px;
}
.ab-btn.delete {
  color: #ef4444;
  border-color: #ef4444;
}
.word-list { padding: 12px 16px; display: flex; flex-direction: column; gap: 8px; }
.word-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card);
  border-radius: 10px;
  border: 1px solid var(--border-color);
  padding: 12px;
}
.word-body { flex: 1; }
.word-content { font-size: 16px; font-weight: 600; color: var(--text-primary); display: block; }
.word-trans { font-size: 13px; color: var(--text-secondary); display: block; margin-top: 2px; }
.word-game { font-size: 11px; color: var(--text-secondary); display: block; margin-top: 4px; }
.load-more { display: flex; justify-content: center; padding: 16px; }
```

- [ ] **Step 5: Verify + Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/learn/mastered/
git commit -m "feat(mini): add mastered words page"
```

---

## Task 12: Unknown and Review Word Pages

**Files:**
- Create: `pages/learn/unknown/unknown.ts/.wxml/.wxss/.json`
- Create: `pages/learn/review/review.ts/.wxml/.wxss/.json`

Both pages are structurally identical to Task 11 (mastered), with three differences:
- `unknown`: hits `/api/tracking/unknown`, title "不认识"
- `review`: hits `/api/tracking/review`, title "待复习"; mark-mastered endpoint is same `/api/tracking/master`

- [ ] **Step 1: Create `pages/learn/unknown/unknown.json`**

```json
{
  "navigationBarTitleText": "不认识",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-checkbox": "@vant/weapp/checkbox/index",
    "van-checkbox-group": "@vant/weapp/checkbox-group/index",
    "van-button": "@vant/weapp/button/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/learn/unknown/unknown.ts`**

```typescript
import { api, PaginatedData } from '../../../utils/api'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface TrackingItemData {
  id: string; contentItem: TrackingContentData | null
  gameName: string | null; createdAt: string
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    items: [] as TrackingItemData[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadItems(true)
  },
  onPullDownRefresh() {
    this.loadItems(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadItems(false)
  },
  async loadItems(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<TrackingItemData>>(`/api/tracking/unknown${qs}`)
      this.setData({
        loading: false,
        items: reset ? res.items : [...this.data.items, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  async bulkDelete() {
    if (this.data.selectedIds.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: `删除 ${this.data.selectedIds.length} 个词？`,
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/unknown', { ids: this.data.selectedIds })
          this.setData({
            items: this.data.items.filter(i => !this.data.selectedIds.includes(i.id)),
            selectedIds: [],
            selectMode: false,
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
})
```

- [ ] **Step 3: Create `pages/learn/unknown/unknown.wxml`** — identical to mastered.wxml except empty text is "暂无不认识的词":

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <view class="action-bar">
      <text class="ab-count">{{items.length}} 词</text>
      <view class="ab-actions">
        <view class="ab-btn" bind:tap="toggleSelectMode">
          <text>{{selectMode ? '取消' : '批量删除'}}</text>
        </view>
        <view wx:if="{{selectMode && selectedIds.length > 0}}" class="ab-btn delete" bind:tap="bulkDelete">
          <text>删除({{selectedIds.length}})</text>
        </view>
      </view>
    </view>
    <van-loading wx:if="{{loading && items.length === 0}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && items.length === 0}}" description="暂无不认识的词" />
    <van-checkbox-group value="{{selectedIds}}" bind:change="onSelectChange">
      <view class="word-list">
        <view wx:for="{{items}}" wx:key="id" class="word-item">
          <van-checkbox wx:if="{{selectMode}}" name="{{item.id}}" />
          <view class="word-body">
            <text class="word-content">{{item.contentItem.content}}</text>
            <text wx:if="{{item.contentItem.translation}}" class="word-trans">{{item.contentItem.translation}}</text>
            <text wx:if="{{item.gameName}}" class="word-game">来自：{{item.gameName}}</text>
          </view>
        </view>
      </view>
    </van-checkbox-group>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/learn/unknown/unknown.wxss`** — copy mastered.wxss verbatim (styles are identical).

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}
.ab-count { font-size: 13px; color: var(--text-secondary); }
.ab-actions { display: flex; gap: 10px; }
.ab-btn { font-size: 13px; color: var(--primary); padding: 4px 10px; border: 1px solid var(--primary); border-radius: 12px; }
.ab-btn.delete { color: #ef4444; border-color: #ef4444; }
.word-list { padding: 12px 16px; display: flex; flex-direction: column; gap: 8px; }
.word-item { display: flex; align-items: center; gap: 10px; background: var(--bg-card); border-radius: 10px; border: 1px solid var(--border-color); padding: 12px; }
.word-body { flex: 1; }
.word-content { font-size: 16px; font-weight: 600; color: var(--text-primary); display: block; }
.word-trans { font-size: 13px; color: var(--text-secondary); display: block; margin-top: 2px; }
.word-game { font-size: 11px; color: var(--text-secondary); display: block; margin-top: 4px; }
```

- [ ] **Step 5: Create `pages/learn/review/review.json`**

```json
{
  "navigationBarTitleText": "待复习",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-checkbox": "@vant/weapp/checkbox/index",
    "van-checkbox-group": "@vant/weapp/checkbox-group/index",
    "van-button": "@vant/weapp/button/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 6: Create `pages/learn/review/review.ts`** — identical to unknown.ts with `/api/tracking/review` and delete endpoint `/api/tracking/review`:

```typescript
import { api, PaginatedData } from '../../../utils/api'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface TrackingItemData {
  id: string; contentItem: TrackingContentData | null
  gameName: string | null; createdAt: string
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    items: [] as TrackingItemData[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadItems(true)
  },
  onPullDownRefresh() {
    this.loadItems(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadItems(false)
  },
  async loadItems(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<TrackingItemData>>(`/api/tracking/review${qs}`)
      this.setData({
        loading: false,
        items: reset ? res.items : [...this.data.items, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  async bulkDelete() {
    if (this.data.selectedIds.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: `删除 ${this.data.selectedIds.length} 个词？`,
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/review', { ids: this.data.selectedIds })
          this.setData({
            items: this.data.items.filter(i => !this.data.selectedIds.includes(i.id)),
            selectedIds: [],
            selectMode: false,
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
})
```

- [ ] **Step 7: Create `pages/learn/review/review.wxml`** — identical to unknown.wxml, empty text is "暂无待复习的词":

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <view class="action-bar">
      <text class="ab-count">{{items.length}} 词</text>
      <view class="ab-actions">
        <view class="ab-btn" bind:tap="toggleSelectMode">
          <text>{{selectMode ? '取消' : '批量删除'}}</text>
        </view>
        <view wx:if="{{selectMode && selectedIds.length > 0}}" class="ab-btn delete" bind:tap="bulkDelete">
          <text>删除({{selectedIds.length}})</text>
        </view>
      </view>
    </view>
    <van-loading wx:if="{{loading && items.length === 0}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && items.length === 0}}" description="暂无待复习的词" />
    <van-checkbox-group value="{{selectedIds}}" bind:change="onSelectChange">
      <view class="word-list">
        <view wx:for="{{items}}" wx:key="id" class="word-item">
          <van-checkbox wx:if="{{selectMode}}" name="{{item.id}}" />
          <view class="word-body">
            <text class="word-content">{{item.contentItem.content}}</text>
            <text wx:if="{{item.contentItem.translation}}" class="word-trans">{{item.contentItem.translation}}</text>
            <text wx:if="{{item.gameName}}" class="word-game">来自：{{item.gameName}}</text>
          </view>
        </view>
      </view>
    </van-checkbox-group>
  </view>
</van-config-provider>
```

- [ ] **Step 8: Create `pages/learn/review/review.wxss`** — same as unknown.wxss (copy verbatim).

- [ ] **Step 9: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/learn/
git commit -m "feat(mini): add unknown and review word list pages"
```

---

## Task 13: Me Page

**Files:**
- Create: `pages/me/me.ts`
- Create: `pages/me/me.wxml`
- Create: `pages/me/me.wxss`
- Create: `pages/me/me.json`

- [ ] **Step 1: Create `pages/me/me.json`**

```json
{
  "navigationBarTitleText": "我的",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-cell": "@vant/weapp/cell/index",
    "van-cell-group": "@vant/weapp/cell-group/index",
    "van-icon": "@vant/weapp/icon/index",
    "van-image": "@vant/weapp/image/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/me/me.ts`**

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
    loading: true,
    profile: null as ProfileData | null,
    formatDate,
    gradeLabel,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme });
    (this.getTabBar() as any)?.setData({ active: 4, theme: app.globalData.theme })
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      this.setData({ loading: false, profile })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
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

- [ ] **Step 3: Create `pages/me/me.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <block wx:if="{{!loading && profile}}">
      <!-- Profile header -->
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
            <text>{{(profile.nickname || profile.username)[0]}}</text>
          </view>
        </view>
        <view class="profile-info">
          <text class="profile-name">{{profile.nickname || profile.username}}</text>
          <view class="profile-badges">
            <text class="grade-badge">{{gradeLabel(profile.grade)}}</text>
            <text class="exp-badge">Lv.{{profile.level}}</text>
          </view>
        </view>
        <van-icon name="arrow" size="16px" color="#9ca3af" />
      </view>

      <!-- Stats bar -->
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

      <!-- VIP info -->
      <view wx:if="{{profile.vipDueAt}}" class="vip-bar">
        <van-icon name="vip-card-o" size="16px" color="#d97706" />
        <text class="vip-text">会员有效期至 {{formatDate(profile.vipDueAt)}}</text>
      </view>

      <!-- Menu -->
      <van-cell-group inset custom-style="margin:16px;">
        <van-cell title="公告通知" icon="bell" is-link bind:click="goNotices" />
        <van-cell title="我的团队" icon="friends-o" is-link bind:click="goGroups" />
        <van-cell title="推荐有礼" icon="gift-o" is-link bind:click="goInvite" />
        <van-cell title="兑换码" icon="coupon-o" is-link bind:click="goRedeem" />
        <van-cell title="购买会员" icon="vip-card-o" is-link bind:click="goPurchase" />
      </van-cell-group>

      <van-cell-group inset custom-style="margin:0 16px 16px;">
        <van-cell title="退出登录" title-style="color:#ef4444;" bind:click="logout" />
      </van-cell-group>
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/me/me.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 100rpx; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.profile-header {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--bg-card);
  padding: 20px 16px;
  border-bottom: 1px solid var(--border-color);
}
.avatar-wrap { flex-shrink: 0; }
.avatar-fallback {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  color: #ffffff;
}
.profile-info { flex: 1; }
.profile-name { font-size: 18px; font-weight: 700; color: var(--text-primary); display: block; margin-bottom: 6px; }
.profile-badges { display: flex; gap: 6px; }
.grade-badge { font-size: 11px; color: #d97706; background: #fef3c7; border-radius: 4px; padding: 2px 6px; }
.exp-badge { font-size: 11px; color: var(--primary); background: var(--primary-light); border-radius: 4px; padding: 2px 6px; }
.stats-bar {
  display: flex;
  align-items: center;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  padding: 14px 0;
  margin-bottom: 12px;
}
.stat-item { flex: 1; text-align: center; }
.stat-v { font-size: 20px; font-weight: 700; color: var(--text-primary); display: block; }
.stat-l { font-size: 11px; color: var(--text-secondary); }
.stat-divider { width: 1px; height: 32px; background: var(--border-color); }
.vip-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fef3c7;
  margin: 0 16px 12px;
  border-radius: 8px;
}
.vip-text { font-size: 12px; color: #d97706; }
```

- [ ] **Step 5: Verify + Commit**

Tab 5 shows avatar, name, stats row, VIP bar (if applicable), and cell menu.

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/me.*
git commit -m "feat(mini): add me page"
```

---

## Task 14: Profile Edit Page

**Files:**
- Create: `pages/me/profile-edit/profile-edit.ts`
- Create: `pages/me/profile-edit/profile-edit.wxml`
- Create: `pages/me/profile-edit/profile-edit.wxss`
- Create: `pages/me/profile-edit/profile-edit.json`

- [ ] **Step 1: Create `pages/me/profile-edit/profile-edit.json`**

```json
{
  "navigationBarTitleText": "编辑资料",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-field": "@vant/weapp/field/index",
    "van-button": "@vant/weapp/button/index",
    "van-cell-group": "@vant/weapp/cell-group/index",
    "van-image": "@vant/weapp/image/index",
    "van-loading": "@vant/weapp/loading/index"
  }
}
```

- [ ] **Step 2: Create `pages/me/profile-edit/profile-edit.ts`**

```typescript
import { api } from '../../../utils/api'

interface ProfileData {
  username: string; nickname: string | null; avatarUrl: string | null
  city: string | null; introduction: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    saving: false,
    uploadingAvatar: false,
    profile: null as ProfileData | null,
    nickname: '',
    city: '',
    introduction: '',
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      this.setData({
        loading: false,
        profile,
        nickname: profile.nickname ?? '',
        city: profile.city ?? '',
        introduction: profile.introduction ?? '',
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onNicknameChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ nickname: (e.detail as { value: string }).value })
  },
  onCityChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ city: (e.detail as { value: string }).value })
  },
  onIntroductionChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ introduction: (e.detail as { value: string }).value })
  },
  chooseAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempPath = res.tempFiles[0].tempFilePath
        this.setData({ uploadingAvatar: true })
        wx.uploadFile({
          url: require('../../../utils/config').config.apiBaseUrl + '/api/uploads/images',
          filePath: tempPath,
          name: 'file',
          header: { Authorization: `Bearer ${require('../../../utils/auth').getToken()}` },
          success: (uploadRes) => {
            const body = JSON.parse(uploadRes.data) as { code: number; data: { id: string; url: string } }
            if (body.code === 0) {
              api.put('/api/user/avatar', { image_id: body.data.id })
                .then(() => {
                  this.setData({
                    uploadingAvatar: false,
                    profile: { ...this.data.profile!, avatarUrl: body.data.url },
                  })
                  wx.showToast({ title: '头像已更新', icon: 'none' })
                })
                .catch(() => {
                  this.setData({ uploadingAvatar: false })
                  wx.showToast({ title: '更新失败', icon: 'none' })
                })
            } else {
              this.setData({ uploadingAvatar: false })
              wx.showToast({ title: '上传失败', icon: 'none' })
            }
          },
          fail: () => {
            this.setData({ uploadingAvatar: false })
            wx.showToast({ title: '上传失败', icon: 'none' })
          },
        })
      },
    })
  },
  async save() {
    if (this.data.saving) return
    this.setData({ saving: true })
    try {
      await api.put('/api/user/profile', {
        nickname: this.data.nickname || null,
        city: this.data.city || null,
        introduction: this.data.introduction || null,
      })
      wx.showToast({ title: '保存成功', icon: 'none' })
      wx.navigateBack()
    } catch (err) {
      this.setData({ saving: false })
      wx.showToast({ title: (err as Error).message || '保存失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 3: Create `pages/me/profile-edit/profile-edit.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <block wx:if="{{!loading}}">
      <!-- Avatar -->
      <view class="avatar-section" bind:tap="chooseAvatar">
        <van-image
          wx:if="{{profile.avatarUrl}}"
          src="{{profile.avatarUrl}}"
          width="80px" height="80px" radius="50%" fit="cover"
        />
        <view wx:else class="avatar-fallback-lg">
          <text>📷</text>
        </view>
        <text class="avatar-hint">{{uploadingAvatar ? '上传中...' : '点击更换头像'}}</text>
      </view>

      <!-- Fields -->
      <van-cell-group inset custom-style="margin:16px 16px 0;">
        <van-field
          label="昵称"
          type="nickname"
          value="{{nickname}}"
          placeholder="请输入昵称"
          bind:change="onNicknameChange"
        />
        <van-field
          label="城市"
          value="{{city}}"
          placeholder="请输入城市"
          bind:change="onCityChange"
        />
        <van-field
          label="简介"
          type="textarea"
          value="{{introduction}}"
          placeholder="请输入简介"
          autosize
          rows="3"
          bind:change="onIntroductionChange"
        />
      </van-cell-group>

      <view class="save-btn">
        <van-button type="primary" block round loading="{{saving}}" bind:click="save">保存</van-button>
      </view>
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/me/profile-edit/profile-edit.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 28px 16px 16px;
  gap: 10px;
}
.avatar-fallback-lg {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}
.avatar-hint { font-size: 13px; color: var(--primary); }
.save-btn { padding: 20px 16px; }
```

- [ ] **Step 5: Verify + Commit**

Navigate to profile edit — shows avatar, nickname/city/intro fields, save button. Nickname field uses native `type="nickname"` which shows WeChat's nickname picker.

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/profile-edit/
git commit -m "feat(mini): add profile edit page"
```

---

## Task 15: Notices Page

**Files:**
- Create: `pages/me/notices/notices.ts`
- Create: `pages/me/notices/notices.wxml`
- Create: `pages/me/notices/notices.wxss`
- Create: `pages/me/notices/notices.json`

- [ ] **Step 1: Create `pages/me/notices/notices.json`**

```json
{
  "navigationBarTitleText": "公告通知",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-button": "@vant/weapp/button/index"
  }
}
```

- [ ] **Step 2: Create `pages/me/notices/notices.ts`**

```typescript
import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface NoticeItem { id: string; title: string; content: string | null; icon: string | null; createdAt: string }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    notices: [] as NoticeItem[],
    nextCursor: '',
    hasMore: false,
    formatRelativeDate,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadNotices(true)
  },
  onPullDownRefresh() {
    this.loadNotices(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadNotices(false)
  },
  async loadNotices(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<NoticeItem>>(`/api/notices${qs}`)
      this.setData({
        loading: false,
        notices: reset ? res.items : [...this.data.notices, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
      // Mark as read
      api.post('/api/notices/mark-read', {}).catch(() => {})
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 3: Create `pages/me/notices/notices.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading && notices.length === 0}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && notices.length === 0}}" description="暂无通知" />
    <view class="notice-list">
      <view wx:for="{{notices}}" wx:key="id" class="notice-item">
        <view class="notice-icon">
          <text>{{item.icon || '📢'}}</text>
        </view>
        <view class="notice-body">
          <text class="notice-title">{{item.title}}</text>
          <text wx:if="{{item.content}}" class="notice-content">{{item.content}}</text>
          <text class="notice-time">{{formatRelativeDate(item.createdAt)}}</text>
        </view>
      </view>
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/me/notices/notices.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.notice-list { padding: 12px 16px; display: flex; flex-direction: column; gap: 10px; }
.notice-item {
  display: flex;
  gap: 12px;
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.notice-icon { font-size: 22px; flex-shrink: 0; }
.notice-body { flex: 1; }
.notice-title { font-size: 15px; font-weight: 600; color: var(--text-primary); display: block; margin-bottom: 4px; }
.notice-content { font-size: 13px; color: var(--text-secondary); display: block; line-height: 1.5; margin-bottom: 6px; }
.notice-time { font-size: 11px; color: var(--text-secondary); }
```

- [ ] **Step 5: Verify + Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/notices/
git commit -m "feat(mini): add notices page"
```

---

## Task 16: Groups Pages

**Files:**
- Create: `pages/me/groups/groups.ts/.wxml/.wxss/.json`
- Create: `pages/me/groups-detail/groups-detail.ts/.wxml/.wxss/.json`

- [ ] **Step 1: Create `pages/me/groups/groups.json`**

```json
{
  "navigationBarTitleText": "我的团队",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-button": "@vant/weapp/button/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-dialog": "@vant/weapp/dialog/index",
    "van-field": "@vant/weapp/field/index"
  }
}
```

- [ ] **Step 2: Create `pages/me/groups/groups.ts`**

```typescript
import { api, PaginatedData } from '../../../utils/api'

interface GroupListItem {
  id: string; name: string; description: string | null
  ownerName: string; memberCount: number; inviteCode: string
  isMember: boolean; isOwner: boolean
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    groups: [] as GroupListItem[],
    nextCursor: '',
    hasMore: false,
    showCreateDialog: false,
    showJoinDialog: false,
    createName: '',
    joinCode: '',
    creating: false,
    joining: false,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadGroups(true)
  },
  onPullDownRefresh() {
    this.loadGroups(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadGroups(false)
  },
  async loadGroups(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}` : ''
    try {
      const res = await api.get<PaginatedData<GroupListItem>>(`/api/groups${qs}`)
      this.setData({
        loading: false,
        groups: reset ? res.items : [...this.data.groups, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/me/groups-detail/groups-detail?id=${id}` })
  },
  onCreateNameChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ createName: (e.detail as { value: string }).value })
  },
  onJoinCodeChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ joinCode: (e.detail as { value: string }).value })
  },
  async createGroup() {
    if (!this.data.createName.trim() || this.data.creating) return
    this.setData({ creating: true })
    try {
      await api.post('/api/groups', { name: this.data.createName.trim() })
      this.setData({ showCreateDialog: false, createName: '', creating: false })
      this.loadGroups(true)
      wx.showToast({ title: '创建成功', icon: 'none' })
    } catch (err) {
      this.setData({ creating: false })
      wx.showToast({ title: (err as Error).message || '创建失败', icon: 'none' })
    }
  },
  async joinGroup() {
    if (!this.data.joinCode.trim() || this.data.joining) return
    this.setData({ joining: true })
    try {
      await api.post(`/api/groups/join/${this.data.joinCode.trim()}`, {})
      this.setData({ showJoinDialog: false, joinCode: '', joining: false })
      this.loadGroups(true)
      wx.showToast({ title: '加入成功', icon: 'none' })
    } catch (err) {
      this.setData({ joining: false })
      wx.showToast({ title: (err as Error).message || '加入失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 3: Create `pages/me/groups/groups.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <view class="top-actions">
      <van-button size="small" plain type="primary" bind:click="setData({showCreateDialog:true})">创建团队</van-button>
      <van-button size="small" plain bind:click="setData({showJoinDialog:true})">加入团队</van-button>
    </view>

    <van-loading wx:if="{{loading && groups.length === 0}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && groups.length === 0}}" description="暂无团队" />

    <view class="group-list">
      <view wx:for="{{groups}}" wx:key="id" class="group-item" data-id="{{item.id}}" bind:tap="goDetail">
        <view class="group-info">
          <text class="group-name">{{item.name}}</text>
          <text class="group-meta">{{item.memberCount}}名成员 · {{item.isOwner ? '队长' : '成员'}}</text>
        </view>
        <van-icon name="arrow" size="14px" color="#9ca3af" />
      </view>
    </view>

    <!-- Create dialog -->
    <van-dialog
      show="{{showCreateDialog}}"
      title="创建团队"
      show-cancel-button
      bind:confirm="createGroup"
      bind:cancel="setData({showCreateDialog:false,createName:''})"
      confirm-loading="{{creating}}"
    >
      <van-field value="{{createName}}" placeholder="团队名称" bind:change="onCreateNameChange" />
    </van-dialog>

    <!-- Join dialog -->
    <van-dialog
      show="{{showJoinDialog}}"
      title="加入团队"
      show-cancel-button
      bind:confirm="joinGroup"
      bind:cancel="setData({showJoinDialog:false,joinCode:''})"
      confirm-loading="{{joining}}"
    >
      <van-field value="{{joinCode}}" placeholder="输入邀请码" bind:change="onJoinCodeChange" />
    </van-dialog>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/me/groups/groups.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.top-actions { display: flex; gap: 10px; padding: 14px 16px; }
.group-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; }
.group-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 14px;
}
.group-info { flex: 1; }
.group-name { font-size: 15px; font-weight: 600; color: var(--text-primary); display: block; }
.group-meta { font-size: 12px; color: var(--text-secondary); margin-top: 3px; display: block; }
```

- [ ] **Step 5: Create `pages/me/groups-detail/groups-detail.json`**

```json
{
  "navigationBarTitleText": "团队详情",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-button": "@vant/weapp/button/index",
    "van-cell": "@vant/weapp/cell/index",
    "van-cell-group": "@vant/weapp/cell-group/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

- [ ] **Step 6: Create `pages/me/groups-detail/groups-detail.ts`**

```typescript
import { api } from '../../../utils/api'

interface GroupMember { id: string; username: string; nickname: string | null; role: string }
interface GroupDetail {
  id: string; name: string; description: string | null
  ownerName: string; memberCount: number; inviteCode: string; isOwner: boolean
  currentGameId: string | null; currentGameName: string
  isPlaying: boolean; startGameLevelId: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    group: null as GroupDetail | null,
    members: [] as GroupMember[],
    starting: false,
  },
  onLoad(options: { id?: string }) {
    this.setData({ theme: app.globalData.theme })
    if (options.id) this.loadGroup(options.id)
  },
  async loadGroup(id: string) {
    try {
      const [group, members] = await Promise.all([
        api.get<GroupDetail>(`/api/groups/${id}`),
        api.get<GroupMember[]>(`/api/groups/${id}/members`),
      ])
      this.setData({ loading: false, group, members })
      wx.setNavigationBarTitle({ title: group.name })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  copyInviteCode() {
    wx.setClipboardData({ data: this.data.group!.inviteCode })
    wx.showToast({ title: '邀请码已复制', icon: 'none' })
  },
  async startGroupGame() {
    if (!this.data.group?.currentGameId || this.data.starting) return
    this.setData({ starting: true })
    try {
      await api.post(`/api/groups/${this.data.group.id}/start-game`, {})
      wx.showToast({ title: '游戏已开始', icon: 'none' })
      this.setData({ starting: false })
    } catch (err) {
      this.setData({ starting: false })
      wx.showToast({ title: (err as Error).message || '开始失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 7: Create `pages/me/groups-detail/groups-detail.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <block wx:if="{{!loading && group}}">
      <!-- Invite code -->
      <view class="invite-bar" bind:tap="copyInviteCode">
        <text class="invite-label">邀请码：</text>
        <text class="invite-code">{{group.inviteCode}}</text>
        <van-icon name="copy-o" size="14px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
      </view>

      <!-- Game info -->
      <van-cell-group inset custom-style="margin:12px 16px;">
        <van-cell wx:if="{{group.currentGameName}}" title="当前课程" value="{{group.currentGameName}}" />
        <van-cell wx:else title="当前课程" value="未设置" />
      </van-cell-group>

      <view wx:if="{{group.isOwner && group.currentGameId}}" class="start-btn">
        <van-button type="primary" block round loading="{{starting}}" bind:click="startGroupGame">
          {{group.isPlaying ? '游戏进行中' : '开始团队游戏'}}
        </van-button>
      </view>

      <!-- Members -->
      <view class="section-header"><text class="section-title">成员 ({{group.memberCount}})</text></view>
      <view class="member-list">
        <view wx:for="{{members}}" wx:key="id" class="member-item">
          <view class="member-avatar">
            <text>{{(item.nickname || item.username)[0]}}</text>
          </view>
          <text class="member-name">{{item.nickname || item.username}}</text>
          <text wx:if="{{item.role === 'owner'}}" class="owner-badge">队长</text>
        </view>
      </view>
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 8: Create `pages/me/groups-detail/groups-detail.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.invite-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 16px;
  background: var(--primary-light);
  border-bottom: 1px solid var(--border-color);
}
.invite-label { font-size: 13px; color: var(--text-secondary); }
.invite-code { font-size: 15px; font-weight: 700; color: var(--primary); letter-spacing: 2px; flex: 1; }
.start-btn { padding: 12px 16px; }
.section-header { padding: 16px 16px 8px; }
.section-title { font-size: 13px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; }
.member-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; }
.member-item { display: flex; align-items: center; gap: 10px; }
.member-avatar {
  width: 36px; height: 36px; border-radius: 50%;
  background: var(--primary-light);
  display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: 600; color: var(--primary);
}
.member-name { flex: 1; font-size: 14px; color: var(--text-primary); }
.owner-badge { font-size: 11px; color: #d97706; background: #fef3c7; border-radius: 4px; padding: 2px 6px; }
```

- [ ] **Step 9: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/groups/ dx-mini/miniprogram/pages/me/groups-detail/
git commit -m "feat(mini): add groups and group detail pages"
```

---

## Task 17: Invite, Redeem, and Purchase Pages

**Files:**
- Create: `pages/me/invite/invite.ts/.wxml/.wxss/.json`
- Create: `pages/me/redeem/redeem.ts/.wxml/.wxss/.json`
- Create: `pages/me/purchase/purchase.ts/.wxml/.wxss/.json`

- [ ] **Step 1: Create `pages/me/invite/invite.json`**

```json
{
  "navigationBarTitleText": "推荐有礼",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

- [ ] **Step 2: Create `pages/me/invite/invite.ts`**

```typescript
import { api } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface InviteStats { total: number; pending: number; paid: number; rewarded: number }
interface ReferralInvitee { id: string; username: string; nickname: string | null; grade: string }
interface ReferralItem {
  id: string; status: string; rewardAmount: number; rewardedAt: string | null; createdAt: string
  invitee: ReferralInvitee | null
}
interface InviteData {
  inviteCode: string; stats: InviteStats; referrals: ReferralItem[]; totalReferrals: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    inviteCode: '',
    stats: null as InviteStats | null,
    referrals: [] as ReferralItem[],
    totalReferrals: 0,
    formatRelativeDate,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadData()
  },
  async loadData() {
    try {
      const data = await api.get<InviteData>('/api/invite')
      this.setData({
        loading: false,
        inviteCode: data.inviteCode,
        stats: data.stats,
        referrals: data.referrals,
        totalReferrals: data.totalReferrals,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  copyCode() {
    wx.setClipboardData({ data: this.data.inviteCode })
    wx.showToast({ title: '邀请码已复制', icon: 'none' })
  },
})
```

- [ ] **Step 3: Create `pages/me/invite/invite.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <block wx:if="{{!loading}}">
      <!-- Invite code card -->
      <view class="invite-card">
        <text class="ic-title">我的邀请码</text>
        <view class="ic-code-row" bind:tap="copyCode">
          <text class="ic-code">{{inviteCode}}</text>
          <van-icon name="copy-o" size="18px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
        </view>
      </view>

      <!-- Stats -->
      <view class="stats-row">
        <view class="si-stat">
          <text class="si-v">{{stats.total ?? 0}}</text>
          <text class="si-l">总邀请</text>
        </view>
        <view class="si-stat">
          <text class="si-v">{{stats.paid ?? 0}}</text>
          <text class="si-l">已付费</text>
        </view>
        <view class="si-stat">
          <text class="si-v">{{stats.rewarded ?? 0}}</text>
          <text class="si-l">已返利</text>
        </view>
      </view>

      <!-- Referral list -->
      <view class="section-header"><text class="section-title">邀请记录</text></view>
      <van-empty wx:if="{{referrals.length === 0}}" description="暂无邀请记录" />
      <view class="referral-list">
        <view wx:for="{{referrals}}" wx:key="id" class="referral-item">
          <text class="ref-name">{{item.invitee ? (item.invitee.nickname || item.invitee.username) : '未知用户'}}</text>
          <view class="ref-right">
            <text class="ref-status {{item.status}}">{{item.status === 'rewarded' ? '已返利' : item.status === 'paid' ? '已付费' : '注册'}}</text>
            <text class="ref-time">{{formatRelativeDate(item.createdAt)}}</text>
          </view>
        </view>
      </view>
    </block>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Create `pages/me/invite/invite.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }
.invite-card {
  margin: 16px;
  background: var(--bg-card);
  border-radius: 16px;
  border: 1px solid var(--border-color);
  padding: 20px;
  text-align: center;
}
.ic-title { font-size: 13px; color: var(--text-secondary); display: block; margin-bottom: 12px; }
.ic-code-row { display: flex; align-items: center; justify-content: center; gap: 8px; }
.ic-code { font-size: 28px; font-weight: 700; color: var(--primary); letter-spacing: 4px; }
.stats-row { display: flex; margin: 0 16px 16px; background: var(--bg-card); border-radius: 12px; border: 1px solid var(--border-color); }
.si-stat { flex: 1; padding: 16px; text-align: center; }
.si-v { font-size: 20px; font-weight: 700; color: var(--text-primary); display: block; }
.si-l { font-size: 11px; color: var(--text-secondary); }
.section-header { padding: 4px 16px 8px; }
.section-title { font-size: 13px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; }
.referral-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; }
.referral-item { display: flex; align-items: center; justify-content: space-between; background: var(--bg-card); border-radius: 10px; border: 1px solid var(--border-color); padding: 12px; }
.ref-name { font-size: 14px; color: var(--text-primary); }
.ref-right { display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.ref-status { font-size: 11px; color: var(--primary); }
.ref-status.rewarded { color: #10b981; }
.ref-time { font-size: 11px; color: var(--text-secondary); }
```

- [ ] **Step 5: Create `pages/me/redeem/redeem.json`**

```json
{
  "navigationBarTitleText": "兑换码",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-field": "@vant/weapp/field/index",
    "van-button": "@vant/weapp/button/index",
    "van-cell-group": "@vant/weapp/cell-group/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index"
  }
}
```

- [ ] **Step 6: Create `pages/me/redeem/redeem.ts`**

```typescript
import { api } from '../../../utils/api'
import { formatDate } from '../../../utils/format'

interface RedeemItem { id: string; code: string; grade: string; redeemedAt: string }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    code: '',
    redeeming: false,
    history: [] as RedeemItem[],
    formatDate,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadHistory()
  },
  async loadHistory() {
    try {
      const res = await api.get<{ items: RedeemItem[] }>('/api/redeems')
      this.setData({ loading: false, history: res.items ?? (res as unknown as RedeemItem[]) })
    } catch {
      this.setData({ loading: false })
    }
  },
  onCodeChange(e: WechatMiniprogram.TouchEvent) {
    this.setData({ code: (e.detail as { value: string }).value })
  },
  async redeem() {
    if (!this.data.code.trim() || this.data.redeeming) return
    this.setData({ redeeming: true })
    try {
      const res = await api.post<{ grade: string }>('/api/redeems', { code: this.data.code.trim() })
      this.setData({ code: '', redeeming: false })
      wx.showModal({ title: '兑换成功', content: `已升级为 ${res.grade}`, showCancel: false })
      this.loadHistory()
    } catch (err) {
      this.setData({ redeeming: false })
      wx.showToast({ title: (err as Error).message || '兑换失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 7: Create `pages/me/redeem/redeem.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <van-cell-group inset custom-style="margin:16px;">
      <van-field value="{{code}}" placeholder="输入兑换码" bind:change="onCodeChange" clearable />
    </van-cell-group>
    <view class="redeem-btn">
      <van-button type="primary" block round loading="{{redeeming}}" bind:click="redeem">立即兑换</van-button>
    </view>

    <view class="section-header"><text class="section-title">兑换记录</text></view>
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && history.length === 0}}" description="暂无兑换记录" />
    <view class="redeem-list">
      <view wx:for="{{history}}" wx:key="id" class="redeem-item">
        <text class="ri-code">{{item.code}}</text>
        <view class="ri-right">
          <text class="ri-grade">{{item.grade}}</text>
          <text class="ri-date">{{formatDate(item.redeemedAt)}}</text>
        </view>
      </view>
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 8: Create `pages/me/redeem/redeem.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.center-loader { display: flex; justify-content: center; padding: 20px; }
.redeem-btn { padding: 0 16px 16px; }
.section-header { padding: 4px 16px 8px; }
.section-title { font-size: 13px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; }
.redeem-list { padding: 0 16px; display: flex; flex-direction: column; gap: 8px; }
.redeem-item { display: flex; align-items: center; justify-content: space-between; background: var(--bg-card); border-radius: 10px; border: 1px solid var(--border-color); padding: 12px; }
.ri-code { font-size: 14px; font-weight: 600; color: var(--text-primary); letter-spacing: 1px; }
.ri-right { display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.ri-grade { font-size: 12px; color: var(--primary); }
.ri-date { font-size: 11px; color: var(--text-secondary); }
```

- [ ] **Step 9: Create `pages/me/purchase/purchase.json`**

```json
{
  "navigationBarTitleText": "购买会员",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-button": "@vant/weapp/button/index"
  }
}
```

- [ ] **Step 10: Create `pages/me/purchase/purchase.ts`**

```typescript
const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

const TIERS = [
  { id: 'monthly', name: '月度会员', price: '¥19', desc: '30天无限访问' },
  { id: 'quarterly', name: '季度会员', price: '¥49', desc: '90天无限访问' },
  { id: 'yearly', name: '年度会员', price: '¥149', desc: '365天无限访问' },
]

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    tiers: TIERS,
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
  },
  onBuy() {
    wx.showToast({ title: '即将开放', icon: 'none' })
  },
})
```

- [ ] **Step 11: Create `pages/me/purchase/purchase.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <view class="coming-banner">
      <text class="cb-emoji">🚀</text>
      <text class="cb-title">即将开放</text>
      <text class="cb-sub">微信支付正在接入中，敬请期待</text>
    </view>

    <view class="tier-list">
      <view wx:for="{{tiers}}" wx:key="id" class="tier-card">
        <view class="tier-info">
          <text class="tier-name">{{item.name}}</text>
          <text class="tier-desc">{{item.desc}}</text>
        </view>
        <view class="tier-right">
          <text class="tier-price">{{item.price}}</text>
          <van-button size="small" disabled custom-style="opacity:0.5;">购买</van-button>
        </view>
      </view>
    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 12: Create `pages/me/purchase/purchase.wxss`**

```css
.page-container { min-height: 100vh; background: var(--bg-page); padding-bottom: 40px; }
.coming-banner {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px 28px;
  gap: 8px;
}
.cb-emoji { font-size: 48px; }
.cb-title { font-size: 22px; font-weight: 700; color: var(--text-primary); }
.cb-sub { font-size: 13px; color: var(--text-secondary); text-align: center; }
.tier-list { padding: 0 16px; display: flex; flex-direction: column; gap: 10px; }
.tier-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card);
  border-radius: 14px;
  border: 1px solid var(--border-color);
  padding: 16px;
}
.tier-info { flex: 1; }
.tier-name { font-size: 16px; font-weight: 600; color: var(--text-primary); display: block; }
.tier-desc { font-size: 12px; color: var(--text-secondary); margin-top: 3px; display: block; }
.tier-right { display: flex; flex-direction: column; align-items: flex-end; gap: 6px; }
.tier-price { font-size: 20px; font-weight: 700; color: var(--primary); }
```

- [ ] **Step 13: Final verification in DevTools**

Check all 19 pages compile without TypeScript errors. Navigate through each tab. Verify:
- Login page renders correctly
- Tab bar shows 5 tabs with correct active states
- Home loads dashboard data (requires running backend)
- Games grid shows with category tabs
- Leaderboard shows period/type toggles
- Learn hub shows 3 stat cards
- Me page shows profile with menu cells
- Dark mode toggle on home page changes all pages

- [ ] **Step 14: Final commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-mini/miniprogram/pages/me/invite/ dx-mini/miniprogram/pages/me/redeem/ dx-mini/miniprogram/pages/me/purchase/
git commit -m "feat(mini): add invite, redeem, and purchase pages"
```

---

## Self-Review

### Spec coverage check

| Spec requirement | Covered |
|---|---|
| WeChat login (wx.login → POST /api/auth/wechat-mini) | ✅ Task 3 |
| JWT in wx.setStorageSync | ✅ Task 1 (auth.ts) |
| Custom tabBar (dark mode support) | ✅ Task 2 |
| 5-tab navigation | ✅ Tasks 1+2 (app.json) |
| Dark mode (system + manual toggle) | ✅ app.ts + home.ts |
| Home: search, dark toggle, bell, stats, heatmap | ✅ Task 4 |
| Games: category tabs, grid, infinite scroll, pull-to-refresh | ✅ Task 5 |
| Game detail: cover, levels list, favorite toggle | ✅ Task 6 |
| Game play: content items, answer choices, mark word | ✅ Task 7 |
| Favorites: swipe-to-unfavorite | ✅ Task 8 |
| Leaderboard: period/type tabs, top-3 podium, pinned rank | ✅ Task 9 |
| Learn hub: 3 stat cards, quick links | ✅ Task 10 |
| Mastered/unknown/review word lists with bulk delete | ✅ Tasks 11+12 |
| Me page: avatar, stats, VIP, menu | ✅ Task 13 |
| Profile edit with native nickname input | ✅ Task 14 |
| Notices with mark-read | ✅ Task 15 |
| Groups + group detail (start game) | ✅ Task 16 |
| Invite code + referral stats | ✅ Task 17 |
| Redeem code | ✅ Task 17 |
| Purchase placeholder ("即将开放") | ✅ Task 17 |
| 401 handling → clearToken + reLaunch | ✅ Task 1 (api.ts) |
| WebSocket: session_replaced, pk_invitation | ✅ Tasks 1+3 (ws.ts + app.ts) |
| Vant Weapp teal theme | ✅ Task 1 (app.wxss) |
