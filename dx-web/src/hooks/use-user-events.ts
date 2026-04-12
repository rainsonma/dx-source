"use client";

import { useWS } from "@/providers/websocket-provider";
import { useTopic } from "@/hooks/use-topic";

export function useUserEvents(
  listeners: Record<string, (data: unknown) => void>,
): void {
  const { userId } = useWS();
  useTopic(userId ? `user:${userId}` : null, listeners);
}
