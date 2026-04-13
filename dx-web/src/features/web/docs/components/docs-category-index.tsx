import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { DocsBreadcrumb } from "./docs-breadcrumb";
import type { DocCategory } from "@/features/web/docs/types";

type Props = { category: DocCategory };

export function DocsCategoryIndex({ category }: Props) {
  const Icon = category.icon;
  return (
    <>
      <DocsBreadcrumb category={category} />
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-3">
          <div
            className={`flex h-12 w-12 items-center justify-center rounded-lg border ${category.accentClass}`}
          >
            <Icon className="h-6 w-6" aria-hidden="true" />
          </div>
          <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            {category.title}
          </h1>
        </div>
        <p className="text-base leading-relaxed text-slate-500">
          {category.description}
        </p>
      </div>
      <div className="h-px w-full bg-slate-200" />
      <div className="flex flex-col gap-3">
        {category.topics.map((topic, i) => (
          <Link
            key={topic.slug}
            href={`/docs/${category.slug}/${topic.slug}`}
            className="group flex items-center gap-4 rounded-[10px] border border-slate-200 bg-white p-5 hover:border-slate-300"
          >
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-100 text-sm font-bold text-slate-600 group-hover:bg-teal-50 group-hover:text-teal-600">
              {i + 1}
            </div>
            <div className="flex flex-1 flex-col gap-1">
              <span className="text-[15px] font-semibold text-slate-900">
                {topic.title}
              </span>
              <span className="text-[13px] text-slate-500">
                {topic.description}
              </span>
            </div>
            <ChevronRight
              className="h-4 w-4 shrink-0 text-slate-400 transition-transform group-hover:translate-x-1"
              aria-hidden="true"
            />
          </Link>
        ))}
      </div>
    </>
  );
}
