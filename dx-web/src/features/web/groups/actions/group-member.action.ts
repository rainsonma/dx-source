import { apiClient } from "@/lib/api-client";
import type { CursorPaginated } from "@/lib/api-client";
import type { GroupMember } from "../types/group";

export const groupMemberApi = {
  async list(groupId: string, cursor?: string) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    const qs = params.toString();
    return apiClient.get<CursorPaginated<GroupMember>>(`/api/groups/${groupId}/members${qs ? `?${qs}` : ""}`);
  },
  async kick(groupId: string, userId: string) {
    return apiClient.delete<null>(`/api/groups/${groupId}/members/${userId}`);
  },
  async leave(groupId: string) {
    return apiClient.post<null>(`/api/groups/${groupId}/leave`);
  },
  async joinByCode(code: string) {
    return apiClient.post<{ group_id: string }>(`/api/groups/join/${code}`);
  },
};
