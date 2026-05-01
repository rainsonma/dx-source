# dx-mini — Promote 消息 to a top-level tab

**Status**: Design  
**Date**: 2026-05-01  
**Scope**: dx-mini (WeChat Mini Program) only — no dx-api, dx-web, deploy, or DB changes

## Goal

Swap 消息 ⇄ 社区 between the home-page hub circles and the bottom tab bar, and promote the existing notice viewer (currently nested at `pages/me/notices/notices` and reached from a home circle plus a bell button on 我的) into a first-class top-level tab. The new tab mirrors the read-only behavior of dx-web `/hall/notices`.

## Non-goals

- No admin (publish / edit / delete) UI on dx-mini. Admin stays on dx-web; the existing dx-api `/api/admin/notices/...` routes are not consumed by dx-mini.
- No real-time push of new notices. The unread state is computed from `lastReadNoticeAt` vs. the latest notice's `createdAt`, refreshed on tab-bar attach.
- No backend, dx-web, or deploy changes.

## Outcome (the visible diff)

| Surface | Before | After |
| --- | --- | --- |
| Tab bar (5 entries, in order) | 首页 / 课程 / 排行榜 / **社区** / 我的 | 首页 / 课程 / 排行榜 / **消息** / 我的 |
| Home page hub circles (in order) | 学习 / 群组 / 打卡 / **消息** / 建议 | 学习 / 群组 / 打卡 / **社区** / 建议 |
| 我的 page top bar | moon/sun toggle + bell shortcut to notices | moon/sun toggle only (bell removed) |
| Notice viewer location | `pages/me/notices/notices` (sub-page reached via `wx.navigateTo`) | `pages/notices/notices` (top-level tab reached via `wx.switchTab`) |
| Notice viewer fidelity | Renders `notice.icon` as plain text (broken — the icon field is a Lucide name string) | Renders via `<dx-icon>` from a curated Lucide allowlist matching dx-web's `iconMap`, falls back to `message-circle-more` |
| Tab-bar unread indicator | None | Red dot on 消息 tab when `lastReadNoticeAt` is null or older than the latest notice's `createdAt`; cleared after `mark-read` resolves |

## File-level change map

### New files

- `dx-mini/miniprogram/pages/notices/notices.{wxml,wxss,ts,json}` — the new tab page.
- `dx-mini/miniprogram/pages/notices/icons.ts` — small helper that maps `notice.icon` to a known dx-icon name with a fallback (mirrors `dx-web/.../notice/helpers/notice-icon.ts`).

### Modified files

- `dx-mini/miniprogram/app.json`
  - `pages`: drop `pages/me/notices/notices`; add `pages/notices/notices`.
  - `tabBar.list`: replace 4th entry (`pages/community/community`) with `pages/notices/notices`.
- `dx-mini/miniprogram/custom-tab-bar/index.{ts,wxml,wxss,json}`
  - 4th tab metadata: `icon: 'bell'`, `text: '消息'`, `path: '/pages/notices/notices'`.
  - Add `unread: boolean` data field.
  - Add `refreshUnread()` method called from `attached()`; fetches `/api/user/profile` + `/api/notices?limit=1` in parallel and toggles the dot.
  - Add `clearUnread()` method invoked by the new tab page after `mark-read` resolves.
  - WXML adds a small absolutely-positioned dot to the 4th icon when `unread` is true.
  - JSON: declare `api`/`auth` as accessible (handled at TS level — no JSON change needed; the file already imports `dx-icon`).
- `dx-mini/miniprogram/pages/home/home.{wxml,ts}`
  - 4th circle becomes 社区 — icon `message-square`, label `社区`, handler `goCommunity()` doing `wx.navigateTo({ url: '/pages/community/community' })`.
  - Remove `goNotices()` (no longer referenced from this page).
- `dx-mini/miniprogram/pages/community/community.ts`
  - Drop the `tabBar.setData({ active: 3, theme })` block in `onShow` — the page is no longer a tab page, there is no tab bar to highlight.
- `dx-mini/miniprogram/pages/me/me.{wxml,ts}`
  - Remove the bell `<dx-icon>` from `.top-bar` and the `goNotices()` method. The `.top-bar` keeps the moon/sun icon.
