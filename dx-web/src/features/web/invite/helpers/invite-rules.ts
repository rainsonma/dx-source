export const inviteRules: string[] = [
  "邀请好友通过您的专属链接、邀请码或二维码注册斗学账号",
  "好友成功注册并完成首次购买会员即算邀请成功",
  "佣金数额实时反馈，邀请所得收入清晰体现，一目了然",
  "邀请人数不设上限，邀请越多佣金越多",
  "佣金收入按时结算，可随时申请提现",
];

export type RewardValue =
  | { kind: "fixed"; amount: number }
  | { kind: "percent"; value: number };

export const commissionRewardKeys = [
  "lifetime",
  "year",
  "season",
  "month",
  "renewal",
] as const;

export type CommissionRewardKey = (typeof commissionRewardKeys)[number];

export const rewardRowLabels: Record<CommissionRewardKey, string> = {
  lifetime: "邀请永久会员",
  year: "邀请年度会员",
  season: "邀请季度会员",
  month: "邀请月度会员",
  renewal: "持续续费返佣",
};

export type InviterTierId = "standard" | "lifetime";

export type CommissionTier = {
  id: InviterTierId;
  label: string;
  sublabel: string;
  rewards: Record<CommissionRewardKey, RewardValue>;
};

export const commissionTiers: CommissionTier[] = [
  {
    id: "standard",
    label: "普通付费会员",
    sublabel: "月度 / 季度 / 年度",
    rewards: {
      lifetime: { kind: "fixed", amount: 500 },
      year: { kind: "percent", value: 30 },
      season: { kind: "percent", value: 30 },
      month: { kind: "percent", value: 30 },
      renewal: { kind: "percent", value: 10 },
    },
  },
  {
    id: "lifetime",
    label: "终身会员",
    sublabel: "永久会员",
    rewards: {
      lifetime: { kind: "fixed", amount: 600 },
      year: { kind: "percent", value: 50 },
      season: { kind: "percent", value: 50 },
      month: { kind: "percent", value: 50 },
      renewal: { kind: "percent", value: 20 },
    },
  },
];

export type InviteeDiscountGrade = "lifetime" | "year" | "season" | "month";

export type InviteeDiscount = {
  grade: InviteeDiscountGrade;
  label: string;
  value: RewardValue;
};

export const inviteeDiscounts: InviteeDiscount[] = [
  { grade: "lifetime", label: "购买永久会员", value: { kind: "fixed", amount: 99 } },
  { grade: "year", label: "购买年度会员", value: { kind: "percent", value: 10 } },
  { grade: "season", label: "购买季度会员", value: { kind: "percent", value: 10 } },
  { grade: "month", label: "购买月度会员", value: { kind: "percent", value: 10 } },
];

export function formatRewardValue(value: RewardValue): string {
  switch (value.kind) {
    case "fixed":
      return `¥${value.amount}`;
    case "percent":
      return `${value.value}%`;
  }
}
