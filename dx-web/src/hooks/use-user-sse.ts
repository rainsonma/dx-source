"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useUserSSE(
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  useEffect(() => { listenersRef.current = listeners; });

  useEffect(() => {
    const url = `${API_URL}/api/user/events`;
    const eventSource = new EventSource(url, { withCredentials: true });

    for (const eventName of Object.keys(listenersRef.current)) {
      eventSource.addEventListener(eventName, (e: MessageEvent) => {
        try {
          const data: unknown = JSON.parse(e.data);
          listenersRef.current[eventName]?.(data);
        } catch {
          // Discard malformed SSE messages
        }
      });
    }

    return () => {
      eventSource.close();
    };
  }, []);
}
