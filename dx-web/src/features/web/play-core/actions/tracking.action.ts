import { trackingApi } from "@/lib/api-client";

export async function markAsMasteredAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  try {
    const res = await trackingApi.markMastered({
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "标记掌握失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "标记掌握失败" };
  }
}

export async function markAsUnknownAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  try {
    const res = await trackingApi.markUnknown({
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "标记生词失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "标记生词失败" };
  }
}

export async function markAsReviewAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  try {
    const res = await trackingApi.markReview({
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "标记复习失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "标记复习失败" };
  }
}
