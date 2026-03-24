"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import {
  ArrowLeft,
  ChevronRight,
  Link2,
  Copy,
  Loader2,
  Pencil,
  Trash2,
} from "lucide-react";
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
import { useGroupDetail } from "../hooks/use-group-detail";
import { useGroupMembers } from "../hooks/use-group-members";
import { MemberList } from "./member-list";
import { SubgroupList } from "./subgroup-list";
import { SubgroupMemberList } from "./subgroup-member-list";
import { ApplicationList } from "./application-list";
import { CreateSubgroupDialog } from "./create-subgroup-dialog";
import { EditGroupDialog } from "./edit-group-dialog";

interface GroupDetailContentProps {
  id: string;
}

function daysSince(dateStr: string) {
  const created = new Date(dateStr);
  const now = new Date();
  return Math.floor((now.getTime() - created.getTime()) / (1000 * 60 * 60 * 24));
}

export function GroupDetailContent({ id }: GroupDetailContentProps) {
  const router = useRouter();
  const {
    group, applications, isLoading,
    fetchDetail, fetchApplications, handleApplication,
    updateGroup, deleteGroup,
  } = useGroupDetail(id);
  const {
    members, subgroups, selectedSubgroup, subgroupMembers,
    fetchMembers, fetchSubgroups, fetchSubgroupMembers,
    kickMember, leaveGroup,
    createSubgroup, deleteSubgroup,
    removeSubgroupMember,
  } = useGroupMembers(id);

  const [createSubgroupOpen, setCreateSubgroupOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    fetchDetail();
    fetchMembers();
    fetchSubgroups();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (group?.is_owner) {
      fetchApplications();
    }
  }, [group?.is_owner]); // eslint-disable-line react-hooks/exhaustive-deps

  async function handleLeave() {
    const ok = await leaveGroup();
    if (ok) {
      toast.success("已退出群组");
      router.push("/hall/groups");
    }
  }

  async function handleDelete() {
    setDeleting(true);
    const ok = await deleteGroup();
    setDeleting(false);
    if (ok) {
      toast.success("群组已删除");
      setDeleteOpen(false);
      router.push("/hall/groups");
    }
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
    { value: String(subgroups.length), label: "小组" },
    { value: `${daysSince(group.created_at)}天`, label: "已创建" },
  ];

  return (
    <>
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link
            href="/hall/groups"
            aria-label="返回"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-border bg-card"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </Link>
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-muted-foreground">学习群</span>
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-sm font-semibold text-foreground">{group.name}</span>
          </div>
        </div>
        {isOwner && (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setEditOpen(true)}
              className="flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted"
            >
              <Pencil className="h-3.5 w-3.5" />
              编辑
            </button>
            <button
              type="button"
              onClick={() => setDeleteOpen(true)}
              className="flex items-center gap-1.5 rounded-lg border border-red-200 px-3 py-1.5 text-xs font-medium text-red-500 hover:bg-red-50"
            >
              <Trash2 className="h-3.5 w-3.5" />
              删除群组
            </button>
          </div>
        )}
      </div>

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
        </div>

        {/* Members list */}
        <MemberList
          groupId={id}
          isOwner={isOwner}
          members={members}
          onKick={kickMember}
          onLeave={handleLeave}
        />

        {/* Sub-groups */}
        <SubgroupList
          subgroups={subgroups}
          isOwner={isOwner}
          selectedId={selectedSubgroup}
          onSelect={fetchSubgroupMembers}
          onCreate={() => setCreateSubgroupOpen(true)}
          onDelete={deleteSubgroup}
        />

        {/* Sub-group members */}
        {selectedSubgroup && (
          <SubgroupMemberList
            members={subgroupMembers}
            isOwner={isOwner}
            onRemove={(userId) => removeSubgroupMember(selectedSubgroup, userId)}
          />
        )}

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
        onCreated={createSubgroup}
      />

      {/* Edit group dialog */}
      {isOwner && (
        <EditGroupDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          name={group.name}
          description={group.description}
          onSave={updateGroup}
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
