# dx-mini: Promote 消息 to a Top-Level Tab — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Swap 消息 ⇄ 社区 between the dx-mini home-page hub circles and the bottom tab bar, and rebuild the existing notice viewer as a top-level tab page that mirrors the read-only behavior of dx-web `/hall/notices` (with a tab-bar unread dot).

**Architecture:** Bottom-up. (1) Pin Lucide icons that the new page references into `scripts/build-icons.mjs` and regenerate `icons.ts`; (2) create the new `pages/notices/` tab page (icons helper + WXML/WXSS/TS/JSON); (3) flip `app.json` `pages` + `tabBar.list` and the custom tab bar's `tabs` array; (4) add the unread-dot logic to the custom tab bar; (5) swap the home-page circle from 消息 to 社区; (6) drop the dead `getTabBar` block from `community.ts`; (7) remove the bell shortcut from 我的; (8) delete the orphaned `pages/me/notices/` tree; (9) full smoke test in WeChat DevTools. Steps must run in order — earlier tasks lay foundations later tasks rely on.

**Tech Stack:** WeChat Mini Program native (glass-easel + Skyline), TypeScript strict, Vant Weapp 1.11.x, `<dx-icon>` Lucide SVG renderer, `lucide-static@0.460.0` (pinned), shared dx-api endpoints `/api/notices`, `/api/notices/mark-read`, `/api/user/profile`.

**Spec:** [2026-05-01-dx-mini-messages-tab-design.md](../specs/2026-05-01-dx-mini-messages-tab-design.md)

**Branch:** Work continues on `main`. The spec was committed there in prior turns; this plan and its commits land on `main` too.

---

## Task 1: Add notice-related Lucide icons to the build script

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Regenerated: `dx-mini/miniprogram/components/dx-icon/icons.ts` (auto-generated; do not hand-edit)

- [ ] **Step 1: Append 10 entries to the `ICONS` array**

Open `dx-mini/scripts/build-icons.mjs`. Locate the `ICONS` constant (the closing line is `]` followed by the `lucideDir` declaration). Insert these 10 tuples just before the closing `]`, preserving the existing comment/format style:

```js
  ['message-circle-more', 'message-circle-more'],
  ['megaphone',           'megaphone'],
  ['rocket',              'rocket'],
  ['shield',              'shield'],
  ['calendar',            'calendar'],
  ['zap',                 'zap'],
  ['party-popper',        'party-popper'],
  ['info',                'info'],
  ['alert-triangle',      'triangle-alert'],   // lucide-static renamed alert-triangle -> triangle-alert
  ['check-circle-2',      'circle-check'],     // lucide-static dropped check-circle-2; circle-check is the closest equivalent
```

- [ ] **Step 2: Regenerate `icons.ts`**

From `dx-mini/`:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected output: `Wrote N icons to miniprogram/components/dx-icon/icons.ts.` where N is the previous count + 10. No errors.

If the script throws `lucide-static is missing "X.svg"`, the right-column filename for that row is wrong; check `dx-mini/node_modules/lucide-static/icons/` for the actual filename and fix the tuple. The 10 mappings above were spot-verified against `lucide-static@0.460.0` at spec time.

- [ ] **Step 3: Sanity-check the regenerated `icons.ts`**

```bash
grep -c '"message-circle-more"\|"megaphone"\|"rocket"\|"shield"\|"calendar"\|"zap"\|"party-popper"\|"info"\|"alert-triangle"\|"check-circle-2"' /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/components/dx-icon/icons.ts
```

Expected: `10`.

- [ ] **Step 4: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add \
  dx-mini/scripts/build-icons.mjs \
  dx-mini/miniprogram/components/dx-icon/icons.ts
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "feat(mini): add notice-related Lucide icons to dx-icon registry"
```

---

## Task 2: Add the `resolveNoticeIcon` helper

**Files:**
- Create: `dx-mini/miniprogram/pages/notices/icons.ts`

- [ ] **Step 1: Create the directory**

```bash
mkdir -p /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/notices
```

- [ ] **Step 2: Write `icons.ts`**

Create `dx-mini/miniprogram/pages/notices/icons.ts` with this exact content:

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

Notes:
- The set mirrors dx-web's `iconMap` keys 1:1.
- All 20 names in `ALLOWED` are declared in `scripts/build-icons.mjs` after Task 1; they will resolve to renderable SVGs via `<dx-icon>`.
- No `?.` / `??` per the dx-mini rule (a saved feedback memory): the early `if (!name) return ...` handles null/undefined; `Set.has` accepts only valid strings here.

- [ ] **Step 3: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors. (Pre-existing errors in `Component({methods})` `this`-typing patterns are baseline noise — note the count *before* this task and confirm it didn't grow.)

- [ ] **Step 4: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/pages/notices/icons.ts
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "feat(mini): add resolveNoticeIcon helper for the notices tab"
```

