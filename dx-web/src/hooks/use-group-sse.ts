"use client";

import { useEffect, useRef } from "react";
import { getToken, refreshAccessToken } from "@/lib/api-client";

const MAX_RETRIES = 10;

function backoffDelay(retryCount: number): number {
  return Math.min(1000 * Math.pow(2, retryCount - 1), 30000);
}

export function useGroupSSE(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  listenersRef.current = listeners;

  useEffect(() => {
    if (!groupId) return;

    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let disposed = false;

    function attachListeners(es: EventSource) {
      for (const eventName of Object.keys(listenersRef.current)) {
        es.addEventListener(eventName, (e: MessageEvent) => {
          try {
            const data: unknown = JSON.parse(e.data);
            listenersRef.current[eventName]?.(data);
          } catch {
            // Discard malformed SSE messages
          }
        });
      }
    }

    function connect(token: string) {
      if (disposed) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "";
      const url = `${apiUrl}/api/groups/${groupId}/events?token=${encodeURIComponent(token)}`;

      eventSource = new EventSource(url);

      attachListeners(eventSource);

      eventSource.onopen = () => {
        retryCount = 0;
      };

      eventSource.onerror = () => {
        if (disposed) return;
        eventSource?.close();
        eventSource = null;
        scheduleReconnect();
      };
    }

    function scheduleReconnect() {
      if (disposed) return;
      retryCount++;
      if (retryCount > MAX_RETRIES) return;

      const delay = backoffDelay(retryCount);
      retryTimer = setTimeout(() => {
        if (disposed) return;
        refreshAndConnect();
      }, delay);
    }

    function refreshAndConnect() {
      if (disposed) return;
      refreshAccessToken()
        .then((token) => {
          if (!disposed) connect(token);
        })
        .catch(() => {
          // On auth failure, refreshAccessToken redirects to /auth/signin
          // (which triggers cleanup via disposed flag).
          // On transient network failure, schedule another retry.
          scheduleReconnect();
        });
    }

    // Initial connection: use existing token or refresh first
    const token = getToken();
    if (token) {
      connect(token);
    } else {
      refreshAndConnect();
    }

    return () => {
      disposed = true;
      if (retryTimer) clearTimeout(retryTimer);
      eventSource?.close();
      eventSource = null;
    };
  }, [groupId]);
}
