import Link from "next/link";
import { Layers } from "lucide-react";

import { GAME_MODE_LABELS } from "@/consts/game-mode";
import type { GameMode } from "@/consts/game-mode";
import type { PublicGameCard } from "@/features/web/games/actions/game.action";

const modeColors: Record<string, string> = {
  "lsrw": "bg-teal-600",
  "vocab-battle": "bg-violet-600",
  "vocab-match": "bg-blue-600",
  "vocab-elimination": "bg-amber-600",
  "listening-challenge": "bg-sky-600",
};

const coverGradients = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

const avatarColors = [
  "bg-teal-500",
  "bg-indigo-500",
  "bg-rose-500",
  "bg-amber-500",
  "bg-cyan-500",
  "bg-fuchsia-500",
];

function hashString(str: string) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 31 + str.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

export function GameCard({ game }: { game: PublicGameCard }) {
  const modeLabel = GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;
  const modeBg = modeColors[game.mode] ?? "bg-slate-600";
  const username = game.user?.username ?? "匿名";

  return (
    <Link
      href={`/hall/games/${game.id}`}
      className="flex w-full flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
    >
      {/* Cover */}
      {game.cover?.url ? (
        <div className="h-[180px] w-full bg-muted">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={game.cover.url}
            alt={game.name}
            className="h-full w-full object-cover"
          />
        </div>
      ) : (
        <div
          className={`flex h-[180px] w-full items-center justify-center ${coverGradients[hashString(game.id) % coverGradients.length]}`}
        >
          <span className="text-2xl font-bold text-white/80">{modeLabel}</span>
        </div>
      )}

      {/* Body */}
      <div className="flex flex-1 flex-col justify-between gap-2 px-3.5 py-3">
        <div className="flex flex-col gap-1">
          <h4 className="text-sm font-bold text-foreground">{game.name}</h4>
          <p className="line-clamp-2 text-[11px] leading-[1.4] text-muted-foreground">
            {game.description}
          </p>
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div
              className={`flex h-5 w-5 items-center justify-center rounded-full text-[10px] font-bold text-white ${avatarColors[hashString(username) % avatarColors.length]}`}
            >
              {username.charAt(0).toUpperCase()}
            </div>
            <span className="text-[11px] font-medium text-muted-foreground">
              {username}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
              <Layers className="h-3 w-3" />
              {game._count.levels}
            </div>
            <span
              className={`rounded-md px-2.5 py-1 text-[11px] font-semibold text-white ${modeBg}`}
            >
              {modeLabel}
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}
