"use client";

import { useEffect, useRef, useMemo } from "react";
import { GAME_MODES } from "@/consts/game-mode";
import { useGroupPlayStore } from "../hooks/use-group-play-store";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { GamePlayProvider, type GamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { GroupPlayLoadingScreen } from "./group-play-loading-screen";
import { GroupPlayTopBar } from "./group-play-top-bar";
import { GroupPlayWaitingScreen } from "./group-play-waiting-screen";
import { GroupPlayResultPanel } from "./group-play-result-panel";
import { GameSettingsModal } from "@/features/web/play-core/components/game-settings-modal";
import { GameResetModal } from "@/features/web/play-core/components/game-reset-modal";
import { GameReportModal } from "@/features/web/play-core/components/game-report-modal";
import { GameExitModal } from "@/features/web/play-core/components/game-exit-modal";
import { GameLsrw } from "@/features/web/play-core/components/game-lsrw";
import { GameVocabMatch } from "@/features/web/play-core/components/game-vocab-match";
import { GameListening } from "@/features/web/play-core/components/game-listening";
import { GameVocabElimination } from "@/features/web/play-core/components/game-vocab-elimination";
import { GameVocabBattle } from "@/features/web/play-core/components/game-vocab-battle";
import { useGroupPlayEvents } from "../hooks/use-group-play-events";
import {
  completeLevelAction,
  recordAnswerAction,
  recordSkipAction,
  markAsReviewAction,
  endSessionAction,
  restartLevelAction,
} from "../actions/session.action";
import { useFullscreen } from "@/features/web/play-core/hooks/use-fullscreen";
import { getToken } from "@/lib/api-client";
import type { ComponentType } from "react";

const modeComponents: Record<string, ComponentType> = {
  [GAME_MODES.LSRW]: GameLsrw,
  [GAME_MODES.VOCAB_MATCH]: GameVocabMatch,
  [GAME_MODES.LISTENING_CHALLENGE]: GameListening,
  [GAME_MODES.VOCAB_ELIMINATION]: GameVocabElimination,
  [GAME_MODES.VOCAB_BATTLE]: GameVocabBattle,
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

  // Score/combo are updated by shared game components via useGameStore
  const score = useGameStore((s) => s.score);
  const combo = useGameStore((s) => s.combo);
  const contentItems = useGameStore((s) => s.contentItems);

  const completedRef = useRef(false);

  const playActions = useMemo<GamePlayActions>(() => ({
    recordAnswer: recordAnswerAction,
    recordSkip: recordSkipAction,
    markAsReview: markAsReviewAction,
    completeLevel: completeLevelAction,
    endSession: endSessionAction,
    restartLevel: restartLevelAction,
  }), []);

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
  useGroupPlayEvents(groupId, {
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

      fetch(`${apiUrl}/api/play-group/${sid}/sync-playtime`, {
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

  // Trigger completeAndWait when entering result phase
  useEffect(() => {
    if (phase === "result" && groupPhase !== "result") {
      completeAndWait();
    }
  }, [phase, groupPhase]);

  if (phase === "loading") {
    return (
      <GamePlayProvider actions={playActions}>
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
      </GamePlayProvider>
    );
  }

  if (phase === "result") {
    if (groupPhase !== "result") {
      return <GroupPlayWaitingScreen groupId={groupId} />;
    }
    return <GroupPlayResultPanel result={groupResult!} groupId={groupId} />;
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <GamePlayProvider actions={playActions}>
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
    </GamePlayProvider>
  );
}
