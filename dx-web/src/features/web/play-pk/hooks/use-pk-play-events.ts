"use client";

import { useRef, useMemo, useEffect } from "react";
import { usePkSSE } from "@/hooks/use-pk-sse";
import type {
  PkLevelCompleteEvent,
  PkForceEndEvent,
  PkNextLevelEvent,
  PkPlayerCompleteEvent,
  PkPlayerActionEvent,
  PkTimeoutWarningEvent,
  PkTimeoutEvent,
} from "../types/pk-play";

type PkPlayEventHandlers = {
  onLevelComplete?: (event: PkLevelCompleteEvent) => void;
  onForceEnd?: (event: PkForceEndEvent) => void;
  onNextLevel?: (event: PkNextLevelEvent) => void;
  onPlayerComplete?: (event: PkPlayerCompleteEvent) => void;
  onPlayerAction?: (event: PkPlayerActionEvent) => void;
  onTimeoutWarning?: (event: PkTimeoutWarningEvent) => void;
  onTimeout?: (event: PkTimeoutEvent) => void;
};

export function usePkPlayEvents(
  pkId: string | null,
  handlers: PkPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  const listeners = useMemo(() => ({
    pk_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as PkLevelCompleteEvent),
    pk_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as PkForceEndEvent),
    pk_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as PkNextLevelEvent),
    pk_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as PkPlayerCompleteEvent),
    pk_player_action: (data: unknown) =>
      handlersRef.current.onPlayerAction?.(data as PkPlayerActionEvent),
    pk_timeout_warning: (data: unknown) =>
      handlersRef.current.onTimeoutWarning?.(data as PkTimeoutWarningEvent),
    pk_timeout: (data: unknown) =>
      handlersRef.current.onTimeout?.(data as PkTimeoutEvent),
  }), []);

  usePkSSE(pkId, listeners);
}
