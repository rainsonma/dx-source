import Link from "next/link";
import type { DocCategory, DocTopic } from "@/features/web/docs/types";

type Props = {
  category: DocCategory;
  topic?: DocTopic;
};

export function DocsBreadcrumb({ category, topic }: Props) {
  return (
    <div className="flex items-center gap-2 text-[13px]">
      <Link href="/wiki" className="text-slate-400 hover:text-slate-600">
        Wiki
      </Link>
      <span className="text-slate-300">/</span>
      {topic ? (
        <>
          <Link
            href={`/wiki/${category.slug}`}
            className="text-slate-400 hover:text-slate-600"
          >
            {category.title}
          </Link>
          <span className="text-slate-300">/</span>
          <span className="font-medium text-slate-900">{topic.title}</span>
        </>
      ) : (
        <span className="font-medium text-slate-900">{category.title}</span>
      )}
    </div>
  );
}
