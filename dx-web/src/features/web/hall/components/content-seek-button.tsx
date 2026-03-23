"use client";

import { useState } from "react";
import { Lightbulb } from "lucide-react";
import { ContentSeekModal } from "@/features/web/hall/components/content-seek-modal";

/** Client button that opens the content seek modal */
export function ContentSeekButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex h-10 items-center gap-1.5 rounded-[10px] border border-border bg-card px-3.5 text-muted-foreground hover:bg-accent"
      >
        <Lightbulb className="h-3.5 w-3.5" />
        <span className="text-[13px] font-semibold">求课程</span>
      </button>
      <ContentSeekModal open={open} onOpenChange={setOpen} />
    </>
  );
}
