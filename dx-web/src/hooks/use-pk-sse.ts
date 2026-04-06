"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function usePkSSE(
  pkId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  useEffect(() => { listenersRef.current = listeners; });

  useEffect(() => {
    if (!pkId) return;

    const url = `${API_URL}/api/play-pk/${pkId}/events`;
    const eventSource = new EventSource(url, { withCredentials: true });

    // Use onmessage instead of addEventListener to avoid Safari named-event bug.
    // The server sends: data: {"type":"pk_player_action","payload":{...}}
    eventSource.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data) as { type: string; payload: unknown };
        listenersRef.current[msg.type]?.(msg.payload);
      } catch {
        // Discard malformed SSE messages
      }
    };

    return () => {
      eventSource.close();
    };
  }, [pkId]);
}
