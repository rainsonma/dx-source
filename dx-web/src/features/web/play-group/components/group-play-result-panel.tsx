"use client";

import { Trophy, Users, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import type {
  GroupLevelCompleteEvent,
  SoloWinner,
  TeamWinner,
} from "@/features/web/groups/types/group";

interface GroupPlayResultPanelProps {
  result: GroupLevelCompleteEvent;
  groupId: string;
}

export function GroupPlayResultPanel({ result, groupId }: GroupPlayResultPanelProps) {
  const isSolo = result.mode === "solo";
  const soloWinner = isSolo ? (result.winner as SoloWinner) : null;
  const teamWinner = !isSolo ? (result.winner as TeamWinner) : null;

  return (
    <div className="flex min-h-full flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-amber-100">
          <Trophy className="h-7 w-7 text-amber-500" />
        </div>

        <h2 className="text-lg font-bold text-foreground">关卡结果</h2>

        <div className="h-px w-full bg-border" />

        {soloWinner && (
          <div className="flex w-full flex-col items-center gap-3 rounded-xl bg-muted px-4 py-4">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-amber-500" />
              <span className="text-sm font-medium text-muted-foreground">冠军</span>
            </div>
            <span className="text-lg font-bold text-foreground">
              {soloWinner.user_name}
            </span>
            <span className="text-2xl font-bold text-teal-600">
              {soloWinner.score} 分
            </span>
          </div>
        )}

        {teamWinner && (
          <div className="flex w-full flex-col items-center gap-3 rounded-xl bg-muted px-4 py-4">
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-amber-500" />
              <span className="text-sm font-medium text-muted-foreground">冠军小组</span>
            </div>
            <span className="text-lg font-bold text-foreground">
              {teamWinner.subgroup_name}
            </span>
            <span className="text-2xl font-bold text-teal-600">
              {teamWinner.total_score} 分
            </span>
            <div className="h-px w-full bg-border" />
            <div className="w-full space-y-1">
              {teamWinner.members.map((m) => (
                <div
                  key={m.user_id}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="text-muted-foreground">{m.user_name}</span>
                  <span className="font-medium text-foreground">{m.score} 分</span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="h-px w-full bg-border" />

        <Button asChild className="w-full bg-teal-600 hover:bg-teal-700">
          <Link href={`/hall/groups/${groupId}`}>返回群组</Link>
        </Button>
      </div>
    </div>
  );
}
