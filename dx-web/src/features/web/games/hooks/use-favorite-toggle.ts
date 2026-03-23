"use client";

import { useState, useRef, useTransition, useCallback } from "react";
import { toast } from "sonner";
import { toggleFavoriteAction } from "@/features/web/games/actions/favorite.action";

const COOLDOWN_MS = 2000;

/** Toggle favorite with optimistic UI, 2-second cooldown, and toast feedback */
export function useFavoriteToggle(
  gameId: string,
  gameName: string,
  initialFavorited: boolean
) {
  const [favorited, setFavorited] = useState(initialFavorited);
  const [isPending, startTransition] = useTransition();
  const lastToggleRef = useRef(0);

  const toggle = useCallback(() => {
    const now = Date.now();
    if (now - lastToggleRef.current < COOLDOWN_MS) {
      toast.warning("操作频繁，请稍后再试");
      return;
    }

    if (isPending) return;

    const prev = favorited;
    setFavorited(!prev);
    lastToggleRef.current = now;

    startTransition(async () => {
      const result = await toggleFavoriteAction(gameId);

      if ("error" in result) {
        setFavorited(prev);
        toast.error(result.error);
        return;
      }

      setFavorited(result.favorited);
      toast.success(
        result.favorited
          ? `已收藏「${gameName}」`
          : `已取消收藏「${gameName}」`
      );
    });
  }, [gameId, gameName, favorited, isPending, startTransition]);

  return { favorited, toggle, isPending };
}
