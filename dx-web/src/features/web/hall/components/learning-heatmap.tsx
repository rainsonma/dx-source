"use client";

import { useState, useTransition } from "react";
import { CalendarDays } from "lucide-react";
import { HeatmapGrid } from "@/features/web/hall/components/heatmap-grid";
import { HeatmapSummary } from "@/features/web/hall/components/heatmap-summary";
import { computeHeatmapStats } from "@/features/web/hall/helpers/heatmap";
import { fetchHeatmapDataAction } from "@/features/web/hall/actions/heatmap.action";
import type { HeatmapDay } from "@/features/web/hall/types/heatmap";

type LearningHeatmapProps = {
  initialYear: number;
  initialDays: HeatmapDay[];
  accountYear: number;
};

/** Learning heatmap block with year selector */
export function LearningHeatmap({
  initialYear,
  initialDays,
  accountYear,
}: LearningHeatmapProps) {
  const currentYear = new Date().getFullYear();
  const years = Array.from(
    { length: currentYear - accountYear + 1 },
    (_, i) => currentYear - i
  );

  const [selectedYear, setSelectedYear] = useState(initialYear);
  const [days, setDays] = useState(initialDays);
  const [isPending, startTransition] = useTransition();

  const stats = computeHeatmapStats(days);

  /** Switch to a different year */
  function handleYearChange(year: number) {
    if (year === selectedYear) return;
    const prevYear = selectedYear;
    setSelectedYear(year);
    startTransition(async () => {
      const result = await fetchHeatmapDataAction(year);
      if (result.data) {
        setDays(result.data.days);
      } else {
        setSelectedYear(prevYear);
      }
    });
  }

  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-teal-50">
          <CalendarDays className="h-5 w-5 text-teal-600" />
        </div>
        <div>
          <h3 className="text-base font-bold text-foreground">学习热力图</h3>
          <p className="text-[13px] text-muted-foreground">
            {selectedYear} 年共学习 {stats.activeDays} 天 · 累计{" "}
            {stats.totalAnswers.toLocaleString()} 题
          </p>
        </div>
      </div>

      {/* Body: grid + summary + year selector */}
      <div className="flex flex-col gap-5 lg:flex-row">
        {/* Left: heatmap grid */}
        <div
          className={`min-w-0 flex-1 overflow-x-auto rounded-xl border border-border p-4 ${
            isPending ? "opacity-50" : ""
          }`}
        >
          <HeatmapGrid year={selectedYear} days={days} />
        </div>

        {/* Right: summary + year selector */}
        <div className="flex w-full flex-col gap-4 lg:w-44 lg:shrink-0">
          {/* Year selector */}
          <div className="flex flex-col items-end gap-2">
            <span className="text-xs text-muted-foreground">年份</span>
            {years.map((year) => (
              <button
                key={year}
                onClick={() => handleYearChange(year)}
                disabled={isPending}
                className={`rounded-full border px-3.5 py-1 text-sm font-medium transition-colors ${
                  year === selectedYear
                    ? "border-teal-500 bg-teal-50 text-teal-700"
                    : "border-border text-muted-foreground hover:border-border"
                }`}
              >
                {year}
              </button>
            ))}
          </div>

          {/* Stats */}
          <HeatmapSummary
            activeDays={stats.activeDays}
            avgPerDay={stats.avgPerDay}
            tierCounts={stats.tierCounts}
          />
        </div>
      </div>
    </div>
  );
}
