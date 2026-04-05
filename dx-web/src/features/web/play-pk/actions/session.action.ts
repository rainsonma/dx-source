import { apiClient } from "@/lib/api-client";

export async function startPkAction(
  gameId: string,
  gameLevelId: string,
  degree: string,
  pattern: string | null,
  difficulty: string
) {
  try {
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      game_level_id: string;
      opponent_id: string;
      opponent_name: string;
      robot_completed: boolean;
    }>("/api/play-pk/start", {
      game_id: gameId,
      game_level_id: gameLevelId,
      degree,
      pattern,
      difficulty,
    });
    if (res.code !== 0) return { data: null, error: res.message || "无法开始PK" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始PK" };
  }
}

export async function completeLevelAction(
  sessionId: string,
  gameLevelId: string,
  data: { score: number; maxCombo: number; totalItems: number }
) {
  try {
    const res = await apiClient.post<{
      next_level_id: string | null;
      next_level_name: string | null;
    }>(
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
  gameSessionId: string;
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
      `/api/play-pk/${data.gameSessionId}/answers`,
      {
        game_session_id: data.gameSessionId,
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

type RestoreData = {
  score: number;
  maxCombo: number;
  correctCount: number;
  wrongCount: number;
  playTime: number;
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

export async function nextPkLevelAction(pkId: string) {
  try {
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      game_level_id: string;
      opponent_id: string;
      opponent_name: string;
      robot_completed: boolean;
    }>(`/api/play-pk/${pkId}/next-level`);
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
