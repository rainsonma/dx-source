# Hall Notification Banner Design

**Date:** 2026-04-09
**Status:** Approved

## Summary

Add a rotating notification banner above the main 3-card row on `/hall` that cycles through the latest 3 notices from `GET /api/notices`. Clicking a notice opens a ShadCN Dialog showing the full notice with a bottom-right timestamp. Also move the existing `StatsRow` down to sit directly above `LearningHeatmap`.

No backend changes. No database changes. No new hooks, actions, or feature folders — reuses existing helpers from the `notice` feature.

## Current State

`dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` renders this vertical order:

```
GreetingTopBar
AdCardsRow
StatsRow
[ GameProgressCard | TodayStarsCard | DailyChallengeCard ]   ← main content row
LearningHeatmap
```

The notices system is already implemented:
- Go API: `GET /api/notices?limit=3` returns cursor-paginated `{ items, nextCursor, hasMore }` (limit supported via `helpers.ParseCursorParams`)
- Frontend: `/hall/notices` page + `features/web/notice/` feature folder with `NoticeItem` type, `formatRelativeTime` helper, `resolveNoticeIcon` helper, and `NoticeItem` presentational component

## New Layout

```
GreetingTopBar
AdCardsRow
NotificationBanner                                           ← NEW
[ GameProgressCard | TodayStarsCard | DailyChallengeCard ]   ← unchanged
StatsRow                                                     ← MOVED
LearningHeatmap
```

Two visual changes:
1. Insert `<NotificationBanner />` directly after `<AdCardsRow />`
2. Move `<StatsRow .../>` from its current position (just below `AdCardsRow`) to sit directly above `LearningHeatmap`

No prop changes to `StatsRow`, no style changes to any existing card, no changes to the 3-card row.

## Visual Design

**Style:** Compact ticker (~56px tall). Single-row horizontal bar matching the project's `rounded-[14px] border border-border bg-card` card treatment.

```
[ ⎕ icon ]  [ 标题 ]  [ 一行截断正文 ]  [ • • • ]  [ 2 小时前 ]
   36px    14px bold   13px muted           dots      12px muted
```

Target overall banner height: ~56px (icon `h-9 w-9` = 36px + button `py-2.5` = 20px vertical padding).

**Rotation:** Auto-advance every 5 seconds via vertical slide-up animation (`animate-in fade-in-0 slide-in-from-bottom-1 duration-300`). Keyed remount on `currentIndex` change triggers the animation each tick.

**Interaction:**
- Hover on banner → pause rotation
- Click anywhere on banner → open dialog (and pause rotation while open)
- Dialog close → resume rotation from the currently-shown index

**Dialog:** ShadCN `Dialog` with constant title `斗学消息通知`, body showing notice icon + notice title + notice content, footer with bottom-right relative timestamp. Body scrolls via `max-h-[60vh] overflow-y-auto` for long content.

## Component Architecture

Two new files in `features/web/hall/components/`:

### `notification-banner.tsx` — container

Client component. Owns: data fetch, rotation state, pause state, dialog open state, and ticker rendering.

**State:**
```ts
const [notices, setNotices] = useState<NoticeItem[]>([]);
const [loaded, setLoaded] = useState(false);
const [currentIndex, setCurrentIndex] = useState(0);
const [hovered, setHovered] = useState(false);
const [dialogNotice, setDialogNotice] = useState<NoticeItem | null>(null);

const paused = hovered || dialogNotice !== null;
```

**Fetch effect** (mount-once):
```ts
useEffect(() => {
  let cancelled = false;
  apiClient
    .get<NoticeListResponse>("/api/notices?limit=3")
    .then((res) => {
      if (cancelled) return;
      if (res.code === 0) setNotices(res.data.items ?? []);
      setLoaded(true);
    })
    .catch(() => {
      if (!cancelled) setLoaded(true);
    });
  return () => { cancelled = true; };
}, []);
```

