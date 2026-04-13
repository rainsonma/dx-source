"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { DOC_CATEGORIES } from "@/features/web/docs/registry";

export function DocsSidebar() {
  const pathname = usePathname();

  return (
    <div className="flex flex-col gap-1">
      <Link
        href="/docs"
        className="mb-3 text-sm font-extrabold tracking-tight text-slate-900 hover:text-teal-600"
      >
        斗学文档
      </Link>
      {DOC_CATEGORIES.map((category, ci) => {
        const catHref = `/docs/${category.slug}`;
        const isActiveCat = pathname?.startsWith(catHref) ?? false;
        const CatIcon = category.icon;
        return (
          <div key={category.slug}>
            {ci > 0 && <div className="my-2 h-px w-full bg-slate-200" />}
            <Link
              href={catHref}
              className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm font-semibold ${
                isActiveCat ? "text-teal-600" : "text-slate-900"
              }`}
            >
              <CatIcon
                className={`h-4 w-4 ${
                  isActiveCat ? "text-teal-600" : "text-slate-400"
                }`}
                aria-hidden="true"
              />
              {category.title}
            </Link>
            <div className="mt-1 flex flex-col gap-1">
              {category.topics.map((topic) => {
                const topicHref = `/docs/${category.slug}/${topic.slug}`;
                const isActiveTopic = pathname === topicHref;
                return (
                  <Link
                    key={topic.slug}
                    href={topicHref}
                    className={`flex items-center gap-2 rounded-md px-3 py-2 text-left text-sm ${
                      isActiveTopic
                        ? "bg-teal-50 font-medium text-teal-600"
                        : "text-slate-500 hover:text-slate-700"
                    }`}
                  >
                    <div
                      className={`h-5 w-[3px] rounded-sm ${
                        isActiveTopic ? "bg-teal-600" : "bg-transparent"
                      }`}
                      aria-hidden="true"
                    />
                    {topic.title}
                  </Link>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
