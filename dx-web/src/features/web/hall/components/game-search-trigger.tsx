"use client";

import { Search } from "lucide-react";
import { GameSearchDialog } from "@/features/web/hall/components/game-search-dialog";
import {KbdGroup} from "@/components/ui/kbd";

/** Clickable search bar that opens the game search dialog */
export function GameSearchTrigger({
  placeholder = "搜索课程游戏...",
}: {
  placeholder?: string;
}) {
  return (
    <>
      <GameSearchDialog />
      <button
        type="button"
        onClick={() => {
          document.dispatchEvent(
            new KeyboardEvent("keydown", { key: "k", metaKey: true })
          );
        }}
        className="flex h-10 w-52 items-center gap-2 rounded-[10px] border border-border bg-card px-3 hover:bg-accent"
      >
        <Search className="h-4 w-4 text-muted-foreground" />
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
      </button>
    </>
  );
}
