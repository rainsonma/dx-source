import {
  Layers,
  CirclePlus,
  Brain,
  BarChart3,
  type LucideIcon,
} from "lucide-react";

const pills: {
  icon: LucideIcon;
  iconClassName: string;
  bgClassName: string;
  label: string;
}[] = [
  {
    icon: Layers,
    iconClassName: "text-violet-600",
    bgClassName: "bg-violet-50 border-violet-200",
    label: "关卡制学习",
  },
  {
    icon: CirclePlus,
    iconClassName: "text-teal-600",
    bgClassName: "bg-teal-50 border-teal-200",
    label: "自定义课程",
  },
  {
    icon: Brain,
    iconClassName: "text-blue-600",
    bgClassName: "bg-blue-50 border-blue-200",
    label: "AI 生成内容",
  },
  {
    icon: BarChart3,
    iconClassName: "text-amber-600",
    bgClassName: "bg-amber-50 border-amber-200",
    label: "进度追踪",
  },
];

export function CoursePlatformSection() {
  return (
    <section className="flex w-full flex-col items-center gap-[60px] bg-gradient-to-b from-teal-50 to-violet-50 px-[120px] py-[100px]">
      <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-violet-500">
          课程体系
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          丰富趣味课程，闯关式学习
        </h2>
      </div>
      <p className="max-w-[720px] text-center text-[15px] leading-relaxed text-slate-500">
        涵盖单词配对、听说读写、听力闯关、语法探索、阅读理解等多种游戏类型。每个课程包含多个关卡，由浅入深，通关即解锁下一关。支持用户自定义创建课程，AI
        智能生成练习内容。
      </p>
      <div className="flex flex-wrap items-center justify-center gap-4">
        {pills.map((pill) => {
          const Icon = pill.icon;
          return (
            <div
              key={pill.label}
              className={`flex items-center gap-2 rounded-xl border px-6 py-3 ${pill.bgClassName}`}
            >
              <Icon className={`h-5 w-5 ${pill.iconClassName}`} />
              <span className="text-[15px] font-medium text-slate-700">
                {pill.label}
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}
