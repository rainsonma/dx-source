# 学习 page back-button + three quick-link pages bring-to-parity + cleanup

**Date:** 2026-05-01
**Scope:** dx-mini only — `pages/learn/learn.{wxml,ts,wxss}`, full rewrite of `pages/learn/{mastered,unknown,review}/*`, deletion of `pages/me/study/`, one line removed from `app.json`.
**Stakes:** UI polish + parity work on already-functional pages. No API changes, no DB changes, no auth changes.

## Goal

The 学习 page (a sub-page reached from the home "学习" tile, NOT a tab) recently got a "我的学习进度" card and a custom-nav lift, but it still has no back affordance — a user lands on it from home and the only escape is the system back gesture or the WeChat capsule's "X". Add a left-chevron back button + centered "学习" title to its top row.

While we're in the area, the three quick-link destination pages (`pages/learn/mastered`, `pages/learn/unknown`, `pages/learn/review`) are functional but lag behind the dx-web reference (`/hall/mastered`, `/hall/unknown`, `/hall/review`) on three counts: no 3-tile stats header, no per-row delete, no timestamps. They also still use the WeChat default nav rather than the custom-nav-with-back-chevron pattern the rest of the page family now uses. Bring all three to dx-web parity in the same batch.

Finally, `pages/me/study` is a "敬请期待" placeholder registered in `app.json` but reachable from nowhere. Delete it.

A side discovery during exploration: `learn.wxml:14` shows `本周 +{{unknownStats.thisWeek || 0}}` but the unknown stats endpoint returns `{ total, today, lastThreeDays }` — no `thisWeek` field, so the figure always renders as `本周 +0`. Fix in this batch since we're touching nearby code.

## Current state

```
[--status-bar-height + 88rpx empty padding]   ← page-container padding-top
[我的学习进度 card]                              ← from prior PR
[stat row   已掌握 / 生词本 / 待复习]
[quick links   已掌握词库 / 生词本词库 / 待复习队列]
```

Tapping a quick link navigates to a destination page with the WeChat default nav bar, a single-line `{{items.length}} 词` count + 批量删除 button, and a flat list of word items (no per-row delete, no timestamps).

`pages/me/study` is reachable from nowhere; `app.json:23` registers it.

## Target state

```
[--status-bar-height empty]                            ← only system status bar cleared
[88rpx nav-bar  ◁ ─────────  学习  ─────────  (capsule)]
[我的学习进度 card]
[stat row   已掌握 / 生词本 / 待复习]                    ← bonus: 生词本 sub-text fixed to 今日 +N
[quick links]
```

Each quick-link destination page mirrors that shell:

```
[--status-bar-height empty]
[88rpx nav-bar  ◁ ─────────  已掌握/生词本/待复习  ─────────  (capsule)]
[3-tile stats header  total / time-window-1 / time-window-2]
[action bar  {{items.length}} 词 + 批量删除]
[word list   each row has trash icon at the right + relative-date underneath]
```

`pages/me/study` is gone; `app.json` no longer references it.

## Architecture decisions

### 1. Back chevron uses `wx.navigateBack()` — no `switchTab`

`pages/learn/learn` is in `app.json` `pages[]` but NOT in `tabBar.list`. The 5 tabs are home/games/leaderboard/notices/me. So 学习 is a regular sub-page reached via `wx.navigateTo` (home `goStudy()` already does this correctly). `wx.navigateBack()` works as expected.

### 2. Centered title + left chevron — title uses absolute centering to ignore the capsule

The WeChat capsule occupies the upper-right ~88rpx. A flex-centered title would be visually pushed left by the capsule's reservation. Use `position: absolute; left: 50%; transform: translateX(-50%);` on the title so it sits in the visual center of the screen regardless of the capsule.

### 3. Reuse `.learn-stat-card.{teal,amber,purple}` styles for the new stats headers

The parent learn page already has the teal/amber/purple tinted-tile pattern with light + dark variants. Don't reinvent. Each destination page declares the same WXSS rules locally (no shared style file — three is too few to warrant extraction; YAGNI).

### 4. Per-row delete uses the existing `trash-2` icon + `wx.showModal`

`trash-2` is already in the `dx-icon` inventory (verified). The confirm flow matches the existing batch-delete handler — `wx.showModal` → `api.delete('/api/tracking/<X>', { ids: [id] })` on confirm → splice from local `items`. No new infrastructure.

### 5. Timestamps pre-computed in `loadItems`, not WXML

Per the project's `feedback_dx_mini_wxml_no_method_calls.md` memory, WXML can't call `formatRelativeDate` directly. Pre-compute `timeText` on each item as it arrives from the API, store on the item shape used by WXML.

