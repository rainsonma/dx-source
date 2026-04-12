"use client";

import { useTopic } from "@/hooks/use-topic";

export function useUserEvents(
  userId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(userId ? `user:${userId}` : null, listeners);
}
