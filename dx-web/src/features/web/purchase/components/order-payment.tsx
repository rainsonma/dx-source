"use client";

import { useEffect, useState } from "react";
import { Check, ScanLine, Clock, AlertCircle } from "lucide-react";
import { orderApi } from "@/lib/api-client";
import { ORDER_STATUSES } from "@/consts/order-status";
import { ORDER_TYPES } from "@/consts/order-type";
import { USER_GRADE_LABELS, type UserGrade } from "@/consts/user-grade";
import { BEAN_PACKAGES } from "@/consts/bean-package";
import {
  PAYMENT_METHODS,
  PAYMENT_METHOD_LABELS,
  type PaymentMethod,
} from "@/consts/payment-method";
import type { Order } from "@/features/web/purchase/types/order.types";
import { AgreementLink } from "@/features/com/legal/components/agreement-link";

function getProductLabel(order: Order): string {
  if (order.type === ORDER_TYPES.MEMBERSHIP) {
    return USER_GRADE_LABELS[order.product as UserGrade] ?? order.product;
  }
  const pkg = BEAN_PACKAGES.find((p) => p.slug === order.product);
  if (pkg) {
    const total = pkg.beans + pkg.bonus;
    return `${total.toLocaleString()} 能量豆`;
  }
  return order.product;
}

function formatAmount(fen: number): string {
  return `\u00a5${(fen / 100).toFixed(2)}`;
}

function useCountdown(expiresAt: string): string {
  const [remaining, setRemaining] = useState("");

  useEffect(() => {
    function update() {
      const diff = new Date(expiresAt).getTime() - Date.now();
      if (diff <= 0) {
        setRemaining("已过期");
        return;
      }
      const mins = Math.floor(diff / 60000);
      const secs = Math.floor((diff % 60000) / 1000);
      setRemaining(`${mins}:${secs.toString().padStart(2, "0")}`);
    }
    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, [expiresAt]);

  return remaining;
}

export function OrderPayment({ orderId }: { orderId: string }) {
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [agreed, setAgreed] = useState(false);
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod>(
    PAYMENT_METHODS.WECHAT,
  );

  useEffect(() => {
    orderApi.getOrder(orderId).then((res) => {
      if (res.code === 0 && res.data) {
        setOrder(res.data as Order);
        if (res.data.paymentMethod) {
          setSelectedMethod(res.data.paymentMethod as PaymentMethod);
        }
      } else {
        setError("订单不存在");
      }
    });
  }, [orderId]);

  const countdown = useCountdown(order?.expiresAt ?? "");

  if (error) {
    return (
      <div className="flex flex-col items-center gap-3 py-20">
        <AlertCircle className="h-10 w-10 text-red-400" />
        <span className="text-base text-slate-600">{error}</span>
      </div>
    );
  }

  if (!order) {
    return (
      <div className="flex items-center justify-center py-20">
        <span className="text-sm text-slate-400">加载中...</span>
      </div>
    );
  }

  if (order.status !== ORDER_STATUSES.PENDING) {
    const statusMessages: Record<string, string> = {
      [ORDER_STATUSES.PAID]: "订单已支付，正在处理中...",
      [ORDER_STATUSES.FULFILLED]: "订单已完成",
      [ORDER_STATUSES.EXPIRED]: "订单已过期，请重新下单",
      [ORDER_STATUSES.CANCELLED]: "订单已取消",
    };
    return (
      <div className="flex flex-col items-center gap-3 py-20">
        <AlertCircle className="h-10 w-10 text-slate-400" />
        <span className="text-base text-slate-600">
          {statusMessages[order.status] ?? "订单状态异常"}
        </span>
      </div>
    );
  }

  const methods: { key: PaymentMethod; color: string; icon: string }[] = [
    { key: PAYMENT_METHODS.WECHAT, color: "bg-[#07C160]", icon: "W" },
    { key: PAYMENT_METHODS.ALIPAY, color: "bg-[#1677FF]", icon: "A" },
  ];

  return (
    <div className="flex w-full max-w-[520px] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
      {/* Order summary */}
      <div className="flex flex-col gap-3.5 px-7 py-6">
        <div className="flex items-center justify-between">
          <span className="text-base font-bold text-slate-900">
            {getProductLabel(order)}
          </span>
          <div className="flex items-center gap-1 text-sm text-slate-500">
            <Clock className="h-3.5 w-3.5" />
            <span>{countdown}</span>
          </div>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-slate-400">订单编号</span>
          <span className="text-xs text-slate-500">
            {order.id.slice(0, 18)}
          </span>
        </div>
        <div className="flex items-center gap-0.5">
          <span className="text-4xl font-bold text-teal-600">
            {formatAmount(order.amount)}
          </span>
        </div>
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* Agreement */}
      <div className="px-7 py-4">
        <label className="flex cursor-pointer gap-2.5">
          <button
            type="button"
            onClick={() => setAgreed(!agreed)}
            className={`flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-[1.5px] ${
              agreed
                ? "border-teal-600 bg-teal-600"
                : "border-slate-300 bg-white"
            }`}
          >
            {agreed && <Check className="h-2.5 w-2.5 text-white" />}
          </button>
          <div className="flex flex-col gap-1">
            <span className="text-sm text-slate-700">
              我已阅读并同意以下协议
            </span>
            <span className="text-xs">
              <AgreementLink slug="product-service" />
            </span>
          </div>
        </label>
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* Payment method */}
      <div className="flex flex-col gap-4 px-7 py-4">
        {methods.map((m) => (
          <button
            key={m.key}
            type="button"
            onClick={() => setSelectedMethod(m.key)}
            className="flex items-center gap-2.5"
          >
            <div
              className={`flex h-[18px] w-[18px] items-center justify-center rounded-full ${
                selectedMethod === m.key
                  ? "bg-teal-600"
                  : "border-[1.5px] border-slate-300 bg-white"
              }`}
            >
              {selectedMethod === m.key && (
                <Check className="h-2.5 w-2.5 text-white" />
              )}
            </div>
            <div
              className={`flex h-[22px] w-[22px] items-center justify-center rounded-[5px] ${m.color}`}
            >
              <span className="text-[9px] font-bold text-white">
                {m.icon}
              </span>
            </div>
            <span
              className={`text-sm ${
                selectedMethod === m.key
                  ? "font-medium text-slate-900"
                  : "text-slate-600"
              }`}
            >
              {PAYMENT_METHOD_LABELS[m.key]}
            </span>
          </button>
        ))}
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* QR code placeholder */}
      <div className="flex flex-col items-center gap-3 px-7 py-6">
        <div className="flex h-[180px] w-[180px] items-center justify-center rounded-lg border border-slate-200 bg-slate-50">
          <div className="flex flex-col items-center gap-2">
            <ScanLine className="h-8 w-8 text-slate-300" />
            <span className="text-xs text-slate-400">支付功能即将开放</span>
          </div>
        </div>
        <span className="text-[13px] text-slate-500">
          {PAYMENT_METHOD_LABELS[selectedMethod]}
          {" \u2014 扫码支付"}
        </span>
      </div>
    </div>
  );
}
