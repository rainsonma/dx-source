"use client";

import { useRef, useMemo, useEffect } from "react";
import { useGroupSSE } from "@/hooks/use-group-sse";
import type {
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
  GroupPlayerCompleteEvent,
  GroupPlayerActionEvent,
  GroupNextLevelEvent,
} from "../types/group-play";

type GroupPlayEventHandlers = {
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onPlayerComplete?: (event: GroupPlayerCompleteEvent) => void;
  onPlayerAction?: (event: GroupPlayerActionEvent) => void;
  onNextLevel?: (event: GroupNextLevelEvent) => void;
  onDismissed?: () => void;
};

export function useGroupPlayEvents(
  groupId: string | null,
  handlers: GroupPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  const listeners = useMemo(() => ({
    group_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    group_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as GroupPlayerCompleteEvent),
    group_player_action: (data: unknown) =>
      handlersRef.current.onPlayerAction?.(data as GroupPlayerActionEvent),
    group_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as GroupNextLevelEvent),
    group_dismissed: () => handlersRef.current.onDismissed?.(),
  }), []);

  useGroupSSE(groupId, listeners);
}
