import { HEATMAP_TIERS, HEATMAP_COLORS } from "@/features/web/hall/types/heatmap";

type HeatmapSummaryProps = {
  activeDays: number;
  avgPerDay: number;
  tierCounts: number[];
};

/** Right-side panel showing yearly activity and intensity breakdown */
export function HeatmapSummary({
  activeDays,
  avgPerDay,
  tierCounts,
}: HeatmapSummaryProps) {
  return (
    <div className="flex flex-col gap-4">
      {/* Yearly activity */}
      <div className="flex flex-col gap-2 rounded-xl border border-border bg-card p-4">
        <h4 className="text-sm font-bold text-foreground">年度活跃</h4>
        <p className="text-[13px] text-muted-foreground">
          活跃天数 {activeDays} 天
        </p>
        <p className="text-[13px] text-muted-foreground">
          日均 {avgPerDay} 题/天
        </p>
      </div>

      {/* Intensity breakdown */}
      <div className="flex flex-col gap-2 rounded-xl border border-border bg-card p-4">
        <h4 className="text-sm font-bold text-foreground">学习强度</h4>
        <div className="flex flex-col gap-1.5">
          {HEATMAP_TIERS.map((tier, i) => (
            <div key={tier.label} className="flex items-center gap-2 text-[13px] text-muted-foreground">
              <span className={`inline-block h-3 w-3 rounded-sm ${HEATMAP_COLORS[i + 1]}`} />
              {tier.label} · {tierCounts[i]} 天
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