---

## Task 3: Create the notices tab page (WXML, WXSS, TS, JSON)

**Files:**
- Create: `dx-mini/miniprogram/pages/notices/notices.json`
- Create: `dx-mini/miniprogram/pages/notices/notices.wxml`
- Create: `dx-mini/miniprogram/pages/notices/notices.wxss`
- Create: `dx-mini/miniprogram/pages/notices/notices.ts`

- [ ] **Step 1: Write `notices.json`**

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

- [ ] **Step 2: Write `notices.wxml`**

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
      description="暂无通知"
    >
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

- [ ] **Step 3: Write `notices.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 140rpx;
}

.status-bar-spacer {
  height: var(--status-bar-height, 20px);
}

.nav-band {
  padding: 6px 18px 12px;
  background: var(--bg-page);
}

.nav-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary, #1a1a1a);
}

.page-container.dark .nav-title {
  color: #f5f5f5;
}

.meta-row {
  padding: 4px 18px 8px;
}

.meta-text {
  font-size: 12px;
  color: #9ca3af;
}

.center-loader {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
}

.notice-list {
  padding: 4px 14px 24px;
}

.notice-item {
  display: flex;
  gap: 12px;
  padding: 14px;
  border-radius: 12px;
  background: var(--bg-card, #ffffff);
  margin-top: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.page-container.dark .notice-item {
  background: var(--bg-card, #1c1c1e);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);
}

.notice-icon-bg {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(13, 148, 136, 0.08);
}

.page-container.dark .notice-icon-bg {
  background: rgba(20, 184, 166, 0.12);
}

.notice-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.notice-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.page-container.dark .notice-title {
  color: #f5f5f5;
}

.notice-content {
  font-size: 13px;
  line-height: 1.5;
  color: #6b7280;
}

.page-container.dark .notice-content {
  color: #9ca3af;
}

.notice-time {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

.load-more {
  padding: 14px 0;
  text-align: center;
}
```

- [ ] **Step 4: Write `notices.ts`**

```ts
import { api, PaginatedData } from '../../utils/api'
import { isLoggedIn } from '../../utils/auth'
import { formatRelativeDate } from '../../utils/format'
import { resolveNoticeIcon } from './icons'

interface NoticeRaw {
  id: string
  title: string
  content: string | null
  icon: string | null
  createdAt: string
}

interface NoticeView extends NoticeRaw {
  iconName: string
  timeText: string
}

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
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    if (this.data.notices.length === 0 && this.data.firstLoading) {
      this.loadNotices(true)
    }
  },
  onPullDownRefresh() {
    this.loadNotices(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loadingMore) {
      this.loadNotices(false)
    }
  },
  async loadNotices(reset: boolean) {
    if (reset) this.setData({ firstLoading: this.data.notices.length === 0 })
    else this.setData({ loadingMore: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${encodeURIComponent(cursor)}&limit=20` : '?limit=20'
    try {
      const res = await api.get<PaginatedData<NoticeRaw>>(`/api/notices${qs}`)
      const view: NoticeView[] = res.items.map((n) => ({
        id: n.id,
        title: n.title,
        content: n.content,
        icon: n.icon,
        createdAt: n.createdAt,
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
      const tabBar = this.getTabBar() as
        | (WechatMiniprogram.Component.TrivialInstance & { clearUnread?: () => void })
        | null
      if (tabBar && tabBar.clearUnread) tabBar.clearUnread()
    }).catch(() => {})
  },
})
```

- [ ] **Step 5: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors beyond the existing `Component({methods})` baseline. Compare error count to the count noted at end of Task 2.

- [ ] **Step 6: Verify the build script's static WXML scan still passes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: clean run. The new `notices.wxml` references `<dx-icon name="megaphone">` (literal) and `<dx-icon name="{{item.iconName}}">` (dynamic; not scanned by the static check). All literal references must already be in `ICONS` after Task 1.

- [ ] **Step 7: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/pages/notices/
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "feat(mini): scaffold pages/notices/ tab page (read-only list)"
```

