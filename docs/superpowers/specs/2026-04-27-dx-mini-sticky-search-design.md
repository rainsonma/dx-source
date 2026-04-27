# dx-mini Sticky 搜索课程 Search Bar — Design Spec

**Date:** 2026-04-27
**Scope:** `dx-mini` (WeChat Mini Program) + one new public endpoint in `dx-api` (`/api/games/search-suggestions`)
**Target pages:** `pages/home/home`, `pages/games/games`, **new** `pages/games/search/search`
**Out of scope:** dx-web, deploy/, DB schema, dx-mini deployment pipeline

## Purpose

The mini program currently has a static "搜索课程" pill embedded in the home page's teal hero band that simply navigates to the courses page (no query passed). The courses page has no search input at all. Both entry points should:

1. Display a fixed-at-top search-launcher pill alongside the WeChat capsule button while content scrolls.
2. Tap the launcher to open a dedicated search page with a real `<input>`, recent-search history, popular-search discovery, and a results grid that reuses the existing game-card style.

The home page keeps a Taobao-style behavior: the launcher is hidden on initial paint and reveals (fade + slide, 200ms) once the hero band scrolls past the top. The courses page keeps the launcher always pinned and additionally pins the existing `<van-tabs>` row directly under it.

## Current State

### dx-mini

- **`pages/home/home.json`**: `"navigationStyle": "custom"` is already set.
- **`pages/home/home.wxml`** (line 22): a static `<view class="search-box" bind:tap="goSearch">` lives inside the teal-wrap, with placeholder text `搜索课程` and a `dx-icon` glyph. `home.wxss:75-100` styles it (`border-radius: 22px`, white pill, drop shadow). `home.ts:97` handler is `goSearch() { wx.navigateTo({ url: '/pages/games/games' }) }` — **no query is passed**.
- **`pages/games/games.json`**: `"navigationStyle": "custom"` is already set.
- **`pages/games/games.wxss:4`** uses `padding-top: calc(var(--status-bar-height, 20px) + 88rpx)` (≈ status bar + 44px) to clear the WeChat capsule area. CSS variable is set inline in `games.wxml:2` via `style="--status-bar-height: {{statusBarHeight}}px"`. No `.fixed-row` exists today.
- **`pages/games/games.wxml:4-18`** renders `<van-tabs>` directly (NOT sticky). Below it: a `<fav-bar>` (line 21), then the inline 2-column game card grid (lines 27-57), then the loading/empty footer (lines 59-63).
- **`utils/api.ts`** exports `api.{get,post,put,delete}` and `interface PaginatedData<T> { items: T[]; nextCursor: string; hasMore: boolean }`. Bearer token from `getToken()` injected into headers. `code: 0` → resolve with `data`; `code === 40104` → modal + relaunch login; other non-zero → reject. `response.code !== 0` rejects with `body.message`.
- **`components/dx-icon/`**: `<dx-icon name="..." size="..." color="..." strokeWidth="1.25" />` renders Lucide SVG via data-URL. Available icons in `scripts/build-icons.mjs:16-53` include `search`, `chevron-left`, `arrow-right`, `star`, `book-open`, etc. **Missing icons we need**: `x`, `circle-x`, `trash-2`, `search-x`, `arrow-left`. (`scripts/build-icons.mjs` enforces declared-icon discipline at build time — every literal `<dx-icon name="...">` must be in the `ICONS` array.)
- **`app.json:2-25`** lists 23 pages. `tabBar.custom: true`; `pages/games/games` is the second tab (custom-tab-bar at index 1).
- **No existing `wx.setStorageSync('dx_recent_searches'`** or similar — the key is unused. Storage keys in active use: `dx_token`, `dx_user_id`, `dx_dev_api_base_url`, `dx_theme`.
- **No existing `<game-card>` component** — the card markup is inlined in `games.wxml:27-57` with class names `.game-grid`, `.game-card`, `.game-cover`, `.game-info`, `.game-name`, `.game-meta`. Styles live in `games.wxss:20-79`.
- **Theme**: every page reads `getApp().globalData.theme` ('light'|'dark'), wraps content in `<van-config-provider theme="{{theme}}">`, toggles `.dark` class on `.page-container`. Light primary `#0d9488`, dark primary `#14b8a6`, dark bg `#1c1c1e`.
- **Loading / empty / error UX** (canonical pattern from `games.ts:46-66`): `loading` boolean in data, `try/catch` wraps `api.get`, on catch `wx.showToast({ title: '加载失败', icon: 'none' })`. Empty state via `<van-empty description="..." />`.