**Rotation effect:**
```ts
useEffect(() => {
  if (notices.length < 2 || paused) return;
  const id = setInterval(() => {
    setCurrentIndex((i) => (i + 1) % notices.length);
  }, ROTATION_INTERVAL_MS); // 5000
  return () => clearInterval(id);
}, [notices.length, paused]);
```

**Render decision tree:**
```
!loaded                       → return null
notices.length === 0          → return null
notices.length === 1          → ticker, no dots, no rotation
notices.length === 2 or 3     → ticker with dots, rotation active
```

**Ticker markup (inside `<button type="button">`):**
- Icon box: `h-9 w-9 rounded-[10px] bg-teal-50` + `resolveNoticeIcon(current.icon)` at `h-[18px] w-[18px] text-teal-600`
- Animated text wrapper: `<div key={currentIndex} className="animate-in fade-in-0 slide-in-from-bottom-1 duration-300 flex min-w-0 flex-1 items-center gap-2.5 sm:gap-3">` — remounts on index change
- Title: `text-sm font-semibold text-foreground shrink-0`
- Content snippet: `hidden truncate text-sm text-muted-foreground sm:inline` (mobile hides snippet to avoid cramping)
- Dots (only if `notices.length > 1`): `hidden shrink-0 items-center gap-1 lg:flex`; active pill = `h-1.5 w-3.5 rounded-full bg-teal-500`, inactive = `h-1.5 w-1.5 rounded-full bg-slate-300`, both `transition-all`
- Timestamp: `shrink-0 text-xs text-muted-foreground` + `formatRelativeTime(current.createdAt)`

**Button classes:**
```
group flex w-full items-center gap-3 rounded-[14px] border border-border bg-card px-4 py-2.5
text-left transition-colors hover:border-teal-200 hover:bg-teal-50/30
focus-visible:ring-2 focus-visible:ring-teal-500/50 focus-visible:ring-offset-2 focus-visible:outline-hidden
sm:px-5
```

With `py-2.5` (10px top + 10px bottom) and a `h-9 w-9` icon box (36px), the computed button height is exactly 56px, matching the mockup approved in brainstorming.

**Accessibility:** `aria-label={\`查看消息通知：${current.title}\`}` on the button; it is a real `<button type="button">` so Enter/Space/Tab/focus ring all work without manual wiring.

### `notification-banner-dialog.tsx` — presentational dialog

**Props:**
```ts
interface NotificationBannerDialogProps {
  notice: NoticeItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}
```

`notice: null` on close is tolerated during the dialog's exit animation (keeps the last notice visible while fading out to prevent flicker).

**Structure:** `DialogContent` with `gap-0 overflow-hidden p-0 sm:max-w-lg` so we can lay out three custom sections:

1. **Header** (`border-b border-border bg-slate-50/60 px-6 py-4`) — `DialogTitle` = `斗学消息通知`, `DialogDescription` = `sr-only` "查看消息通知内容" (silences radix a11y warning without visible text)
2. **Body** (`flex max-h-[60vh] flex-col gap-3 overflow-y-auto px-6 py-5`) — notice icon + notice title, then `<p className="text-sm leading-relaxed whitespace-pre-wrap text-muted-foreground">` for notice content (preserves newlines)
3. **Footer** (`flex items-center justify-end border-t border-border px-6 py-3`) — `<span className="text-xs text-muted-foreground">{formatRelativeTime(notice.createdAt)}</span>` pinned bottom-right per spec

Default `showCloseButton` (shadcn `X` in top-right) is kept.

## Responsive Behavior

| Viewport | Banner elements visible |
|---|---|
| `< sm` (mobile) | icon, title, timestamp |
| `sm` to `< lg` | icon, title, truncated content snippet, timestamp |
| `lg+` (desktop) | icon, title, truncated content snippet, dot indicators, timestamp |