### 6. `unknownStats` gets its own type

Today's `learn.ts` types both master and unknown stats as the same `Stats` interface, which is wrong for unknown. Split into `MasterStats` (or rename to `Stats` continuing existing usage for master) and `UnknownStats`.

### 7. No backend changes

The list endpoints (`/api/tracking/master`, `/api/tracking/unknown`, `/api/tracking/review`), the stats endpoints (`/api/tracking/master/stats`, etc.), and the delete endpoints (`DELETE /api/tracking/<X>` with `{ids: [...]}` body) all exist already and have the right shapes. Nothing to add.

### 8. Search bar — out of scope

dx-web has a `PageTopBar` with a search input. On mini, the data sizes (per-user mastered/unknown/review words) don't justify a search input as core. Skip — easy to add later.

## Per-page contract

### `pages/learn/learn` — minimal additions

**Insertions only; no removals beyond the WXSS padding-top adjustment.**

#### `learn.wxml`

Insert immediately under the existing `<van-config-provider theme="{{theme}}">`/`<view class="page-container ...">` opening:

```xml
<view class="nav-bar">
  <view class="back-btn" bind:tap="goBack">
    <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
  </view>
  <text class="nav-title">学习</text>
</view>
```

(This goes ABOVE the existing `<van-loading wx:if="{{loading}}" ... />` block.)

Also change the existing line:

```xml
<text class="lsc-sub">本周 +{{unknownStats.thisWeek || 0}}</text>
```

to:

```xml
<text class="lsc-sub">今日 +{{unknownStats.today || 0}}</text>
```

#### `learn.ts`

Add a new `goBack` method:

```ts
goBack() { wx.navigateBack() },
```

Replace the misleading shared `Stats` typing for unknown by adding a dedicated interface:

```ts
interface UnknownStats { total: number; today: number; lastThreeDays: number }
```

Update the `data` declaration:

```ts
unknownStats: null as UnknownStats | null,
```

And the `loadAll` line:

```ts
api.get<UnknownStats>('/api/tracking/unknown/stats'),
```

Master and review stats are already correctly typed; no change there.

#### `learn.wxss`

