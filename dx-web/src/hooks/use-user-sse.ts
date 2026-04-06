"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useUserSSE(
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  useEffect(() => { listenersRef.current = listeners; });

  useEffect(() => {
    // Mark online immediately (bridges gap before SSE connects)
    fetch(`${API_URL}/api/user/ping`, { method: "POST", credentials: "include" }).catch(() => {});

    const url = `${API_URL}/api/user/events`;
    const eventSource = new EventSource(url, { withCredentials: true });

    // Use onmessage instead of addEventListener to avoid Safari named-event bug.
    // The server sends: data: {"type":"pk_invitation","payload":{...}}
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
  }, []);
}
