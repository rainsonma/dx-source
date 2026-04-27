# dx-mini 社区 Tab — Design

**Date:** 2026-04-27
**Author:** rainson + Claude
**Status:** Draft (pending user review)
**Touches:** `dx-mini` (Phase 1, primary), `dx-web` (Phase 2, follow-up fixes)

## Goal

Replace the dx-mini stub at `/pages/me/community/community` ("敬请期待") with a full community feature that mirrors dx-web `/hall/community` in functionality, while adopting native mini-program UX idioms (bottom-popup composer, sticky comment input, tap-to-detail).

Reach functional parity with dx-api's existing community endpoints, plus include image upload (which dx-web has staged in schema but not yet exposed in UI).

## Non-goals (v1)

- Edit / delete UI for own posts and comments → deferred to Phase 2.
- Real-time WebSocket updates for new comments / posts → out of scope.
- User profile pages reachable from author taps → no-op for v1, matches dx-web.
- Multi-image posts, @mentions, hashtag autocomplete, report/block.
- Content-safety review (`msg_sec_check`) — required before WeChat shop submission, but separate workstream.

## Source-of-truth surveys

The design is grounded in three exploratory surveys (run during brainstorming):

- dx-web `/hall/community` feature inventory: pages, components, hooks, schemas, types, all wire calls. **No duplicate community pages elsewhere in dx-web** — only `/wiki/community` (docs), `home/components/community-section.tsx` (marketing), and a daily-challenge-card cross-link. All are intentional and stay.
- dx-api community endpoint reference: routes, request/response shapes, models, error codes, pagination semantics. Endpoints fully cover the feature surface; no new server work needed for Phase 1.
- dx-mini structural conventions: page registration, custom tab bar, theme tokens, Vant Weapp components in use, icon registration via `scripts/build-icons.mjs`, `wx.uploadFile` pattern in `pages/me/profile-edit/`.

## Architecture

### File layout (Phase 1)

```
dx-mini/miniprogram/
├── pages/community/
│   ├── community.{ts,wxml,wxss,json}             # tab page — feed
│   ├── detail/
│   │   └── detail.{ts,wxml,wxss,json}             # sub-page — single post + comments
│   ├── components/
│   │   ├── post-card/                             # used by feed list
│   │   ├── post-block/                            # full post in detail header (no truncation)
│   │   ├── comment-item/                          # comment + nested replies
│   │   └── composer-popup/                        # bottom-popup compose UI
│   └── types.ts                                   # Post, Comment, CommentWithReplies, FeedTab
└── scripts/build-icons.mjs                        # extend ICONS array
```

### Routing wire-up

- `app.json` — register `pages/community/community` and `pages/community/detail/detail`. Mark feed page `enablePullDownRefresh: true`. Remove `pages/me/community/community` registration.
- `custom-tab-bar/index.ts` & `index.wxml` — point tab 3 (`message-square`, "社区") at `/pages/community/community`. Tab order, icon, label unchanged.
- Delete `dx-mini/miniprogram/pages/me/community/` folder entirely after the swap (no other references — the only inbound link was the old tab-bar entry).

### Component placement rationale

Local components live under `pages/community/components/` (matching the existing dx-mini convention where `home-*` sub-components live next to `pages/home/`). They are community-specific and not reused elsewhere; keeping them adjacent narrows blast radius and matches the codebase's existing pattern.

## Data layer

### API mapping (dx-mini → dx-api)