### dx-api

- **`routes/api.go:47-52`** registers public game routes inside `r.Prefix("/api").Group(...)`:

  ```go
  router.Prefix("/games").Group(func(games route.Router) {
      games.Get("/", gameController.List)
      games.Get("/search", gameController.Search)
  })
  ```

  Protected `GET /api/games/{id}` and `GET /api/games/played` live further down inside `protected.Middleware(middleware.JwtAuth())` (lines 106-107). The new `search-suggestions` endpoint slots into the public group next to `search`.

- **`GET /api/games?q=&cursor=&limit=&categoryIds=&pressId=&mode=`** — already supports name search via `ILIKE` on `q` (case-insensitive), cursor pagination via `services.ListPublishedGames`, and returns `helpers.Paginated(ctx, games, nextCursor, hasMore)` envelope. Used today by `pages/games/games.ts:56`. **No backend change needed** for the result list on the search page.

- **`Game` model** (`app/models/game.go`) fields: `ID, Name, Description, UserID, Mode, GameCategoryID, GamePressID, Icon, CoverURL, Order, IsActive, Status, IsSelective, IsPrivate`. **No `play_count` column.**

- **`GameSession` model** (`app/models/game_session.go`) has `GameID`. Existing aggregation patterns in `app/services/api/game_stats_service.go` and `app/services/api/leaderboard_service.go` use `facades.Orm().Query().Raw(...).Scan(&rows)` with PostgreSQL `COUNT(*)` and `GROUP BY`. Pattern to mirror.

- **Redis helpers** (`app/helpers/redis.go`): `RedisGet(key) (string, error)`, `RedisSet(key, value string, ttl time.Duration) error`, `RedisDel(key) error`. **No existing aggregation cache pattern in services**, but the email rate-limit pattern in `app/services/api/email_service.go` shows the call shape. We will be the first to do JSON-encoded SQL-result caching; pattern below.

- **Response envelope** — `helpers.Success(ctx, data)` returns `{code:0,message:"ok",data:...}`. `helpers.Paginated(...)` for cursor-paginated lists.

- **Request validators** in `app/http/requests/api/game_request.go` follow Goravel's `FormRequest` pattern with `Authorize`, `Rules`, `Filters`, `Messages`. The new endpoint takes no params, so a validator is unnecessary — controller can skip directly to service.

## Goals

1. One reusable component (`dx-search-bar`) handles capsule-aware fixed positioning on home + courses + search pages. Single source of truth for the geometry that `wx.getMenuButtonBoundingClientRect()` yields.
2. Home page: launcher pinned at top, hidden on cold open, fades + slides in (200ms) once the hero scrolls past. The existing in-hero static pill stays and also navigates to the search page (now with no query).
3. Courses page: launcher pinned at top; tabs row pinned directly under the launcher; favorites row + grid scroll under both. Existing tab/grid logic untouched.
4. New `pages/games/search/search` page: real `<input>` in a custom navbar row `[取消][input ...][capsule]`; idle state shows 最近搜索 + 搜索发现 sections; submit-only triggers (keyboard 搜索 OR chip tap); results render as a 2-column grid using the same card style as the courses page; cursor-paginated via `onReachBottom`.
5. New `GET /api/games/search-suggestions` returns up to 12 string chips (top game names by play count + top categories), Redis-cached 1h.
6. No regression to existing flows. No `?.` / `??` in TS or WXML (per `feedback_dx_mini_no_optional_chaining.md`). No `console.log` in production code.

## Non-Goals

- No new DB tables, no schema migrations.
- No search-log table; popularity is computed live and cached.
- No personalized "猜你想学" — use device-local recents; not user-scoped.
- No live (debounced) search — submit-only, by explicit user choice.
- No replacement of dx-web's hall search dialog. dx-web is untouched.
- No `<game-card>` component refactor for the courses page — inline copy of the same WXML markup in the search page is acceptable and avoids cross-cutting churn. (Promote to component later if a third page needs it.)

## Design

### 1. New shared component: `dx-search-bar`

Location: `dx-mini/miniprogram/components/dx-search-bar/{index.ts, index.wxml, index.wxss, index.json}`.

**Purpose:** the fixed-top launcher row visible on home + courses, plus the row used by the search page itself (with input slot instead of placeholder text).