---

## Task 4: Register the new tab page in `app.json` (no removals yet)

**Files:**
- Modify: `dx-mini/miniprogram/app.json`

This step adds the new page registration but keeps the old `pages/me/notices/notices` entry around so the app keeps booting between tasks. The actual tab-bar swap is in Step 2 below — both `pages` entries coexist briefly.

- [ ] **Step 1: Add `pages/notices/notices` to the `pages` array**

In `dx-mini/miniprogram/app.json`, find the `"pages": [...]` array. Insert `"pages/notices/notices"` immediately after `"pages/leaderboard/leaderboard"`. Resulting `pages` value:

```json
"pages": [
  "pages/login/login",
  "pages/home/home",
  "pages/games/games",
  "pages/games/detail/detail",
  "pages/games/play/play",
  "pages/games/favorites/favorites",
  "pages/games/search/search",
  "pages/leaderboard/leaderboard",
  "pages/notices/notices",
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
  "pages/community/community",
  "pages/community/detail/detail",
  "pages/me/feedback/feedback"
]
```

- [ ] **Step 2: Replace the 4th `tabBar.list` entry**

In the same file, find `"tabBar": { ..., "list": [...] }`. Replace `{ "pagePath": "pages/community/community" }` with `{ "pagePath": "pages/notices/notices" }`. Resulting `list`:

```json
"list": [
  { "pagePath": "pages/home/home" },
  { "pagePath": "pages/games/games" },
  { "pagePath": "pages/leaderboard/leaderboard" },
  { "pagePath": "pages/notices/notices" },
  { "pagePath": "pages/me/me" }
]
```

- [ ] **Step 3: Validate JSON syntax**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && python3 -c "import json; json.load(open('dx-mini/miniprogram/app.json'))"
```

Expected: silent (valid JSON). Any error prints a line/column.

- [ ] **Step 4: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/app.json
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "feat(mini): register pages/notices/ and swap tab bar slot 3 to it"
```

---

## Task 5: Update the custom tab bar (path/icon/text + unread dot)

**Files:**
- Modify: `dx-mini/miniprogram/custom-tab-bar/index.ts`
- Modify: `dx-mini/miniprogram/custom-tab-bar/index.wxml`
- Modify: `dx-mini/miniprogram/custom-tab-bar/index.wxss`

- [ ] **Step 1: Replace `index.ts` with the new content**

Overwrite `dx-mini/miniprogram/custom-tab-bar/index.ts` with:

```ts
import { api } from '../utils/api'
import { isLoggedIn } from '../utils/auth'

interface TabItem {
  icon: string
  text: string
  path: string
}

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
          api.get<{
            items: { id: string; createdAt: string }[]
            nextCursor: string
            hasMore: boolean
          }>('/api/notices?limit=1'),
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
    clearUnread() {
      this.setData({ unread: false })
    },
  },
})
```

Key changes vs. existing file: the 4th `tabs` entry now points at `/pages/notices/notices` with `icon: 'bell'` and `text: '消息'`; an `unread: false` data field; `attached()` calls `refreshUnread()`; new `refreshUnread` and `clearUnread` methods.

- [ ] **Step 2: Replace `index.wxml` with the dot-aware version**

Overwrite `dx-mini/miniprogram/custom-tab-bar/index.wxml` with:

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

- [ ] **Step 3: Append the dot styles to `index.wxss`**

At the end of `dx-mini/miniprogram/custom-tab-bar/index.wxss`, add:

```css
.icon-wrap {
  position: relative;
  line-height: 0;
}
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
.tab-bar.dark .unread-dot {
  border-color: #1c1c1e;
}
```

- [ ] **Step 4: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors beyond the baseline. The `Component({methods})` `this`-typing pattern noise will already be present from the existing tab bar code; confirm the count matches the post-Task-3 count.

