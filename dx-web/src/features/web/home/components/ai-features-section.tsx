import { Sparkles, MessageCircle, type LucideIcon } from "lucide-react";

const aiFeatures: {
  icon: LucideIcon;
  iconClassName: string;
  bgClassName: string;
  tagClassName: string;
  title: string;
  description: string;
  tags: string[];
}[] = [
  {
    icon: Sparkles,
    iconClassName: "text-teal-600",
    bgClassName: "bg-teal-100",
    tagClassName: "border-teal-200 bg-teal-50 text-teal-700",
    title: "AI 随心学",
    description:
      "自由定制游戏内容，AI 根据你的水平与兴趣智能生成词汇、话题和关卡。从日常会话到专业领域，打造专属于你的学习路径。",
    tags: ["内容定制", "智能生成", "专属关卡"],
  },
  {
    icon: MessageCircle,
    iconClassName: "text-violet-600",
    bgClassName: "bg-violet-100",
    tagClassName: "border-violet-200 bg-violet-50 text-violet-700",
    title: "AI 随心练",
    description:
      "与 AI 自由对话，支持自定义话题。对话结束后获取详细学习报告，词汇、语法、流利度一目了然。",
    tags: ["自由对话", "学习报告", "自定义话题"],
  },
];

function AiFeatureCard({
  feature,
}: {
  feature: (typeof aiFeatures)[number];
}) {
  const Icon = feature.icon;

  return (
    <div className="flex flex-1 flex-col gap-6 rounded-2xl border border-slate-200 bg-white p-10">
      <div
        className={`flex h-14 w-14 items-center justify-center rounded-2xl ${feature.bgClassName}`}
      >
        <Icon className={`h-7 w-7 ${feature.iconClassName}`} />
      </div>
      <h3 className="text-2xl font-bold text-slate-900">{feature.title}</h3>
      <p className="text-[15px] leading-relaxed text-slate-500">
        {feature.description}
      </p>
      <div className="flex flex-wrap gap-2.5">
        {feature.tags.map((tag) => (
          <span
            key={tag}
            className={`rounded-full border px-4 py-1.5 text-[13px] font-medium ${feature.tagClassName}`}
          >
            {tag}
          </span>
        ))}
      </div>
    </div>
  );
}

export function AiFeaturesSection() {
  return (
    <section className="w-full bg-gradient-to-b from-slate-50 to-teal-50 py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-[60px] px-[120px]">
        <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-teal-600">
          AI 驱动
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          AI 陪你练，开口说英语
        </h2>
        <p className="max-w-[580px] text-center text-lg text-slate-500">
          不只是练习，更是真实体验。AI 智能学习，建立英语自信。
        </p>
      </div>
      <div className="flex w-full flex-col gap-6 md:flex-row">
        {aiFeatures.map((feature) => (
          <AiFeatureCard key={feature.title} feature={feature} />
        ))}
      </div>
      </div>
    </section>
  );
}
