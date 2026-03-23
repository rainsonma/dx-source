"use client";

import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { useValueFlash } from "@/features/web/play/hooks/use-value-flash";

interface StatRowProps {
  icon: LucideIcon;
  iconClass: string;
  label: string;
  value: number;
  valueClass: string;
  flashColorClass: string;
}

const whooshStyle = { animation: "whoosh 400ms ease-out forwards" } as const;

/** A single stat row with a decorative whoosh progress bar. */
export function StatRow({
  icon: Icon,
  iconClass,
  label,
  value,
  valueClass,
  flashColorClass,
}: StatRowProps) {
  const { flashKey } = useValueFlash(value);

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
      <span className={cn("text-xs font-bold shrink-0", valueClass)}>{value}</span>
    </div>
  );
}
