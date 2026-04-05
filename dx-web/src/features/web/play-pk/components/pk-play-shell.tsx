"use client";

import { useEffect, useRef, useMemo, useState, useCallback } from "react";
import { GAME_MODES } from "@/consts/game-mode";
import { usePkPlayStore } from "../hooks/use-pk-play-store";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { GamePlayProvider, type GamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { PkPlayLoadingScreen } from "./pk-play-loading-screen";
import { PkPlayTopBar } from "./pk-play-top-bar";
import { PkPlayResultPanel } from "./pk-play-result-panel";
import { GameSettingsModal } from "@/features/web/play-core/components/game-settings-modal";
import { GameReportModal } from "@/features/web/play-core/components/game-report-modal";
import { GameExitModal } from "@/features/web/play-core/components/game-exit-modal";
import { GameWordSentence } from "@/features/web/play-core/components/game-word-sentence";
import { GameVocabMatch } from "@/features/web/play-core/components/game-vocab-match";
import { GameListening } from "@/features/web/play-core/components/game-listening";
import { GameVocabElimination } from "@/features/web/play-core/components/game-vocab-elimination";
import { GameVocabBattle } from "@/features/web/play-core/components/game-vocab-battle";
import { usePkPlayEvents } from "../hooks/use-pk-play-events";
import {
  completeLevelAction,
  recordAnswerAction,
  markAsReviewAction,
  pausePkAction,
  resumePkAction,
} from "../actions/session.action";
import { useFullscreen } from "@/features/web/play-core/hooks/use-fullscreen";
import type { ComponentType } from "react";
import type { PkLevelCompleteEvent } from "../types/pk-play";

const modeComponents: Record<string, ComponentType> = {
  [GAME_MODES.WORD_SENTENCE]: GameWordSentence,
  [GAME_MODES.VOCAB_MATCH]: GameVocabMatch,
  [GAME_MODES.LISTENING_CHALLENGE]: GameListening,
  [GAME_MODES.VOCAB_ELIMINATION]: GameVocabElimination,
  [GAME_MODES.VOCAB_BATTLE]: GameVocabBattle,
};

interface PkPlayShellProps {
  game: {
    id: string;
    name: string;
    mode: string;
    levels: { id: string; name: string; order: number }[];
  };
  player: { id: string; nickname: string; avatarUrl: string | null };
  degree: string;
  pattern: string | null;
  levelId: string;
  difficulty: string;
}

export function PkPlayShell({
  game,
  player,
  degree,
  pattern,
  levelId,
  difficulty,
}: PkPlayShellProps) {
  // Phase and overlay managed via useGameStore so shared modals work
  const phase = useGameStore((s) => s.phase);
  const overlay = useGameStore((s) => s.overlay);
  const showOverlay = useGameStore((s) => s.showOverlay);
  const closeOverlay = useGameStore((s) => s.closeOverlay);
  const setPhase = useGameStore((s) => s.setPhase);
  const storeGameId = useGameStore((s) => s.gameId);
  const storeLevelId = useGameStore((s) => s.levelId);
  const exitGame = useGameStore((s) => s.exitGame);

  // PK-specific state
  const pkPhase = usePkPlayStore((s) => s.pkPhase);
  const pkResult = usePkPlayStore((s) => s.pkResult);
  const pkId = usePkPlayStore((s) => s.pkId);
  const sessionId = usePkPlayStore((s) => s.sessionId);
  const opponentName = usePkPlayStore((s) => s.opponentName);
  const lastOpponentAction = usePkPlayStore((s) => s.lastOpponentAction);
  const timeoutCountdown = usePkPlayStore((s) => s.timeoutCountdown);
  const pkNextLevelId = usePkPlayStore((s) => s.nextLevelId);
  const setPkResult = usePkPlayStore((s) => s.setPkResult);
  const trackOpponentAction = usePkPlayStore((s) => s.trackOpponentAction);
  const setTimeoutCountdown = usePkPlayStore((s) => s.setTimeoutCountdown);

  const score = useGameStore((s) => s.score);
  const combo = useGameStore((s) => s.combo);
  const contentItems = useGameStore((s) => s.contentItems);

  const completedRef = useRef(false);
  const [countdownValue, setCountdownValue] = useState<number | null>(null);

  // No-op stubs for actions not applicable to PK mode
  const noopSkip = useCallback(async () => ({ data: null, error: null }), []);
  const noopEndSession = useCallback(async () => ({ data: null, error: null }), []);
  const noopRestartLevel = useCallback(async () => ({ data: null, error: null }), []);

  const playActions = useMemo<GamePlayActions>(() => ({
    recordAnswer: recordAnswerAction,
    recordSkip: noopSkip,
    markAsReview: markAsReviewAction,
    completeLevel: completeLevelAction,
    endSession: noopEndSession,
    restartLevel: noopRestartLevel,
    competitive: true,
  }), [noopSkip, noopEndSession, noopRestartLevel]);

  const targetLevel =
    game.levels.find((l) => l.id === levelId) ?? game.levels[0];
  const targetLevelId = targetLevel?.id ?? levelId;
  const levelName = targetLevel?.name ?? game.name;

  const { isFullscreen, toggleFullscreen } = useFullscreen();

  async function completeAndSetResult() {
    if (completedRef.current || !sessionId || !targetLevelId) return;
    if (usePkPlayStore.getState().levelId !== targetLevelId) return;
    completedRef.current = true;
    const result = await completeLevelAction(sessionId, targetLevelId, {
      score,
      maxCombo: combo.maxCombo,
      totalItems: contentItems?.length ?? 0,
    });
    if (result.error) {
      // Retry once
      await completeLevelAction(sessionId, targetLevelId, {
        score,
        maxCombo: combo.maxCombo,
        totalItems: contentItems?.length ?? 0,
      });
    }
    // Store next level info from complete response for result panel
    if (result.data) {
      const store = usePkPlayStore.getState();
      if (!store.pkResult) {
        // Result not yet set via SSE — will be set when pk_player_complete arrives
        // But store nextLevel info for when it does
        usePkPlayStore.setState({
          nextLevelId: result.data.next_level_id ?? null,
          nextLevelName: result.data.next_level_name ?? null,
        });
      }
    }
  }

  // SSE: listen for PK events
  usePkPlayEvents(pkId, {
    onForceEnd: () => {
      setPhase("result");
    },
    onPlayerComplete: (event) => {
      const currentLevelId = usePkPlayStore.getState().levelId;
      if (event.game_level_id === currentLevelId) {
        // First-to-complete: build result and immediately show result
        const store = usePkPlayStore.getState();
        const result: PkLevelCompleteEvent = {
          game_level_id: event.game_level_id,
          winner: {
            user_id: event.user_id,
            user_name: event.user_name,
            score: event.score,
          },
          participants: [
            { user_id: event.user_id, user_name: event.user_name, score: event.score },
            { user_id: store.opponentId ?? "", user_name: store.opponentName ?? "", score: store.opponentScore },
          ],
        };
        setPkResult(result, store.nextLevelId, store.nextLevelName);
        setPhase("result");
      }
    },
    onPlayerAction: (event) => {
      trackOpponentAction(event);
    },
    onTimeoutWarning: (event) => {
      setTimeoutCountdown(event.countdown);
    },
    onTimeout: () => {
      setPhase("result");
      completeAndSetResult();
    },
  });

  // Timeout countdown interval
  useEffect(() => {
    if (timeoutCountdown === null || timeoutCountdown <= 0) {
      setCountdownValue(null);
      return;
    }
    setCountdownValue(timeoutCountdown);
    const interval = setInterval(() => {
      setCountdownValue((prev) => {
        if (prev === null || prev <= 1) {
          clearInterval(interval);
          return null;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [timeoutCountdown]);

  // Reset on game/level change
  useEffect(() => {
    const isDifferentGame = storeGameId !== null && storeGameId !== game.id;
    const isDifferentLevel =
      storeLevelId !== null && storeLevelId !== targetLevelId;
    const isStaleState = storeGameId === game.id && phase !== "loading";

    if (isDifferentGame || isDifferentLevel || isStaleState) {
      exitGame();
      completedRef.current = false;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [game.id, targetLevelId]);

  // Flush playtime on tab close / navigation
  useEffect(() => {
    const handleBeforeUnload = () => {
      const sid = usePkPlayStore.getState().sessionId;
      if (!sid) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

      fetch(`${apiUrl}/api/play-pk/${sid}/sync-playtime`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          play_time: usePkPlayStore.getState().playTime,
        }),
        keepalive: true,
      });
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, []);

  // Trigger completeAndSetResult when entering result phase
  useEffect(() => {
    if (phase === "result" && pkPhase !== "result") {
      completeAndSetResult();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [phase, pkPhase]);

  // Pause/resume handlers
  function handlePause() {
    if (pkId) pausePkAction(pkId);
    showOverlay("paused");
  }

  function handleCloseOverlay() {
    if (overlay === "paused" && pkId) resumePkAction(pkId);
    closeOverlay();
  }

  if (phase === "loading") {
    return (
      <GamePlayProvider actions={playActions}>
        <PkPlayLoadingScreen
          gameId={game.id}
          gameName={game.name}
          gameMode={game.mode}
          degree={degree}
          pattern={pattern}
          levelId={targetLevelId}
          levelName={levelName}
          difficulty={difficulty}
        />
      </GamePlayProvider>
    );
  }

  if (phase === "result") {
    return (
      <PkPlayResultPanel
        result={pkResult}
        pkId={pkId!}
        gameId={game.id}
        levelName={levelName}
        nextLevelId={pkNextLevelId}
      />
    );
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <GamePlayProvider actions={playActions}>
      <div className="flex h-screen w-full flex-col bg-muted">
        <PkPlayTopBar
          player={player}
          playerId={player.id}
          opponentName={opponentName ?? "对手"}
          lastOpponentAction={lastOpponentAction}
          onExit={() => showOverlay("exit")}
          onPause={handlePause}
          onSettings={() => showOverlay("settings")}
          onReport={() => showOverlay("report")}
          onFullscreen={toggleFullscreen}
          isFullscreen={isFullscreen}
        />
        <div className="flex flex-1 flex-col items-center justify-center gap-6 overflow-y-auto px-4 py-10">
          <GameComponent />
        </div>

        {/* Timeout countdown overlay */}
        {countdownValue !== null && countdownValue > 0 && (
          <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/60">
            <div className="flex flex-col items-center gap-3 rounded-2xl bg-card px-8 py-6">
              <span className="text-4xl font-extrabold tabular-nums text-red-500">
                {countdownValue}
              </span>
              <p className="text-sm font-medium text-muted-foreground">
                你将在 {countdownValue} 秒后输掉本场 PK
              </p>
            </div>
          </div>
        )}

        {overlay === "paused" && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
            <div className="flex w-full max-w-[320px] flex-col items-center gap-5 rounded-[20px] bg-card px-8 py-9">
              <h2 className="text-base font-bold text-foreground">已暂停</h2>
              <button
                type="button"
                onClick={handleCloseOverlay}
                className="flex h-12 w-full items-center justify-center rounded-xl bg-teal-600"
              >
                <span className="text-[15px] font-semibold text-white">继续游戏</span>
              </button>
            </div>
          </div>
        )}
        {overlay === "settings" && <GameSettingsModal />}
        {overlay === "report" && <GameReportModal />}
        {overlay === "exit" && <GameExitModal gameId={game.id} />}
      </div>
    </GamePlayProvider>
  );
}
