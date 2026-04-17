"use client";

import { useState } from "react";
import { AgreementDialog } from "./agreement-dialog";
import { getAgreementBySlug } from "@/features/com/legal/registry";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  slug: LegalAgreementSlug;
  className?: string;
};

export function AgreementLink({ slug, className }: Props) {
  const [open, setOpen] = useState(false);
  const agreement = getAgreementBySlug(slug);
  return (
    <>
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          setOpen(true);
        }}
        className={
          className ??
          "text-teal-600 underline-offset-2 hover:text-teal-700 hover:underline"
        }
      >
        {agreement.shortTitle}
      </button>
      <AgreementDialog slug={slug} open={open} onOpenChange={setOpen} />
    </>
  );
}
