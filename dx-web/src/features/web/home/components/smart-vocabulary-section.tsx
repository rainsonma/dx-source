import {
  BookOpen,
  RefreshCw,
  CircleCheck,
  type LucideIcon,
} from "lucide-react";

const vocabularyFeatures: {
  icon: LucideIcon;
  iconColor: string;
  bgClassName: string;
  title: string;
  description: string;
}[] = [
  {
    icon: BookOpen,
    iconColor: "#EC4899",
    bgClassName: "bg-pink-500/10",
    title: "生词本",
    description:
      "自动收集你不会的单词，分级分类，随时复习。支持搜索、排序、难度筛选。",
  },
  {
    icon: RefreshCw,
    iconColor: "#7B61FF",
    bgClassName: "bg-violet-500/10",
    title: "复习本",
    description:
      "艾宾浩斯遗忘曲线算法，智能推送复习提醒。在最佳时机巩固记忆，事半功倍。",
  },
  {
    icon: CircleCheck,
    iconColor: "#0d9488",
    bgClassName: "bg-teal-500/10",
    title: "已掌握",
    description:
      "看着你的词汇量持续增长，成就感满满。清晰的掌握进度，激励你不断前进。",
  },
];

function VocabularyCard({
  feature,
}: {
  feature: (typeof vocabularyFeatures)[number];
}) {
  const Icon = feature.icon;

  return (
    <div className="flex flex-col items-center gap-5 rounded-2xl border border-slate-200 bg-white px-8 py-9 text-center shadow-[0_4px_16px_rgba(15,23,42,0.03)]">
      <div
        className={`flex h-14 w-14 items-center justify-center rounded-[14px] ${feature.bgClassName}`}
      >
        <Icon className="h-7 w-7" style={{ color: feature.iconColor }} />
      </div>
      <h3 className="text-xl font-bold text-slate-900">{feature.title}</h3>
      <p className="text-[15px] leading-relaxed text-slate-500">
        {feature.description}
      </p>
    </div>
  );
}

export function SmartVocabularySection() {
  return (
    <section className="w-full bg-gradient-to-b from-violet-50 to-pink-50 py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-[60px] px-[120px]">
        <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-pink-500">
          智能词库
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          你的专属词汇管家
        </h2>
        <p className="max-w-[600px] text-center text-[15px] leading-relaxed text-slate-500">
          科学记忆算法，精准追踪每个单词的掌握程度。
        </p>
      </div>
      <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {vocabularyFeatures.map((feature) => (
          <VocabularyCard key={feature.title} feature={feature} />
        ))}
      </div>
      </div>
    </section>
  );
}
