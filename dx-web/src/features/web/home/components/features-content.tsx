import {
  Zap,
  Swords,
  Keyboard,
  Shuffle,
  Bomb,
  CheckCircle2,
  BrainCircuit,
  BarChart3,
  CalendarClock,
  Trophy,
  MessageSquare,
  UsersRound,
} from "lucide-react";
import Link from "next/link";

const gameCards = [
  { icon: Swords, iconBg: "bg-blue-600/10", iconColor: "text-blue-600", title: "词汇配对", desc: "在限定时间内快速配对英文单词和中文释义，提升词汇反应速度。", bullets: ["限时挑战模式", "多人实时对战", "自动适配难度"] },
  { icon: Keyboard, iconBg: "bg-purple-600/10", iconColor: "text-purple-600", title: "连词成句", desc: "根据提示连词成句，比拼速度和正确率，训练英语组句能力。", bullets: ["连词组句", "AI 语音播报", "拼写错误回顾"] },
  { icon: Shuffle, iconBg: "bg-pink-500/10", iconColor: "text-pink-500", title: "词汇消消乐", desc: "趣味消除游戏与词汇记忆结合，在娱乐中巩固单词记忆。", bullets: ["趣味消除玩法", "词汇巩固复习", "连击加分机制"] },
];

const bigCard = {
  icon: Bomb,
  iconBg: "bg-red-500/10",
  iconColor: "text-red-500",
  title: "词汇对轰",
  desc: "与好友或随机对手进行实时词汇对战，在紧张刺激的对轰中扩大词汇量。支持多种对战模式，包括限时赛、积分赛和淘汰赛。",
  bullets: ["实时在线对战", "多种对战模式", "赛季排行榜"],
};

const coursePills = [
  { label: "渐进体系", color: "bg-purple-600/10 text-purple-600" },
  { label: "AI 生成", color: "bg-teal-600/10 text-teal-600" },
  { label: "智能评测", color: "bg-blue-500/10 text-blue-500" },
  { label: "自定义创建", color: "bg-amber-500/10 text-amber-500" },
];

const vocabCards = [
  { icon: BrainCircuit, iconBg: "bg-pink-500/10", iconColor: "text-pink-500", title: "智能生词本", desc: "游戏中自动收录新词，支持分组管理和自定义标签。" },
  { icon: CalendarClock, iconBg: "bg-purple-600/10", iconColor: "text-purple-600", title: "间隔复习引擎", desc: "基于艾宾浩斯曲线，在遗忘临界点推送复习提醒。" },
  { icon: BarChart3, iconBg: "bg-teal-600/10", iconColor: "text-teal-600", title: "掌握度追踪", desc: "可视化展示每个单词的掌握程度和复习进度。" },
];

const socialCards = [
  { icon: Trophy, iconBg: "bg-amber-500/10", iconColor: "text-amber-500", title: "学习排行榜", desc: "每周和每月排名激励，与全国学友同台竞技。" },
  { icon: MessageSquare, iconBg: "bg-orange-600/10", iconColor: "text-orange-600", title: "斗学社论坛", desc: "分享学习心得，讨论学习方法，找到志同道合的伙伴。" },
  { icon: UsersRound, iconBg: "bg-blue-500/10", iconColor: "text-blue-500", title: "学习小组", desc: "创建或加入学习小组，一起打卡学习，互相监督进步。" },
];

function SectionHeader({
  label,
  labelColor,
  title,
  subtitle,
}: {
  label: string;
  labelColor: string;
  title: string;
  subtitle: string;
}) {
  return (
    <div className="flex flex-col items-center gap-4">
      <span className={`text-[13px] font-semibold tracking-[3px] ${labelColor}`}>
        {label}
      </span>
      <h2 className="text-center text-3xl font-extrabold tracking-tighter text-slate-900 md:text-5xl">
        {title}
      </h2>
      <p className="max-w-[840px] text-center text-base leading-relaxed text-slate-500 md:text-lg">
        {subtitle}
      </p>
    </div>
  );
}