Change `.page-container`'s top padding so the new `.nav-bar` row OCCUPIES the 88rpx that used to be empty:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: var(--status-bar-height, 20px);
  padding-bottom: 100rpx;
}
```

(Was `calc(var(--status-bar-height, 20px) + 88rpx)`.)

Append:

```css
.nav-bar {
  position: relative;
  height: 88rpx;
  display: flex;
  align-items: center;
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
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
```

### `pages/learn/mastered` — full rewrite

#### `mastered.json`

```json
{
  "navigationStyle": "custom",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-checkbox": "@vant/weapp/checkbox/index",
    "van-checkbox-group": "@vant/weapp/checkbox-group/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-loading": "@vant/weapp/loading/index",
    "dx-icon": "/components/dx-icon/index"
  }
}
```

(`navigationBarTitleText` removed; `dx-icon` registered; `van-button` dropped — was unused.)

#### `mastered.ts`

```ts
import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface MasterItemData {
  id: string
  contentItem: TrackingContentData | null
  gameName: string | null
  masteredAt: string | null
  createdAt: string
}
interface MasterItemView extends MasterItemData {
  timeText: string
}
interface MasterStats { total: number; thisWeek: number; thisMonth: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    loading: true,
    stats: { total: 0, thisWeek: 0, thisMonth: 0 } as MasterStats,
    items: [] as MasterItemView[],
    nextCursor: '',
    hasMore: false,
    selectedIds: [] as string[],
    selectMode: false,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const theme = app.globalData.theme
    this.setData({
      statusBarHeight: sys.statusBarHeight || 20,
      theme,
      primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
    })
    this.loadAll(true)
  },
  onPullDownRefresh() {
    this.loadAll(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadAll(false)
  },
  async loadAll(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? '?cursor=' + cursor + '&limit=20' : '?limit=20'
    const results = await Promise.allSettled([
      api.get<PaginatedData<MasterItemData>>('/api/tracking/master' + qs),
      reset ? api.get<MasterStats>('/api/tracking/master/stats') : Promise.resolve(this.data.stats),
    ])

    const listOk = results[0].status === 'fulfilled'
    const list = listOk ? results[0].value : { items: [] as MasterItemData[], nextCursor: '', hasMore: false }
    const newViews: MasterItemView[] = list.items.map((it) => ({
      ...it,
      timeText: formatRelativeDate(it.masteredAt || it.createdAt),
    }))

    const stats = results[1].status === 'fulfilled' ? results[1].value : this.data.stats

    this.setData({
      loading: false,
      items: reset ? newViews : [...this.data.items, ...newViews],
      nextCursor: list.nextCursor,
      hasMore: list.hasMore,
      stats,
    })

    if (results.some((r) => r.status === 'rejected')) {
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goBack() { wx.navigateBack() },
  toggleSelectMode() {
    this.setData({ selectMode: !this.data.selectMode, selectedIds: [] })
  },
  onSelectChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ selectedIds: e.detail as string[] })
  },
  onDeleteOne(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (!id) return
    wx.showModal({
      title: '确认删除',
      content: '删除该词？',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids: [id] })
          this.setData({
            items: this.data.items.filter((i) => i.id !== id),
            stats: { ...this.data.stats, total: Math.max(0, this.data.stats.total - 1) },
          })
          wx.showToast({ title: '已删除', icon: 'none' })
        } catch {
          wx.showToast({ title: '删除失败', icon: 'none' })
        }
      },
    })
  },
  async bulkDelete() {
    const ids = this.data.selectedIds
    if (ids.length === 0) return
    wx.showModal({
      title: '确认删除',
      content: '删除 ' + ids.length + ' 个词？',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await api.delete('/api/tracking/master', { ids })
          this.setData({
            items: this.data.items.filter((i) => !ids.includes(i.id)),
            selectedIds: [],
            selectMode: false,
            stats: { ...this.data.stats, total: Math.max(0, this.data.stats.total - ids.length) },
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

(Note the `Promise.allSettled` for the initial load — list and stats both fetched in parallel; failure of either still updates the other. Subsequent `loadMore` calls only fetch the list, since stats don't change between page-loads.)

#### `mastered.wxml`

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="nav-bar">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">已掌握</text>
    </view>

    <view class="content">
      <view class="stat-card-row">
        <view class="learn-stat-card teal">
          <text class="lsc-value">{{stats.total}}</text>
          <text class="lsc-label">已掌握总数</text>
        </view>
        <view class="learn-stat-card amber">
          <text class="lsc-value">{{stats.thisWeek}}</text>
          <text class="lsc-label">本周掌握</text>
        </view>
        <view class="learn-stat-card purple">
          <text class="lsc-value">{{stats.thisMonth}}</text>
          <text class="lsc-label">本月掌握</text>
        </view>
      </view>

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

      <van-loading wx:if="{{loading && items.length === 0}}" size="30px" color="{{primaryColor}}" class="center-loader" />
      <van-empty wx:if="{{!loading && items.length === 0}}" description="暂无已掌握的词" />

      <van-checkbox-group value="{{selectedIds}}" bind:change="onSelectChange">
        <view class="word-list">
          <view wx:for="{{items}}" wx:key="id" class="word-item">
            <van-checkbox wx:if="{{selectMode}}" name="{{item.id}}" />
            <view class="word-body">
              <block wx:if="{{item.contentItem}}">
                <text class="word-content">{{item.contentItem.content}}</text>
                <text wx:if="{{item.contentItem.translation}}" class="word-trans">{{item.contentItem.translation}}</text>
              </block>
              <view class="word-meta">
                <text wx:if="{{item.gameName}}" class="word-game">来自：{{item.gameName}}</text>
                <text wx:if="{{item.timeText}}" class="word-time">{{item.timeText}}</text>
              </view>
            </view>
            <view wx:if="{{!selectMode}}" class="word-delete" data-id="{{item.id}}" bind:tap="onDeleteOne">
              <dx-icon name="trash-2" size="18px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
            </view>
          </view>
        </view>
      </van-checkbox-group>

      <view wx:if="{{loading && items.length > 0}}" class="load-more">
        <van-loading size="20px" color="{{primaryColor}}" />
      </view>
    </view>
  </view>
</van-config-provider>
```

(The trash icon is hidden in select mode to avoid two competing delete affordances.)

#### `mastered.wxss`

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: var(--status-bar-height, 20px);
  padding-bottom: 40px;
}
.nav-bar {
  position: relative;
  height: 88rpx;
  display: flex;
  align-items: center;
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
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.content { padding: 0 16px; }
.center-loader { display: flex; justify-content: center; padding: 40px; }

.stat-card-row { display: flex; gap: 10px; margin-bottom: 16px; }
.learn-stat-card {
  flex: 1;
  border-radius: 12px;
  padding: 14px 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.learn-stat-card.teal { background: rgba(13, 148, 136, 0.12); }
.learn-stat-card.amber { background: rgba(245, 158, 11, 0.12); }
.learn-stat-card.purple { background: rgba(99, 102, 241, 0.12); }
.dark .learn-stat-card.teal { background: rgba(20, 184, 166, 0.12); }
.dark .learn-stat-card.amber { background: rgba(245, 158, 11, 0.10); }
.dark .learn-stat-card.purple { background: rgba(99, 102, 241, 0.10); }
.lsc-value { font-size: 22px; font-weight: 700; color: var(--text-primary); }
.lsc-label { font-size: 12px; color: var(--text-primary); font-weight: 500; }

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: var(--bg-card);
  border-radius: 10px;
  border: 1px solid var(--border-color);
  margin-bottom: 12px;
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
  color: var(--van-danger-color);
  border-color: var(--van-danger-color);
}

.word-list { display: flex; flex-direction: column; gap: 8px; }
.word-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card);
  border-radius: 10px;
  border: 1px solid var(--border-color);
  padding: 12px;
}
.word-body { flex: 1; min-width: 0; }
.word-content { font-size: 16px; font-weight: 600; color: var(--text-primary); display: block; }
.word-trans { font-size: 13px; color: var(--text-secondary); display: block; margin-top: 2px; }
.word-meta { display: flex; gap: 10px; margin-top: 4px; align-items: center; }
.word-game { font-size: 11px; color: var(--text-secondary); }
.word-time { font-size: 11px; color: var(--text-secondary); }
.word-delete {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}
.load-more { display: flex; justify-content: center; padding: 16px; }
```

### `pages/learn/unknown` — same shape, three differences

Identical to `mastered` except:

1. **`unknown.json`** — `navigationBarTitleText` removed (not strictly necessary but tidier — title is now in WXML), `navigationStyle: "custom"`, same `usingComponents`. Note `unknown.json` was already updated in the prior PR to set the title to "生词本" — that line gets removed entirely once the WXML carries the title.

2. **`unknown.ts`** — interface `UnknownStats { total: number; today: number; lastThreeDays: number }`; endpoint `/api/tracking/unknown/stats`; list endpoint `/api/tracking/unknown`; delete endpoint `/api/tracking/unknown`. `MasterItemData` becomes `UnknownItemData` with field `createdAt` (no `masteredAt`); `timeText` derives from `createdAt` only.

3. **`unknown.wxml`** — title text is `生词本`; stat tiles render `stats.total / stats.today / stats.lastThreeDays` with labels `全部生词 / 今日添加 / 最近三天`; empty-state description is `生词本里还空着` (preserves the prior PR's copy alignment).

4. **`unknown.wxss`** — identical to mastered.wxss.

### `pages/learn/review` — same shape, three differences

Identical to `mastered` except:

1. **`review.json`** — `navigationBarTitleText` removed, `navigationStyle: "custom"`, same `usingComponents`.

2. **`review.ts`** — interface `ReviewStats { pending: number; overdue: number; reviewedToday: number }`; endpoint `/api/tracking/review/stats`; list endpoint `/api/tracking/review`; delete endpoint `/api/tracking/review`. Item interface uses `createdAt` only; `timeText` derives from `createdAt`.

3. **`review.wxml`** — title `待复习`; stat tiles render `stats.pending / stats.overdue / stats.reviewedToday` with labels `待复习 / 逾期 / 今日复习`; empty-state description `暂无待复习的词`.

4. **`review.wxss`** — identical to mastered.wxss.

### `pages/me/study` — delete

- Delete `dx-mini/miniprogram/pages/me/study/study.json`
- Delete `dx-mini/miniprogram/pages/me/study/study.ts`
- Delete `dx-mini/miniprogram/pages/me/study/study.wxml`
- Delete `dx-mini/miniprogram/pages/me/study/study.wxss`
- Delete the directory `dx-mini/miniprogram/pages/me/study/`

### `app.json` — remove the registration

Remove the line `"pages/me/study/study",` from `dx-mini/miniprogram/app.json` `pages[]` (currently line 23). Adjust nothing else.

## Edge cases

| Case | Behavior |
|---|---|
| User taps back chevron from `learn` arrived via `goStudy` | `wx.navigateBack()` returns to home tab. Standard behavior. |
| User taps back chevron from `mastered`/`unknown`/`review` | `wx.navigateBack()` returns to learn page. Standard. |
| Stats fetch fails but list succeeds | List renders normally; stats keep their default `0` values (or previous values on refresh). Single toast surfaces "加载失败". |
| List fetch fails but stats succeed | Empty list state; stats render. Single toast. |
| Both fail | Both default; single toast. |
| Per-row delete on the last item of a page | Item disappears; if the list is now empty, `wx.stopPullDownRefresh` is unaffected (it only fires on user-pulled refresh); stats `total` decrements by 1. |
| Bulk delete | Same flow; `total` decrements by `ids.length`. |
| User toggles select mode mid-scroll | Trash icon hides on each row; checkbox appears. Works because both visibility flags depend on `selectMode`. |
| User deletes an item then immediately reaches bottom | `loadItems(false)` proceeds normally — `nextCursor` and `hasMore` are unchanged by the local splice. (Server-side pagination is independent of client-side filtering — there may be one-row gap or duplicate possibility in extreme cases, but acceptable for this scale.) |
| Theme switch while page is open | `theme` re-reads on `onShow`; CSS variables propagate. Trash icon color updates on next `setData` (or accept the slight delay — not worth wiring full reactivity for an edge case). |

## Verification

Must all pass before declaring done:

1. `cd dx-mini && npx tsc --noEmit -p tsconfig.json` — total error count unchanged from baseline (143 from prior batch); zero errors attributable to any of the touched files.
2. `cd dx-mini && npm run build:icons` — clean run; static WXML scan accepts the `chevron-left` and `trash-2` references (both already in inventory).
3. WeChat DevTools manual smoke (use 预览, NOT 真机调试):
   - Open 学习 from home → back chevron + "学习" title visible at top, capsule on right doesn't overlap.
   - Tap back chevron → returns to home tab.
   - Tap each of the three quick links → destination page renders with custom-nav row, back chevron, correct centered title, 3-tile stats header, list with per-row trash icon, timestamp under each word.
   - Tap trash on a row → confirm modal → confirm → row disappears, total decrements.
   - Toggle 批量删除 mode → trash icons hide, checkboxes appear; pick some → 删除(N) appears → confirm → rows disappear.
   - Pull down to refresh → list + stats reload.
   - Scroll to bottom → next page loads (if `hasMore`).
   - Switch to dark mode → all chrome readable.
   - Switch tabs out and back to home → home still works (no tabBar regression).
   - Confirm `pages/me/study` is unregistered (DevTools should not list it; navigating to its path manually should fail).
4. The `本周 +0` figure on the parent learn page's `生词本` stat tile is now `今日 +N`.

## Out of scope

- No PageTopBar / search box on the three quick-link pages.
- No backend changes.
- No new icons added to inventory.
- No deeper "unknown → vocab" rename in `play.wxml` or elsewhere.
- No extraction of shared `<dx-tracking-list>` Component — three near-identical pages is too few; revisit if a 4th appears.
- No conversion of `pages/learn/learn` to a tab page (and no removal from `pages[]` — it stays as a sub-page reachable from home).
- No change to `home.ts goStudy()` — already correct.

## File list

**dx-mini (10 changed, 4 deleted, 1 directory removed):**

| File | Operation |
|---|---|
| `miniprogram/pages/learn/learn.wxml` | modify (insert nav-bar; fix one label) |
| `miniprogram/pages/learn/learn.ts` | modify (add `goBack`; split `UnknownStats` type; update endpoint typing) |
| `miniprogram/pages/learn/learn.wxss` | modify (adjust top-pad; add `.nav-bar`/`.back-btn`/`.nav-title`) |
| `miniprogram/pages/learn/mastered/mastered.json` | rewrite |
| `miniprogram/pages/learn/mastered/mastered.ts` | rewrite |
| `miniprogram/pages/learn/mastered/mastered.wxml` | rewrite |
| `miniprogram/pages/learn/mastered/mastered.wxss` | rewrite |
| `miniprogram/pages/learn/unknown/unknown.json` | rewrite |
| `miniprogram/pages/learn/unknown/unknown.ts` | rewrite |
| `miniprogram/pages/learn/unknown/unknown.wxml` | rewrite |
| `miniprogram/pages/learn/unknown/unknown.wxss` | rewrite |
| `miniprogram/pages/learn/review/review.json` | rewrite |
| `miniprogram/pages/learn/review/review.ts` | rewrite |
| `miniprogram/pages/learn/review/review.wxml` | rewrite |
| `miniprogram/pages/learn/review/review.wxss` | rewrite |
| `miniprogram/pages/me/study/study.json` | delete |
| `miniprogram/pages/me/study/study.ts` | delete |
| `miniprogram/pages/me/study/study.wxml` | delete |
| `miniprogram/pages/me/study/study.wxss` | delete |
| `miniprogram/pages/me/study/` | rmdir |
| `miniprogram/app.json` | modify (remove `pages/me/study/study` line) |

**dx-api:** no changes.
