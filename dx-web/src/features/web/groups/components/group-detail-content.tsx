import {
  ArrowLeft,
  ChevronRight,
  Play,
  Link2,
  QrCode,
  Copy,
  Check,
  Users,
} from "lucide-react";

const groupInfo = {
  name: "六年级英语冲刺群",
  avatar: "六",
  creator: "张老师",
  desc: "六年级英语冲刺复习群，一起努力备战期末考试！每天坚持打卡学习，互相监督共同进步。",
  stats: [
    { value: "128", label: "成员" },
    { value: "3", label: "小组" },
    { value: "42天", label: "已创建" },
  ],
  game: { name: "词汇对战", level: "六年级上册", mode: "竞技模式" },
};

const members = [
  { name: "张老师", role: "群主", avatar: "张", bg: "bg-teal-600", textColor: "text-white", isOwner: true },
  { name: "李明", role: "管理员", avatar: "李", bg: "bg-teal-100", textColor: "text-teal-700", checked: false },
  { name: "王小红", role: "成员", avatar: "王", bg: "bg-amber-100", textColor: "text-amber-700", checked: true },
  { name: "刘宇航", role: "成员", avatar: "刘", bg: "bg-indigo-100", textColor: "text-indigo-700", checked: false },
  { name: "赵梦琪", role: "成员", avatar: "赵", bg: "bg-red-100", textColor: "text-red-700", checked: false },
  { name: "陈思雨", role: "成员", avatar: "陈", bg: "bg-fuchsia-100", textColor: "text-fuchsia-700", checked: true },
  { name: "马天宇", role: "成员", avatar: "马", bg: "bg-green-100", textColor: "text-green-700", checked: false },
];

const subGroups = [
  { name: "词汇提高小组", members: 32, avatar: "词", bg: "bg-blue-50", rank: 1 },
  { name: "背诵打卡小组", members: 28, avatar: "背", bg: "bg-amber-100", rank: 2 },
  { name: "语法突破小组", members: 24, avatar: "语", bg: "bg-green-50", rank: 3 },
  { name: "阅读拓展小组", members: 18, avatar: "阅", bg: "bg-red-100", rank: 4 },
];

const subGroupMembers = [
  { name: "李明", role: "组长", avatar: "李", bg: "bg-teal-100", textColor: "text-teal-700" },
  { name: "张伟", role: "成员", avatar: "张", bg: "bg-indigo-100", textColor: "text-indigo-700" },
  { name: "赵梦琪", role: "成员", avatar: "赵", bg: "bg-red-100", textColor: "text-red-700" },
  { name: "陈思雨", role: "成员", avatar: "陈", bg: "bg-fuchsia-100", textColor: "text-fuchsia-700" },
  { name: "刘洋", role: "成员", avatar: "刘", bg: "bg-amber-100", textColor: "text-amber-700" },
];

interface GroupDetailContentProps {
  id: string;
}

