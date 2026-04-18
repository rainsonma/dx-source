# AI Custom Confirm-Dialog Error Visibility

## Problem

On the ai-custom game detail page (`/hall/ai-custom/[id]`), clicking **撤回** on a published game can return `HTTP 400` from `POST /api/course-games/{id}/withdraw` with a Chinese message such as `"还有 N 个进行中的游戏会话，请等待结束后再撤回"`. The backend intentionally blocks withdraw when active (`ended_at IS NULL`) single-play sessions exist on the game (`dx-api/app/services/api/course_game_service.go:408-437`). That's correct behavior and stays unchanged.

The frontend, however, **silently swallows that error**. The user sees nothing — no toast, no inline text, no state change. Only the browser console records the 400.

Root cause: the withdraw confirmation button in `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx:324` is a shadcn `AlertDialogAction`, which is Radix's `AlertDialog.Action` — a `DialogClose` under the hood. Clicking it synchronously closes the dialog before the async request resolves. The error state written by `useWithdrawGame` (`dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts:71-87`) then lands in an unmounted `AlertDialogContent`, so `{withdrawAction.error && <p>...</p>}` never renders.

The identical bug exists in `useDeleteGame` and the 删除 `AlertDialogAction` (same file, lines 15-34 / 267-276). It is hidden today only because a successful delete navigates the user away — on error, a user would see the same silent failure.

## Scope

- **In scope:** frontend-only fix on the two ai-custom confirmation dialogs (撤回 and 删除) so that server-returned errors display inline inside the existing shadcn `AlertDialog`, and the dialog closes only on success.
- **Out of scope:** any backend change (the 400 semantics are correct), any scheduler for stale sessions, the publish dialog (already works via `toast.error`), and the delete/withdraw dialogs anywhere outside the ai-custom feature.

## Design

### Hook change pattern (both hooks)

`useWithdrawGame` and `useDeleteGame` in `use-game-actions.ts` go from returning `{ execute, isPending, error }` to `{ execute, isPending, error, clearError }`, with three adjustments:

1. `execute` accepts an optional `onSuccess` callback. The caller passes `() => setDialogOpen(false)` so the dialog closes **only** after a successful response.
2. `execute` clears the local error at the start of each invocation (`setError(null)`) so retrying after a dismissed error doesn't briefly flash the stale message.
3. A stable `clearError` is exposed so the dialog can reset the error when it reopens (see **Error-reset on reopen** below).

Shape after change (withdraw shown; delete mirrors it):

```ts
export function useWithdrawGame(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute(onSuccess?: () => void) {
    startTransition(async () => {
      setError(null)
      const result = await withdrawGameAction(gameId)
      if (result.error) {
        setError(result.error)
        return
      }
      await swrMutate("/api/course-games")
      onSuccess?.()
    })
  }

  const clearError = useCallback(() => setError(null), [])

  return { execute, isPending, error, clearError }
}
```

`useDeleteGame` today also navigates away on success (`router.replace("/hall/ai-custom")`). That stays; the navigation replaces the page, so passing a no-op `onSuccess` is fine (or the navigation itself is the "success" action). The relevant change is that on **error**, the dialog now stays open with the message instead of closing silently.

### Dialog wiring (both dialogs)

In `game-hero-card.tsx`, each confirmation button changes from:

```tsx
<AlertDialogAction onClick={action.execute} ...>
```

to:

```tsx
<AlertDialogAction
  onClick={(e) => {
    e.preventDefault()
    action.execute(() => setDialogOpen(false))
  }}
  ...
>
```

`e.preventDefault()` cancels Radix's default `DialogClose` behavior (documented for `AlertDialog.Action` / `AlertDialog.Cancel`). The dialog now stays open during the request, shows the error if the backend rejects, and closes only when `onSuccess` fires.

### Error-reset on reopen (both dialogs)

When the user dismisses a dialog with an error showing and reopens it, the previous message should not flash. Wrap `onOpenChange` to clear the error:

```tsx
<AlertDialog
  open={withdrawOpen}
  onOpenChange={(open) => {
    if (!open) withdrawAction.clearError()
    setWithdrawOpen(open)
  }}
>
```

This requires exposing `clearError: () => setError(null)` from each hook. Same treatment for the delete dialog.

### What stays the same

- The inline `{action.error && <p className="text-sm text-red-500">...</p>}` paragraphs inside `AlertDialogContent` — already in place at lines 260-262 (delete) and 317-319 (withdraw). No markup to add.
- `usePublishGame` and its 发布 dialog — unchanged. `toast.error` already works for publish; touching it would be churn.
- The cancel buttons (`AlertDialogCancel`) — unchanged. Radix's default close-on-cancel is the desired behavior.
- Backend routes, services, models, migrations — unchanged.

## Data Flow

```
User clicks 确定撤回
        │
        ▼
AlertDialogAction onClick
        │ e.preventDefault()   ← dialog stays open
        │ withdrawAction.execute(onClose)
        ▼
POST /api/course-games/{id}/withdraw
        │
   ┌────┴────┐
   │         │
   ▼         ▼
 2xx       4xx/5xx
   │         │
   │         └─► setError(message)   → dialog shows inline red text; user can cancel or retry
   │
   └─► swrMutate() → onClose() → setWithdrawOpen(false)   → dialog closes on success
```

## Error Handling

- Server errors (400 "还有 N 个进行中的游戏会话…", 403 "无权操作", 409 "游戏名称已存在", 500 "操作失败") propagate through `withdrawGameAction` / `deleteGameAction` → `result.error` → inline `<p>` in the dialog. Chinese messages render as-is.
- Network errors (thrown from `apiFetch`) are already caught in the action functions and surfaced as `{ error: "撤回失败" }` / `{ error: "删除游戏失败" }`. Same path — render inline.
- No new failure modes introduced.

## Testing

- `npm run lint` clean.
- `npm run build` clean (TypeScript check).

Manual verification (dev server, browser):

1. Start dx-api with a game that has ≥ 1 active session (`ended_at IS NULL`) you own.
2. Open `/hall/ai-custom/{id}` as the owner → click 撤回 → 确定撤回.
   - Expected: dialog stays open, red error text shows the backend message (e.g. "还有 1 个进行中的游戏会话，请等待结束后再撤回"), game status stays `published`, 确定撤回 is re-enabled.
3. Cancel the dialog, end/force-complete the active session, reopen the dialog.
   - Expected: no stale error on reopen; 确定撤回 succeeds; dialog closes; 撤回 button disappears, 发布 appears.
4. In the delete flow on an unpublished game, induce a 500 (e.g. by stopping the DB and retrying) and confirm the error displays inline and the dialog stays open; retry after DB is back and confirm success closes the dialog and navigates to `/hall/ai-custom`.
5. Regression: publish flow still uses toast and works unchanged.

## Files Touched

- `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts` — `useWithdrawGame` and `useDeleteGame`: add `onSuccess` param to `execute`, add `clearError`, reset error at start of `execute`.
- `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx` — two `AlertDialogAction` buttons (撤回, 删除) get `e.preventDefault()` + `onSuccess` wiring; two `AlertDialog` wrappers get `onOpenChange` that calls `clearError` on close.

No other files change.