Dialog is always centered with `sm:max-w-lg`.

## Data Strategy

**Approach:** Reuse existing `GET /api/notices` endpoint. No backend changes.

- Endpoint: `GET /api/notices?limit=3`
- Auth: protected route (user JWT required; `/hall` is already inside the protected layout)
- Pagination: cursor-based, but only the first page is requested with `limit=3`
- Frequency: one fetch per mount of `NotificationBanner` (no polling, no SSE, no SSR)
- Error handling: silent — on non-zero `code` or thrown error, banner hides
- Response type: inlined as `NoticeListResponse` matching the same shape used at `dx-web/src/app/(web)/hall/(main)/notices/page.tsx:26`

**Why client-side fetch:** Matches the existing `/hall/(home)/page.tsx` pattern (which client-fetches `/api/hall/dashboard` and `/api/hall/heatmap`). Keeps the hall page's data concerns in one place (the page component and its child components).

**Why no mark-as-read:** The banner is a "latest 3" preview, not an inbox. Mark-as-read behavior remains on `/hall/notices` via the existing `MarkNoticesRead` component. The banner must not touch `users.last_read_notice_at`.

## Edge Cases

| Condition | Behavior |
|---|---|
| First fetch in flight (`!loaded`) | Return `null` — no skeleton, no layout shift |
| Fetch throws | Catch → `loaded = true` with empty array → banner hidden |
| API returns `code !== 0` | `loaded = true` with empty array → banner hidden |
| `notices.length === 0` | Return `null` — silent |
| `notices.length === 1` | Ticker renders, rotation effect early-returns, no dots rendered |
| `notices.length === 2 or 3` | Ticker renders with dots, rotation active every 5s |
| User hovers ticker | `hovered = true` → `paused = true` → interval torn down |
| User clicks ticker | `dialogNotice` set → `paused = true` → dialog opens |
| User closes dialog | `dialogNotice = null` → `paused` recomputed (still hovered?) → rotation may resume |
| Banner unmounts during fetch | `cancelled` flag prevents stale `setState` |
| Banner unmounts during rotation | Effect cleanup clears interval |
| React strict-mode double-invoke | `cancelled` flag handles the teardown cleanly |
| Content has newlines | `whitespace-pre-wrap` in dialog body preserves them |
| Very long content in dialog | `max-h-[60vh] overflow-y-auto` on body → internal scroll |

## File Changes

### New Files

| File | Purpose |
|---|---|
| `dx-web/src/features/web/hall/components/notification-banner.tsx` | Container: fetch, state, rotation, ticker render, dialog wiring |
| `dx-web/src/features/web/hall/components/notification-banner-dialog.tsx` | Presentational ShadCN dialog for a single notice |

### Modified Files

| File | Change |
|---|---|
| `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` | Add `NotificationBanner` import; insert `<NotificationBanner />` after `<AdCardsRow />`; move `<StatsRow .../>` to sit directly above `LearningHeatmap` |

### Reused (import, not copy)

- `NoticeItem` type from `@/features/web/notice/actions/notice.action`
- `formatRelativeTime` from `@/features/web/notice/helpers/notice-time`
- `resolveNoticeIcon` from `@/features/web/notice/helpers/notice-icon`
- `apiClient.get` from `@/lib/api-client`
- `cn` from `@/lib/utils`
- `Dialog`, `DialogContent`, `DialogDescription`, `DialogHeader`, `DialogTitle` from `@/components/ui/dialog`

### No Backend Changes

The dx-api Go backend already exposes `GET /api/notices?limit=3` via `helpers.ParseCursorParams(ctx, 20)` at `dx-api/app/helpers/pagination.go:16-31`. The notices table schema is untouched. The code-level FK constraint pattern is untouched (notices table has no FKs). No migrations, no new routes, no new models, no new controllers, no new services.

### No Deploy Changes

