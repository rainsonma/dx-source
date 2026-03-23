"use client";

import { useRef, useCallback, useEffect } from "react";
import { useGameSettings } from "@/features/web/play/hooks/use-game-settings";

const SOUND_URL = "/sounds/typewriter-1.mp3";

/** Play a typewriter key-strike sound via Web Audio API */
export function useKeySound() {
  const ctxRef = useRef<AudioContext | null>(null);
  const bufferRef = useRef<AudioBuffer | null>(null);
  const typingSoundEnabled = useGameSettings((s) => s.typingSoundEnabled);

  /** Ensure the AudioContext is created and resumed (handles autoplay policy) */
  const ensureContext = useCallback((): AudioContext => {
    if (!ctxRef.current) {
      ctxRef.current = new AudioContext();
    }
    if (ctxRef.current.state === "suspended") {
      ctxRef.current.resume();
    }
    return ctxRef.current;
  }, []);

  /** Fetch and decode the audio file into an AudioBuffer */
  useEffect(() => {
    let cancelled = false;

    async function loadSound() {
      try {
        const ctx = ensureContext();
        const response = await fetch(SOUND_URL);
        const arrayBuffer = await response.arrayBuffer();
        const audioBuffer = await ctx.decodeAudioData(arrayBuffer);
        if (!cancelled) {
          bufferRef.current = audioBuffer;
        }
      } catch {
        // Silently ignore — sound is non-critical
      }
    }

    loadSound();

    return () => {
      cancelled = true;
    };
  }, [ensureContext]);

  /** Clean up AudioContext on unmount */
  useEffect(() => {
    return () => {
      if (ctxRef.current) {
        ctxRef.current.close();
        ctxRef.current = null;
      }
    };
  }, []);

  /** Play the typewriter sound instantly by creating a new source node */
  const playKeySound = useCallback(() => {
    if (!typingSoundEnabled) return;

    const ctx = ctxRef.current;
    const buffer = bufferRef.current;
    if (!ctx || !buffer) return;

    if (ctx.state === "suspended") {
      ctx.resume();
    }

    const source = ctx.createBufferSource();
    source.buffer = buffer;
    source.connect(ctx.destination);
    source.start(0);
  }, [typingSoundEnabled]);

  return { playKeySound };
}
