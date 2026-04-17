"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { getAgreementBySlug } from "@/features/com/legal/registry";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  slug: LegalAgreementSlug;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function AgreementDialog({ slug, open, onOpenChange }: Props) {
  const agreement = getAgreementBySlug(slug);
  const Body = agreement.Component;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="flex h-[85vh] max-h-[85vh] w-full max-w-[960px] flex-col gap-0 overflow-hidden p-0 sm:max-w-[960px]"
        showCloseButton
      >
        <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-6 py-4">
          <div className="flex flex-col gap-1">
            <DialogTitle className="text-lg font-bold text-slate-900">
              {agreement.title}
            </DialogTitle>
            <DialogDescription className="text-xs text-slate-500">
              生效日期：{agreement.effectiveDate} · 最近更新：{agreement.lastUpdated}
            </DialogDescription>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto px-6 py-5">
          <div className="flex flex-col gap-6 text-[14px] leading-[1.75] text-slate-700">
            <Body />
          </div>
        </div>
        <div className="flex items-center justify-end border-t border-slate-200 bg-slate-50 px-6 py-3">
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="h-9 rounded-md bg-teal-600 px-4 text-sm font-semibold text-white hover:bg-teal-700"
          >
            我已阅读
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
