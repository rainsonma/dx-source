import {
  Lightbulb,
  Gamepad2,
  Brain,
  Users,
  ChevronRight,
} from "lucide-react";

const sidebarSections = [
  {
    title: "快速开始",
    items: [
      { label: "简介", active: true },
      { label: "安装与配置", active: false },
      { label: "项目结构", active: false },
      { label: "第一个课程", active: false },
    ],
  },
  {
    title: "核心概念",
    items: [
      { label: "课程体系", active: false },
      { label: "游戏模式", active: false },
      { label: "AI 练习", active: false },
      { label: "词汇系统", active: false },
      { label: "社区功能", active: false },
    ],
  },
  {
    title: "API 参考",
    items: [
      { label: "认证接口", active: false },
      { label: "用户管理", active: false },
      { label: "课程数据", active: false },
      { label: "Webhook", active: false },
    ],
  },
];

const tocItems = [
  { label: "什么是斗学？", active: true },
  { label: "核心功能", active: false },
  { label: "快速上手", active: false },
  { label: "示例代码", active: false },
];

const features = [
  { icon: Gamepad2, iconColor: "text-teal-600", title: "游戏化学习", desc: "通过多种游戏模式让学习变得有趣高效" },
  { icon: Brain, iconColor: "text-purple-600", title: "AI 驱动", desc: "智能推荐和个性化学习路径定制" },
  { icon: Users, iconColor: "text-amber-500", title: "社区互动", desc: "与学友一起学习，互相激励共同进步" },
];

const steps = [
  { num: 1, title: "注册账户", desc: "在官网注册你的斗学账户，完善个人学习偏好设置" },
  { num: 2, title: "选择课程", desc: "浏览课程列表，选择适合你等级和兴趣的英语课程" },
  { num: 3, title: "开始游戏", desc: "通过词汇对战、连词成句等游戏模式开始你的学习之旅" },
];

const codeLines = [
  { text: "// 初始化斗学 SDK", color: "text-slate-400" },
  { text: "import { DouxueSDK } from '@douxue/sdk';", color: "text-sky-300" },
  { text: "", color: "" },
  { text: "const app = new DouxueSDK({", color: "text-slate-200" },
  { text: "  apiKey: 'your-api-key',", color: "text-amber-200" },
  { text: "  locale: 'zh-CN'", color: "text-amber-200" },
  { text: "});", color: "text-slate-200" },
];

