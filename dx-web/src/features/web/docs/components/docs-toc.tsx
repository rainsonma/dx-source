"use client";

import { useEffect, useState } from "react";

type Heading = {
  id: string;
  label: string;
};

export function DocsToc() {
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    const h2s = Array.from(
      document.querySelectorAll<HTMLHeadingElement>("main h2[id]"),
    );
    const hs: Heading[] = h2s.map((h) => ({
      id: h.id,
      label: h.textContent ?? "",
    }));
    // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: reading DOM headings after mount is a one-shot sync from the external system (DOM) into component state
    setHeadings(hs);

    if (hs.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort(
            (a, b) => a.boundingClientRect.top - b.boundingClientRect.top,
          );
        if (visible.length > 0) {
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: "0px 0px -70% 0px", threshold: 0 },
    );

    h2s.forEach((h) => observer.observe(h));
    return () => observer.disconnect();
  }, []);

  if (headings.length === 0) return null;

  return (
    <aside className="hidden w-[220px] shrink-0 border-l border-slate-200 px-5 py-10 xl:block">
      <div className="sticky top-4 flex flex-col gap-3">
        <span className="text-xs font-semibold tracking-wide text-slate-900">
          本页目录
        </span>
        <div className="mt-1 flex flex-col gap-3">
          {headings.map((h) => (
            <a
              key={h.id}
              href={`#${h.id}`}
              className="flex items-center gap-2 text-left"
            >
              <div
                className={`h-4 w-0.5 rounded-sm ${
                  activeId === h.id ? "bg-teal-600" : "bg-transparent"
                }`}
                aria-hidden="true"
              />
              <span
                className={`text-[13px] ${
                  activeId === h.id
                    ? "font-medium text-teal-600"
                    : "text-slate-500"
                }`}
              >
                {h.label}
              </span>
            </a>
          ))}
        </div>
      </div>
    </aside>
  );
}
