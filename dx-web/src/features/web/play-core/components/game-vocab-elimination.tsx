"use client";

import { Sparkles, ArrowRight } from "lucide-react";
import { useVocabElimination } from "@/features/web/play-core/hooks/use-vocab-elimination";

export function GameVocabElimination() {
  const {
    gridRows,
    selectedTileId,
    eliminatedIndices,
    wrongPair,
    progress,
    combo,
    selectTile,
    isBatchComplete,
    isLastBatch,
    nextBatch,
  } = useVocabElimination();

  if (gridRows.length === 0) return null;

  return (
    <div className="flex w-full max-w-[700px] flex-col items-center gap-5">
      {/* Status row */}
      <div className="flex w-full flex-col items-center gap-3 sm:flex-row sm:justify-between">
        <span className="text-sm font-semibold text-foreground">
          已消除 {progress.current}/{progress.total} 对
        </span>
        <div className="h-1.5 w-full max-w-[300px] rounded-full bg-border">
          <div
            className="h-1.5 rounded-full bg-gradient-to-r from-pink-500 to-teal-500 transition-all duration-300"
            style={{
              width: `${(progress.current / Math.max(progress.total, 1)) * 100}%`,
            }}
          />
        </div>
        {combo.streak >= 2 && (
          <div className="flex items-center gap-1.5 rounded-lg border border-pink-500/20 bg-pink-50 px-3 py-1">
            <Sparkles className="h-3.5 w-3.5 text-pink-500" />
            <span className="text-xs font-bold text-pink-500">
              连击 &times;{combo.streak}
            </span>
          </div>
        )}
      </div>

      {/* Grid */}
      <div className="flex w-full flex-col gap-2.5 rounded-[20px] border border-border bg-card p-4 shadow-sm md:p-6">
        {gridRows.map((row, ri) => (
          <div key={ri} className="flex gap-2.5">
            {row.map((tile) => {
              const isEliminated = eliminatedIndices.has(tile.itemIndex);
              const isSelected = selectedTileId === tile.id;
              const isWrong =
                wrongPair?.t1 === tile.id || wrongPair?.t2 === tile.id;

              return (
                <button
                  key={tile.id}
                  type="button"
                  disabled={isEliminated}
                  onClick={() => selectTile(tile.id)}
                  className={`flex h-14 flex-1 items-center justify-center rounded-xl border transition-all md:h-16 ${
                    isEliminated
                      ? "border-border/20 bg-muted opacity-40"
                      : isWrong
                        ? "animate-[shake_0.3s_ease-in-out] border-2 border-red-400 bg-red-50"
                        : isSelected
                          ? "border-2 border-pink-500 bg-pink-50"
                          : "border-[1.5px] border-border bg-card hover:bg-muted/50"
                  }`}
                >
                  <span
                    className={`text-sm font-medium ${
                      isEliminated
                        ? "text-muted-foreground line-through"
                        : isWrong
                          ? "text-red-600"
                          : isSelected
                            ? "text-pink-600"
                            : "text-foreground"
                    }`}
                  >
                    {tile.text}
                  </span>
                </button>
              );
            })}
          </div>
        ))}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          点击两个匹配的方块进行消除
        </p>
        {!isLastBatch && (
          <button
            type="button"
            disabled={!isBatchComplete}
            onClick={nextBatch}
            className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-40"
          >
            下一组
            <ArrowRight className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
