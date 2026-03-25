"use client";

import { Loader2 } from "lucide-react";

export function GroupWaitingScreen() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <p className="text-lg text-muted-foreground">等待其他选手完成...</p>
    </div>
  );
}
