# Play Page Fullscreen Toggle Design

**Date**: 2026-03-15

## Overview

Add fullscreen toggle to the game play page using the Browser Fullscreen API. The fullscreen button already exists visually in the top bar (`Maximize` icon) but has no handler. This design wires it up.

## Approach

Browser Fullscreen API (`document.documentElement.requestFullscreen()` / `document.exitFullscreen()`). ESC key exit is handled natively by the browser. Top bar remains visible in fullscreen.

## Components

### `use-fullscreen.ts` (new hook)

- `isFullscreen` boolean state
- `toggleFullscreen()` — enters or exits fullscreen
- Listens to `fullscreenchange` event to sync state (covers ESC exit)
- Handles Safari `webkit` prefix
- Cleans up listener on unmount

### `game-top-bar.tsx` (modify)

- Add `onFullscreen` callback prop and `isFullscreen` boolean prop
- Wire `onFullscreen` into `actionHandlers` for `"fullscreen"` action
- Swap icon: `Minimize` when fullscreen, `Maximize` when not

### `game-play-shell.tsx` (modify)

- Call `useFullscreen()` hook
- Pass `onFullscreen={toggleFullscreen}` and `isFullscreen` to `GameTopBar`

## What doesn't change

- Game store, timer, overlays, game components — no modifications needed
- Layout already uses `h-screen w-full`, fills fullscreen naturally
