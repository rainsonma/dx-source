"use client";

import { useRef } from "react";
import { Zap, SkipForward, Check, Volume2, Star, Shield } from "lucide-react";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { useVocabBattle } from "@/features/web/play-core/hooks/use-vocab-battle";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import {
  markAsMasteredAction,
  markAsUnknownAction,
} from "@/features/web/play-core/actions/tracking.action";
import { toast } from "sonner";

export function GameVocabBattle() {
  const {
    targetWord,
    translation,
    letterSlots,
    keyboardLetters,
    usedKeyIndices,
    isRevealed,
    opponentSlots,
    competitive,
    progress,
    combo,
    pressLetter,
    advanceAfterReveal,
    skipItem,
  } = useVocabBattle();

  const gameId = useGameStore((s) => s.gameId);
  const levelId = useGameStore((s) => s.levelId);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const contentItems = useGameStore((s) => s.contentItems);
  const currentItem = contentItems?.[currentIndex] ?? null;

  const masteredIdsRef = useRef(new Set<string>());
  const unknownIdsRef = useRef(new Set<string>());

  if (!targetWord) return null;

  function handleMastered() {
    if (!gameId || !levelId || !currentItem) return;
    if (masteredIdsRef.current.has(currentItem.id)) return;
    masteredIdsRef.current.add(currentItem.id);
    markAsMasteredAction({ contentItemId: currentItem.id, gameId, gameLevelId: levelId });
    toast.success("已掌握", { duration: 1500 });
  }

  function handleUnknown() {
    if (!gameId || !levelId || !currentItem) return;
    if (unknownIdsRef.current.has(currentItem.id)) return;
    unknownIdsRef.current.add(currentItem.id);
    markAsUnknownAction({ contentItemId: currentItem.id, gameId, gameLevelId: levelId });
    toast.success("已加入生词", { duration: 1500 });
  }

  return (
    <div className="flex w-full max-w-[760px] flex-col rounded-[20px] border border-border bg-card shadow-sm">
      {/* Opponent zone — each letter is covered by a shield until revealed */}
      <div
        className={`flex flex-col items-center gap-3 px-6 py-7 md:px-8 ${
          !competitive ? "pointer-events-none opacity-40" : ""
        }`}
      >
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🤖 对手</span>
        </div>
        <div className="flex items-center justify-center gap-2.5">
          {opponentSlots.map((slot, i) =>
            slot.letter === " " ? (
              <div key={i} className="flex h-10 w-5 items-center justify-center">
                <span className="text-sm text-slate-300">/</span>
              </div>
            ) : (
              <div
                key={i}
                className={`flex h-10 w-10 items-center justify-center rounded-lg border transition-all ${
                  slot.revealed
                    ? "border-red-300 bg-red-50"
                    : "border-red-200 bg-red-100"
                }`}
              >
                {slot.revealed ? (
                  <span className="text-sm font-semibold text-red-600">
                    {slot.letter}
                  </span>
                ) : (
                  <Shield className="h-4 w-4 text-red-400" />
                )}
              </div>
            )
          )}
        </div>
      </div>

      {/* Translation zone */}
      <div className="flex flex-col items-center gap-2.5 bg-gradient-to-b from-red-50/0 via-red-50 to-red-50/0 px-6 py-4 md:px-8">
        <p className="text-center text-2xl font-extrabold tracking-wider text-foreground md:text-[32px]">
          {translation}
        </p>
        <div className="h-0.5 w-full rounded-full bg-gradient-to-r from-red-500/0 via-red-500/30 via-30% via-teal-500/30 via-70% to-teal-500/0" />
      </div>

      {/* Player zone — each letter is covered by a shield until correctly typed */}
      <div className="flex flex-col items-center gap-3 px-6 py-5 md:px-8">
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🎯 我</span>
        </div>
        <div className="flex items-center justify-center gap-2.5">
          {letterSlots.map((slot, i) =>
            slot.letter === " " ? (
              <div key={i} className="flex h-10 w-5 items-center justify-center">
                <span className="text-sm text-slate-300">/</span>
              </div>
            ) : (
              <div
                key={i}
                className={`flex h-10 w-10 items-center justify-center rounded-lg border transition-all ${
                  slot.filled
                    ? "border-teal-300 bg-teal-50"
                    : "border-teal-200 bg-teal-100"
                }`}
              >
                {slot.filled ? (
                  <span className="text-sm font-semibold text-teal-600">
                    {slot.filledLetter}
                  </span>
                ) : (
                  <Shield className="h-4 w-4 text-teal-400" />
                )}
              </div>
            )
          )}
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Combo row */}
      {combo.streak >= 3 && (
        <div className="flex items-center justify-center gap-3 px-6 py-2 md:px-8">
          <span className="text-[13px] font-medium text-muted-foreground">连击</span>
          <div className="flex items-center gap-1.5 rounded-lg bg-red-500 px-3 py-1">
            <Zap className="h-3 w-3 text-white" />
            <span className="text-xs font-bold text-white">
              &times;{combo.streak}
            </span>
          </div>
        </div>
      )}

      {/* Keyboard + actions */}
      <div className="flex flex-col items-center gap-8 px-6 pb-6 pt-3 md:px-8">
        <span className="text-xs font-medium text-muted-foreground">
          {competitive ? "点击字母发射炮弹击碎对手护盾" : "拼写单词"}
        </span>

        {/* Letter keyboard + space bar */}
        {!isRevealed && (
          <div className="flex flex-wrap items-center justify-center gap-2.5">
            {keyboardLetters.map((letter, i) => {
              const isUsed = usedKeyIndices.has(i);
              return (
                <button
                  key={i}
                  type="button"
                  disabled={isUsed}
                  onClick={() => pressLetter(i)}
                  className={`flex h-12 w-12 items-center justify-center rounded-xl shadow-md transition-opacity md:h-14 md:w-14 ${
                    isUsed
                      ? "bg-slate-400 opacity-40"
                      : "bg-slate-800 hover:bg-slate-700"
                  }`}
                >
                  <span className={`font-bold text-white ${letter === " " ? "text-xs" : "text-lg"}`}>
                    {letter === " " ? "空格" : letter}
                  </span>
                </button>
              );
            })}
          </div>
        )}

        {/* Revealed: show full word + advance */}
        {isRevealed && (
          <div className="flex flex-col items-center gap-3">
            <span className="text-lg font-bold text-teal-600">{targetWord}</span>
            <button
              type="button"
              onClick={advanceAfterReveal}
              className="flex items-center gap-2 rounded-xl bg-teal-600 px-9 py-3"
            >
              <Check className="h-4 w-4 text-white" />
              <span className="text-xs font-semibold text-white">
                {progress.current >= progress.total ? "查看结果" : "下一题"}
              </span>
            </button>
          </div>
        )}

        {/* Action buttons — same as word-sentence */}
        <div className="flex flex-wrap items-center justify-center gap-3">
          <button
            type="button"
            className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
          >
            <Volume2 className="h-4 w-4 text-muted-foreground" />
            <span className="text-xs font-medium text-muted-foreground">发音</span>
          </button>
          <button
            type="button"
            onClick={handleUnknown}
            className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
          >
            <Star className="h-4 w-4 text-muted-foreground" />
            <span className="text-xs font-medium text-muted-foreground">生词</span>
          </button>
          <button
            type="button"
            onClick={handleMastered}
            className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
          >
            <Check className="h-4 w-4 text-muted-foreground" />
            <span className="text-xs font-medium text-muted-foreground">掌握</span>
          </button>
          {competitive ? (
            <HoverCard openDelay={200}>
              <HoverCardTrigger asChild>
                <button
                  type="button"
                  disabled
                  className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3 opacity-40 cursor-not-allowed"
                >
                  <SkipForward className="h-4 w-4 text-muted-foreground" />
                  <span className="text-xs font-medium text-muted-foreground">跳过</span>
                </button>
              </HoverCardTrigger>
              <HoverCardContent className="w-auto px-3 py-1.5 text-sm" side="top">
                竞技模式禁用
              </HoverCardContent>
            </HoverCard>
          ) : (
            <button
              type="button"
              onClick={skipItem}
              className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
            >
              <SkipForward className="h-4 w-4 text-muted-foreground" />
              <span className="text-xs font-medium text-muted-foreground">跳过</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