- [ ] **Step 5: Verify the static WXML icon scan**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: clean. The tab bar references `<dx-icon name="{{item.icon}}">` (dynamic, not scanned), and all icons in the data array (`home`, `book-text`, `trophy`, `bell`, `user`) were already declared.

- [ ] **Step 6: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/custom-tab-bar/
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "feat(mini): swap 4th tab to 消息 with unread red dot"
```

---

## Task 6: Swap the home-page hub circle from 消息 to 社区

**Files:**
- Modify: `dx-mini/miniprogram/pages/home/home.wxml`
- Modify: `dx-mini/miniprogram/pages/home/home.ts`

- [ ] **Step 1: Update the 4th `circle-item` in `home.wxml`**

Open `dx-mini/miniprogram/pages/home/home.wxml`. Find this block (around lines 81–84):

```xml
      <view class="circle-item" bind:tap="goNotices">
        <view class="circle c4"><dx-icon name="bell" size="22px" color="#ffffff" /></view>
        <text class="circle-label">消息</text>
      </view>
```

Replace it with:

```xml
      <view class="circle-item" bind:tap="goCommunity">
        <view class="circle c4"><dx-icon name="message-square" size="22px" color="#ffffff" /></view>
        <text class="circle-label">社区</text>
      </view>
```

The `c4` gradient class stays — it's a decorative palette index, not semantic.

- [ ] **Step 2: Replace `goNotices` with `goCommunity` in `home.ts`**

Open `dx-mini/miniprogram/pages/home/home.ts`. Find the line:

```ts
  goNotices() { wx.navigateTo({ url: '/pages/me/notices/notices' }) },
```

Replace it with:

```ts
  goCommunity() { wx.navigateTo({ url: '/pages/community/community' }) },
```

- [ ] **Step 3: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors. (The previous `goNotices` reference was unused outside this file; nothing else references it.)

- [ ] **Step 4: Verify the static WXML icon scan**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: clean. `message-square` is already in `ICONS` (declared from prior tab bar work).

- [ ] **Step 5: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/pages/home/
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "refactor(mini): swap home hub 4th circle from 消息 to 社区"
```

---

## Task 7: Drop the dead tab-bar block from the community page

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/community.ts`

The community page is no longer a tab page. `getTabBar()` returns `undefined` on non-tab pages, and the existing `if (tabBar)` guard already prevents a crash, so this is dead-code cleanup rather than a bug fix.

- [ ] **Step 1: Remove the tab-bar block in `onShow`**

Open `dx-mini/miniprogram/pages/community/community.ts`. Find `onShow` (around lines 28–39):

```ts
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme: app.globalData.theme })
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    if (this.data.posts.length === 0 && !this.data.loading) {
      this.loadFeed(true)
    }
  },
```

Replace it with:

```ts
  onShow() {
    this.setData({ theme: app.globalData.theme })
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    if (this.data.posts.length === 0 && !this.data.loading) {
      this.loadFeed(true)
    }
  },
```

- [ ] **Step 2: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors.

- [ ] **Step 3: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/pages/community/community.ts
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "refactor(mini): drop dead tab-bar setData call from community page"
```

---

## Task 8: Remove the bell shortcut from 我的 page

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.wxml`
- Modify: `dx-mini/miniprogram/pages/me/me.ts`

- [ ] **Step 1: Remove the bell `<dx-icon>` from the top bar in `me.wxml`**

Open `dx-mini/miniprogram/pages/me/me.wxml`. Find the `top-bar` block (around lines 3–16):

```xml
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
```

Replace it with (drops the second `<dx-icon>`, keeps the moon/sun toggle):

```xml
    <view class="top-bar">
      <dx-icon
        name="{{theme === 'dark' ? 'sun' : 'moon'}}"
        size="22px"
        color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        bind:click="toggleTheme"
      />
    </view>
```

- [ ] **Step 2: Remove `goNotices` from `me.ts`**

Open `dx-mini/miniprogram/pages/me/me.ts`. Find this line (around line 69):

```ts
  goNotices() { wx.navigateTo({ url: '/pages/me/notices/notices' }) },
```

Delete it. (Leave the surrounding `goProfileEdit`, `goGroups`, etc. alone.)

- [ ] **Step 3: Type-check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit
```

Expected: no new errors.

