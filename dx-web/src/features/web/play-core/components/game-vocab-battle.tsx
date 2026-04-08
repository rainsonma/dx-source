"use client";

import { Zap, SkipForward, Check } from "lucide-react";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { useVocabBattle } from "@/features/web/play-core/hooks/use-vocab-battle";

export function GameVocabBattle() {
  const {
    targetWord,
    translation,
    letterSlots,
    keyboardLetters,
    usedKeyIndices,
    hasError,
    isRevealed,
    playerShields,
    opponentShields,
    opponentSlots,
    competitive,
    progress,
    combo,
    pressLetter,
    advanceAfterReveal,
    skipItem,
  } = useVocabBattle();

  if (!targetWord) return null;

  return (
    <div className="flex w-full max-w-[760px] flex-col rounded-[20px] border border-border bg-card shadow-sm">
      {/* Opponent zone */}
      <div
        className={`flex flex-col items-center gap-4 px-6 py-7 md:px-8 ${
          !competitive ? "pointer-events-none opacity-40" : ""
        }`}
      >
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🤖 对手</span>
        </div>
        <div className="flex items-center justify-center gap-2">
          {opponentShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 transition-colors ${
                active
                  ? "border-red-400 bg-red-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center justify-center gap-2.5">
          {opponentSlots.map((slot, i) => (
            <div
              key={i}
              className="flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-muted"
            >
              <span className="text-sm font-medium text-slate-300">
                {slot.filled ? slot.letter : "?"}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Translation zone */}
      <div className="flex flex-col items-center gap-2.5 bg-gradient-to-b from-red-50/0 via-red-50 to-red-50/0 px-6 py-4 md:px-8">
        <p className="text-center text-2xl font-extrabold tracking-wider text-foreground md:text-[32px]">
          {translation}
        </p>
        <div className="h-0.5 w-full rounded-full bg-gradient-to-r from-red-500/0 via-red-500/30 via-30% via-teal-500/30 via-70% to-teal-500/0" />
      </div>

      {/* Player zone */}
      <div className="flex flex-col items-center gap-4 px-6 py-5 md:px-8">
        <div
          className={`flex items-center justify-center gap-2.5 ${
            hasError ? "animate-[shake_0.3s_ease-in-out]" : ""
          }`}
        >
          {letterSlots.map((slot, i) => (
            <div
              key={i}
              className={`flex h-10 w-10 items-center justify-center rounded-lg border transition-colors ${
                slot.filled
                  ? "border-teal-300 bg-teal-50"
                  : "border-border bg-muted"
              }`}
            >
              <span
                className={`text-sm font-semibold ${
                  slot.filled ? "text-teal-600" : "text-slate-300"
                }`}
              >
                {slot.filled ? slot.filledLetter : "_"}
              </span>
            </div>
          ))}
        </div>
        <div className="flex items-center justify-center gap-2">
          {playerShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 transition-colors ${
                active
                  ? "border-teal-400 bg-teal-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🎯 我</span>
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

      {/* Hint + action row */}
      <div className="flex flex-col items-center gap-3 px-6 pb-6 pt-3 md:px-8">
        <span className="text-xs font-medium text-muted-foreground">
          {competitive ? "点击字母发射炮弹击碎对手护盾" : "拼写单词"}
        </span>

        {/* Letter keyboard */}
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
                  <span className="text-lg font-bold text-white">{letter}</span>
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

        {/* Skip button (non-competitive only) */}
        {!isRevealed && (
          <div className="flex items-center gap-3">
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
        )}
      </div>
    </div>
  );
}
