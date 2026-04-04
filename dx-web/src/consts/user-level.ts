export const MAX_LEVEL = 100;

const BASE_EXP = 100;
const MULTIPLIER = 1.05;

/** Flat EXP cost for Lv.0 → Lv.1, separate from the exponential curve. */
const INTRO_EXP = 100;

export type UserLevel = {
  level: number;
  expRequired: number;
};

function generateLevels(): UserLevel[] {
  const levels: UserLevel[] = [
    { level: 0, expRequired: 0 },
    { level: 1, expRequired: INTRO_EXP },
  ];
  let cumulative = INTRO_EXP;

  for (let i = 2; i <= MAX_LEVEL; i++) {
    cumulative += Math.floor(BASE_EXP * Math.pow(MULTIPLIER, i - 2));
    levels.push({ level: i, expRequired: cumulative });
  }

  return levels;
}

export const USER_LEVELS: UserLevel[] = generateLevels();

export function getLevel(exp: number): number {
  if (exp < 0) {
    throw new Error("exp must be non-negative");
  }
  for (let i = USER_LEVELS.length - 1; i >= 0; i--) {
    if (exp >= USER_LEVELS[i].expRequired) {
      return USER_LEVELS[i].level;
    }
  }
  // Unreachable: USER_LEVELS[0].expRequired is always 0 and exp is non-negative.
  return 0;
}

export function getExpForLevel(level: number): number {
  if (!Number.isInteger(level) || level < 0 || level > MAX_LEVEL) {
    throw new Error(`Level must be an integer between 0 and ${MAX_LEVEL}`);
  }
  return USER_LEVELS[level].expRequired;
}
