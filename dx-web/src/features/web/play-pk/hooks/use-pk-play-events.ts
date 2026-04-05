"use client";

import { useRef, useMemo, useEffect } from "react";
import { usePkSSE } from "@/hooks/use-pk-sse";
import type {
  PkForceEndEvent,
  PkPlayerCompleteEvent,
  PkPlayerActionEvent,
  PkTimeoutWarningEvent,
  PkTimeoutEvent,
} from "../types/pk-play";

type PkPlayEventHandlers = {
  onForceEnd?: (event: PkForceEndEvent) => void;
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
    pk_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as PkForceEndEvent),
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
