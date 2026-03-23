import { apiClient } from "@/lib/api-client";

/** Content item shape from Go API */
type ContentItem = {
  content: string;
  translation: string | null;
  contentType: string;
};

/** Unknown item shape from Go API TrackingItemData */
export type UnknownItem = {
  id: string;
  contentItem: ContentItem;
  gameName: string;
  createdAt: string;
};

/** Unknown stats shape from Go API */
export type UnknownStats = {
  total: number;
  today: number;
  lastThreeDays: number;
};

/** Fetch next page of unknown words via Go API */
export async function fetchUnknownsAction(cursor?: string): Promise<{
  items: UnknownItem[];
  nextCursor: string | null;
}> {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  const qs = params.toString();

  const res = await apiClient.get<{
    items: UnknownItem[];
    nextCursor: string;
    hasMore: boolean;
  }>(`/api/tracking/unknown${qs ? `?${qs}` : ""}`);

  if (res.code !== 0) {
    return { items: [], nextCursor: null };
  }

  return {
    items: res.data.items ?? [],
    nextCursor: res.data.hasMore ? res.data.nextCursor : null,
  };
}

/** Fetch unknown stats via Go API */
export async function fetchUnknownStatsAction(): Promise<UnknownStats> {
  const res = await apiClient.get<UnknownStats>("/api/tracking/unknown/stats");

  if (res.code !== 0) {
    return { total: 0, today: 0, lastThreeDays: 0 };
  }

  return res.data;
}

/** Delete a single unknown entry via Go API */
export async function deleteUnknownAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const res = await apiClient.delete(`/api/tracking/unknown/${id}`);

    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple unknown entries via Go API */
export async function deleteUnknownsAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const res = await apiClient.delete<{ count: number }>("/api/tracking/unknown", { ids });

    if (res.code !== 0) return { error: res.message };
    return { success: true, count: res.data.count };
  } catch {
    return { error: "删除失败" };
  }
}
