import { HEATMAP_TIERS, type HeatmapDay } from "@/features/web/hall/types/heatmap";

/** Get the intensity tier (0-4) for an answer count */
export function getTier(count: number): number {
  if (count === 0) return 0;
  const idx = HEATMAP_TIERS.findIndex((t) => count >= t.min && count <= t.max);
  return idx === -1 ? 0 : idx + 1;
}

/** Build a Map of "YYYY-MM-DD" → count from HeatmapDay[] */
export function buildDayMap(days: HeatmapDay[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const d of days) {
    map.set(d.date, d.count);
  }
  return map;
}

/** Format a Date as "YYYY-MM-DD" */
export function formatDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Get the Monday on or before a given date */
function toMonday(d: Date): Date {
  const copy = new Date(d);
  const day = copy.getDay(); // 0=Sun, 1=Mon, ...
  const diff = day === 0 ? 6 : day - 1;
  copy.setDate(copy.getDate() - diff);
  return copy;
}

/** Generate all grid cells for a year (weeks × 7 days) */
export function generateGridCells(year: number): { date: string; row: number; col: number }[] {
  const jan1 = new Date(year, 0, 1);
  const dec31 = new Date(year, 11, 31);
  const gridStart = toMonday(jan1);

  const cells: { date: string; row: number; col: number }[] = [];
  const cursor = new Date(gridStart);
  let col = 0;

  while (cursor <= dec31 || cursor.getDay() !== 1) {
    for (let row = 0; row < 7; row++) {
      cells.push({ date: formatDate(cursor), row, col });
      cursor.setDate(cursor.getDate() + 1);
    }
    col++;
    if (col > 53) break;
  }

  return cells;
}

/** Get month label positions for the grid header */
export function getMonthLabels(year: number): { label: string; col: number }[] {
  const jan1 = new Date(year, 0, 1);
  const gridStart = toMonday(jan1);
  const labels: { label: string; col: number }[] = [];

  for (let month = 0; month < 12; month++) {
    const firstOfMonth = new Date(year, month, 1);
    const daysDiff = Math.floor(
      (firstOfMonth.getTime() - gridStart.getTime()) / (1000 * 60 * 60 * 24)
    );
    const col = Math.floor(daysDiff / 7);
    labels.push({ label: `${month + 1}月`, col });
  }

  return labels;
}

/** Compute summary stats from heatmap days */
export function computeHeatmapStats(days: HeatmapDay[]) {
  const activeDays = days.length;
  const totalAnswers = days.reduce((sum, d) => sum + d.count, 0);
  const avgPerDay = activeDays > 0 ? Math.round(totalAnswers / activeDays) : 0;

  const tierCounts = [0, 0, 0, 0];
  for (const d of days) {
    const tier = getTier(d.count);
    if (tier > 0) tierCounts[tier - 1]++;
  }

  return { activeDays, totalAnswers, avgPerDay, tierCounts };
}
