"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { groupApi } from "../actions/group.action";
import type { GroupDetail, GroupApplication } from "../types/group";

export function useGroupDetail(groupId: string) {
  const [group, setGroup] = useState<GroupDetail | null>(null);
  const [applications, setApplications] = useState<GroupApplication[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchDetail = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await groupApi.detail(groupId);
      if (res.code !== 0) {
        toast.error(res.message);
        return;
      }
      setGroup(res.data);
    } catch {
      toast.error("加载失败");
    } finally {
      setIsLoading(false);
    }
  }, [groupId]);

  const fetchApplications = useCallback(async () => {
    const res = await groupApi.listApplications(groupId);
    if (res.code === 0) {
      setApplications(res.data.items);
    }
  }, [groupId]);

  const handleApplication = useCallback(async (appId: string, action: "accept" | "reject") => {
    const res = await groupApi.handleApplication(groupId, appId, action);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success(action === "accept" ? "已通过" : "已拒绝");
    fetchApplications();
    fetchDetail();
  }, [groupId, fetchApplications, fetchDetail]);

  const updateGroup = useCallback(async (name: string, description?: string) => {
    const res = await groupApi.update(groupId, { name, description });
    if (res.code !== 0) {
      toast.error(res.message);
      return false;
    }
    toast.success("更新成功");
    fetchDetail();
    return true;
  }, [groupId, fetchDetail]);

  const deleteGroup = useCallback(async () => {
    const res = await groupApi.delete(groupId);
    if (res.code !== 0) {
      toast.error(res.message);
      return false;
    }
    return true;
  }, [groupId]);

  return {
    group, applications, isLoading,
    fetchDetail, fetchApplications, handleApplication,
    updateGroup, deleteGroup,
  };
}
