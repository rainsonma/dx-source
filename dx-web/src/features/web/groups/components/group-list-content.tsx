"use client";

import { useState, useEffect } from "react";
import useSWR from "swr";
import { apiClient } from "@/lib/api-client";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
import { Plus, Loader2 } from "lucide-react";
import { toast } from "sonner";
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
import type { Group } from "../types/group";
import { GroupCard } from "./group-card";
import { CreateGroupDialog } from "./create-group-dialog";
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog";

type Tab = "all" | "created" | "joined";

interface GroupListResponse {
  items: Group[];
  nextCursor: string;
  hasMore: boolean;
}

const tabs: { label: string; value: Tab }[] = [
  { label: "全部", value: "all" },
  { label: "我建的群", value: "created" },
  { label: "我加的群", value: "joined" },
];

export function GroupListContent() {
  const [tab, setTab] = useState<Tab>("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [applyTarget, setApplyTarget] = useState<Group | null>(null);
  const [applying, setApplying] = useState(false);
  const [isVip, setIsVip] = useState(false);
  const [upgradeOpen, setUpgradeOpen] = useState(false);

  useEffect(() => {
    apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile").then((res) => {
      if (res.code === 0 && res.data) {
        setIsVip(isVipActive(res.data.grade as UserGrade, res.data.vip_due_at));
      }
    });
  }, []);

  const swrKey = `/api/groups?tab=${tab}`;
  const { data, isLoading } = useSWR<GroupListResponse>(swrKey);

  const groups = data?.items ?? [];

  function changeTab(newTab: Tab) {
    setTab(newTab);
  }

  async function handleApplyConfirm() {
    if (!applyTarget) return;
    setApplying(true);
    const res = await groupApi.apply(applyTarget.id);
    setApplying(false);
    if (res.code !== 0) {
      toast.error(res.message);
      setApplyTarget(null);
      return;
    }
    toast.success("申请已提交，等待群主审核");
    setApplyTarget(null);
    await swrMutate("/api/groups");
  }

  return (
    <>
      {/* Tab row */}
      <div className="flex w-full flex-col gap-3 border-b border-border md:flex-row md:items-center md:gap-0">
        <div className="flex items-center">
          {tabs.map((t) => (
            <button
              key={t.value}
              type="button"
              onClick={() => changeTab(t.value)}
              className={`px-5 py-2 text-sm font-medium ${
                tab === t.value
                  ? "border-b-2 border-teal-600 text-teal-600"
                  : "text-muted-foreground"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <div className="hidden flex-1 md:block" />
        <button
          type="button"
          onClick={() => isVip ? setCreateOpen(true) : setUpgradeOpen(true)}
          className="flex w-full items-center justify-center gap-1.5 rounded-lg bg-teal-600 px-3.5 py-2 text-sm font-medium text-white hover:bg-teal-700 md:w-auto"
        >
          <Plus className="h-4 w-4" />
          创建学习群
        </button>
      </div>

      {/* Loading state */}
      {isLoading && groups.length === 0 && (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
        </div>
      )}

      {/* Empty state */}
      {!isLoading && groups.length === 0 && (
        <div className="flex flex-col items-center justify-center gap-2 py-20 text-muted-foreground">
          <p className="text-sm">暂无群组</p>
          <p className="text-xs">点击上方按钮创建一个学习群吧</p>
        </div>
      )}

      {/* Card grid */}
      {groups.length > 0 && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-3">
          {groups.map((group) => (
            <GroupCard
              key={group.id}
              group={group}
              isMember={group.is_member}
              isVip={isVip}
              onJoin={() => isVip ? setApplyTarget(group) : setUpgradeOpen(true)}
              onUpgrade={() => setUpgradeOpen(true)}
            />
          ))}
        </div>
      )}

      {/* Create dialog */}
      <CreateGroupDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={() => swrMutate("/api/groups")}
      />

      <UpgradeDialog
        open={upgradeOpen}
        onOpenChange={setUpgradeOpen}
        title="会员专属功能"
        message="升级会员即可创建和加入学习群，与同学一起学习"
      />

      {/* Apply confirm dialog */}
      <AlertDialog open={!!applyTarget} onOpenChange={(open) => { if (!open) setApplyTarget(null); }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>申请加入群组</AlertDialogTitle>
            <AlertDialogDescription>
              确认申请加入「{applyTarget?.name}」群？需等待群主确认！
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={applying}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleApplyConfirm} disabled={applying} className="bg-teal-600 hover:bg-teal-700">
              {applying && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认申请
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
