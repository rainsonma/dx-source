import { apiClient } from "@/lib/api-client";

/** Parse a value that may be a JSON string into an object, or return as-is */
function parseJsonField(value: unknown): unknown {
  if (typeof value === "string") {
    try {
      return JSON.parse(value);
    } catch {
      return value;
    }
  }
  return value ?? null;
}

/** Map Go API flat ContentItemData to the nested shape expected by the game store */
function toContentItem(item: Record<string, unknown>) {
  return {
    id: item.id,
    content: item.content,
    contentType: item.contentType,
    translation: item.translation ?? null,
    definition: item.definition ?? null,
    explanation: item.explanation ?? null,
    items: parseJsonField(item.items),
    structure: parseJsonField(item.structure),
    ukAudio: item.ukAudioUrl ? { url: item.ukAudioUrl } : null,
    usAudio: item.usAudioUrl ? { url: item.usAudioUrl } : null,
  };
}

/** Fetch content items for a game level from the Go API */
export async function fetchLevelContentAction(
  gameId: string,
  gameLevelId: string,
  degree?: string
) {
  try {
    const params = degree ? `?degree=${degree}` : "";
    const res = await apiClient.get<Record<string, unknown>[]>(
      `/api/games/${gameId}/levels/${gameLevelId}/content${params}`
    );

    if (res.code !== 0) {
      return { data: null, error: res.message };
    }

    return { data: (res.data ?? []).map(toContentItem), error: null };
  } catch {
    return { data: null, error: "加载内容失败" };
  }
}
