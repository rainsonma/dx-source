import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { DocsBreadcrumb } from "@/features/web/docs/components/docs-breadcrumb";
import { DocsPrevNext } from "@/features/web/docs/components/docs-prev-next";
import { DocsToc } from "@/features/web/docs/components/docs-toc";
import { DOC_CATEGORIES, findTopic } from "@/features/web/docs/registry";

type Params = { category: string; topic: string };

export function generateStaticParams(): Params[] {
  return DOC_CATEGORIES.flatMap((c) =>
    c.topics.map((t) => ({ category: c.slug, topic: t.slug })),
  );
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { category, topic } = await params;
  const found = findTopic(category, topic);
  if (!found) return { title: "未找到 — 斗学帮助中心" };
  return {
    title: `${found.ref.topic.title} — 斗学帮助中心`,
    description: found.ref.topic.description,
  };
}

export default async function TopicPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { category, topic } = await params;
  const found = findTopic(category, topic);
  if (!found) notFound();
  const { ref, prev, next } = found;
  const TopicComponent = ref.topic.Component;

  return (
    <div className="flex gap-8">
      <div className="flex flex-1 flex-col gap-8">
        <DocsBreadcrumb category={ref.category} topic={ref.topic} />
        <div className="flex flex-col gap-3">
          <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            {ref.topic.title}
          </h1>
          <p className="text-base leading-relaxed text-slate-500">
            {ref.topic.description}
          </p>
        </div>
        <div className="h-px w-full bg-slate-200" />
        <TopicComponent />
        <div className="h-px w-full bg-slate-200" />
        <DocsPrevNext prev={prev} next={next} />
      </div>
      <DocsToc />
    </div>
  );
}
