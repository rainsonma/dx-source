import type { Metadata } from "next";
import { DocsHomeHero } from "@/features/web/docs/components/docs-home-hero";
import { DocsCategoryGrid } from "@/features/web/docs/components/docs-category-grid";

export const metadata: Metadata = {
  title: "斗学帮助中心",
  description: "了解每一项功能的详细用法，让学习更顺畅。",
};

export default function DocsLandingPage() {
  return (
    <>
      <DocsHomeHero />
      <DocsCategoryGrid />
    </>
  );
}
