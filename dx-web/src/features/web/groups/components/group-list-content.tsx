"use client";

import { useEffect, useState } from "react";
import { Plus, Loader2 } from "lucide-react";
import { useGroups } from "../hooks/use-groups";
import { GroupCard } from "./group-card";
import { CreateGroupDialog } from "./create-group-dialog";

type Tab = "all" | "created" | "joined";

const tabs: { label: string; value: Tab }[] = [
  { label: "全部", value: "all" },
  { label: "我建的群", value: "created" },
  { label: "我加的群", value: "joined" },
];

export function GroupListContent() {
  const { groups, isLoading, tab, hasMore, fetchGroups, changeTab, loadMore } = useGroups();
  const [createOpen, setCreateOpen] = useState(false);

  useEffect(() => {
    fetchGroups();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
          onClick={() => setCreateOpen(true)}
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
            <GroupCard key={group.id} group={group} isMember />
          ))}
        </div>
      )}

      {/* Load more */}
      {hasMore && (
        <div className="flex justify-center">
          <button
            type="button"
            onClick={loadMore}
            disabled={isLoading}
            className="flex items-center gap-1.5 rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted-foreground hover:bg-muted disabled:opacity-50"
          >
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            加载更多
          </button>
        </div>
      )}

      {/* Create dialog */}
      <CreateGroupDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={() => fetchGroups()}
      />
    </>
  );
}
