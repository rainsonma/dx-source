/** Single day in the heatmap */
export type HeatmapDay = {
  date: string; // "YYYY-MM-DD"
  count: number;
};

/** Full heatmap dataset for a year */
export type HeatmapData = {
  year: number;
  days: HeatmapDay[];
};

/** Intensity tier boundaries (answer counts) */
export const HEATMAP_TIERS = [
  { min: 1, max: 10, label: "1~10 题" },
  { min: 11, max: 30, label: "11~30 题" },
  { min: 31, max: 60, label: "31~60 题" },
  { min: 61, max: Infinity, label: "60+ 题" },
] as const;

/** Teal color palette for heatmap tiers (0=empty, 1-4=tiers) */
export const HEATMAP_COLORS = [
  "bg-muted",       // 0: no activity
  "bg-teal-200",    // 1: 1-10
  "bg-teal-400",    // 2: 11-30
  "bg-teal-600",    // 3: 31-60
  "bg-teal-800",    // 4: 60+
] as const;
