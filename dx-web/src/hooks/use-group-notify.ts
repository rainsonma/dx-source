"use client";

import { useEffect, useRef } from "react";
import { getToken, refreshAccessToken } from "@/lib/api-client";

const MAX_RETRIES = 10;

function backoffDelay(retryCount: number): number {
  return Math.min(1000 * Math.pow(2, retryCount - 1), 30000);
}

export function useGroupNotify(
  groupId: string | null,
  onUpdate: (scope: string) => void
): void {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  useEffect(() => {
    if (!groupId) return;

    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let disposed = false;

    function connect(token: string) {
      if (disposed) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "";
      const url = `${apiUrl}/api/groups/${groupId}/notify?token=${encodeURIComponent(token)}`;

      eventSource = new EventSource(url);

      eventSource.addEventListener("group_updated", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data) as { scope: string };
          callbackRef.current(data.scope);
        } catch {
          // Discard malformed SSE messages
        }
      });

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
