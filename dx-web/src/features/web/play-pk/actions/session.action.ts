import { apiClient } from "@/lib/api-client";

export async function startPkAction(
  gameId: string,
  degree: string,
  pattern: string | null,
  levelId: string,
  difficulty: string
) {
  try {
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      opponent_id: string;
      opponent_name: string;
    }>("/api/play-pk/start", {
      game_id: gameId,
      degree,
      pattern,
      level_id: levelId,
      difficulty,
    });
    if (res.code !== 0) return { data: null, error: res.message || "无法开始PK" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始PK" };
  }
}

export async function startLevelAction(
  sessionId: string,
  gameLevelId: string,
  degree: string,
  pattern: string | null
) {
  try {
    const res = await apiClient.post<{ id: string; currentContentItemId?: string | null }>(
      `/api/play-pk/${sessionId}/levels/start`,
      { game_level_id: gameLevelId, degree, pattern }
    );
    if (res.code !== 0) return { data: null, error: res.message || "开始关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "开始关卡失败" };
  }
}

export async function completeLevelAction(
  sessionId: string,
  gameLevelId: string,
  data: { score: number; maxCombo: number; totalItems: number }
) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-pk/${sessionId}/levels/${gameLevelId}/complete`,
      { score: data.score, max_combo: data.maxCombo, total_items: data.totalItems }
    );
    if (res.code !== 0) return { data: null, error: res.message || "完成关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "完成关卡失败" };
  }
}

export async function recordAnswerAction(data: {
  gameSessionTotalId: string;
  gameSessionLevelId: string;
  gameLevelId: string;
  contentItemId: string;
  isCorrect: boolean;
  userAnswer: string;
  sourceAnswer: string;
  baseScore: number;
  comboScore: number;
  score: number;
  maxCombo: number;
  playTime: number;
  nextContentItemId: string | null;
  duration: number;
}) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-pk/${data.gameSessionTotalId}/answers`,
      {
        game_session_level_id: data.gameSessionLevelId,
        game_level_id: data.gameLevelId,
        content_item_id: data.contentItemId,
        is_correct: data.isCorrect,
        user_answer: data.userAnswer,
        source_answer: data.sourceAnswer,
        base_score: data.baseScore,
        combo_score: data.comboScore,
        score: data.score,
        max_combo: data.maxCombo,
        play_time: data.playTime,
        next_content_item_id: data.nextContentItemId,
        duration: data.duration,
      }
    );
    if (res.code !== 0) return { data: null, error: res.message || "记录失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "记录失败" };
  }
}

export async function recordSkipAction(data: {
  gameSessionTotalId: string;
  gameLevelId: string;
  playTime: number;
  nextContentItemId: string | null;
}) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-pk/${data.gameSessionTotalId}/skips`,
      {
        game_level_id: data.gameLevelId,
        play_time: data.playTime,
        next_content_item_id: data.nextContentItemId,
      }
    );
    if (res.code !== 0) return { data: null, error: res.message || "跳过失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "跳过失败" };
  }
}

type RestoreData = {
  sessionLevel?: {
    score: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
    playTime: number;
  };
};

export async function restoreSessionDataAction(sessionId: string) {
  try {
    const res = await apiClient.get<RestoreData>(
      `/api/play-pk/${sessionId}/restore`
    );
    if (res.code !== 0) return { data: null, error: res.message || "恢复会话数据失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "恢复会话数据失败" };
  }
}

export async function endPkAction(pkId: string) {
  try {
    const res = await apiClient.post<unknown>(`/api/play-pk/${pkId}/end`);
    if (res.code !== 0) return { data: null, error: res.message || "结束PK失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束PK失败" };
  }
}

export async function nextLevelAction(pkId: string, currentLevelId: string) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-pk/${pkId}/next-level`,
      { current_level_id: currentLevelId }
    );
    if (res.code !== 0) return { data: null, error: res.message || "下一关失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "下一关失败" };
  }
}

export async function pausePkAction(pkId: string) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${pkId}/pause`);
  } catch {
    // Fire-and-forget
  }
}

export async function resumePkAction(pkId: string) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${pkId}/resume`);
  } catch {
    // Fire-and-forget
  }
}

export async function fetchLevelContentAction(
  gameId: string,
  levelId: string,
  degree?: string
) {
  try {
    const params = degree ? `?degree=${degree}` : "";
    const res = await apiClient.get<Record<string, unknown>[]>(
      `/api/games/${gameId}/levels/${levelId}/content${params}`
    );
    if (res.code !== 0) return { data: null, error: res.message || "加载内容失败" };
    return { data: res.data ?? [], error: null };
  } catch {
    return { data: null, error: "加载内容失败" };
  }
}

export async function markAsReviewAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  // Uses shared tracking API — not play-pk specific
  try {
    await apiClient.post<unknown>("/api/tracking/review", {
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
  } catch {
    // Fire-and-forget
  }
}
