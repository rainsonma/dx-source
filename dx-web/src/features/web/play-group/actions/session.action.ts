import { apiClient } from "@/lib/api-client";

export async function startSessionAction(
  gameId: string,
  gameLevelId: string,
  degree: string,
  pattern: string | null,
  gameGroupId: string
) {
  try {
    const body: Record<string, string> = {
      game_id: gameId,
      game_level_id: gameLevelId,
      degree,
      game_group_id: gameGroupId,
    };
    if (pattern) body.pattern = pattern;
    const res = await apiClient.post<{ id: string; currentContentItemId?: string | null }>("/api/play-group/start", body);
    if (res.code !== 0) return { data: null, error: res.message || "无法开始游戏" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始游戏" };
  }
}

export async function completeLevelAction(
  sessionId: string,
  gameLevelId: string,
  data: { score: number; maxCombo: number; totalItems: number }
) {
  try {
    const res = await apiClient.post<{ nextLevelId?: string; nextLevelName?: string }>(
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
      `/api/play-group/${data.gameSessionId}/answers`,
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

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export async function recordSkipAction(data: {
  gameSessionId: string;
  gameLevelId: string;
  playTime: number;
  nextContentItemId: string | null;
}) {
  // Skip is disabled for group play — no-op stub to satisfy GamePlayActions interface
  return { data: null, error: null };
}

export async function syncPlayTimeAction(
  sessionId: string,
  playTime: number
) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-group/${sessionId}/sync-playtime`,
      { play_time: playTime }
    );
    if (res.code !== 0) return { data: null, error: res.message || "同步时间失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "同步时间失败" };
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
    const res = await apiClient.put<unknown>(
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
    score: number;
    exp: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
  }
) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-group/${sessionId}/end`,
      {
        score: data.score,
        exp: data.exp,
        max_combo: data.maxCombo,
        correct_count: data.correctCount,
        wrong_count: data.wrongCount,
        skip_count: data.skipCount,
      }
    );
    if (res.code !== 0) return { data: null, error: res.message || "结束会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束会话失败" };
  }
}

export async function nextGroupLevelAction(groupId: string) {
  try {
    const res = await apiClient.post<null>(`/api/groups/${groupId}/next-level`);
    if (res.code !== 0) return { data: null, error: res.message || "进入下一关失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "进入下一关失败" };
  }
}

export async function restartLevelAction(sessionId: string, levelId: string) {
  try {
    const res = await apiClient.post<unknown>(
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
    await apiClient.post<unknown>("/api/tracking/review", {
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
    const res = await apiClient.get<Record<string, unknown>[]>(
      `/api/games/${gameId}/levels/${levelId}/content${params}`
    );
    if (res.code !== 0) return { data: null, error: res.message || "加载内容失败" };
    return { data: res.data ?? [], error: null };
  } catch {
    return { data: null, error: "加载内容失败" };
  }
}