**Public properties:**

| Prop | Type | Default | Notes |
|---|---|---|---|
| `theme` | `String` (`'light' \| 'dark'`) | `'light'` | Pulled from page-level `theme`. |
| `pinned` | `Boolean` | `true` | When `true`, host is `position: fixed; top:0; left:0; right:0; z-index: 90`. When `false`, lives in document flow (used by home's in-hero pill). |
| `revealed` | `Boolean` | `true` | Drives `.revealed` class. When `false`: `opacity:0; transform: translateY(-4px); pointer-events:none;`. CSS transitions: `opacity 200ms ease, transform 200ms ease`. |
| `placeholder` | `String` | `'搜索课程'` | Shown inside the launcher pill (display-only; not a real input). |
| `mode` | `String` (`'launcher' \| 'input'`) | `'launcher'` | `launcher` renders the static pill (tap → emit `tap`). `input` renders the row with a `<slot/>` for the page to inject its `<input>` and `<×>` icon. |
| `showCancel` | `Boolean` | `false` | When `true` (search page), renders a `取消` text on the far left of the row, which emits `cancel`. |

**Events:**

| Event | When |
|---|---|
| `bind:tap` | User taps the launcher pill (only when `mode='launcher'`). |
| `bind:cancel` | User taps `取消` (only when `mode='input' && showCancel`). |

**Computed geometry (set in `attached()` and re-set in `wx.onWindowResize` if available):**

```ts
const sys = wx.getSystemInfoSync()
const cap = wx.getMenuButtonBoundingClientRect()
const statusBarHeight = sys.statusBarHeight || 20
const rowHeight = (cap.bottom - statusBarHeight) + 8 // a few px breathing room below capsule for the row
const pillRight = sys.windowWidth - cap.left + 8 // distance from page right edge to pill's right edge
this.setData({ statusBarHeight, rowHeight, pillRight })
```

These data values are written into inline CSS variables on the host:

```html
<view
  class="bar-host {{theme}} {{pinned ? 'pinned' : ''}} {{revealed ? 'revealed' : 'hidden'}}"
  style="--sb-h: {{statusBarHeight}}px; --row-h: {{rowHeight}}px; --pill-right: {{pillRight}}px;"
>
  <view class="bar-spacer"></view>
  <view class="bar-row">
    <view wx:if="{{showCancel}}" class="bar-cancel" bind:tap="onCancel">取消</view>
    <view wx:if="{{mode === 'launcher'}}" class="bar-pill" bind:tap="onTap">
      <dx-icon name="search" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
      <text class="bar-placeholder">{{placeholder}}</text>
    </view>
    <view wx:else class="bar-pill bar-pill-input">
      <slot/>
    </view>
  </view>
</view>
```

**WXSS skeleton:**

```css
.bar-host { position: relative; }
.bar-host.pinned {
  position: fixed; top: 0; left: 0; right: 0; z-index: 90;
  background: var(--bg-page);
  transition: opacity 200ms ease, transform 200ms ease;
}
.bar-host.pinned.hidden { opacity: 0; transform: translateY(-4px); pointer-events: none; }
.bar-spacer { height: var(--sb-h, 20px); }
.bar-row {
  height: var(--row-h, 40px);
  display: flex; align-items: center; gap: 8px;
  padding: 0 var(--pill-right, 102px) 0 12px;
}
.bar-cancel { font-size: 14px; color: var(--text-primary); padding-right: 8px; }
.bar-pill {
  flex: 1; height: 32px; border-radius: 16px;
  background: #ffffff;
  display: flex; align-items: center; gap: 8px; padding: 0 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}
.bar-pill-input { padding: 0 8px 0 12px; }
.bar-placeholder { font-size: 13px; color: #9ca3af; }
.bar-host.dark .bar-pill { background: #2c2c2e; }
.bar-host.dark .bar-placeholder { color: #6b7280; }
```

**`index.json`:**

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

**TS quirks** (per CLAUDE.md / memory): use `Component({...})` with `properties` + `methods`. No `?.` / `??`. Tolerate the existing typed-`this` pattern (don't introduce *new* tsc errors).

### 2. Home page changes (`pages/home/`)

**`home.wxml`** — add the pinned launcher above the existing teal-wrap; keep the in-hero pill but make it tap to the new search page (NO query passed):

```html
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <!-- NEW: pinned launcher (Taobao-reveal) -->
    <dx-search-bar
      theme="{{theme}}"
      pinned
      revealed="{{compactRevealed}}"
      bind:tap="goSearch"
    />

    <!-- Existing teal-wrap unchanged except its inner search-box still taps goSearch -->
    <view class="teal-wrap">
      <view class="status-bar-spacer"></view>
      <!-- ... greeting, badges (unchanged) ... -->
      <view class="search-row">
        <view class="search-box" bind:tap="goSearch">
          <dx-icon name="search" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
          <text class="search-placeholder">搜索课程</text>
        </view>
      </view>
    </view>

    <!-- ... rest of home (combined-card, circle-row, marketing sections) unchanged ... -->
  </view>
</van-config-provider>
```

**`home.json`** — add `"dx-search-bar": "/components/dx-search-bar/index"` to `usingComponents`.

**`home.ts`** — add scroll-driven reveal logic:

```ts
data: {
  // ... existing fields ...
  compactRevealed: false,
  heroBottomPx: 0, // measured once after first paint
}

onReady() {
  // Measure where the in-hero pill ends. The launcher should reveal once
  // the user has scrolled past it (pill bottom <= 0 in viewport coordinates).
  wx.createSelectorQuery()
    .in(this)
    .select('.search-row')
    .boundingClientRect((rect) => {
      if (rect && typeof rect.bottom === 'number') {
        this.setData({ heroBottomPx: rect.bottom })
      }
    })
    .exec()
}

onPageScroll(e) {
  const threshold = this.data.heroBottomPx
  if (threshold <= 0) return
  const shouldReveal = e.scrollTop >= threshold
  if (shouldReveal !== this.data.compactRevealed) {
    this.setData({ compactRevealed: shouldReveal })
  }
}

goSearch() {
  wx.navigateTo({ url: '/pages/games/search/search' })
}
```

The `setData` only fires on threshold crossings (no per-frame thrash). `goSearch` body changes from `wx.navigateTo({ url: '/pages/games/games' })` to `wx.navigateTo({ url: '/pages/games/search/search' })`.

### 3. Courses page changes (`pages/games/`)

**`games.wxml`** — wrap tabs in a sticky container; insert pinned launcher above:

```html
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <!-- NEW: pinned launcher -->
    <dx-search-bar theme="{{theme}}" pinned revealed bind:tap="goSearch" />

    <!-- NEW: sticky tabs row, pinned directly under the launcher -->
    <view class="sticky-tabs">
      <van-tabs
        active="{{activeCategoryId}}"
        bind:click="onCategoryChange"
        color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        background="{{theme === 'dark' ? '#1c1c1e' : '#ffffff'}}"
        title-active-color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        scrollable
      >
        <van-tab wx:for="{{categories}}" wx:key="id" title="{{item.name}}" name="{{item.id}}" />
      </van-tabs>
    </view>

    <!-- fav-bar + game-grid + loading/empty (unchanged) -->
  </view>
</van-config-provider>
```

**`games.wxss`** — replace `padding-top` and add sticky-tabs styles:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  /* status bar + 40px launcher row + ~88rpx tabs row */
  padding-top: calc(var(--status-bar-height, 20px) + 40px + 88rpx);
  padding-bottom: 100rpx;
}
.sticky-tabs {
  position: fixed;
  top: calc(var(--status-bar-height, 20px) + 40px); /* directly under the launcher */
  left: 0; right: 0;
  z-index: 80; /* below the launcher (z-90) so launcher always paints over */
  background: var(--bg-page);
}
/* existing .fav-bar, .game-grid, .game-card, etc. unchanged */
```

**`games.json`** — add `"dx-search-bar": "/components/dx-search-bar/index"` to `usingComponents`.

**`games.ts`** — add the navigation handler. No other changes; the existing `loadCategories` / `loadGames` / `onPullDownRefresh` / `onReachBottom` continue working under the pinned overlays.

```ts
goSearch() {
  wx.navigateTo({ url: '/pages/games/search/search' })
}
```

### 4. New search page (`pages/games/search/`)

**Files:**

- `search.ts`
- `search.wxml`
- `search.wxss`
- `search.json`
- `history.ts` — small helper for wx.storage I/O (extracted to keep `search.ts` focused; per CLAUDE.md "many small files").

**`app.json`** — append `"pages/games/search/search"` to the `pages` array (line 8 area, sibling to `pages/games/favorites/favorites`).

**`search.json`:**

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index",
    "dx-search-bar": "/components/dx-search-bar/index"
  }
}
```

(`van-empty` is intentionally excluded — the empty-results state has custom artwork; reusing `<van-empty>` with description text is also acceptable but the spec uses a custom one to ship the friendly subtitle.)

**`history.ts`:**

```ts
const KEY = 'dx_recent_searches'
const MAX = 10

export function loadHistory(): string[] {
  const raw = wx.getStorageSync(KEY)
  return Array.isArray(raw) ? (raw as string[]).slice(0, MAX) : []
}

export function pushHistory(term: string): string[] {
  const trimmed = term.trim()
  if (!trimmed) return loadHistory()
  const cur = loadHistory().filter(t => t !== trimmed)
  const next = [trimmed, ...cur].slice(0, MAX)
  wx.setStorageSync(KEY, next)
  return next
}

export function clearHistory(): void {
  wx.removeStorageSync(KEY)
}
```

**`search.wxml`** (sketch):

```html
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}">
    <!-- Top row: 取消 + input + clear-X (capsule on the right is system-rendered) -->
    <dx-search-bar
      theme="{{theme}}"
      pinned
      revealed
      mode="input"
      show-cancel
      bind:cancel="onCancel"
    >
      <view class="input-wrap">
        <dx-icon name="search" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
        <input
          class="search-input"
          type="text"
          confirm-type="search"
          focus="{{autoFocus}}"
          value="{{query}}"
          placeholder="搜索课程"
          placeholder-class="search-input-placeholder"
          bindinput="onInput"
          bindconfirm="onSubmit"
        />
        <view wx:if="{{query.length > 0}}" class="clear-btn" bind:tap="onClear">
          <dx-icon name="circle-x" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
        </view>
      </view>
    </dx-search-bar>

    <!-- Body offset by row height -->
    <view class="body" style="padding-top: calc(var(--status-bar-height, 20px) + 40px);">

      <!-- IDLE: suggestions -->
      <block wx:if="{{mode === 'idle'}}">
        <view wx:if="{{recents.length > 0}}" class="section">
          <view class="section-head">
            <text class="section-title">最近搜索</text>
            <view class="section-action" bind:tap="onClearHistoryTap">
              <dx-icon name="trash-2" size="16px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
            </view>
          </view>
          <view class="chip-row">
            <view wx:for="{{recents}}" wx:key="*this" class="chip" data-term="{{item}}" bind:tap="onChipTap">
              <text class="chip-text">{{item}}</text>
            </view>
          </view>
        </view>

        <view wx:if="{{suggestions.length > 0}}" class="section">
          <view class="section-head"><text class="section-title">搜索发现</text></view>
          <view class="chip-row">
            <view wx:for="{{suggestions}}" wx:key="*this" class="chip chip-discover" data-term="{{item}}" bind:tap="onChipTap">
              <text class="chip-text">{{item}}</text>
            </view>
          </view>
        </view>

        <view wx:if="{{recents.length === 0 && suggestions.length === 0 && !suggestionsLoading}}" class="hint">
          <text>搜一搜你感兴趣的课程</text>
        </view>
        <view wx:if="{{suggestionsLoading}}" class="hint"><van-loading size="20px" color="#0d9488" /></view>
      </block>

      <!-- LOADING: results being fetched for the first time -->
      <view wx:if="{{mode === 'loading'}}" class="hint"><van-loading size="20px" color="#0d9488" /></view>

      <!-- RESULTS -->
      <block wx:if="{{mode === 'results'}}">
        <view class="game-grid">
          <view wx:for="{{games}}" wx:key="id" class="game-card" data-id="{{item.id}}" bind:tap="goDetail">
            <view class="game-cover">
              <van-image wx:if="{{item.coverUrl}}" src="{{item.coverUrl}}" width="100%" height="120px" fit="cover" radius="8px 8px 0 0" />
              <view wx:else class="cover-placeholder">
                <dx-icon name="book-open" size="28px" color="#9ca3af" />
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
        <view wx:if="{{loadingMore}}" class="hint"><van-loading size="20px" color="#0d9488" /></view>
      </block>

      <!-- EMPTY (no results for non-empty query) -->
      <view wx:if="{{mode === 'empty'}}" class="empty-state">
        <dx-icon name="search-x" size="48px" color="{{theme === 'dark' ? '#374151' : '#d1d5db'}}" />
        <text class="empty-title">没找到相关课程</text>
        <text class="empty-sub">换个关键词试试</text>
      </view>

    </view>
  </view>
</van-config-provider>
```

**`search.ts` data shape:**

```ts
type Mode = 'idle' | 'loading' | 'results' | 'empty'

interface GameCardData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number; author: string | null
}

data: {
  theme: 'light' as 'light' | 'dark',
  autoFocus: true, // true on onLoad so keyboard pops up
  query: '',
  mode: 'idle' as Mode,
  recents: [] as string[],
  suggestions: [] as string[],
  suggestionsLoading: false,
  games: [] as GameCardData[],
  nextCursor: '',
  hasMore: false,
  loadingMore: false,
}
```

**`search.ts` lifecycle and methods:**

- `onLoad`: read theme; load history into `recents`; fetch `/api/games/search-suggestions` into `suggestions`. On suggestion-fetch failure, silently leave the array empty (do NOT toast — suggestions are non-critical).
- `onShow`: re-read theme.
- `onReachBottom`: only when `mode === 'results' && hasMore && !loadingMore` — call `loadMore()`.
- `onInput(e)`: `setData({ query: e.detail.value })`. Does NOT trigger search. **Important:** if the new value is empty AND `mode === 'results' || mode === 'empty'`, also reset `mode='idle'` so the suggestions reappear.
- `onClear()`: `setData({ query: '', mode: 'idle' })` and re-focus the input (set `autoFocus` to false then true on the next tick to re-trigger).
- `onSubmit(e)`: read `(e.detail.value || this.data.query).trim()`. If empty → no-op (input keeps focus). Otherwise call `runSearch(term)`.
- `onChipTap(e)`: read `data-term` from event currentTarget dataset; call `runSearch(term)`.
- `runSearch(term)`: `setData({ query: term, mode: 'loading' })`; push to history; await `api.get('/api/games?' + qs)` with `q=` and `limit=20`; on success, set `mode = items.length ? 'results' : 'empty'` and store cursor + hasMore. On failure: keep previous mode (or revert to `idle` if previous was `loading`), `wx.showToast({ title: '加载失败', icon: 'none' })`.
- `loadMore()`: `setData({ loadingMore: true })`; same endpoint with current `nextCursor`; concat results; turn off loadingMore.
- `onClearHistoryTap()`: `wx.showModal({ title: '清除最近搜索?', confirmText: '清除', cancelText: '取消' })`. On confirm, `clearHistory()` and `setData({ recents: [] })`.
- `onCancel()`: `wx.navigateBack({ delta: 1 })`. If history is empty (deep-link), fallback to `wx.switchTab({ url: '/pages/games/games' })`.
- `goDetail(e)`: `wx.navigateTo({ url: '/pages/games/detail/detail?id=' + e.currentTarget.dataset['id'] })`.

**`search.wxss`** — local styles for sections, chips, empty state. Reuse the courses-page card classes by copying the `.game-grid`/`.game-card`/etc. CSS verbatim (or extract the shared chunk into a small `wxss` partial later — out of scope for this spec).

### 5. dx-api: new `GET /api/games/search-suggestions`

**Route registration** (`routes/api.go`, inside the existing `/games` group):

```go
router.Prefix("/games").Group(func(games route.Router) {
    games.Get("/", gameController.List)
    games.Get("/search", gameController.Search)
    games.Get("/search-suggestions", gameController.SearchSuggestions) // NEW
})
```

Public (no JWT) — same access level as `/api/games` and `/api/games/search`.

**Controller** (`app/http/controllers/api/game_controller.go`):

```go
// SearchSuggestions returns popular search terms for the dx-mini search page.
// Combines top game names by play count + top categories. Cached 1h.
func (c *GameController) SearchSuggestions(ctx contractshttp.Context) contractshttp.Response {
    suggestions, err := services.GetSearchSuggestions()
    if err != nil {
        return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to load search suggestions")
    }
    return helpers.Success(ctx, suggestions)
}
```

No request validator — no params accepted.

**Service** (`app/services/api/game_service.go`):

```go
const (
    searchSuggestionsCacheKey = "dx:search:suggestions"
    searchSuggestionsCacheTTL = time.Hour
    searchSuggestionsMaxItems = 12
    searchSuggestionsTopGames  = 8
    searchSuggestionsTopCats   = 4
)

// GetSearchSuggestions returns up to 12 search-term chips for the dx-mini
// search page. Cached in Redis for 1h. On cache miss: top published-game
// names by aggregated game_sessions count, plus top game-category names by
// number of published games. Strings deduped; total capped at 12.
func GetSearchSuggestions() ([]string, error) {
    if cached, err := helpers.RedisGet(searchSuggestionsCacheKey); err == nil && cached != "" {
        var out []string
        if jerr := json.Unmarshal([]byte(cached), &out); jerr == nil {
            return out, nil
        }
        // fall through to recompute on bad payload
    }

    type nameRow struct{ Name string `gorm:"column:name"` }

    var topGames []nameRow
    if err := facades.Orm().Query().Raw(`
        SELECT g.name AS name
        FROM games g
        LEFT JOIN game_sessions gs ON gs.game_id = g.id
        WHERE g.status = ? AND g.is_active = TRUE AND g.is_private = FALSE
        GROUP BY g.id, g.name, g.created_at
        ORDER BY COUNT(gs.id) DESC, g.created_at DESC
        LIMIT ?
    `, consts.GameStatusPublished, searchSuggestionsTopGames).Scan(&topGames); err != nil {
        return nil, fmt.Errorf("failed to load top game names: %w", err)
    }

    var topCats []nameRow
    if err := facades.Orm().Query().Raw(`
        SELECT gc.name AS name
        FROM game_categories gc
        INNER JOIN games g ON g.game_category_id = gc.id
        WHERE g.status = ? AND g.is_active = TRUE AND g.is_private = FALSE
        GROUP BY gc.id, gc.name
        ORDER BY COUNT(g.id) DESC
        LIMIT ?
    `, consts.GameStatusPublished, searchSuggestionsTopCats).Scan(&topCats); err != nil {
        return nil, fmt.Errorf("failed to load top categories: %w", err)
    }

    seen := make(map[string]bool, searchSuggestionsMaxItems)
    out := make([]string, 0, searchSuggestionsMaxItems)
    push := func(s string) {
        if s == "" || seen[s] || len(out) >= searchSuggestionsMaxItems {
            return
        }
        seen[s] = true
        out = append(out, s)
    }
    for _, r := range topGames {
        push(r.Name)
    }
    for _, r := range topCats {
        push(r.Name)
    }

    if buf, jerr := json.Marshal(out); jerr == nil {
        _ = helpers.RedisSet(searchSuggestionsCacheKey, string(buf), searchSuggestionsCacheTTL)
    }
    return out, nil
}
```

**Response envelope** for the new endpoint: `{ "code": 0, "message": "ok", "data": ["term1","term2",...] }`.

**Cache invalidation:** the 1h TTL is good enough. No manual invalidation hook this round (admins editing games or categories simply wait up to an hour). If staleness becomes an issue, a follow-up can call `helpers.RedisDel(searchSuggestionsCacheKey)` from the publish/withdraw flows in `services.api.course_game_service.go`.

### 6. New icons to add

Edit `dx-mini/scripts/build-icons.mjs` `ICONS` array:

```js
['x',          'x'],
['circle-x',   'circle-x'],
['trash-2',    'trash-2'],
['search-x',   'search-x'],
['arrow-left', 'arrow-left'],
```

Then run `cd dx-mini && npm run build:icons`. The script's WXML scan ensures any literal `<dx-icon name="...">` we add to WXML is declared; this guards against typos at commit time.

### 7. Theme + safe-area + edge cases

- **Cold open**: home → only the WeChat capsule visible at top; teal hero renders normally; launcher is `revealed=false` (opacity 0). After scrolling past the in-hero pill (`scrollTop >= heroBottomPx`), launcher fades in.
- **Pull-to-refresh on courses**: still works; the launcher and tabs remain visible, the grid below them refreshes (existing `onPullDownRefresh` logic untouched).
- **Tab switch back to courses**: existing `onShow` re-syncs theme + tab-bar; launcher revealed state persists (always-pinned on this page).
- **Search page deep-link** (e.g., from a future scheme URL): `wx.navigateBack` would have no entry; fallback to `wx.switchTab`.
- **Network error** during result fetch: toast `加载失败`, `mode` falls back to `idle`. Suggestions are unaffected.
- **Network error** during suggestion fetch: silent — `suggestions` stays `[]`. The idle screen still shows recents (if any) plus the placeholder hint.
- **iOS dynamic font / safe-area**: row height is computed at runtime from `wx.getMenuButtonBoundingClientRect()`, so iPhone-with-notch and Android with no notch both align correctly.
- **Dark mode**: each component reads `theme` and applies dark-mode pill / chip / text colors. Verified primary palette: light `#0d9488`, dark `#14b8a6`, dark surface `#1c1c1e`.
- **Empty query submit**: `term.trim() === ''` → no-op; input keeps focus; mode unchanged.
- **Whitespace-only query in history**: `pushHistory` trims; empty term is rejected.
- **Duplicate submissions** (same query twice): dedupe via `pushHistory` (filter then unshift).

## Implementation Plan (high level)

The detailed plan with phases, files, and sub-tasks will be produced by `superpowers:writing-plans` after this spec is signed off. The expected ordering is:

1. **dx-api** — add validator-free controller + service + route + service-level unit test for `GetSearchSuggestions` covering cache-miss, cache-hit (mock RedisGet), and dedupe/cap. Run `go test -race ./...`. Public endpoint, no DB migration.
2. **dx-mini icons** — extend `scripts/build-icons.mjs`, run `npm run build:icons`, commit the regenerated `icons.ts`.
3. **dx-mini component** — `components/dx-search-bar/{index.ts, index.wxml, index.wxss, index.json}`.
4. **dx-mini search page** — register in `app.json`; create `pages/games/search/{search.ts, search.wxml, search.wxss, search.json, history.ts}`.
5. **dx-mini home page** — wire `dx-search-bar` + scroll-driven `compactRevealed`; redirect `goSearch` to the new page.
6. **dx-mini courses page** — wire `dx-search-bar` + sticky tabs + `goSearch`.
7. **Manual QA** — smoke test (see below) on iOS + Android in WeChat DevTools 真机调试 / 预览 (per memory `project_wechat_devtools_realdevice_bug.md`, prefer 预览 over 真机调试 if 真机调试 misbehaves).

## Acceptance / Smoke Test Plan

1. Cold-open home page on iPhone with notch: only WeChat capsule visible at top. Scroll down — once the in-hero `搜索课程` pill leaves the viewport, the pinned launcher fades + slides in (≤200ms). Scroll back up — it disappears.
2. Cold-open home page on Android: same behavior, no overlap or gap with capsule. The pill width adjusts to the device's window width.
3. Cold-open courses page: launcher visible immediately at top; tabs row immediately under it; both stay pinned while grid scrolls.
4. Tap any of: home pinned launcher / home in-hero pill / courses pinned launcher → search page navigates in with keyboard up.
5. Search page idle state: 最近搜索 chips render if storage has any (use DevTools Storage panel to verify), 搜索发现 chips render from API. If both are empty, "搜一搜你感兴趣的课程" hint shows.
6. Tap a chip → input fills in, `mode='loading'`, then `mode='results'` (or `mode='empty'`). Suggestions disappear; results render in 2-column grid.
7. Type `vocab` + tap keyboard 搜索 → submit fires once; results render. Scroll to bottom → next page loads via `onReachBottom`.
8. Tap × in input → `mode='idle'`, suggestions reappear.
9. Submit empty / whitespace-only query → no API call (verify via DevTools Network); input keeps focus.
10. Tap trash icon in 最近搜索 → modal `清除最近搜索?` shows; confirm → history cleared, section disappears.
11. Tap a result → `/pages/games/detail/detail?id=<id>`.
12. Tap 取消 → returns to entry page (home or courses).
13. Force kill the network → submit → toast `加载失败`, `mode` returns to idle.
14. Toggle 深色模式 in WeChat → all three pages re-render with dark surfaces; chip and pill contrast meet WCAG AA.
15. dx-api: `curl http://localhost:3001/api/games/search-suggestions` returns `{"code":0,"message":"ok","data":["..."]}`; second call within 1h returns identical payload (cache hit).
16. `go test -race ./...` passes for the new service test.
17. `cd dx-mini && npm run build:icons` succeeds (verifying every literal `<dx-icon name="...">` is declared).

## Notes / Open follow-ups (post-merge)

- If a third page needs the same card layout, promote `.game-grid`/`.game-card`/etc. into a `<game-card>` component and refactor courses + search to use it.
- If admin staleness becomes a real concern for `dx:search:suggestions`, hook `RedisDel` into the publish/withdraw transitions in `services.api.course_game_service.go`.
- Consider a `wx.getMenuButtonBoundingClientRect()` debounce for `wx.onWindowResize` if device-rotation support is ever desired (most mini programs lock portrait, so deferred).
