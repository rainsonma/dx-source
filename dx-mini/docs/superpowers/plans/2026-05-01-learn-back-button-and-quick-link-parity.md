# 学习 page back-button + three quick-link pages bring-to-parity + cleanup — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a back chevron + centered "学习" title to the dx-mini 学习 sub-page; bring the three destination pages (`pages/learn/{mastered,unknown,review}`) up to dx-web parity (custom-nav + back button + 3-tile stats header + per-row delete + relative-date timestamp); fix a pre-existing `unknownStats.thisWeek` rendering bug on the parent learn page; delete the orphaned `pages/me/study/` placeholder.

**Architecture:** dx-mini-only changes. The back chevron uses the standard `wx.navigateBack()` (`pages/learn/learn` is a sub-page, not a tab — verified). The three destination pages each get a full rewrite that mirrors the parent learn page's custom-nav shell, declares its own page-specific stats interface (master/unknown/review have three different shapes), reuses the existing `.learn-stat-card.{teal,amber,purple}` tinted-tile pattern for the new stats header, and adds an inline trash-icon per row alongside the existing batch-delete flow. No backend changes; all endpoints already exist.

**Tech Stack:** WeChat Mini Program (native + TypeScript + WXML + WXSS), Vant Weapp 1.11.x, the project's `<dx-icon>` Lucide SVG component. Both `chevron-left` and `trash-2` are already in the icon inventory — no `npm run build:icons` regeneration needed.

**Reference spec:** `dx-mini/docs/superpowers/specs/2026-05-01-learn-back-button-and-quick-link-parity-design.md`

---

## File Structure

**dx-mini (10 changed, 4 deleted, 1 directory removed):**

| File | Operation | Task |
|---|---|---|
| `miniprogram/pages/learn/learn.ts` | modify (add `UnknownStats`; type endpoint; add `goBack`) | 1, 2 |
| `miniprogram/pages/learn/learn.wxml` | modify (fix label; insert nav-bar) | 1, 2 |
| `miniprogram/pages/learn/learn.wxss` | modify (adjust padding-top; add nav-bar styles) | 2 |
| `miniprogram/pages/learn/mastered/mastered.{json,ts,wxml,wxss}` | rewrite | 3 |
| `miniprogram/pages/learn/unknown/unknown.{json,ts,wxml,wxss}` | rewrite | 4 |
| `miniprogram/pages/learn/review/review.{json,ts,wxml,wxss}` | rewrite | 5 |
| `miniprogram/pages/me/study/{study.json,study.ts,study.wxml,study.wxss}` + dir | delete | 6 |
| `miniprogram/app.json` | modify (drop one line from `pages[]`) | 6 |

The plan ships in 7 tasks. Tasks 1-2 patch the parent learn page (smallest blast radius first). Tasks 3-5 rewrite the three destination pages (each independent, but ordered mastered → unknown → review for predictability). Task 6 is the cleanup. Task 7 is the manual smoke test.

---

## Task 1: Fix `unknownStats` rendering bug on the learn page

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.ts`
- Modify: `dx-mini/miniprogram/pages/learn/learn.wxml`

The parent learn page currently shows `本周 +{{unknownStats.thisWeek || 0}}` on the middle stat card, but the `/api/tracking/unknown/stats` endpoint returns `{ total, today, lastThreeDays }` — there is no `thisWeek` field, so the figure always renders as `本周 +0`. Fix the type and the WXML label.

- [ ] **Step 1: Capture the tsc baseline**

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`

Record this number — the post-edit count must match.

- [ ] **Step 2: Add the `UnknownStats` interface to `learn.ts`**

In `dx-mini/miniprogram/pages/learn/learn.ts`, locate the existing two interface declarations near the top:

```ts
interface Stats { total: number; thisWeek: number; thisMonth: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }
```

Insert a new line immediately after `Stats` and before `ReviewStats`:

```ts
interface Stats { total: number; thisWeek: number; thisMonth: number }
interface UnknownStats { total: number; today: number; lastThreeDays: number }
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }
```

