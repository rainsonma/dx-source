"use client";

import { Gem } from "lucide-react";

interface BeanPackageCardProps {
  beans: number;
  bonus: number;
  price: number;
  tag?: string;
  highlight?: boolean;
  disabled?: boolean;
  onPurchase?: () => void;
}

export function BeanPackageCard({
  beans,
  bonus,
  price,
  tag,
  highlight,
  disabled,
  onPurchase,
}: BeanPackageCardProps) {
  const priceYuan = price / 100;
  const totalDisplay = (beans + bonus).toLocaleString();

  return (
    <div
      className={`relative flex min-w-[180px] flex-1 flex-col items-center gap-3 rounded-[14px] border p-5 ${
        highlight
          ? "border-orange-400 bg-slate-50 shadow-[0_0_0_1px_rgba(251,146,60,0.3)]"
          : "border-slate-200 bg-slate-50"
      }`}
    >
      {tag && (
        <span
          className={`absolute -top-0 right-3 rounded-b-lg px-2.5 py-1 text-[11px] font-bold text-white ${
            highlight
              ? "bg-gradient-to-r from-orange-400 to-red-500"
              : "bg-gradient-to-r from-violet-500 to-pink-500"
          }`}
        >
          {tag}
        </span>
      )}

      <Gem className="h-10 w-10 text-blue-400" />

      <div className="flex flex-col items-center gap-0.5">
        <span className="text-2xl font-extrabold text-slate-900">
          {totalDisplay}
        </span>
        {bonus > 0 && (
          <span className="text-xs text-green-600">
            +{bonus.toLocaleString()} 赠送
          </span>
        )}
        <span className="text-xs text-slate-500">能量豆</span>
      </div>

      <span className="text-xl font-bold text-orange-500">¥{priceYuan}</span>

      <button
        type="button"
        disabled={disabled}
        onClick={onPurchase}
        className={`w-full rounded-[10px] py-2.5 text-center text-sm font-semibold ${
          highlight
            ? "bg-gradient-to-r from-orange-400 to-red-500 text-white hover:from-orange-500 hover:to-red-600"
            : "border border-slate-300 bg-white text-slate-700 hover:bg-slate-100"
        }`}
      >
        立即购买
      </button>
    </div>
  );
}