export function DocsPageContent() {
  return (
    <div className="flex flex-1 bg-slate-50">
      <div className="mx-auto flex w-full max-w-[1280px]">
      {/* Left sidebar */}
      <aside className="hidden w-[260px] shrink-0 border-r border-slate-200 bg-white px-5 py-6 lg:block">
        <div className="flex flex-col gap-1">
          {sidebarSections.map((section, si) => (
            <div key={section.title}>
              {si > 0 && (
                <>
                  <div className="my-2 h-px w-full bg-slate-200" />
                  <div className="h-2" />
                </>
              )}
              <span className="text-xs font-semibold tracking-wide text-slate-900">
                {section.title}
              </span>
              <div className="mt-1 flex flex-col gap-1">
                {section.items.map((item) => (
                  <button
                    key={item.label}
                    type="button"
                    className={`flex items-center gap-2 rounded-md px-3 py-2 text-left text-sm ${
                      item.active
                        ? "bg-teal-50 font-medium text-teal-600"
                        : "text-slate-500"
                    }`}
                  >
                    {item.active && (
                      <div className="h-5 w-[3px] rounded-sm bg-teal-600" />
                    )}
                    {item.label}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      </aside>

      {/* Main content */}
      <main className="flex flex-1 flex-col gap-8 px-4 py-6 md:px-8 lg:px-14 lg:py-10">
        {/* Breadcrumb */}
        <div className="flex items-center gap-2 text-[13px]">
          <span className="text-slate-400">文档</span>
          <span className="text-slate-300">/</span>
          <span className="text-slate-400">快速开始</span>
          <span className="text-slate-300">/</span>
          <span className="font-medium text-slate-900">简介</span>
        </div>

        {/* Title */}
        <div className="flex flex-col gap-3">
          <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            简介
          </h1>
          <p className="text-base leading-relaxed text-slate-500">
            欢迎来到斗学文档！了解如何使用我们的平台，将英语学习变成一场充满乐趣的冒险。
          </p>
        </div>

        <div className="h-px w-full bg-slate-200" />

        {/* Section 1 */}
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-bold tracking-tight text-slate-900 md:text-2xl">
            什么是斗学？
          </h2>
          <p className="text-[15px] leading-[1.7] text-slate-600">
            斗学是一款创新的英语学习平台，将游戏化机制与 AI 驱动的个性化学习完美结合。通过词汇对战、连词成句、听力闯关等多种游戏模式，让学习过程变得轻松有趣。
          </p>
          <div className="flex gap-3 rounded-lg border border-teal-200 bg-teal-50 p-4">
            <Lightbulb className="h-5 w-5 shrink-0 text-teal-600" />
            <div className="flex flex-col gap-1">
              <span className="text-[13px] font-semibold text-teal-600">提示</span>
              <span className="text-sm leading-snug text-teal-700">
                斗学支持网页端和移动端，随时随地开始学习。推荐使用 Chrome、Safari 或 Edge 浏览器获得最佳体验。
              </span>
            </div>
          </div>
        </div>

        {/* Section 2 */}
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-bold tracking-tight text-slate-900 md:text-2xl">
            核心功能
          </h2>
          <p className="text-[15px] leading-[1.7] text-slate-600">
            斗学平台包含以下核心功能模块，每个模块都经过精心设计以最大化学习效果：
          </p>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((feat) => (
              <div
                key={feat.title}
                className="flex flex-col gap-2.5 rounded-[10px] border border-slate-200 bg-white p-5"
              >
                <feat.icon className={`h-7 w-7 ${feat.iconColor}`} />
                <span className="text-[15px] font-semibold text-slate-900">{feat.title}</span>
                <span className="text-[13px] leading-snug text-slate-500">{feat.desc}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Section 3 */}
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-bold tracking-tight text-slate-900 md:text-2xl">
            快速上手
          </h2>
          <p className="text-[15px] leading-[1.7] text-slate-600">
            只需简单几步即可开始你的学习之旅：
          </p>
          <div className="flex flex-col gap-3">
            {steps.map((step) => (
              <div key={step.num} className="flex gap-3">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-teal-600">
                  <span className="text-xs font-bold text-white">{step.num}</span>
                </div>
                <div className="flex flex-col gap-1">
                  <span className="text-sm font-semibold text-slate-900">{step.title}</span>
                  <span className="text-[13px] text-slate-500">{step.desc}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Code block */}
        <div className="flex flex-col gap-3">
          <h3 className="text-base font-semibold text-slate-900">示例代码</h3>
          <div className="flex flex-col gap-1 overflow-x-auto rounded-lg bg-slate-800 px-5 py-4">
            {codeLines.map((line, i) => (
              <span key={i} className={`font-mono text-[13px] ${line.color}`}>
                {line.text || "\u00A0"}
              </span>
            ))}
          </div>
        </div>

        <div className="h-px w-full bg-slate-200" />

        {/* Page nav */}
        <div className="flex items-center justify-end">
          <button
            type="button"
            className="flex flex-col items-end gap-1 rounded-lg border border-slate-200 px-4 py-3"
          >
            <span className="text-xs text-slate-400">下一页</span>
            <div className="flex items-center gap-1.5">
              <span className="text-sm font-medium text-slate-700">安装与配置</span>
              <ChevronRight className="h-3.5 w-3.5 text-slate-500" />
            </div>
          </button>
        </div>
      </main>

      {/* Right TOC */}
      <div className="hidden w-[220px] shrink-0 border-l border-slate-200 px-5 py-10 xl:block">
        <div className="flex flex-col gap-3">
          <span className="text-xs font-semibold tracking-wide text-slate-900">
            本页目录
          </span>
          <div className="mt-1 flex flex-col gap-3">
            {tocItems.map((item) => (
              <button
                key={item.label}
                type="button"
                className="flex items-center gap-2 text-left"
              >
                <div
                  className={`h-4 w-0.5 rounded-sm ${
                    item.active ? "bg-teal-600" : "bg-transparent"
                  }`}
                />
                <span
                  className={`text-[13px] ${
                    item.active ? "font-medium text-teal-600" : "text-slate-500"
                  }`}
                >
                  {item.label}
                </span>
              </button>
            ))}
          </div>
        </div>
      </div>
      </div>
    </div>
  );
}
