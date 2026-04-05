import { create } from "zustand";
import {
  createComboState,
  processAnswer,
  type ComboState,
} from "@/features/web/play-core/helpers/scoring";
import type { ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import type { GroupLevelCompleteEvent, GroupPlayerCompleteEvent, Participants, GroupPlayerActionEvent } from "../types/group-play";

export type GamePhase = "loading" | "playing" | "result";
export type GameOverlay = "paused" | "settings" | "reset" | "report" | "exit" | null;

export type { ContentItem };

interface GroupPlayState {
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
  playTime: number;
  gameGroupId: string | null;
  groupPhase: "playing" | "result" | null;
  groupResult: GroupLevelCompleteEvent | null;
  participants: Participants | null;
  completedPlayerIds: string[];
  lastPlayerAction: GroupPlayerActionEvent | null;
  nextLevelId: string | null;
  nextLevelName: string | null;
}

interface GroupPlayActions {
  initSession: (data: {
    sessionId: string;
    gameId: string;
    gameMode: string;
    degree: string;
    pattern: string | null;
    levelId: string;
    contentItems: ContentItem[];
    startFromIndex: number;
    gameGroupId: string;
    participants?: Participants | null;
    restored?: {
      score: number;
      maxCombo: number;
      correctCount: number;
      wrongCount: number;
      playTime: number;
    };
  }) => void;
  nextItem: () => void;
  recordResult: (isCorrect: boolean) => void;
  setPhase: (phase: GamePhase) => void;
  showOverlay: (overlay: GameOverlay) => void;
  closeOverlay: () => void;
  resetGame: () => void;
  exitGame: () => void;
  setGroupResult: (result: GroupLevelCompleteEvent) => void;
  setGroupResultFromWinner: (event: GroupPlayerCompleteEvent) => void;
  clearGroupPhase: () => void;
  setParticipants: (data: Participants) => void;
  addCompletedPlayer: (userId: string) => void;
  setLastPlayerAction: (action: GroupPlayerActionEvent) => void;
  setNextLevel: (id: string, name: string) => void;
}

const initialState: GroupPlayState = {
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
  playTime: 0,
  gameGroupId: null,
  groupPhase: null,
  groupResult: null,
  participants: null,
  completedPlayerIds: [],
  lastPlayerAction: null,
  nextLevelId: null,
  nextLevelName: null,
};

export const useGroupPlayStore = create<GroupPlayState & GroupPlayActions>()(
  (set) => ({
    ...initialState,

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
        playTime: data.restored?.playTime ?? 0,
        gameGroupId: data.gameGroupId,
        groupPhase: "playing",
        groupResult: null,
        participants: data.participants ?? null,
        completedPlayerIds: [],
        lastPlayerAction: null,
        nextLevelId: null,
        nextLevelName: null,
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
        playTime: 0,
        lastPlayerAction: null,
        nextLevelId: null,
        nextLevelName: null,
      }),

    exitGame: () => set({ ...initialState }),

    setGroupResult: (result) => set({ groupPhase: "result", groupResult: result, completedPlayerIds: [] }),
    setGroupResultFromWinner: (event) =>
      set({
        groupPhase: "result",
        groupResult: {
          game_level_id: event.game_level_id,
          mode: "group_solo",
          winner: { user_id: event.user_id, user_name: event.user_name, score: event.score },
          participants: [{ user_id: event.user_id, user_name: event.user_name, score: event.score }],
        },
        completedPlayerIds: [],
      }),
    clearGroupPhase: () => set({ groupPhase: null, groupResult: null, completedPlayerIds: [], nextLevelId: null, nextLevelName: null }),
    setParticipants: (data) => set({ participants: data }),
    addCompletedPlayer: (userId) =>
      set((s) => ({
        completedPlayerIds: s.completedPlayerIds.includes(userId)
          ? s.completedPlayerIds
          : [...s.completedPlayerIds, userId],
      })),

    setLastPlayerAction: (action) => set({ lastPlayerAction: action }),

    setNextLevel: (id, name) => set({ nextLevelId: id, nextLevelName: name }),
  })
);
