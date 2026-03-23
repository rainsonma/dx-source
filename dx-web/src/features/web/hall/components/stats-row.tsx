import { Zap, Flame, BookOpen, Target } from "lucide-react";

interface StatsRowProps {
  exp: number;
  currentPlayStreak: number;
  masteredTotal: number;
  masteredThisWeek: number;
  reviewPending: number;
}

/** Dashboard stats row showing EXP, streak, mastered, and review counts */
export function StatsRow({
  exp,
  currentPlayStreak,
  masteredTotal,
  masteredThisWeek,
  reviewPending,
}: StatsRowProps) {
  const stats = [
    {
      icon: Zap,
      iconColor: "text-teal-600",
      label: "总经验值",
      value: exp.toLocaleString(),
      sub: "累计获得经验值",
    },
    {
      icon: Flame,
      iconColor: "text-orange-500",
      label: "连续学习",
      value: `${currentPlayStreak} 天`,
      sub: "保持连续学习记录！",
    },
    {
      icon: BookOpen,
      iconColor: "text-violet-500",
      label: "已掌握词汇",
      value: masteredTotal.toLocaleString(),
      sub: `本周新增 ${masteredThisWeek} 个`,
    },
    {
      icon: Target,
      iconColor: "text-teal-600",
      label: "待复习词汇",
      value: reviewPending.toLocaleString(),
      sub: "今日待复习",
    },
  ];

  return (
    <div className="grid w-full grid-cols-2 gap-4 lg:grid-cols-4">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="flex w-full flex-col gap-2 rounded-[14px] border border-border bg-card p-5"
        >
          <div className="flex items-center gap-2">
            <stat.icon className={`h-[18px] w-[18px] ${stat.iconColor}`} />
            <span className="text-[13px] font-medium text-muted-foreground">
              {stat.label}
            </span>
          </div>
          <span className="text-[28px] font-extrabold text-foreground">
            {stat.value}
          </span>
          <span className="text-xs text-muted-foreground">{stat.sub}</span>
        </div>
      ))}
    </div>
  );
}
