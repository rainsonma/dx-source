# AI Custom Confirm-Dialog Error Visibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the ai-custom game detail page so that server-side errors from 撤回 (withdraw) and 删除 (delete) actions display inline inside the shadcn `AlertDialog` instead of being silently swallowed when the dialog auto-closes.

**Architecture:** Frontend-only change in two files. Each hook gains an optional `onSuccess` callback plus a stable `clearError` function, and each confirmation `AlertDialogAction` uses `e.preventDefault()` + `onSuccess` wiring so the dialog closes only on a successful response. On an error response, the dialog stays open and renders the backend's message via the existing `{action.error && <p>…</p>}` paragraph.

**Tech Stack:** Next.js 16 (App Router, client components), React 19, shadcn/ui `AlertDialog` (wrapping Radix `AlertDialog`), SWR, TypeScript. No test framework configured for dx-web — verification is `npm run lint`, `npm run build`, and manual browser checks against a running `dx-api`.

**Spec:** `docs/superpowers/specs/2026-04-18-ai-custom-confirm-dialog-error-visibility-design.md`

---

## File Structure

Two files change. Both already exist — no new files.

Frontend (dx-web):
- `src/features/web/ai-custom/hooks/use-game-actions.ts` — `useWithdrawGame` and `useDeleteGame` each gain `onSuccess` parameter on `execute`, reset error at start of `execute`, and expose `clearError`.
- `src/features/web/ai-custom/components/game-hero-card.tsx` — the 撤回 and 删除 `AlertDialog` wrappers get `onOpenChange` handlers that call `clearError` on close, and their `AlertDialogAction` buttons use `e.preventDefault()` + `onSuccess` to close only on success.

No backend files change. No new routes, services, models, migrations, or tests.

---

