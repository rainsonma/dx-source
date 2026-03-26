import { apiClient } from "@/lib/api-client";

/** Submit a content feedback report */
export async function submitReportAction(data: {
  gameId: string;
  gameLevelId: string;
  contentItemId: string;
  reason: string;
  note?: string;
}) {
  try {
    const res = await apiClient.post<{ id: string; count: number }>("/api/reports", {
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
      content_item_id: data.contentItemId,
      reason: data.reason,
      note: data.note,
    });

    if (res.code !== 0) {
      return { data: null, error: res.message };
    }

    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "提交反馈失败" };
  }
}
