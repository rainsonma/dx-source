import { apiClient } from "@/lib/api-client";

/** Content item shape from Go API */
type ContentItem = {
  content: string;
  translation: string | null;
  contentType: string;
};

/** Review item shape from Go API ReviewItemData */
export type ReviewItem = {
  id: string;
  contentItem: ContentItem;
  gameId: string;
  gameName: string;
  lastReviewAt: string | null;
  nextReviewAt: string | null;
  reviewCount: number;
  createdAt: string;
};

/** Review stats shape from Go API */
export type ReviewStats = {
  pending: number;
  overdue: number;
  reviewedToday: number;
};

/** Fetch next page of review words via Go API */
export async function fetchReviewsAction(cursor?: string): Promise<{
  items: ReviewItem[];
  nextCursor: string | null;
}> {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  const qs = params.toString();

  const res = await apiClient.get<{
    items: ReviewItem[];
    nextCursor: string;
    hasMore: boolean;
  }>(`/api/tracking/review${qs ? `?${qs}` : ""}`);

  if (res.code !== 0) {
    return { items: [], nextCursor: null };
  }

  return {
    items: res.data.items ?? [],
    nextCursor: res.data.hasMore ? res.data.nextCursor : null,
  };
}

/** Fetch review stats via Go API */
export async function fetchReviewStatsAction(): Promise<ReviewStats> {
  const res = await apiClient.get<ReviewStats>("/api/tracking/review/stats");

  if (res.code !== 0) {
    return { pending: 0, overdue: 0, reviewedToday: 0 };
  }

  return res.data;
}

/** Delete a single review entry via Go API */
export async function deleteReviewAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const res = await apiClient.delete(`/api/tracking/review/${id}`);

    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple review entries via Go API */
export async function deleteReviewsAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const res = await apiClient.delete<{ count: number }>("/api/tracking/review", { ids });

    if (res.code !== 0) return { error: res.message };
    return { success: true, count: res.data.count };
  } catch {
    return { error: "删除失败" };
  }
}
