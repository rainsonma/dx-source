"use client";

import { Trophy, Users, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { GroupLevelCompleteEvent, SoloWinner, TeamWinner } from "@/features/web/groups/types/group";

interface GroupResultPanelProps {
  result: GroupLevelCompleteEvent;
  hasMoreLevels: boolean;
  onNextLevel: () => void;
  onFinish: () => void;
}

export function GroupResultPanel({ result, hasMoreLevels, onNextLevel, onFinish }: GroupResultPanelProps) {
  const isSolo = result.mode === "solo";
  const soloWinner = isSolo ? (result.winner as SoloWinner) : null;
  const teamWinner = !isSolo ? (result.winner as TeamWinner) : null;

  return (
    <div className="flex flex-col items-center justify-center h-full gap-6 px-4">
      <div className="flex flex-col items-center gap-3">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-amber-100">
          <Trophy className="h-8 w-8 text-amber-500" />
        </div>
        <h2 className="text-xl font-bold">关卡结果</h2>
      </div>

      {soloWinner && (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-border bg-card p-6 w-full max-w-xs">
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-medium text-muted-foreground">冠军</span>
          </div>
          <span className="text-lg font-bold">{soloWinner.user_name}</span>
          <span className="text-2xl font-bold text-teal-600">{soloWinner.score} 分</span>
        </div>
      )}

      {teamWinner && (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-border bg-card p-6 w-full max-w-xs">
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-medium text-muted-foreground">冠军小组</span>
          </div>
          <span className="text-lg font-bold">{teamWinner.subgroup_name}</span>
          <span className="text-2xl font-bold text-teal-600">{teamWinner.total_score} 分</span>
          <div className="w-full space-y-1 pt-2 border-t border-border">
            {teamWinner.members.map((m) => (
              <div key={m.user_id} className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">{m.user_name}</span>
                <span className="font-medium">{m.score} 分</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex gap-3">
        {hasMoreLevels ? (
          <Button onClick={onNextLevel}>下一关</Button>
        ) : (
          <Button onClick={onFinish}>查看最终结果</Button>
        )}
      </div>
    </div>
  );
}
