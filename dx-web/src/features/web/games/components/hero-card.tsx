import Link from "next/link";
import { Play, Heart, Layers, Users, Swords } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";

const GRADIENT_COVERS = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

function getGradient(id: string) {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
  return GRADIENT_COVERS[Math.abs(hash) % GRADIENT_COVERS.length];
}

export function HeroCard({
  id,
  title,
  description,
  mode,
  levelCount,
  playerCount,
  coverUrl,
  resumeLabel,
  onStart,
  onPkStart,
  isFavorited,
  onFavoriteToggle,
  isFavoritePending,
}: {
  id: string;
  title: string;
  description: string;
  mode: string;
  levelCount: number;
  playerCount: string;
  coverUrl: string | null;
  resumeLabel?: string | null;
  onStart?: () => void;
  onPkStart?: () => void;
  isFavorited: boolean;
  onFavoriteToggle: () => void;
  isFavoritePending: boolean;
}) {
  const modeLabel = GAME_MODE_LABELS[mode as GameMode] ?? mode;

  return (
    <div className="overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="flex w-full flex-col gap-5 p-4 lg:flex-row lg:gap-7 lg:p-6">
        {/* Cover image / fallback */}
        {coverUrl ? (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={coverUrl}
            alt={title}
            className="h-[200px] w-full shrink-0 rounded-xl object-cover lg:h-[280px] lg:w-[280px]"
          />
        ) : (
          <div
            className={`flex h-[200px] w-full shrink-0 items-center justify-center rounded-xl lg:h-[280px] lg:w-[280px] ${getGradient(id)}`}
          >
            <span className="text-3xl font-bold text-white/80">
              {modeLabel}
            </span>
          </div>
        )}

        {/* Right content */}
        <div className="flex flex-col gap-4 lg:flex-1 lg:justify-between">
          <div className="flex flex-col gap-4">
            <h2 className="text-[28px] font-extrabold text-foreground">
              {title}
            </h2>
            <p className="text-sm leading-[1.6] text-muted-foreground">
              {description}
            </p>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <Layers className="h-4 w-4 text-muted-foreground" />
                <span className="text-[13px] font-medium text-muted-foreground">
                  {levelCount} 课/关
                </span>
              </div>
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-muted-foreground" />
                <span className="text-[13px] font-medium text-muted-foreground">
                  {playerCount} 人在玩
                </span>
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={onStart}
              className="flex items-center gap-2 rounded-[10px] bg-teal-600 px-7 py-3 text-[15px] font-bold text-white hover:bg-teal-700"
            >
              {resumeLabel ? `继续学习「${resumeLabel.length > 6 ? resumeLabel.slice(0, 6) + "…" : resumeLabel}」` : "开始游戏"}
              <Play className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={onPkStart}
              className="flex items-center gap-2 rounded-[10px] border border-border bg-card px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent"
            >
              <Swords className="h-4 w-4" />
              PK
            </button>
            <Link
              href="/hall/groups"
              className="flex items-center gap-2 rounded-[10px] border border-border bg-card px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent"
            >
              <Users className="h-4 w-4" />
              群组
            </Link>
            <button
              type="button"
              onClick={onFavoriteToggle}
              disabled={isFavoritePending}
              className={`flex items-center gap-2 rounded-[10px] border px-5 py-3 text-sm font-medium transition-colors ${
                isFavorited
                  ? "border-rose-200 bg-rose-50 text-rose-500 hover:bg-rose-100"
                  : "border-border bg-card text-muted-foreground hover:bg-accent"
              } disabled:opacity-50`}
            >
              <Heart
                className={`h-4 w-4 ${isFavorited ? "fill-current" : ""}`}
              />
              {isFavorited ? "已收藏" : "收藏"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