- `dx-mini/scripts/build-icons.mjs`
  - Append to `ICONS` (logical-name, lucide-static-filename pairs):
    - `message-circle-more` / `message-circle-more`
    - `megaphone` / `megaphone`
    - `rocket` / `rocket`
    - `shield` / `shield`
    - `calendar` / `calendar`
    - `zap` / `zap`
    - `party-popper` / `party-popper`
    - `info` / `info`
    - `alert-triangle` / `triangle-alert` (lucide-static rename)
    - `check-circle-2` / `circle-check` (lucide-static rename — `circle-check-2` doesn't exist in 0.460; rendering close equivalent under the same logical name preserves API compatibility)

### Deleted files

- `dx-mini/miniprogram/pages/me/notices/` (whole directory: `notices.wxml`, `notices.wxss`, `notices.ts`, `notices.json`).

### Untouched

- `dx-api/**`, `dx-web/**`, `deploy/**`, `docs/**` (other than this spec), `dx-mini/miniprogram/miniprogram_npm/**`.

## Detailed design

### 1. Tab bar swap

The 4-th tab entry in both `app.json` `tabBar.list` and `custom-tab-bar/index.ts` `tabs` array is changed from community to notices. Both must agree (WeChat looks at `app.json` for `wx.switchTab` validity; the custom tab bar component renders the visible bar).

The other four entries are unchanged.

### 2. Hub circle swap

The 4th circle in `pages/home/home.wxml` (`circle.c4` block) is changed from notices/bell to community/message-square. The `circle.c4` gradient (`linear-gradient(135deg, #059669, #14b8a6)`) is kept — the gradient palette is decorative, not semantic.

`pages/home/home.ts`:

- Replace `goNotices()` with `goCommunity()`.
- Implementation: `wx.navigateTo({ url: '/pages/community/community' })`.

### 3. The notices tab page

Layout follows the community-tab pattern (status-bar spacer + nav band) and dx-web's `/hall/notices` content shape, adapted to mini idioms.

#### 3.1 `notices.wxml`

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="status-bar-spacer"></view>
    <view class="nav-band">
      <text class="nav-title">消息</text>
    </view>

    <view class="meta-row" wx:if="{{!firstLoading}}">
      <text class="meta-text">共 {{notices.length}} 条通知{{hasMore ? '+' : ''}}</text>
    </view>

    <van-loading wx:if="{{firstLoading}}" size="30px" color="{{primaryColor}}" class="center-loader" />

    <van-empty
      wx:if="{{!firstLoading && notices.length === 0}}"
      description="暂无通知">
      <view slot="image" class="empty-icon">
        <dx-icon name="megaphone" size="40px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
      </view>
    </van-empty>

    <view class="notice-list">
      <view wx:for="{{notices}}" wx:key="id" class="notice-item">
        <view class="notice-icon-bg">
          <dx-icon name="{{item.iconName}}" size="18px" color="{{primaryColor}}" />
        </view>
        <view class="notice-body">
          <text class="notice-title">{{item.title}}</text>
          <text wx:if="{{item.content}}" class="notice-content">{{item.content}}</text>
          <text class="notice-time">{{item.timeText}}</text>
        </view>
      </view>
    </view>

    <view wx:if="{{loadingMore}}" class="load-more">
      <van-loading size="20px" color="{{primaryColor}}" />
    </view>
  </view>
</van-config-provider>
```

#### 3.2 `notices.ts`

```ts
import { api, PaginatedData } from '../../utils/api'
import { isLoggedIn } from '../../utils/auth'
import { formatRelativeDate } from '../../utils/format'
import { resolveNoticeIcon } from './icons'

interface NoticeRaw { id: string; title: string; content: string | null; icon: string | null; createdAt: string }
interface NoticeView extends NoticeRaw { iconName: string; timeText: string }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    firstLoading: true,
    loadingMore: false,
    notices: [] as NoticeView[],
    nextCursor: '',
    hasMore: false,
    markedRead: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({ statusBarHeight: sys.statusBarHeight || 20 })
  },
  onShow() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme })
    if (!isLoggedIn()) { wx.reLaunch({ url: '/pages/login/login' }); return }
    if (this.data.notices.length === 0 && this.data.firstLoading) this.loadNotices(true)
  },
  onPullDownRefresh() { this.loadNotices(true).then(() => wx.stopPullDownRefresh()) },
  onReachBottom() { if (this.data.hasMore && !this.data.loadingMore) this.loadNotices(false) },
  async loadNotices(reset: boolean) {
    if (reset) this.setData({ firstLoading: this.data.notices.length === 0 })
    else this.setData({ loadingMore: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${encodeURIComponent(cursor)}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<NoticeRaw>>(`/api/notices${qs}`)
      const view: NoticeView[] = res.items.map((n) => ({
        ...n,
        iconName: resolveNoticeIcon(n.icon),
        timeText: formatRelativeDate(n.createdAt),
      }))
      this.setData({
        firstLoading: false,
        loadingMore: false,
        notices: reset ? view : [...this.data.notices, ...view],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
      if (reset && !this.data.markedRead) this.markRead()
    } catch {
      this.setData({ firstLoading: false, loadingMore: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  markRead() {
    api.post('/api/notices/mark-read', {}).then(() => {
      this.setData({ markedRead: true })
      const tabBar = this.getTabBar() as (WechatMiniprogram.Component.TrivialInstance & { clearUnread?: () => void }) | null
      if (tabBar && tabBar.clearUnread) tabBar.clearUnread()
    }).catch(() => {})
  },
})
```

#### 3.3 `notices.wxss` (sketch — final values match the existing community/me styling tone)

- `.page-container` — `min-height: 100vh; background: var(--bg-page); padding-bottom: 140rpx;` (matches community).
- `.nav-band` — same teal/title style as community's `.nav-band`.
- `.meta-row` — `padding: 8px 16px 0; .meta-text { font-size: 12px; color: #9ca3af; }`.
- `.notice-list` — `padding: 8px 16px 24px;`.
- `.notice-item` — `display: flex; gap: 12px; padding: 14px; border-radius: 12px; background: var(--bg-card); margin-top: 10px;`.
- `.notice-icon-bg` — 36x36, `border-radius: 10px;`, `background: rgba(13,148,136,0.08)` (dark: `rgba(20,184,166,0.12)`).
- `.notice-body` — flex column, gap 4, title 14px font-weight 600, content 13px line-height 1.5, time 11px color #9ca3af.
- `.center-loader` — same centered style as the existing `me/notices.wxss`.
- `.empty-icon` — flex centered, 40x40 wrapper.
- `.load-more` — `padding: 14px 0; text-align: center;`.

#### 3.4 `notices.json`

```json
{
  "navigationStyle": "custom",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading":         "@vant/weapp/loading/index",
    "van-empty":           "@vant/weapp/empty/index",
    "dx-icon":             "/components/dx-icon/index"
  }
}
```

(Matches `pages/community/community.json` shape — no `navigationBarTitleText` needed because `navigationStyle: "custom"` hides the WeChat default nav bar entirely; the in-page `.nav-band` provides the title.)

### 4. Custom tab bar — unread dot

#### 4.1 `custom-tab-bar/index.ts`

```ts
import { api } from '../utils/api'
import { isLoggedIn } from '../utils/auth'

interface TabItem { icon: string; text: string; path: string }

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    unread: false,
    tabs: [
      { icon: 'home',          text: '首页',   path: '/pages/home/home' },
      { icon: 'book-text',     text: '课程',   path: '/pages/games/games' },
      { icon: 'trophy',        text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'bell',          text: '消息',   path: '/pages/notices/notices' },
      { icon: 'user',          text: '我的',   path: '/pages/me/me' },
    ] as TabItem[],
  },
  lifetimes: {
    attached() {
      const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
      this.setData({ theme: app.globalData.theme })
      this.refreshUnread()
    },
  },
  methods: {
    switchTab(e: WechatMiniprogram.TouchEvent) {
      const path = e.currentTarget.dataset['path'] as string
      wx.switchTab({ url: path })
    },
    async refreshUnread() {
      if (!isLoggedIn()) return
      try {
        const [profile, list] = await Promise.all([
          api.get<{ last_read_notice_at: string | null }>('/api/user/profile'),
          api.get<{ items: { id: string; createdAt: string }[]; nextCursor: string; hasMore: boolean }>('/api/notices?limit=1'),
        ])
        const latest = list.items[0]
        if (!latest) return
        const lastRead = profile.last_read_notice_at
        const unread = !lastRead || new Date(lastRead).getTime() < new Date(latest.createdAt).getTime()
        this.setData({ unread })
      } catch {
        // ignore — leave dot in its previous state
      }
    },
    clearUnread() { this.setData({ unread: false }) },
  },
})
```

#### 4.2 `custom-tab-bar/index.wxml`

```xml
<view class="tab-bar {{theme === 'dark' ? 'dark' : ''}}">
  <view
    wx:for="{{tabs}}"
    wx:key="path"
    class="tab-item {{active === index ? 'active' : ''}}"
    data-path="{{item.path}}"
    bind:tap="switchTab"
  >
    <view class="icon-wrap">
      <dx-icon
        name="{{item.icon}}"
        size="22px"
        color="{{active === index ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : '#9ca3af'}}"
      />
      <view wx:if="{{unread && index === 3}}" class="unread-dot"></view>
    </view>
    <text class="tab-label {{active === index ? 'active-label' : ''}}">{{item.text}}</text>
  </view>
</view>
```

#### 4.3 `custom-tab-bar/index.wxss` additions

```css
.icon-wrap { position: relative; line-height: 0; }
.unread-dot {
  position: absolute;
  top: -2px;
  right: -4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ef4444;
  border: 2px solid #ffffff;
}
.tab-bar.dark .unread-dot { border-color: #1c1c1e; }
```

### 5. Icon allowlist (`pages/notices/icons.ts`)

```ts
const ALLOWED = new Set([
  'message-circle-more', 'swords', 'bell', 'megaphone', 'trophy', 'gift',
  'rocket', 'star', 'shield', 'book-open', 'calendar', 'user-plus',
  'heart', 'zap', 'party-popper', 'info', 'alert-triangle', 'check-circle-2',
  'sparkles', 'crown',
])

export function resolveNoticeIcon(name: string | null | undefined): string {
  if (!name) return 'message-circle-more'
  return ALLOWED.has(name) ? name : 'message-circle-more'
}
```

20 entries, matching dx-web's `iconMap` exactly.

The 10 already declared in `scripts/build-icons.mjs`: `bell`, `crown`, `gift`, `trophy`, `star`, `book-open`, `heart`, `sparkles`, `swords`, `user-plus`.

The 10 to add (logical → lucide-static filename, last two pending verification):

| Logical name | lucide-static filename |
| --- | --- |
| `message-circle-more` | `message-circle-more` |
| `megaphone` | `megaphone` |
| `rocket` | `rocket` |
| `shield` | `shield` |
| `calendar` | `calendar` |
| `zap` | `zap` |
| `party-popper` | `party-popper` |
| `info` | `info` |
| `alert-triangle` | `triangle-alert` |
| `check-circle-2` | `circle-check` |

Both `triangle-alert` and `circle-check` exist in the pinned `lucide-static@0.460.0` (verified with `ls dx-mini/node_modules/lucide-static/icons`). `circle-check-2` does not exist in 0.460; we render the close-equivalent `circle-check` SVG under the logical name `check-circle-2` so dx-web `notice.icon` strings keep working unchanged.

## Error handling

- **List load failure** → `wx.showToast({ title: '加载失败', icon: 'none' })`. Both initial load and reach-bottom load surface the toast; data state stays consistent.
- **Pull-down refresh failure** → toast as above; `stopPullDownRefresh` always called via `.then` chain.
- **Mark-read failure** → silently swallowed via `.catch(() => {})`. Matches dx-web `markNoticesReadAction`. Unread dot persists until next tab-bar attach + new attempt.
- **Tab-bar unread refresh failure** → silently swallowed; dot stays in its previous (likely default `false`) state.
- **Auth missing on tab `onShow`** → `wx.reLaunch` to login (mirrors `community.ts`).
- **Tab-bar before sign-in** → `refreshUnread()` early-returns when `!isLoggedIn()`. Dot stays `false` until first sign-in + tab-bar re-attach.

## Edge cases

- **Tab page entry rules.** WeChat will silently fail on `wx.navigateTo` to a tab page. After this change, no code references `/pages/me/notices/notices` (deleted) or calls `wx.navigateTo` for `/pages/notices/notices`. Validation: a final repo-wide grep before commit.
- **`pages/community/community` no longer a tab.** The existing `getTabBar()` block in `onShow` is guarded by `if (tabBar) ...`; on a non-tab page `getTabBar()` returns `undefined`, so the line is dead code rather than a crash, but it should still be removed for clarity. The community page is now reached via `wx.navigateTo`, exposing the WeChat back-arrow nav bar. All in-page features (compose, like, comment, bookmark, follow, detail navigation) keep working.
- **Custom tab-bar must mirror `app.json` order.** Both define 5 entries with 消息 at index 3.
- **First-time user (no `last_read_notice_at`).** Dot shows on tab; visiting 消息 once clears it.
- **Empty notices DB.** `refreshUnread()` returns early after seeing `list.items[0]` undefined; dot stays `false`. Page shows the empty state.
- **Stored icon name not in allowlist.** Falls back to `message-circle-more` (same UX as dx-web).
- **No `?.` / `??`.** All optional-property accesses in TS use explicit guards or `||` (per the dx-mini rule). WXML uses ternary expressions only.
- **Theme switch** while sitting on the new tab → `onShow` re-applies. The tab bar component's theme is set via `tabBar.setData({ theme })` from the page's `onShow`, same as existing pages.
- **Tab-bar component re-attach behavior.** WeChat re-creates the custom tab-bar component on each tab switch; `attached()` re-runs, so `refreshUnread()` polls fresh on every tab switch. This is desired (cheap, eventually consistent) and the spec relies on it.
- **lucide-static rename surprises.** Build script's static check will throw a descriptive error; fix the right-column filename mapping inline.

## Testing plan (manual, in WeChat DevTools)

1. **Layout flip** — tab bar shows 首页/课程/排行榜/消息/我的; home circles show 学习/群组/打卡/社区/建议. Tap each tab and circle, confirm correct routing.
2. **消息 tab — empty.** Mark all `notices` rows `is_active=false` (or use a fresh DB), open tab → empty state shows megaphone + "暂无通知".
3. **消息 tab — populated.** Seed notices with `icon: "swords"`, `icon: "alert-triangle"`, `icon: null`, `icon: "nonexistent"`. All four render: first two with their own icons, last two with the default `message-circle-more`.
4. **Pagination.** Seed >20 notices; scroll to bottom, confirm next page loads, no duplicates.
5. **Pull-down refresh.** Pull down on the tab page → spinner → list resets.
6. **Mark-read + unread dot.** Fresh user (no `last_read_notice_at`): tab bar shows red dot at index 3 on first attach. Open 消息 tab once → dot disappears immediately after `mark-read` resolves. Switch to home and back → dot stays gone.
7. **Bell removed from 我的.** 我的 page top bar shows only the moon/sun icon.
8. **社区 reachable from home.** Tap 社区 circle → opens community page with WeChat back arrow, no bottom tab bar; compose / like / comment / bookmark / follow all still work.
9. **Theme.** Toggle dark mode on 我的, then visit 消息: nav band, icons, and unread dot all theme correctly.
10. **Lint / typecheck.** `cd dx-mini && npx tsc --noEmit` produces no new errors beyond the pre-existing `Component({methods})` `this`-typing pattern. `cd dx-mini && npm run build:icons` runs clean (or fails with a clear filename error to fix in place).
11. **No broken references.** `grep -r 'goNotices\|/pages/me/notices/notices' dx-mini/miniprogram` returns no hits.

## Out of scope

- dx-web, dx-api, deploy, postgres, nginx untouched.
- Admin publish/edit/delete UI on dx-mini — stays only on dx-web.
- Real-time push of new notices — none. Read-only polling on tab-bar attach is sufficient.
- Caching the unread count across app launches — not needed; one extra request per tab-bar attach is cheap.
- Migrating existing `notice.icon` rows to the allowlist — non-allowlist names already fall back gracefully.
