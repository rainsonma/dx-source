"use client";

import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface GroupStatRowProps {
  icon: LucideIcon;
  iconClass: string;
  label: string;
  displayText: string | null;
  flashKey: number;
  flashColorClass: string;
}

const whooshStyle = { animation: "whoosh 400ms ease-out forwards" } as const;

/** A stat row for group play — shows the acting player's name with whoosh animation. */
export function GroupStatRow({
  icon: Icon,
  iconClass,
  label,
  displayText,
  flashKey,
  flashColorClass,
}: GroupStatRowProps) {
  return (
    <div className="flex items-center gap-1.5">
      <Icon className={cn("h-3.5 w-3.5 shrink-0", iconClass)} />
      <span className="text-xs text-muted-foreground shrink-0">{label}</span>
      <div className="flex-1 mx-1 h-1 rounded-full bg-border overflow-hidden">
        {flashKey > 0 && (
          <div
            key={flashKey}
            className={cn("h-full rounded-full", flashColorClass)}
            style={whooshStyle}
          />
        )}
      </div>
      <span className="text-xs font-bold shrink-0 text-foreground max-w-16 truncate">
        {displayText ?? "\u2014"}
      </span>
    </div>
  );
}
