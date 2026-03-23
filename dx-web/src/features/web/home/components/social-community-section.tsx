import {
  Trophy,
  MessageSquare,
  Users,
  type LucideIcon,
} from "lucide-react";

const communityFeatures: {
  icon: LucideIcon;
  iconColor: string;
  bgClassName: string;
  title: string;
  description: string;
}[] = [
  {
    icon: Trophy,
    iconColor: "#F59E0B",
    bgClassName: "bg-amber-500/10",
    title: "排行榜",
    description:
      "每周、每月排名，与全球学习者一决高下。登上领奖台，赢取专属称号和奖励。",
  },
  {
    icon: MessageSquare,
    iconColor: "#EA580C",
    bgClassName: "bg-orange-600/10",
    title: "斗学社",
    description:
      "分享学习心得，交流学习技巧，互相激励。发帖、评论、点赞，学习也能很社交。",
  },
  {
    icon: Users,
    iconColor: "#3B82F6",
    bgClassName: "bg-blue-500/10",
    title: "学习群",
    description:
      "组建学习小组，一起闯关，互相督促。支持群内发起课程游戏，实时协作学习。",
  },
];

function CommunityCard({
  feature,
}: {
  feature: (typeof communityFeatures)[number];
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

export function SocialCommunitySection() {
  return (
    <section className="flex w-full flex-col items-center gap-[60px] bg-gradient-to-b from-pink-50 to-orange-50 px-[120px] py-[100px]">
      <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-orange-600">
          社交学习
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          和百万学友一起进步
        </h2>
        <p className="max-w-[600px] text-center text-[15px] leading-relaxed text-slate-500">
          学习不孤单，社区互助，组队挑战，排行争霸。
        </p>
      </div>
      <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {communityFeatures.map((feature) => (
          <CommunityCard key={feature.title} feature={feature} />
        ))}
      </div>
    </section>
  );
}
