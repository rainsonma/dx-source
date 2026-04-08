"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore, type ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const BATCH_SIZE = 8;
const COLUMNS = 4;

export type Tile = {
  id: string;
  type: "en" | "zh";
  text: string;
  itemIndex: number;
};

function shuffleTiles(tiles: Tile[], seed: number): Tile[] {
  const result = [...tiles];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

export function useVocabElimination() {
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const setPhase = useGameStore((s) => s.setPhase);
  const combo = useGameStore((s) => s.combo);

  const {
    recordAnswer: recordAnswerAction,
    markAsReview: markAsReviewAction,
  } = useGamePlayActions();

  const [selectedTileId, setSelectedTileId] = useState<string | null>(null);
  const [eliminatedIndices, setEliminatedIndices] = useState<Set<number>>(new Set());
  const [wrongPair, setWrongPair] = useState<{ t1: string; t2: string } | null>(null);

  const wrongTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const advanceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());

  const totalItems = contentItems?.length ?? 0;

  const batchStart = currentIndex;
  const batchEnd = Math.min(batchStart + BATCH_SIZE, totalItems);
  const batchItems: ContentItem[] = useMemo(
    () => contentItems?.slice(batchStart, batchEnd) ?? [],
    [contentItems, batchStart, batchEnd]
  );

  const tiles = useMemo(() => {
    const raw: Tile[] = [];
    batchItems.forEach((item, i) => {
      raw.push({ id: `en-${i}`, type: "en", text: item.content, itemIndex: i });
      raw.push({ id: `zh-${i}`, type: "zh", text: (item.translation as string) ?? "", itemIndex: i });
    });
    return shuffleTiles(raw, batchStart + 1);
  }, [batchItems, batchStart]);

  const gridRows = useMemo(() => {
    const rows: Tile[][] = [];
    for (let i = 0; i < tiles.length; i += COLUMNS) {
      rows.push(tiles.slice(i, i + COLUMNS));
    }
    return rows;
  }, [tiles]);

  const totalEliminated = currentIndex + eliminatedIndices.size;
  const progress = {
    current: totalEliminated,
    total: totalItems,
  };

  useEffect(() => {
    setSelectedTileId(null);
    setEliminatedIndices(new Set());
    setWrongPair(null);
    itemStartTimeRef.current = Date.now();
  }, [batchStart]);

  useEffect(() => {
    return () => {
      if (wrongTimerRef.current) clearTimeout(wrongTimerRef.current);
      if (advanceTimerRef.current) clearTimeout(advanceTimerRef.current);
    };
  }, []);

  const fireServerRecord = useCallback(
    (item: ContentItem, isCorrect: boolean, batchIdx: number) => {
      if (!sessionId || !levelId) return;

      const prevScore = useGameStore.getState().score;
      recordResult(isCorrect);

      const latestState = useGameStore.getState();
      const pointsEarned = latestState.score - prevScore;
      const baseScore = isCorrect ? SCORING.CORRECT_ANSWER : 0;
      const comboScore = pointsEarned - baseScore;

      const nextItemId =
        contentItems?.[batchStart + batchIdx + 1]?.id ??
        contentItems?.[batchEnd]?.id ??
        null;

      const duration = Math.round(
        (Date.now() - itemStartTimeRef.current) / 1000
      );

      recordAnswerAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        contentItemId: item.id,
        isCorrect,
        userAnswer: isCorrect ? item.content : "",
        sourceAnswer: item.content,
        baseScore,
        comboScore,
        score: latestState.score,
        maxCombo: latestState.combo.maxCombo,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
        duration,
      });

      if (!isCorrect && gameId && !reviewedIdsRef.current.has(item.id)) {
        reviewedIdsRef.current.add(item.id);
        markAsReviewAction({
          contentItemId: item.id,
          gameId,
          gameLevelId: levelId,
        });
      }

      itemStartTimeRef.current = Date.now();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [sessionId, levelId, gameId, contentItems, batchStart, batchEnd, recordResult]
  );

  const advanceBatch = useCallback(() => {
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);

  const selectTile = useCallback(
    (tileId: string) => {
      if (wrongPair) return;

      const tile = tiles.find((t) => t.id === tileId);
      if (!tile) return;
      if (eliminatedIndices.has(tile.itemIndex)) return;

      if (selectedTileId === tileId) {
        setSelectedTileId(null);
        return;
      }

      if (selectedTileId === null) {
        setSelectedTileId(tileId);
        return;
      }

      const firstTile = tiles.find((t) => t.id === selectedTileId);
      if (!firstTile) {
        setSelectedTileId(tileId);
        return;
      }

      const item = batchItems[firstTile.itemIndex];
      if (!item) return;

      const isMatch =
        firstTile.itemIndex === tile.itemIndex && firstTile.type !== tile.type;

      if (isMatch) {
        const newEliminated = new Set(eliminatedIndices);
        newEliminated.add(firstTile.itemIndex);
        setEliminatedIndices(newEliminated);
        setSelectedTileId(null);

        fireServerRecord(item, true, firstTile.itemIndex);

        if (newEliminated.size === batchItems.length) {
          advanceTimerRef.current = setTimeout(advanceBatch, 600);
        }
      } else {
        setWrongPair({ t1: selectedTileId, t2: tileId });
        fireServerRecord(item, false, firstTile.itemIndex);

        wrongTimerRef.current = setTimeout(() => {
          setWrongPair(null);
          setSelectedTileId(null);
        }, 500);
      }
    },
    [selectedTileId, eliminatedIndices, tiles, batchItems, wrongPair, fireServerRecord, advanceBatch]
  );

  return {
    gridRows,
    tiles,
    selectedTileId,
    eliminatedIndices,
    wrongPair,
    progress,
    combo,
    selectTile,
  };
}
