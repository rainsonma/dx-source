"use client";

import { useState } from "react";
import { Loader2, TicketPlus } from "lucide-react";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { USER_GRADE_LABELS } from "@/consts/user-grade";

const GRADE_OPTIONS = [
  { value: "month", label: USER_GRADE_LABELS.month },
  { value: "season", label: USER_GRADE_LABELS.season },
  { value: "year", label: USER_GRADE_LABELS.year },
  { value: "lifetime", label: USER_GRADE_LABELS.lifetime },
];

const QUANTITY_OPTIONS = [
  { value: "10", label: "10" },
  { value: "50", label: "50" },
  { value: "100", label: "100" },
  { value: "500", label: "500" },
];

type GenerateCodesModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onGenerate: (input: { grade: string; quantity: string }) => Promise<boolean>;
};

/** Modal form for generating redeem codes (admin only) */
export function GenerateCodesModal({ open, onOpenChange, onGenerate }: GenerateCodesModalProps) {
  const [grade, setGrade] = useState(GRADE_OPTIONS[0].value);
  const [quantity, setQuantity] = useState(QUANTITY_OPTIONS[0].value);
  const [submitting, setSubmitting] = useState(false);

  /** Handle generation submit */
  async function handleSubmit() {
    setSubmitting(true);
    const ok = await onGenerate({ grade, quantity });
    setSubmitting(false);
    if (ok) {
      onOpenChange(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[520px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 px-7 pt-7 pb-6">
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-foreground">
            <TicketPlus className="h-[18px] w-[18px] text-teal-600" />
            生成兑换码
          </DialogTitle>
          <DialogDescription className="sr-only">
            选择兑换码类型和数量
          </DialogDescription>

          <div className="h-px bg-border" />

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <label htmlFor="generate-grade" className="text-[13px] font-medium text-foreground">
                生成类型 *
              </label>
              <select
                id="generate-grade"
                value={grade}
                onChange={(e) => setGrade(e.target.value)}
                disabled={submitting}
                className="h-10 rounded-lg border border-border bg-card px-3.5 text-[13px] text-foreground outline-none transition-colors focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              >
                {GRADE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-2">
              <label htmlFor="generate-quantity" className="text-[13px] font-medium text-foreground">
                生成数量 *
              </label>
              <select
                id="generate-quantity"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                disabled={submitting}
                className="h-10 rounded-lg border border-border bg-card px-3.5 text-[13px] text-foreground outline-none transition-colors focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              >
                {QUANTITY_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2.5">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
              className="rounded-lg border border-border px-4 py-2 text-[13px] font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting}
              className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {submitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              确认生成
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
