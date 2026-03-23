"use client";

import { useState, useTransition } from "react";
import { Ticket, ShoppingCart, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { redeemCodeAction } from "@/features/web/redeem/actions/redeem.action";
import { swrMutate } from "@/lib/swr";

/** Card with code input field and redeem button */
export function RedeemInputCard() {
  const [code, setCode] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Handle redeem submission */
  const handleRedeem = () => {
    if (!code.trim()) return;

    startTransition(async () => {
      const result = await redeemCodeAction({ code: code.trim() });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      toast.success("兑换成功");
      setCode("");
      swrMutate("/api/redeems", "/api/admin/redeems");
    });
  };

  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-border bg-card p-5 shadow-sm md:p-7">
      <div className="flex flex-col gap-1.5">
        <span className="text-base font-semibold text-foreground">输入兑换码</span>
        <span className="text-[13px] text-muted-foreground">
          兑换码区分大小写，请按格式正确输入（如：ABCD-1234-EFGH-5678）
        </span>
      </div>
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="flex h-12 flex-1 items-center rounded-[10px] border border-border bg-muted px-4">
          <Ticket className="mr-2.5 h-[18px] w-[18px] text-muted-foreground" />
          <input
            type="text"
            placeholder="请输入兑换码"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            disabled={isPending}
            maxLength={19}
            className="flex-1 bg-transparent text-[13px] text-foreground outline-none placeholder:text-muted-foreground disabled:opacity-50"
          />
        </div>
        <button
          type="button"
          onClick={handleRedeem}
          disabled={isPending || !code.trim()}
          className="flex h-12 items-center justify-center gap-1.5 rounded-[10px] bg-teal-600 px-6 text-sm font-semibold text-white hover:bg-teal-700 disabled:opacity-50"
        >
          {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          立即兑换
        </button>
      </div>
      <div className="flex items-center justify-center gap-2">
        <span className="text-[13px] text-muted-foreground">没有兑换码？</span>
        <button
          type="button"
          className="flex items-center gap-1.5 rounded-lg bg-teal-50 px-4 py-2 text-[13px] font-medium text-teal-600"
        >
          <ShoppingCart className="h-3.5 w-3.5" />
          购买会员
        </button>
      </div>
    </div>
  );
}
