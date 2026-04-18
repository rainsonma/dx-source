# WeChat Mini Program Design — 斗学

Date: 2026-04-18

## Overview

Build the `dx-mini` WeChat mini program frontend for 斗学. It exposes all authenticated operations from the platform (games, learning tracking, leaderboard, groups, profile, referrals, purchases) using WeChat-native login. The web landing page, wiki, and public pages are out of scope — mini program users are always authenticated.

The existing dx-api backend is extended with WeChat auth and platform tracking. The mini program uses Vant Weapp as its component library, themed to match the existing teal-600 design system from dx-web (light + dark mode).

---

## Backend Changes (dx-api)

### 1. Modified migrations (in-place)

**`database/migrations/20260322000001_create_users_table.go`** — add two fields:

```
openid   text, unique, nullable   -- WeChat mini program user identity
unionid  text, nullable           -- WeChat union identity across apps
```

Add index on `openid`.

**`database/migrations/20260322000009_create_user_logins_table.go`** — add one field:

```
platform  text, nullable   -- "website" | "mini" | "ios" | "android"
```

### 2. New constants

**`app/consts/platform.go`**

```go
const (
    PlatformWebsite  = "website"
    PlatformMini     = "mini"
    PlatformIOS      = "ios"
    PlatformAndroid  = "android"
)
```

### 3. Updated User model

**`app/models/user.go`** — add two fields:

```go
OpenID  *string `gorm:"column:openid"  json:"-"`
UnionID *string `gorm:"column:unionid" json:"-"`
```

Both excluded from JSON responses.

### 4. Updated UserLogin model

**`app/models/user_login.go`** — add one field:

```go
Platform *string `gorm:"column:platform" json:"platform"`
```

### 5. Updated RecordLogin

**`app/services/api/auth_service.go`** — `RecordLogin` accepts a `platform string` parameter and writes it to the new column.

All existing callers (website signin) pass `consts.PlatformWebsite`. The new WeChat handler passes `consts.PlatformMini`.

### 6. New WeChat mini auth endpoint

**Route:** `POST /api/auth/wechat-mini` (public, no JWT required)

**Request:**
```json
{ "code": "wx_login_code" }
```

**Flow:**
1. Exchange `code` with WeChat server:
   `GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=CODE&grant_type=authorization_code`
2. Receive `{ openid, session_key, unionid? }`
3. Find user by `openid`:
   - **Found** → sign in directly, issue JWT
   - **Not found** → auto-register:
     - `username` = `"wx_" + openid[:8]`; suffix with random 4-char code if taken
     - `nickname` = nickname from request (WeChat display name)
     - Store `openid`, `unionid`
     - `password` = random 16-char string (hashed)
     - `invite_code` = random 8-char code
4. Call `RecordLogin(userID, ip, userAgent, consts.PlatformMini)` asynchronously
5. Return `{ token, user }` — same shape as existing auth endpoints

**New files:**
- `app/http/controllers/api/wechat_auth_controller.go`
- `app/http/requests/api/wechat_auth_request.go`
- `app/services/api/wechat_auth_service.go`

**New env vars (dx-api `.env`):**
```
WECHAT_MINI_APP_ID=wx6ffd2fe38aaf0c96
WECHAT_MINI_APP_SECRET=
```

**New config file `config/wechat.go`:**
```go
"wechat_mini_app_id":     os.Getenv("WECHAT_MINI_APP_ID"),
"wechat_mini_app_secret": os.Getenv("WECHAT_MINI_APP_SECRET"),
```

---

## Mini Program Architecture (dx-mini)

### Framework & tooling

- **WXML/WXSS/TypeScript** — native WeChat (already configured with glass-easel)
- **Vant Weapp** (`@vant/weapp`) — installed via npm, themed to teal
- **No Taro / uni-app** — vanilla mini program only

### Project structure

```
miniprogram/
├── app.json                  # Pages + custom tabBar config
├── app.ts                    # App lifecycle, token check, theme init
├── app.wxss                  # Global styles, Vant CSS variable overrides
├── custom-tab-bar/           # Custom tabBar component (dark mode support)
│   ├── index.ts
│   ├── index.wxml
│   ├── index.wxss
│   └── index.json
├── pages/
│   ├── login/                # WeChat login (pre-auth)
│   ├── home/                 # Tab 1: 首页
│   ├── games/
│   │   ├── games/            # Tab 2: 课程 (root)
│   │   ├── detail/           # Game detail
│   │   ├── play/             # Single-player session
│   │   └── favorites/        # Favorited games
│   ├── leaderboard/          # Tab 3: 排行榜 (root)
│   ├── learn/
│   │   ├── learn/            # Tab 4: 学习 (root)
│   │   ├── mastered/
│   │   ├── unknown/
│   │   └── review/
│   └── me/
│       ├── me/               # Tab 5: 我的 (root)
│       ├── profile-edit/
│       ├── notices/
│       ├── groups/           # Groups list
│       ├── groups-detail/    # Group detail
│       ├── invite/
│       ├── redeem/
│       └── purchase/         # Placeholder
└── utils/
    ├── api.ts                # HTTP client wrapper
    ├── auth.ts               # Token read/write/clear
    ├── config.ts             # BASE_URL per envVersion (develop/trial/release)
    ├── ws.ts                 # WebSocket wrapper
    └── format.ts             # Date/number formatters
```

