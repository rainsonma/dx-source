import { sessionApi } from "@/lib/api-client";

export async function startSessionAction(
  gameId: string,
  gameLevelId: string,
  degree?: string,
  pattern?: string,
  gameGroupId?: string
) {
  try {
    const res = await sessionApi.startSession({
      game_id: gameId,
      game_level_id: gameLevelId,
      degree,
      pattern,
      ...(gameGroupId && { game_group_id: gameGroupId }),
    });
    if (res.code !== 0) return { data: null, error: res.message || "无法开始游戏" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始游戏" };
  }
}

export async function checkActiveSessionAction(
  gameLevelId: string,
  degree: string,
  pattern: string | null
) {
  try {
    const res = await sessionApi.checkActive(gameLevelId, degree, pattern);
    if (res.code !== 0) return { data: null, error: res.message || "检查会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "检查会话失败" };
  }
}

export async function checkAnyActiveSessionAction(gameId: string) {
  try {
    const res = await sessionApi.checkAnyActive(gameId);
    if (res.code !== 0) return { data: null, error: res.message || "检查会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "检查会话失败" };
  }
}

export async function forceCompleteSessionAction(sessionId: string) {
  try {
    const res = await sessionApi.forceComplete(sessionId);
    if (res.code !== 0) return { data: null, error: res.message || "结束会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束会话失败" };
  }
}

export async function restartLevelSessionAction(
  sessionId: string,
  gameLevelId: string
) {
  try {
    const res = await sessionApi.restartLevel(sessionId, gameLevelId);
    if (res.code !== 0) return { data: null, error: res.message || "重启关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "重启关卡失败" };
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
    const res = await sessionApi.recordAnswer(data.gameSessionId, {
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
    });
    if (res.code !== 0) return { data: null, error: res.message || "记录失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "记录失败" };
  }
}

export async function recordSkipAction(data: {
  gameSessionId: string;
  gameLevelId: string;
  playTime: number;
  nextContentItemId: string | null;
}) {
  try {
    const res = await sessionApi.recordSkip(data.gameSessionId, {
      game_session_id: data.gameSessionId,
      game_level_id: data.gameLevelId,
      play_time: data.playTime,
      next_content_item_id: data.nextContentItemId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "跳过失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "跳过失败" };
  }
}

export async function completeLevelAction(
  sessionId: string,
  gameLevelId: string,
  data: {
    score: number;
    maxCombo: number;
    totalItems: number;
  }
) {
  try {
    const res = await sessionApi.completeLevel(sessionId, gameLevelId, {
      score: data.score,
      max_combo: data.maxCombo,
      total_items: data.totalItems,
    });
    if (res.code !== 0) return { data: null, error: res.message || "完成关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "完成关卡失败" };
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
    const res = await sessionApi.endSession(sessionId, {
      score: data.score,
      exp: data.exp,
      max_combo: data.maxCombo,
      correct_count: data.correctCount,
      wrong_count: data.wrongCount,
      skip_count: data.skipCount,
    });
    if (res.code !== 0) return { data: null, error: res.message || "结束游戏失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束游戏失败" };
  }
}

export async function updateSessionContentItemAction(
  sessionId: string,
  contentItemId: string | null
) {
  try {
    const res = await sessionApi.updateContentItem(sessionId, contentItemId);
    if (res.code !== 0) return { data: null, error: res.message || "更新进度失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "更新进度失败" };
  }
}

export async function fetchSessionRestoreDataAction(sessionId: string) {
  try {
    const res = await sessionApi.restore(sessionId);
    if (res.code !== 0) return { data: null, error: res.message || "恢复会话数据失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "恢复会话数据失败" };
  }
}

export async function syncPlayTimeAction(
  sessionId: string,
  playTime: number
) {
  try {
    const res = await sessionApi.syncPlayTime(sessionId, {
      play_time: playTime,
    });
    if (res.code !== 0) return { data: null, error: res.message || "同步时间失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "同步时间失败" };
  }
}
