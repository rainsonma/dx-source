export const ORDER_TYPES = {
  MEMBERSHIP: "membership",
  BEANS: "beans",
} as const;

export type OrderType = (typeof ORDER_TYPES)[keyof typeof ORDER_TYPES];
