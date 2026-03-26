"use client";

import { useState, useEffect, useRef } from "react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import {
  completeLevelAction,
  endSessionAction,
} from "@/features/web/play-core/actions/session.action";

interface UseGameResultParams {
  levels: { id: string; order: number }[];
}

/** Auto-fires completeLevel (and endSession if last level) on mount */
export function useGameResult({ levels }: UseGameResultParams) {
  const [completing, setCompleting] = useState(true);
  const firedRef = useRef(false);

  const sessionId = useGameStore((s) => s.sessionId);
  const gameId = useGameStore((s) => s.gameId);
  const levelId = useGameStore((s) => s.levelId);
  const score = useGameStore((s) => s.score);
  const combo = useGameStore((s) => s.combo);
  const correctCount = useGameStore((s) => s.correctCount);
  const wrongCount = useGameStore((s) => s.wrongCount);
  const skipCount = useGameStore((s) => s.skipCount);
  const contentItems = useGameStore((s) => s.contentItems);

  const totalItems = contentItems?.length ?? 0;

  const sortedLevels = [...levels].sort((a, b) => a.order - b.order);
  const isLastLevel = sortedLevels[sortedLevels.length - 1]?.id === levelId;

  useEffect(() => {
    if (firedRef.current || !sessionId || !levelId || !gameId) return;
    firedRef.current = true;

    async function complete() {
      await completeLevelAction(sessionId!, levelId!, {
        score,
        maxCombo: combo.maxCombo,
        totalItems,
      });

      if (isLastLevel) {
        await endSessionAction(sessionId!, {
          gameId: gameId!,
          score,
          exp: 0,
          maxCombo: combo.maxCombo,
          correctCount,
          wrongCount,
          skipCount,
          allLevelsCompleted: true,
        });
      }

      setCompleting(false);
    }

    complete();
  }, [
    sessionId,
    levelId,
    gameId,
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
