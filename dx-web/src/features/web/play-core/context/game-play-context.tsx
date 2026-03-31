"use client";

import { createContext, useContext, type ReactNode } from "react";

// Action function types that each shell (play-single, play-group) provides
export type RecordAnswerFn = (data: {
  gameSessionTotalId: string;
  gameSessionLevelId: string;
  gameLevelId: string;
  contentItemId: string;
  isCorrect: boolean;
  userAnswer: string;
  sourceAnswer: string;
  baseScore: number;
  comboScore: number;
  score: number;
  maxCombo: number;
  playTime: number;
  nextContentItemId: string | null;
  duration: number;
}) => Promise<{ data: unknown; error: string | null }>;

export type RecordSkipFn = (data: {
  gameSessionTotalId: string;
  gameLevelId: string;
  playTime: number;
  nextContentItemId: string | null;
}) => Promise<{ data: unknown; error: string | null }>;

export type MarkAsReviewFn = (data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) => Promise<unknown>;

export type CompleteLevelFn = (
  sessionId: string,
  gameLevelId: string,
  data: { score: number; maxCombo: number; totalItems: number }
) => Promise<{ data: unknown; error: string | null }>;

export type EndSessionFn = (
  sessionId: string,
  data: {
    gameId: string;
    score: number;
    exp: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
    allLevelsCompleted: boolean;
  }
) => Promise<{ data: unknown; error: string | null }>;

export type RestartLevelFn = (
  sessionId: string,
  levelId: string
) => Promise<{ data: unknown; error: string | null }>;

export interface GamePlayActions {
  recordAnswer: RecordAnswerFn;
  recordSkip: RecordSkipFn;
  markAsReview: MarkAsReviewFn;
  completeLevel: CompleteLevelFn;
  endSession: EndSessionFn;
  restartLevel: RestartLevelFn;
}

const GamePlayContext = createContext<GamePlayActions | null>(null);

export function GamePlayProvider({
  actions,
  children,
}: {
  actions: GamePlayActions;
  children: ReactNode;
}) {
  return (
    <GamePlayContext.Provider value={actions}>
      {children}
    </GamePlayContext.Provider>
  );
}

export function useGamePlayActions(): GamePlayActions {
  const ctx = useContext(GamePlayContext);
  if (!ctx) {
    throw new Error("useGamePlayActions must be used within a GamePlayProvider");
  }
  return ctx;
}