- [ ] **Step 3: Re-type the `unknownStats` data field**

In the same file, locate inside the `Page({ data: { ... } })` block:

```ts
    unknownStats: null as Stats | null,
```

Change to:

```ts
    unknownStats: null as UnknownStats | null,
```

- [ ] **Step 4: Re-type the `unknownStats` API call in `loadAll`**

In the same file, inside the `loadAll` function's `Promise.allSettled` array, locate:

```ts
      api.get<Stats>('/api/tracking/unknown/stats'),
```

Change to:

```ts
      api.get<UnknownStats>('/api/tracking/unknown/stats'),
```

- [ ] **Step 5: Fix the WXML label**

In `dx-mini/miniprogram/pages/learn/learn.wxml`, locate the `learn-stat-card amber` block (around line 13-16). It currently reads:

```xml
        <view class="learn-stat-card amber" bind:tap="goUnknown">
          <text class="lsc-value">{{unknownStats.total || 0}}</text>
          <text class="lsc-label">生词本</text>
          <text class="lsc-sub">本周 +{{unknownStats.thisWeek || 0}}</text>
        </view>
```

Change the third `<text>` line to:

```xml
          <text class="lsc-sub">今日 +{{unknownStats.today || 0}}</text>
```

So the block becomes:

```xml
        <view class="learn-stat-card amber" bind:tap="goUnknown">
          <text class="lsc-value">{{unknownStats.total || 0}}</text>
          <text class="lsc-label">生词本</text>
          <text class="lsc-sub">今日 +{{unknownStats.today || 0}}</text>
        </view>
```

- [ ] **Step 6: Verify tsc baseline holds**

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`

Expected: same count as Step 1's baseline. No errors should mention `pages/learn/learn.ts`.

- [ ] **Step 7: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.ts dx-mini/miniprogram/pages/learn/learn.wxml
git commit -m "fix(mini): show today's unknown count instead of always-zero this-week"
```

---

## Task 2: Add back chevron + 学习 title to the learn page

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/learn.ts`
- Modify: `dx-mini/miniprogram/pages/learn/learn.wxml`
- Modify: `dx-mini/miniprogram/pages/learn/learn.wxss`

Add an 88rpx nav-bar row at the top of the page with a left-aligned back chevron and a center-aligned "学习" title. The new row OCCUPIES the 88rpx that previously was empty padding above the progress card.

- [ ] **Step 1: Add the `goBack` handler in `learn.ts`**

In `dx-mini/miniprogram/pages/learn/learn.ts`, locate the existing handlers at the bottom of the `Page({...})` block. The current order is:

```ts
  goGame(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (id) wx.navigateTo({ url: '/pages/games/detail/detail?id=' + id })
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
```

Insert a new handler before `goGame`:

```ts
  goBack() { wx.navigateBack() },
  goGame(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string | undefined
    if (id) wx.navigateTo({ url: '/pages/games/detail/detail?id=' + id })
  },
  goMastered() { wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
  goReview() { wx.navigateTo({ url: '/pages/learn/review/review' }) },
})
```

- [ ] **Step 2: Insert the nav-bar row in `learn.wxml`**

In `dx-mini/miniprogram/pages/learn/learn.wxml`, locate the opening lines:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
```

Insert a new `<view class="nav-bar">...</view>` block IMMEDIATELY AFTER the `<view class="page-container...">` opening tag and BEFORE the `<van-loading...>` element:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="nav-bar">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">学习</text>
    </view>
    <van-loading wx:if="{{loading}}" size="30px" color="#0d9488" class="center-loader" />
```

(The rest of the file is unchanged.)

- [ ] **Step 3: Adjust `.page-container` padding-top + add nav-bar styles in `learn.wxss`**

In `dx-mini/miniprogram/pages/learn/learn.wxss`, locate the `.page-container` rule at the top:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
  padding-bottom: 100rpx;
}
```

