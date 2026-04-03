import { useState } from "react";
import { Lock, Star } from "lucide-react";
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog";

function LevelCell({
  level,
  name,
  status,
  onClick,
}: {
  level: number;
  name: string;
  status: "completed" | "current" | "locked";
  onClick?: () => void;
}) {
  const shortName = name.slice(0, 4);

  if (status === "completed") {
    return (
      <div
        onClick={onClick}
        className="flex h-[67px] w-[67px] cursor-pointer flex-col items-center justify-center gap-0.5 rounded-[10px] bg-teal-600"
      >
        <span className="text-lg font-extrabold text-white">{level}</span>
        <span className="text-[9px] font-medium text-white/80">
          {shortName}
        </span>
        <StarGroup count={level % 3 === 0 ? 3 : level % 2 === 0 ? 2 : 1} />
        <span className="text-[7px] font-semibold text-white/70">已完成</span>
      </div>
    );
  }

  if (status === "current") {
    return (
      <div
        onClick={onClick}
        className="flex h-[67px] w-[67px] cursor-pointer flex-col items-center justify-center gap-0.5 rounded-[10px] border-2 border-teal-600 bg-card"
      >
        <span className="text-lg font-extrabold text-teal-600">{level}</span>
        <span className="text-[9px] font-medium text-teal-600/80">
          {shortName}
        </span>
      </div>
    );
  }

  return (
    <div
      onClick={onClick}
      className={`relative flex h-[67px] w-[67px] flex-col items-center justify-center gap-0.5 rounded-[10px] bg-muted${onClick ? " cursor-pointer" : ""}`}
    >
      <div className="absolute -left-[5px] -top-[5px] flex h-4 w-4 items-center justify-center rounded-full bg-amber-500">
        <Lock className="h-[9px] w-[9px] text-white" />
      </div>
      <span className="text-lg font-extrabold text-muted-foreground">{level}</span>
      <span className="text-[9px] font-medium text-muted-foreground">
        {shortName}
      </span>
    </div>
  );
}

function StarGroup({ count }: { count: number }) {
  return (
    <div className="flex gap-0.5">
      {[1, 2, 3].map((i) => (
        <Star
          key={i}
          className={`h-3 w-3 ${
            i <= count
              ? "fill-amber-400 text-amber-400"
              : "text-muted-foreground"
          }`}
        />
      ))}
    </div>
  );
}

export function LevelGrid({
  levels,
  completedLevels,
  isVip,
  onLevelClick,
}: {
  levels: { id: string; name: string; order: number }[];
  completedLevels: number;
  isVip: boolean;
  onLevelClick?: (levelId: string, levelName: string) => void;
}) {
  const [upgradeOpen, setUpgradeOpen] = useState(false);
  const totalLevels = levels.length;

  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-4 lg:p-6">
      {/* Header */}
      <div className="flex w-full items-center justify-between">
        <h3 className="text-lg font-bold text-foreground">选择关卡</h3>
        <span className="text-[13px] text-muted-foreground">
          共 {totalLevels} 个关卡
        </span>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-4 gap-1.5 md:grid-cols-5 lg:grid-cols-4 xl:grid-cols-7 2xl:grid-cols-10">
        {levels.map((lv, idx) => {
          const levelNum = idx + 1;
          const status = isVip
            ? "current"
            : levelNum <= completedLevels
              ? "completed"
              : levelNum === completedLevels + 1
                ? "current"
                : "locked";

          return (
            <LevelCell
              key={lv.id}
              level={levelNum}
              name={lv.name}
              status={status}
              onClick={
                status === "locked"
                  ? () => setUpgradeOpen(true)
                  : onLevelClick
                    ? () => onLevelClick(lv.id, lv.name)
                    : undefined
              }
            />
          );
        })}
      </div>

      <UpgradeDialog open={upgradeOpen} onOpenChange={setUpgradeOpen} />
    </div>
  );
}
