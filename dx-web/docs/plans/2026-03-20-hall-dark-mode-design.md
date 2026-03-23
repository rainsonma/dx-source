# Hall Dark Mode Design

## Goal

Add light/dark mode toggle to all hall pages (including play pages) via the existing Sun icon button in `top-actions.tsx`. Must not break any existing functionality.

## Current State

- `next-themes` v0.4.6 installed but ThemeProvider not configured
- Dark CSS variables fully defined in `globals.css` (`.dark` class with oklch values)
- Dark variant configured: `@custom-variant dark (&:is(.dark *));`
- Sun button in `top-actions.tsx` exists but has no functionality
- ~296 hardcoded light-mode color references across ~39 files in hall/play features
- Semantic color tokens (bg-card, text-foreground, etc.) almost unused in hall/play

## Decisions

- **Scope**: All pages under `/hall` including play pages
- **Provider placement**: Hall-scoped ThemeProvider (not root-level) to avoid affecting auth/admin pages
- **Color strategy**: Replace hardcoded colors with semantic tokens (not dark: prefix)
- **Accent colors**: Keep teal, blue, emerald, violet unchanged — they work in both modes
- **Persistence**: localStorage via next-themes (storageKey: "hall-theme")

## Architecture

### ThemeProvider

New client component `src/features/web/hall/components/hall-theme-provider.tsx`:

```tsx
"use client";
import { ThemeProvider } from "next-themes";

export function HallThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="light" storageKey="hall-theme">
      {children}
    </ThemeProvider>
  );
}
```

Wraps children in `src/app/(web)/hall/layout.tsx`.

### Theme Toggle Button

New client component `src/features/web/hall/components/theme-toggle-button.tsx`:
- Uses `useTheme()` from `next-themes`
- Toggles between `light` and `dark`
- Shows Sun icon in light mode, Moon icon in dark mode
- Same visual styling as existing icon buttons (h-10 w-10 rounded-[10px] border)

Replaces the static Sun button in `top-actions.tsx`.

### Semantic Token Migration

| Hardcoded | Semantic Replacement |
|-----------|---------------------|
| `bg-white` | `bg-card` |
| `bg-slate-50` | `bg-muted` |
| `bg-slate-100` | `bg-muted` |
| `text-slate-900` | `text-foreground` |
| `text-slate-700` | `text-foreground` |
| `text-slate-500` | `text-muted-foreground` |
| `text-slate-400` | `text-muted-foreground` |
| `border-slate-200` | `border-border` |
| `border-slate-100` | `border-border` |
| `hover:bg-slate-50` | `hover:bg-accent` |
| `hover:bg-slate-100` | `hover:bg-accent` |
| `bg-slate-900/50` (overlays) | Keep as-is (works both modes) |

## Files Affected

- **New files**: hall-theme-provider.tsx, theme-toggle-button.tsx
- **Modified infrastructure**: hall layout.tsx, top-actions.tsx
- **Color migration**: ~39 files in features/web/hall/ and features/web/play/
- **CSS**: No changes needed — dark variables already defined

## Risks

- **Visual regressions**: Some components may look odd with semantic tokens if the mapping isn't exact. Visual testing after each batch of changes is essential.
- **SSR flash**: next-themes handles this via script injection, but hall-scoped provider means the script only runs for hall pages.
- **Third-party components**: shadcn/ui components already use semantic tokens and should work automatically. Custom components need manual migration.
