import Link from "next/link";
import { DOC_CATEGORIES } from "@/features/web/docs/registry";

export function DocsCategoryGrid() {
  return (
    <div className="flex flex-col gap-4">
      <h2 className="text-xl font-bold tracking-tight text-slate-900 md:text-2xl">
        全部分类
      </h2>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {DOC_CATEGORIES.map((category) => {
          const Icon = category.icon;
          return (
            <Link
              key={category.slug}
              href={`/wiki/${category.slug}`}
              className="flex flex-col gap-2.5 rounded-[10px] border border-slate-200 bg-white p-5 hover:border-slate-300"
            >
              <div
                className={`flex h-10 w-10 items-center justify-center rounded-lg border ${category.accentClass}`}
              >
                <Icon className="h-5 w-5" aria-hidden="true" />
              </div>
              <span className="text-[15px] font-semibold text-slate-900">
                {category.title}
              </span>
              <span className="text-[13px] leading-snug text-slate-500">
                {category.description}
              </span>
              <span className="mt-auto pt-2 text-[12px] text-slate-400">
                {category.topics.length} 个主题 →
              </span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
