"use client";

import { useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useUserSSE } from "@/hooks/use-user-sse";
import { acceptPkInviteAction, declinePkInviteAction } from "../actions/invite.action";
import { PkInvitationPopup } from "./pk-invitation-popup";

interface PkInvitation {
  pk_id: string;
  game_id: string;
  game_name: string;
  game_mode: string;
  level_name: string;
  initiator_id: string;
  initiator_name: string;
}

export function PkInvitationProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [invitation, setInvitation] = useState<PkInvitation | null>(null);

  useUserSSE({
    pk_invitation: (data) => {
      const inv = data as PkInvitation;
      setInvitation(inv);
    },
  });

  const handleAccept = useCallback(async () => {
    if (!invitation) return;
    const result = await acceptPkInviteAction(invitation.pk_id);
    setInvitation(null);
    if (result.data) {
      const params = new URLSearchParams({
        degree: result.data.degree,
        level: result.data.game_level_id,
        pkId: invitation.pk_id,
        sessionId: result.data.session_id,
      });
      if (result.data.pattern) params.set("pattern", result.data.pattern);
      router.push(`/hall/pk-room/${invitation.pk_id}?${params}`);
    }
  }, [invitation, router]);

  const handleDecline = useCallback(async () => {
    if (!invitation) return;
    await declinePkInviteAction(invitation.pk_id);
    setInvitation(null);
  }, [invitation]);

  return (
    <>
      {children}
      {invitation && (
        <PkInvitationPopup
          pkId={invitation.pk_id}
          gameName={invitation.game_name}
          levelName={invitation.level_name}
          initiatorName={invitation.initiator_name}
          onAccept={handleAccept}
          onDecline={handleDecline}
        />
      )}
    </>
  );
}