## Task 1: Fix the 撤回 (withdraw) flow

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts` (`useWithdrawGame`, lines 71-87)
- Modify: `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx` (withdraw dialog, lines 98 / 102 / 309-336)

- [ ] **Step 1: Add `useCallback` to the React imports in `use-game-actions.ts`**

Open `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts`. The current first import line is:

```ts
import { useState, useTransition } from "react"
```

Replace it with:

```ts
import { useCallback, useState, useTransition } from "react"
```

- [ ] **Step 2: Rewrite `useWithdrawGame` to support `onSuccess` + `clearError`**

Still in `use-game-actions.ts`, replace the current `useWithdrawGame` block (lines 71-87):

```ts
export function useWithdrawGame(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute() {
    startTransition(async () => {
      const result = await withdrawGameAction(gameId)
      if (result.error) {
        setError(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending, error }
}
```

with:

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

- [ ] **Step 3: Wire `onOpenChange` on the withdraw `AlertDialog` to reset the error**

Open `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx`. Locate the withdraw dialog opening at line 309:

```tsx
<AlertDialog open={withdrawOpen} onOpenChange={setWithdrawOpen}>
```

Replace it with:

```tsx
<AlertDialog
  open={withdrawOpen}
  onOpenChange={(open) => {
    if (!open) withdrawAction.clearError()
    setWithdrawOpen(open)
  }}
>
```

- [ ] **Step 4: Prevent the `AlertDialogAction` auto-close on the withdraw confirm button**

Still in `game-hero-card.tsx`, locate the withdraw confirm button (lines 324-333):

```tsx
<AlertDialogAction
  onClick={withdrawAction.execute}
  disabled={withdrawAction.isPending}
  className="!bg-amber-600 hover:!bg-amber-700"
>
  {withdrawAction.isPending && (
    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
  )}
  确定撤回
</AlertDialogAction>
```

Replace it with:

```tsx
<AlertDialogAction
  onClick={(e) => {
    e.preventDefault()
    withdrawAction.execute(() => setWithdrawOpen(false))
  }}
  disabled={withdrawAction.isPending}
  className="!bg-amber-600 hover:!bg-amber-700"
>
  {withdrawAction.isPending && (
    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
  )}
  确定撤回
</AlertDialogAction>
```

Note: `e.preventDefault()` cancels Radix's default `DialogClose` behavior (`AlertDialogAction` is a `DialogClose` under the hood), so the dialog stays open until `setWithdrawOpen(false)` runs in the success path.

- [ ] **Step 5: Run lint to verify the changes compile cleanly**

Run from `dx-web/`:

```bash
npm run lint
```

Expected: exits 0 with no new errors or warnings attributed to `use-game-actions.ts` or `game-hero-card.tsx`.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts \
        dx-web/src/features/web/ai-custom/components/game-hero-card.tsx
git commit -m "fix(web): surface withdraw errors inline in ai-custom confirm dialog"
```

---

## Task 2: Fix the 删除 (delete) flow

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts` (`useDeleteGame`, lines 15-34)
- Modify: `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx` (delete dialog, lines 96 / 100 / 252-279)

- [ ] **Step 1: Rewrite `useDeleteGame` to support `onSuccess` + `clearError`**

Open `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts`. Replace the current `useDeleteGame` block (lines 15-34):

```ts
export function useDeleteGame(gameId: string) {
  const router = useRouter()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function execute() {
    setIsPending(true)
    setError(null)
    const result = await deleteGameAction(gameId)
    if (result.error) {
      setError(result.error)
      setIsPending(false)
    } else {
      await swrMutate("/api/course-games")
      router.replace("/hall/ai-custom")
    }
  }

  return { execute, isPending, error }
}
```

with:

```ts
export function useDeleteGame(gameId: string) {
  const router = useRouter()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function execute(onSuccess?: () => void) {
    setIsPending(true)
    setError(null)
    const result = await deleteGameAction(gameId)
    if (result.error) {
      setError(result.error)
      setIsPending(false)
      return
    }
    await swrMutate("/api/course-games")
    onSuccess?.()
    router.replace("/hall/ai-custom")
  }

  const clearError = useCallback(() => setError(null), [])

  return { execute, isPending, error, clearError }
}
```

Note: `useCallback` is already imported from Task 1 Step 1; no additional import change.

- [ ] **Step 2: Wire `onOpenChange` on the delete `AlertDialog` to reset the error**

Open `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx`. Locate the delete dialog opening at line 252:

```tsx
<AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
```

Replace it with:

```tsx
<AlertDialog
  open={deleteOpen}
  onOpenChange={(open) => {
    if (!open) deleteAction.clearError()
    setDeleteOpen(open)
  }}
>
```

- [ ] **Step 3: Prevent the `AlertDialogAction` auto-close on the delete confirm button**

Still in `game-hero-card.tsx`, locate the delete confirm button (lines 267-276):

```tsx
<AlertDialogAction
  onClick={deleteAction.execute}
  disabled={deleteAction.isPending}
  className="bg-red-600 hover:bg-red-700"
>
  {deleteAction.isPending && (
    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
  )}
  确定删除
</AlertDialogAction>
```

Replace it with:

```tsx
<AlertDialogAction
  onClick={(e) => {
    e.preventDefault()
    deleteAction.execute(() => setDeleteOpen(false))
  }}
  disabled={deleteAction.isPending}
  className="bg-red-600 hover:bg-red-700"
>
  {deleteAction.isPending && (
    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
  )}
  确定删除
</AlertDialogAction>
```

- [ ] **Step 4: Run lint to verify the changes compile cleanly**

Run from `dx-web/`:

```bash
npm run lint
```

Expected: exits 0 with no new errors or warnings.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts \
        dx-web/src/features/web/ai-custom/components/game-hero-card.tsx
git commit -m "fix(web): surface delete errors inline in ai-custom confirm dialog"
```

---

## Task 3: Build + manual verification

**Files:** none modified. This task confirms the two flows behave correctly end-to-end.

- [ ] **Step 1: Type-check the whole dx-web build**

Run from `dx-web/`:

```bash
npm run build
```

Expected: build succeeds with no TypeScript errors. (This catches any prop-type drift in `AlertDialog` wrappers that `eslint` alone would miss.)

- [ ] **Step 2: Start dx-api and dx-web**

In one shell:

```bash
cd dx-api && go run .
```

In another:

```bash
cd dx-web && npm run dev
```

Expected: api listens on `:3001`, web on `:3000`, health endpoint `GET /api/health` returns `{"db":true,"redis":true}`.

- [ ] **Step 3: Create an active game session on a published game you own**

As the owner, in a browser:

1. Sign in as the game owner.
2. Navigate to `/hall/ai-custom` and open a published game.
3. Click 去玩 to land on `/hall/games/{id}`.
4. Start a level — once the session opens, close the tab **without finishing**. This leaves `game_sessions.ended_at IS NULL` for that game.

- [ ] **Step 4: Verify the withdraw error displays inline**

1. Go back to `/hall/ai-custom/{id}` for the same game.
2. Click 撤回, then 确定撤回.

Expected:
- The confirm dialog **stays open**.
- A red error paragraph appears inside the dialog with text like "还有 1 个进行中的游戏会话，请等待结束后再撤回".
- The 确定撤回 button re-enables (no permanent spinner).
- The game's status stays `published` (banner and 撤回 button still visible after closing the dialog).

- [ ] **Step 5: Verify the withdraw success flow closes the dialog**

1. Force-end the stale session: `POST /api/play-single/{sessionID}/force-complete` (using the session id you noted, or by completing it in the UI).
2. Back on `/hall/ai-custom/{id}`, click 撤回, then 确定撤回.

Expected:
- Dialog closes.
- Game status flips to `withdraw` (撤回 button disappears, 发布 button appears).
- No stale error paragraph if you reopen the withdraw dialog later (it won't be shown anyway once withdrawn, but confirms no state leaked).

- [ ] **Step 6: Verify the delete error-surface path on a draft game**

1. Create a draft game (don't publish).
2. From the detail page, click 删除.
3. Simulate a backend error: stop `dx-api` (Ctrl-C the `go run .` process).
4. Click 确定删除.

Expected:
- Dialog **stays open**.
- Error paragraph in red: "删除游戏失败" (from the network-error catch in `deleteGameAction`).
- 确定删除 button re-enables.

- [ ] **Step 7: Verify the delete success flow still navigates away**

1. Restart `dx-api` (`go run .`).
2. Click 确定删除 again.

Expected:
- Dialog closes momentarily as the page navigates to `/hall/ai-custom`.
- The deleted game no longer appears in the list.

- [ ] **Step 8: Verify no regressions in the publish flow**

1. On a draft game with at least one level and content, click 发布 → 确定发布.

Expected:
- Publish succeeds.
- Dialog closes (normal Radix auto-close — publish uses `toast`, not inline errors, and was not modified).
- Game flips to `published` status.

- [ ] **Step 9: Verify stale-error reset on dialog reopen**

1. On a published game with an active session (recreate one via Step 3 if needed), open 撤回 and see the inline error.
2. Click 取消 to close the dialog.
3. Click 撤回 again without finishing the session.

Expected: dialog opens fresh with **no** error text visible before you click 确定撤回 again. (The `clearError` on `onOpenChange(false)` cleared the prior message.)

- [ ] **Step 10: Final confirmation**

Report back: all nine manual checks pass with no console errors tied to the modified code paths. The `ws://localhost/api/ws failed` WebSocket message in the console is pre-existing and unrelated (out of scope for this plan).

---

## Self-Review Notes

- **Spec coverage:** Withdraw flow covered in Task 1; delete flow in Task 2; manual verification covers the "Testing" section of the spec (withdraw error display, success close, delete error display, success navigation, reopen reset, publish regression). All spec bullets under "Files Touched" map 1:1 to edits in the tasks.
- **No placeholders:** Every step has the exact code or exact command. No "TBD", no "similar to Task N without code".
- **Type consistency:** Both hooks return `{ execute, isPending, error, clearError }`. Both dialogs read `clearError` off the action and pass a close-callback via `execute(onSuccess)`. `useCallback` imported once in Task 1 Step 1 and reused in Task 2 Step 1 — the imports file is only touched once to avoid churn.
