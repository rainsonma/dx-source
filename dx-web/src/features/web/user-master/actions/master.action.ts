import { apiClient } from "@/lib/api-client";

/** Content item shape from Go API */
type ContentItem = {
  content: string;
  translation: string | null;
  contentType: string;
};

/** Master item shape from Go API TrackingItemData */
export type MasterItem = {
  id: string;
  contentItem: ContentItem;
  gameName: string;
  masteredAt: string | null;
  createdAt: string;
};

/** Master stats shape from Go API MasterStatsData */
export type MasterStats = {
  total: number;
  thisWeek: number;
  thisMonth: number;
};

/** Fetch next page of mastered words via Go API */
export async function fetchMastersAction(cursor?: string): Promise<{
  items: MasterItem[];
  nextCursor: string | null;
}> {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  const qs = params.toString();

  const res = await apiClient.get<{
    items: MasterItem[];
    nextCursor: string;
    hasMore: boolean;
  }>(`/api/tracking/master${qs ? `?${qs}` : ""}`);

  if (res.code !== 0) {
    return { items: [], nextCursor: null };
  }

  return {
    items: res.data.items ?? [],
    nextCursor: res.data.hasMore ? res.data.nextCursor : null,
  };
}

/** Fetch master stats via Go API */
export async function fetchMasterStatsAction(): Promise<MasterStats> {
  const res = await apiClient.get<MasterStats>("/api/tracking/master/stats");

  if (res.code !== 0) {
    return { total: 0, thisWeek: 0, thisMonth: 0 };
  }

  return res.data;
}

/** Delete a single master entry via Go API */
export async function deleteMasterAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const res = await apiClient.delete(`/api/tracking/master/${id}`);

    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple master entries via Go API */
export async function deleteMastersAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const res = await apiClient.delete<{ count: number }>("/api/tracking/master", { ids });

    if (res.code !== 0) return { error: res.message };
    return { success: true, count: res.data.count };
  } catch {
    return { error: "删除失败" };
  }
}
