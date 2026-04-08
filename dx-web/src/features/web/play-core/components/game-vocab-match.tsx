"use client";

import { Zap, Circle, CheckCircle2 } from "lucide-react";
import { useVocabMatch } from "@/features/web/play-core/hooks/use-vocab-match";

export function GameVocabMatch() {
  const {
    batchItems,
    shuffledDefs,
    selectedWordIndex,
    matchedIndices,
    wrongPair,
    progress,
    combo,
    selectWord,
    selectDef,
  } = useVocabMatch();

  if (batchItems.length === 0) return null;

  return (
    <div className="flex w-full max-w-3xl flex-col gap-7 rounded-[20px] border border-border bg-card p-6 shadow-sm md:p-8">
      {/* Progress */}
      <div className="flex flex-col gap-2.5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">
              进度 {progress.current}/{progress.total}
            </span>
          </div>
          {combo.streak >= 3 && (
            <div className="flex items-center gap-1.5 rounded-lg bg-teal-600/10 px-3 py-1">
              <Zap className="h-3.5 w-3.5 text-teal-600" />
              <span className="text-xs font-bold text-teal-600">
                连击 &times;{combo.streak}
              </span>
            </div>
          )}
        </div>
        <div className="h-1.5 w-full rounded-full bg-border">
          <div
            className="h-1.5 rounded-full bg-gradient-to-r from-blue-500 to-teal-500 transition-all duration-300"
            style={{
              width: `${(progress.current / Math.max(progress.total, 1)) * 100}%`,
            }}
          />
        </div>
      </div>

      {/* Match area */}
      <div className="flex flex-col gap-4 sm:flex-row sm:gap-6">
        {/* English words */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            英文单词
          </span>
          {batchItems.map((item, i) => {
            const isMatched = matchedIndices.has(i);
            const isSelected = selectedWordIndex === i;
            const isWrong = wrongPair?.word === i;
            return (
              <button
                key={item.id}
                type="button"
                disabled={isMatched}
                onClick={() => selectWord(i)}
                className={`flex items-center gap-2.5 rounded-xl border px-4 py-3 transition-colors ${
                  isMatched
                    ? "border-emerald-300 bg-emerald-50"
                    : isWrong
                      ? "animate-[shake_0.3s_ease-in-out] border-red-400 bg-red-50"
                      : isSelected
                        ? "border-blue-400 bg-blue-50"
                        : "border-border bg-card hover:bg-muted/50"
                }`}
              >
                {isMatched ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                ) : (
                  <Circle
                    className={`h-4 w-4 ${isSelected ? "text-blue-400" : "text-slate-300"}`}
                  />
                )}
                <span
                  className={`text-sm font-medium ${
                    isMatched
                      ? "text-emerald-600"
                      : isSelected
                        ? "text-blue-600"
                        : "text-foreground"
                  }`}
                >
                  {item.content}
                </span>
              </button>
            );
          })}
        </div>

        {/* Chinese definitions */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            中文释义
          </span>
          {shuffledDefs.map((def) => {
            const isMatched = matchedIndices.has(def.batchIndex);
            const isWrong = wrongPair?.def === def.batchIndex;
            return (
              <button
                key={`def-${def.batchIndex}`}
                type="button"
                disabled={isMatched}
                onClick={() => selectDef(def.batchIndex)}
                className={`flex items-center justify-center rounded-xl border px-4 py-3 transition-colors ${
                  isMatched
                    ? "border-emerald-300 bg-emerald-50"
                    : isWrong
                      ? "animate-[shake_0.3s_ease-in-out] border-red-400 bg-red-50"
                      : "border-border bg-card hover:bg-muted/50"
                }`}
              >
                <span
                  className={`text-sm font-medium ${
                    isMatched ? "text-emerald-600" : "text-foreground"
                  }`}
                >
                  {def.translation}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      <p className="text-center text-xs text-muted-foreground">
        点击左侧单词，再点击右侧匹配的释义
      </p>
    </div>
  );
}
