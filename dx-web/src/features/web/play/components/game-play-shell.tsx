"use client";

import { useEffect } from "react";
import { GAME_MODES } from "@/consts/game-mode";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { GameLoadingScreen } from "@/features/web/play/components/game-loading-screen";
import { GameTopBar } from "@/features/web/play/components/game-top-bar";
import { GameResultCard } from "@/features/web/play/components/game-result-card";
import { GamePauseOverlay } from "@/features/web/play/components/game-pause-overlay";
import { GameSettingsModal } from "@/features/web/play/components/game-settings-modal";
import { GameResetModal } from "@/features/web/play/components/game-reset-modal";
import { GameReportModal } from "@/features/web/play/components/game-report-modal";
import { LsrwGame } from "@/features/web/play/components/lsrw-game";
import { VocabMatchGame } from "@/features/web/play/components/vocab-match-game";
import { ListeningGame } from "@/features/web/play/components/listening-game";
import { VocabEliminationGame } from "@/features/web/play/components/vocab-elimination-game";
import { VocabBattleGame } from "@/features/web/play/components/vocab-battle-game";
import { GameExitModal } from "@/features/web/play/components/game-exit-modal";
import { useGameTimer, getElapsedSeconds } from "@/features/web/play/hooks/use-game-timer";
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

interface GamePlayShellProps {
  game: {
    id: string;
    name: string;
    mode: string;
    levels: { id: string; name: string; order: number }[];
  };
  player: { nickname: string; avatarUrl: string | null };
  degree?: string;
  pattern?: string;
  levelId?: string;
  groupId?: string;
  levelTimeLimit?: number;
}

export function GamePlayShell({ game, player, degree, pattern, levelId, groupId, levelTimeLimit }: GamePlayShellProps) {
  const phase = useGameStore((s) => s.phase);
  const overlay = useGameStore((s) => s.overlay);
  const showOverlay = useGameStore((s) => s.showOverlay);
  const closeOverlay = useGameStore((s) => s.closeOverlay);
  const setPhase = useGameStore((s) => s.setPhase);

  const targetLevelId = levelId ?? game.levels[0]?.id;
  const targetLevel = game.levels.find((l) => l.id === targetLevelId);
  const levelName = targetLevel?.name ?? game.name;

  const storeGameId = useGameStore((s) => s.gameId);
  const storeLevelId = useGameStore((s) => s.levelId);
  const exitGame = useGameStore((s) => s.exitGame);
  const playTime = useGameStore((s) => s.playTime);
  const { formatted: elapsedTime } = useGameTimer(playTime);
  const { isFullscreen, toggleFullscreen } = useFullscreen();

  useEffect(() => {
    const isDifferentGame = storeGameId !== null && storeGameId !== game.id;
    const isDifferentLevel =
      storeLevelId !== null && storeLevelId !== targetLevelId;
    const isStaleState =
      storeGameId === game.id && phase !== "loading";

    if (isDifferentGame || isDifferentLevel || isStaleState) {
      exitGame();
    }
    // Store values (storeGameId, storeLevelId, phase) intentionally excluded —
    // we only re-run when external props change (soft navigation), not on
    // internal store transitions (loading → playing → result).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [game.id, targetLevelId]);

  // Flush playtime to server on tab close / navigation
  useEffect(() => {
    const handleBeforeUnload = () => {
      const sid = useGameStore.getState().sessionId;
      const lid = useGameStore.getState().levelId;
      if (!sid || !lid) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";
      const token = getToken();
      const headers: Record<string, string> = { "Content-Type": "application/json" };
      if (token) headers["Authorization"] = `Bearer ${token}`;

      fetch(`${apiUrl}/api/sessions/${sid}/sync-playtime`, {
        method: "POST",
        headers,
        body: JSON.stringify({
          game_level_id: lid,
          play_time: getElapsedSeconds(),
        }),
        keepalive: true,
      });
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, []);

  if (phase === "loading") {
    return (
      <GameLoadingScreen
        gameId={game.id}
        gameName={game.name}
        gameMode={game.mode}
        degree={degree}
        pattern={pattern}
        levelId={targetLevelId}
        levelName={targetLevel?.name}
        groupId={groupId}
        levelTimeLimit={levelTimeLimit}
      />
    );
  }

  if (phase === "result") {
    return <GameResultCard game={game} elapsedTime={elapsedTime} />;
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <div className="flex h-screen w-full flex-col bg-muted">
      <GameTopBar
        player={player}
        levelName={levelName}
        elapsedTime={elapsedTime}
        onExit={() => showOverlay("exit")}
        onReset={() => showOverlay("reset")}
        onSettings={() => showOverlay("settings")}
        onPause={() => showOverlay("paused")}
        onReport={() => showOverlay("report")}
        onFullscreen={toggleFullscreen}
        isFullscreen={isFullscreen}
        levelTimeLimit={levelTimeLimit}
        onLevelTimeUp={() => {
          // Time's up — transition to result phase which triggers completeLevel
          setPhase("result");
        }}
      />
      <div className="flex flex-1 flex-col items-center justify-center gap-6 overflow-y-auto px-4 py-10">
        <GameComponent />
      </div>

      {overlay === "paused" && (
        <GamePauseOverlay
          elapsedTime={elapsedTime}
          gameId={game.id}
          onResume={closeOverlay}
        />
      )}
      {overlay === "settings" && <GameSettingsModal />}
      {overlay === "reset" && <GameResetModal />}
      {overlay === "report" && <GameReportModal />}
      {overlay === "exit" && <GameExitModal gameId={game.id} />}
    </div>
  );
}
