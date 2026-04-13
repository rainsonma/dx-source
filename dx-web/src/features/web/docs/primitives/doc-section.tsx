import type { ReactNode } from "react";

type Props = {
  id: string;
  title: string;
  children: ReactNode;
};

export function DocSection({ id, title, children }: Props) {
  return (
    <section className="flex flex-col gap-4">
      <h2
        id={id}
        className="scroll-mt-24 text-xl font-bold tracking-tight text-slate-900 md:text-2xl"
      >
        {title}
      </h2>
      <div className="flex flex-col gap-4 text-[15px] leading-[1.7] text-slate-600">
        {children}
      </div>
    </section>
  );
}
