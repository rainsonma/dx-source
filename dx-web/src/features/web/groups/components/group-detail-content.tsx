"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import useSWR from "swr";
import { toast } from "sonner";
import {
  Link2,
  Copy,
  Loader2,
  Pencil,
  Trash2,
} from "lucide-react";
import { BreadcrumbTopBar } from "@/features/web/hall/components/breadcrumb-top-bar";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { swrMutate } from "@/lib/swr";
import { groupApi } from "../actions/group.action";
import { groupMemberApi } from "../actions/group-member.action";
import { groupSubgroupApi } from "../actions/group-subgroup.action";
import type { GroupDetail, GroupMember, Subgroup, SubgroupMember, GroupApplication } from "../types/group";
import { MemberList } from "./member-list";
import { SubgroupList } from "./subgroup-list";
import { SubgroupMemberList } from "./subgroup-member-list";
import { ApplicationList } from "./application-list";
import { CreateSubgroupDialog } from "./create-subgroup-dialog";
import { EditGroupDialog } from "./edit-group-dialog";

interface GroupDetailContentProps {
  id: string;
}

interface MembersResponse {
  items: GroupMember[];
  nextCursor: string;
  hasMore: boolean;
}

interface ApplicationsResponse {
  items: GroupApplication[];
  nextCursor: string;
  hasMore: boolean;
}

function daysSince(dateStr: string) {
  const created = new Date(dateStr);
  const now = new Date();
  return Math.floor((now.getTime() - created.getTime()) / (1000 * 60 * 60 * 24));
}

