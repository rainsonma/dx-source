"use client";

import { useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { GAME_MODES } from "@/consts/game-mode";
import { useGroupPlayStore } from "../hooks/use-group-play-store";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { GamePlayProvider, type GamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { GroupPlayLoadingScreen } from "./group-play-loading-screen";
import { GroupPlayTopBar } from "./group-play-top-bar";
// GroupPlayWaitingScreen removed — first-to-complete means immediate result
import { GroupPlayResultPanel } from "./group-play-result-panel";
import { GameSettingsModal } from "@/features/web/play-core/components/game-settings-modal";
import { GameResetModal } from "@/features/web/play-core/components/game-reset-modal";
import { GameReportModal } from "@/features/web/play-core/components/game-report-modal";
import { GameExitModal } from "@/features/web/play-core/components/game-exit-modal";
import { GameWordSentence } from "@/features/web/play-core/components/game-word-sentence";
import { GameVocabMatch } from "@/features/web/play-core/components/game-vocab-match";
import { GameListening } from "@/features/web/play-core/components/game-listening";
import { GameVocabElimination } from "@/features/web/play-core/components/game-vocab-elimination";
import { GameVocabBattle } from "@/features/web/play-core/components/game-vocab-battle";
import { toast } from "sonner";
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
import type { ComponentType } from "react";

const modeComponents: Record<string, ComponentType> = {
  [GAME_MODES.WORD_SENTENCE]: GameWordSentence,
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
  player: { id: string; nickname: string; avatarUrl: string | null };
  degree: string;
  pattern: string | null;
  levelId: string;
  groupId: string;
  gameMode: string | null;
}

export function GroupPlayShell({
  game,
  player,
  degree,
  pattern,
  levelId,
  groupId,
}: GroupPlayShellProps) {
  const router = useRouter();

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
  const setGroupResult = useGroupPlayStore((s) => s.setGroupResult);
  const setGroupResultFromWinner = useGroupPlayStore((s) => s.setGroupResultFromWinner);
  const addCompletedPlayer = useGroupPlayStore((s) => s.addCompletedPlayer);
  const setLastPlayerAction = useGroupPlayStore((s) => s.setLastPlayerAction);
  const setNextLevel = useGroupPlayStore((s) => s.setNextLevel);
  const nextLevelId = useGroupPlayStore((s) => s.nextLevelId);
  const nextLevelName = useGroupPlayStore((s) => s.nextLevelName);


  const playActions = useMemo<GamePlayActions>(() => ({
    recordAnswer: recordAnswerAction,
    recordSkip: recordSkipAction,
    markAsReview: markAsReviewAction,
    completeLevel: async (...args: Parameters<typeof completeLevelAction>) => {
      const result = await completeLevelAction(...args);
      if (result.data) {
        if (result.data.nextLevelId && result.data.nextLevelName) {
          setNextLevel(result.data.nextLevelId, result.data.nextLevelName);
        }
        // Immediately show group result for the winner before phase changes
        const store = useGroupPlayStore.getState();
        if (store.groupPhase !== "result") {
          setGroupResultFromWinner(
            { user_id: player.id, user_name: player.nickname, game_level_id: args[1], score: useGameStore.getState().score, participants: [], next_level_id: result.data.nextLevelId ?? null, next_level_name: result.data.nextLevelName ?? null },
            [{ user_id: player.id, user_name: player.nickname, score: useGameStore.getState().score }],
          );
        }
      }
      return result;
    },
    endSession: endSessionAction,
    restartLevel: restartLevelAction,
    competitive: true,
  }), [player.id, player.nickname, setNextLevel, setGroupResultFromWinner]);

  const targetLevel =
    game.levels.find((l) => l.id === levelId) ?? game.levels[0];
  const targetLevelId = targetLevel?.id ?? levelId;
  const levelName = targetLevel?.name ?? game.name;

  const { isFullscreen, toggleFullscreen } = useFullscreen();

  // SSE: listen for group level complete, force-end, and player events
  useGroupPlayEvents(groupId, {
    onLevelComplete: (event) => {
      setGroupResult(event);
    },
    onForceEnd: (event) => {
      // Force-end: show the last level result (or first if available)
      const lastResult = event.results[event.results.length - 1];
      if (lastResult) {
        setGroupResult(lastResult);
      }
      setPhase("result");
    },
    onPlayerComplete: (event) => {
      const currentLevelId = useGroupPlayStore.getState().levelId;
      if (event.game_level_id === currentLevelId) {
        addCompletedPlayer(event.user_id);
        // First-to-complete: winner event triggers immediate result
        const store = useGroupPlayStore.getState();
        if (store.groupPhase !== "result") {
          setGroupResultFromWinner(event, event.participants);
          if (event.next_level_id && event.next_level_name) {
            setNextLevel(event.next_level_id, event.next_level_name);
          }
          setPhase("result");
        }
      }
    },
    onPlayerAction: (event) => {
      setLastPlayerAction(event);
    },
    onDismissed: () => {
      toast.error("群组已被解散");
      router.push("/hall/groups");
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

      fetch(`${apiUrl}/api/play-group/${sid}/sync-playtime`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
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
      />
      </GamePlayProvider>
    );
  }

  if (phase === "result" && groupPhase === "result") {
    return (
      <GroupPlayResultPanel
        result={groupResult!}
        groupId={groupId}
        levelName={levelName}
        nextLevelId={nextLevelId}
        nextLevelName={nextLevelName}
      />
    );
  }

  const GameComponent = modeComponents[game.mode];
  if (!GameComponent) return null;

  return (
    <GamePlayProvider actions={playActions}>
    <div className="flex h-screen w-full flex-col bg-muted">
      <GroupPlayTopBar
        player={player}
        playerId={player.id}
        levelName={levelName}
        onExit={() => showOverlay("exit")}
        onReset={() => showOverlay("reset")}
        onSettings={() => showOverlay("settings")}
        onReport={() => showOverlay("report")}
        onFullscreen={toggleFullscreen}
        isFullscreen={isFullscreen}
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
