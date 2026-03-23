export type ScoreRating = {
  label: string;
  colorClass: string;
  bgClass: string;
};

export const SCORE_RATINGS = [
  { minAccuracy: 0.9, label: "优秀", colorClass: "text-teal-600", bgClass: "bg-teal-50" },
  { minAccuracy: 0.7, label: "良好", colorClass: "text-blue-600", bgClass: "bg-blue-50" },
  { minAccuracy: 0.6, label: "及格", colorClass: "text-amber-600", bgClass: "bg-amber-50" },
  { minAccuracy: 0, label: "继续加油", colorClass: "text-rose-500", bgClass: "bg-rose-50" },
] as const;

/** Get score rating label and color based on accuracy (0-1) */
export function getScoreRating(accuracy: number): ScoreRating {
  const match = SCORE_RATINGS.find((r) => accuracy >= r.minAccuracy);
  return match ?? { label: "继续加油", colorClass: "text-rose-500", bgClass: "bg-rose-50" };
}
