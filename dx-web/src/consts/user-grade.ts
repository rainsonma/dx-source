export const USER_GRADES = {
  FREE: "free",
  MONTH: "month",
  SEASON: "season",
  YEAR: "year",
  LIFETIME: "lifetime",
} as const;

export type UserGrade = (typeof USER_GRADES)[keyof typeof USER_GRADES];

export const USER_GRADE_PRICES: Record<UserGrade, number> = {
  free: 0,
  month: 39,
  season: 99,
  year: 309,
  lifetime: 1999,
};

export const USER_GRADE_LABELS: Record<UserGrade, string> = {
  free: "免费会员",
  month: "月度会员",
  season: "季度会员",
  year: "年度会员",
  lifetime: "终身会员",
};

/** Number of months each grade adds when redeemed */
export const USER_GRADE_MONTHS: Record<UserGrade, number | null> = {
  free: null,
  month: 1,
  season: 3,
  year: 12,
  lifetime: null,
};
