import { create } from "zustand";
import {
  createComboState,
  processAnswer,
  type ComboState,
} from "@/features/web/play-core/helpers/scoring";
import type { ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import type { GroupLevelCompleteEvent } from "../types/group-play";

export type GamePhase = "loading" | "playing" | "result";
export type GameOverlay = "paused" | "settings" | "reset" | "report" | "exit" | null;

export type { ContentItem };

interface GroupPlayState {
  phase: GamePhase;
  overlay: GameOverlay;
  sessionId: string | null;
  levelSessionId: string | null;
  gameId: string | null;
  gameMode: string | null;
  degree: string | null;
  pattern: string | null;
  levelId: string | null;
  contentItems: ContentItem[] | null;
  startFromIndex: number;
  currentIndex: number;
  score: number;
  combo: ComboState;
  correctCount: number;
  wrongCount: number;
  skipCount: number;
  playTime: number;
  gameGroupId: string | null;
  levelTimeLimit: number | null;
  groupPhase: "playing" | "waiting" | "result" | null;
  groupResult: GroupLevelCompleteEvent | null;
}

interface GroupPlayActions {
  initSession: (data: {
    sessionId: string;
    levelSessionId: string;
    gameId: string;
    gameMode: string;
    degree: string;
    pattern: string | null;
    levelId: string;
    contentItems: ContentItem[];
    startFromIndex: number;
    gameGroupId: string;
    levelTimeLimit: number;
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
  setPhase: (phase: GamePhase) => void;
  showOverlay: (overlay: GameOverlay) => void;
  closeOverlay: () => void;
  resetGame: () => void;
  exitGame: () => void;
  setGroupWaiting: () => void;
  setGroupResult: (result: GroupLevelCompleteEvent) => void;
  clearGroupPhase: () => void;
}

const initialState: GroupPlayState = {
  phase: "loading",
  overlay: null,
  sessionId: null,
  levelSessionId: null,
  gameId: null,
  gameMode: null,
  degree: null,
  pattern: null,
  levelId: null,
  contentItems: null,
  startFromIndex: 0,
  currentIndex: 0,
  score: 0,
  combo: createComboState(),
  correctCount: 0,
  wrongCount: 0,
  skipCount: 0,
  playTime: 0,
  gameGroupId: null,
  levelTimeLimit: null,
  groupPhase: null,
  groupResult: null,
};

export const useGroupPlayStore = create<GroupPlayState & GroupPlayActions>()(
  (set) => ({
    ...initialState,

    initSession: (data) =>
      set({
        phase: "playing",
        sessionId: data.sessionId,
        levelSessionId: data.levelSessionId,
        gameId: data.gameId,
        gameMode: data.gameMode,
        degree: data.degree,
        pattern: data.pattern,
        levelId: data.levelId,
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
        gameGroupId: data.gameGroupId,
        levelTimeLimit: data.levelTimeLimit,
        groupPhase: "playing",
        groupResult: null,
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

    setPhase: (phase) => set({ phase }),

    showOverlay: (overlay) => set({ overlay }),

    closeOverlay: () => set({ overlay: null }),

    resetGame: () =>
      set({
        phase: "loading",
        overlay: null,
        levelSessionId: null,
        currentIndex: 0,
        score: 0,
        combo: createComboState(),
        correctCount: 0,
        wrongCount: 0,
        skipCount: 0,
        playTime: 0,
      }),

    exitGame: () => set({ ...initialState }),

    setGroupWaiting: () => set({ groupPhase: "waiting" }),
    setGroupResult: (result) => set({ groupPhase: "result", groupResult: result }),
    clearGroupPhase: () => set({ groupPhase: null, groupResult: null }),
  })
);
