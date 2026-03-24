"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { groupMemberApi } from "../actions/group-member.action";
import { groupSubgroupApi } from "../actions/group-subgroup.action";
import type { GroupMember, Subgroup, SubgroupMember } from "../types/group";

export function useGroupMembers(groupId: string) {
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [subgroups, setSubgroups] = useState<Subgroup[]>([]);
  const [selectedSubgroup, setSelectedSubgroup] = useState<string | null>(null);
  const [subgroupMembers, setSubgroupMembers] = useState<SubgroupMember[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchMembers = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await groupMemberApi.list(groupId);
      if (res.code === 0) {
        setMembers(res.data.items);
      }
    } finally {
      setIsLoading(false);
    }
  }, [groupId]);

  const fetchSubgroups = useCallback(async () => {
    const res = await groupSubgroupApi.list(groupId);
    if (res.code === 0) {
      setSubgroups(res.data);
    }
  }, [groupId]);

  const fetchSubgroupMembers = useCallback(async (subgroupId: string) => {
    setSelectedSubgroup(subgroupId);
    const res = await groupSubgroupApi.listMembers(groupId, subgroupId);
    if (res.code === 0) {
      setSubgroupMembers(res.data);
    }
  }, [groupId]);

  const kickMember = useCallback(async (userId: string) => {
    const res = await groupMemberApi.kick(groupId, userId);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("已移除");
    fetchMembers();
    fetchSubgroups();
  }, [groupId, fetchMembers, fetchSubgroups]);

  const leaveGroup = useCallback(async () => {
    const res = await groupMemberApi.leave(groupId);
    if (res.code !== 0) {
      toast.error(res.message);
      return false;
    }
    return true;
  }, [groupId]);

  const createSubgroup = useCallback(async (name: string) => {
    const res = await groupSubgroupApi.create(groupId, { name });
    if (res.code !== 0) {
      toast.error(res.message);
      return false;
    }
    toast.success("小组创建成功");
    fetchSubgroups();
    return true;
  }, [groupId, fetchSubgroups]);

  const deleteSubgroup = useCallback(async (subgroupId: string) => {
    const res = await groupSubgroupApi.delete(groupId, subgroupId);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("已删除小组");
    fetchSubgroups();
    if (selectedSubgroup === subgroupId) {
      setSelectedSubgroup(null);
      setSubgroupMembers([]);
    }
  }, [groupId, fetchSubgroups, selectedSubgroup]);

  const assignMembers = useCallback(async (subgroupId: string, userIds: string[]) => {
    const res = await groupSubgroupApi.assign(groupId, subgroupId, userIds);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("已分配成员");
    fetchSubgroups();
    if (selectedSubgroup === subgroupId) {
      fetchSubgroupMembers(subgroupId);
    }
  }, [groupId, fetchSubgroups, selectedSubgroup, fetchSubgroupMembers]);

  const removeSubgroupMember = useCallback(async (subgroupId: string, userId: string) => {
    const res = await groupSubgroupApi.removeMember(groupId, subgroupId, userId);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("已移除");
    fetchSubgroups();
    if (selectedSubgroup === subgroupId) {
      fetchSubgroupMembers(subgroupId);
    }
  }, [groupId, fetchSubgroups, selectedSubgroup, fetchSubgroupMembers]);

  return {
    members, subgroups, selectedSubgroup, subgroupMembers, isLoading,
    fetchMembers, fetchSubgroups, fetchSubgroupMembers,
    kickMember, leaveGroup,
    createSubgroup, deleteSubgroup,
    assignMembers, removeSubgroupMember,
  };
}
