"use client";

import { useEffect, useMemo } from "react";
import { GAME_MODES } from "@/consts/game-mode";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { GamePlayProvider, type GamePlayActions } from "@/features/web/play-core/context/game-play-context";
import {
  recordAnswerAction,
  recordSkipAction,
  completeLevelAction,
  endSessionAction,
  restartLevelSessionAction,
} from "@/features/web/play-single/actions/session.action";
import { markAsReviewAction } from "@/features/web/play-core/actions/tracking.action";
import { GameLoadingScreen } from "@/features/web/play-single/components/game-loading-screen";
import { GameTopBar } from "@/features/web/play-single/components/game-top-bar";
import { GameResultCard } from "@/features/web/play-single/components/game-result-card";
import { GamePauseOverlay } from "@/features/web/play-core/components/game-pause-overlay";
import { GameSettingsModal } from "@/features/web/play-core/components/game-settings-modal";
import { GameResetModal } from "@/features/web/play-core/components/game-reset-modal";
import { GameReportModal } from "@/features/web/play-core/components/game-report-modal";
import { GameWordSentence } from "@/features/web/play-core/components/game-word-sentence";
import { GameVocabMatch } from "@/features/web/play-core/components/game-vocab-match";
import { GameListening } from "@/features/web/play-core/components/game-listening";
import { GameVocabElimination } from "@/features/web/play-core/components/game-vocab-elimination";
import { GameVocabBattle } from "@/features/web/play-core/components/game-vocab-battle";
import { GameExitModal } from "@/features/web/play-core/components/game-exit-modal";
import { useGameTimer, getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { useFullscreen } from "@/features/web/play-core/hooks/use-fullscreen";
import type { ComponentType } from "react";

const modeComponents: Record<string, ComponentType> = {
  [GAME_MODES.WORD_SENTENCE]: GameWordSentence,
  [GAME_MODES.VOCAB_MATCH]: GameVocabMatch,
  [GAME_MODES.LISTENING_CHALLENGE]: GameListening,
  [GAME_MODES.VOCAB_ELIMINATION]: GameVocabElimination,
  [GAME_MODES.VOCAB_BATTLE]: GameVocabBattle,
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
}

export function GamePlayShell({ game, player, degree, pattern, levelId }: GamePlayShellProps) {
  const phase = useGameStore((s) => s.phase);
  const overlay = useGameStore((s) => s.overlay);
  const showOverlay = useGameStore((s) => s.showOverlay);
  const closeOverlay = useGameStore((s) => s.closeOverlay);

  const targetLevelId = levelId ?? game.levels[0]?.id;
  const targetLevel = game.levels.find((l) => l.id === targetLevelId);
  const levelName = targetLevel?.name ?? game.name;

  const storeGameId = useGameStore((s) => s.gameId);
  const storeLevelId = useGameStore((s) => s.levelId);
  const exitGame = useGameStore((s) => s.exitGame);
  const playTime = useGameStore((s) => s.playTime);
  const { formatted: elapsedTime } = useGameTimer(playTime);
  const { isFullscreen, toggleFullscreen } = useFullscreen();

  const playActions = useMemo<GamePlayActions>(() => ({
    recordAnswer: recordAnswerAction,
    recordSkip: recordSkipAction,
    markAsReview: markAsReviewAction,
    completeLevel: completeLevelAction,
    endSession: endSessionAction,
    restartLevel: restartLevelSessionAction,
  }), []);

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

      fetch(`${apiUrl}/api/play-single/${sid}/sync-playtime`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
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
      <GamePlayProvider actions={playActions}>
      <GameLoadingScreen
        gameId={game.id}
        gameName={game.name}
        gameMode={game.mode}
        degree={degree}
        pattern={pattern}
        levelId={targetLevelId}
        levelName={targetLevel?.name}
      />
      </GamePlayProvider>
    );
  }

  if (phase === "result") {
    return (
      <GamePlayProvider actions={playActions}>
        <GameResultCard game={game} elapsedTime={elapsedTime} />
      </GamePlayProvider>
    );
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <GamePlayProvider actions={playActions}>
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
    </GamePlayProvider>
  );
}
