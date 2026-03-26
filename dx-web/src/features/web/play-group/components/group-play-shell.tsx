"use client";

import { useEffect, useRef } from "react";
import { GAME_MODES } from "@/consts/game-mode";
import { useGroupPlayStore } from "../hooks/use-group-play-store";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { GroupPlayLoadingScreen } from "./group-play-loading-screen";
import { GroupPlayTopBar } from "./group-play-top-bar";
import { GroupPlayWaitingScreen } from "./group-play-waiting-screen";
import { GroupPlayResultPanel } from "./group-play-result-panel";
import { GameSettingsModal } from "@/features/web/play/components/game-settings-modal";
import { GameResetModal } from "@/features/web/play/components/game-reset-modal";
import { GameReportModal } from "@/features/web/play/components/game-report-modal";
import { GameExitModal } from "@/features/web/play/components/game-exit-modal";
import { LsrwGame } from "@/features/web/play/components/lsrw-game";
import { VocabMatchGame } from "@/features/web/play/components/vocab-match-game";
import { ListeningGame } from "@/features/web/play/components/listening-game";
import { VocabEliminationGame } from "@/features/web/play/components/vocab-elimination-game";
import { VocabBattleGame } from "@/features/web/play/components/vocab-battle-game";
import { useGroupEvents } from "@/features/web/groups/hooks/use-group-events";
import { completeLevelAction } from "../actions/session.action";
import { useFullscreen } from "@/features/web/play/hooks/use-fullscreen";
import { getToken } from "@/lib/api-client";
import type { ComponentType } from "react";

const modeComponents: Record<string, ComponentType> = {
  [GAME_MODES.LSRW]: LsrwGame,
  [GAME_MODES.VOCAB_MATCH]: VocabMatchGame,
  [GAME_MODES.LISTENING_CHALLENGE]: ListeningGame,
  [GAME_MODES.VOCAB_ELIMINATION]: VocabEliminationGame,
  [GAME_MODES.VOCAB_BATTLE]: VocabBattleGame,
};

interface GroupPlayShellProps {
  game: {
    id: string;
    name: string;
    mode: string;
    levels: { id: string; name: string; order: number }[];
  };
  player: { nickname: string; avatarUrl: string | null };
  degree: string;
  pattern: string | null;
  levelId: string;
  groupId: string;
  levelTimeLimit: number;
}

export function GroupPlayShell({
  game,
  player,
  degree,
  pattern,
  levelId,
  groupId,
  levelTimeLimit,
}: GroupPlayShellProps) {
  // Phase and overlay managed via useGameStore so shared modals work
  const phase = useGameStore((s) => s.phase);
  const overlay = useGameStore((s) => s.overlay);
  const showOverlay = useGameStore((s) => s.showOverlay);
  const setPhase = useGameStore((s) => s.setPhase);
  const storeGameId = useGameStore((s) => s.gameId);
  const storeLevelId = useGameStore((s) => s.levelId);
  const exitGame = useGameStore((s) => s.exitGame);

  // Group-specific state from useGroupPlayStore
  const groupPhase = useGroupPlayStore((s) => s.groupPhase);
  const groupResult = useGroupPlayStore((s) => s.groupResult);
  const setGroupWaiting = useGroupPlayStore((s) => s.setGroupWaiting);
  const setGroupResult = useGroupPlayStore((s) => s.setGroupResult);
  const sessionId = useGroupPlayStore((s) => s.sessionId);
  const score = useGroupPlayStore((s) => s.score);
  const combo = useGroupPlayStore((s) => s.combo);
  const contentItems = useGroupPlayStore((s) => s.contentItems);

  const completedRef = useRef(false);

  const targetLevel =
    game.levels.find((l) => l.id === levelId) ?? game.levels[0];
  const targetLevelId = targetLevel?.id ?? levelId;
  const levelName = targetLevel?.name ?? game.name;

  const { isFullscreen, toggleFullscreen } = useFullscreen();

  async function completeAndWait() {
    if (completedRef.current || !sessionId || !targetLevelId) return;
    completedRef.current = true;
    setGroupWaiting();
    await completeLevelAction(sessionId, targetLevelId, {
      score,
      maxCombo: combo.maxCombo,
      totalItems: contentItems?.length ?? 0,
    });
  }

  // SSE: listen for group level complete result
  useGroupEvents(groupId, {
    onLevelComplete: (event) => {
      setGroupResult(event);
    },
  });

  useEffect(() => {
    const isDifferentGame = storeGameId !== null && storeGameId !== game.id;
    const isDifferentLevel =
      storeLevelId !== null && storeLevelId !== targetLevelId;
    const isStaleState = storeGameId === game.id && phase !== "loading";

    if (isDifferentGame || isDifferentLevel || isStaleState) {
      exitGame();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [game.id, targetLevelId]);

  // Flush playtime to server on tab close / navigation
  useEffect(() => {
    const handleBeforeUnload = () => {
      const sid = useGroupPlayStore.getState().sessionId;
      if (!sid) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
      const token = getToken();
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (token) headers["Authorization"] = `Bearer ${token}`;

      fetch(`${apiUrl}/api/group-play/sessions/${sid}/sync-playtime`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          play_time: useGroupPlayStore.getState().playTime,
        }),
        keepalive: true,
      });
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, []);

  if (phase === "loading") {
    return (
      <GroupPlayLoadingScreen
        gameId={game.id}
        gameName={game.name}
        gameMode={game.mode}
        degree={degree}
        pattern={pattern}
        levelId={targetLevelId}
        levelName={levelName}
        gameGroupId={groupId}
        levelTimeLimit={levelTimeLimit}
      />
    );
  }

  if (phase === "result") {
    if (groupPhase !== "result") {
      completeAndWait();
      return <GroupPlayWaitingScreen groupId={groupId} />;
    }
    return <GroupPlayResultPanel result={groupResult!} groupId={groupId} />;
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <div className="flex h-screen w-full flex-col bg-muted">
      <GroupPlayTopBar
        player={player}
        levelName={levelName}
        levelTimeLimit={levelTimeLimit}
        onExit={() => showOverlay("exit")}
        onReset={() => showOverlay("reset")}
        onSettings={() => showOverlay("settings")}
        onReport={() => showOverlay("report")}
        onFullscreen={toggleFullscreen}
        isFullscreen={isFullscreen}
        onLevelTimeUp={() => {
          setPhase("result");
          completeAndWait();
        }}
      />
      <div className="flex flex-1 flex-col items-center justify-center gap-6 overflow-y-auto px-4 py-10">
        <GameComponent />
      </div>

      {overlay === "settings" && <GameSettingsModal />}
      {overlay === "reset" && <GameResetModal />}
      {overlay === "report" && <GameReportModal />}
      {overlay === "exit" && <GameExitModal gameId={game.id} />}
    </div>
  );
}
