# dx-mini 社区 Tab — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dx-mini stub at `pages/me/community/community` with a full community feature mirroring dx-web `/hall/community` (4 tabs, post composer with image upload, like / comment / reply / bookmark / follow, tap-to-detail page with sticky comment input). Then in Phase 2, fix the `bookmarks`/`bookmarked` tab-name bug on dx-web and add edit/delete UI on both clients.

**Architecture:** Two new top-level pages (`pages/community/community` tab + `pages/community/detail/detail` sub-page) with four local components (`post-card`, `post-block`, `comment-item`, `composer-popup`), a small `utils/avatar.ts` shared helper, and ~8 new Lucide icons. dx-api endpoints already exist; no server changes for Phase 1.

**Tech Stack:** WeChat mini program (Glass-easel), TypeScript strict mode, Vant Weapp 1.11.x, Lucide SVGs via `<dx-icon>`, custom tab bar, `wx.chooseMedia` + `wx.uploadFile` for image uploads, cursor-paginated feeds via `onReachBottom`.

**Spec:** [docs/superpowers/specs/2026-04-27-dx-mini-community-tab-design.md](../specs/2026-04-27-dx-mini-community-tab-design.md)

**dx-mini has no unit test framework.** Verification is `tsc --noEmit` (or DevTools compile), the icon-build script's static WXML scanner, and manual smoke in WeChat DevTools simulator.

---

## Phase 1 — dx-mini community feature

### Task 1: Add new Lucide icons + verify build

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Generated (do not hand-edit): `dx-mini/miniprogram/components/dx-icon/icons.ts`

- [ ] **Step 1: Append the new icons to the ICONS array**

In `dx-mini/scripts/build-icons.mjs`, find the closing `]` of the ICONS array (currently after `['search-x', 'search-x'],` near line 57) and insert these rows just before it (preserve the existing alignment style):

```js
  ['heart',           'heart'],
  ['message-circle',  'message-circle'],
  ['bookmark',        'bookmark'],
  ['user-plus',       'user-plus'],
  ['user-check',      'user-check'],
  ['send',            'send'],
  ['image',           'image'],
  ['plus',            'plus'],
```

- [ ] **Step 2: Run the icon build**

```
cd dx-mini && npm run build:icons
```

Expected: `Wrote N icons to miniprogram/components/dx-icon/icons.ts.` where N is the prior count + 8. No errors. (Errors here would mean lucide-static is missing one of these names — check the package version in `dx-mini/package.json`.)

- [ ] **Step 3: Spot-check the generated file**

Run:
```
grep -E '^  "(heart|send|user-plus|plus)"' dx-mini/miniprogram/components/dx-icon/icons.ts
```
Expected: 4 matching lines (each is one of the new names with its inline SVG payload).

- [ ] **Step 4: Commit**

```
git add dx-mini/scripts/build-icons.mjs dx-mini/miniprogram/components/dx-icon/icons.ts
git commit -m "feat(mini): register Lucide icons for community feature"
```

---

### Task 2: Add `utils/avatar.ts` (shared deterministic-color helper)

**Files:**
- Create: `dx-mini/miniprogram/utils/avatar.ts`

- [ ] **Step 1: Create the helper**

Write `dx-mini/miniprogram/utils/avatar.ts` (port of `dx-web/src/lib/avatar.ts`):

```ts
const avatarColors = [
  '#ef4444', '#f97316', '#f59e0b', '#eab308', '#84cc16',
  '#22c55e', '#14b8a6', '#06b6d4', '#0ea5e9', '#3b82f6',
  '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
]

export function getAvatarColor(id: string): string {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0
  }
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

export function getAvatarLetter(nickname: string | null | undefined): string {
  if (!nickname) return '?'
  return nickname.charAt(0)
}
```

- [ ] **Step 2: Commit**

```
git add dx-mini/miniprogram/utils/avatar.ts
git commit -m "feat(mini): add deterministic avatar color helper"
```

---

### Task 3: Add community types

**Files:**
- Create: `dx-mini/miniprogram/pages/community/types.ts`

- [ ] **Step 1: Create the types file**

Write `dx-mini/miniprogram/pages/community/types.ts`:

```ts
export interface PostAuthor {
  id: string
  nickname: string
  avatar_url: string | null
}

export interface Post {
  id: string
  content: string
  image_url: string | null
  tags: string[]
  like_count: number
  comment_count: number
  is_liked: boolean
  is_bookmarked: boolean
  author: PostAuthor
  created_at: string
}

export interface Comment {
  id: string
  content: string
  author: PostAuthor
  parent_id: string | null
  created_at: string
}

export interface CommentWithReplies {
  comment: Comment
  replies: Comment[]
}

export type FeedTab = 'latest' | 'hot' | 'following' | 'bookmarked'

export const FEED_TABS: { name: FeedTab; title: string }[] = [
  { name: 'latest',     title: '最新' },
  { name: 'hot',        title: '热门' },
  { name: 'following',  title: '关注' },
  { name: 'bookmarked', title: '收藏' },
]
```

- [ ] **Step 2: Commit**

```
git add dx-mini/miniprogram/pages/community/types.ts
git commit -m "feat(mini): add community types and tab constants"
```

---

### Task 4: Routing — register new pages, swap tab bar, scaffold empty pages

**Goal:** Move the 社区 tab from the stub at `pages/me/community/community` to a fresh top-level `pages/community/community`, scaffold a sub-page `pages/community/detail/detail`, and delete the old stub. After this task the simulator should compile and tab 3 should show an empty placeholder; nothing else changes yet.

**Files:**
- Modify: `dx-mini/miniprogram/app.json`
- Modify: `dx-mini/miniprogram/custom-tab-bar/index.ts`
- Create: `dx-mini/miniprogram/pages/community/community.ts`
- Create: `dx-mini/miniprogram/pages/community/community.wxml`
- Create: `dx-mini/miniprogram/pages/community/community.wxss`
- Create: `dx-mini/miniprogram/pages/community/community.json`
- Create: `dx-mini/miniprogram/pages/community/detail/detail.ts`
- Create: `dx-mini/miniprogram/pages/community/detail/detail.wxml`
- Create: `dx-mini/miniprogram/pages/community/detail/detail.wxss`
- Create: `dx-mini/miniprogram/pages/community/detail/detail.json`
- Delete: `dx-mini/miniprogram/pages/me/community/` (entire folder)

- [ ] **Step 1: Scaffold `community.json`**

```json
{
  "navigationStyle": "custom",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Scaffold `community.ts`**

```ts
const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
    })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
    const tabBar = this.getTabBar() as WechatMiniprogram.Component.TrivialInstance | null
    if (tabBar) tabBar.setData({ active: 3, theme: app.globalData.theme })
  },
})
```

- [ ] **Step 3: Scaffold `community.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="status-bar-spacer"></view>
    <view class="nav-band">
      <text class="nav-title">社区</text>
    </view>
    <view class="placeholder">scaffolded — feed coming in Task 5</view>
  </view>
</van-config-provider>
```

- [ ] **Step 4: Scaffold `community.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: calc(56px + env(safe-area-inset-bottom) + 32rpx);
}
.status-bar-spacer {
  height: var(--status-bar-height, 20px);
}
.nav-band {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.nav-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}
.placeholder {
  padding: 32px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}
```

- [ ] **Step 5: Scaffold `detail.json`**

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 6: Scaffold `detail.ts`**

```ts
const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    postId: '',
  },
  onLoad(query: Record<string, string>) {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
      postId: query.id || '',
    })
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
  },
  goBack() {
    wx.navigateBack()
  },
})
```

- [ ] **Step 7: Scaffold `detail.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="status-bar-spacer"></view>
    <view class="nav-band">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">帖子</text>
      <view class="back-btn"></view>
    </view>
    <view class="placeholder">postId={{postId}} — coming in Task 13</view>
  </view>
</van-config-provider>
```

- [ ] **Step 8: Scaffold `detail.wxss`**

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
}
.status-bar-spacer {
  height: var(--status-bar-height, 20px);
}
.nav-band {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
}
.back-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}
.nav-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}
.placeholder {
  padding: 32px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}
```

- [ ] **Step 9: Update `app.json` — register pages, remove stub**

Open `dx-mini/miniprogram/app.json`. Replace the line `"pages/me/community/community",` with the two new pages. Replace the tabBar entry that points at `pages/me/community/community` with the new path. The `pages` array should now include both new entries; the deleted line should not appear.

After edit, `pages` array contains (at minimum) — leave existing order otherwise:
```
"pages/community/community",
"pages/community/detail/detail",
```
…and the original `"pages/me/community/community",` is **removed**.

`tabBar.list` 4th item changes from:
```json
{ "pagePath": "pages/me/community/community" },
```
to:
```json
{ "pagePath": "pages/community/community" },
```

- [ ] **Step 10: Update `custom-tab-bar/index.ts` — point tab 3 at the new path**

In `dx-mini/miniprogram/custom-tab-bar/index.ts`, change the 4th tab's `path` from `'/pages/me/community/community'` to `'/pages/community/community'`. Leave `icon`, `text`, and order unchanged.

- [ ] **Step 11: Delete the old stub folder**

```
rm -rf dx-mini/miniprogram/pages/me/community
```

- [ ] **Step 12: Verify in WeChat DevTools**

Open the project in WeChat DevTools. Compile (默认 自动). Expected: zero TS errors, zero WXML errors. Tap tab 3 (社区) → see the "scaffolded" placeholder. Other tabs unaffected.

- [ ] **Step 13: Commit**

```
git add -A
git commit -m "refactor(mini): move 社区 tab from pages/me/community to top-level pages/community"
```

---

### Task 5: Profile-edit avatar upload — fix missing `role` form field

