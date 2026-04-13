import type { ReactNode } from "react";

export function DocSlug({ children }: { children: ReactNode }) {
  return (
    <code className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[13px] text-slate-700">
      {children}
    </code>
  );
}
