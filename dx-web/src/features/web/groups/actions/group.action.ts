import { apiClient } from "@/lib/api-client";
import type { CursorPaginated } from "@/lib/api-client";
import type { Group, GroupDetail, GroupApplication, GroupGameSearchItem } from "../types/group";

export const groupApi = {
  async list(params?: { tab?: string; cursor?: string; limit?: number }) {
    const query = new URLSearchParams();
    if (params?.tab) query.set("tab", params.tab);
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.limit) query.set("limit", String(params.limit));
    const qs = query.toString();
    return apiClient.get<CursorPaginated<Group>>(`/api/groups${qs ? `?${qs}` : ""}`);
  },
  async create(data: { name: string; description?: string }) {
    return apiClient.post<{ id: string; invite_code: string }>("/api/groups", data);
  },
  async detail(id: string) {
    return apiClient.get<GroupDetail>(`/api/groups/${id}`);
  },
  async update(id: string, data: { name: string; description?: string }) {
    return apiClient.put<null>(`/api/groups/${id}`, data);
  },
  async delete(id: string) {
    return apiClient.delete<null>(`/api/groups/${id}`);
  },
  async apply(id: string) {
    return apiClient.post<{ id: string }>(`/api/groups/${id}/apply`);
  },
  async cancelApply(id: string) {
    return apiClient.delete<null>(`/api/groups/${id}/apply`);
  },
  async listApplications(id: string, cursor?: string) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    const qs = params.toString();
    return apiClient.get<CursorPaginated<GroupApplication>>(`/api/groups/${id}/applications${qs ? `?${qs}` : ""}`);
  },
  async handleApplication(groupId: string, appId: string, action: "accept" | "reject") {
    return apiClient.put<null>(`/api/groups/${groupId}/applications/${appId}`, { action });
  },
  async searchGamesForGroup(groupId: string, q?: string, limit?: number) {
    const params = new URLSearchParams();
    if (q) params.set("q", q);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return apiClient.get<GroupGameSearchItem[]>(`/api/groups/${groupId}/games/search${qs ? `?${qs}` : ""}`);
  },
  async setGame(groupId: string, gameId: string, gameMode: string) {
    return apiClient.put<null>(`/api/groups/${groupId}/game`, { game_id: gameId, game_mode: gameMode });
  },
  async clearGame(groupId: string) {
    return apiClient.delete<null>(`/api/groups/${groupId}/game`);
  },
};
