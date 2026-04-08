"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const SHIELD_COUNT = 5;
const MIN_KEYBOARD_SIZE = 6;

function seededShuffle<T>(arr: T[], seed: number): T[] {
  const result = [...arr];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

function buildKeyboard(word: string, seed: number): string[] {
  const letters = word.toUpperCase().split("");
  const distractors = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  let s = seed;
  while (letters.length < MIN_KEYBOARD_SIZE) {
    s = (s * 16807 + 0) % 2147483647;
    const ch = distractors[s % distractors.length];
    if (!letters.includes(ch)) {
      letters.push(ch);
    }
  }
  return seededShuffle(letters, seed + 7);
}

export function useVocabBattle() {
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const recordSkipStore = useGameStore((s) => s.recordSkip);
  const nextItem = useGameStore((s) => s.nextItem);
  const setPhase = useGameStore((s) => s.setPhase);
  const combo = useGameStore((s) => s.combo);

  const {
    recordAnswer: recordAnswerAction,
    recordSkip: recordSkipAction,
    markAsReview: markAsReviewAction,
    competitive,
  } = useGamePlayActions();

  const [filledLetters, setFilledLetters] = useState<string[]>([]);
  const [usedKeyIndices, setUsedKeyIndices] = useState<Set<number>>(new Set());
  const [hasError, setHasError] = useState(false);
  const [isRevealed, setIsRevealed] = useState(false);
  const [playerShields, setPlayerShields] = useState<boolean[]>([]);
  const [opponentShields, setOpponentShields] = useState<boolean[]>([]);
  const [opponentFilledCount, setOpponentFilledCount] = useState(0);

  const hadWrongAttemptRef = useRef(false);
  const errorTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isProcessingRef = useRef(false);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());

  const totalItems = contentItems?.length ?? 0;
  const currentItem = contentItems?.[currentIndex] ?? null;
  const targetWord = (currentItem?.content ?? "").toUpperCase();
  const translation = (currentItem?.translation as string) ?? "";

  const phonetic = useMemo(() => {
    const raw = currentItem?.items;
    const items = Array.isArray(raw)
      ? raw
      : typeof raw === "string"
        ? (() => { try { return JSON.parse(raw); } catch { return []; } })()
        : [];
    return items[0]?.phonetic ?? null;
  }, [currentItem]);

  const keyboardLetters = useMemo(
    () => (targetWord ? buildKeyboard(targetWord, currentIndex + 1) : []),
    [targetWord, currentIndex]
  );

  const letterSlots = useMemo(() => {
    return targetWord.split("").map((letter, i) => ({
      letter,
      filled: i < filledLetters.length,
      filledLetter: filledLetters[i] ?? null,
    }));
  }, [targetWord, filledLetters]);

  const opponentSlots = useMemo(() => {
    return targetWord.split("").map((letter, i) => ({
      letter,
      filled: i < opponentFilledCount,
    }));
  }, [targetWord, opponentFilledCount]);

  const progress = {
    current: currentIndex + 1,
    total: totalItems,
  };

  useEffect(() => {
    setFilledLetters([]);
    setUsedKeyIndices(new Set());
    setHasError(false);
    setIsRevealed(false);
    setPlayerShields(Array(SHIELD_COUNT).fill(true));
    setOpponentShields(Array(SHIELD_COUNT).fill(true));
    setOpponentFilledCount(0);
    hadWrongAttemptRef.current = false;
    isProcessingRef.current = false;
    itemStartTimeRef.current = Date.now();
  }, [currentIndex]);

  useEffect(() => {
    return () => {
      if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
    };
  }, []);

  const fireServerRecord = useCallback(
    (isCorrect: boolean) => {
      if (!sessionId || !levelId || !currentItem) return;

      const prevScore = useGameStore.getState().score;
      recordResult(isCorrect);

      const latestState = useGameStore.getState();
      const pointsEarned = latestState.score - prevScore;
      const baseScore = isCorrect ? SCORING.CORRECT_ANSWER : 0;
      const comboScore = pointsEarned - baseScore;

      const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
      const duration = Math.round(
        (Date.now() - itemStartTimeRef.current) / 1000
      );

      recordAnswerAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        contentItemId: currentItem.id,
        isCorrect,
        userAnswer: filledLetters.join(""),
        sourceAnswer: currentItem.content,
        baseScore,
        comboScore,
        score: latestState.score,
        maxCombo: latestState.combo.maxCombo,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
        duration,
      });

      if (!isCorrect && gameId && !reviewedIdsRef.current.has(currentItem.id)) {
        reviewedIdsRef.current.add(currentItem.id);
        markAsReviewAction({
          contentItemId: currentItem.id,
          gameId,
          gameLevelId: levelId,
        });
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [sessionId, levelId, gameId, currentItem, contentItems, currentIndex, filledLetters, recordResult]
  );

  const pressLetter = useCallback(
    (keyIndex: number) => {
      if (isRevealed) return;
      if (usedKeyIndices.has(keyIndex)) return;

      const letter = keyboardLetters[keyIndex];
      if (!letter) return;

      const nextPos = filledLetters.length;
      const expectedLetter = targetWord[nextPos];

      if (letter === expectedLetter) {
        const newFilled = [...filledLetters, letter];
        setFilledLetters(newFilled);
        setUsedKeyIndices(new Set([...usedKeyIndices, keyIndex]));

        setOpponentShields((prev) => {
          const idx = prev.lastIndexOf(true);
          if (idx === -1) return prev;
          const next = [...prev];
          next[idx] = false;
          return next;
        });

        if (hasError) setHasError(false);

        if (newFilled.length === targetWord.length) {
          const isItemCorrect = !hadWrongAttemptRef.current;
          setIsRevealed(true);
          fireServerRecord(isItemCorrect);
        }
      } else {
        hadWrongAttemptRef.current = true;
        setHasError(true);
        if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
        errorTimerRef.current = setTimeout(() => setHasError(false), 400);

        setPlayerShields((prev) => {
          const idx = prev.lastIndexOf(true);
          if (idx === -1) return prev;
          const next = [...prev];
          next[idx] = false;
          return next;
        });
      }
    },
    [isRevealed, usedKeyIndices, keyboardLetters, filledLetters, targetWord, hasError, fireServerRecord]
  );

  const advanceAfterReveal = useCallback(() => {
    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    if (currentIndex + 1 >= totalItems) {
      setPhase("result");
    } else {
      nextItem();
    }
  }, [currentIndex, totalItems, setPhase, nextItem]);

  const skipItem = useCallback(() => {
    if (!currentItem) return;

    if (isRevealed) {
      advanceAfterReveal();
      return;
    }

    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    recordSkipStore();

    if (sessionId && levelId) {
      const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
      recordSkipAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
      });
    }

    setIsRevealed(true);
    isProcessingRef.current = false;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentItem, isRevealed, recordSkipStore, sessionId, levelId, contentItems, currentIndex, advanceAfterReveal]);

  useEffect(() => {
    if (!isRevealed) return;
    const handler = (e: KeyboardEvent) => {
      if (e.repeat) return;
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        advanceAfterReveal();
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [isRevealed, advanceAfterReveal]);

  return {
    targetWord,
    translation,
    phonetic,
    letterSlots,
    keyboardLetters,
    usedKeyIndices,
    filledLetters,
    hasError,
    isRevealed,
    playerShields,
    opponentShields,
    opponentSlots,
    opponentFilledCount,
    competitive: competitive ?? false,
    progress,
    combo,
    pressLetter,
    advanceAfterReveal,
    skipItem,
  };
}