### API client (`utils/api.ts`)

Mirrors dx-web's `apiClient`:

```typescript
// BASE_URL comes from utils/config.ts which exports the right value per __wxConfig.envVersion
// ('develop' → http://localhost:3001, 'trial'/'release' → production URL)

const BASE_URL = config.apiBaseUrl

function request<T>(method: string, path: string, data?: object): Promise<T>
  // Reads JWT from wx.getStorageSync('dx_token')
  // Sets Authorization: Bearer {token} header
  // Wraps wx.request()
  // On code !== 0: throws with message
  // On HTTP 401 + code 40104: wx.showModal "账号已在其他设备登录" → clear token → redirect to login
  // On HTTP 401 other: clear token → redirect to login

export const api = {
  get:    <T>(path: string)              => request<T>('GET', path),
  post:   <T>(path: string, data: object) => request<T>('POST', path, data),
  put:    <T>(path: string, data: object) => request<T>('PUT', path, data),
  delete: <T>(path: string, data?: object) => request<T>('DELETE', path, data),
}
```

### Auth flow (`utils/auth.ts`)

```typescript
getToken():   string | null   // wx.getStorageSync('dx_token')
setToken(t):  void            // wx.setStorageSync('dx_token', t)
clearToken(): void            // wx.removeStorageSync('dx_token')
isLoggedIn(): boolean         // !!getToken()
```

**Login page flow:**
1. User taps "使用微信登录" button
2. `wx.login()` → get `code`
3. `POST /api/auth/wechat-mini` with `{ code }` — no nickname sent
4. Backend derives username as `"wx_" + openid[:8]` on auto-register; nickname left null initially
5. `setToken(data.token)`, store `data.user.id`
6. `wx.reLaunch({ url: '/pages/home/home' })`

Note: `wx.getUserProfile()` is deprecated (base library ≥ 2.21.2). Nickname is left blank on registration and set by the user via `pages/me/profile-edit` using the native `<input type="nickname">` component.

**App launch check (`app.ts` `onLaunch`):**
- If `!isLoggedIn()` → `wx.reLaunch({ url: '/pages/login/login' })`

### Dark mode

- `wx.getSystemSetting().theme` → `'dark' | 'light'`
- Subscribe to `wx.onThemeChange()` to react dynamically
- Store preference in `wx.setStorageSync('dx_theme', theme)`
- Feed into `<van-config-provider theme="dark|light">` wrapping each page
- Manual toggle on home navbar writes to storage and calls `setData({ theme })`
- **Custom tabBar required**: the native `tabBar` does not support dynamic color changes. A custom tabBar component (`custom-tab-bar/`) replaces the native one, reading the current theme from the app-level store and applying teal/dark styles accordingly.

### WebSocket (`utils/ws.ts`)

Thin wrapper over `wx.connectSocket()` implementing the same topic/event protocol as dx-web's WebSocket provider:

```typescript
connect(token: string): void     // ws://api/ws with Authorization header
subscribe(topic: string): void   // send { type: 'subscribe', topic }
on(event: string, cb): void      // event listener
disconnect(): void
```

Used for: PK invitations, group game state updates, session-replaced kicks.

---

## Pages & Navigation

### Tab bar (`app.json`)

```json
{
  "tabBar": {
    "custom": true,
    "list": [
      { "pagePath": "pages/home/home" },
      { "pagePath": "pages/games/games" },
      { "pagePath": "pages/leaderboard/leaderboard" },
      { "pagePath": "pages/learn/learn" },
      { "pagePath": "pages/me/me" }
    ]
  }
}
```

`"custom": true` delegates rendering to `custom-tab-bar/index` which applies teal/dark styles dynamically. The `list` entries are still required by WeChat for routing but `color`/`backgroundColor` etc. are ignored when custom is true.

### Page inventory

