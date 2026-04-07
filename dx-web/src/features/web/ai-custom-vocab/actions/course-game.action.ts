import { apiClient } from "@/lib/api-client";

export type LevelContentItem = {
  id: string;
  content: string;
  contentType: string;
  translation: string | null;
  items: unknown;
  order: number;
};

export type LevelMeta = {
  id: string;
  sourceData: string;
  translation: string | null;
  sourceFrom: string;
  sourceType: string;
  isBreakDone: boolean;
  isItemDone: boolean;
  order: number;
  itemCount: number;
};

export type CreateCourseGameResult = {
  success?: boolean;
  gameId?: string;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

/** Create a new course game via Go API. */
export async function createCourseGameAction(
  _prev: CreateCourseGameResult,
  formData: FormData
): Promise<CreateCourseGameResult> {
  try {
    const body = {
      gameCategoryId: formData.get("gameCategoryId") as string,
      gamePressId: formData.get("gamePressId") as string,
      gameMode: formData.get("gameMode") as string,
      name: formData.get("name") as string,
      description: (formData.get("description") as string) || undefined,
      coverId: (formData.get("coverId") as string) || undefined,
    };

    const res = await apiClient.post<{ id: string }>("/api/course-games", body);

    if (res.code !== 0) {
      return { error: res.message };
    }

    return { success: true, gameId: res.data.id };
  } catch {
    return { error: "创建失败" };
  }
}

export type GameLevelActionResult = {
  success?: boolean;
  levelId?: string;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

/** Create a game level via Go API. */
export async function createGameLevelAction(
  gameId: string,
  _prev: GameLevelActionResult,
  formData: FormData
): Promise<GameLevelActionResult> {
  try {
    const body = {
      name: formData.get("name") as string,
      description: (formData.get("description") as string) || undefined,
    };

    const res = await apiClient.post<{ id: string }>(
      `/api/course-games/${gameId}/levels`,
      body
    );

    if (res.code !== 0) {
      return { error: res.message };
    }

    return { success: true, levelId: res.data.id };
  } catch {
    return { error: "创建关卡失败" };
  }
}

export type SimpleActionResult = {
  success?: boolean;
  error?: string;
};

/** Delete a game level via Go API. */
export async function deleteGameLevelAction(
  gameId: string,
  levelId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(
      `/api/course-games/${gameId}/levels/${levelId}`
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除关卡失败" };
  }
}

/** Delete a course game via Go API. */
export async function deleteGameAction(
  gameId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(`/api/course-games/${gameId}`);
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除游戏失败" };
  }
}

/** Publish a course game via Go API. */
export async function publishGameAction(
  gameId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.post<null>(`/api/course-games/${gameId}/publish`);
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "发布失败" };
  }
}

/** Withdraw a published game via Go API. */
export async function withdrawGameAction(
  gameId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.post<null>(`/api/course-games/${gameId}/withdraw`);
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "撤回失败" };
  }
}

export type UpdateGameResult = {
  success?: boolean;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

/** Update a course game via Go API. */
export async function updateCourseGameAction(
  gameId: string,
  _prev: UpdateGameResult,
  formData: FormData
): Promise<UpdateGameResult> {
  try {
    const body = {
      gameCategoryId: formData.get("gameCategoryId") as string,
      gamePressId: formData.get("gamePressId") as string,
      gameMode: formData.get("gameMode") as string,
      name: formData.get("name") as string,
      description: (formData.get("description") as string) || undefined,
      coverId: (formData.get("coverId") as string) || undefined,
    };

    const res = await apiClient.put<null>(`/api/course-games/${gameId}`, body);

    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "更新失败" };
  }
}

export type SaveMetadataResult = {
  success?: boolean;
  count?: number;
  error?: string;
};

/** Save content metadata batch via Go API. */
export async function saveMetadataAction(
  gameId: string,
  data: { gameLevelId: string; entries: unknown[]; sourceFrom: string }
): Promise<SaveMetadataResult> {
  try {
    const res = await apiClient.post<{ count: number }>(
      `/api/course-games/${gameId}/levels/${data.gameLevelId}/metadata`,
      data
    );

    if (res.code !== 0) return { error: res.message };
    return { success: true, count: res.data.count };
  } catch {
    return { error: "保存失败" };
  }
}

/** Reorder content metadata via Go API. */
export async function reorderMetaAction(
  gameId: string,
  metaId: string,
  newOrder: number
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.put<null>(
      `/api/course-games/${gameId}/metadata/reorder`,
      { metaId, newOrder }
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "排序失败" };
  }
}

export type FetchContentItemsResult = {
  items: LevelContentItem[];
  error?: string;
};

/** Fetch content items grouped by metadata via Go API. */
export async function fetchContentItemsAction(
  gameId: string,
  levelId: string
): Promise<FetchContentItemsResult> {
  try {
    const res = await apiClient.get<LevelContentItem[]>(
      `/api/course-games/${gameId}/levels/${levelId}/content-items`
    );

    if (res.code !== 0) return { items: [], error: res.message };
    return { items: res.data ?? [] };
  } catch {
    return { items: [], error: "加载失败" };
  }
}

/** Reorder a content item via Go API. */
export async function reorderItemAction(
  gameId: string,
  itemId: string,
  newOrder: number
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.put<null>(
      `/api/course-games/${gameId}/content-items/reorder`,
      { itemId, newOrder }
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "排序失败" };
  }
}

/** Update content item text via Go API. */
export async function updateContentItemTextAction(
  gameId: string,
  itemId: string,
  content: string,
  translation: string | null
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.put<null>(
      `/api/course-games/${gameId}/content-items/${itemId}`,
      { content, translation }
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "更新失败" };
  }
}

export type InsertContentItemResult = {
  item?: LevelContentItem;
  error?: string;
};

/** Insert a content item via Go API. */
export async function insertContentItemAction(
  gameId: string,
  data: {
    gameLevelId: string;
    contentMetaId: string;
    content: string;
    contentType: string;
    translation: string | null;
    referenceItemId: string;
    direction: "above" | "below";
  }
): Promise<InsertContentItemResult> {
  try {
    const res = await apiClient.post<LevelContentItem>(
      `/api/course-games/${gameId}/levels/${data.gameLevelId}/content-items`,
      data
    );

    if (res.code !== 0) return { error: res.message };
    return { item: res.data };
  } catch {
    return { error: "插入失败" };
  }
}

/** Delete a content item via Go API. */
export async function deleteContentItemAction(
  gameId: string,
  itemId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(
      `/api/course-games/${gameId}/content-items/${itemId}`
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete all content from a level via Go API. */
export async function deleteAllLevelContentAction(
  gameId: string,
  levelId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(
      `/api/course-games/${gameId}/levels/${levelId}/content-items`
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "清空失败" };
  }
}
