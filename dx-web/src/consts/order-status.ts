export const ORDER_STATUSES = {
  PENDING: "pending",
  PAID: "paid",
  FULFILLED: "fulfilled",
  EXPIRED: "expired",
  CANCELLED: "cancelled",
} as const;

export type OrderStatus = (typeof ORDER_STATUSES)[keyof typeof ORDER_STATUSES];

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: "待支付",
  paid: "已支付",
  fulfilled: "已完成",
  expired: "已过期",
  cancelled: "已取消",
};