- [ ] **Step 4: Verify static WXML icon scan**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: clean. Both `bell` and the moon/sun pair stay declared regardless (used by the tab bar and remaining toggle); we just removed one *use site*.

- [ ] **Step 5: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/pages/me/me.wxml dx-mini/miniprogram/pages/me/me.ts
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "refactor(mini): remove bell shortcut from 我的 top bar"
```

---

## Task 9: Delete the orphaned `pages/me/notices/` tree

**Files:**
- Delete: `dx-mini/miniprogram/pages/me/notices/notices.wxml`
- Delete: `dx-mini/miniprogram/pages/me/notices/notices.wxss`
- Delete: `dx-mini/miniprogram/pages/me/notices/notices.ts`
- Delete: `dx-mini/miniprogram/pages/me/notices/notices.json`
- Modify: `dx-mini/miniprogram/app.json`

- [ ] **Step 1: Confirm no live references remain**

```bash
grep -rn "/pages/me/notices/notices\|goNotices" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/ 2>/dev/null
```

Expected: no output. If anything matches, stop and investigate before deleting — it likely means an earlier task missed a callsite.

- [ ] **Step 2: Drop the entry from `app.json` `pages`**

Open `dx-mini/miniprogram/app.json` and remove the `"pages/me/notices/notices"` line from the `pages` array. The `pages` array now reads:

```json
"pages": [
  "pages/login/login",
  "pages/home/home",
  "pages/games/games",
  "pages/games/detail/detail",
  "pages/games/play/play",
  "pages/games/favorites/favorites",
  "pages/games/search/search",
  "pages/leaderboard/leaderboard",
  "pages/notices/notices",
  "pages/learn/learn",
  "pages/learn/mastered/mastered",
  "pages/learn/unknown/unknown",
  "pages/learn/review/review",
  "pages/me/me",
  "pages/me/profile-edit/profile-edit",
  "pages/me/groups/groups",
  "pages/me/groups-detail/groups-detail",
  "pages/me/invite/invite",
  "pages/me/redeem/redeem",
  "pages/me/purchase/purchase",
  "pages/me/study/study",
  "pages/me/tasks/tasks",
  "pages/community/community",
  "pages/community/detail/detail",
  "pages/me/feedback/feedback"
]
```

- [ ] **Step 3: Validate JSON**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && python3 -c "import json; json.load(open('dx-mini/miniprogram/app.json'))"
```

Expected: silent.

- [ ] **Step 4: Delete the directory**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source rm -r dx-mini/miniprogram/pages/me/notices
```

Expected output: 4 files staged for removal (`notices.wxml`, `notices.wxss`, `notices.ts`, `notices.json`).

- [ ] **Step 5: Type-check & icon-build sanity**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit && npm run build:icons
```

Expected: both clean. Type errors here would mean an external file imports from the deleted directory; investigate before continuing.

- [ ] **Step 6: Commit**

```bash
git -C /Users/rainsen/Programs/Projects/douxue/dx-source add dx-mini/miniprogram/app.json
git -C /Users/rainsen/Programs/Projects/douxue/dx-source commit -m "chore(mini): delete obsolete pages/me/notices/ tree"
```

---

## Task 10: Smoke test in WeChat DevTools

This task is verification-only — no code changes, no commit. The agent should pause here and ask the human to run the manual checks against a running mini program (since DevTools cannot be driven from the CLI).

**Pre-condition:** Open the project at `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini` in WeChat Developer Tools. Confirm: 详情 → 本地设置 → "不校验合法域名..." is checked (dev only). In the DevTools console, point the dev API at a running dx-api instance:

```js
require('./utils/config').setDevApiBaseUrl('http://<lan-ip>')
```

Refresh the simulator. Sign in.

- [ ] **Check 1 — Layout flip**

Tab bar shows in order: 首页 / 课程 / 排行榜 / 消息 / 我的. Home circles show: 学习 / 群组 / 打卡 / 社区 / 建议. Tap each tab → correct page loads. Tap each home circle → correct page loads.

- [ ] **Check 2 — 消息 tab, populated**

Seed at least 3 notices via dx-web `/hall/notices` (admin user `rainson`) with a mix of icons:
- one with `icon: "swords"`
- one with `icon: "alert-triangle"`
- one with `icon: null`
- one with `icon: "nonexistent-name-foo"` (fallback test)

