"use client";

import { useRef, useMemo, useEffect } from "react";
import { useGroupEvents } from "@/hooks/use-group-events";
import type {
  GroupGameStartEvent,
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
  RoomMemberEvent,
} from "../types/group";

type GroupEventHandlers = {
  onGameStart?: (event: GroupGameStartEvent) => void;
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onRoomMemberJoined?: (event: RoomMemberEvent) => void;
  onRoomMemberLeft?: (event: RoomMemberEvent) => void;
  onDismissed?: () => void;
};

export function useGroupRoomEvents(
  groupId: string | null,
  handlers: GroupEventHandlers
) {
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  const listeners = useMemo(() => ({
    group_game_start: (data: unknown) =>
      handlersRef.current.onGameStart?.(data as GroupGameStartEvent),
    group_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    room_member_joined: (data: unknown) =>
      handlersRef.current.onRoomMemberJoined?.(data as RoomMemberEvent),
    room_member_left: (data: unknown) =>
      handlersRef.current.onRoomMemberLeft?.(data as RoomMemberEvent),
    group_dismissed: () => handlersRef.current.onDismissed?.(),
  }), []);

  useGroupEvents(groupId, listeners);
}
