"use client";

import { useRef, useCallback, useEffect, useState } from "react";
import type {
  SpellingItem,
  TypedWord,
} from "@/features/web/play-core/types/spelling";
import { useKeySound } from "@/features/web/play-core/hooks/use-key-sound";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";

interface SpellingInputRowProps {
  typedWords: TypedWord[];
  inputValue: string;
  hasError: boolean;
  isRevealed: boolean;
  currentWord: SpellingItem | null;
  showAnswer: boolean;
  onInputChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}

export function SpellingInputRow({
  typedWords,
  inputValue,
  hasError,
  isRevealed,
  currentWord,
  showAnswer,
  onInputChange,
  onKeyDown,
}: SpellingInputRowProps) {
  const { playKeySound } = useKeySound();
  const overlay = useGameStore((s) => s.overlay);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const isDragging = useRef(false);
  const dragStartX = useRef(0);
  const scrollStartX = useRef(0);
  const [isFocused, setIsFocused] = useState(false);

  // Auto-scroll to rightmost on new word
  useEffect(() => {
    const el = containerRef.current;
    if (el) {
      el.scrollLeft = el.scrollWidth;
    }
  }, [typedWords.length]);

  // Auto-focus input (skip when an overlay is open)
  useEffect(() => {
    if (!isRevealed && !overlay) {
      inputRef.current?.focus();
    }
  }, [typedWords.length, isRevealed, overlay]);

  // Drag handlers — skip if target is the input
  const handleDragStart = useCallback(
    (clientX: number, target: EventTarget) => {
      if (inputRef.current && inputRef.current === target) return;
      isDragging.current = true;
      dragStartX.current = clientX;
      scrollStartX.current = containerRef.current?.scrollLeft ?? 0;
    },
    []
  );

  const handleDragMove = useCallback((clientX: number) => {
    if (!isDragging.current || !containerRef.current) return;
    const dx = dragStartX.current - clientX;
    containerRef.current.scrollLeft = scrollStartX.current + dx;
  }, []);

  const handleDragEnd = useCallback(() => {
    isDragging.current = false;
  }, []);

  const onMouseDown = useCallback(
    (e: React.MouseEvent) => handleDragStart(e.clientX, e.target),
    [handleDragStart]
  );
  const onMouseMove = useCallback(
    (e: React.MouseEvent) => handleDragMove(e.clientX),
    [handleDragMove]
  );
  const onTouchStart = useCallback(
    (e: React.TouchEvent) =>
      handleDragStart(e.touches[0].clientX, e.target),
    [handleDragStart]
  );
  const onTouchMove = useCallback(
    (e: React.TouchEvent) => handleDragMove(e.touches[0].clientX),
    [handleDragMove]
  );

  // Count how many leading characters match the target word
  const correctPrefixLen = currentWord
    ? (() => {
        const input = inputValue.toLowerCase();
        const target = currentWord.item.toLowerCase();
        let count = 0;
        for (let i = 0; i < input.length && i < target.length; i++) {
          if (input[i] === target[i]) count++;
          else break;
        }
        return count;
      })()
    : 0;

  const answerLen = showAnswer && currentWord ? currentWord.item.length + 1 : 0;
  const inputWidthCh = Math.max(inputValue.length + 1, answerLen, 3);
  const isCorrectPrefix = correctPrefixLen > 0 && correctPrefixLen === inputValue.length;
  const isWrongInput = inputValue.length > 0 && correctPrefixLen < inputValue.length;
  const tealPercent =
    inputValue.length > 0
      ? (correctPrefixLen / inputValue.length) * 100
      : 0;

  /** Keys that should not trigger the typewriter sound */
  const SILENT_KEYS = new Set([
    "Escape",
    "Meta",
    "Control",
    "Alt",
    "Shift",
  ]);

  /** Play key sound on printable keystrokes, then delegate to parent handler */
  const handleKeyDownWithSound = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (!e.repeat && !SILENT_KEYS.has(e.key)) {
        playKeySound();
      }
      onKeyDown(e);
    },
    [playKeySound, onKeyDown]
  );

  return (
    <div
      ref={containerRef}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={handleDragEnd}
      onMouseLeave={handleDragEnd}
      onTouchStart={onTouchStart}
      onTouchMove={onTouchMove}
      onTouchEnd={handleDragEnd}
      className="flex items-center gap-2 overflow-x-hidden rounded-[14px] border border-border bg-muted px-4 py-3.5 md:gap-3 md:px-6"
      style={{
        maskImage:
          "linear-gradient(to right, transparent, black 40px, black)",
        WebkitMaskImage:
          "linear-gradient(to right, transparent, black 40px, black)",
      }}
    >
      <div className="min-w-0 flex-1" />

      {typedWords.map((word, i) => (
        <span
          key={`${i}-${word.text}`}
          className={`shrink-0 text-base ${
            word.isAnswer
              ? "font-medium text-foreground"
              : "font-medium text-muted-foreground"
          }`}
        >
          {word.text}
        </span>
      ))}

      <div
        className={`relative shrink-0 ${isRevealed ? "invisible" : ""} ${hasError ? "animate-[shake_0.4s_ease-in-out]" : ""}`}
        style={{ width: `${inputWidthCh}ch` }}
      >
        {/* Ghost text — answer hint */}
        {showAnswer && currentWord && (
          <span
            className="pointer-events-none absolute inset-0 flex items-center justify-center text-base font-bold text-slate-300"
            aria-hidden="true"
          >
            {currentWord.item}
          </span>
        )}
        <input
          ref={inputRef}
          type="text"
          value={inputValue}
          onChange={(e) => onInputChange(e.target.value)}
          onKeyDown={handleKeyDownWithSound}
          onFocus={() => setIsFocused(true)}
          onBlur={() => {
            if (!isRevealed && !overlay) {
              requestAnimationFrame(() => inputRef.current?.focus());
            } else {
              setIsFocused(false);
            }
          }}
          aria-label="输入单词"
          autoComplete="off"
          autoCapitalize="off"
          spellCheck={false}
          className={`relative z-10 w-full px-1 text-center text-base font-bold outline-none ${
            showAnswer ? "bg-transparent" : "bg-border"
          } ${
            hasError || isWrongInput
              ? "text-red-600"
              : isCorrectPrefix
                ? "text-teal-600"
                : "text-foreground"
          }`}
        />
        <div className="flex h-[3px] overflow-hidden rounded-full">
          {isFocused && !(hasError || isWrongInput) && tealPercent > 0 && (
            <div
              className="bg-teal-600 transition-all duration-150"
              style={{ width: `${tealPercent}%` }}
            />
          )}
          <div
            className={`flex-1 transition-colors ${
              hasError || isWrongInput
                ? "bg-red-500"
                : isFocused
                  ? "bg-slate-900"
                  : "bg-slate-400"
            }`}
          />
        </div>
      </div>
    </div>
  );
}
