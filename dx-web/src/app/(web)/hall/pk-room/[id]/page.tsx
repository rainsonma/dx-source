"use client";

import { useEffect, useState, useCallback, use } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Swords, Loader2, ArrowLeft, User } from "lucide-react";
import { usePkSSE } from "@/hooks/use-pk-sse";
import { fetchPkDetailsAction } from "@/features/web/play-pk/actions/invite.action";
import { endPkAction } from "@/features/web/play-pk/actions/session.action";
import { getAvatarColor } from "@/lib/avatar";

type PkDetails = {
  pk_id: string;
  game_id: string;
  game_name: string;
  game_mode: string;
  level_id: string;
  level_name: string;
  degree: string;
  pattern: string | null;
  initiator_id: string;
  initiator_name: string;
  opponent_id: string;
  opponent_name: string;
  invitation_status: string;
};

export default function PkRoomPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id: pkId } = use(params);
  const router = useRouter();
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("sessionId") ?? "";

  const [details, setDetails] = useState<PkDetails | null>(null);
  const [status, setStatus] = useState<
    "waiting" | "accepted" | "declined" | "timeout"
  >("waiting");
  const [timeLeft, setTimeLeft] = useState(30);
  const [countdown, setCountdown] = useState<number | null>(null);

  // Fetch PK details
  useEffect(() => {
    fetchPkDetailsAction(pkId).then((res) => {
      if (res.data) {
        setDetails(res.data);
        if (res.data.invitation_status === "accepted") {
          setStatus("accepted");
          setCountdown(1);
        }
      }
    });
  }, [pkId]);

  // 30s waiting timeout
  useEffect(() => {
    if (status !== "waiting") return;
    if (timeLeft <= 0) {
      const timer = setTimeout(() => {
        setStatus("timeout");
        endPkAction(pkId);
      }, 0);
      return () => clearTimeout(timer);
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, status, pkId]);

  // Countdown after accepted → navigate to play
  useEffect(() => {
    if (countdown === null || !details) return;
    if (countdown <= 0) {
      const p = new URLSearchParams({
        degree: details.degree,
        level: details.level_id,
        pkId: details.pk_id,
        sessionId,
      });
      if (details.pattern) p.set("pattern", details.pattern);
      router.push(`/hall/play-pk/${details.game_id}?${p}`);
      return;
    }
    const timer = setTimeout(() => setCountdown((c) => (c ?? 1) - 1), 1000);
    return () => clearTimeout(timer);
  }, [countdown, details, sessionId, router]);

  // SSE listeners
  usePkSSE(status === "waiting" ? pkId : null, {
    pk_invitation_accepted: (data: unknown) => {
      const event = data as { opponent_name: string };
      setStatus("accepted");
      if (details) {
        setDetails({
          ...details,
          opponent_name: event.opponent_name,
          invitation_status: "accepted",
        });
      }
      setCountdown(1);
    },
    pk_invitation_declined: () => {
      setStatus("declined");
    },
  });

  const handleCancel = useCallback(async () => {
    await endPkAction(pkId);
    router.back();
  }, [pkId, router]);

  const handleBack = useCallback(() => {
    router.back();
  }, [router]);

  if (!details) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
      </div>
    );
  }

  const initiatorColor = getAvatarColor(details.initiator_name);
  const opponentColor = getAvatarColor(details.opponent_name);

  return (
    <div className="flex h-screen w-full flex-col items-center justify-center gap-8 bg-[radial-gradient(ellipse_at_center,#1E1B4B,#0F0A2E)] px-4">
      <div className="flex flex-col items-center gap-2">
        <Swords className="h-8 w-8 text-teal-400" />
        <h1 className="text-2xl font-bold text-white">{details.game_name}</h1>
        <p className="text-sm text-slate-400">{details.level_name} · PK 对战</p>
      </div>

      <div className="flex items-center gap-8">
        <div className="flex flex-col items-center gap-3">
          <div
            className="flex h-16 w-16 items-center justify-center rounded-full text-xl font-bold text-white"
            style={{ backgroundColor: initiatorColor }}
          >
            {details.initiator_name[0]}
          </div>
          <span className="text-sm font-medium text-white">
            {details.initiator_name}
          </span>
        </div>

        <span className="text-2xl font-bold text-slate-500">VS</span>

        <div className="flex flex-col items-center gap-3">
          {status === "accepted" ? (
            <div
              className="flex h-16 w-16 items-center justify-center rounded-full text-xl font-bold text-white"
              style={{ backgroundColor: opponentColor }}
            >
              {details.opponent_name[0]}
            </div>
          ) : (
            <div className="flex h-16 w-16 items-center justify-center rounded-full border-2 border-dashed border-slate-600">
              <User className="h-6 w-6 text-slate-600" />
            </div>
          )}
          <span className="text-sm font-medium text-slate-400">
            {status === "accepted" ? details.opponent_name : "等待对手..."}
          </span>
        </div>
      </div>

      {status === "waiting" && (
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="h-5 w-5 animate-spin text-teal-400" />
          <p className="text-sm text-slate-400">
            等待对方接受邀请... {timeLeft}s
          </p>
          <button
            type="button"
            onClick={handleCancel}
            className="mt-2 flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            取消
          </button>
        </div>
      )}

      {status === "accepted" && countdown !== null && (
        <p className="text-lg font-bold text-teal-400">
          {countdown > 0 ? `${countdown}s 后开始...` : "开始!"}
        </p>
      )}

      {status === "declined" && (
        <div className="flex flex-col items-center gap-3">
          <p className="text-sm font-medium text-red-400">对方已拒绝</p>
          <button
            type="button"
            onClick={handleBack}
            className="flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </button>
        </div>
      )}

      {status === "timeout" && (
        <div className="flex flex-col items-center gap-3">
          <p className="text-sm font-medium text-amber-400">对方未响应</p>
          <button
            type="button"
            onClick={handleBack}
            className="flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </button>
        </div>
      )}
    </div>
  );
}
