"use client";

import { useTopic } from "@/hooks/use-topic";

export function useGroupEvents(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(groupId ? `group:${groupId}` : null, listeners);
}
