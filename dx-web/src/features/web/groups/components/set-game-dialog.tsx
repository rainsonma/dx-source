"use client";

import { useState, useEffect, useCallback } from "react";
import { Gamepad2, Search, User, Users, Check, X, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { apiClient } from "@/lib/api-client";
import { swrMutate } from "@/lib/swr";
import { groupApi } from "../actions/group.action";
import type { GroupGameSearchItem } from "../types/group";

interface SetGameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  groupId: string;
  currentGameId?: string | null;
  currentGameMode?: string | null;
  currentStartLevelId?: string | null;
}

export function SetGameDialog({
  open,
  onOpenChange,
  groupId,
  currentGameId,
  currentGameMode,
  currentStartLevelId,
}: SetGameDialogProps) {
  const [selectedGameId, setSelectedGameId] = useState<string | null>(null);
  const [selectedMode, setSelectedMode] = useState<"group_solo" | "group_team">("group_solo");
  const [searchQuery, setSearchQuery] = useState("");
  const [games, setGames] = useState<GroupGameSearchItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [levels, setLevels] = useState<{ id: string; name: string; order: number }[]>([]);
  const [selectedLevelId, setSelectedLevelId] = useState<string | null>(null);

  const fetchGames = useCallback(
    async (q: string, limit?: number) => {
      setLoading(true);
      const res = await groupApi.searchGamesForGroup(groupId, q || undefined, limit);
      setLoading(false);
      if (res.code === 0) setGames(res.data);
    },
    [groupId]
  );

  // On open: initialize and load latest 3 games
  useEffect(() => {
    if (open) {
      setSelectedGameId(currentGameId ?? null);
      setSelectedMode(currentGameMode === "group_team" ? "group_team" : "group_solo");
      setSearchQuery("");
      setIsSearching(false);
      fetchGames("", 3);
    }
  }, [open, currentGameId, currentGameMode, fetchGames]);

  // Debounced search
  useEffect(() => {
    if (!open) return;
    if (searchQuery === "") {
      setIsSearching(false);
      fetchGames("", 3);
      return;
    }
    setIsSearching(true);
    const timer = setTimeout(() => {
      fetchGames(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, open, fetchGames]);

  // Fetch levels when a game is selected
  useEffect(() => {
    if (!selectedGameId) {
      setLevels([]);
      setSelectedLevelId(null);
      return;
    }
    async function fetchLevels() {
      const res = await apiClient.get<{ levels?: { id: string; name: string; order: number }[] }>(`/api/games/${selectedGameId}`);
      if (res.code === 0 && res.data?.levels) {
        const lvls = res.data.levels.map((l) => ({ id: l.id, name: l.name, order: l.order }));
        setLevels(lvls);
        setSelectedLevelId(currentStartLevelId ?? lvls[0]?.id ?? null);
      }
    }
    fetchLevels();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- only re-fetch when selected game changes
  }, [selectedGameId]);

  async function handleConfirm() {
    if (!selectedGameId) {
      toast.error("请选择一个游戏");
      return;
    }
    setConfirming(true);
    const res = await groupApi.setGame(groupId, selectedGameId, selectedMode, selectedLevelId ?? undefined);
    setConfirming(false);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("课程游戏已设置");
    await swrMutate(`/api/groups/${groupId}`);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className="gap-0 p-0 sm:max-w-lg" aria-describedby={undefined}>
        <DialogTitle className="sr-only">设置群课程游戏</DialogTitle>
        {/* Header */}
        <div className="flex items-center justify-between border-b px-6 py-5">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-teal-100">
              <Gamepad2 className="h-4.5 w-4.5 text-teal-600" />
            </div>
            <span className="text-[15px] font-semibold text-foreground">设置群课程游戏</span>
          </div>
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Body */}
        <div className="flex flex-col gap-4 px-6 py-5">
          {/* Search input */}
          <div className="flex h-[42px] items-center gap-2 rounded-[10px] border border-border bg-slate-50 px-3">
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="搜索游戏..."
              className="flex-1 bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery("")}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>

          {/* Section label */}
          <span className="text-[11px] font-medium text-muted-foreground">
            {isSearching ? "搜索结果" : "最新游戏"}
          </span>

          {/* Game list */}
          <div className="h-64 overflow-y-auto rounded-[10px] border border-border">
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
              </div>
            ) : games.length === 0 ? (
              <div className="flex items-center justify-center py-8 text-xs text-muted-foreground">
                {isSearching ? "未找到匹配的游戏" : "暂无游戏"}
              </div>
            ) : (
              games.map((game, idx) => {
                const isSelected = selectedGameId === game.id;
                return (
                  <div key={game.id}>
                    {idx > 0 && <div className="h-px bg-border" />}
                    <button
                      type="button"
                      onClick={() => setSelectedGameId(game.id)}
                      className={`flex w-full items-center gap-3 px-4 py-3 text-left transition-colors ${
                        isSelected ? "bg-teal-50" : "hover:bg-muted/50"
                      }`}
                    >
                      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-teal-100">
                        <span className="text-xs font-bold text-teal-600">{game.name[0]}</span>
                      </div>
                      <div className="flex flex-1 flex-col gap-0.5 overflow-hidden">
                        <span className="truncate text-[13px] font-medium text-foreground">{game.name}</span>
                        {game.category_name && (
                          <span className="text-[11px] text-muted-foreground">{game.category_name}</span>
                        )}
                      </div>
                      <div
                        className={`flex h-4.5 w-4.5 shrink-0 items-center justify-center rounded-full border-2 ${
                          isSelected ? "border-teal-600 bg-teal-600" : "border-border"
                        }`}
                      >
                        {isSelected && <div className="h-1.5 w-1.5 rounded-full bg-white" />}
                      </div>
                    </button>
                  </div>
                );
              })
            )}
          </div>

          {/* Level selector */}
          {levels.length > 0 && (
            <div className="flex items-center gap-3">
              <Gamepad2 className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="shrink-0 text-[13px] font-medium text-foreground">起始关卡</span>
              <Select value={selectedLevelId ?? ""} onValueChange={setSelectedLevelId}>
                <SelectTrigger className="h-9 flex-1 text-sm">
                  <SelectValue placeholder="选择关卡" />
                </SelectTrigger>
                <SelectContent>
                  {levels.map((level) => (
                    <SelectItem key={level.id} value={level.id}>
                      {level.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Mode selector */}
          <div className="flex gap-1 rounded-[10px] bg-slate-100 p-1">
            <button
              type="button"
              onClick={() => setSelectedMode("group_solo")}
              className={`flex h-9 flex-1 items-center justify-center gap-1.5 rounded-lg text-[13px] font-medium transition-colors ${
                selectedMode === "group_solo"
                  ? "bg-teal-600 text-white"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <User className="h-3.5 w-3.5" />
              单人对抗
            </button>
            <button
              type="button"
              onClick={() => setSelectedMode("group_team")}
              className={`flex h-9 flex-1 items-center justify-center gap-1.5 rounded-lg text-[13px] font-medium transition-colors ${
                selectedMode === "group_team"
                  ? "bg-teal-600 text-white"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <Users className="h-3.5 w-3.5" />
              小组对抗
            </button>
          </div>

          {/* Confirm button */}
          <button
            type="button"
            onClick={handleConfirm}
            disabled={confirming || !selectedGameId}
            className="flex h-[50px] w-full items-center justify-center gap-2 rounded-[12px] bg-teal-600 text-[14px] font-semibold text-white transition-opacity hover:bg-teal-700 disabled:opacity-50"
          >
            {confirming ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Check className="h-4 w-4" />
            )}
            确认设置
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
