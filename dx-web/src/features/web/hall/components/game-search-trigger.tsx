"use client";

import { usePathname } from "next/navigation";
import { Search, X } from "lucide-react";
import { GameSearchDialog } from "@/features/web/hall/components/game-search-dialog";
import { useGameSearchText } from "@/features/web/games/stores/game-search-store";
import { KbdGroup } from "@/components/ui/kbd";

/** Clickable search bar that opens the game search dialog */
export function GameSearchTrigger({
  placeholder = "搜索课程游戏...",
}: {
  placeholder?: string;
}) {
  const pathname = usePathname();
  const q = useGameSearchText((s) => s.q);
  const clearQ = useGameSearchText((s) => s.clearQ);

  const showActiveSearch = q && pathname === "/hall/games";

  function openDialog() {
    document.dispatchEvent(
      new KeyboardEvent("keydown", { key: "k", metaKey: true })
    );
  }

  return (
    <>
      <GameSearchDialog />
      <button
        type="button"
        onClick={openDialog}
        className="flex h-10 w-52 items-center gap-2 rounded-[10px] border border-border bg-card px-3 hover:bg-accent"
      >
        <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
        {showActiveSearch ? (
          <>
            <span className="flex-1 truncate text-left text-[13px] text-foreground">
              {q}
            </span>
            <button
              type="button"
              aria-label="清除搜索"
              onClick={(e) => {
                e.stopPropagation();
                clearQ();
              }}
              className="flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground hover:text-foreground"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </>
        ) : (
          <>
            <span className="flex-1 text-left text-[13px] text-muted-foreground">
              {placeholder}
            </span>
            <KbdGroup>
              <kbd className="pointer-events-none hidden h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 font-mono text-muted-foreground sm:inline-flex">
                ⌘
              </kbd>
              <kbd className="pointer-events-none hidden h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 text-[10px] font-mono text-muted-foreground sm:inline-flex">
                k
              </kbd>
            </KbdGroup>
          </>
        )}
      </button>
    </>
  );
}