Change `padding-top` to drop the `+ 88rpx` (the new `.nav-bar` row now OCCUPIES that 88rpx):

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: var(--status-bar-height, 20px);
  padding-bottom: 100rpx;
}
```

Then, immediately AFTER the `.page-container` rule and BEFORE `.center-loader`, insert the three new nav-bar style rules:

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

So the top of the file now reads:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-top: var(--status-bar-height, 20px);
  padding-bottom: 100rpx;
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
.center-loader { display: flex; justify-content: center; padding: 40px; }
.content { padding: 0 16px; }
... rest unchanged ...
```

- [ ] **Step 4: Verify the icon inventory still covers the WXML**

Run: `cd dx-mini && npm run build:icons`

Expected: clean run; the static WXML scan should pass silently. Both `chevron-left` (already in inventory at `scripts/build-icons.mjs`) and the existing icons (`list-checks`, `gamepad-2`, `check`, `help-circle`, `clock`, `chevron-right`) all resolve.

If `git status` shows `icons.ts` modified after this step, that's a no-op rewrite (timestamp). Verify with `git diff dx-mini/miniprogram/components/dx-icon/icons.ts` — if there's no logical diff, run `git checkout dx-mini/miniprogram/components/dx-icon/icons.ts` to drop the noise.

- [ ] **Step 5: Verify tsc baseline holds**

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`

Expected: same baseline (143 from prior batch). Zero errors should mention `pages/learn/learn.ts`.

- [ ] **Step 6: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/learn.ts dx-mini/miniprogram/pages/learn/learn.wxml dx-mini/miniprogram/pages/learn/learn.wxss
git commit -m "feat(mini): add back chevron + 学习 title to learn page nav bar"
```

---

## Task 3: Rewrite `pages/learn/mastered` for parity

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/mastered/mastered.json`
- Modify: `dx-mini/miniprogram/pages/learn/mastered/mastered.ts`
- Modify: `dx-mini/miniprogram/pages/learn/mastered/mastered.wxml`
- Modify: `dx-mini/miniprogram/pages/learn/mastered/mastered.wxss`

Full rewrite of all four files. Adds: custom-nav top shell, back chevron + "已掌握" title, 3-tile stats header (已掌握总数 / 本周掌握 / 本月掌握), per-row trash icon, relative-date timestamp under each word. Preserves: cursor pagination, pull-down refresh, batch-delete with confirm modal, theme support.

- [ ] **Step 1: Overwrite `mastered.json`**

Replace `dx-mini/miniprogram/pages/learn/mastered/mastered.json` entirely with:

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

(Removed: `navigationBarTitleText` — title is now in WXML; `van-button` — was unused in the previous WXML. Added: `dx-icon` — needed for chevron and trash.)

- [ ] **Step 2: Overwrite `mastered.ts`**

Replace `dx-mini/miniprogram/pages/learn/mastered/mastered.ts` entirely with:

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

- [ ] **Step 3: Overwrite `mastered.wxml`**

Replace `dx-mini/miniprogram/pages/learn/mastered/mastered.wxml` entirely with:

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

- [ ] **Step 4: Overwrite `mastered.wxss`**

Replace `dx-mini/miniprogram/pages/learn/mastered/mastered.wxss` entirely with:

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

- [ ] **Step 5: Verify the icon inventory + tsc**

Run: `cd dx-mini && npm run build:icons`
Expected: clean. Static WXML scan accepts `chevron-left` and `trash-2` (both pre-existing in inventory).

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`
Expected: same baseline. No errors mentioning `pages/learn/mastered/`.

- [ ] **Step 6: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/mastered/
git commit -m "feat(mini): bring mastered page to dx-web parity (custom nav + stats + per-row delete)"
```

---

## Task 4: Rewrite `pages/learn/unknown` for parity

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/unknown/unknown.json`
- Modify: `dx-mini/miniprogram/pages/learn/unknown/unknown.ts`
- Modify: `dx-mini/miniprogram/pages/learn/unknown/unknown.wxml`
- Modify: `dx-mini/miniprogram/pages/learn/unknown/unknown.wxss`

