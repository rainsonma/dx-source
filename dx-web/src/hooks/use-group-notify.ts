"use client";

import { useTopic } from "@/hooks/use-topic";

export function useGroupNotify(
  groupId: string | null,
  onScope: (scope: string) => void,
): void {
  useTopic(
    groupId ? `group:${groupId}:notify` : null,
    {
      group_updated: (data) => {
        const d = data as { scope?: string };
        if (d.scope) onScope(d.scope);
      },
    },
  );
}
