import Link from "next/link";
import { Play, Trophy, Clock } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import { formatPlayTime } from "@/lib/format";
type PlayedGameCardType = {
  id: string;
  name: string;
  description: string | null;
  mode: string;
  cover: { url: string } | null;
  category: { name: string } | null;
  user: { username: string } | null;
  highestScore: number;
  totalPlayTime: number;
};

const GRADIENT_COVERS = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

/** Deterministic gradient based on id hash */
function getGradient(id: string) {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
  return GRADIENT_COVERS[Math.abs(hash) % GRADIENT_COVERS.length];
}

export function PlayedGameCard({ game }: { game: PlayedGameCardType }) {
  const modeLabel = GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;

  return (
    <Link
      href={`/hall/games/${game.id}`}
      className="flex w-full flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
    >
      {game.cover ? (
        /* eslint-disable-next-line @next/next/no-img-element */
        <img
          src={game.cover.url}
          alt={game.name}
          className="h-[180px] w-full object-cover"
        />
      ) : (
        <div
          className={`flex h-[180px] w-full items-center justify-center ${getGradient(game.id)}`}
        >
          <span className="text-lg font-bold text-white/80">{modeLabel}</span>
        </div>
      )}

      <div className="flex flex-1 flex-col justify-between gap-2 px-3.5 py-3">
        <div className="flex flex-col gap-1">
          <h4 className="text-sm font-bold text-foreground">{game.name}</h4>
          <p className="line-clamp-2 text-[11px] leading-[1.4] text-muted-foreground">
            {game.description}
          </p>
        </div>
        <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
          <span className="flex items-center gap-1">
            <Trophy className="h-3 w-3" />
            最高 {game.highestScore} 分
          </span>
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {formatPlayTime(game.totalPlayTime)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-medium text-muted-foreground">
            {game.category?.name ?? modeLabel}
          </span>
          <span className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1.5 text-[11px] font-semibold text-white">
            <Play className="h-3 w-3" />
            开始
          </span>
        </div>
      </div>
    </Link>
  );
}
