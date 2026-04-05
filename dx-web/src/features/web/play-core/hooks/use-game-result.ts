"use client";

import { useState, useEffect, useRef } from "react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";

interface UseGameResultParams {
  levels: { id: string; order: number }[];
}

/** Auto-fires completeLevel (and endSession if last level) on mount */
export function useGameResult({ levels }: UseGameResultParams) {
  const [completing, setCompleting] = useState(true);
  const firedRef = useRef(false);

  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const score = useGameStore((s) => s.score);
  const combo = useGameStore((s) => s.combo);
  const correctCount = useGameStore((s) => s.correctCount);
  const wrongCount = useGameStore((s) => s.wrongCount);
  const skipCount = useGameStore((s) => s.skipCount);
  const contentItems = useGameStore((s) => s.contentItems);

  const { completeLevel: completeLevelAction, endSession: endSessionAction } = useGamePlayActions();

  const totalItems = contentItems?.length ?? 0;

  const sortedLevels = [...levels].sort((a, b) => a.order - b.order);
  const isLastLevel = sortedLevels[sortedLevels.length - 1]?.id === levelId;

  useEffect(() => {
    if (firedRef.current || !sessionId || !levelId) return;
    firedRef.current = true;

    async function complete() {
      await completeLevelAction(sessionId!, levelId!, {
        score,
        maxCombo: combo.maxCombo,
        totalItems,
      });

      if (isLastLevel) {
        await endSessionAction(sessionId!, {
          score,
          exp: 0,
          maxCombo: combo.maxCombo,
          correctCount,
          wrongCount,
          skipCount,
        });
      }

      setCompleting(false);
    }

    complete();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- action functions are stable, fire-once via firedRef guard
  }, [
    sessionId,
    levelId,
    score,
    combo.maxCombo,
    totalItems,
    isLastLevel,
    correctCount,
    wrongCount,
    skipCount,
  ]);

  return { completing, isLastLevel };
}
