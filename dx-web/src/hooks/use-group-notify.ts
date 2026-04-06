"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useGroupNotify(
  groupId: string | null,
  onUpdate: (scope: string) => void
): void {
  const callbackRef = useRef(onUpdate);
  useEffect(() => {
    callbackRef.current = onUpdate;
  });

  useEffect(() => {
    if (!groupId) return;

    const url = `${API_URL}/api/groups/${groupId}/notify`;
    const eventSource = new EventSource(url, { withCredentials: true });

    eventSource.addEventListener("group_updated", (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data) as { scope: string };
        callbackRef.current(data.scope);
      } catch {
        // Discard malformed SSE messages
      }
    });

    return () => {
      eventSource.close();
    };
  }, [groupId]);
}
