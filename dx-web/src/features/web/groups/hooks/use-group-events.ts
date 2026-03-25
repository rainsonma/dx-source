"use client";

import { useEffect, useRef } from "react";
import type {
  GroupGameStartEvent,
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
} from "../types/group";

type GroupEventHandlers = {
  onGameStart?: (event: GroupGameStartEvent) => void;
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
};

export function useGroupEvents(
  groupId: string | null,
  handlers: GroupEventHandlers
) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    if (!groupId) return;

    const token = localStorage.getItem("dx_token");
    if (!token) return;

    const apiUrl = process.env.NEXT_PUBLIC_API_URL;
    const url = `${apiUrl}/api/groups/${groupId}/events?token=${encodeURIComponent(token)}`;

    const eventSource = new EventSource(url);

    eventSource.addEventListener("group_game_start", (e) => {
      const data: GroupGameStartEvent = JSON.parse(e.data);
      handlersRef.current.onGameStart?.(data);
    });

    eventSource.addEventListener("group_level_complete", (e) => {
      const data: GroupLevelCompleteEvent = JSON.parse(e.data);
      handlersRef.current.onLevelComplete?.(data);
    });

    eventSource.addEventListener("group_game_force_end", (e) => {
      const data: GroupForceEndEvent = JSON.parse(e.data);
      handlersRef.current.onForceEnd?.(data);
    });

    return () => eventSource.close();
  }, [groupId]);
}
