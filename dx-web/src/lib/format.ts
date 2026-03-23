/** Format a Date to zh-CN locale string, handles serialized dates from server actions */
export function formatDate(date: Date | string | null): string {
  if (!date) return "-";
  return new Date(date).toLocaleDateString("zh-CN");
}

/** Format seconds into human-readable duration (e.g., "3h 25m", "48m", "< 1m") */
export function formatPlayTime(seconds: number): string {
  if (seconds < 60) return "< 1m";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h`;
  return `${minutes}m`;
}
