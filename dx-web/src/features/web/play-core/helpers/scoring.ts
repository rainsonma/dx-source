import { SCORING, COMBO_THRESHOLDS } from "@/consts/scoring";

export type ComboState = {
  streak: number;
  cyclePosition: number;
  totalScore: number;
  maxCombo: number;
};

export function createComboState(): ComboState {
  return { streak: 0, cyclePosition: 0, totalScore: 0, maxCombo: 0 };
}

export function processAnswer(
  state: ComboState,
  isCorrect: boolean
): { state: ComboState; pointsEarned: number; comboBonus: number } {
  if (!isCorrect) {
    return {
      state: { ...state, streak: 0, cyclePosition: 0 },
      pointsEarned: 0,
      comboBonus: 0,
    };
  }

  let points = SCORING.CORRECT_ANSWER;
  let bonus = 0;
  const newStreak = state.streak + 1;
  let newCyclePosition = state.cyclePosition + 1;

  for (const threshold of COMBO_THRESHOLDS) {
    if (newCyclePosition === threshold.streak) {
      bonus += threshold.bonus;
    }
  }

  if (newCyclePosition >= SCORING.COMBO_CYCLE_LENGTH) {
    newCyclePosition = 0;
  }

  points += bonus;

  return {
    state: {
      streak: newStreak,
      cyclePosition: newCyclePosition,
      totalScore: state.totalScore + points,
      maxCombo: Math.max(state.maxCombo, newStreak),
    },
    pointsEarned: points,
    comboBonus: bonus,
  };
}
