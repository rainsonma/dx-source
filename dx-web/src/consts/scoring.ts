export const SCORING = {
  CORRECT_ANSWER: 1,
  COMBO_3_BONUS: 3,
  COMBO_5_BONUS: 5,
  COMBO_10_BONUS: 10,
  COMBO_CYCLE_LENGTH: 10,
  LEVEL_COMPLETE_EXP: 10,
  EXP_ACCURACY_THRESHOLD: 0.6,
} as const;

export const COMBO_THRESHOLDS = [
  { streak: 3, bonus: SCORING.COMBO_3_BONUS },
  { streak: 5, bonus: SCORING.COMBO_5_BONUS },
  { streak: 10, bonus: SCORING.COMBO_10_BONUS },
] as const;
