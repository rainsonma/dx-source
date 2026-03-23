import { apiClient } from "@/lib/api-client";

export type GameSearchResult = {
  id: string;
  name: string;
  mode: string;
  category: { name: string } | null;
};

export type GameSearchActionResult = {
  games: GameSearchResult[];
  error?: string;
};

/** Map Go API flat search/recent result to the nested GameSearchResult shape */
function toGameSearchResult(item: any): GameSearchResult {
  return {
    id: item.id,
    name: item.name,
    mode: item.mode,
    category: item.categoryName ? { name: item.categoryName } : null,
  };
}

/** Search published games by name */
export async function searchGamesAction(
  query: string
): Promise<GameSearchActionResult> {
  try {
    const trimmed = query.trim();
    if (!trimmed) return { games: [] };

    const params = new URLSearchParams({
      q: trimmed,
      limit: "8",
    });
    const res = await apiClient.get<any[]>(`/api/games/search?${params}`);

    if (res.code !== 0) {
      return { games: [], error: res.message };
    }

    return { games: (res.data ?? []).map(toGameSearchResult) };
  } catch {
    return { games: [], error: "搜索失败，请重试" };
  }
}

/** Get current user's recently played games */
export async function getRecentGamesAction(): Promise<GameSearchActionResult> {
  try {
    const res = await apiClient.get<any[]>("/api/games/recent");

    if (res.code !== 0) {
      return { games: [], error: res.message };
    }

    return { games: (res.data ?? []).map(toGameSearchResult) };
  } catch {
    return { games: [], error: "加载失败，请重试" };
  }
}
