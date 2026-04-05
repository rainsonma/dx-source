import { create } from "zustand";
import {
  createComboState,
  processAnswer,
  type ComboState,
} from "@/features/web/play-core/helpers/scoring";
import type { ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import type { PkLevelCompleteEvent, PkPlayerActionEvent } from "../types/pk-play";

export type { ContentItem };

interface PkPlayState {
  // Session
  pkId: string | null;
  sessionId: string | null;
  levelSessionId: string | null;
  gameId: string | null;
  gameMode: string | null;
  degree: string | null;
  pattern: string | null;
  levelId: string | null;
  difficulty: string | null;

  // Content
  contentItems: ContentItem[] | null;
  currentIndex: number;
  startFromIndex: number;

  // Scoring
  score: number;
  combo: ComboState;
  correctCount: number;
  wrongCount: number;
  skipCount: number;
  playTime: number;

  // PK-specific
  opponentId: string | null;
  opponentName: string | null;
  pkPhase: "playing" | "waiting" | "result" | null;
  pkResult: PkLevelCompleteEvent | null;
  opponentCompleted: boolean;
  opponentScore: number;
  opponentCombo: number;
  opponentItemsPlayed: number;
  lastOpponentAction: PkPlayerActionEvent | null;
  timeoutCountdown: number | null;
}

interface PkPlayActions {
  initSession: (data: {
    pkId: string;
    sessionId: string;
    levelSessionId: string;
    gameId: string;
    gameMode: string;
    degree: string;
    pattern: string | null;
    levelId: string;
    difficulty: string;
    opponentId: string;
    opponentName: string;
    contentItems: ContentItem[];
    startFromIndex: number;
    restored?: {
      score: number;
      maxCombo: number;
      correctCount: number;
      wrongCount: number;
      skipCount: number;
      playTime: number;
    };
  }) => void;
  nextItem: () => void;
  recordResult: (isCorrect: boolean) => void;
  recordSkip: () => void;
  setPkWaiting: () => void;
  setPkResult: (result: PkLevelCompleteEvent) => void;
  setOpponentCompleted: (score?: number) => void;
  setLastOpponentAction: (action: PkPlayerActionEvent) => void;
  trackOpponentAction: (action: PkPlayerActionEvent) => void;
  setTimeoutCountdown: (seconds: number | null) => void;
  exitGame: () => void;
}

const initialState: PkPlayState = {
  pkId: null,
  sessionId: null,
  levelSessionId: null,
  gameId: null,
  gameMode: null,
  degree: null,
  pattern: null,
  levelId: null,
  difficulty: null,
  contentItems: null,
  currentIndex: 0,
  startFromIndex: 0,
  score: 0,
  combo: createComboState(),
  correctCount: 0,
  wrongCount: 0,
  skipCount: 0,
  playTime: 0,
  opponentId: null,
  opponentName: null,
  pkPhase: null,
  pkResult: null,
  opponentCompleted: false,
  opponentScore: 0,
  opponentCombo: 0,
  opponentItemsPlayed: 0,
  lastOpponentAction: null,
  timeoutCountdown: null,
};

export const usePkPlayStore = create<PkPlayState & PkPlayActions>()(
  (set) => ({
    ...initialState,

    initSession: (data) =>
      set({
        pkId: data.pkId,
        sessionId: data.sessionId,
        levelSessionId: data.levelSessionId,
        gameId: data.gameId,
        gameMode: data.gameMode,
        degree: data.degree,
        pattern: data.pattern,
        levelId: data.levelId,
        difficulty: data.difficulty,
        opponentId: data.opponentId,
        opponentName: data.opponentName,
        contentItems: data.contentItems,
        startFromIndex: data.startFromIndex,
        currentIndex: data.startFromIndex,
        score: data.restored?.score ?? 0,
        combo: data.restored
          ? {
              streak: 0,
              cyclePosition: 0,
              totalScore: data.restored.score,
              maxCombo: data.restored.maxCombo,
            }
          : createComboState(),
        correctCount: data.restored?.correctCount ?? 0,
        wrongCount: data.restored?.wrongCount ?? 0,
        skipCount: data.restored?.skipCount ?? 0,
        playTime: data.restored?.playTime ?? 0,
        pkPhase: "playing",
        pkResult: null,
        opponentCompleted: false,
        opponentScore: 0,
        opponentCombo: 0,
        opponentItemsPlayed: 0,
        lastOpponentAction: null,
        timeoutCountdown: null,
      }),

    nextItem: () => set((s) => ({ currentIndex: s.currentIndex + 1 })),

    recordResult: (isCorrect) =>
      set((s) => {
        const { state, pointsEarned } = processAnswer(s.combo, isCorrect);
        return {
          combo: state,
          score: s.score + pointsEarned,
          correctCount: s.correctCount + (isCorrect ? 1 : 0),
          wrongCount: s.wrongCount + (isCorrect ? 0 : 1),
        };
      }),

    recordSkip: () =>
      set((s) => ({
        combo: { ...s.combo, streak: 0, cyclePosition: 0 },
        skipCount: s.skipCount + 1,
      })),

    setPkWaiting: () => set({ pkPhase: "waiting" }),

    setPkResult: (result) =>
      set({ pkPhase: "result", pkResult: result, opponentCompleted: false }),

    setOpponentCompleted: (score) =>
      set({
        opponentCompleted: true,
        ...(score !== undefined ? { opponentScore: score } : {}),
      }),

    setLastOpponentAction: (action) => set({ lastOpponentAction: action }),

    trackOpponentAction: (action) =>
      set((s) => {
        const isScore = action.action === "score";
        const isCombo = action.action === "combo";
        return {
          lastOpponentAction: action,
          opponentItemsPlayed: s.opponentItemsPlayed + 1,
          opponentScore: isScore ? s.opponentScore + 1 : s.opponentScore,
          opponentCombo: isCombo
            ? (action.combo_streak ?? s.opponentCombo)
            : isScore
              ? s.opponentCombo
              : 0,
        };
      }),

    setTimeoutCountdown: (seconds) => set({ timeoutCountdown: seconds }),

    exitGame: () => set({ ...initialState }),
  })
);
