import Link from "next/link";
import { Play } from "lucide-react";

import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";

type GameCard = {
  id: string;
  name: string;
  description?: string | null;
  mode: string;
  status: string;
  createdAt?: string | Date | null;
  coverUrl?: string | null;
  levelCount?: number;
  // Legacy Prisma shape compatibility
  cover?: { url: string } | null;
  _count?: { levels: number };
};

const coverColors = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

function pickCoverColor(id: string) {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0;
  }
  return coverColors[Math.abs(hash) % coverColors.length];
}

type StatusVariant = "published" | "withdraw" | "draft";

const statusStyles: Record<StatusVariant, { bg: string; label: string }> = {
  published: { bg: "bg-green-600", label: "已发布" },
  withdraw: { bg: "bg-amber-600", label: "已撤回" },
  draft: { bg: "bg-slate-500", label: "未发布" },
};

export function GameCardItem({ game, asDiv, onClick }: { game: GameCard; asDiv?: boolean; onClick?: () => void }) {
  const status = (game.status === "published" ? "published" : game.status === "withdraw" ? "withdraw" : "draft") as StatusVariant;
  const s = statusStyles[status];
  const modeLabel = GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;
  const dateStr = game.createdAt
    ? new Date(game.createdAt).toLocaleDateString("zh-CN", {
        year: "numeric",
        month: "2-digit",
      })
    : "";

  const cardContent = (
    <>
      {/* Cover */}
      <div className={`relative flex h-[120px] items-center justify-center ${(game.coverUrl || game.cover?.url) ? "bg-border" : pickCoverColor(game.id)}`}>
        {(game.coverUrl || game.cover?.url) ? (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={game.coverUrl ?? game.cover?.url ?? ""}
            alt={game.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <span className="text-2xl font-bold text-white/80">{modeLabel}</span>
        )}
        <div className="absolute inset-x-0 top-0 flex h-9 items-start justify-between px-2 pt-2">
          <span
            className={`rounded-md px-2 py-0.5 text-[10px] font-semibold text-white ring-1 ring-white/30 drop-shadow-md ${s.bg}`}
          >
            {s.label}
          </span>
        </div>
      </div>

      {/* Body */}
      <div className="flex flex-1 flex-col justify-between gap-2 p-3">
        <div className="flex flex-col gap-1">
          <span className="text-sm font-bold text-foreground">{game.name}</span>
          {game.description && (
            <span className="line-clamp-2 text-[11px] leading-snug text-muted-foreground">
              {game.description}
            </span>
          )}
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-[11px] text-muted-foreground">
            {game.levelCount ?? game._count?.levels ?? 0} 个学习单元
          </span>
          <span className="text-[11px] text-muted-foreground">{dateStr}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="rounded-[10px] bg-teal-50 px-2 py-0.5 text-[11px] font-medium text-teal-600">
            {modeLabel}
          </span>
          <span className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white">
            <Play className="h-3 w-3" />
            进入
          </span>
        </div>
      </div>
    </>
  );

  if (asDiv) {
    return (
      <div
        onClick={onClick}
        className="flex cursor-pointer flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
      >
        {cardContent}
      </div>
    );
  }

  return (
    <Link
      href={`/hall/ai-custom/${game.id}`}
      className="flex flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
    >
      {cardContent}
    </Link>
  );
}
