import { favoriteApi } from "@/lib/api-client";

export type ToggleFavoriteResult =
  | { success: true; favorited: boolean }
  | { error: string };

/** Toggle a game's favorite state via Go API */
export async function toggleFavoriteAction(
  gameId: string
): Promise<ToggleFavoriteResult> {
  try {
    const res = await favoriteApi.toggle(gameId);
    if (res.code !== 0) {
      return { error: res.message || "操作失败" };
    }
    return { success: true, favorited: res.data.favorited };
  } catch {
    return { error: "操作失败" };
  }
}
