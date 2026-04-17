import Link from "next/link";

export function DocsHomeHero() {
  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
        斗学帮助中心
      </h1>
      <p className="text-base leading-relaxed text-slate-500">
        了解每一项功能的详细用法，让学习更顺畅。
      </p>
      <div className="flex flex-wrap gap-3">
        <Link
          href="/wiki/getting-started/what-is-douxue"
          className="inline-flex items-center gap-2 rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white hover:bg-teal-700"
        >
          快速开始
        </Link>
        <Link
          href="/wiki/account/faq"
          className="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
        >
          常见问题
        </Link>
      </div>
    </div>
  );
}