export function GroupDetailContent({ id }: GroupDetailContentProps) {
  const router = useRouter();

  // SWR data fetching
  const { data: group, isLoading } = useSWR<GroupDetail>(`/api/groups/${id}`);
  const { data: membersData } = useSWR<MembersResponse>(`/api/groups/${id}/members`);
  const { data: subgroups } = useSWR<Subgroup[]>(`/api/groups/${id}/subgroups`);
  const { data: appsData } = useSWR<ApplicationsResponse>(
    group?.is_owner ? `/api/groups/${id}/applications` : null
  );

  const members = membersData?.items ?? [];
  const subgroupList = subgroups ?? [];
  const applications = appsData?.items ?? [];

  // Local UI state
  const [selectedSubgroup, setSelectedSubgroup] = useState<string | null>(null);
  const [subgroupMembers, setSubgroupMembers] = useState<SubgroupMember[]>([]);
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set());
  const [createSubgroupOpen, setCreateSubgroupOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  function invalidateAll() {
    swrMutate(`/api/groups/${id}`, "/api/groups");
  }

  function toggleSelectUser(userId: string) {
    setSelectedUserIds((prev) => {
      const next = new Set(prev);
      if (next.has(userId)) next.delete(userId);
      else next.add(userId);
      return next;
    });
  }

  async function handleAssignToSubgroup(subgroupId: string) {
    if (selectedUserIds.size === 0) return;
    const res = await groupSubgroupApi.assign(id, subgroupId, Array.from(selectedUserIds));
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("已分配到小组");
    setSelectedUserIds(new Set());
    await swrMutate(`/api/groups/${id}`);
    if (selectedSubgroup === subgroupId) fetchSubgroupMembers(subgroupId);
  }

  async function fetchSubgroupMembers(subgroupId: string) {
    setSelectedSubgroup(subgroupId);
    const res = await groupSubgroupApi.listMembers(id, subgroupId);
    if (res.code === 0) setSubgroupMembers(res.data);
  }

  async function handleKick(userId: string) {
    const res = await groupMemberApi.kick(id, userId);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("已移除");
    invalidateAll();
  }

  async function handleLeave() {
    const res = await groupMemberApi.leave(id);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("已退出群组");
    router.push("/hall/groups");
  }

  async function handleDelete() {
    setDeleting(true);
    const res = await groupApi.delete(id);
    setDeleting(false);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("群组已删除");
    setDeleteOpen(false);
    await swrMutate("/api/groups");
    router.push("/hall/groups");
  }

  async function handleUpdateGroup(name: string, description?: string) {
    const res = await groupApi.update(id, { name, description });
    if (res.code !== 0) { toast.error(res.message); return false; }
    toast.success("更新成功");
    invalidateAll();
    return true;
  }

  async function handleCreateSubgroup(name: string) {
    const res = await groupSubgroupApi.create(id, { name });
    if (res.code !== 0) { toast.error(res.message); return false; }
    toast.success("小组创建成功");
    await swrMutate(`/api/groups/${id}`);
    return true;
  }

  async function handleDeleteSubgroup(subgroupId: string) {
    const res = await groupSubgroupApi.delete(id, subgroupId);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("已删除小组");
    await swrMutate(`/api/groups/${id}`);
    if (selectedSubgroup === subgroupId) {
      setSelectedSubgroup(null);
      setSubgroupMembers([]);
    }
  }

  async function handleRemoveSubgroupMember(subgroupId: string, userId: string) {
    const res = await groupSubgroupApi.removeMember(id, subgroupId, userId);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("已移除");
    await swrMutate(`/api/groups/${id}`);
    if (selectedSubgroup === subgroupId) fetchSubgroupMembers(subgroupId);
  }

  async function handleApplication(appId: string, action: "accept" | "reject") {
    const res = await groupApi.handleApplication(id, appId, action);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success(action === "accept" ? "已通过" : "已拒绝");
    invalidateAll();
  }

  async function handleCopyInvite() {
    if (!group) return;
    const link = `${window.location.origin}/hall/groups/join/${group.invite_code}`;
    await navigator.clipboard.writeText(link);
    toast.success("已复制邀请链接");
  }

  if (isLoading && !group) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-20 text-muted-foreground">
        <p className="text-sm">群组不存在或已被删除</p>
        <Link href="/hall/groups" className="text-sm text-teal-600 hover:underline">返回群组列表</Link>
      </div>
    );
  }

  const isOwner = group.is_owner;
  const stats = [
    { value: String(group.member_count), label: "成员" },
    { value: String(subgroupList.length), label: "小组" },
    { value: `${daysSince(group.created_at)}天`, label: "已创建" },
  ];

  return (
    <>
      {/* Top bar */}
      <BreadcrumbTopBar
        backHref="/hall/groups"
        items={[
          { label: "学习群", href: "/hall/groups", maxChars: 10 },
          { label: group.name, maxChars: 20 },
        ]}
      />

      {/* Multi-column layout */}
      <div className="grid flex-1 grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Left: Group info */}
        <div className="flex w-full flex-col gap-4 overflow-y-auto rounded-[14px] border border-border bg-card p-4">
          <div className="flex items-center gap-3.5">
            <div className="flex h-[52px] w-[52px] shrink-0 items-center justify-center rounded-[14px] bg-teal-100">
              <span className="text-[22px] font-bold text-teal-600">{group.name[0]}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-lg font-bold text-foreground">{group.name}</span>
              <span className="text-xs text-muted-foreground">由 {group.owner_name} 创建</span>
            </div>
          </div>

          {group.description && (
            <p className="text-[13px] leading-relaxed text-muted-foreground">{group.description}</p>
          )}

          <div className="flex gap-2.5">
            {stats.map((stat) => (
              <div key={stat.label} className="flex flex-1 flex-col items-center gap-0.5 rounded-[10px] bg-muted py-2.5">
                <span className="text-lg font-extrabold text-teal-600">{stat.value}</span>
                <span className="text-[10px] text-muted-foreground">{stat.label}</span>
              </div>
            ))}
          </div>

          <div className="h-px bg-border" />

          {/* Invite link */}
          <div className="flex flex-col gap-1.5 px-1">
            <div className="flex items-center gap-1.5">
              <Link2 className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-[11px] font-semibold text-muted-foreground">邀请链接</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-border bg-muted px-2.5 py-2">
              <span className="flex-1 truncate text-[11px] text-muted-foreground">
                {typeof window !== "undefined"
                  ? `${window.location.origin}/hall/groups/join/${group.invite_code}`
                  : `/hall/groups/join/${group.invite_code}`}
              </span>
              <button type="button" onClick={handleCopyInvite}>
                <Copy className="h-3.5 w-3.5 text-muted-foreground hover:text-foreground" />
              </button>
            </div>
          </div>

          {/* Owner actions */}
          {isOwner && (
            <>
              <div className="h-px bg-border" />
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setEditOpen(true)}
                  className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-border py-2 text-xs font-medium text-muted-foreground hover:bg-muted"
                >
                  <Pencil className="h-3.5 w-3.5" />
                  编辑
                </button>
                <button
                  type="button"
                  onClick={() => setDeleteOpen(true)}
                  className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-red-200 py-2 text-xs font-medium text-red-500 hover:bg-red-50"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  删除群组
                </button>
              </div>
            </>
          )}
        </div>

        {/* Members list */}
        <MemberList
          isOwner={isOwner}
          members={members}
          subgroups={subgroupList}
          selectedUserIds={selectedUserIds}
          onToggleSelect={toggleSelectUser}
          onAssignToSubgroup={handleAssignToSubgroup}
          onKick={handleKick}
          onLeave={handleLeave}
        />

        {/* Sub-groups */}
        <SubgroupList
          subgroups={subgroupList}
          isOwner={isOwner}
          selectedId={selectedSubgroup}
          onSelect={fetchSubgroupMembers}
          onCreate={() => setCreateSubgroupOpen(true)}
          onDelete={handleDeleteSubgroup}
        />

        {/* Sub-group members */}
        <SubgroupMemberList
          members={subgroupMembers}
          isOwner={isOwner}
          onRemove={selectedSubgroup
            ? (userId) => handleRemoveSubgroupMember(selectedSubgroup, userId)
            : undefined}
          emptyText={selectedSubgroup ? "暂无组成员" : "请选择一个小组查看成员"}
        />

        {/* Applications (owner only) */}
        {isOwner && applications.length > 0 && (
          <ApplicationList
            applications={applications}
            onAccept={(appId) => handleApplication(appId, "accept")}
            onReject={(appId) => handleApplication(appId, "reject")}
          />
        )}
      </div>

      {/* Create subgroup dialog */}
      <CreateSubgroupDialog
        open={createSubgroupOpen}
        onOpenChange={setCreateSubgroupOpen}
        onCreated={handleCreateSubgroup}
      />

      {/* Edit group dialog */}
      {isOwner && (
        <EditGroupDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          name={group.name}
          description={group.description}
          onSave={handleUpdateGroup}
        />
      )}

      {/* Delete confirmation */}
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除群组</AlertDialogTitle>
            <AlertDialogDescription>
              删除后所有成员和小组数据将被清除，此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