Open the 消息 tab. All 4 render. The first two show their own icons; the last two show the default `message-circle-more` icon. Title, content (when present), and relative time all appear. The `共 N 条通知` row shows the correct count.

- [ ] **Check 3 — 消息 tab, empty**

Either temporarily soft-delete every notice in the DB (`UPDATE notices SET is_active = false`) or sign in as a fresh user with no visible notices. Open 消息 tab. Empty state shows: megaphone icon + "暂无通知". No `共 0 条` row (because `firstLoading` flips to false but the notices array is empty — confirm the `meta-row` does not render).

- [ ] **Check 4 — Pagination**

Seed >20 notices. Open 消息. Scroll to bottom. The footer spinner appears, then a second page loads with no duplicates. Continue until `hasMore` becomes false; confirm the spinner stops appearing.

- [ ] **Check 5 — Pull-down refresh**

On the 消息 tab, drag down. Spinner appears, then list resets to the freshest 20.

- [ ] **Check 6 — Mark-read + unread dot**

Sign in as a fresh user (no `last_read_notice_at`) with at least one active notice in DB. After sign-in, the bottom 消息 tab icon shows a red dot at top-right of the bell. Tap into 消息. The dot disappears within ~500 ms (after `mark-read` resolves). Switch to 首页, then back to 消息 — the dot stays gone.

Restart the mini program after publishing a new notice from dx-web. On the next launch the dot re-appears (`refreshUnread` re-runs on each tab-bar `attached`, sees `lastReadNoticeAt < latest.createdAt`, sets the dot back). Within a single live session, real-time dot re-appearance after a *new* notice is intentionally out of scope — pull-down on 消息 will load it; the dot stays gone until next launch.

- [ ] **Check 7 — Bell removed from 我的**

Open 我的. The top bar shows only the moon/sun icon, aligned right. No bell. No layout shift.

- [ ] **Check 8 — 社区 reachable from home**

On home, tap the 社区 circle. The community page opens with a WeChat-default back arrow in its nav bar (no bottom tab bar). All in-page actions still work: open the composer, post, like a post, comment, reply, bookmark, follow another user, open the detail page, edit/delete your own post.

- [ ] **Check 9 — Theme**

Open 我的 → tap moon/sun → switch to dark mode. Visit 消息: nav title, item icons, item background, and the unread-dot border color all theme correctly. Switch back to light. Re-verify.

- [ ] **Check 10 — Final lint sweep**

From the CLI:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx tsc --noEmit && npm run build:icons
grep -rn "/pages/me/notices/notices\|goNotices" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/ 2>/dev/null
```

Expected: tsc clean (baseline error count unchanged), build-icons clean, grep silent.

If any check fails, stop and diagnose before claiming completion.

---

## Self-review checklist (run before declaring the plan done)

- **Spec coverage** — every modified/created/deleted file listed in the spec's "File-level change map" appears in this plan: `app.json` (T4, T9), `custom-tab-bar/*` (T5), `home.{wxml,ts}` (T6), `community.ts` (T7), `me.{wxml,ts}` (T8), `pages/me/notices/*` deleted (T9), `pages/notices/*` created (T2, T3), `scripts/build-icons.mjs` + auto-generated `icons.ts` (T1). ✓
- **Placeholder scan** — no "TBD", "TODO", "implement later", or "similar to Task N" references. Every code block is complete. ✓
- **Ordering safety** — between any two consecutive tasks, the app boots: T4 keeps the old `me/notices` registered while adding the new one; T9 deletes the old tree only after T8 removes the last in-app reference. ✓
- **Type & name consistency** — `clearUnread()`, `refreshUnread()`, and `markedRead` are named identically across the tab bar (T5), notices.ts (T3), and tests (T10). `resolveNoticeIcon` signature matches between T2 and T3. The `iconName` / `timeText` fields on `NoticeView` are consumed in T3's WXML exactly as produced by T3's TS. ✓
- **Lucide rename mappings** — verified against `lucide-static@0.460.0`: `triangle-alert.svg` and `circle-check.svg` exist; `message-circle-more.svg`, `megaphone.svg`, `rocket.svg`, `shield.svg`, `calendar.svg`, `zap.svg`, `party-popper.svg`, `info.svg` exist directly. ✓
