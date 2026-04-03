import type { OrderType } from "@/consts/order-type";
import type { OrderStatus } from "@/consts/order-status";
import type { PaymentMethod } from "@/consts/payment-method";

export type Order = {
  id: string;
  type: OrderType;
  product: string;
  amount: number;
  status: OrderStatus;
  paymentMethod: PaymentMethod | null;
  expiresAt: string;
  createdAt: string;
};