Same shape as Task 3, with three differences: title is `生词本`; stats interface is `{ total, today, lastThreeDays }`; item interface uses only `createdAt` (no `masteredAt`).

- [ ] **Step 1: Overwrite `unknown.json`**

Replace `dx-mini/miniprogram/pages/learn/unknown/unknown.json` entirely with:

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

(`navigationBarTitleText: "生词本"` from the prior PR is removed — the WXML now carries the title.)

- [ ] **Step 2: Overwrite `unknown.ts`**

Replace `dx-mini/miniprogram/pages/learn/unknown/unknown.ts` entirely with:

```ts
import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface UnknownItemData {
  id: string
  contentItem: TrackingContentData | null
  gameName: string | null
  createdAt: string
}
interface UnknownItemView extends UnknownItemData {
  timeText: string
}
interface UnknownStats { total: number; today: number; lastThreeDays: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    loading: true,
    stats: { total: 0, today: 0, lastThreeDays: 0 } as UnknownStats,
    items: [] as UnknownItemView[],
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
      api.get<PaginatedData<UnknownItemData>>('/api/tracking/unknown' + qs),
      reset ? api.get<UnknownStats>('/api/tracking/unknown/stats') : Promise.resolve(this.data.stats),
    ])

    const listOk = results[0].status === 'fulfilled'
    const list = listOk ? results[0].value : { items: [] as UnknownItemData[], nextCursor: '', hasMore: false }
    const newViews: UnknownItemView[] = list.items.map((it) => ({
      ...it,
      timeText: formatRelativeDate(it.createdAt),
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
          await api.delete('/api/tracking/unknown', { ids: [id] })
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
          await api.delete('/api/tracking/unknown', { ids })
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

- [ ] **Step 3: Overwrite `unknown.wxml`**

Replace `dx-mini/miniprogram/pages/learn/unknown/unknown.wxml` entirely with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="nav-bar">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">生词本</text>
    </view>

    <view class="content">
      <view class="stat-card-row">
        <view class="learn-stat-card teal">
          <text class="lsc-value">{{stats.total}}</text>
          <text class="lsc-label">全部生词</text>
        </view>
        <view class="learn-stat-card amber">
          <text class="lsc-value">{{stats.today}}</text>
          <text class="lsc-label">今日添加</text>
        </view>
        <view class="learn-stat-card purple">
          <text class="lsc-value">{{stats.lastThreeDays}}</text>
          <text class="lsc-label">最近三天</text>
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
      <van-empty wx:if="{{!loading && items.length === 0}}" description="生词本里还空着" />

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

- [ ] **Step 4: Overwrite `unknown.wxss`**

Replace `dx-mini/miniprogram/pages/learn/unknown/unknown.wxss` entirely with the SAME content as Task 3 Step 4 (`mastered.wxss`):

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

- [ ] **Step 5: Verify the icon inventory + tsc**

Run: `cd dx-mini && npm run build:icons`
Expected: clean.

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`
Expected: same baseline. No errors mentioning `pages/learn/unknown/`.

- [ ] **Step 6: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/unknown/
git commit -m "feat(mini): bring 生词本 page to dx-web parity (custom nav + stats + per-row delete)"
```

---

## Task 5: Rewrite `pages/learn/review` for parity

**Files:**
- Modify: `dx-mini/miniprogram/pages/learn/review/review.json`
- Modify: `dx-mini/miniprogram/pages/learn/review/review.ts`
- Modify: `dx-mini/miniprogram/pages/learn/review/review.wxml`
- Modify: `dx-mini/miniprogram/pages/learn/review/review.wxss`

Same shape as Task 3, with three differences: title is `待复习`; stats interface is `{ pending, overdue, reviewedToday }`; item interface uses only `createdAt`.

- [ ] **Step 1: Overwrite `review.json`**

Replace `dx-mini/miniprogram/pages/learn/review/review.json` entirely with:

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

- [ ] **Step 2: Overwrite `review.ts`**

Replace `dx-mini/miniprogram/pages/learn/review/review.ts` entirely with:

```ts
import { api, PaginatedData } from '../../../utils/api'
import { formatRelativeDate } from '../../../utils/format'

