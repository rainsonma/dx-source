/** Fixed spaced-repetition intervals in days, indexed by reviewCount */
export const REVIEW_INTERVAL_DAYS = [1, 3, 7, 14, 30, 90] as const;

/** Compute the next review date from the current review count */
export function getNextReviewAt(reviewCount: number): Date {
  const index = Math.min(reviewCount, REVIEW_INTERVAL_DAYS.length - 1);
  const days = REVIEW_INTERVAL_DAYS[index];
  const next = new Date();
  next.setDate(next.getDate() + days);
  return next;
}
