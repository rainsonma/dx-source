import { apiClient } from "@/lib/api-client";

export async function verifyOpponentAction(username: string) {
  try {
    const res = await apiClient.post<{
      user_id: string;
      nickname: string;
      is_online: boolean;
      is_vip: boolean;
    }>("/api/users/verify-online", { username });
    if (res.code !== 0) return { data: null, error: res.message || "验证失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "验证失败" };
  }
}

export async function invitePkAction(data: {
  gameId: string;
  gameLevelId: string;
  degree: string;
  pattern: string | null;
  opponentId: string;
}) {
  try {
    const body: Record<string, string> = {
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
      degree: data.degree,
      opponent_id: data.opponentId,
    };
    if (data.pattern) body.pattern = data.pattern;
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      game_level_id: string;
    }>("/api/play-pk/invite", body);
    if (res.code !== 0) return { data: null, error: res.message || "邀请失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "邀请失败" };
  }
}

export async function acceptPkInviteAction(pkId: string) {
  try {
    const res = await apiClient.post<{
      session_id: string;
      game_id: string;
      game_level_id: string;
      degree: string;
      pattern: string | null;
    }>(`/api/play-pk/invite/${pkId}/accept`);
    if (res.code !== 0) return { data: null, error: res.message || "接受失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "接受失败" };
  }
}

export async function declinePkInviteAction(pkId: string) {
  try {
    const res = await apiClient.post<unknown>(`/api/play-pk/invite/${pkId}/decline`);
    if (res.code !== 0) return { data: null, error: res.message || "拒绝失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "拒绝失败" };
  }
}

export async function fetchPkDetailsAction(pkId: string) {
  try {
    const res = await apiClient.get<{
      pk_id: string;
      session_id: string;
      game_id: string;
      game_name: string;
      game_mode: string;
      level_id: string;
      level_name: string;
      degree: string;
      pattern: string | null;
      initiator_id: string;
      initiator_name: string;
      opponent_id: string;
      opponent_name: string;
      invitation_status: string;
    }>(`/api/play-pk/${pkId}/details`);
    if (res.code !== 0) return { data: null, error: res.message || "获取详情失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "获取详情失败" };
  }
}
