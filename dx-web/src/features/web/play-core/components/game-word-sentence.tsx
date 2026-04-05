"use client";

import { useRef } from "react";
import {
  EyeOff,
  Volume2,
  Check,
  Star,
  SkipForward,
  BookOpen,
  ListChecks,
  Keyboard,
} from "lucide-react";
import { useWordSentence } from "@/features/web/play-core/hooks/use-word-sentence";
import { SpellingInputRow } from "@/features/web/play-core/components/spelling-input-row";
import type { SpellingItem } from "@/features/web/play-core/types/spelling";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import {
  markAsMasteredAction,
  markAsUnknownAction,
} from "@/features/web/play-core/actions/tracking.action";
import { toast } from "sonner";

export function GameWordSentence() {
  const {
    inputValue,
    setInputValue,
    typedWords,
    hasError,
    isRevealed,
    currentWord,
    currentItem,
    progress,
    wordProgress,
    showAnswer,
    handleKeyDown,
    toggleAnswer,
    submitWord,
    skipItem,
    advanceAfterReveal,
  } = useWordSentence();

  const { competitive } = useGamePlayActions();

  const gameId = useGameStore((s) => s.gameId);
  const levelId = useGameStore((s) => s.levelId);

  const masteredIdsRef = useRef(new Set<string>());
  const unknownIdsRef = useRef(new Set<string>());

  if (!currentItem) return null;

  // Pre-compute sorted spelling items and whether any have phonetic
  const rawItemsData = currentItem.items;
  const rawItems = Array.isArray(rawItemsData)
    ? rawItemsData
    : typeof rawItemsData === "string"
      ? (() => { try { const p = JSON.parse(rawItemsData); return Array.isArray(p) ? p : []; } catch { return []; } })()
      : [];
  const sortedSpellingItems = (rawItems as SpellingItem[])
    .filter((si) => si.position >= 1)
    .sort((a, b) => a.position - b.position);
  const hasAnyPhonetic = sortedSpellingItems.some((si) => si.phonetic?.uk);

  /** Fire-and-forget: mark current content as mastered (deduped per item) */
  function handleMastered() {
    if (!gameId || !levelId) return;
    if (masteredIdsRef.current.has(currentItem!.id)) return;
    masteredIdsRef.current.add(currentItem!.id);

    markAsMasteredAction({
      contentItemId: currentItem!.id,
      gameId,
      gameLevelId: levelId,
    });
    toast.success("已掌握", { duration: 1500 });
  }

  /** Fire-and-forget: mark current content as unknown (deduped per item) */
  function handleUnknown() {
    if (!gameId || !levelId) return;
    if (unknownIdsRef.current.has(currentItem!.id)) return;
    unknownIdsRef.current.add(currentItem!.id);

    markAsUnknownAction({
      contentItemId: currentItem!.id,
      gameId,
      gameLevelId: levelId,
    });
    toast.success("已加入生词", { duration: 1500 });
  }

  return (
    <div className="flex w-full max-w-3xl flex-col gap-8 rounded-[20px] border border-border bg-card p-6 shadow-sm md:p-9">
      {/* Progress */}
      <div className="flex flex-col gap-2.5">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-1 text-xs font-medium text-muted-foreground">
            <ListChecks className="h-3.5 w-3.5" />
            当前关卡进度
          </span>
          <span className="text-xs font-medium text-muted-foreground">
            ( {progress.current} / {progress.total} )
          </span>
        </div>
        <div className="h-1.5 w-full rounded-full bg-border">
          <div
            className="h-1.5 rounded-full bg-teal-500 transition-all duration-300"
            style={{
              width: `${(progress.current / Math.max(progress.total, 1)) * 100}%`,
            }}
          />
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Prompt */}
      <div className="flex flex-col items-center gap-4">
        <p className="text-center text-xl font-extrabold tracking-tight text-foreground md:text-[28px]">
          {currentItem.translation}
        </p>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Hidden mask / Revealed text — always render content for stable height */}
      <div className="relative rounded-2xl border border-border bg-muted px-6 py-5">
        {/* Revealed content — invisible when hidden to reserve height */}
        <div
          className={`flex flex-wrap items-start justify-center gap-x-3 gap-y-2 ${
            isRevealed ? "animate-[fadeIn_0.4s_ease-in]" : "invisible"
          }`}
        >
          {sortedSpellingItems.map((si) => (
              <div key={si.position} className="flex flex-col items-center gap-0.5">
                {/* UK phonetic — invisible spacer keeps alignment when absent */}
                {hasAnyPhonetic && (
                  <span className={`text-xs ${si.phonetic?.uk ? "text-teal-600" : "invisible"}`}>
                    {si.phonetic?.uk || "\u00A0"}
                  </span>
                )}
                {/* Word */}
                <span className="text-lg font-semibold text-foreground">
                  {si.item}
                </span>
                {/* POS pill */}
                {si.pos && (
                  <span className="rounded-full bg-indigo-100 px-1.5 py-0.5 text-[10px] text-indigo-600">
                    {si.pos}
                  </span>
                )}
              </div>
            ))}
        </div>
        {/* Dots overlay — centered on top when hidden */}
        {!isRevealed && (
          <div className="absolute inset-0 flex items-center justify-center gap-2.5">
            <span className="text-base font-medium text-slate-300">
              · · · · · ·
            </span>
            <EyeOff className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-base font-medium text-slate-300">
              · · · · · ·
            </span>
          </div>
        )}
      </div>

      {/* Spelling input row */}
      <div className="flex flex-col gap-3">
        <SpellingInputRow
          typedWords={typedWords}
          inputValue={inputValue}
          hasError={hasError}
          isRevealed={isRevealed}
          currentWord={currentWord}
          showAnswer={showAnswer}
          onInputChange={setInputValue}
          onKeyDown={handleKeyDown}
        />
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <Keyboard className="h-3.5 w-3.5" />
            输入提示内容，空格或回车确认
          </span>
          <span className="text-xs font-bold text-muted-foreground">
            ( {wordProgress.current} / {wordProgress.total} )
          </span>
        </div>
      </div>

      {/* Action buttons */}
      <div className="flex flex-wrap items-center justify-center gap-3">
        {/* Pronunciation */}
        <button
          type="button"
          className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
        >
          <Volume2 className="h-4 w-4 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground">发音</span>
        </button>
        {/* Unknown word (生词) — left position */}
        <button
          type="button"
          onClick={handleUnknown}
          className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
        >
          <Star className="h-4 w-4 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground">生词</span>
        </button>
        {/* Mastered (掌握) — right position */}
        <button
          type="button"
          onClick={handleMastered}
          className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
        >
          <Check className="h-4 w-4 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground">掌握</span>
        </button>
        {/* Answer toggle — hidden in competitive modes */}
        {!competitive && (
          <button
            type="button"
            onClick={toggleAnswer}
            className={`flex items-center gap-2 rounded-xl px-5 py-3 ${
              showAnswer
                ? "border border-teal-600 bg-muted"
                : "border border-border bg-muted"
            }`}
          >
            <BookOpen
              className={`h-4 w-4 ${showAnswer ? "text-teal-600" : "text-muted-foreground"}`}
            />
            <span
              className={`text-xs font-medium ${showAnswer ? "text-teal-600" : "text-muted-foreground"}`}
            >
              答案
            </span>
          </button>
        )}
        {/* Skip — hidden in competitive modes */}
        {!competitive && (
          <button
            type="button"
            onClick={skipItem}
            className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
          >
            <SkipForward className="h-4 w-4 text-muted-foreground" />
            <span className="text-xs font-medium text-muted-foreground">跳过</span>
          </button>
        )}
        {/* Confirm */}
        <button
          type="button"
          onClick={() =>
            isRevealed ? advanceAfterReveal() : submitWord(inputValue)
          }
          className="flex items-center gap-2 rounded-xl bg-teal-600 px-9 py-3"
        >
          <Check className="h-4 w-4 text-white" />
          <span className="text-xs font-semibold text-white">确认</span>
        </button>
      </div>
    </div>
  );
}
