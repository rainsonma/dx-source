"use client";

import { useRef, useMemo } from "react";
import { useGroupSSE } from "@/hooks/use-group-sse";
import type {
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
  GroupNextLevelEvent,
  GroupPlayerCompleteEvent,
  GroupPlayerActionEvent,
} from "../types/group-play";

type GroupPlayEventHandlers = {
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onNextLevel?: (event: GroupNextLevelEvent) => void;
  onPlayerComplete?: (event: GroupPlayerCompleteEvent) => void;
  onPlayerAction?: (event: GroupPlayerActionEvent) => void;
  onDismissed?: () => void;
};

export function useGroupPlayEvents(
  groupId: string | null,
  handlers: GroupPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  const listeners = useMemo(() => ({
    group_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    group_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as GroupNextLevelEvent),
    group_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as GroupPlayerCompleteEvent),
    group_player_action: (data: unknown) =>
      handlersRef.current.onPlayerAction?.(data as GroupPlayerActionEvent),
    group_dismissed: () => handlersRef.current.onDismissed?.(),
  }), []);

  useGroupSSE(groupId, listeners);
}
