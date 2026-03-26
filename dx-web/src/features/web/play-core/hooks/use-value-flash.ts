"use client";

import { useRef, useState, useEffect } from "react";

/** Detects value changes and returns a flash key for restarting animations. */
export function useValueFlash(value: number) {
  const prevRef = useRef(value);
  const [flashKey, setFlashKey] = useState(0);

  useEffect(() => {
    if (value !== prevRef.current) {
      prevRef.current = value;
      setFlashKey((k) => k + 1);
    }
  }, [value]);

  return { flashKey };
}