| Path | Content |
|------|---------|
| `pages/login/login` | WeChat avatar · "使用微信登录" button · wx.login flow |
| `pages/home/home` | Search bar · dark toggle · bell → notices · greeting · streak/level/exp stats · recommended game · 7-week heatmap |
| `pages/games/games` | Tab 2 root · category filter tabs · game card grid · favorites toggle · pull-to-refresh · infinite scroll |
| `pages/games/detail` | Game cover · description · levels list · per-level stats · favorite button |
| `pages/games/play` | Full-screen game session · answer choices · score · combo · mark word (mastered/unknown/review) |
| `pages/games/favorites` | Favorited games list · unfavorite swipe action |
| `pages/leaderboard/leaderboard` | Tab 3 root · period (日/周/月) · type (经验值/游戏时长) · top-3 podium · ranked list · current user pinned |
| `pages/learn/learn` | Tab 4 root · 3 stat cards (mastered/unknown/review) · weekly summary · nav to each list |
| `pages/learn/mastered` | Mastered words list · count · bulk delete |
| `pages/learn/unknown` | Unknown words list · count · bulk delete |
| `pages/learn/review` | Review queue · count · bulk delete |
| `pages/me/me` | Tab 5 root · avatar · nickname · membership badge · beans · VIP expiry · menu list |
| `pages/me/profile-edit` | Edit nickname · city · introduction · avatar upload (native `<input type="nickname">`) |
| `pages/me/notices` | Announcements list · mark read · unread count badge |
| `pages/me/groups` | My groups · create group · join by invite code |
| `pages/me/groups-detail` | Group info · members · set game · start group session |
| `pages/me/invite` | Invite code · share · referral list · conversion stats · commissions |
| `pages/me/redeem` | Input redeem code · history |
| `pages/me/purchase` | Membership tiers · beans packages · "即将开放" placeholder, no payment logic |

---

## UI Design System

### Colors

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| Primary | `#0d9488` (teal-600) | `#14b8a6` (teal-400) | Buttons, active tab, accents |
| Background | `#f5f5f5` | `#0f0f0f` | Page background |
| Card | `#ffffff` | `#1c1c1e` | Card surfaces |
| Card border | `#f0f0f0` | `rgba(255,255,255,0.06)` | Card edges |
| Text primary | `#1a1a1a` | `#f5f5f5` | Headings, body |
| Text secondary | `#888888` | `#6b7280` | Captions, labels |
| Tint surface | `#f0fdfa` | `rgba(20,184,166,0.12)` | Stat card backgrounds |
| Destructive | `#ef4444` | `#ef4444` | Delete, errors |

### Vant CSS variable overrides (`app.wxss`)

```css
page {
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
}

page.dark {
  --van-primary-color: #14b8a6;
  --van-background: #0f0f0f;
  --van-background-2: #1c1c1e;
  --van-text-color: #f5f5f5;
  --van-text-color-2: #6b7280;
  --van-border-color: rgba(255,255,255,0.06);
}
```

### Typography

- Font: system default (`-apple-system`, `PingFang SC`)
- Page title: 17px semibold
- Section heading: 13px semibold, `#888` uppercase
- Body: 14px regular
- Caption: 11px, secondary color

### Spacing & radius

- Page padding: 16px horizontal
- Card border-radius: 12px
- Button border-radius: 20px (pill) for primary CTAs, 8px for secondary
- Gap between cards: 12px

---

## Key Flows

### WeChat login

```
App launch
  → isLoggedIn()? NO → pages/login/login
  → tap "使用微信登录"
  → wx.login() → code
  → POST /api/auth/wechat-mini { code }
  → setToken(data.token), store userId
  → wx.reLaunch({ url: '/pages/home/home' })
```

### Game play session

```
pages/games/games → tap game card
  → pages/games/detail (gameId)
  → tap level → POST /api/play-single/start { gameId, levelId, degree, pattern }
  → pages/games/play (sessionId)
  → for each item:
      → POST /api/play-single/{id}/answers
      → optional: POST /api/tracking/master|unknown|review
  → POST /api/play-single/{id}/end { score, exp, ... }
  → show results modal → back to detail
```

### 401 handling

```
api.request() receives HTTP 401
  → code === 40104: wx.showModal("账号已在其他设备登录") → clearToken() → reLaunch login
  → other 401: clearToken() → reLaunch login
```

### Realtime (WebSocket)

```
app.ts onLaunch (if logged in)
  → ws.connect(token)
  → ws.subscribe('user::' + userId)
  → on 'session_replaced': clearToken() → reLaunch login
  → on 'pk_invitation': wx.showModal → accept/decline
```

---

## Feature Scope

### Included

All authenticated operations from dx-web except the two below.

### Out of scope for mini

| Feature | Reason |
|---------|--------|
| AI custom game creation | Multi-step content generation wizard — desktop UX only |
| Community posts/comments | Content creation doesn't fit mini program UX |
| Admin functions | No admin interface in mini |

### Payment (placeholder)

`pages/me/purchase/purchase` shows membership tiers and beans packages with pricing, but all purchase buttons show "即将开放" (coming soon). No order creation or payment logic. The WeChat payment callback endpoint already exists in dx-api and is unaffected.

---

## Out-of-scope changes

- No changes to dx-web
- No changes to existing dx-api routes or business logic
- No changes to existing migrations other than adding the three new fields
