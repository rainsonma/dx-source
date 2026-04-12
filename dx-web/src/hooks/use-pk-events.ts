"use client";

import { useTopic } from "@/hooks/use-topic";

export function usePkEvents(
  pkId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(pkId ? `pk:${pkId}` : null, listeners);
}
