import { sessionApi } from "@/lib/api-client";

export async function startSessionAction(
  gameId: string,
  degree?: string,
  levelId?: string,
  pattern?: string,
  gameGroupId?: string
) {
  try {
    const res = await sessionApi.startSession({
      game_id: gameId,
      degree,
      level_id: levelId,
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
  gameId: string,
  degree: string,
  pattern: string | null
) {
  try {
    const res = await sessionApi.checkActive(gameId, degree, pattern);
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

export async function checkActiveLevelSessionAction(
  gameId: string,
  degree: string,
  pattern: string | null,
  gameLevelId: string
) {
  try {
    const res = await sessionApi.checkActiveLevel(gameId, degree, pattern, gameLevelId);
    if (res.code !== 0) return { data: null, error: res.message || "检查关卡会话失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "检查关卡会话失败" };
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
    const res = await sessionApi.recordAnswer(data.gameSessionTotalId, {
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
    });
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
    const res = await sessionApi.recordSkip(data.gameSessionTotalId, {
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

export async function startSessionLevelAction(
  sessionId: string,
  gameLevelId: string,
  degree: string,
  pattern?: string
) {
  try {
    const res = await sessionApi.startLevel(sessionId, {
      game_level_id: gameLevelId,
      degree,
      pattern,
    });
    if (res.code !== 0) return { data: null, error: res.message || "开始关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "开始关卡失败" };
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
    const res = await sessionApi.endSession(sessionId, {
      game_id: data.gameId,
      score: data.score,
      exp: data.exp,
      max_combo: data.maxCombo,
      correct_count: data.correctCount,
      wrong_count: data.wrongCount,
      skip_count: data.skipCount,
      all_levels_completed: data.allLevelsCompleted,
    });
    if (res.code !== 0) return { data: null, error: res.message || "结束游戏失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束游戏失败" };
  }
}

export async function advanceSessionLevelAction(
  sessionId: string,
  nextLevelId: string
) {
  try {
    // Use the current level as the route param; the next level is in the body
    const res = await sessionApi.advanceLevel(sessionId, nextLevelId, nextLevelId);
    if (res.code !== 0) return { data: null, error: res.message || "切换关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "切换关卡失败" };
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

export async function fetchSessionRestoreDataAction(
  sessionId: string,
  gameLevelId: string
) {
  try {
    const res = await sessionApi.restore(sessionId, gameLevelId);
    if (res.code !== 0) return { data: null, error: res.message || "恢复会话数据失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "恢复会话数据失败" };
  }
}

export async function syncPlayTimeAction(
  sessionId: string,
  gameLevelId: string,
  playTime: number
) {
  try {
    const res = await sessionApi.syncPlayTime(sessionId, {
      game_level_id: gameLevelId,
      play_time: playTime,
    });
    if (res.code !== 0) return { data: null, error: res.message || "同步时间失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "同步时间失败" };
  }
}

