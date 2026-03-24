import { apiClient } from "@/lib/api-client";
import type { Subgroup, SubgroupMember } from "../types/group";

export const groupSubgroupApi = {
  async list(groupId: string) {
    return apiClient.get<Subgroup[]>(`/api/groups/${groupId}/subgroups`);
  },
  async create(groupId: string, data: { name: string }) {
    return apiClient.post<{ id: string }>(`/api/groups/${groupId}/subgroups`, data);
  },
  async update(groupId: string, subgroupId: string, data: { name: string }) {
    return apiClient.put<null>(`/api/groups/${groupId}/subgroups/${subgroupId}`, data);
  },
  async delete(groupId: string, subgroupId: string) {
    return apiClient.delete<null>(`/api/groups/${groupId}/subgroups/${subgroupId}`);
  },
  async listMembers(groupId: string, subgroupId: string) {
    return apiClient.get<SubgroupMember[]>(`/api/groups/${groupId}/subgroups/${subgroupId}/members`);
  },
  async assign(groupId: string, subgroupId: string, userIds: string[]) {
    return apiClient.post<null>(`/api/groups/${groupId}/subgroups/${subgroupId}/members`, { user_ids: userIds });
  },
  async removeMember(groupId: string, subgroupId: string, userId: string) {
    return apiClient.delete<null>(`/api/groups/${groupId}/subgroups/${subgroupId}/members/${userId}`);
  },
};