function FeatureCard({
  icon: Icon,
  iconBg,
  iconColor,
  title,
  desc,
  bullets,
}: {
  icon: React.ElementType;
  iconBg: string;
  iconColor: string;
  title: string;
  desc: string;
  bullets?: string[];
}) {
  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm md:p-8">
      <div className={`flex h-14 w-14 items-center justify-center rounded-xl ${iconBg}`}>
        <Icon className={`h-7 w-7 ${iconColor}`} />
      </div>
      <h3 className="text-xl font-bold text-slate-900">{title}</h3>
      <p className="text-[15px] leading-[1.7] text-slate-500">{desc}</p>
      {bullets && (
        <div className="flex flex-col gap-2">
          {bullets.map((b) => (
            <div key={b} className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-teal-500" />
              <span className="text-sm text-slate-600">{b}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function FeaturesContent() {
  return (
    <>
      {/* Hero */}
      <div className="bg-gradient-to-b from-teal-50 via-blue-50 via-45% via-purple-50 via-70% to-white">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col gap-6 px-4 pb-16 pt-12 sm:px-8 md:px-16 md:pb-[100px] md:pt-20 lg:px-[120px]">
          <div className="flex items-center gap-2 self-start rounded-full bg-teal-600/10 px-5 py-2">
            <Zap className="h-4 w-4 text-teal-600" />
            <span className="text-[13px] font-semibold text-teal-600">
              全方位沉浸式英语学习功能
            </span>
          </div>
          <h1 className="whitespace-pre-line text-3xl font-extrabold leading-[1.15] tracking-tighter text-slate-900 md:text-[56px]">
            {"让每一项功能\n都成为你的超能力"}
          </h1>
          <p className="max-w-[791px] text-base leading-relaxed text-slate-500 md:text-lg">
            从游戏化对战到 AI 智能练习，从科学词汇管理到社区互动，斗学为你打造最完整的英语学习体验。
          </p>
        </div>
      </div>

      {/* Game Modes */}
      <section className="w-full bg-gradient-to-b from-white to-teal-50 py-16 md:py-[100px]">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-4 sm:px-8 md:gap-[60px] md:px-16 lg:px-[120px]">
        <SectionHeader
          label="多重学习模式"
          labelColor="text-purple-600"
          title="玩中学，主动练"
          subtitle="每种学习模式专攻不同英语技能维度，结合即时反馈与竞技排名，让你在不知不觉中突破瓶颈。"
        />
        <div className="flex w-full flex-col gap-6">
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            {gameCards.map((card) => (
              <FeatureCard key={card.title} {...card} />
            ))}
          </div>
          <FeatureCard {...bigCard} />
        </div>
        </div>
      </section>

      {/* Course Platform */}
      <section className="w-full bg-gradient-to-b from-teal-50 to-purple-50 py-16 md:py-[100px]">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-4 sm:px-8 md:gap-[60px] md:px-16 lg:px-[120px]">
        <SectionHeader
          label="闯关式课程"
          labelColor="text-purple-500"
          title="系统化学习，关卡式成长"
          subtitle="从零基础到高级进阶，每个课程都是一条精心铺设的学习路径。通过游戏闯关完成知识点学习，支持自定义课程创建。"
        />
        <div className="flex w-full flex-col items-center gap-8">
          <div className="flex flex-wrap items-center justify-center gap-4">
            {coursePills.map((pill) => (
              <span
                key={pill.label}
                className={`rounded-full px-4 py-2 text-sm font-medium ${pill.color}`}
              >
                {pill.label}
              </span>
            ))}
          </div>
          <p className="max-w-[737px] text-center text-sm leading-[1.7] text-slate-600 md:text-base">
            每个课程由多个主题单元组成，每个单元包含 5-10 个游戏关卡。关卡类型涵盖词汇配对、连词成句、听力闯关、语法探索和阅读理解。通关当前关卡才能解锁下一关，每个单元结束后有综合测试，确保知识点真正掌握。支持用户自定义创建课程，AI 根据指定主题自动生成关卡内容与难度梯度。
          </p>
        </div>
        </div>
      </section>

      {/* Smart Vocab */}
      <section className="w-full bg-gradient-to-b from-purple-50 to-pink-50 py-16 md:py-[100px]">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-4 sm:px-8 md:gap-[60px] md:px-16 lg:px-[120px]">
        <SectionHeader
          label="科学记忆系统"
          labelColor="text-pink-500"
          title="让每个单词真正住进你的大脑"
          subtitle="基于艾宾浩斯遗忘曲线算法，结合你的个人记忆模式，在最佳时机推送复习提醒，告别死记硬背。"
        />
        <div className="grid w-full grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {vocabCards.map((card) => (
            <FeatureCard key={card.title} {...card} />
          ))}
        </div>
        </div>
      </section>

      {/* Social Community */}
      <section className="w-full bg-gradient-to-b from-pink-50 to-orange-50 py-16 md:py-[100px]">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-4 sm:px-8 md:gap-[60px] md:px-16 lg:px-[120px]">
        <SectionHeader
          label="学习不孤单"
          labelColor="text-orange-600"
          title="与百万学友共同进步"
          subtitle="学习最大的敌人是孤独。斗学社区让你找到志同道合的学习伙伴，在竞争中成长，在互助中突破。"
        />
        <div className="grid w-full grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {socialCards.map((card) => (
            <FeatureCard key={card.title} {...card} />
          ))}
        </div>
        </div>
      </section>

      {/* CTA */}
      <section className="w-full bg-gradient-to-b from-orange-50 to-teal-50 py-16 md:py-[100px]">
        <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-4 sm:px-8 md:px-16 lg:px-[120px]">
        <div className="hidden h-1 w-full max-w-[800px] rounded-full bg-gradient-to-r from-teal-400/0 via-teal-400 via-30% via-purple-500 via-70% to-purple-500/0 md:block" />
        <h2 className="text-center text-3xl font-extrabold tracking-tighter text-slate-900 md:text-5xl">
          每一项功能，都为你的进步而生
        </h2>
        <p className="max-w-[648px] text-center text-base leading-relaxed text-slate-500 md:text-lg">
          立即注册，免费解锁所有核心功能。加入50万+学习者，开启你的英语进化之旅！
        </p>
        <div className="flex flex-col items-center gap-4 sm:flex-row">
          <Link
            href="/auth/signup"
            className="rounded-xl bg-teal-600 px-9 py-4 text-base font-semibold text-white shadow-lg shadow-teal-600/25 hover:bg-teal-700"
          >
            免费注册
          </Link>
          <Link
            href="/wiki"
            className="rounded-xl border border-slate-300 px-9 py-4 text-base font-medium text-slate-700 hover:bg-slate-50"
          >
            查看 Wiki
          </Link>
        </div>
        </div>
      </section>
    </>
  );
}
