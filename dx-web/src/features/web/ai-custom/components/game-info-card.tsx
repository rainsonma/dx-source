import { Info } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import {
  GAME_STATUS_LABELS,
  type GameStatus,
} from "@/consts/game-status";

type GameInfoCardProps = {
  game: {
    status: string;
    mode: string;
    createdAt: Date;
    updatedAt: Date;
    category: { name: string } | null;
    press: { name: string } | null;
    user: { id: string; username: string } | null;
    _count: { stats: number };
  };
};

const STATUS_STYLES: Record<string, string> = {
  published: "bg-emerald-100 text-emerald-600",
  withdraw: "bg-amber-100 text-amber-600",
  draft: "bg-muted text-muted-foreground",
};

function formatDate(date: Date) {
  return new Date(date).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

export function GameInfoCard({ game }: GameInfoCardProps) {
  const modeLabel =
    GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;
  const statusLabel =
    GAME_STATUS_LABELS[game.status as GameStatus] ?? game.status;
  const statusStyle = STATUS_STYLES[game.status] ?? STATUS_STYLES.draft;

  const rows = [
    {
      label: "作者",
      value: game.user?.username ?? "未知",
      hasAvatar: true,
      avatarLetter: game.user?.username?.[0]?.toUpperCase() ?? "?",
    },
    { label: "分类", value: game.category?.name ?? "-" },
    { label: "模式", value: modeLabel },
    { label: "出版社", value: game.press?.name ?? "-" },
    { label: "参与人数", value: `${game._count.stats} 人` },
    { label: "创建时间", value: formatDate(game.createdAt) },
    { label: "修改时间", value: formatDate(game.updatedAt) },
  ];

  return (
    <div className="flex w-full shrink-0 flex-col gap-4 rounded-[14px] border border-border bg-card p-4 lg:w-80 lg:p-6">
      <div className="flex items-center gap-2">
        <Info className="h-4 w-4 text-teal-600" />
        <span className="text-[15px] font-bold text-foreground">
          课程游戏信息
        </span>
      </div>
      <div className="flex flex-col gap-3">
        {rows.map((row) => (
          <div key={row.label} className="flex items-center justify-between">
            <span className="text-[13px] text-muted-foreground">{row.label}</span>
            {row.hasAvatar ? (
              <div className="flex items-center gap-1.5">
                <div className="flex h-5 w-5 items-center justify-center rounded-[10px] bg-indigo-100">
                  <span className="text-[10px] font-semibold text-indigo-600">
                    {row.avatarLetter}
                  </span>
                </div>
                <span className="text-[13px] font-medium text-foreground">
                  {row.value}
                </span>
              </div>
            ) : (
              <span className="text-[13px] font-medium text-foreground">
                {row.value}
              </span>
            )}
          </div>
        ))}
        {/* Status row */}
        <div className="flex items-center justify-between">
          <span className="text-[13px] text-muted-foreground">状态</span>
          <span
            className={`rounded-md px-2 py-0.5 text-xs font-medium ${statusStyle}`}
          >
            {statusLabel}
          </span>
        </div>
      </div>
    </div>
  );
}
