"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, CheckCircle2, Users } from "lucide-react";
import { toast } from "sonner";
import { getAccessToken } from "@/lib/token";
import { groupMemberApi } from "../actions/group-member.action";

interface GroupInfo {
  id: string;
  name: string;
  description: string | null;
  member_count: number;
  owner_name: string;
}

type JoinState = "idle" | "loading" | "applied" | "already_applied";

interface Props {
  code: string;
}

export function GroupInviteContent({ code }: Props) {
  const router = useRouter();
  const [group, setGroup] = useState<GroupInfo | null>(null);
  const [fetching, setFetching] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [joinState, setJoinState] = useState<JoinState>("idle");
  const isLoggedIn = !!getAccessToken();

  useEffect(() => {
    async function loadGroup() {
      const res = await groupMemberApi.getGroupByInviteCode(code);
      setFetching(false);
      if (res.code !== 0) {
        setNotFound(true);
        return;
      }
      setGroup(res.data);
    }
    loadGroup();
  }, [code]);

  async function handleJoin() {
    setJoinState("loading");
    const res = await groupMemberApi.joinByCode(code);
    if (res.code === 0) {
      setJoinState("applied");
      return;
    }
    if (res.code === 40009) {
      // Already a member — redirect to group page
      router.push(`/hall/groups/${res.data?.group_id ?? ""}`);
      return;
    }
    if (res.code === 40010) {
      // Already applied
      setJoinState("already_applied");
      return;
    }
    setJoinState("idle");
    toast.error(res.message);
  }

  if (fetching) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
      </div>
    );
  }

  if (notFound || !group) {
    return (
      <div className="w-full max-w-md rounded-[14px] border border-border bg-card p-8 text-center">
        <p className="text-sm font-medium text-foreground">邀请链接无效</p>
        <p className="mt-1 text-xs text-muted-foreground">该群组邀请码不存在或已过期</p>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md rounded-[14px] border border-border bg-card p-8">
      {/* Avatar + name */}
      <div className="flex flex-col items-center gap-3">
        <div className="flex h-16 w-16 items-center justify-center rounded-[18px] bg-teal-100">
          <span className="text-[28px] font-bold text-teal-600">{group.name[0]}</span>
        </div>
        <div className="flex flex-col items-center gap-1">
          <h1 className="text-lg font-bold text-foreground">{group.name}</h1>
          <p className="text-xs text-muted-foreground">由 {group.owner_name} 创建</p>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Users className="h-3.5 w-3.5" />
            <span>{group.member_count} 名成员</span>
          </div>
        </div>
      </div>

      {/* Description */}
      {group.description && (
        <p className="mt-5 text-center text-[13px] leading-relaxed text-muted-foreground">
          {group.description}
        </p>
      )}

      {/* Action button */}
      <div className="mt-6">
        {!isLoggedIn ? (
          <button
            type="button"
            onClick={() => router.push(`/auth/signin?redirect=/g/${code}`)}
            className="flex w-full items-center justify-center rounded-xl bg-teal-600 py-3 text-sm font-semibold text-white hover:bg-teal-700"
          >
            登录后加入
          </button>
        ) : joinState === "applied" ? (
          <button
            type="button"
            disabled
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-teal-100 py-3 text-sm font-semibold text-teal-700"
          >
            <CheckCircle2 className="h-4 w-4" />
            申请已提交，等待群主审核
          </button>
        ) : joinState === "already_applied" ? (
          <button
            type="button"
            disabled
            className="flex w-full items-center justify-center rounded-xl bg-muted py-3 text-sm font-semibold text-muted-foreground"
          >
            申请审核中...
          </button>
        ) : (
          <button
            type="button"
            onClick={handleJoin}
            disabled={joinState === "loading"}
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-teal-600 py-3 text-sm font-semibold text-white hover:bg-teal-700 disabled:opacity-60"
          >
            {joinState === "loading" && <Loader2 className="h-4 w-4 animate-spin" />}
            加入群组
          </button>
        )}
      </div>
    </div>
  );
}
