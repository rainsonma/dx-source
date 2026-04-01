import { Twitter, Instagram, Star, type LucideIcon } from "lucide-react";

const testimonials: {
  socialIcon: LucideIcon;
  quote: string;
  author: string;
  level: string;
}[] = [
  {
    socialIcon: Twitter,
    quote:
      "\u201C连词成句和词汇对战让我彻底上瘾了！从几乎看不懂菜单到三个月内能进行完整对话。PvP 对战让我真的想每天练习。\u201D",
    author: "sarah.lin",
    level: "Lv.28",
  },
  {
    socialIcon: Instagram,
    quote:
      "\u201CAI 随心练太惊艳了！像和真人对话一样，结束后还有详细报告。我的口语从结结巴巴到现在能流利聊天，半年进步巨大。\u201D",
    author: "小明明同学",
    level: "Lv.34",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201C单词消消乐简直太上头了！配对游戏让背单词变成了享受，不知不觉记住了上千个单词。孩子也抢着玩。\u201D",
    author: "emilyyyy",
    level: "Lv.15",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201C加入学习群后，和小伙伴们一起闯关、互相督促，学习动力爆棚。排行榜的竞争感让我根本停不下来！\u201D",
    author: "DavidChen",
    level: "Lv.42",
  },
  {
    socialIcon: Instagram,
    quote:
      "\u201C生词本和复习本功能太贴心了。科学记忆算法帮我精准复习，半年下来词汇量翻了三倍，雅思从5.5提到了7分。\u201D",
    author: "anna.k",
    level: "Lv.56",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201C听力闯关帮我突破了最大的瓶颈。从听不懂外国人说话到现在能看无字幕美剧，感觉像换了个耳朵。\u201D",
    author: "小瑞语",
    level: "Lv.32",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201C词汇对轰的 AI 对战模式太刺激了！和 AI 互射炮弹，紧张又好玩。不知不觉就练了一个小时拼写。\u201D",
    author: "阿杰学英语",
    level: "Lv.20",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201CAI 随心配让我根据自己的职业需求定制学习内容，商务英语提升飞快。面试外企再也不怕了！\u201D",
    author: "晓薇",
    level: "Lv.45",
  },
  {
    socialIcon: Twitter,
    quote:
      "\u201C课程游戏的闯关模式太适合我了。每天通两关，半年下来从零基础到能和外国朋友正常聊天。\u201D",
    author: "kev.m",
    level: "Lv.18",
  },
];

function StarRating() {
  return (
    <div className="flex gap-1">
      {Array.from({ length: 5 }).map((_, i) => (
        <Star
          key={i}
          className="h-4 w-4 fill-yellow-400 text-yellow-400"
        />
      ))}
    </div>
  );
}

function TestimonialCard({
  testimonial,
}: {
  testimonial: (typeof testimonials)[number];
}) {
  const SocialIcon = testimonial.socialIcon;

  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-8 shadow-[0_4px_16px_rgba(15,23,42,0.03)]">
      <div className="flex items-center justify-between">
        <StarRating />
        <SocialIcon className="h-5 w-5 text-slate-400" />
      </div>
      <p className="text-[15px] leading-relaxed text-slate-600">
        {testimonial.quote}
      </p>
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-full bg-slate-200" />
        <div className="flex flex-col">
          <span className="text-sm font-semibold text-slate-900">
            {testimonial.author}
          </span>
          <span className="text-xs text-slate-400">{testimonial.level}</span>
        </div>
      </div>
    </div>
  );
}

export function TestimonialsSection() {
  return (
    <section className="flex w-full flex-col items-center gap-[60px] bg-gradient-to-b from-violet-50 to-pink-50 px-[120px] py-[100px]">
      <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-orange-600">
          用户评价
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          他们的英语，在游戏中开挂
        </h2>
      </div>
      <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {testimonials.map((testimonial) => (
          <TestimonialCard
            key={testimonial.author}
            testimonial={testimonial}
          />
        ))}
      </div>
    </section>
  );
}
