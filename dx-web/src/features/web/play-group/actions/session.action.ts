import { apiClient } from "@/lib/api-client";

export async function startSessionAction(
  gameId: string,
  degree: string,
  pattern: string | null,
  levelId: string,
  gameGroupId: string
) {
  try {
    const res = await apiClient.post<any>("/api/play-group/start", {
      game_id: gameId,
      degree,
      pattern,
      level_id: levelId,
      game_group_id: gameGroupId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "无法开始游戏" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始游戏" };
  }
}

export async function startLevelAction(
  sessionId: string,
  gameLevelId: string,
  degree: string,
  pattern: string | null
) {
  try {
    const res = await apiClient.post<any>(
      `/api/play-group/${sessionId}/levels/start`,
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
    const res = await apiClient.post<any>(
      `/api/play-group/${sessionId}/levels/${gameLevelId}/complete`,
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
    const res = await apiClient.post<any>(
      `/api/play-group/${data.gameSessionTotalId}/answers`,
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
    const res = await apiClient.post<any>(
      `/api/play-group/${data.gameSessionTotalId}/skips`,
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

export async function syncPlayTimeAction(
  sessionId: string,
  playTime: number
) {
  try {
    const res = await apiClient.post<any>(
      `/api/play-group/${sessionId}/sync-playtime`,
      { play_time: playTime }
    );
    if (res.code !== 0) return { data: null, error: res.message || "同步时间失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "同步时间失败" };
  }
}

export async function restoreSessionDataAction(sessionId: string) {
  try {
    const res = await apiClient.get<any>(
      `/api/play-group/${sessionId}/restore`
    );
    if (res.code !== 0) return { data: null, error: res.message || "恢复会话数据失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "恢复会话数据失败" };
  }
}

export async function updateContentItemAction(
  sessionId: string,
  contentItemId: string | null
) {
  try {
    const res = await apiClient.put<any>(
      `/api/play-group/${sessionId}/content-item`,
      { content_item_id: contentItemId }
    );
    if (res.code !== 0) return { data: null, error: res.message || "更新进度失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "更新进度失败" };
  }
}

export async function endSessionAction(
  sessionId: string,
  data: {
    gameId: string;
    score: number;
    exp: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
    allLevelsCompleted: boolean;
  }
) {
  try {
    const res = await apiClient.post<any>(
      `/api/play-group/${sessionId}/end`,
      {
        game_id: data.gameId,
        score: data.score,
        exp: data.exp,
        max_combo: data.maxCombo,
        correct_count: data.correctCount,
        wrong_count: data.wrongCount,
        skip_count: data.skipCount,
        all_levels_completed: data.allLevelsCompleted,
      }
    );
    if (res.code !== 0) return { data: null, error: res.message || "结束会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束会话失败" };
  }
}

export async function restartLevelAction(sessionId: string, levelId: string) {
  try {
    const res = await apiClient.post<any>(
      `/api/play-group/${sessionId}/levels/${levelId}/restart`
    );
    if (res.code !== 0) return { data: null, error: res.message || "重置关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "重置关卡失败" };
  }
}

export async function markAsReviewAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  // Uses shared tracking API — not play-group specific
  try {
    await apiClient.post<any>("/api/tracking/review", {
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
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
    const res = await apiClient.get<any[]>(
      `/api/games/${gameId}/levels/${levelId}/content${params}`
    );
    if (res.code !== 0) return { data: null, error: res.message || "加载内容失败" };
    return { data: res.data ?? [], error: null };
  } catch {
    return { data: null, error: "加载内容失败" };
  }
}
