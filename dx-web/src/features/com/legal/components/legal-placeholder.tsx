import type { ReactNode } from "react";

export function LegalPlaceholder({ children }: { children: ReactNode }) {
  return (
    <span className="rounded bg-amber-50 px-1 py-0.5 font-mono text-[13px] text-amber-700">
      {children}
    </span>
  );
}
