import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { DocsCategoryIndex } from "@/features/web/docs/components/docs-category-index";
import { DOC_CATEGORIES, findCategory } from "@/features/web/docs/registry";

type Params = { category: string };

export function generateStaticParams(): Params[] {
  return DOC_CATEGORIES.map((c) => ({ category: c.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { category } = await params;
  const found = findCategory(category);
  if (!found) return { title: "未找到 — 斗学帮助中心" };
  return {
    title: `${found.title} — 斗学帮助中心`,
    description: found.description,
  };
}

export default async function CategoryPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { category } = await params;
  const found = findCategory(category);
  if (!found) notFound();
  return <DocsCategoryIndex category={found} />;
}
