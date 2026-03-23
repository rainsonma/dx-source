"use client";

import { useState, useEffect, useRef } from "react";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";

/** Module-level elapsed time — readable without re-renders */
let currentElapsed = 0;

/** Get current elapsed seconds without subscribing to state changes */
export function getElapsedSeconds(): number {
  return currentElapsed;
}

/** Format seconds into H:MM:SS string */
export function formatElapsedTime(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const mm = String(minutes).padStart(2, "0");
  const ss = String(seconds).padStart(2, "0");
  return `${hours}:${mm}:${ss}`;
}

/** Accumulating game timer that pauses on overlays and supports restoring from server */
export function useGameTimer(initialSeconds: number = 0) {
  const phase = useGameStore((s) => s.phase);
  const overlay = useGameStore((s) => s.overlay);
  const [elapsed, setElapsed] = useState(initialSeconds);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const isRunning = phase === "playing" && overlay === null;

  // Reset or restore on game restart / resume
  useEffect(() => {
    if (phase === "loading" || initialSeconds > elapsed) {
      setElapsed(initialSeconds);
      currentElapsed = initialSeconds;
    }
  }, [phase, initialSeconds]);

  // Start/stop interval based on running state
  useEffect(() => {
    if (isRunning) {
      intervalRef.current = setInterval(() => {
        setElapsed((prev) => {
          const next = prev + 1;
          currentElapsed = next;
          return next;
        });
      }, 1000);
    } else if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [isRunning]);

  return { elapsed, formatted: formatElapsedTime(elapsed) };
}
