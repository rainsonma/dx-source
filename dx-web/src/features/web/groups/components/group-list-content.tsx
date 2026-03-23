"use client";

import { Plus, ChevronRight } from "lucide-react";

type GroupVariant = "teal" | "amber" | "indigo" | "red" | "green" | "purple";

const variantClasses: Record<GroupVariant, { avatarBg: string; avatarColor: string; tagBg: string; tagColor: string }> = {
  teal: { avatarBg: "bg-teal-100", avatarColor: "text-teal-700", tagBg: "bg-teal-50", tagColor: "text-teal-600" },
  amber: { avatarBg: "bg-amber-100", avatarColor: "text-amber-700", tagBg: "bg-amber-50", tagColor: "text-amber-700" },
  indigo: { avatarBg: "bg-indigo-100", avatarColor: "text-indigo-700", tagBg: "bg-indigo-50", tagColor: "text-indigo-700" },
  red: { avatarBg: "bg-red-100", avatarColor: "text-red-700", tagBg: "bg-red-50", tagColor: "text-red-700" },
  green: { avatarBg: "bg-green-50", avatarColor: "text-green-700", tagBg: "bg-green-50", tagColor: "text-green-700" },
  purple: { avatarBg: "bg-purple-50", avatarColor: "text-purple-700", tagBg: "bg-purple-50", tagColor: "text-purple-700" },
};

const tabs = [
  { label: "全部", active: true },
  { label: "我建的群", active: false },
  { label: "我加的群", active: false },
];

const groups = [
  { letter: "六", name: "六年级英语冲刺群", creator: "王老师", members: "128 人", desc: "专注六年级英语考试提分，群组课程游戏每日打卡", tags: ["英语", "六年级"], joined: true, highlighted: true, variant: "teal" as GroupVariant },
  { letter: "三", name: "三年级口语角", creator: "李明", members: "86 人", desc: "每天练习英语口语对话，纠正发音，培养语感", tags: ["口语", "三年级"], joined: true, highlighted: false, variant: "amber" as GroupVariant },
  { letter: "听", name: "听力达人养成营", creator: "张华", members: "215 人", desc: "沉浸式英语听力训练，从听懂到听会", tags: ["听力", "全年级"], joined: false, highlighted: false, variant: "indigo" as GroupVariant },
  { letter: "拼", name: "拼写王者争霸赛", creator: "王老师", members: "64 人", desc: "竞技式拼写挑战，在对抗中巩固单词记忆", tags: ["拼写", "竞赛"], joined: true, highlighted: false, variant: "red" as GroupVariant },
  { letter: "中", name: "中考英语突击班", creator: "刘芳", members: "342 人", desc: "针对中考英语高频考点，系统刷题提分", tags: ["中考", "提分"], joined: false, highlighted: false, variant: "green" as GroupVariant },
  { letter: "A", name: "AI 词汇探索营", creator: "陈思", members: "178 人", desc: "AI 驱动的智能词汇学习，个性化推荐记忆方案", tags: ["AI", "词汇"], joined: false, highlighted: false, variant: "purple" as GroupVariant },
];

export function GroupListContent() {
  return (
    <>
      {/* Tab row */}
      <div className="flex w-full flex-col gap-3 border-b border-border md:flex-row md:items-center md:gap-0">
        <div className="flex items-center">
          {tabs.map((tab) => (
            <button
              key={tab.label}
              type="button"
              className={`px-5 py-2 text-sm font-medium ${
                tab.active
                  ? "border-b-2 border-teal-600 text-teal-600"
                  : "text-muted-foreground"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <div className="hidden flex-1 md:block" />
        <button
          type="button"
          className="flex w-full items-center justify-center gap-1.5 rounded-lg bg-teal-600 px-3.5 py-2 text-sm font-medium text-white hover:bg-teal-700 md:w-auto"
        >
          <Plus className="h-4 w-4" />
          创建学习群
        </button>
      </div>

      {/* Card grid */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-3">
        {groups.map((group) => {
          const v = variantClasses[group.variant];
          return (
            <div
              key={group.name}
              className={`flex flex-col gap-3.5 rounded-[14px] p-5 ${
                group.highlighted
                  ? "border-2 border-teal-600 bg-teal-50/50"
                  : "border border-border bg-card"
              }`}
            >
              {/* Header */}
              <div className="flex items-center gap-3">
                <div
                  className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl ${v.avatarBg}`}
                >
                  <span className={`text-lg font-bold ${v.avatarColor}`}>{group.letter}</span>
                </div>
                <div className="flex flex-1 flex-col gap-0.5">
                  <span className="text-[15px] font-semibold text-foreground">{group.name}</span>
                  <div className="flex items-center gap-1.5">
                    <div className={`flex h-[18px] w-[18px] items-center justify-center rounded-full ${v.avatarBg}`}>
                      <span className={`text-[8px] font-semibold ${v.avatarColor}`}>{group.creator[0]}</span>
                    </div>
                    <span className="text-xs font-medium text-muted-foreground">{group.creator}</span>
                    <span className="text-xs text-muted-foreground">·</span>
                    <span className="text-xs text-muted-foreground">{group.members}</span>
                  </div>
                </div>
                {group.joined ? (
                  <span className="rounded-md bg-teal-600/10 px-2.5 py-1 text-[11px] font-medium text-teal-600">已加入</span>
                ) : (
                  <button type="button" className="flex items-center gap-1 rounded-md bg-muted px-2.5 py-1 text-[11px] font-medium text-muted-foreground">
                    <Plus className="h-[11px] w-[11px]" />
                    加入
                  </button>
                )}
              </div>

              <p className="text-[13px] leading-relaxed text-muted-foreground">{group.desc}</p>

              <div className="flex items-center justify-between">
                <div className="flex gap-1.5">
                  {group.tags.map((tag) => (
                    <span key={tag} className={`rounded-md px-2 py-1 text-[11px] font-medium ${v.tagBg} ${v.tagColor}`}>{tag}</span>
                  ))}
                </div>
                <div className="flex h-7 w-7 items-center justify-center rounded-full bg-slate-400/10">
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </>
  );
}
