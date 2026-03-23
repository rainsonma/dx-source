"use client";

import { useState } from "react";
import { Flag } from "lucide-react";
import { FeedbackModal } from "@/features/web/hall/components/feedback-modal";

/** Client button that opens the feedback modal */
export function FeedbackButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-border bg-card text-muted-foreground hover:bg-accent"
      >
        <Flag className="h-[18px] w-[18px]" />
      </button>
      <FeedbackModal open={open} onOpenChange={setOpen} />
    </>
  );
}