interface TrackingContentData { content: string; translation: string | null; contentType: string }
interface ReviewItemData {
  id: string
  contentItem: TrackingContentData | null
  gameName: string | null
  createdAt: string
}
interface ReviewItemView extends ReviewItemData {
  timeText: string
}
interface ReviewStats { pending: number; overdue: number; reviewedToday: number }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    statusBarHeight: 20,
    loading: true,
    stats: { pending: 0, overdue: 0, reviewedToday: 0 } as ReviewStats,
    items: [] as ReviewItemView[],
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
      api.get<PaginatedData<ReviewItemData>>('/api/tracking/review' + qs),
      reset ? api.get<ReviewStats>('/api/tracking/review/stats') : Promise.resolve(this.data.stats),
    ])

    const listOk = results[0].status === 'fulfilled'
    const list = listOk ? results[0].value : { items: [] as ReviewItemData[], nextCursor: '', hasMore: false }
    const newViews: ReviewItemView[] = list.items.map((it) => ({
      ...it,
      timeText: formatRelativeDate(it.createdAt),
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
          await api.delete('/api/tracking/review', { ids: [id] })
          this.setData({
            items: this.data.items.filter((i) => i.id !== id),
            stats: { ...this.data.stats, pending: Math.max(0, this.data.stats.pending - 1) },
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
          await api.delete('/api/tracking/review', { ids })
          this.setData({
            items: this.data.items.filter((i) => !ids.includes(i.id)),
            selectedIds: [],
            selectMode: false,
            stats: { ...this.data.stats, pending: Math.max(0, this.data.stats.pending - ids.length) },
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

(Note: review's `total`-equivalent is `pending`; `Math.max(0, pending - n)` keeps the figure non-negative after a delete.)

- [ ] **Step 3: Overwrite `review.wxml`**

Replace `dx-mini/miniprogram/pages/learn/review/review.wxml` entirely with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <view class="nav-bar">
      <view class="back-btn" bind:tap="goBack">
        <dx-icon name="chevron-left" size="22px" color="{{theme === 'dark' ? '#f5f5f5' : '#1a1a1a'}}" />
      </view>
      <text class="nav-title">待复习</text>
    </view>

    <view class="content">
      <view class="stat-card-row">
        <view class="learn-stat-card teal">
          <text class="lsc-value">{{stats.pending}}</text>
          <text class="lsc-label">待复习</text>
        </view>
        <view class="learn-stat-card amber">
          <text class="lsc-value">{{stats.overdue}}</text>
          <text class="lsc-label">逾期</text>
        </view>
        <view class="learn-stat-card purple">
          <text class="lsc-value">{{stats.reviewedToday}}</text>
          <text class="lsc-label">今日复习</text>
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
      <van-empty wx:if="{{!loading && items.length === 0}}" description="暂无待复习的词" />

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

- [ ] **Step 4: Overwrite `review.wxss`**

Replace `dx-mini/miniprogram/pages/learn/review/review.wxss` entirely with the SAME content as Task 3 Step 4 (`mastered.wxss`):

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

- [ ] **Step 5: Verify the icon inventory + tsc**

Run: `cd dx-mini && npm run build:icons`
Expected: clean.

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`
Expected: same baseline. No errors mentioning `pages/learn/review/`.

- [ ] **Step 6: Commit**

```bash
git add dx-mini/miniprogram/pages/learn/review/
git commit -m "feat(mini): bring 待复习 page to dx-web parity (custom nav + stats + per-row delete)"
```

---

## Task 6: Delete `pages/me/study` and unregister from `app.json`

**Files:**
- Delete: `dx-mini/miniprogram/pages/me/study/study.json`
- Delete: `dx-mini/miniprogram/pages/me/study/study.ts`
- Delete: `dx-mini/miniprogram/pages/me/study/study.wxml`
- Delete: `dx-mini/miniprogram/pages/me/study/study.wxss`
- Delete: `dx-mini/miniprogram/pages/me/study/` (the now-empty directory)
- Modify: `dx-mini/miniprogram/app.json`

The page is a "敬请期待" placeholder reachable from nowhere. Verified earlier — no other page navigates to `/pages/me/study/study` (only `app.json:23`'s registration line matches).

- [ ] **Step 1: Confirm nothing else references the page**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source && grep -rn "pages/me/study" dx-mini/miniprogram/`

Expected: exactly ONE line — `dx-mini/miniprogram/app.json:23: "pages/me/study/study",`

If you see any other matches, STOP and report DONE_WITH_CONCERNS — there's an unexpected caller and we shouldn't delete the page until that caller is investigated.

- [ ] **Step 2: Delete the four files and the directory**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
rm dx-mini/miniprogram/pages/me/study/study.json
rm dx-mini/miniprogram/pages/me/study/study.ts
rm dx-mini/miniprogram/pages/me/study/study.wxml
rm dx-mini/miniprogram/pages/me/study/study.wxss
rmdir dx-mini/miniprogram/pages/me/study
```

`rmdir` (not `rm -rf`) ensures we only delete an empty directory — fails loudly if any file unexpectedly remained.

- [ ] **Step 3: Remove the registration line from `app.json`**

In `dx-mini/miniprogram/app.json`, locate the `pages[]` array. The current state has the line:

```json
    "pages/me/study/study",
```

Delete that single line. Be careful with JSON commas — the line above and the line below should still be valid JSON. The current surrounding context is:

```json
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
  ],
```

After deletion:

```json
    "pages/me/profile-edit/profile-edit",
    "pages/me/groups/groups",
    "pages/me/groups-detail/groups-detail",
    "pages/me/invite/invite",
    "pages/me/redeem/redeem",
    "pages/me/purchase/purchase",
    "pages/me/tasks/tasks",
    "pages/community/community",
    "pages/community/detail/detail",
    "pages/me/feedback/feedback"
  ],
```

The neighboring `"pages/me/purchase/purchase",` keeps its trailing comma (it now precedes the next entry, `pages/me/tasks/tasks`).

- [ ] **Step 4: Validate JSON + verify references gone**

Run: `python3 -c "import json; json.load(open('dx-mini/miniprogram/app.json'))"`
Expected: no output (valid JSON).

Run: `grep -rn "pages/me/study" dx-mini/miniprogram/`
Expected: NO matches (the registration is gone, the directory is gone).

- [ ] **Step 5: Verify the icon scan + tsc still clean**

Run: `cd dx-mini && npm run build:icons`
Expected: clean.

Run: `cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -c "error TS"`
Expected: same baseline.

- [ ] **Step 6: Commit**

```bash
git add -A dx-mini/miniprogram/pages/me/study dx-mini/miniprogram/app.json
git commit -m "chore(mini): delete unused me/study placeholder page"
```

(Note: `git add -A <path>` correctly stages deletions for that path.)

---

## Task 7: Manual smoke test in WeChat DevTools

**Files:** none (verification only — no commit unless a regression is found and fixed)

Per the project memory `project_wechat_devtools_realdevice_bug.md`, use **预览 (Preview)** with the WeChat 小程序助手 phone app, NOT 真机调试.

- [ ] **Step 1: Confirm dx-api is running**

```bash
cd dx-api && go run .
```

(Or use `air`.) Confirm port 3001 is reachable on the LAN IP that DevTools `setDevApiBaseUrl(...)` points at.

- [ ] **Step 2: Open the project in WeChat DevTools, set dev URL if needed**

```js
require('./utils/config').setDevApiBaseUrl('http://<lan-ip>')
```

- [ ] **Step 3: Verify the parent learn page**

- [ ] Tap home → tap the "学习" circle → land on the learn page.
- [ ] Top of page now shows: ◁ back chevron on the LEFT, 学习 title CENTERED, WeChat capsule on the RIGHT — all in the same horizontal band.
- [ ] No overlap between the back-chevron and the capsule, no overlap between the title and the capsule.
- [ ] Tap back chevron → returns to home tab.
- [ ] On the middle stat card under the progress card, the sub-text now reads `今日 +N` (where N is your today-added count), NOT `本周 +0`.

- [ ] **Step 4: Verify the three quick-link pages**

For EACH of mastered / unknown / review:

- [ ] Tap the corresponding quick-link → land on the destination page.
- [ ] Top row: back chevron LEFT, correct title CENTERED (已掌握 / 生词本 / 待复习), capsule RIGHT.
- [ ] Tap back chevron → returns to learn page.
- [ ] 3-tile stats header renders just below the nav row, with the page-specific labels:
  - mastered: 已掌握总数 / 本周掌握 / 本月掌握
  - unknown: 全部生词 / 今日添加 / 最近三天
  - review: 待复习 / 逾期 / 今日复习
- [ ] Action bar shows `{{N}} 词` and `批量删除` button.
- [ ] Word list renders with content + translation + 来自：<game> + relative-date timestamp under each.
- [ ] Per-row trash icon appears on the right of each row.
- [ ] Tap trash icon → confirm modal "删除该词？" → 确定 → row disappears, total stat decrements by 1.
- [ ] Tap 批量删除 → trash icons hide, checkboxes appear → check 2-3 → 删除(N) button appears → tap → confirm → those rows disappear, total decrements by N.
- [ ] Pull down to refresh → list + stats both reload.
- [ ] Scroll to bottom (if you have 20+ items) → next page loads.

- [ ] **Step 5: Verify dark theme on all four pages**

- [ ] Toggle dark mode (from 我的 page's 深色模式 row, or the toggle the project uses).
- [ ] Re-visit each of the four pages (learn, mastered, unknown, review).
- [ ] Verify: back chevron readable, title readable, stat tile colors readable, list rows readable, trash icon visible.
- [ ] No white flashes, no clipped text, no invisible icons.

- [ ] **Step 6: Verify `me/study` is gone**

- [ ] In DevTools' page list / file tree, confirm `pages/me/study/` is not present.
- [ ] Try manually navigating: `wx.navigateTo({ url: '/pages/me/study/study' })` from the DevTools console. Expected: error (page not registered).

- [ ] **Step 7: Sanity-check no regressions on other tabs**

- [ ] Tap each of 首页 / 课程 / 排行 / 消息 / 我的 → each tab loads as before.

- [ ] **Step 8: Final cleanup**

If any defect surfaced, fix it inline (returning to the relevant earlier task) and commit as `fix(mini): ...`. If everything passed, no commit needed.

---

## Self-review notes (already applied)

- **Spec coverage:** Every spec section maps to a task — back button + nav title (Task 2), three quick-link page parity (Tasks 3-5), me/study removal (Task 6), the bonus unknownStats label fix (Task 1), manual smoke (Task 7).
- **Placeholder scan:** No "TBD", "TODO", "fill in", or "similar to Task N". The wxss appears verbatim in Tasks 3, 4, 5 (intentional repetition per the writing-plans skill — no shared CSS file, three is too few to extract).
- **Type consistency:** Every interface (`MasterStats` / `UnknownStats` / `ReviewStats` / `MasterItemData` / `UnknownItemData` / `ReviewItemData` / `MasterItemView` / `UnknownItemView` / `ReviewItemView`) is declared in the file that uses it. Field names cross-reference correctly: WXML reads `stats.total` / `stats.thisWeek` / `stats.thisMonth` for mastered, `stats.total` / `stats.today` / `stats.lastThreeDays` for unknown, `stats.pending` / `stats.overdue` / `stats.reviewedToday` for review — all match each `data.stats` initial value AND the corresponding interface AND the dx-api response shapes verified during exploration.
- **Convention checks:** No optional chaining (`?.`) or nullish coalescing (`??`) anywhere; no method calls in WXML (display strings pre-computed via `timeText` in `loadAll`); icon names already in inventory (`chevron-left`, `trash-2`); manual smoke uses 预览 not 真机调试.
