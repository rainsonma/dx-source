"use client";

import { useEffect, useRef } from "react";
import { getAccessToken } from "@/lib/token";
import type { GroupLevelCompleteEvent, GroupForceEndEvent, GroupNextLevelEvent } from "../types/group-play";

type GroupPlayEventHandlers = {
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onNextLevel?: (event: GroupNextLevelEvent) => void;
};

export function useGroupPlayEvents(
  groupId: string | null,
  handlers: GroupPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    if (!groupId) return;

    const token = getAccessToken();
    if (!token) return;

    const apiUrl = process.env.NEXT_PUBLIC_API_URL;
    const url = `${apiUrl}/api/groups/${groupId}/events?token=${encodeURIComponent(token)}`;

    const eventSource = new EventSource(url);

    eventSource.addEventListener("group_level_complete", (e) => {
      const data: GroupLevelCompleteEvent = JSON.parse(e.data);
      handlersRef.current.onLevelComplete?.(data);
    });

    eventSource.addEventListener("group_game_force_end", (e) => {
      const data: GroupForceEndEvent = JSON.parse(e.data);
      handlersRef.current.onForceEnd?.(data);
    });

    eventSource.addEventListener("group_next_level", (e) => {
      const data: GroupNextLevelEvent = JSON.parse(e.data);
      handlersRef.current.onNextLevel?.(data);
    });

    return () => eventSource.close();
  }, [groupId]);
}
