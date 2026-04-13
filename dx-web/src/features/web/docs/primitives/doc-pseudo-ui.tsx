import type { ReactNode } from "react";

type Props = {
  label?: string;
  children: ReactNode;
};

export function DocPseudoUI({ label, children }: Props) {
  return (
    <div className="relative rounded-[10px] border border-slate-200 bg-slate-50 p-5">
      {label && (
        <span className="absolute -top-2.5 left-4 rounded bg-white px-2 text-[11px] font-medium text-slate-500">
          {label}
        </span>
      )}
      <div className="flex flex-col gap-2 text-sm text-slate-700">
        {children}
      </div>
    </div>
  );
}
