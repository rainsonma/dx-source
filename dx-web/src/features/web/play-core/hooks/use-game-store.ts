import { create } from "zustand";
import {
  createComboState,
  processAnswer,
  type ComboState,
} from "@/features/web/play-core/helpers/scoring";

export type GamePhase = "loading" | "playing" | "result";
export type GameOverlay = "paused" | "settings" | "reset" | "report" | "exit" | null;

export type ContentItem = {
  id: string;
  content: string;
  contentType: string;
  translation: string | null;
  definition: string | null;
  explanation: string | null;
  items: unknown;
  structure: unknown;
  ukAudio: { url: string } | null;
  usAudio: { url: string } | null;
};

interface GameState {
  phase: GamePhase;
  overlay: GameOverlay;
  sessionId: string | null;
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
}

interface GameActions {
  initSession: (data: {
    sessionId: string;
    gameId: string;
    gameMode: string;
    degree: string;
    pattern: string | null;
    levelId: string;
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
  setPhase: (phase: GamePhase) => void;
  showOverlay: (overlay: GameOverlay) => void;
  closeOverlay: () => void;
  resetGame: () => void;
  exitGame: () => void;
}

const initialGameState: GameState = {
  phase: "loading",
  overlay: null,
  sessionId: null,
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
};

export const useGameStore = create<GameState & GameActions>()((set) => ({
  ...initialGameState,

  initSession: (data) =>
    set({
      phase: "playing",
      sessionId: data.sessionId,
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
    }),

  nextItem: () =>
    set((s) => ({ currentIndex: s.currentIndex + 1 })),

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
      currentIndex: 0,
      score: 0,
      combo: createComboState(),
      correctCount: 0,
      wrongCount: 0,
      skipCount: 0,
      playTime: 0,
    }),

  exitGame: () => set({ ...initialGameState }),
}));