| Action | Endpoint | Shape |
|---|---|---|
| List feed | `GET /api/posts?tab={tab}&cursor={c}&limit=20` | `PaginatedData<Post>` |
| Get post | `GET /api/posts/{id}` | `Post` |
| Create post | `POST /api/posts` body `{content, image_url?, tags?}` | `Post` |
| Toggle like | `POST /api/posts/{id}/like` | `{liked, like_count}` |
| Toggle bookmark | `POST /api/posts/{id}/bookmark` | `{bookmarked}` |
| Toggle follow author | `POST /api/users/{uid}/follow` | `{followed}` |
| List comments | `GET /api/posts/{id}/comments?cursor={c}&limit=20` | `PaginatedData<CommentWithReplies>` |
| Create comment / reply | `POST /api/posts/{id}/comments` body `{content, parent_id?}` | `Comment` |
| Upload post image | `POST /api/uploads/images` multipart `{file, role:"post-image"}` | `{url}` |

### Tab values on the wire

`latest | hot | following | bookmarked` — note `bookmarked` (not `bookmarks`). dx-web sends `bookmarks` and silently falls through to `latest` server-side; the mini will use the value the server actually accepts so the 收藏 tab works from day one. (See Phase 2 follow-up to fix dx-web.)

### Types

Defined in `pages/community/types.ts`, mirroring dx-web's structures:

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

