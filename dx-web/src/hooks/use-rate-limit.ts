import { useRef, useCallback } from "react";
import { toast } from "sonner";

/**
 * Cooldown-based rate limiter for mutation buttons.
 * Returns a `check()` function — call it before executing the action.
 * Returns `true` if allowed, `false` (with toast) if still in cooldown.
 */
export function useRateLimit(cooldownMs = 2_000) {
  const lastOpRef = useRef(0);

  const check = useCallback((): boolean => {
    const now = Date.now();
    if (now - lastOpRef.current < cooldownMs) {
      toast.error("操作过于频繁，请稍后再试");
      return false;
    }
    lastOpRef.current = now;
    return true;
  }, [cooldownMs]);

  return check;
}
