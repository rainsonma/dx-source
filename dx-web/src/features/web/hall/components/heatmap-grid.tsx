import { HEATMAP_COLORS } from "@/features/web/hall/types/heatmap";
import type { HeatmapDay } from "@/features/web/hall/types/heatmap";
import {
  generateGridCells,
  getMonthLabels,
  buildDayMap,
  getTier,
} from "@/features/web/hall/helpers/heatmap";

type HeatmapGridProps = {
  year: number;
  days: HeatmapDay[];
};

const DAY_LABELS = ["一", "", "三", "", "五", "", ""];

/** GitHub-style heatmap grid with month and day labels */
export function HeatmapGrid({ year, days }: HeatmapGridProps) {
  const cells = generateGridCells(year);
  const dayMap = buildDayMap(days);
  const monthLabels = getMonthLabels(year);
  const totalCols = cells.length > 0 ? cells[cells.length - 1].col + 1 : 53;

  return (
    <div className="flex flex-col gap-2">
      {/* Month labels */}
      <div
        className="grid gap-[3px] text-xs text-muted-foreground"
        style={{
          gridTemplateColumns: `24px repeat(${totalCols}, 1fr)`,
        }}
      >
        <span /> {/* spacer for day labels column */}
        {Array.from({ length: totalCols }, (_, col) => {
          const label = monthLabels.find((m) => m.col === col);
          return (
            <span key={col} className="text-center text-[11px]">
              {label?.label ?? ""}
            </span>
          );
        })}
      </div>

      {/* Grid with day labels */}
      <div className="flex gap-[3px]">
        {/* Day labels */}
        <div className="flex w-6 flex-col gap-[3px]">
          {DAY_LABELS.map((label, i) => (
            <div
              key={i}
              className="flex h-[14px] items-center text-[11px] text-muted-foreground"
            >
              {label}
            </div>
          ))}
        </div>

        {/* Grid cells */}
        <div
          className="grid flex-1 gap-[3px]"
          style={{
            gridTemplateRows: "repeat(7, 14px)",
            gridTemplateColumns: `repeat(${totalCols}, 1fr)`,
            gridAutoFlow: "column",
          }}
        >
          {cells.map((cell) => {
            const count = dayMap.get(cell.date) ?? 0;
            const tier = getTier(count);
            const isInYear =
              cell.date.startsWith(String(year));

            return (
              <div
                key={cell.date}
                className={`h-[14px] rounded-[3px] ${
                  isInYear ? HEATMAP_COLORS[tier] : "bg-transparent"
                }`}
                title={
                  isInYear
                    ? `${cell.date}: ${count > 0 ? `${count} 题` : "无记录"}`
                    : undefined
                }
              />
            );
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
        <span>少</span>
        {HEATMAP_COLORS.map((color, i) => (
          <div
            key={i}
            className={`h-[12px] w-[12px] rounded-[2px] ${color}`}
          />
        ))}
        <span>多</span>
      </div>
    </div>
  );
}
