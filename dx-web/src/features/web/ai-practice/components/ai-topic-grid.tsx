"use client";

import {
  Bot,
  MessageSquarePlus,
  Send,
  Coffee,
  Building2,
  Plane,
  Utensils,
  HandHeart,
  ShoppingBag,
  Briefcase,
  Stethoscope,
  GraduationCap,
  Clapperboard,
  Music,
  Presentation,
} from "lucide-react";

const categories = [
  { label: "全部", active: true },
  { label: "日常生活", active: false },
  { label: "学习教育", active: false },
  { label: "职场商务", active: false },
  { label: "旅行出行", active: false },
  { label: "文化娱乐", active: false },
  { label: "社交情感", active: false },
  { label: "健康运动", active: false },
];

const topics = [
  { icon: Coffee, iconColor: "text-amber-600", iconBg: "bg-amber-100", title: "咖啡店点餐", desc: "练习在咖啡店点单、询问推荐和定制饮品", tags: ["日常", "初级"] },
  { icon: Building2, iconColor: "text-blue-600", iconBg: "bg-blue-100", title: "酒店入住", desc: "办理入住退房、客房服务和设施咨询", tags: ["旅行", "中级"] },
  { icon: Plane, iconColor: "text-pink-600", iconBg: "bg-pink-100", title: "机场出行", desc: "值机登机、安检过关和航班延误沟通", tags: ["旅行", "中级"] },
  { icon: Utensils, iconColor: "text-green-600", iconBg: "bg-green-50", title: "餐厅点菜", desc: "阅读菜单、点餐下单和特殊饮食需求沟通", tags: ["日常", "初级"] },
  { icon: HandHeart, iconColor: "text-teal-600", iconBg: "bg-teal-100", title: "日常问候", desc: "打招呼、自我介绍和寒暄聊天", tags: ["日常", "入门"] },
  { icon: ShoppingBag, iconColor: "text-red-600", iconBg: "bg-red-100", title: "购物砍价", desc: "逛街购物、比较价格和退换货交流", tags: ["日常", "中级"] },
  { icon: Briefcase, iconColor: "text-sky-600", iconBg: "bg-sky-50", title: "面试对话", desc: "英语面试问答、自我介绍和职业描述", tags: ["职场", "高级"] },
  { icon: Stethoscope, iconColor: "text-emerald-600", iconBg: "bg-emerald-50", title: "看病就医", desc: "描述症状、预约挂号和药房买药交流", tags: ["日常", "中级"] },
  { icon: GraduationCap, iconColor: "text-orange-600", iconBg: "bg-orange-50", title: "校园生活", desc: "课堂讨论、社团活动和校园日常交流", tags: ["学习", "初级"] },
  { icon: Clapperboard, iconColor: "text-fuchsia-600", iconBg: "bg-fuchsia-50", title: "电影讨论", desc: "聊电影剧情、推荐影片和影评分享", tags: ["娱乐", "中级"] },
  { icon: Music, iconColor: "text-teal-600", iconBg: "bg-teal-50", title: "音乐分享", desc: "讨论音乐风格、推荐歌曲和演唱会体验", tags: ["娱乐", "初级"] },
  { icon: Presentation, iconColor: "text-blue-600", iconBg: "bg-blue-100", title: "商务会议", desc: "会议讨论、项目汇报和团队协作表达", tags: ["职场", "高级"] },
];

const tagColors: Record<string, { bg: string; text: string }> = {
  "日常": { bg: "bg-teal-50", text: "text-teal-600" },
  "旅行": { bg: "bg-blue-50", text: "text-blue-600" },
  "职场": { bg: "bg-purple-50", text: "text-purple-600" },
  "学习": { bg: "bg-orange-50", text: "text-orange-600" },
  "娱乐": { bg: "bg-pink-50", text: "text-pink-600" },
  "入门": { bg: "bg-green-50", text: "text-green-600" },
  "初级": { bg: "bg-emerald-50", text: "text-emerald-600" },
  "中级": { bg: "bg-amber-50", text: "text-amber-600" },
  "高级": { bg: "bg-red-50", text: "text-red-600" },
};

export function AiTopicGrid() {
  return (
    <>
      {/* Hero Banner */}
      <div className="flex flex-col gap-4 rounded-2xl bg-gradient-to-br from-teal-600 via-teal-700 to-teal-900 p-5 md:p-8">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-[14px] bg-white/10">
            <Bot className="h-7 w-7 text-white" />
          </div>
          <span className="text-xl font-extrabold tracking-tight text-white md:text-[28px]">
            AI 随心练
          </span>
        </div>
        <p className="text-sm leading-relaxed text-white/80">
          AI 为你量身打造英语对话场景，沉浸式练习口语表达，随时随地开启智能学习之旅
        </p>
        <div className="flex flex-col gap-3 sm:flex-row">
          <div className="flex h-11 flex-1 items-center gap-2 rounded-xl bg-white px-4">
            <MessageSquarePlus className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-[13px] text-muted-foreground">
              输入你想练习的话题，或从下方选择...
            </span>
          </div>
          <button
            type="button"
            className="flex items-center justify-center gap-1.5 rounded-xl bg-white px-5 py-2.5"
          >
            <Send className="h-4 w-4 text-teal-600" />
            <span className="text-[13px] font-semibold text-teal-600">开始对话</span>
          </button>
        </div>
      </div>

      {/* Category tabs */}
      <div className="flex flex-wrap items-center gap-2.5">
        {categories.map((cat) => (
          <button
            key={cat.label}
            type="button"
            className={`rounded-lg px-3.5 py-1.5 text-[13px] font-medium ${
              cat.active
                ? "bg-teal-600 text-white"
                : "border border-border bg-muted text-muted-foreground"
            }`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* Topic grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {topics.map((topic) => (
          <div
            key={topic.title}
            className="flex gap-3.5 rounded-[14px] border border-border bg-card p-4"
          >
            <div className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl ${topic.iconBg}`}>
              <topic.icon className={`h-[22px] w-[22px] ${topic.iconColor}`} />
            </div>
            <div className="flex flex-col gap-1.5">
              <span className="text-sm font-bold text-foreground">{topic.title}</span>
              <span className="text-xs leading-snug text-muted-foreground">{topic.desc}</span>
              <div className="flex gap-1.5">
                {topic.tags.map((tag) => {
                  const c = tagColors[tag] ?? { bg: "bg-muted", text: "text-muted-foreground" };
                  return (
                    <span
                      key={tag}
                      className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${c.bg} ${c.text}`}
                    >
                      {tag}
                    </span>
                  );
                })}
              </div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
