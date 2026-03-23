"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowRight, Gamepad2, ListChecks } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { getProgressColor } from "@/features/web/hall/helpers/progress-color.helper";
type SessionProgressItem = {
  id: string;
  gameId: string;
  gameName: string;
  gameMode: string;
  degree: string;
  pattern: string | null;
  playedLevelsCount: number;
  totalLevelsCount: number;
  score: number;
  exp: number;
  lastPlayedAt: Date;
  endedAt: Date | null;
};

const PAGE_SIZE = 5;

/** Compute progress percentage from played/total levels */
function calcProgress(played: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((played / total) * 100);
}

type GameProgressCardProps = {
  sessions: SessionProgressItem[];
};

/** Dashboard card listing user's game session progress with pagination */
export function GameProgressCard({ sessions }: GameProgressCardProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const totalPages = Math.max(1, Math.ceil(sessions.length / PAGE_SIZE));

  const start = (currentPage - 1) * PAGE_SIZE;
  const pageItems = sessions.slice(start, start + PAGE_SIZE);
  const isEmpty = sessions.length === 0;

  return (
    <div className="flex w-full flex-col gap-5 rounded-[14px] border border-border bg-card p-6">
      {/* Header */}
      <div className="flex w-full items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-teal-50">
            <ListChecks className="h-5 w-5 text-teal-600" />
          </div>
          <h3 className="text-base font-bold text-foreground">我的学习进度</h3>
        </div>
        {!isEmpty && (
          <Link
            href="/hall/games/mine"
            className="flex items-center gap-1 text-[13px] font-semibold text-teal-600 hover:text-teal-700"
          >
            查看全部
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        )}
      </div>

      {isEmpty ? (
        <div className="flex min-h-[288px] flex-col items-center justify-center gap-3 text-muted-foreground">
          <Gamepad2 className="h-10 w-10 stroke-1" />
          <p className="text-sm">还没有学习进度，去发现课程游戏吧</p>
        </div>
      ) : (
        <>
          {/* Progress list — min-h: 5 rows × ~48px + 4 gaps × 12px = 288px */}
          <div className="flex min-h-[288px] flex-col gap-3">
            {pageItems.map((session, i) => {
              const modeLabel =
                GAME_MODE_LABELS[session.gameMode as GameMode] ??
                session.gameMode;
              const progress = calcProgress(
                session.playedLevelsCount,
                session.totalLevelsCount
              );
              const color = getProgressColor(start + i);

              return (
                <Link
                  key={session.id}
                  href={`/hall/games/${session.gameId}`}
                  className="flex flex-col gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-accent"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-[13px] font-medium text-foreground">
                      {session.gameName} · {modeLabel}
                    </span>
                    <span className="text-[13px] font-semibold text-muted-foreground">
                      {progress}%
                    </span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className={`h-full rounded-full ${color}`}
                      style={{ width: `${progress}%` }}
                    />
                  </div>
                </Link>
              );
            })}
          </div>

          {/* Pagination */}
          <DataTablePagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
          />
        </>
      )}
    </div>
  );
}