export type FeedTab = "latest" | "hot" | "following" | "bookmarked"
```

### Pagination

Cursor-based, matching dx-mini's existing `PaginatedData<T>` from `utils/api.ts`. `onReachBottom()` invokes `loadMore()` when `hasMore && !loading`. Pull-down (`onPullDownRefresh`) resets the cursor and refetches.

### Optimistic updates

Like / bookmark / follow flip local state immediately; on API failure, roll back and surface `wx.showToast({title: err.message, icon: 'none'})`. Mirrors dx-web's `PostActions` pattern.

Comment / reply creates append a temporary item with `pending: true`, replaced by the server response on success or removed + toast on error. Post `comment_count` bumps locally and propagates back to the feed via `wx.getOpenerEventChannel()` so the user sees the updated count without re-fetching the feed.

### Auth gating

Both pages require login (dx-api enforces user JWT on every endpoint used). On `onShow()`, check `isLoggedIn()` from `utils/auth.ts`. If not logged in:
- Feed (tab page) → `wx.reLaunch({ url: '/pages/login/login' })`.
- Detail (sub-page) → `wx.redirectTo({ url: '/pages/login/login' })`.

`utils/api.ts` already auto-redirects on HTTP 401 as a defense-in-depth fallback.

### Errors

| Layer | Behavior |
|---|---|
| 401 | `utils/api.ts` clears token + reLaunch to login (existing behavior) |
| 4xx with envelope message | `wx.showToast({title: body.message, icon: 'none'})` |
| Network failure | `wx.showToast({title: '网络异常，请重试', icon: 'none'})` |
| Empty list | `van-empty description="暂无帖子"` (or `"暂无评论，来抢沙发"`) |

## UI: feed page (`pages/community/community`)

### Layout (top → bottom)

```
status bar spacer (statusBarHeight)
┌─ custom nav band ────────────────────────────────┐
│  title "社区" centered                            │  navigationStyle: "custom"
├─ van-tabs (sticky, capsule style) ───────────────┤
│   最新   热门   关注   收藏                       │
├─ post list ──────────────────────────────────────┤
│   [post-card] ...                                 │  infinite list
│   van-loading at bottom while fetching            │
│   van-empty if list is empty                      │
├─ FAB: floating "+" bottom-right ─────────────────┤  opens composer-popup
└─ tab bar (custom, height 56px + safe-area) ──────┘
```

### Behaviors

- **Tab switch** (`van-tabs` `bind:change`) → updates `tab` in data, resets cursor + items, refetches.
- **Post card tap** → `wx.navigateTo({ url: '/pages/community/detail/detail?id=' + postId })`. Tap bubbling is stopped on the like / bookmark / follow controls so they don't trigger the navigation.
- **Like / bookmark** → optimistic toggle, server confirms.
- **Follow author** → optimistic toggle on the post-card, button shows pending spinner while in flight.
- **Pull-to-refresh** → resets list. `wx.stopPullDownRefresh()` after the refetch resolves.
- **Reach-bottom** → `onReachBottom()` paginates if `hasMore && !loading`.
- **FAB +** → opens `composer-popup`. Defensive `isLoggedIn()` check before opening (the page is already auth-gated, but cheap to repeat).
- **Theme** → `van-config-provider theme="{{theme}}"` reads `app.globalData.theme`.
- **Status bar** → measured in `onLoad`, exposed as CSS var `--status-bar-height`. Same pattern as `pages/leaderboard/leaderboard`.

### post-card component

Renders one post. Props: `post: Post`. Emits:
- `bind:tap` (whole card minus interactive controls) → page navigates to detail.
- `like-toggle`, `bookmark-toggle`, `follow-toggle` → page handles API + optimistic state.

Layout:
```
┌─────────────────────────────────────────────────┐
│ avatar • nickname • 3小时前 • [+ 关注] btn       │  header row
│ content (line-clamp 6, "展开 →" if overflow)    │  taps to detail
│ image (if present, rounded, full width)         │
│ #tag1  #tag2  #tag3                              │  teal chips
│ ─────────────                                    │
│ ♥ 12      💬 5      🔖                          │  actions row
└─────────────────────────────────────────────────┘
```

Avatar: if `author.avatar_url` is set, render via `van-image`. Otherwise render initial-circle with deterministic background. Port `dx-web/src/lib/avatar.ts` `getAvatarColor` to a new shared util at `dx-mini/miniprogram/utils/avatar.ts` (single source of truth — feed cards, detail page, comment items all consume it).

Time: uses the existing `formatRelativeDate` from `utils/format.ts` as-is. Output buckets are `今天` / `昨天` / `N天前` / `N周前` / `N月前` / `N年前`. No bespoke 30-day fallback — the helper already covers the full range.

Content overflow: 6-line clamp via `-webkit-line-clamp: 6` plus `text-overflow: ellipsis`. A "展开 →" link at the trailing edge (when content overflows) is purely visual — tapping anywhere on the card body navigates to detail.

## UI: composer popup (`composer-popup` component)

`van-popup` with `position="bottom"`, `round`, `safe-area-inset-bottom`, dynamic height capped at ~85vh.

### Layout

```
┌─ header bar ──────────────────────────────────────┐
│  取消              发布                             │  left dismiss / right submit
├─ body (scrollable) ───────────────────────────────┤
│  textarea, placeholder "分享你的想法…"              │
│    autosize, max-length 2000                       │
│    char counter "1234/2000" bottom-right of body   │
│                                                     │
│  image preview row (only if image attached):       │
│   [ 80×80 thumbnail with × delete badge ]           │
│                                                     │
│  tag chips:                                         │
│   [ #数学 × ]  [ #高考 × ]  + 添加标签              │
├─ action bar ──────────────────────────────────────┤
│  [📷 添加图片]                                      │  image picker trigger
└────────────────────────────────────────────────────┘
```

### Fields (matching dx-api validators)

- `content` — required, 1–2000 chars after trim. Submit button disabled while empty after trim.
- `image_url` — optional, single image. dx-api column is `text NULL`, single-valued.
- `tags` — optional, max 5, each ≤20 chars. Tag input commits on Enter only (not on whitespace) since Chinese IME may emit spaces. Duplicate tags suppressed.

### Image picker flow

1. Tap 📷 → `wx.chooseMedia({ count: 1, mediaType: ['image'], sizeType: ['compressed'], sourceType: ['album','camera'] })`.
2. Receive `tempFilePath` and `tempFile.size`.
3. Local validation **first**: size ≤ 2 MB, extension is `.jpg`/`.jpeg`/`.png`. If oversize or wrong type, toast and abort (no upload attempted).
4. Set `uploading: true`. Show local thumbnail immediately (using `tempFilePath`) so the user has visual feedback while the upload runs.
5. `wx.uploadFile({ url: config.apiBaseUrl + '/api/uploads/images', filePath, name: 'file', formData: { role: 'post-image' }, header: { Authorization: 'Bearer ' + getToken() } })`.
6. Parse `JSON.parse(uploadRes.data)` as `{code, data: {url}}`. On `code === 0`, store `image_url = data.url`, set `uploading: false`. On non-zero, toast `body.message` and clear the local thumbnail.
7. Thumbnail has an × delete badge that clears both `image_url` and the local preview.
8. Submit is blocked while `uploading: true` so we never post before the URL resolves.

### Submit flow

1. Validate locally (length, tag count, tag length). First failure surfaces as a toast.
2. Close popup, `wx.showLoading({ title: '发布中…' })`.
3. `api.post<Post>('/api/posts', { content, image_url, tags })`.
4. On success: `wx.hideLoading()`, `wx.showToast('已发布')`, `triggerEvent('post-created', { post })`. The feed page (parent) handles the event by prepending the new post to its in-memory list — no full refetch.
5. On error: `wx.hideLoading()`, re-open popup with previous content preserved, surface error toast.

(`composer-popup` is a Component nested in the feed Page, so the `triggerEvent` → page-level `bind:postcreated` handler is the right plumbing — eventChannel is for cross-page communication only.)

### State reset

Form state clears only on successful submit or explicit cancel (with confirm dialog if any field is non-empty).

## UI: post-detail page (`pages/community/detail/detail`)

### Layout

```
status bar spacer
┌─ custom nav: ← 返回    "帖子"     ────────────────┐
├─ post-block ─────────────────────────────────────┤
│  avatar • nickname • 完整时间                      │
│  full content (no truncation, no clamp)            │
│  image (full width, tap to wx.previewImage)        │
│  #tags                                              │
│  ─────────────                                      │
│  ♥ N    💬 N    🔖    [关注/已关注]                 │
├─ "评论 N" section header ────────────────────────┤
│  ┌─ comment-item ────────────────────────┐         │
│  │ avatar • nickname • 1小时前             │         │
│  │ content                                 │         │
│  │ [回复]                                  │         │
│  │ ┌─ replies (indented, left border) ──┐  │         │
│  │ │ avatar • nickname • 30分钟前         │  │         │
│  │ │ content                              │  │         │
│  │ └──────────────────────────────────────┘  │         │
│  └────────────────────────────────────────────┘         │
│  ... infinite list, van-empty if zero          │         │
├─ sticky input bar (safe-area-inset-bottom) ──────┤
│  [ placeholder ▾ ]                       [发送]   │
└────────────────────────────────────────────────────┘
```

### Lifecycle

- `onLoad(query)` → read `id` from `query`, `Promise.all([loadPost(), loadComments(true)])`.
- `onReachBottom` → paginates comments if `hasMore && !loading`.
- `onShow` → re-checks auth.

### Behaviors

- **Post-level interactions** (like, bookmark, follow) — same optimistic logic as feed card.
- **State propagation back to feed** — uses `wx.getOpenerEventChannel().emit('post-updated', { id, patch })` on each interaction, where `patch: Partial<Post>` carries only the fields that changed (e.g. `{ is_liked: true, like_count: 13 }`). Feed registers the listener in `onShow` (via `getCurrentPages` lookup of the opener channel set up when navigating to detail) and merges patches into its in-memory list. Cheaper than re-fetching. If the channel is unavailable (e.g. detail entered via deep link), the page silently no-ops.
- **Image preview** — tap post image → `wx.previewImage({ urls: [absoluteUrl] })` where `absoluteUrl = config.apiBaseUrl + post.image_url`.
- **Reply mode**:
  - Default placeholder `"说点什么…"`. Send with `parent_id: null` (top-level comment).
  - Tap "回复" on a top-level comment → placeholder swaps to `"回复 @{nickname}："`. `replyingTo: { commentId, nickname }` is set in page data. Send with `parent_id: commentId`.
  - Cancel reply: tap a small "×" inline next to the placeholder, or tap-outside on the comments area.
  - "回复" affordance is hidden on reply items (those with `parent_id !== null`) — server rejects nested replies with `ErrNestedReply`, UI prevents reaching that state.
- **Comment send** — disabled while empty after trim. Optimistic insert with `pending: true`, replaced by server response. On error, remove + toast. Bumps `post.comment_count` locally and propagates to feed via eventChannel.
- **Empty state** — `van-empty description="暂无评论,来抢沙发"`.
- **Keyboard handling** — input bar uses `cursor-spacing` and `adjust-position={{true}}` so the bar lifts above the keyboard.

### Component split

- **post-block** — full post (no truncation), image preview wired, action row.
- **comment-item** — one top-level comment plus its replies block. Emits `reply` up to the page (which sets `replyingTo` and focuses the input).

## Icon registration

Add to `dx-mini/scripts/build-icons.mjs` ICONS array, then run `npm run build:icons` to regenerate `components/dx-icon/icons.ts`:

```
['heart',           'heart'],
['message-circle',  'message-circle'],
['bookmark',        'bookmark'],
['user-plus',       'user-plus'],
['user-check',      'user-check'],
['send',            'send'],
['image',           'image'],
['plus',            'plus'],
```

(Phase 2 will add `more-horizontal` when we land edit/delete UI — registered then, not now.)

Filled vs outline for like / bookmark active state: rendered SVGs use `currentColor` for stroke. Active state applies CSS `fill: currentColor` on the inner `<svg>` to fill the shape — no separate icon file needed. The renderer in `components/dx-icon/index.ts` already supports `color` prop binding; the fill toggle happens in the consuming component via WXSS.

## Theming & styling

Match dx-mini conventions exactly:
- `van-config-provider theme="{{theme}}"` wraps each page; reads `app.globalData.theme`.
- CSS custom properties from `app.wxss`: `--primary` (#0d9488 light / #14b8a6 dark), `--bg-page`, `--bg-card`, `--border-color`, `--text-primary`, `--text-secondary`, `--destructive`.
- Page padding-bottom `140rpx` for tab-bar clearance on the feed; detail page uses `safe-area-inset-bottom` on the sticky input bar instead.
- Card borders `1px solid var(--border-color)`, radius `12px`.
- Tag chip: teal background `var(--primary-light)`, teal text `var(--primary)`, radius `4px`, padding `2px 8px`.
- Like color when active: `#ef4444` (matches `--destructive`). Bookmark color when active: `var(--primary)`.

## Adjacent fix included in Phase 1

**`pages/me/profile-edit/profile-edit.ts`** — current `wx.uploadFile` call to `/api/uploads/images` does not pass the `role` form field that dx-api requires (`upload_request.go` validates against an allowlist). Add `formData: { role: 'user-avatar' }` so the avatar upload validates server-side. Same change pattern as the new post-image upload. Separate commit, scoped to that one line; manually verified by uploading a new avatar after the change.

## Lint / TS hygiene

- Strict TS throughout. No new `// @ts-ignore`. Tolerate the existing `Component({methods})` `this` typing limitation (a known WeChat SDK constraint already accepted by the codebase) — do not introduce *new* tsc errors beyond that pattern.
- No `console.log` in production code.
- WXML / WXSS: no `?.` or `??` (existing dx-mini constraint — neither WXS nor the WXML expression parser accept them).
- Imports ordered: stdlib → wx-typings → utils → components.
- Per-page `*.json` declares only the components actually used.

## Verification before claiming complete

WeChat Developer Tools, Glass-easel rendering:
- 编译 cleanly with zero TS errors and zero WXML errors.
- Smoke each tab in the feed: 最新 / 热门 / 关注 / 收藏 → scroll → 下拉刷新 → 上拉加载更多.
- Compose flow:
  - text-only post
  - text + image post
  - text + tags post
  - text + image + tags post
  - cancel mid-compose with unsaved content (confirm dialog appears)
- Detail flow: open post → like → bookmark → follow → comment → reply → load more comments → previewImage on the post image.
- Auth flow: log out → tap 社区 tab → reLaunches to login.
- Dark mode toggle on each page (uses existing theme switcher in `pages/me/me`).
- Profile-edit avatar upload still works after the `role` field fix.

## Phase 2 follow-up (after Phase 1 lands)

Separate workstream, not part of Phase 1 commits. Tracked here for record.

### Phase 2.1 — Fix dx-web `bookmarks` tab name bug

**Problem:** `dx-web/src/features/web/community/types/post.ts:33` declares `FeedTab = "latest" | "hot" | "following" | "bookmarks"`, but `dx-api/app/services/api/post_service.go:81` only accepts `"bookmarked"`. dx-web's 收藏 tab silently falls through to `latest` — broken since shipped.

**Fix:** Rename the dx-web type to `"bookmarked"` (single source of truth, matching the wire value). Search-and-replace `"bookmarks"` → `"bookmarked"` in:
- `dx-web/src/features/web/community/types/post.ts`
- `dx-web/src/features/web/community/components/feed-tabs.tsx` (TABS array key)
- Any test fixtures or stories.

**Verification:** Bookmark a few posts on dx-web, switch to 收藏 tab, confirm only bookmarked posts appear. Smoke other tabs to ensure no regression.

### Phase 2.2 — Edit / delete UI on dx-web and dx-mini

dx-api supports `PUT /api/posts/{id}`, `DELETE /api/posts/{id}`, `PUT /api/posts/{id}/comments/{cid}`, `DELETE /api/posts/{id}/comments/{cid}` with ownership checks. UI exposure on both clients.

**Affordances:**
- **dx-web** — `more-horizontal` icon button on `PostCard` and `CommentItem`, opens a small dropdown menu with "编辑" / "删除". Edit reuses the existing `CreatePostDialog` / `CommentInput` in edit mode. Delete shows confirm dialog.
- **dx-mini** — `more-horizontal` icon button on the post-detail `post-block` (not on the feed card — keeps the feed card minimal) and on each `comment-item`. Opens `wx.showActionSheet({ itemList: ['编辑', '删除'] })`. Edit re-opens the composer-popup pre-filled. Delete shows `wx.showModal({ title: '确认删除', cancelText: '取消', confirmText: '删除', confirmColor: '#ef4444' })`.

**Visibility rule:** affordance shown only when the post / comment author ID matches the current user's ID (from `getUserId()` on mini, from auth context on web).

**Server already enforces** ownership via `403 Forbidden` — UI is just convenience and to avoid showing a control that always fails.

**Verification:**
- Author edits / deletes own post and comment → succeeds.
- Other user attempts via crafted request → server still rejects (server-side enforcement is the security boundary, UI is only a convenience).
- Soft-delete behavior on posts (sets `is_active=false`) is unchanged — list endpoints already filter.

## Open questions / accepted tradeoffs

- **Hot-tab cursor is a numeric offset on dx-api.** It's offset-pagination dressed as a cursor and can drift if posts shift rank during user scroll. dx-web has the same tradeoff. Accepted, no action.
- **State propagation feed ↔ detail** uses `wx.getOpenerEventChannel`. If the user navigates A → B → A back, the channel still works since detail emits to its opener. If we ever add deep-linking to detail (e.g., from a notification), the eventChannel won't exist; the detail page should swallow the `undefined` channel without throwing — covered in code.
- **Avatar fallback** uses initial + deterministic color. Port `dx-web/src/lib/avatar.ts:getAvatarColor` to `pages/community/utils/avatar.ts` (or inline) so feed and detail share one source of truth. The dx-mini codebase doesn't already have this helper.
