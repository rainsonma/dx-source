import { TrendingUp } from "lucide-react";

export function MyStatsCard({
  stats,
}: {
  stats: { label: string; value: string }[];
}) {
  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-6">
      <div className="flex items-center gap-2">
        <TrendingUp className="h-[18px] w-[18px] text-teal-600" />
        <h3 className="text-base font-bold text-foreground">我的战绩</h3>
        <span className="ml-auto rounded-full bg-teal-100 px-2 py-0.5 text-xs font-medium text-teal-700 dark:bg-teal-900/30 dark:text-teal-400">
          近一月
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="flex flex-col gap-1 rounded-[10px] bg-muted p-3.5"
          >
            <span className="text-[22px] font-extrabold text-foreground">
              {stat.value}
            </span>
            <span className="text-xs text-muted-foreground">{stat.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
