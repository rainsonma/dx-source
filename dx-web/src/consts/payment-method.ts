export const PAYMENT_METHODS = {
  WECHAT: "wechat",
  ALIPAY: "alipay",
} as const;

export type PaymentMethod =
  (typeof PAYMENT_METHODS)[keyof typeof PAYMENT_METHODS];

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  wechat: "微信支付",
  alipay: "支付宝",
};