**Goal:** Adjacent bug — `wx.uploadFile` to `/api/uploads/images` requires a `role` form field that the existing avatar upload omits. The server validates against an allowlist; without `role` the upload is rejected. Add `formData: { role: 'user-avatar' }` so the existing flow validates server-side.

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/profile-edit/profile-edit.ts:61-65`

- [ ] **Step 1: Add `formData` to the `wx.uploadFile` call**

Replace the existing call (around lines 61–66, the `wx.uploadFile({...})` block inside `chooseAvatar`) with the version below — the only change is the new `formData` line:

```ts
        wx.uploadFile({
          url: config.apiBaseUrl + '/api/uploads/images',
          filePath: tempPath,
          name: 'file',
          formData: { role: 'user-avatar' },
          header: { Authorization: `Bearer ${getToken()}` },
          success: (uploadRes) => {
```

(The body of `success` and `fail` callbacks is unchanged.)

- [ ] **Step 2: Smoke test in DevTools**

Open the project, log in, navigate to 我的 → 编辑资料 → tap avatar → 选择图片 → upload. Expected: `头像已更新` toast, the new avatar replaces the old one in the page header. Without the fix the server returned a 4xx with envelope code 40000.

- [ ] **Step 3: Commit**

```
git add dx-mini/miniprogram/pages/me/profile-edit/profile-edit.ts
git commit -m "fix(mini): pass required 'role' field on profile avatar upload"
```

---

### Task 6: Feed page — load latest posts, render placeholder list

**Goal:** Wire `loadFeed` for the `latest` tab, render an empty/loading/error state. No tabs, no post-card yet — list cells render a single line of text just so we know the data flows.

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/community.ts`
- Modify: `dx-mini/miniprogram/pages/community/community.wxml`
- Modify: `dx-mini/miniprogram/pages/community/community.wxss`

- [ ] **Step 1: Replace `community.ts` with the data-aware version**

```ts
import { api, PaginatedData } from '../../utils/api'
import { isLoggedIn } from '../../utils/auth'
import type { Post, FeedTab } from './types'

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    tab: 'latest' as FeedTab,
    posts: [] as Post[],
    nextCursor: '',
    hasMore: false,
    loading: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
    })
  },
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
  onPullDownRefresh() {
    this.loadFeed(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadFeed(false)
    }
  },
  async loadFeed(reset: boolean) {
    if (this.data.loading) return
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const parts = ['limit=20', `tab=${this.data.tab}`]
    if (cursor) parts.push(`cursor=${encodeURIComponent(cursor)}`)
    try {
      const res = await api.get<PaginatedData<Post>>(`/api/posts?${parts.join('&')}`)
      this.setData({
        posts: reset ? res.items : [...this.data.posts, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loading: false,
      })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: (err as Error).message || '加载失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 2: Replace `community.wxml` with the placeholder list**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="status-bar-spacer"></view>
    <view class="nav-band">
      <text class="nav-title">社区</text>
    </view>

    <view class="post-list">
      <view
        wx:for="{{posts}}"
        wx:key="id"
        class="post-row"
      >
        <text class="post-author">{{item.author.nickname}}</text>
        <text class="post-content">{{item.content}}</text>
      </view>
    </view>

    <van-loading wx:if="{{loading}}" size="24px" color="#0d9488" class="center-loader" />
    <van-empty wx:if="{{!loading && posts.length === 0}}" description="暂无帖子" />
  </view>
</van-config-provider>
```

- [ ] **Step 3: Append placeholder list styles to `community.wxss`**

Append to `community.wxss` (keep the existing rules):

```css
.post-list {
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.post-row {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.post-author {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}
.post-content {
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 6;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.center-loader {
  display: flex;
  justify-content: center;
  padding: 24px;
}

/* Remove the .placeholder rule from earlier scaffolding — no longer used. */
```

Also delete the now-unused `.placeholder` rule introduced in Task 4.

- [ ] **Step 4: Compile + smoke**

In WeChat DevTools: compile clean, log in, tap 社区 tab. Expected: posts list shows author + content for each post, or `暂无帖子` if the DB is empty. Pull-to-refresh works (visible spinner). Scroll → load more works when there are >20 posts.

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml dx-mini/miniprogram/pages/community/community.wxss
git commit -m "feat(mini): load latest posts on community tab with cursor pagination"
```

---

### Task 7: post-card component — display only

**Goal:** Build a reusable `post-card` component that renders avatar, nickname, time, content (line-clamped), image, tags, and a stub action row. No interactions yet — that's Task 9 / 10.

**Files:**
- Create: `dx-mini/miniprogram/pages/community/components/post-card/index.ts`
- Create: `dx-mini/miniprogram/pages/community/components/post-card/index.wxml`
- Create: `dx-mini/miniprogram/pages/community/components/post-card/index.wxss`
- Create: `dx-mini/miniprogram/pages/community/components/post-card/index.json`

- [ ] **Step 1: Create `index.json`**

```json
{
  "component": true,
  "usingComponents": {
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Create `index.ts`**

```ts
import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatRelativeDate, formatNumber } from '../../../../utils/format'
import { config } from '../../../../utils/config'
import type { Post } from '../../types'

Component({
  options: { addGlobalClass: true },
  properties: {
    post: { type: Object, value: null },
    theme: { type: String, value: 'light' },
  },
  data: {
    avatarColor: '#999',
    avatarLetter: '?',
    timeText: '',
    likeText: '',
    commentText: '',
    imageAbsoluteUrl: '',
  },
  observers: {
    post(post: Post | null) {
      if (!post) return
      this.setData({
        avatarColor: getAvatarColor(post.author.id),
        avatarLetter: getAvatarLetter(post.author.nickname),
        timeText: formatRelativeDate(post.created_at),
        likeText: post.like_count > 0 ? formatNumber(post.like_count) : '',
        commentText: post.comment_count > 0 ? formatNumber(post.comment_count) : '',
        imageAbsoluteUrl: post.image_url ? config.apiBaseUrl + post.image_url : '',
      })
    },
  },
  methods: {
    onCardTap() {
      this.triggerEvent('opendetail', { id: (this.data as { post: Post }).post.id })
    },
    onLikeTap(e: WechatMiniprogram.TouchEvent) {
      // catchtap on the WXML side blocks bubbling; this is a no-op stub for Task 9.
      void e
    },
    onCommentTap() {
      this.triggerEvent('opendetail', { id: (this.data as { post: Post }).post.id })
    },
    onBookmarkTap(e: WechatMiniprogram.TouchEvent) {
      // stub for Task 9
      void e
    },
    onFollowTap(e: WechatMiniprogram.TouchEvent) {
      // stub for Task 10
      void e
    },
  },
})
```

- [ ] **Step 3: Create `index.wxml`**

```xml
<view class="post-card" bind:tap="onCardTap">
  <view class="header">
    <view class="author">
      <view class="avatar" wx:if="{{!post.author.avatar_url}}" style="background: {{avatarColor}};">{{avatarLetter}}</view>
      <van-image
        wx:if="{{post.author.avatar_url}}"
        src="{{post.author.avatar_url}}"
        width="44px"
        height="44px"
        round
        fit="cover"
      />
      <view class="meta">
        <text class="nickname">{{post.author.nickname}}</text>
        <text class="time">{{timeText}}</text>
      </view>
    </view>
    <view class="follow-stub" catch:tap="onFollowTap">
      <dx-icon name="user-plus" size="14px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
      <text>关注</text>
    </view>
  </view>

  <text class="content">{{post.content}}</text>

  <view class="image-wrap" wx:if="{{imageAbsoluteUrl}}">
    <van-image src="{{imageAbsoluteUrl}}" width="100%" height="220px" fit="cover" radius="10" />
  </view>

  <view class="tags" wx:if="{{post.tags.length > 0}}">
    <text wx:for="{{post.tags}}" wx:key="*this" class="tag">#{{item}}</text>
  </view>

  <view class="divider"></view>

  <view class="actions">
    <view class="action" catch:tap="onLikeTap">
      <dx-icon
        name="heart"
        size="18px"
        color="{{post.is_liked ? '#ef4444' : (theme === 'dark' ? '#9ca3af' : '#6b7280')}}"
        custom-style="{{post.is_liked ? 'fill: #ef4444;' : ''}}"
      />
      <text class="action-text">{{likeText}}</text>
    </view>
    <view class="action" catch:tap="onCommentTap">
      <dx-icon name="message-circle" size="18px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
      <text class="action-text">{{commentText}}</text>
    </view>
    <view class="action" catch:tap="onBookmarkTap">
      <dx-icon
        name="bookmark"
        size="18px"
        color="{{post.is_bookmarked ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : (theme === 'dark' ? '#9ca3af' : '#6b7280')}}"
        custom-style="{{post.is_bookmarked ? (theme === 'dark' ? 'fill: #14b8a6;' : 'fill: #0d9488;') : ''}}"
      />
    </view>
  </view>
</view>
```

- [ ] **Step 4: Create `index.wxss`**

```css
.post-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
  margin: 0 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.author {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
}
.meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nickname {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.time {
  font-size: 12px;
  color: var(--text-secondary);
}
.follow-stub {
  display: flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--primary);
  color: var(--primary);
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
}
.content {
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 6;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.image-wrap {
  border-radius: 10px;
  overflow: hidden;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.tag {
  background: var(--primary-light);
  color: var(--primary);
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}
.divider {
  height: 1px;
  background: var(--border-color);
}
.actions {
  display: flex;
  align-items: center;
  justify-content: space-around;
}
.action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 16px;
}
.action-text {
  font-size: 13px;
  color: var(--text-secondary);
}
```

- [ ] **Step 5: Wire post-card into the feed page**

In `dx-mini/miniprogram/pages/community/community.json`, add:
```json
"post-card": "./components/post-card/index"
```
to `usingComponents`.

In `community.wxml`, replace the `.post-list` block with:

```xml
<view class="post-list">
  <post-card
    wx:for="{{posts}}"
    wx:key="id"
    post="{{item}}"
    theme="{{theme}}"
    bind:opendetail="onOpenDetail"
  />
</view>
```

In `community.ts`, add the handler inside the methods (alongside `loadFeed`):

```ts
  onOpenDetail(e: WechatMiniprogram.CustomEvent) {
    const id = (e.detail as { id: string }).id
    wx.navigateTo({ url: `/pages/community/detail/detail?id=${id}` })
  },
```

In `community.wxss`, delete the `.post-row`, `.post-author`, and `.post-content` rules (post-card now owns its own styles).

- [ ] **Step 6: Compile + smoke**

In DevTools: compile clean. Tap 社区 → see styled post cards (avatar, nickname, time, content, optional image, tags, actions). Tap a card → navigates to detail (still scaffolded). Like / bookmark / follow icons render but are no-ops.

- [ ] **Step 7: Commit**

```
git add dx-mini/miniprogram/pages/community/components/post-card dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml dx-mini/miniprogram/pages/community/community.wxss dx-mini/miniprogram/pages/community/community.json
git commit -m "feat(mini): post-card component renders posts in community feed"
```

---

### Task 8: Tab switching (4 feed tabs)

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/community.ts`
- Modify: `dx-mini/miniprogram/pages/community/community.wxml`
- Modify: `dx-mini/miniprogram/pages/community/community.wxss`

- [ ] **Step 1: Add tab handler in `community.ts`**

Inside `Page({ … })`, alongside existing methods:

```ts
  onTabChange(e: WechatMiniprogram.TouchEvent) {
    const name = (e.detail as { name: string }).name
    this.setData({
      tab: name as FeedTab,
      posts: [],
      nextCursor: '',
      hasMore: false,
    })
    this.loadFeed(true)
  },
```

Also add the tab list to the page data block — append after `tab` initialization:

```ts
import { FEED_TABS, type FeedTab } from './types'
```
(replace any earlier `import type { FeedTab }` with the combined import).

And inside `data:`:
```ts
    feedTabs: FEED_TABS,
```

- [ ] **Step 2: Add the tab pills in `community.wxml`**

Just below the `nav-band` view, before `.post-list`:

```xml
<view class="tabs-row">
  <van-tabs
    active="{{tab}}"
    bind:click="onTabChange"
    color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
    swipeable
  >
    <van-tab
      wx:for="{{feedTabs}}"
      wx:key="name"
      title="{{item.title}}"
      name="{{item.name}}"
    />
  </van-tabs>
</view>
```

- [ ] **Step 3: Style the tabs**

Append to `community.wxss`:

```css
.tabs-row {
  background: var(--bg-page);
  --tabs-nav-background-color: transparent;
}
```

- [ ] **Step 4: Smoke**

In DevTools: tap each tab → spinner → list refreshes. `最新` shows newest, `热门` shows hot-ranked, `关注` shows posts from followed users (empty state if you follow no one), `收藏` shows your bookmarks (empty if none).

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml dx-mini/miniprogram/pages/community/community.wxss
git commit -m "feat(mini): wire 4 community feed tabs (最新/热门/关注/收藏)"
```

---

### Task 9: Post-card like + bookmark interactions (optimistic)

**Goal:** Like and bookmark toggles flip local state immediately, call API, roll back + toast on error. The state is updated on the page-level `posts` array via component → page event.

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/components/post-card/index.ts`
- Modify: `dx-mini/miniprogram/pages/community/components/post-card/index.wxml` (no change if already calling `onLikeTap` / `onBookmarkTap`)
- Modify: `dx-mini/miniprogram/pages/community/community.ts`

- [ ] **Step 1: Replace stub handlers in `post-card/index.ts`**

Replace `onLikeTap` and `onBookmarkTap` with versions that emit events upward (the page handles the API to keep components dumb):

```ts
    onLikeTap() {
      this.triggerEvent('toggle-like', { id: (this.data as { post: Post }).post.id })
    },
    onBookmarkTap() {
      this.triggerEvent('toggle-bookmark', { id: (this.data as { post: Post }).post.id })
    },
```

- [ ] **Step 2: Page-level handlers in `community.ts`**

Add to the methods block:

```ts
  async onToggleLike(e: WechatMiniprogram.CustomEvent) {
    const id = (e.detail as { id: string }).id
    const idx = this.data.posts.findIndex((p) => p.id === id)
    if (idx < 0) return
    const before = this.data.posts[idx]
    const optimistic: Post = {
      ...before,
      is_liked: !before.is_liked,
      like_count: before.is_liked ? Math.max(before.like_count - 1, 0) : before.like_count + 1,
    }
    this.patchPost(idx, optimistic)
    try {
      const res = await api.post<{ liked: boolean; like_count: number }>(`/api/posts/${id}/like`, {})
      this.patchPost(idx, { ...optimistic, is_liked: res.liked, like_count: res.like_count })
    } catch (err) {
      this.patchPost(idx, before)
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  async onToggleBookmark(e: WechatMiniprogram.CustomEvent) {
    const id = (e.detail as { id: string }).id
    const idx = this.data.posts.findIndex((p) => p.id === id)
    if (idx < 0) return
    const before = this.data.posts[idx]
    const optimistic: Post = { ...before, is_bookmarked: !before.is_bookmarked }
    this.patchPost(idx, optimistic)
    try {
      const res = await api.post<{ bookmarked: boolean }>(`/api/posts/${id}/bookmark`, {})
      this.patchPost(idx, { ...optimistic, is_bookmarked: res.bookmarked })
    } catch (err) {
      this.patchPost(idx, before)
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  patchPost(index: number, patch: Post) {
    const next = this.data.posts.slice()
    next[index] = patch
    this.setData({ posts: next })
  },
```

(Make sure `Post` is imported at the top: `import type { Post, FeedTab } from './types'`. Already done in Task 6.)

- [ ] **Step 3: Bind events in `community.wxml`**

Update the `<post-card .../>` block:

```xml
<post-card
  wx:for="{{posts}}"
  wx:key="id"
  post="{{item}}"
  theme="{{theme}}"
  bind:opendetail="onOpenDetail"
  bind:toggle-like="onToggleLike"
  bind:toggle-bookmark="onToggleBookmark"
/>
```

- [ ] **Step 4: Smoke**

Tap heart → fills red instantly, count bumps, no flicker. Tap bookmark → fills teal. With server reachable, network round-trip confirms. Force a server failure (stop dx-api locally) and confirm rollback + toast.

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/components/post-card/index.ts dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml
git commit -m "feat(mini): optimistic like/bookmark toggles on community post cards"
```

---

### Task 10: Post-card follow author button

**Goal:** Toggle follow on the author of a post card. Track local state per post (since two cards can show the same author and follow state may shift).

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/components/post-card/index.ts`
- Modify: `dx-mini/miniprogram/pages/community/components/post-card/index.wxml`
- Modify: `dx-mini/miniprogram/pages/community/components/post-card/index.wxss`
- Modify: `dx-mini/miniprogram/pages/community/community.ts`

- [ ] **Step 1: Add `followed` and `followPending` properties to post-card**

In `post-card/index.ts`, add to `properties`:

```ts
    followed: { type: Boolean, value: false },
    followPending: { type: Boolean, value: false },
```

Replace `onFollowTap`:

```ts
    onFollowTap() {
      if ((this.data as { followPending: boolean }).followPending) return
      this.triggerEvent('toggle-follow', {
        userId: (this.data as { post: Post }).post.author.id,
      })
    },
```

- [ ] **Step 2: Reflect `followed` in the WXML**

In `index.wxml`, replace the `follow-stub` block with:

```xml
<view class="follow-btn {{followed ? 'followed' : ''}}" catch:tap="onFollowTap">
  <block wx:if="{{followed}}">
    <dx-icon name="user-check" size="14px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
    <text>已关注</text>
  </block>
  <block wx:else>
    <dx-icon name="user-plus" size="14px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
    <text>关注</text>
  </block>
</view>
```

In `index.wxss`, replace the `.follow-stub` rule with:

```css
.follow-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--primary);
  color: var(--primary);
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
}
.follow-btn.followed {
  border-color: var(--border-color);
  color: var(--text-secondary);
  background: var(--primary-light);
}
```

- [ ] **Step 3: Track follow state in page data**

In `community.ts`:
- Add `followedUserIds: {} as Record<string, boolean>` to `data`.
- In the WXML pass `followed="{{followedUserIds[item.author.id]}}"`.
- Add the page handler:

```ts
  async onToggleFollow(e: WechatMiniprogram.CustomEvent) {
    const userId = (e.detail as { userId: string }).userId
    const before = this.data.followedUserIds[userId] || false
    this.setData({
      followedUserIds: { ...this.data.followedUserIds, [userId]: !before },
    })
    try {
      const res = await api.post<{ followed: boolean }>(`/api/users/${userId}/follow`, {})
      this.setData({
        followedUserIds: { ...this.data.followedUserIds, [userId]: res.followed },
      })
    } catch (err) {
      this.setData({
        followedUserIds: { ...this.data.followedUserIds, [userId]: before },
      })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
```

- [ ] **Step 4: Bind in `community.wxml`**

```xml
<post-card
  wx:for="{{posts}}"
  wx:key="id"
  post="{{item}}"
  theme="{{theme}}"
  followed="{{followedUserIds[item.author.id]}}"
  bind:opendetail="onOpenDetail"
  bind:toggle-like="onToggleLike"
  bind:toggle-bookmark="onToggleBookmark"
  bind:toggle-follow="onToggleFollow"
/>
```

- [ ] **Step 5: Smoke**

Tap 关注 → flips to 已关注 instantly, network confirms. Tap again → unfollows. Force a network error → rollback + toast.

- [ ] **Step 6: Commit**

```
git add dx-mini/miniprogram/pages/community/components/post-card dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml
git commit -m "feat(mini): follow author toggle on post card with optimistic state"
```

---

### Task 11: composer-popup — textarea + char counter + tag chips

**Goal:** Build the composer popup as a Component. Wire the textarea + char counter + tag chip input. No image picker yet (Task 12), no submit yet (Task 13).

**Files:**
- Create: `dx-mini/miniprogram/pages/community/components/composer-popup/index.ts`
- Create: `dx-mini/miniprogram/pages/community/components/composer-popup/index.wxml`
- Create: `dx-mini/miniprogram/pages/community/components/composer-popup/index.wxss`
- Create: `dx-mini/miniprogram/pages/community/components/composer-popup/index.json`

- [ ] **Step 1: Create `index.json`**

```json
{
  "component": true,
  "usingComponents": {
    "van-popup": "@vant/weapp/popup/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Create `index.ts`**

```ts
Component({
  options: { addGlobalClass: true },
  properties: {
    show: { type: Boolean, value: false },
    theme: { type: String, value: 'light' },
  },
  data: {
    content: '',
    tagInput: '',
    tags: [] as string[],
    imageUrl: '',
    uploading: false,
  },
  methods: {
    onContentInput(e: WechatMiniprogram.Input) {
      this.setData({ content: (e.detail as { value: string }).value })
    },
    onTagInput(e: WechatMiniprogram.Input) {
      this.setData({ tagInput: (e.detail as { value: string }).value })
    },
    onTagConfirm() {
      const raw = (this.data as { tagInput: string }).tagInput.trim().replace(/^#/, '')
      if (!raw) return
      const tags = (this.data as { tags: string[] }).tags
      if (tags.length >= 5) {
        wx.showToast({ title: '最多5个标签', icon: 'none' })
        return
      }
      if (raw.length > 20) {
        wx.showToast({ title: '标签不超过20字', icon: 'none' })
        return
      }
      if (tags.indexOf(raw) >= 0) {
        this.setData({ tagInput: '' })
        return
      }
      this.setData({ tags: tags.concat([raw]), tagInput: '' })
    },
    onTagRemove(e: WechatMiniprogram.TouchEvent) {
      const tag = e.currentTarget.dataset['tag'] as string
      const tags = (this.data as { tags: string[] }).tags.filter((t) => t !== tag)
      this.setData({ tags })
    },
    onClose() {
      this.triggerEvent('close')
    },
    onSubmit() {
      // wired in Task 13
    },
    onPickImage() {
      // wired in Task 12
    },
    onRemoveImage() {
      this.setData({ imageUrl: '' })
    },
  },
})
```

- [ ] **Step 3: Create `index.wxml`**

```xml
<van-popup
  show="{{show}}"
  position="bottom"
  round
  custom-style="max-height: 85vh;"
  safe-area-inset-bottom
  bind:close="onClose"
  close-on-click-overlay="{{false}}"
>
  <view class="composer">
    <view class="composer-header">
      <view class="header-btn cancel" bind:tap="onClose">取消</view>
      <view
        class="header-btn submit {{(content.length === 0 || uploading) ? 'disabled' : ''}}"
        bind:tap="onSubmit"
      >发布</view>
    </view>

    <view class="composer-body">
      <textarea
        class="textarea"
        value="{{content}}"
        placeholder="分享你的想法…"
        maxlength="2000"
        auto-height
        adjust-position="{{true}}"
        cursor-spacing="20"
        bind:input="onContentInput"
      />
      <view class="counter">{{content.length}}/2000</view>

      <view class="image-row" wx:if="{{imageUrl}}">
        <image src="{{imageUrl}}" class="thumb" mode="aspectFill" />
        <view class="thumb-x" catch:tap="onRemoveImage">
          <dx-icon name="x" size="14px" color="#ffffff" />
        </view>
      </view>

      <view class="tag-row">
        <view wx:for="{{tags}}" wx:key="*this" class="tag-chip">
          <text>#{{item}}</text>
          <view class="tag-x" catch:tap="onTagRemove" data-tag="{{item}}">
            <dx-icon name="x" size="12px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
          </view>
        </view>
        <input
          wx:if="{{tags.length < 5}}"
          class="tag-input"
          value="{{tagInput}}"
          placeholder="+ 添加标签"
          confirm-type="done"
          bind:input="onTagInput"
          bind:confirm="onTagConfirm"
          bind:blur="onTagConfirm"
        />
      </view>
    </view>

    <view class="composer-action">
      <view class="image-pick" bind:tap="onPickImage">
        <dx-icon name="image" size="20px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
        <text>添加图片</text>
      </view>
    </view>
  </view>
</van-popup>
```

- [ ] **Step 4: Create `index.wxss`**

```css
.composer {
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
}
.composer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}
.header-btn {
  font-size: 15px;
  padding: 4px 8px;
}
.header-btn.cancel {
  color: var(--text-secondary);
}
.header-btn.submit {
  color: var(--primary);
  font-weight: 600;
}
.header-btn.submit.disabled {
  opacity: 0.45;
}
.composer-body {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  position: relative;
}
.textarea {
  width: 100%;
  min-height: 120px;
  font-size: 15px;
  line-height: 1.55;
  color: var(--text-primary);
  background: transparent;
}
.counter {
  position: absolute;
  right: 16px;
  bottom: 8px;
  font-size: 11px;
  color: var(--text-secondary);
}
.image-row {
  margin-top: 12px;
  position: relative;
  width: 80px;
  height: 80px;
}
.thumb {
  width: 100%;
  height: 100%;
  border-radius: 8px;
}
.thumb-x {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}
.tag-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 16px;
}
.tag-chip {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
}
.tag-x { display: flex; align-items: center; }
.tag-input {
  font-size: 13px;
  color: var(--text-primary);
  border: 1px dashed var(--border-color);
  border-radius: 999px;
  padding: 4px 12px;
  min-width: 100px;
}
.composer-action {
  border-top: 1px solid var(--border-color);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.image-pick {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--primary);
}
```

- [ ] **Step 5: Wire the popup into the feed page (closed by default)**

In `community.json`, add:
```json
"composer-popup": "./components/composer-popup/index"
```

In `community.ts`, add to `data`:
```ts
    composerOpen: false,
```
…and methods:
```ts
  openComposer() {
    if (!isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.setData({ composerOpen: true })
  },
  onComposerClose() {
    this.setData({ composerOpen: false })
  },
```

In `community.wxml`, add inside the page-container, after the post list:
```xml
<view class="fab" bind:tap="openComposer">
  <dx-icon name="plus" size="22px" color="#ffffff" />
</view>

<composer-popup
  show="{{composerOpen}}"
  theme="{{theme}}"
  bind:close="onComposerClose"
/>
```

In `community.wxss`, add:
```css
.fab {
  position: fixed;
  right: 24rpx;
  bottom: calc(56px + env(safe-area-inset-bottom) + 32rpx);
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 18px rgba(13, 148, 136, 0.32);
  z-index: 10;
}
```

- [ ] **Step 6: Smoke**

Tap FAB → popup slides up. Type content, watch counter. Add a tag, hit done → chip appears. Tap × on chip → removes. Tap 取消 or backdrop-tap (won't work yet, `close-on-click-overlay=false`) — use 取消. Submit / 添加图片 are visible but no-op.

- [ ] **Step 7: Commit**

```
git add dx-mini/miniprogram/pages/community/components/composer-popup dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml dx-mini/miniprogram/pages/community/community.wxss dx-mini/miniprogram/pages/community/community.json
git commit -m "feat(mini): composer popup with textarea, char counter, tag chips"
```

---

### Task 12: composer-popup — image picker + upload

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/components/composer-popup/index.ts`

- [ ] **Step 1: Implement `onPickImage`**

Replace the stub `onPickImage` with the real flow. At the top of the file, add the imports:

```ts
import { config } from '../../../../utils/config'
import { getToken } from '../../../../utils/auth'
```

Replace the body of `onPickImage`:

```ts
    onPickImage() {
      const self = this
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success(res) {
          const file = res.tempFiles[0]
          if (file.size > 2 * 1024 * 1024) {
            wx.showToast({ title: '图片不超过 2MB', icon: 'none' })
            return
          }
          const lower = file.tempFilePath.toLowerCase()
          if (!/\.(jpg|jpeg|png)$/.test(lower)) {
            wx.showToast({ title: '仅支持 JPG/PNG', icon: 'none' })
            return
          }
          self.setData({ uploading: true, imageUrl: file.tempFilePath })
          wx.uploadFile({
            url: config.apiBaseUrl + '/api/uploads/images',
            filePath: file.tempFilePath,
            name: 'file',
            formData: { role: 'post-image' },
            header: { Authorization: 'Bearer ' + getToken() },
            success(uploadRes) {
              try {
                const body = JSON.parse(uploadRes.data) as { code: number; message: string; data: { url: string } }
                if (body.code === 0) {
                  self.setData({ imageUrl: body.data.url, uploading: false })
                } else {
                  self.setData({ imageUrl: '', uploading: false })
                  wx.showToast({ title: body.message || '上传失败', icon: 'none' })
                }
              } catch {
                self.setData({ imageUrl: '', uploading: false })
                wx.showToast({ title: '上传失败', icon: 'none' })
              }
            },
            fail() {
              self.setData({ imageUrl: '', uploading: false })
              wx.showToast({ title: '上传失败', icon: 'none' })
            },
          })
        },
      })
    },
```

(Note: while uploading, `imageUrl` holds the local `tempFilePath` for instant preview. On success it's swapped to the server-relative URL; the WXML `image` tag handles both.)

- [ ] **Step 2: Smoke**

Open composer → 添加图片 → choose photo → preview appears immediately → after upload completes, `imageUrl` is the server URL (verify by inspecting devtools data panel; or by submitting in Task 13). Try a >2MB image → toast and abort. Try a non-jpg/png → toast and abort.

- [ ] **Step 3: Commit**

```
git add dx-mini/miniprogram/pages/community/components/composer-popup/index.ts
git commit -m "feat(mini): composer popup image picker uploads to /api/uploads/images"
```

---

### Task 13: composer-popup — submit + post-created event

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/components/composer-popup/index.ts`
- Modify: `dx-mini/miniprogram/pages/community/community.ts`
- Modify: `dx-mini/miniprogram/pages/community/community.wxml`

- [ ] **Step 1: Implement `onSubmit` in composer**

At the top, add:
```ts
import { api } from '../../../../utils/api'
import type { Post } from '../../types'
```

Replace `onSubmit`:

```ts
    async onSubmit() {
      const d = this.data as { content: string; imageUrl: string; tags: string[]; uploading: boolean }
      const content = d.content.trim()
      if (!content) {
        wx.showToast({ title: '请输入内容', icon: 'none' })
        return
      }
      if (content.length > 2000) {
        wx.showToast({ title: '内容不超过 2000 字', icon: 'none' })
        return
      }
      if (d.uploading) {
        wx.showToast({ title: '图片上传中…', icon: 'none' })
        return
      }
      wx.showLoading({ title: '发布中…', mask: true })
      try {
        const post = await api.post<Post>('/api/posts', {
          content,
          image_url: d.imageUrl || null,
          tags: d.tags.length > 0 ? d.tags : null,
        })
        wx.hideLoading()
        wx.showToast({ title: '已发布', icon: 'success' })
        this.triggerEvent('postcreated', { post })
        this.setData({ content: '', tagInput: '', tags: [], imageUrl: '', uploading: false })
      } catch (err) {
        wx.hideLoading()
        wx.showToast({ title: (err as Error).message || '发布失败', icon: 'none' })
      }
    },
```

- [ ] **Step 2: Bind on the page**

In `community.wxml`, update the popup tag:
```xml
<composer-popup
  show="{{composerOpen}}"
  theme="{{theme}}"
  bind:close="onComposerClose"
  bind:postcreated="onPostCreated"
/>
```

In `community.ts`, add:
```ts
  onPostCreated(e: WechatMiniprogram.CustomEvent) {
    const post = (e.detail as { post: Post }).post
    this.setData({
      composerOpen: false,
      posts: [post, ...this.data.posts],
    })
  },
```

- [ ] **Step 3: Smoke**

Compose a text-only post → 发布 → loading → toast → popup closes → new post appears at top of feed. Repeat with image attached. Repeat with tags. Repeat with all three.

- [ ] **Step 4: Commit**

```
git add dx-mini/miniprogram/pages/community/components/composer-popup/index.ts dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml
git commit -m "feat(mini): composer submit creates post and prepends to feed"
```

---

### Task 14: composer-popup — cancel-with-unsaved-content confirm

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/components/composer-popup/index.ts`

- [ ] **Step 1: Replace `onClose`**

```ts
    onClose() {
      const d = this.data as { content: string; tags: string[]; imageUrl: string }
      const dirty = d.content.trim().length > 0 || d.tags.length > 0 || d.imageUrl.length > 0
      if (!dirty) {
        this.triggerEvent('close')
        return
      }
      const self = this
      wx.showModal({
        title: '放弃编辑？',
        content: '已输入的内容将丢失',
        confirmText: '放弃',
        cancelText: '继续编辑',
        confirmColor: '#ef4444',
        success(res) {
          if (res.confirm) {
            self.setData({ content: '', tagInput: '', tags: [], imageUrl: '', uploading: false })
            self.triggerEvent('close')
          }
        },
      })
    },
```

- [ ] **Step 2: Smoke**

Open composer, type some content, tap 取消 → modal asks. 继续编辑 keeps state. 放弃 clears and closes. Open with no content, tap 取消 → closes immediately.

- [ ] **Step 3: Commit**

```
git add dx-mini/miniprogram/pages/community/components/composer-popup/index.ts
git commit -m "feat(mini): confirm before discarding unsaved post composer content"
```

---

### Task 15: post-block component (full post display) + detail page chrome + loadPost

**Files:**
- Create: `dx-mini/miniprogram/pages/community/components/post-block/index.{ts,wxml,wxss,json}`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.{ts,wxml,wxss,json}`

- [ ] **Step 1: Create post-block `index.json`**

```json
{
  "component": true,
  "usingComponents": {
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Create post-block `index.ts`**

```ts
import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatDate, formatNumber } from '../../../../utils/format'
import { config } from '../../../../utils/config'
import type { Post } from '../../types'

Component({
  options: { addGlobalClass: true },
  properties: {
    post: { type: Object, value: null },
    theme: { type: String, value: 'light' },
    followed: { type: Boolean, value: false },
  },
  data: {
    avatarColor: '#999',
    avatarLetter: '?',
    timeText: '',
    likeText: '',
    commentText: '',
    imageAbsoluteUrl: '',
  },
  observers: {
    post(post: Post | null) {
      if (!post) return
      this.setData({
        avatarColor: getAvatarColor(post.author.id),
        avatarLetter: getAvatarLetter(post.author.nickname),
        timeText: formatDate(post.created_at),
        likeText: post.like_count > 0 ? formatNumber(post.like_count) : '',
        commentText: post.comment_count > 0 ? formatNumber(post.comment_count) : '',
        imageAbsoluteUrl: post.image_url ? config.apiBaseUrl + post.image_url : '',
      })
    },
  },
  methods: {
    onPreviewImage() {
      const url = (this.data as { imageAbsoluteUrl: string }).imageAbsoluteUrl
      if (url) wx.previewImage({ urls: [url] })
    },
    onLikeTap() {
      this.triggerEvent('toggle-like')
    },
    onBookmarkTap() {
      this.triggerEvent('toggle-bookmark')
    },
    onFollowTap() {
      this.triggerEvent('toggle-follow')
    },
  },
})
```

- [ ] **Step 3: Create post-block `index.wxml`**

```xml
<view class="block">
  <view class="header">
    <view class="author">
      <view class="avatar" wx:if="{{!post.author.avatar_url}}" style="background: {{avatarColor}};">{{avatarLetter}}</view>
      <van-image
        wx:if="{{post.author.avatar_url}}"
        src="{{post.author.avatar_url}}"
        width="48px"
        height="48px"
        round
        fit="cover"
      />
      <view class="meta">
        <text class="nickname">{{post.author.nickname}}</text>
        <text class="time">{{timeText}}</text>
      </view>
    </view>
    <view class="follow-btn {{followed ? 'followed' : ''}}" bind:tap="onFollowTap">
      <block wx:if="{{followed}}">
        <dx-icon name="user-check" size="14px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
        <text>已关注</text>
      </block>
      <block wx:else>
        <dx-icon name="user-plus" size="14px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
        <text>关注</text>
      </block>
    </view>
  </view>

  <text class="content">{{post.content}}</text>

  <view class="image-wrap" wx:if="{{imageAbsoluteUrl}}" bind:tap="onPreviewImage">
    <van-image src="{{imageAbsoluteUrl}}" width="100%" height="240px" fit="cover" radius="10" />
  </view>

  <view class="tags" wx:if="{{post.tags.length > 0}}">
    <text wx:for="{{post.tags}}" wx:key="*this" class="tag">#{{item}}</text>
  </view>

  <view class="divider"></view>

  <view class="actions">
    <view class="action" bind:tap="onLikeTap">
      <dx-icon
        name="heart"
        size="20px"
        color="{{post.is_liked ? '#ef4444' : (theme === 'dark' ? '#9ca3af' : '#6b7280')}}"
        custom-style="{{post.is_liked ? 'fill: #ef4444;' : ''}}"
      />
      <text class="action-text">{{likeText}}</text>
    </view>
    <view class="action">
      <dx-icon name="message-circle" size="20px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
      <text class="action-text">{{commentText}}</text>
    </view>
    <view class="action" bind:tap="onBookmarkTap">
      <dx-icon
        name="bookmark"
        size="20px"
        color="{{post.is_bookmarked ? (theme === 'dark' ? '#14b8a6' : '#0d9488') : (theme === 'dark' ? '#9ca3af' : '#6b7280')}}"
        custom-style="{{post.is_bookmarked ? (theme === 'dark' ? 'fill: #14b8a6;' : 'fill: #0d9488;') : ''}}"
      />
    </view>
  </view>
</view>
```

- [ ] **Step 4: Create post-block `index.wxss`**

```css
.block {
  background: var(--bg-card);
  margin: 0 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.author { display: flex; align-items: center; gap: 10px; }
.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
}
.meta { display: flex; flex-direction: column; gap: 2px; }
.nickname { font-size: 15px; font-weight: 600; color: var(--text-primary); }
.time { font-size: 12px; color: var(--text-secondary); }
.follow-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--primary);
  color: var(--primary);
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
}
.follow-btn.followed {
  border-color: var(--border-color);
  color: var(--text-secondary);
  background: var(--primary-light);
}
.content {
  font-size: 15px;
  line-height: 1.65;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}
.image-wrap { border-radius: 10px; overflow: hidden; }
.tags { display: flex; flex-wrap: wrap; gap: 6px; }
.tag {
  background: var(--primary-light);
  color: var(--primary);
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}
.divider { height: 1px; background: var(--border-color); }
.actions {
  display: flex;
  align-items: center;
  justify-content: space-around;
}
.action { display: flex; align-items: center; gap: 6px; padding: 4px 16px; }
.action-text { font-size: 13px; color: var(--text-secondary); }
```

- [ ] **Step 5: Update detail `detail.json`**

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index",
    "post-block": "../components/post-block/index"
  }
}
```

- [ ] **Step 6: Replace `detail.ts` to load the post**

```ts
import { api } from '../../../utils/api'
import type { Post } from '../types'

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    postId: '',
    post: null as Post | null,
    loading: false,
    followed: false,
  },
  onLoad(query: Record<string, string>) {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
      postId: query.id || '',
    })
    if (query.id) this.loadPost()
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
  },
  goBack() {
    wx.navigateBack()
  },
  async loadPost() {
    this.setData({ loading: true })
    try {
      const post = await api.get<Post>(`/api/posts/${this.data.postId}`)
      this.setData({ post, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: (err as Error).message || '加载失败', icon: 'none' })
    }
  },
})
```

- [ ] **Step 7: Replace `detail.wxml`**

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="status-bar-spacer"></view>
    <view class="nav-band">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">帖子</text>
      <view class="back-btn"></view>
    </view>

    <van-loading wx:if="{{loading}}" size="24px" color="#0d9488" class="center-loader" />

    <block wx:if="{{!loading && post}}">
      <post-block post="{{post}}" theme="{{theme}}" followed="{{followed}}" />
    </block>

    <van-empty wx:if="{{!loading && !post}}" description="帖子不存在" />
  </view>
</van-config-provider>
```

- [ ] **Step 8: Append to `detail.wxss`**

```css
.center-loader { display: flex; justify-content: center; padding: 32px; }
```

- [ ] **Step 9: Smoke**

Tap a post-card on the feed → detail loads → see full content (no clamp), image, tags, action row.

- [ ] **Step 10: Commit**

```
git add dx-mini/miniprogram/pages/community/components/post-block dx-mini/miniprogram/pages/community/detail
git commit -m "feat(mini): post detail page with full post block"
```

---

### Task 16: Detail page interactions (like / bookmark / follow) with eventChannel back to feed

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.ts`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxml`
- Modify: `dx-mini/miniprogram/pages/community/community.ts`
- Modify: `dx-mini/miniprogram/pages/community/community.wxml`

- [ ] **Step 1: Update navigation from feed to pass an eventChannel**

In `community.ts`, replace `onOpenDetail`:

```ts
  onOpenDetail(e: WechatMiniprogram.CustomEvent) {
    const id = (e.detail as { id: string }).id
    wx.navigateTo({
      url: `/pages/community/detail/detail?id=${id}`,
      events: {
        'post-updated': (payload: { id: string; patch: Partial<Post> }) => {
          const idx = this.data.posts.findIndex((p) => p.id === payload.id)
          if (idx < 0) return
          const next = this.data.posts.slice()
          next[idx] = { ...next[idx], ...payload.patch }
          this.setData({ posts: next })
        },
      },
    })
  },
```

- [ ] **Step 2: Add interaction handlers to `detail.ts`**

Add to the methods block:

```ts
  emitUpdate(patch: Partial<Post>) {
    if (!this.data.post) return
    let channel: WechatMiniprogram.EventChannel | null = null
    try { channel = this.getOpenerEventChannel() } catch { channel = null }
    if (channel) channel.emit('post-updated', { id: this.data.post.id, patch })
  },
  async onToggleLike() {
    if (!this.data.post) return
    const before = this.data.post
    const optimistic: Post = {
      ...before,
      is_liked: !before.is_liked,
      like_count: before.is_liked ? Math.max(before.like_count - 1, 0) : before.like_count + 1,
    }
    this.setData({ post: optimistic })
    try {
      const res = await api.post<{ liked: boolean; like_count: number }>(`/api/posts/${before.id}/like`, {})
      const next = { ...optimistic, is_liked: res.liked, like_count: res.like_count }
      this.setData({ post: next })
      this.emitUpdate({ is_liked: res.liked, like_count: res.like_count })
    } catch (err) {
      this.setData({ post: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  async onToggleBookmark() {
    if (!this.data.post) return
    const before = this.data.post
    const optimistic: Post = { ...before, is_bookmarked: !before.is_bookmarked }
    this.setData({ post: optimistic })
    try {
      const res = await api.post<{ bookmarked: boolean }>(`/api/posts/${before.id}/bookmark`, {})
      const next = { ...optimistic, is_bookmarked: res.bookmarked }
      this.setData({ post: next })
      this.emitUpdate({ is_bookmarked: res.bookmarked })
    } catch (err) {
      this.setData({ post: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  async onToggleFollow() {
    if (!this.data.post) return
    const before = this.data.followed
    this.setData({ followed: !before })
    try {
      const res = await api.post<{ followed: boolean }>(`/api/users/${this.data.post.author.id}/follow`, {})
      this.setData({ followed: res.followed })
    } catch (err) {
      this.setData({ followed: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
```

- [ ] **Step 3: Bind events in `detail.wxml`**

Update the post-block tag:
```xml
<post-block
  post="{{post}}"
  theme="{{theme}}"
  followed="{{followed}}"
  bind:toggle-like="onToggleLike"
  bind:toggle-bookmark="onToggleBookmark"
  bind:toggle-follow="onToggleFollow"
/>
```

- [ ] **Step 4: Smoke**

Open detail → tap heart → fills, count bumps. Go back to feed → the same post on the feed reflects the new state. Repeat for bookmark. Tap follow → flips. Force network failure on the call → rollback + toast.

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/detail dx-mini/miniprogram/pages/community/community.ts dx-mini/miniprogram/pages/community/community.wxml
git commit -m "feat(mini): post-detail like/bookmark/follow propagate back to feed"
```

---

### Task 17: comment-item component (display only) + replies inline

**Files:**
- Create: `dx-mini/miniprogram/pages/community/components/comment-item/index.{ts,wxml,wxss,json}`

- [ ] **Step 1: Create `index.json`**

```json
{
  "component": true,
  "usingComponents": {
    "van-image": "@vant/weapp/image/index"
  }
}
```

- [ ] **Step 2: Create `index.ts`**

```ts
import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatRelativeDate } from '../../../../utils/format'
import type { Comment, CommentWithReplies } from '../../types'

interface DisplayComment {
  comment: Comment
  replies: Comment[]
}

Component({
  options: { addGlobalClass: true },
  properties: {
    item: { type: Object, value: null },
    theme: { type: String, value: 'light' },
  },
  data: {
    parentColor: '#999',
    parentLetter: '?',
    parentTime: '',
    replyDecor: [] as { color: string; letter: string; time: string }[],
  },
  observers: {
    item(item: DisplayComment | null) {
      if (!item) return
      this.setData({
        parentColor: getAvatarColor(item.comment.author.id),
        parentLetter: getAvatarLetter(item.comment.author.nickname),
        parentTime: formatRelativeDate(item.comment.created_at),
        replyDecor: item.replies.map((r) => ({
          color: getAvatarColor(r.author.id),
          letter: getAvatarLetter(r.author.nickname),
          time: formatRelativeDate(r.created_at),
        })),
      })
    },
  },
  methods: {
    onReply() {
      const item = (this.data as { item: CommentWithReplies }).item
      this.triggerEvent('reply', {
        commentId: item.comment.id,
        nickname: item.comment.author.nickname,
      })
    },
  },
})
```

- [ ] **Step 3: Create `index.wxml`**

```xml
<view class="comment">
  <view class="parent-row">
    <view class="avatar" wx:if="{{!item.comment.author.avatar_url}}" style="background: {{parentColor}};">{{parentLetter}}</view>
    <van-image
      wx:if="{{item.comment.author.avatar_url}}"
      src="{{item.comment.author.avatar_url}}"
      width="32px"
      height="32px"
      round
      fit="cover"
    />
    <view class="comment-body">
      <view class="comment-meta">
        <text class="nickname">{{item.comment.author.nickname}}</text>
        <text class="time">{{parentTime}}</text>
      </view>
      <text class="content">{{item.comment.content}}</text>
      <view class="reply-link" bind:tap="onReply">回复</view>
    </view>
  </view>

  <view class="replies" wx:if="{{item.replies.length > 0}}">
    <view
      wx:for="{{item.replies}}"
      wx:for-item="reply"
      wx:for-index="i"
      wx:key="id"
      class="reply-row"
    >
      <view class="avatar small" wx:if="{{!reply.author.avatar_url}}" style="background: {{replyDecor[i].color}};">{{replyDecor[i].letter}}</view>
      <van-image
        wx:if="{{reply.author.avatar_url}}"
        src="{{reply.author.avatar_url}}"
        width="28px"
        height="28px"
        round
        fit="cover"
      />
      <view class="reply-body">
        <view class="comment-meta">
          <text class="nickname">{{reply.author.nickname}}</text>
          <text class="time">{{replyDecor[i].time}}</text>
        </view>
        <text class="content">{{reply.content}}</text>
      </view>
    </view>
  </view>
</view>
```

- [ ] **Step 4: Create `index.wxss`**

```css
.comment {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}
.parent-row { display: flex; gap: 10px; }
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.avatar.small {
  width: 28px;
  height: 28px;
  font-size: 11px;
}
.comment-body, .reply-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.comment-meta { display: flex; align-items: center; gap: 8px; }
.nickname { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.time { font-size: 11px; color: var(--text-secondary); }
.content {
  font-size: 13px;
  line-height: 1.55;
  color: var(--text-primary);
  white-space: pre-wrap;
}
.reply-link {
  font-size: 12px;
  color: var(--text-secondary);
  align-self: flex-start;
}
.replies {
  margin-left: 42px;
  padding-left: 12px;
  border-left: 2px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.reply-row { display: flex; gap: 10px; }
```

- [ ] **Step 5: Commit (component is unwired; will be consumed in Task 18)**

```
git add dx-mini/miniprogram/pages/community/components/comment-item
git commit -m "feat(mini): comment-item component (display + reply trigger)"
```

---

### Task 18: Detail page — load comments + render list with pagination

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.ts`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxml`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.json`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxss`

- [ ] **Step 1: Register comment-item in `detail.json`**

```json
"comment-item": "../components/comment-item/index"
```
(append to existing `usingComponents`).

- [ ] **Step 2: Extend `detail.ts` data + load**

Add to imports:
```ts
import { PaginatedData } from '../../../utils/api'
import type { CommentWithReplies } from '../types'
```

Add to `data`:
```ts
    comments: [] as CommentWithReplies[],
    commentsCursor: '',
    commentsHasMore: false,
    commentsLoading: false,
```

Add `onReachBottom`:
```ts
  onReachBottom() {
    if (this.data.commentsHasMore && !this.data.commentsLoading) {
      this.loadComments(false)
    }
  },
```

In `onLoad` after `if (query.id) this.loadPost()`, add:
```ts
    if (query.id) this.loadComments(true)
```

Add the loader:
```ts
  async loadComments(reset: boolean) {
    if (this.data.commentsLoading) return
    this.setData({ commentsLoading: true })
    const cursor = reset ? '' : this.data.commentsCursor
    const parts = ['limit=20']
    if (cursor) parts.push(`cursor=${encodeURIComponent(cursor)}`)
    try {
      const res = await api.get<PaginatedData<CommentWithReplies>>(
        `/api/posts/${this.data.postId}/comments?${parts.join('&')}`
      )
      this.setData({
        comments: reset ? res.items : [...this.data.comments, ...res.items],
        commentsCursor: res.nextCursor,
        commentsHasMore: res.hasMore,
        commentsLoading: false,
      })
    } catch (err) {
      this.setData({ commentsLoading: false })
      wx.showToast({ title: (err as Error).message || '加载评论失败', icon: 'none' })
    }
  },
```

- [ ] **Step 3: Render comments in `detail.wxml`**

After the `<post-block .../>` block, before `<van-empty .../>`:
```xml
<view class="comments-header" wx:if="{{post}}">
  <text>评论 {{post.comment_count}}</text>
</view>

<block wx:if="{{post}}">
  <comment-item
    wx:for="{{comments}}"
    wx:key="comment.id"
    item="{{item}}"
    theme="{{theme}}"
  />
  <van-loading wx:if="{{commentsLoading}}" size="20px" color="#0d9488" class="center-loader" />
  <van-empty wx:if="{{!commentsLoading && comments.length === 0}}" description="暂无评论，来抢沙发" />
</block>
```

- [ ] **Step 4: Append styles to `detail.wxss`**

```css
.comments-header {
  margin: 16px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
```

- [ ] **Step 5: Smoke**

Open a post with comments → see them list. Open one without → empty state. Long thread → reach-bottom paginates.

- [ ] **Step 6: Commit**

```
git add dx-mini/miniprogram/pages/community/detail
git commit -m "feat(mini): post detail loads and renders paginated comments"
```

---

### Task 19: Sticky comment input bar + send (top-level)

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.ts`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxml`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxss`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.json`

- [ ] **Step 1: Add input state to `detail.ts`**

```ts
import type { Comment } from '../types'
```

Add to `data`:
```ts
    inputValue: '',
    sending: false,
```

Add methods:
```ts
  onInput(e: WechatMiniprogram.Input) {
    this.setData({ inputValue: (e.detail as { value: string }).value })
  },
  async onSend() {
    const v = this.data.inputValue.trim()
    if (!v || this.data.sending) return
    this.setData({ sending: true })
    try {
      const created = await api.post<Comment>(`/api/posts/${this.data.postId}/comments`, {
        content: v,
        parent_id: null,
      })
      this.setData({
        sending: false,
        inputValue: '',
        comments: [{ comment: created, replies: [] }, ...this.data.comments],
        post: this.data.post
          ? { ...this.data.post, comment_count: this.data.post.comment_count + 1 }
          : this.data.post,
      })
      if (this.data.post) {
        this.emitUpdate({ comment_count: this.data.post.comment_count })
      }
    } catch (err) {
      this.setData({ sending: false })
      wx.showToast({ title: (err as Error).message || '评论失败', icon: 'none' })
    }
  },
```

- [ ] **Step 2: Add input bar in `detail.wxml`**

Inside the page-container, just before the closing `</view>` of `.page-container`:

```xml
<view class="input-bar">
  <input
    class="input"
    value="{{inputValue}}"
    placeholder="说点什么…"
    maxlength="500"
    confirm-type="send"
    cursor-spacing="20"
    adjust-position="{{true}}"
    bind:input="onInput"
    bind:confirm="onSend"
  />
  <view class="send-btn {{(inputValue.length === 0 || sending) ? 'disabled' : ''}}" bind:tap="onSend">
    <dx-icon name="send" size="18px" color="#ffffff" />
  </view>
</view>
```

- [ ] **Step 3: Style the input bar**

Append to `detail.wxss`:

```css
.page-container {
  padding-bottom: calc(64px + env(safe-area-inset-bottom));
}
.input-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--bg-card);
  border-top: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px calc(8px + env(safe-area-inset-bottom));
  z-index: 50;
}
.input {
  flex: 1;
  background: var(--bg-page);
  border-radius: 999px;
  padding: 8px 14px;
  font-size: 14px;
  color: var(--text-primary);
}
.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}
.send-btn.disabled { opacity: 0.4; }
```

- [ ] **Step 4: Smoke**

Open detail → type → 发送 → comment appears at the top of the list, count bumps, return to feed → comment_count on the card matches.

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/detail
git commit -m "feat(mini): sticky comment input bar with top-level send"
```

---

### Task 20: Reply mode (parent_id, placeholder swap, cancel)

**Files:**
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.ts`
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.wxml`

- [ ] **Step 1: Track replyingTo state**

Add to `data`:
```ts
    replyingTo: null as { commentId: string; nickname: string } | null,
    inputPlaceholder: '说点什么…',
```

Add a method:
```ts
  onReplyTo(e: WechatMiniprogram.CustomEvent) {
    const detail = e.detail as { commentId: string; nickname: string }
    this.setData({
      replyingTo: detail,
      inputPlaceholder: `回复 @${detail.nickname}：`,
    })
  },
  onCancelReply() {
    this.setData({ replyingTo: null, inputPlaceholder: '说点什么…' })
  },
```

Modify `onSend` to use `parent_id`:

```ts
  async onSend() {
    const v = this.data.inputValue.trim()
    if (!v || this.data.sending) return
    this.setData({ sending: true })
    const parentId = this.data.replyingTo ? this.data.replyingTo.commentId : null
    try {
      const created = await api.post<Comment>(`/api/posts/${this.data.postId}/comments`, {
        content: v,
        parent_id: parentId,
      })
      if (parentId) {
        const idx = this.data.comments.findIndex((c) => c.comment.id === parentId)
        if (idx >= 0) {
          const next = this.data.comments.slice()
          next[idx] = { ...next[idx], replies: [...next[idx].replies, created] }
          this.setData({ comments: next })
        }
      } else {
        this.setData({ comments: [{ comment: created, replies: [] }, ...this.data.comments] })
      }
      this.setData({
        sending: false,
        inputValue: '',
        replyingTo: null,
        inputPlaceholder: '说点什么…',
        post: this.data.post
          ? { ...this.data.post, comment_count: this.data.post.comment_count + 1 }
          : this.data.post,
      })
      if (this.data.post) {
        this.emitUpdate({ comment_count: this.data.post.comment_count })
      }
    } catch (err) {
      this.setData({ sending: false })
      wx.showToast({ title: (err as Error).message || '评论失败', icon: 'none' })
    }
  },
```

- [ ] **Step 2: Bind events in `detail.wxml`**

Update the comment-item tag to forward replies:
```xml
<comment-item
  wx:for="{{comments}}"
  wx:key="comment.id"
  item="{{item}}"
  theme="{{theme}}"
  bind:reply="onReplyTo"
/>
```

Update the input bar — swap placeholder and add cancel chip:
```xml
<view class="input-bar">
  <view wx:if="{{replyingTo}}" class="reply-chip" bind:tap="onCancelReply">
    <text>回复 {{replyingTo.nickname}}</text>
    <dx-icon name="x" size="12px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
  </view>
  <input
    class="input"
    value="{{inputValue}}"
    placeholder="{{inputPlaceholder}}"
    maxlength="500"
    confirm-type="send"
    cursor-spacing="20"
    adjust-position="{{true}}"
    bind:input="onInput"
    bind:confirm="onSend"
  />
  <view class="send-btn {{(inputValue.length === 0 || sending) ? 'disabled' : ''}}" bind:tap="onSend">
    <dx-icon name="send" size="18px" color="#ffffff" />
  </view>
</view>
```

- [ ] **Step 3: Style the reply chip**

Append to `detail.wxss`:
```css
.reply-chip {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
}
```

- [ ] **Step 4: Smoke**

Tap 回复 on a top-level comment → input placeholder shows `回复 @<name>：` and the reply chip appears. Type, send → reply appears under that comment's `replies` block. Tap the chip → cancels reply mode (back to top-level placeholder). Server rejects nested-replies — the UI prevents reaching that state because reply links are only on top-level comments.

- [ ] **Step 5: Commit**

```
git add dx-mini/miniprogram/pages/community/detail
git commit -m "feat(mini): reply mode in comment input (parent_id + placeholder swap)"
```

---

### Task 21: Final smoke + dark mode pass

**Files:**
- Possibly minor tweaks to: any of the above WXSS files

- [ ] **Step 1: Run the full smoke checklist from the spec**

In WeChat DevTools simulator, log in, then walk through:
- 社区 tab loads, all 4 tabs work, pull-to-refresh and reach-bottom paginate.
- Compose: text-only, text+image, text+tags, text+image+tags. Cancel-with-content prompts confirm.
- Detail: open from a post card, like/bookmark/follow flip and propagate back to feed when navigating back. Image preview works (`wx.previewImage`). Comments paginate. Top-level send works. Reply mode works. Empty-comment state shows.
- Auth: log out → tap 社区 → reLaunches to login.
- Profile-edit: avatar upload still succeeds.

- [ ] **Step 2: Toggle dark mode (`/pages/me/me` → 主题)**

Walk every page once in dark. Common issues: hard-coded white/black colors; missing dark-mode rules. Fix any.

- [ ] **Step 3: TS / WXML cleanliness**

```
cd dx-mini && npx tsc --noEmit -p tsconfig.json
```
Expected: zero errors **except the existing `Component({ methods })` `this` typing limitation already accepted by the codebase**. Do not introduce new `// @ts-ignore`.

Then re-run icon validator (catches accidental new `<dx-icon name="…" />` references not in ICONS):
```
cd dx-mini && npm run build:icons
```

- [ ] **Step 4: Commit any final polish**

```
git add -A
git commit -m "chore(mini): dark-mode polish and smoke test fixups for community"
```

---

## Phase 2 — Adjacent fixes (after Phase 1 lands)

### Task 22: Fix dx-web `bookmarks` → `bookmarked` tab name

**Files:**
- Modify: `dx-web/src/features/web/community/types/post.ts`
- Modify: `dx-web/src/features/web/community/components/feed-tabs.tsx`
- Search-and-replace: `dx-web/src/features/web/community/`

- [ ] **Step 1: Rename in the type definition**

In `dx-web/src/features/web/community/types/post.ts`, line 33:

```ts
export type FeedTab = "latest" | "hot" | "following" | "bookmarked";
```

(Was `"bookmarks"`.)

- [ ] **Step 2: Update the FeedTabs label table**

Open `dx-web/src/features/web/community/components/feed-tabs.tsx`. Find the TABS array and change the entry whose `key` is `"bookmarks"` so that `key: "bookmarked"`. Leave the visible label `"收藏"` as-is.

- [ ] **Step 3: Grep for any remaining references**

```
rg '"bookmarks"' dx-web/src/features/web/community/
rg '"bookmarks"' dx-web/src/components/
```
Expected: no results — TypeScript will also flag the rename.

- [ ] **Step 4: TS compile**

```
cd dx-web && npx tsc --noEmit
```
Expected: clean.

- [ ] **Step 5: Manual verify**

`cd dx-web && npm run dev`. Bookmark 2-3 posts on `/hall/community`. Click the 收藏 tab. Expected: only those bookmarked posts appear (was previously: same as 最新).

- [ ] **Step 6: Commit**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/community
git commit -m "fix(web): correct community feed tab value to 'bookmarked' to match server"
```

---

### Task 23: Edit / delete UI — dx-mini

**Goal:** Add a more-horizontal action button on the post-detail post-block and on each comment-item, visible only when the current user owns the row. Tap → action sheet with `编辑` / `删除`. Edit re-opens the composer-popup pre-filled (post) or swaps the input bar to edit mode (comment). Delete shows a confirm modal then calls the API.

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs` (add `more-horizontal`)
- Modify: `dx-mini/miniprogram/pages/community/components/post-block/index.{ts,wxml,wxss}`
- Modify: `dx-mini/miniprogram/pages/community/components/comment-item/index.{ts,wxml}`
- Modify: `dx-mini/miniprogram/pages/community/components/composer-popup/index.{ts,wxml}` (add edit mode)
- Modify: `dx-mini/miniprogram/pages/community/detail/detail.{ts,wxml}`

- [ ] **Step 1: Register `more-horizontal` icon**

In `dx-mini/scripts/build-icons.mjs`, append `['more-horizontal', 'more-horizontal'],` to the ICONS array. Run:
```
cd dx-mini && npm run build:icons
```

- [ ] **Step 2: post-block — owner-aware more button**

In `post-block/index.ts`, add property:
```ts
    isOwner: { type: Boolean, value: false },
```

In `index.ts` add the method:
```ts
    onMoreTap() {
      this.triggerEvent('open-actions')
    },
```

In `index.wxml`, add a more button inside the header, right of the follow-btn (only when `isOwner`):
```xml
<view class="more-btn" wx:if="{{isOwner}}" catch:tap="onMoreTap">
  <dx-icon name="more-horizontal" size="20px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
</view>
```

Add the `dx-icon` reference is already present via `index.json`. Append style:
```css
.more-btn { padding: 4px; }
```

- [ ] **Step 3: detail page — owner check + action sheet**

Add import:
```ts
import { getUserId } from '../../../utils/auth'
```

Add to `data`:
```ts
    isOwner: false,
```

In `loadPost`, after setting `post`, set:
```ts
this.setData({ isOwner: getUserId() === post.author.id })
```

Add the action handler:
```ts
  onPostActions() {
    if (!this.data.post) return
    const self = this
    wx.showActionSheet({
      itemList: ['编辑', '删除'],
      success(res) {
        if (res.tapIndex === 0) self.editPost()
        if (res.tapIndex === 1) self.deletePost()
      },
    })
  },
  editPost() {
    if (!this.data.post) return
    this.setData({
      composerOpen: true,
      composerEdit: { ...this.data.post },
    })
  },
  async deletePost() {
    if (!this.data.post) return
    const id = this.data.post.id
    const confirmed = await new Promise<boolean>((resolve) => {
      wx.showModal({
        title: '确认删除',
        content: '删除后无法恢复',
        confirmText: '删除',
        cancelText: '取消',
        confirmColor: '#ef4444',
        success(res) { resolve(res.confirm) },
      })
    })
    if (!confirmed) return
    try {
      await api.delete(`/api/posts/${id}`)
      let channel: WechatMiniprogram.EventChannel | null = null
      try { channel = this.getOpenerEventChannel() } catch { channel = null }
      if (channel) channel.emit('post-deleted', { id })
      wx.showToast({ title: '已删除', icon: 'success' })
      wx.navigateBack()
    } catch (err) {
      wx.showToast({ title: (err as Error).message || '删除失败', icon: 'none' })
    }
  },
```

Add `composerOpen` and `composerEdit` to `data`:
```ts
    composerOpen: false,
    composerEdit: null as Post | null,
```

In `detail.wxml`, bind the post-block more event:
```xml
<post-block
  ...
  is-owner="{{isOwner}}"
  bind:open-actions="onPostActions"
/>
```

And add a composer instance for edit mode (also need to add `composer-popup` to `detail.json` usingComponents):
```json
"composer-popup": "../components/composer-popup/index"
```

```xml
<composer-popup
  show="{{composerOpen}}"
  theme="{{theme}}"
  edit-post="{{composerEdit}}"
  bind:close="onComposerClose"
  bind:postupdated="onPostUpdated"
/>
```

```ts
  onComposerClose() {
    this.setData({ composerOpen: false, composerEdit: null })
  },
  onPostUpdated(e: WechatMiniprogram.CustomEvent) {
    const updated = (e.detail as { post: Post }).post
    this.setData({ post: updated, composerOpen: false, composerEdit: null })
    this.emitUpdate({
      content: updated.content,
      image_url: updated.image_url,
      tags: updated.tags,
    })
  },
```

- [ ] **Step 4: composer-popup edit mode**

In `composer-popup/index.ts`, add property:
```ts
    editPost: { type: Object, value: null },
```

Add observer to pre-fill:
```ts
  observers: {
    editPost(p: Post | null) {
      if (p) {
        this.setData({
          content: p.content,
          imageUrl: p.image_url ? '' : '',
          tags: p.tags.slice(),
        })
        if (p.image_url) {
          // store the absolute preview URL so the user sees the existing image
          this.setData({ imageUrl: '__keep__:' + p.image_url })
        }
      }
    },
  },
```

Branch the submit between create and update:
```ts
    async onSubmit() {
      const d = this.data as { content: string; imageUrl: string; tags: string[]; uploading: boolean; editPost: Post | null }
      const content = d.content.trim()
      if (!content) { wx.showToast({ title: '请输入内容', icon: 'none' }); return }
      if (d.uploading) { wx.showToast({ title: '图片上传中…', icon: 'none' }); return }
      let imageUrl: string | null = null
      if (d.imageUrl.startsWith('__keep__:')) imageUrl = d.imageUrl.slice('__keep__:'.length)
      else if (d.imageUrl) imageUrl = d.imageUrl

      wx.showLoading({ title: d.editPost ? '保存中…' : '发布中…', mask: true })
      try {
        if (d.editPost) {
          await api.put(`/api/posts/${d.editPost.id}`, {
            content,
            image_url: imageUrl,
            tags: d.tags.length > 0 ? d.tags : null,
          })
          const updated: Post = {
            ...d.editPost,
            content,
            image_url: imageUrl,
            tags: d.tags,
          }
          wx.hideLoading()
          wx.showToast({ title: '已保存', icon: 'success' })
          this.triggerEvent('postupdated', { post: updated })
        } else {
          const post = await api.post<Post>('/api/posts', {
            content,
            image_url: imageUrl,
            tags: d.tags.length > 0 ? d.tags : null,
          })
          wx.hideLoading()
          wx.showToast({ title: '已发布', icon: 'success' })
          this.triggerEvent('postcreated', { post })
        }
        this.setData({ content: '', tagInput: '', tags: [], imageUrl: '', uploading: false })
      } catch (err) {
        wx.hideLoading()
        wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
      }
    },
```

Update `onPickImage` to overwrite the `__keep__:` placeholder when user picks a new image (no code change needed — `imageUrl` is replaced wholesale).

In the WXML thumbnail, render the kept image correctly:
```xml
<view class="image-row" wx:if="{{imageUrl}}">
  <image
    src="{{imageUrl[0] === '_' ? imageBaseForKeep : imageUrl}}"
    class="thumb"
    mode="aspectFill"
  />
  <view class="thumb-x" catch:tap="onRemoveImage">
    <dx-icon name="x" size="14px" color="#ffffff" />
  </view>
</view>
```

Actually — simpler: store a derived `imagePreview` in the component data. In `index.ts`, add to `data`:
```ts
    imagePreview: '',
```

Compute it via observers when imageUrl changes:
```ts
  observers: {
    editPost(p: Post | null) { /* same as above */ },
    imageUrl(v: string) {
      if (!v) this.setData({ imagePreview: '' })
      else if (v.startsWith('__keep__:')) {
        const path = v.slice('__keep__:'.length)
        this.setData({ imagePreview: require('../../../../utils/config').config.apiBaseUrl + path })
      }
      else this.setData({ imagePreview: v })
    },
  },
```

(Remove the awkward inline conditional from the WXML.)

WXML now becomes:
```xml
<view class="image-row" wx:if="{{imageUrl}}">
  <image src="{{imagePreview}}" class="thumb" mode="aspectFill" />
  <view class="thumb-x" catch:tap="onRemoveImage">
    <dx-icon name="x" size="14px" color="#ffffff" />
  </view>
</view>
```

- [ ] **Step 5: comment-item — owner-aware action sheet**

In `comment-item/index.ts`, add property:
```ts
    isOwner: { type: Boolean, value: false },
```

Add method:
```ts
    onMore() {
      this.triggerEvent('open-actions', {
        commentId: (this.data as { item: CommentWithReplies }).item.comment.id,
      })
    },
```

In WXML, add a more-button row:
```xml
<view class="reply-link" bind:tap="onReply">回复</view>
<view class="more-link" wx:if="{{isOwner}}" catch:tap="onMore">
  <dx-icon name="more-horizontal" size="14px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
</view>
```

Append style:
```css
.more-link { padding: 4px; align-self: flex-start; }
```

- [ ] **Step 6: Detail page — comment edit/delete actions**

In `detail.ts`, evaluate ownership per comment via the page-level `currentUserId`:
```ts
    currentUserId: '',
```

In `onLoad`:
```ts
this.setData({ currentUserId: getUserId() || '' })
```

Pass `is-owner` to comment-item via WXML:
```xml
<comment-item
  wx:for="{{comments}}"
  wx:key="comment.id"
  item="{{item}}"
  theme="{{theme}}"
  is-owner="{{currentUserId === item.comment.author.id}}"
  bind:reply="onReplyTo"
  bind:open-actions="onCommentActions"
/>
```

Add the action handler:
```ts
  onCommentActions(e: WechatMiniprogram.CustomEvent) {
    const commentId = (e.detail as { commentId: string }).commentId
    const self = this
    wx.showActionSheet({
      itemList: ['编辑', '删除'],
      success(res) {
        if (res.tapIndex === 0) self.editComment(commentId)
        if (res.tapIndex === 1) self.deleteComment(commentId)
      },
    })
  },
  editComment(commentId: string) {
    const target = this.data.comments.find((c) => c.comment.id === commentId)
    if (!target) return
    this.setData({
      replyingTo: null,
      inputValue: target.comment.content,
      inputPlaceholder: '编辑评论',
      editingCommentId: commentId,
    })
  },
  async deleteComment(commentId: string) {
    const confirmed = await new Promise<boolean>((resolve) => {
      wx.showModal({
        title: '确认删除',
        content: '删除后无法恢复',
        confirmText: '删除',
        cancelText: '取消',
        confirmColor: '#ef4444',
        success(res) { resolve(res.confirm) },
      })
    })
    if (!confirmed) return
    try {
      await api.delete(`/api/posts/${this.data.postId}/comments/${commentId}`)
      this.setData({
        comments: this.data.comments.filter((c) => c.comment.id !== commentId),
        post: this.data.post
          ? { ...this.data.post, comment_count: Math.max(this.data.post.comment_count - 1, 0) }
          : this.data.post,
      })
      if (this.data.post) {
        this.emitUpdate({ comment_count: this.data.post.comment_count })
      }
      wx.showToast({ title: '已删除', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: (err as Error).message || '删除失败', icon: 'none' })
    }
  },
```

Add to `data`:
```ts
    editingCommentId: '',
```

Modify `onSend` to branch on `editingCommentId`:
```ts
  async onSend() {
    const v = this.data.inputValue.trim()
    if (!v || this.data.sending) return
    this.setData({ sending: true })
    try {
      if (this.data.editingCommentId) {
        await api.put(`/api/posts/${this.data.postId}/comments/${this.data.editingCommentId}`, {
          content: v,
        })
        const idx = this.data.comments.findIndex((c) => c.comment.id === this.data.editingCommentId)
        if (idx >= 0) {
          const next = this.data.comments.slice()
          next[idx] = {
            ...next[idx],
            comment: { ...next[idx].comment, content: v },
          }
          this.setData({ comments: next })
        }
        this.setData({
          sending: false,
          inputValue: '',
          editingCommentId: '',
          inputPlaceholder: '说点什么…',
        })
        wx.showToast({ title: '已保存', icon: 'success' })
        return
      }
      const parentId = this.data.replyingTo ? this.data.replyingTo.commentId : null
      const created = await api.post<Comment>(`/api/posts/${this.data.postId}/comments`, {
        content: v,
        parent_id: parentId,
      })
      if (parentId) {
        const idx = this.data.comments.findIndex((c) => c.comment.id === parentId)
        if (idx >= 0) {
          const next = this.data.comments.slice()
          next[idx] = { ...next[idx], replies: [...next[idx].replies, created] }
          this.setData({ comments: next })
        }
      } else {
        this.setData({ comments: [{ comment: created, replies: [] }, ...this.data.comments] })
      }
      this.setData({
        sending: false,
        inputValue: '',
        replyingTo: null,
        inputPlaceholder: '说点什么…',
        post: this.data.post
          ? { ...this.data.post, comment_count: this.data.post.comment_count + 1 }
          : this.data.post,
      })
      if (this.data.post) {
        this.emitUpdate({ comment_count: this.data.post.comment_count })
      }
    } catch (err) {
      this.setData({ sending: false })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
```

- [ ] **Step 7: Smoke**

Log in as the post author. Open detail → tap more on the post → 编辑 → composer pre-fills → save → post updates. Tap more again → 删除 → confirm → navigates back, post gone from feed. Repeat with comments. Log in as a different user → no more buttons appear.

- [ ] **Step 8: Commit**

```
git add dx-mini
git commit -m "feat(mini): edit/delete UI for own posts and comments on detail page"
```

---

### Task 24: Edit / delete UI — dx-web

**Files:**
- Modify: `dx-web/src/features/web/community/components/post-card.tsx`
- Modify: `dx-web/src/features/web/community/components/comment-item.tsx`
- Modify: `dx-web/src/features/web/community/components/create-post-dialog.tsx` (extend to edit mode)
- Modify: `dx-web/src/features/web/community/actions/post.action.ts` (already exposes update / delete; verify)
- Possibly add: `dx-web/src/features/web/community/components/post-actions-menu.tsx` (small dropdown)

- [ ] **Step 1: Verify update/delete in `post.action.ts` are exposed**

Confirm the file already has `postApi.update`, `postApi.delete`, `postApi.updateComment`, `postApi.deleteComment`. If any are missing, add them in the same shape as `postApi.create`.

- [ ] **Step 2: Add a dropdown menu component**

Create `dx-web/src/features/web/community/components/post-actions-menu.tsx` using shadcn/ui's `DropdownMenu` primitive:

```tsx
"use client"

import { MoreHorizontal } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface Props {
  onEdit: () => void
  onDelete: () => void
  size?: "sm" | "md"
}

export function PostActionsMenu({ onEdit, onDelete, size = "md" }: Props) {
  const dim = size === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="rounded-full p-1.5 text-muted-foreground hover:bg-muted"
          aria-label="more"
        >
          <MoreHorizontal className={dim} />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-28">
        <DropdownMenuItem onClick={onEdit}>编辑</DropdownMenuItem>
        <DropdownMenuItem onClick={onDelete} className="text-red-600 focus:text-red-600">
          删除
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
```

(If `dropdown-menu` is not yet imported in `components/ui/`, run `npx shadcn@latest add dropdown-menu` from the `dx-web` directory first.)

- [ ] **Step 3: Wire into `post-card.tsx`**

Pull current user via `useUser` (or however the auth hook is exposed in dx-web — check `dx-web/src/lib/auth.ts` and any `use*User*` hooks; if none, accept `currentUserId?: string` as a new prop and pass it from the page using server-rendered auth).

Where the header renders today (lines ~62-94), add (only when current user owns the post):

```tsx
{currentUserId === post.author.id && (
  <PostActionsMenu
    onEdit={() => setEditOpen(true)}
    onDelete={handleDelete}
  />
)}
```

Add `editOpen` state and `handleDelete`:
```tsx
const [editOpen, setEditOpen] = useState(false)

async function handleDelete() {
  if (!confirm("确认删除？")) return
  try {
    const res = await postApi.delete(post.id)
    if (res.code !== 0) { toast.error(res.message); return }
    toast.success("已删除")
    onMutate?.()
  } catch {
    toast.error("删除失败")
  }
}
```

Render the edit dialog conditionally:
```tsx
{editOpen && (
  <CreatePostDialog
    open={editOpen}
    onOpenChange={setEditOpen}
    onCreated={() => onMutate?.()}
    editPost={post}
  />
)}
```

- [ ] **Step 4: Extend `create-post-dialog.tsx` for edit mode**

Add a new prop `editPost?: Post`. When set:
- Title becomes `编辑帖子` instead of `发布帖子`.
- Initial state pre-fills `content`, `tags` from `editPost`.
- Submit branches to `postApi.update(editPost.id, ...)` instead of `postApi.create(...)`.
- Success toast: `已保存` instead of `发布成功`.

Reset state when `editPost` changes (use `useEffect` keyed on `editPost?.id`).

- [ ] **Step 5: Wire into `comment-item.tsx`**

Add similar logic: a `currentUserId` prop, render `PostActionsMenu size="sm"` next to the 回复 link when owned, with `onEdit` toggling an inline edit `CommentInput` pre-filled with content (extend `CommentInput` with optional `commentId` and pre-filled `initialContent` for edit mode), and `onDelete` calling `postApi.deleteComment(postId, comment.id)`.

- [ ] **Step 6: Pass `currentUserId` from the page**

In `dx-web/src/app/(web)/hall/(main)/community/page.tsx` (or wherever the top-level community page lives), grab the user from `auth()` and pass `currentUserId={user.id}` down to `<CommunityFeed currentUserId={user.id} />`. The page is a server component already; just thread the prop through `CommunityFeed` → `PostCard`.

- [ ] **Step 7: TS compile + dev test**

```
cd dx-web && npx tsc --noEmit && npm run dev
```

Smoke: log in as a post author. Hover/tap more → 编辑 / 删除. Edit a post → save → list updates. Delete → list shrinks. Log in as another user → no more menu appears on others' posts.

- [ ] **Step 8: Commit**

```
git add dx-web
git commit -m "feat(web): edit/delete UI on community posts and comments"
```

---

## Self-Review

**Spec coverage check:**
- Task 1 → Icon registration (spec §"Icon registration") ✓
- Task 2 → Avatar util (spec §"UI: feed page" / "Open questions") ✓
- Task 3 → Types (spec §"Types") ✓
- Task 4 → Routing wire-up (spec §"Routing wire-up") ✓
- Task 5 → Adjacent profile-edit role fix (spec §"Adjacent fix included in Phase 1") ✓
- Tasks 6–10 → Feed page (spec §"UI: feed page") ✓
- Tasks 11–14 → Composer popup (spec §"UI: composer popup") ✓
- Tasks 15–20 → Post-detail page (spec §"UI: post-detail page") ✓
- Task 21 → Final verification (spec §"Verification") ✓
- Task 22 → Phase 2.1 dx-web bookmarks bug ✓
- Tasks 23–24 → Phase 2.2 edit/delete UI ✓

**Placeholder scan:** No `TBD` / `TODO` / `appropriate handling` left. Each step has concrete code.

**Type / signature consistency:** `Post`, `Comment`, `CommentWithReplies`, `FeedTab` are defined once in Task 3 and consumed identically afterward. The page-level `patchPost` signature, `emitUpdate({patch: Partial<Post>})` payload, and event names (`opendetail`, `toggle-like`, `toggle-bookmark`, `toggle-follow`, `postcreated`, `postupdated`, `postdeleted`, `reply`, `open-actions`) match between component triggers and page bindings.
