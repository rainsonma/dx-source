"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";
import type { SpellingItem, TypedWord } from "@/features/web/play-core/types/spelling";

const REVEAL_DELAY_MS = 1200;
const ERROR_ANIMATION_MS = 400;

export function useLsrw() {
  const currentIndex = useGameStore((s) => s.currentIndex);
  const contentItems = useGameStore((s) => s.contentItems);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelSessionId = useGameStore((s) => s.levelSessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const recordSkip = useGameStore((s) => s.recordSkip);
  const nextItem = useGameStore((s) => s.nextItem);
  const setPhase = useGameStore((s) => s.setPhase);

  const { recordAnswer: recordAnswerAction, recordSkip: recordSkipAction, markAsReview: markAsReviewAction } = useGamePlayActions();

  const [wordIndex, setWordIndex] = useState(0);
  const [typedWords, setTypedWords] = useState<TypedWord[]>([]);
  const [inputValue, setInputValue] = useState("");
  const [hasError, setHasError] = useState(false);
  const [hadWrongAttempt, setHadWrongAttempt] = useState(false);
  const [isRevealed, setIsRevealed] = useState(false);
  const [showAnswer, setShowAnswer] = useState(false);

  const errorTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const revealTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const hadWrongAttemptRef = useRef(false);
  const isProcessingRef = useRef(false);
  const itemStartTimeRef = useRef<number>(Date.now());

  const currentItem = contentItems?.[currentIndex] ?? null;
  const items: SpellingItem[] = Array.isArray(currentItem?.items) ? currentItem.items : [];
  const currentWord = items[wordIndex] ?? null;
  const totalItems = contentItems?.length ?? 0;

  const progress = {
    current: currentIndex + 1,
    total: totalItems,
  };

  const wordProgress = {
    current: typedWords.length,
    total: items.length,
  };

  // Auto-skip leading answer:false items starting from a given index
  const skipNonAnswers = useCallback(
    (fromIndex: number, currentItems: SpellingItem[]) => {
      let idx = fromIndex;
      const autoWords: TypedWord[] = [];
      while (idx < currentItems.length && !currentItems[idx].answer) {
        autoWords.push({ text: currentItems[idx].item, isAnswer: false });
        idx++;
      }
      return { nextIndex: idx, autoWords };
    },
    []
  );

  // Initialize/reset when currentIndex changes
  useEffect(() => {
    const item = contentItems?.[currentIndex] ?? null;
    const spellingItems: SpellingItem[] =
      Array.isArray(item?.items) ? item.items : [];

    if (revealTimerRef.current) {
      clearTimeout(revealTimerRef.current);
      revealTimerRef.current = null;
    }

    setInputValue("");
    setHasError(false);
    setHadWrongAttempt(false);
    hadWrongAttemptRef.current = false;
    setIsRevealed(false);

    isProcessingRef.current = false;
    itemStartTimeRef.current = Date.now();

    const { nextIndex, autoWords } = skipNonAnswers(0, spellingItems);
    setWordIndex(nextIndex);
    setTypedWords(autoWords);
  }, [currentIndex, contentItems, skipNonAnswers]);

  // Cleanup timers
  useEffect(() => {
    return () => {
      if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
      if (revealTimerRef.current) clearTimeout(revealTimerRef.current);
    };
  }, []);

  const submitWord = useCallback(
    (input: string) => {
      const trimmed = input.trim();
      if (isRevealed) return;

      // Empty input — trigger error shake, keep red until user types
      if (!trimmed) {
        if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
        setHasError(true);
        return;
      }

      if (!currentWord) return;

      const isCorrect =
        trimmed.toLowerCase() === currentWord.item.toLowerCase();

      if (!isCorrect) {
        setHadWrongAttempt(true);
        hadWrongAttemptRef.current = true;
        setHasError(true);
        return;
      }

      // Correct answer
      const newTypedWords: TypedWord[] = [
        ...typedWords,
        { text: currentWord.item, isAnswer: true },
      ];

      // Auto-skip subsequent answer:false items
      const { nextIndex, autoWords } = skipNonAnswers(
        wordIndex + 1,
        items
      );
      const allTypedWords = [...newTypedWords, ...autoWords];

      setTypedWords(allTypedWords);
      setInputValue("");
      setWordIndex(nextIndex);

      // Check if all words in this item are done
      if (nextIndex >= items.length) {
        const isItemCorrect = !hadWrongAttemptRef.current;
        setIsRevealed(true);

        // Capture score before update to derive per-record scores
        const prevScore = useGameStore.getState().score;
        recordResult(isItemCorrect);

        // Derive per-record scores from state delta
        const latestState = useGameStore.getState();
        const pointsEarned = latestState.score - prevScore;
        const baseScore = isItemCorrect ? SCORING.CORRECT_ANSWER : 0;
        const comboScore = pointsEarned - baseScore;

        // Record to server (fire-and-forget)
        if (sessionId && levelSessionId && levelId && currentItem) {
          const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
          const duration = Math.round(
            (Date.now() - itemStartTimeRef.current) / 1000
          );
          recordAnswerAction({
            gameSessionTotalId: sessionId,
            gameSessionLevelId: levelSessionId!,
            gameLevelId: levelId,
            contentItemId: currentItem.id,
            isCorrect: isItemCorrect,
            userAnswer: allTypedWords
              .filter((w) => w.isAnswer)
              .map((w) => w.text)
              .join(" "),
            sourceAnswer: currentItem.content as string,
            baseScore,
            comboScore,
            score: latestState.score,
            maxCombo: latestState.combo.maxCombo,
            playTime: getElapsedSeconds(),
            nextContentItemId: nextItemId,
            duration,
          });
        }

        // Save incorrect item for review (fire-and-forget)
        if (!isItemCorrect && currentItem && gameId && levelId) {
          markAsReviewAction({
            contentItemId: currentItem.id,
            gameId,
            gameLevelId: levelId,
          });
        }

        // Wait for user to press Enter/Space to advance (handled in handleKeyDown)
      }
    },
    [
      currentWord,
      isRevealed,
      typedWords,
      wordIndex,
      items,
      recordResult,
      sessionId,
      levelSessionId,
      levelId,
      gameId,
      currentItem,
      contentItems,
      currentIndex,
      totalItems,
      setPhase,
      nextItem,
      skipNonAnswers,
    ]
  );

  /** Advance to the next item after reveal */
  const advanceAfterReveal = useCallback(() => {
    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    if (currentIndex + 1 >= totalItems) {
      setPhase("result");
    } else {
      nextItem();
    }
  }, [currentIndex, totalItems, setPhase, nextItem]);

  /** Skip current item — record as neutral skip, no GameRecord or UserReview */
  const skipItem = useCallback(() => {
    if (!currentItem) return;

    // If already revealed, just advance to next item
    if (isRevealed) {
      advanceAfterReveal();
      return;
    }

    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    // Client: increment skipCount, reset combo
    recordSkip();

    // Server: increment skipCount on GameSessionLevel + GameSession (fire-and-forget)
    if (sessionId && levelId) {
      const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
      recordSkipAction({
        gameSessionTotalId: sessionId,
        gameLevelId: levelId,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
      });
    }

    // Show reveal — user must press again to advance
    setIsRevealed(true);
    isProcessingRef.current = false;
  }, [
    currentItem,
    isRevealed,
    recordSkip,
    sessionId,
    levelId,
    contentItems,
    currentIndex,
    advanceAfterReveal,
  ]);

  /** Toggle answer hint visibility */
  const toggleAnswer = useCallback(() => {
    setShowAnswer((prev) => !prev);
  }, []);

  // Listen for Enter/Space/Tab on document when revealed (input is unmounted)
  useEffect(() => {
    if (!isRevealed) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Tab" || e.key === " ") e.preventDefault();
      if (e.repeat) return;
      if (e.key === "Enter" || e.key === " ") {
        advanceAfterReveal();
      } else if (e.key === "Tab") {
        advanceAfterReveal();
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [isRevealed, advanceAfterReveal]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Tab" || e.key === " ") e.preventDefault();
      if (e.repeat) return;
      if (e.key === "Enter" || e.key === " ") {
        submitWord(inputValue);
      } else if (e.key === "Tab") {
        skipItem();
      }
    },
    [submitWord, inputValue, skipItem]
  );

  /** Clear error state when user starts typing */
  const handleInputChange = useCallback(
    (value: string) => {
      setInputValue(value);
      if (value && hasError) {
        if (errorTimerRef.current) {
          clearTimeout(errorTimerRef.current);
          errorTimerRef.current = null;
        }
        setHasError(false);
      }
    },
    [hasError]
  );

  return {
    // State
    inputValue,
    setInputValue: handleInputChange,
    typedWords,
    hasError,
    isRevealed,
    currentWord,
    currentItem,
    progress,
    wordProgress,
    showAnswer,
    // Actions
    submitWord,
    handleKeyDown,
    toggleAnswer,
    skipItem,
    advanceAfterReveal,
  };
}
