"use client";

import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from "@/components/ui/command";
import { GAME_MODE_LABELS } from "@/consts/game-mode";
import { useGameSearch } from "@/features/web/hall/hooks/use-game-search";

/** Cmd+K game search dialog with server-side fuzzy matching and recent suggestions */
export function GameSearchDialog() {
  const router = useRouter();
  const {
    isOpen,
    setIsOpen,
    query,
    setQuery,
    displayItems,
    groupLabel,
    showGroup,
    isLoading,
  } = useGameSearch();

  /** Navigate to game detail and close dialog */
  function handleSelect(gameId: string) {
    router.push(`/hall/games/${gameId}`);
    setIsOpen(false);
  }

  /** Format the mode label for display */
  function getModeLabel(mode: string): string {
    return GAME_MODE_LABELS[mode as keyof typeof GAME_MODE_LABELS] ?? mode;
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogHeader className="sr-only">
        <DialogTitle>搜索课程游戏</DialogTitle>
        <DialogDescription>输入课程游戏名称搜索</DialogDescription>
      </DialogHeader>
      <DialogContent className="top-[10%] translate-y-0 overflow-hidden p-2" showCloseButton={false}>
        <Command
          shouldFilter={false}
          className="[&_[cmdk-group-heading]]:text-muted-foreground **:data-[slot=command-input-wrapper]:h-12 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group]]:px-2 [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5"
        >
          <CommandInput
            placeholder="输入课程游戏名称搜索..."
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {isLoading && (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}

            {!isLoading && displayItems.length === 0 && (
              <CommandEmpty>
                {query.trim() ? "未找到相关数据" : "暂无数据"}
              </CommandEmpty>
            )}

            {!isLoading && showGroup && (
              <CommandGroup heading={groupLabel}>
                {displayItems.map((game) => (
                  <CommandItem
                    key={game.id}
                    value={game.name}
                    onSelect={() => handleSelect(game.id)}
                    className="cursor-pointer"
                  >
                    <div className="flex flex-col gap-0.5">
                      <span className="text-sm font-medium">{game.name}</span>
                      <span className="text-xs text-muted-foreground">
                        {getModeLabel(game.mode)}
                        {game.category && ` · ${game.category.name}`}
                      </span>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
