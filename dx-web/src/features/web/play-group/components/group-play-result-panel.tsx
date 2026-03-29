"use client";

import { Crown, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import Link from "next/link";
import type {
  GroupLevelCompleteEvent,
  SoloWinner,
  TeamWinner,
} from "../types/group-play";

interface GroupPlayResultPanelProps {
  result: GroupLevelCompleteEvent;
  groupId: string;
}

/** Teal palette for podium columns */
const PODIUM_STYLES = [
  {
    // 1st place (center)
    colWidth: "w-[84px]",
    avatarSize: "h-[42px] w-[42px]",
    avatarBorder: "border-2 border-[#2dd4bf]",
    nameColor: "text-[#2dd4bf]",
    nameSize: "text-xs",
    barHeight: "h-[90px]",
    barGradient: "from-[#0d9488] to-[#0f766e]",
    rankSize: "text-2xl",
    rankColor: "text-[#2dd4bf]",
  },
  {
    // 2nd place (left)
    colWidth: "w-[76px]",
    avatarSize: "h-[34px] w-[34px]",
    avatarBorder: "border-2 border-[#14b8a6]",
    nameColor: "text-[#5eead4]",
    nameSize: "text-[10px]",
    barHeight: "h-[64px]",
    barGradient: "from-[#115e59] to-[#134e4a]",
    rankSize: "text-xl",
    rankColor: "text-[#5eead4]",
  },
  {
    // 3rd place (right)
    colWidth: "w-[76px]",
    avatarSize: "h-[34px] w-[34px]",
    avatarBorder: "border-2 border-[#0f766e]",
    nameColor: "text-[#99f6e4]",
    nameSize: "text-[10px]",
    barHeight: "h-[44px]",
    barGradient: "from-[#134e4a] to-[#1a3a38]",
    rankSize: "text-lg",
    rankColor: "text-[#99f6e4]",
  },
];

/** Render order: 2nd, 1st, 3rd */
const PODIUM_ORDER = [1, 0, 2];

interface PodiumEntry {
  id: string;
  name: string;
  score: number;
  rank: number;
}

function buildSoloPodium(participants: SoloWinner[]): PodiumEntry[] {
  return participants.map((p, i) => ({
    id: p.user_id,
    name: p.user_name,
    score: p.score,
    rank: i + 1,
  }));
}

function buildTeamPodium(teams: TeamWinner[]): PodiumEntry[] {
  return teams.map((t, i) => ({
    id: t.subgroup_id,
    name: t.subgroup_name,
    score: t.total_score,
    rank: i + 1,
  }));
}

export function GroupPlayResultPanel({
  result,
  groupId,
}: GroupPlayResultPanelProps) {
  const isSolo = result.mode === "group_solo";
  const entries = isSolo
    ? buildSoloPodium(result.participants)
    : buildTeamPodium(result.teams ?? []);

  const podiumEntries = entries.slice(0, 3);
  const restEntries = entries.slice(3);

  return (
    <div className="flex h-screen flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <h2 className="text-lg font-bold text-foreground">关卡结果</h2>

        {/* Podium */}
        <div className="flex items-end justify-center gap-1">
          {PODIUM_ORDER.map((rank) => {
            const entry = podiumEntries[rank];
            if (!entry) return null;
            const style = PODIUM_STYLES[rank];
            const avatarBg = isSolo
              ? getAvatarColor(entry.id)
              : undefined;

            return (
              <div
                key={entry.id}
                className={`flex flex-col items-center ${style.colWidth}`}
              >
                {rank === 0 && (
                  <Crown className="mb-0.5 h-[18px] w-[18px] text-amber-400" />
                )}
                <Avatar
                  className={`${style.avatarSize} ${style.avatarBorder}`}
                  style={
                    isSolo
                      ? { backgroundColor: avatarBg }
                      : { backgroundColor: "#0f766e" }
                  }
                >
                  <AvatarFallback
                    className="text-white font-bold text-sm"
                    style={
                      isSolo
                        ? { backgroundColor: avatarBg }
                        : { backgroundColor: "#0f766e" }
                    }
                  >
                    {entry.name[0]}
                  </AvatarFallback>
                </Avatar>
                <span
                  className={`mt-1 truncate text-center font-medium leading-tight ${style.nameColor} ${style.nameSize} max-w-full`}
                >
                  {entry.name}
                </span>
                <span className="text-xs font-bold text-foreground">
                  {entry.score} 分
                </span>
                <div
                  className={`mt-1.5 w-full rounded-t-md bg-gradient-to-b ${style.barGradient} ${style.barHeight} flex items-center justify-center`}
                >
                  <span
                    className={`font-bold ${style.rankSize} ${style.rankColor}`}
                  >
                    {entry.rank}
                  </span>
                </div>
              </div>
            );
          })}
        </div>

        {/* Remaining participants/teams */}
        {restEntries.length > 0 && (
          <>
            <div className="h-px w-full bg-border" />
            <div className="flex w-full flex-col gap-1.5">
              {restEntries.map((entry) => {
                const avatarBg = isSolo
                  ? getAvatarColor(entry.id)
                  : "#115e59";
                return (
                  <div
                    key={entry.id}
                    className="flex items-center gap-2 text-sm"
                  >
                    <Avatar size="sm" style={{ backgroundColor: avatarBg }}>
                      <AvatarFallback
                        className="text-white text-xs font-bold"
                        style={{ backgroundColor: avatarBg }}
                      >
                        {entry.name[0]}
                      </AvatarFallback>
                    </Avatar>
                    <span className="flex-1 truncate text-muted-foreground">
                      {entry.name}
                    </span>
                    <span className="font-medium text-[#5eead4]">
                      {entry.score} 分
                    </span>
                  </div>
                );
              })}
            </div>
          </>
        )}

        {/* All participant avatars */}
        <div className="h-px w-full bg-border" />
        <div className="flex w-full flex-col items-center gap-1.5">
          <span className="text-[10px] text-muted-foreground">
            全部参赛成员
          </span>
          <div className="flex flex-wrap justify-center gap-1">
            {result.participants.map((p) => {
              const bg = getAvatarColor(p.user_id);
              return (
                <Avatar
                  key={p.user_id}
                  size="sm"
                  style={{ backgroundColor: bg }}
                >
                  <AvatarFallback
                    className="text-white text-[10px] font-bold"
                    style={{ backgroundColor: bg }}
                  >
                    {p.user_name[0]}
                  </AvatarFallback>
                </Avatar>
              );
            })}
          </div>
        </div>

        {/* Return button */}
        <div className="h-px w-full bg-border" />
        <Button asChild className="w-full bg-teal-600 hover:bg-teal-700">
          <Link href={`/hall/groups/${groupId}`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回
          </Link>
        </Button>
      </div>
    </div>
  );
}
