// dx-web/src/features/web/home/hooks/use-rotating-scene.ts
"use client";

import { useEffect, useState } from "react";

interface Options {
  total: number;
  intervalMs?: number;
  paused?: boolean;
}

/** Cycles an index [0..total) while not paused. */
export function useRotatingScene({
  total,
  intervalMs = 6000,
  paused = false,
}: Options): { index: number; setIndex: (i: number) => void } {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (paused || total <= 1) return;
    const id = window.setInterval(() => {
      setIndex((i) => (i + 1) % total);
    }, intervalMs);
    return () => window.clearInterval(id);
  }, [paused, total, intervalMs]);

  return { index, setIndex };
}
