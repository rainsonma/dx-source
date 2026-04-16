"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore, type ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const BATCH_SIZE = 5;

function shuffleArray<T>(arr: T[], seed: number): T[] {
  const result = [...arr];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

export function useVocabMatch() {
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

  const [selectedWordIndex, setSelectedWordIndex] = useState<number | null>(null);
  const [matchedIndices, setMatchedIndices] = useState<Set<number>>(new Set());
  const [wrongPair, setWrongPair] = useState<{ word: number; def: number } | null>(null);

  const wrongTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const advanceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());
  const advanceCalledRef = useRef(false);

  const totalItems = contentItems?.length ?? 0;

  const batchStart = currentIndex;
  const batchEnd = Math.min(batchStart + BATCH_SIZE, totalItems);
  const batchItems: ContentItem[] = useMemo(
    () => contentItems?.slice(batchStart, batchEnd) ?? [],
    [contentItems, batchStart, batchEnd]
  );

  const shuffledDefs = useMemo(() => {
    const defs = batchItems.map((item, i) => ({
      batchIndex: i,
      translation: (item.translation as string) ?? "",
    }));
    return shuffleArray(defs, batchStart + 1);
  }, [batchItems, batchStart]);

  const totalMatched = currentIndex + matchedIndices.size;

  const progress = {
    current: totalMatched,
    total: totalItems,
  };

  useEffect(() => {
    setSelectedWordIndex(null);
    setMatchedIndices(new Set());
    setWrongPair(null);
    itemStartTimeRef.current = Date.now();
    advanceCalledRef.current = false;
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
    if (advanceCalledRef.current) return;
    advanceCalledRef.current = true;
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);

  const selectWord = useCallback(
    (batchIndex: number) => {
      if (matchedIndices.has(batchIndex)) return;
      if (wrongPair) return;
      setSelectedWordIndex(batchIndex);
    },
    [matchedIndices, wrongPair]
  );

  const selectDef = useCallback(
    (defBatchIndex: number) => {
      if (selectedWordIndex === null) return;
      if (matchedIndices.has(defBatchIndex)) return;
      if (wrongPair) return;

      const item = batchItems[selectedWordIndex];
      if (!item) return;

      const isCorrect = defBatchIndex === selectedWordIndex;

      if (isCorrect) {
        const newMatched = new Set(matchedIndices);
        newMatched.add(selectedWordIndex);
        setMatchedIndices(newMatched);
        setSelectedWordIndex(null);

        fireServerRecord(item, true, selectedWordIndex);

        if (newMatched.size === batchItems.length) {
          advanceTimerRef.current = setTimeout(advanceBatch, 600);
        }
      } else {
        setWrongPair({ word: selectedWordIndex, def: defBatchIndex });
        fireServerRecord(item, false, selectedWordIndex);

        wrongTimerRef.current = setTimeout(() => {
          setWrongPair(null);
          setSelectedWordIndex(null);
        }, 500);
      }
    },
    [selectedWordIndex, matchedIndices, batchItems, wrongPair, fireServerRecord, advanceBatch]
  );

  const isBatchComplete = matchedIndices.size === batchItems.length && batchItems.length > 0;
  const isLastBatch = batchEnd >= totalItems;

  const nextBatch = useCallback(() => {
    if (advanceTimerRef.current) {
      clearTimeout(advanceTimerRef.current);
      advanceTimerRef.current = null;
    }
    advanceBatch();
  }, [advanceBatch]);

  return {
    batchItems,
    shuffledDefs,
    selectedWordIndex,
    matchedIndices,
    wrongPair,
    progress,
    combo,
    selectWord,
    selectDef,
    isBatchComplete,
    isLastBatch,
    nextBatch,
  };
}