`deploy/docker-compose.dev.yml` / `docker-compose.prod.yml` untouched. The feature is a pure client-side addition to an existing dynamic Next.js route; no build-time env vars, no nginx config changes.

## Styling

- Teal accent: `bg-teal-50` icon background, `text-teal-600` icon color, `bg-teal-500` active dot, `hover:border-teal-200 hover:bg-teal-50/30` on button — matches existing hall cards (`today-stars-card.tsx`, `notice-item.tsx`)
- Card treatment: `rounded-[14px] border border-border bg-card` — identical to `StatsRow`, `GameProgressCard`, `TodayStarsCard`, `DailyChallengeCard`
- Muted text: `text-muted-foreground` for snippet and timestamp
- Dialog tint: `bg-slate-50/60` on the header for a subtle "bulletin board" feel without being loud
- Dialog body uses `whitespace-pre-wrap` to preserve newlines in notice content
- No emoji icons in code — the real `Lucide` icon from `resolveNoticeIcon` is used

## Non-Goals (Explicit)

To prevent scope creep during implementation:

- No mark-as-read integration (keeps `/hall/notices` as the sole inbox)
- No real-time push / SSE / polling
- No unread-count badge on the banner
- No "View all" link in the banner (greeting bar already links to `/hall/notices`)
- No admin-editor changes
- No changes to the existing `features/web/notice/` feature folder
- No internationalization — hardcoded Chinese strings per existing convention
- No unit tests — dx-web has no React testing infrastructure; adding one for a single component is out of scope
- No change to default rotation interval — fixed at 5s, not user-configurable

## Constraints

- **No lint issues** — `npm run lint` must pass with 0 warnings
- **No type errors** — `npm run build` must succeed with strict TypeScript
- **No breaking changes** — all existing `/hall` functionality, routes, and API calls continue to work
- **No backend changes** — dx-api untouched
- **No database changes** — no migrations, no schema edits
- **Preserve the code-level FK constraint pattern** — not impacted by this change, but called out for awareness
- **Responsive** — must work on mobile (`< 640px`) and desktop (`≥ 1024px`)
- **Accessible** — real `<button>` element, `aria-label`, `DialogDescription` for screen readers, focus-visible ring
- **Matches existing conventions** — kebab-case filenames, PascalCase exports, `"use client"` top directive, `@/` import alias, no barrel imports

## Verification

### Static checks (must pass before PR)

```bash
cd dx-web
npm run lint            # ESLint — 0 warnings
npm run build           # Next.js build — 0 type errors
```

### Manual test plan

1. With 3+ active notices → banner appears between `AdCardsRow` and the 3-card row, rotates every 5s, dots animate
2. Hover pauses rotation; mouse-leave resumes
3. Click opens dialog with header `斗学消息通知`, body shows notice icon + title + content, footer shows relative time pinned bottom-right
4. Close dialog via X button, overlay click, or ESC → rotation resumes from the currently displayed index
5. Soft-delete all notices in admin (set `is_active=false`) → banner disappears on next page load, rest of hall page renders normally
6. Create exactly 1 active notice → banner shows it without dots, no rotation
7. Create exactly 2 active notices → banner rotates between them with 2 dots visible on `lg+`
8. Mobile viewport (< 640px) → content snippet hidden, dots hidden, layout stays clean with icon + title + timestamp
9. Tablet viewport (640–1023px) → content snippet visible, dots hidden
10. Desktop viewport (≥ 1024px) → all elements visible
11. API returns error (simulate via DevTools network throttle / block) → banner absent, rest of hall page renders normally
12. Verify `StatsRow` now sits directly above `LearningHeatmap` with correct spacing
13. Verify no console warnings from radix-ui about missing `DialogDescription`
14. Verify Tab key navigation: can tab to the banner, banner shows focus-visible ring, Enter/Space opens dialog
15. Verify dialog scroll: create a notice with very long content (> 500 chars) → dialog body scrolls internally, footer stays pinned
