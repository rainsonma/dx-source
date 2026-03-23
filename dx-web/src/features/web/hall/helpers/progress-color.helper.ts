const PROGRESS_COLORS = [
  "bg-teal-500",
  "bg-blue-500",
  "bg-amber-500",
  "bg-pink-500",
  "bg-violet-500",
  "bg-cyan-500",
];

/** Cycle through progress bar colors by index */
export function getProgressColor(index: number): string {
  return PROGRESS_COLORS[index % PROGRESS_COLORS.length];
}
