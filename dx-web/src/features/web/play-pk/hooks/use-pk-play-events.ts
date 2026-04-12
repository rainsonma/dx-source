"use client";

import { useRef, useMemo, useEffect } from "react";
import { usePkEvents } from "@/hooks/use-pk-events";
import type {
  PkForceEndEvent,
  PkPlayerCompleteEvent,
  PkPlayerActionEvent,
  PkNextLevelEvent,
} from "../types/pk-play";

type PkPlayEventHandlers = {
  onForceEnd?: (event: PkForceEndEvent) => void;
  onPlayerComplete?: (event: PkPlayerCompleteEvent) => void;
  onPlayerAction?: (event: PkPlayerActionEvent) => void;
  onNextLevel?: (event: PkNextLevelEvent) => void;
};

export function usePkPlayEvents(
  pkId: string | null,
  handlers: PkPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  const listeners = useMemo(() => ({
    pk_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as PkForceEndEvent),
    pk_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as PkPlayerCompleteEvent),
    pk_player_action: (data: unknown) =>
      handlersRef.current.onPlayerAction?.(data as PkPlayerActionEvent),
    pk_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as PkNextLevelEvent),
  }), []);

  usePkEvents(pkId, listeners);
}
