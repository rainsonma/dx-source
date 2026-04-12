import {
  Keyboard,
  Swords,
  Shuffle,
  Crosshair,
  type LucideIcon,
} from "lucide-react";

const features: {
  icon: LucideIcon;
  iconClassName: string;
  bgClassName: string;
  title: string;
  description: string;
}[] = [
  {
    icon: Keyboard,
    iconClassName: "text-violet-600",
    bgClassName: "bg-violet-100",
    title: "连词成句",
    description:
      "连词成句全方位训练，看看你能多快组成正确句子。速度与准确并重，越快越高分。",
  },
  {
    icon: Swords,
    iconClassName: "text-teal-600",
    bgClassName: "bg-blue-100",
    title: "词汇配对",
    description:
      "与全球学习者实时对战，在紧张刺激的词汇配对比拼中提升你的词汇量和反应速度。",
  },
  {
    icon: Shuffle,
    iconClassName: "text-emerald-600",
    bgClassName: "bg-pink-100",
    title: "词汇消消乐",
    description:
      "记忆配对消除游戏，将英文单词与中文释义快速匹配。考验记忆力与反应速度，越快消除分数越高。",
  },
  {
    icon: Crosshair,
    iconClassName: "text-red-500",
    bgClassName: "bg-red-100",
    title: "词汇对轰",
    description:
      "与 AI 对手展开词汇炮弹大战！快速拼写正确单词发射炮弹，攻防兼备，紧张刺激的对战体验。",
  },
];

function FeatureCard({
  feature,
}: {
  feature: (typeof features)[number];
}) {
  const Icon = feature.icon;

  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-8">
      <div
        className={`flex h-14 w-14 items-center justify-center rounded-2xl ${feature.bgClassName}`}
      >
        <Icon className={`h-7 w-7 ${feature.iconClassName}`} />
      </div>
      <h3 className="text-xl font-bold text-slate-900">{feature.title}</h3>
      <p className="text-[15px] leading-relaxed text-slate-500">
        {feature.description}
      </p>
    </div>
  );
}

export function FeaturesSection() {
  return (
    <section
      id="features"
      className="w-full bg-gradient-to-b from-white to-slate-50 py-[100px]"
    >
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-[60px] px-[120px]">
        <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-violet-500">
          核心玩法
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          多种游戏，全方位提升
        </h2>
        <p className="max-w-[520px] text-center text-lg text-slate-500">
          每个游戏针对不同英语技能，从词汇到听力到拼写，全面覆盖。
        </p>
      </div>
      <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2">
        {features.map((feature) => (
          <FeatureCard key={feature.title} feature={feature} />
        ))}
      </div>
      </div>
    </section>
  );
}