export function GroupDetailContent({ id }: GroupDetailContentProps) {
  return (
    <>
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            type="button"
            aria-label="返回"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-border bg-card"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </button>
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-muted-foreground">学习群</span>
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-sm font-semibold text-foreground">{groupInfo.name} #{id}</span>
          </div>
        </div>
      </div>

      {/* Multi-column layout */}
      <div className="grid flex-1 grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Left: Group info */}
        <div className="flex w-full flex-col gap-4 overflow-y-auto rounded-[14px] border border-border bg-card p-4">
          <div className="flex items-center gap-3.5">
            <div className="flex h-[52px] w-[52px] shrink-0 items-center justify-center rounded-[14px] bg-teal-100">
              <span className="text-[22px] font-bold text-teal-600">{groupInfo.avatar}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-lg font-bold text-foreground">{groupInfo.name}</span>
              <span className="text-xs text-muted-foreground">由 {groupInfo.creator} 创建</span>
            </div>
          </div>

          <p className="text-[13px] leading-relaxed text-muted-foreground">{groupInfo.desc}</p>

          <div className="flex gap-2.5">
            {groupInfo.stats.map((stat) => (
              <div key={stat.label} className="flex flex-1 flex-col items-center gap-0.5 rounded-[10px] bg-muted py-2.5">
                <span className="text-lg font-extrabold text-teal-600">{stat.value}</span>
                <span className="text-[10px] text-muted-foreground">{stat.label}</span>
              </div>
            ))}
          </div>

          <div className="h-px bg-border" />

          <div className="flex flex-col gap-2.5">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold text-muted-foreground">当前课程游戏</span>
              <button type="button" className="rounded-md bg-teal-600/10 px-2.5 py-1 text-[11px] font-medium text-teal-600">更换</button>
            </div>
            <div className="flex items-center gap-3 rounded-[10px] border border-border bg-muted px-3 py-2.5">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-teal-100">
                <span className="text-xs font-bold text-teal-600">词</span>
              </div>
              <div className="flex flex-1 flex-col gap-0.5">
                <span className="text-[13px] font-semibold text-foreground">{groupInfo.game.name}</span>
                <span className="text-[11px] text-muted-foreground">{groupInfo.game.level}</span>
              </div>
              <span className="rounded-md bg-amber-500/10 px-2 py-1 text-[10px] font-semibold text-amber-600">{groupInfo.game.mode}</span>
            </div>
          </div>

          <div className="h-px bg-border" />

          <button type="button" className="flex w-full items-center justify-center gap-2 rounded-[10px] bg-teal-600 px-4 py-2.5 text-sm font-semibold text-white">
            <Play className="h-4 w-4" />
            进入游戏
          </button>

          <div className="h-px bg-border" />

          <div className="flex flex-col gap-1.5 px-1">
            <div className="flex items-center gap-1.5">
              <Link2 className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-[11px] font-semibold text-muted-foreground">邀请链接</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-border bg-muted px-2.5 py-2">
              <span className="flex-1 truncate text-[11px] text-muted-foreground">https://douxue.com/g/abc123</span>
              <Copy className="h-3.5 w-3.5 text-muted-foreground" />
            </div>
          </div>

          <div className="h-px bg-border" />

          <div className="flex flex-col items-center gap-2 px-1">
            <div className="flex items-center gap-1.5">
              <QrCode className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-[11px] font-semibold text-muted-foreground">二维码邀请</span>
            </div>
            <div className="flex h-[120px] w-[120px] items-center justify-center rounded-[10px] border border-border bg-muted">
              <QrCode className="h-12 w-12 text-muted-foreground" />
            </div>
          </div>
        </div>

        {/* Members list */}
        <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
          <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
            <span className="text-sm font-semibold text-foreground">群成员（128）</span>
            <button type="button" className="flex items-center gap-1 rounded-lg bg-teal-600 px-2.5 py-1 text-[11px] font-semibold text-white">
              <Users className="h-3 w-3" />
              分组
            </button>
          </div>
          <div className="flex-1 overflow-y-auto">
            {members.map((m, i) => (
              <div key={m.name}>
                {m.isOwner ? (
                  <div className="flex items-center gap-3 border-l-[3px] border-teal-600/20 bg-teal-50 px-5 py-3">
                    <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${m.bg}`}>
                      <span className={`text-sm font-semibold ${m.textColor}`}>{m.avatar}</span>
                    </div>
                    <div className="flex flex-1 flex-col gap-0.5">
                      <span className="text-[13px] font-semibold text-foreground">{m.name}</span>
                      <span className="text-[11px] text-muted-foreground">{m.role}</span>
                    </div>
                    <span className="text-xs text-amber-500">👑</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-3 px-5 py-2.5">
                    <div className={`h-[18px] w-[18px] rounded ${m.checked ? "flex items-center justify-center bg-teal-600" : "border-[1.5px] border-border bg-card"}`}>
                      {m.checked && <Check className="h-3 w-3 text-white" />}
                    </div>
                    <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${m.bg}`}>
                      <span className={`text-xs font-semibold ${m.textColor}`}>{m.avatar}</span>
                    </div>
                    <div className="flex flex-1 flex-col gap-0.5">
                      <span className="text-[13px] font-semibold text-foreground">{m.name}</span>
                      <span className="text-[11px] text-muted-foreground">{m.role}</span>
                    </div>
                  </div>
                )}
                {i < members.length - 1 && <div className="h-px bg-border" />}
              </div>
            ))}
          </div>
        </div>

        {/* Sub-groups */}
        <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
          <div className="flex items-center gap-2.5 border-b border-border px-5 py-3.5">
            <span className="text-[15px] font-semibold text-foreground">群小组（4）</span>
          </div>
          <div className="flex-1 overflow-y-auto">
            {subGroups.map((sg, i) => (
              <div key={sg.name}>
                {i > 0 && <div className="h-px bg-border" />}
                <div className="flex items-center gap-3.5 px-5 py-3.5">
                  <div className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-[10px] ${sg.bg}`}>
                    <span className="text-sm font-semibold text-foreground">{sg.avatar}</span>
                  </div>
                  <div className="flex flex-1 flex-col gap-0.5">
                    <span className="text-[13px] font-semibold text-foreground">{sg.name}</span>
                    <span className="text-[11px] text-muted-foreground">{sg.members} 名成员</span>
                  </div>
                  <span className="text-sm font-bold text-teal-600">{sg.rank}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Sub-group members */}
        <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
          <div className="border-b border-border px-5 py-3.5">
            <span className="text-sm font-semibold text-foreground">组成员（32）</span>
          </div>
          <div className="flex-1 overflow-y-auto">
            {subGroupMembers.map((m, i) => (
              <div key={m.name}>
                {i > 0 && <div className="h-px bg-border" />}
                <div className="flex items-center gap-3 px-5 py-2.5">
                  <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${m.bg}`}>
                    <span className={`text-xs font-semibold ${m.textColor}`}>{m.avatar}</span>
                  </div>
                  <div className="flex flex-1 flex-col gap-0.5">
                    <span className="text-[13px] font-semibold text-foreground">{m.name}</span>
                    <span className="text-[11px] text-muted-foreground">{m.role}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
