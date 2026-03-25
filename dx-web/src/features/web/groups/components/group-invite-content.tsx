"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, Users } from "lucide-react";
import { toast } from "sonner";
import { groupMemberApi } from "../actions/group-member.action";

interface GroupInfo {
  id: string;
  name: string;
  description: string | null;
  member_count: number;
  owner_name: string;
}

interface Props {
  code: string;
  isLoggedIn: boolean;
}

export function GroupInviteContent({ code, isLoggedIn }: Props) {
  const router = useRouter();
  const [group, setGroup] = useState<GroupInfo | null>(null);
  const [fetching, setFetching] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [applying, setApplying] = useState(false);

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

  async function handleApply() {
    setApplying(true);
    const res = await groupMemberApi.joinByCode(code);
    setApplying(false);

    if (res.code === 0) {
      toast.success("申请已提交，等待群主审核");
      router.push("/hall/groups");
      return;
    }
    // Already a member
    if (res.code === 40009) {
      toast.info("您已是该群成员");
      router.push(`/hall/groups/${group?.id ?? ""}`);
      return;
    }
    // Already applied
    if (res.code === 40010) {
      toast.info("您已提交过申请，请等待群主审核");
      router.push("/hall/groups");
      return;
    }
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
        {isLoggedIn ? (
          <button
            type="button"
            onClick={handleApply}
            disabled={applying}
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-teal-600 py-3 text-sm font-semibold text-white hover:bg-teal-700 disabled:opacity-60"
          >
            {applying && <Loader2 className="h-4 w-4 animate-spin" />}
            申请加入
          </button>
        ) : (
          <button
            type="button"
            onClick={() => router.push(`/auth/signin?redirect=/g/${code}`)}
            className="flex w-full items-center justify-center rounded-xl bg-teal-600 py-3 text-sm font-semibold text-white hover:bg-teal-700"
          >
            登录后加入
          </button>
        )}
      </div>
    </div>
  );
}
