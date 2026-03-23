const changelog = [
  {
    version: "v2.3.0",
    date: "2026 年 2 月 10 日",
    latest: true,
    changes: [
      "新增：AI 智能推荐课程功能上线",
      "新增：单词本支持自定义分组",
      "优化：游戏加载速度提升 40%",
      "修复：听说读写计时器偶发错误",
    ],
  },
  {
    version: "v2.2.0",
    date: "2026 年 1 月 15 日",
    latest: false,
    changes: [
      "新增：词汇对战排行榜",
      "新增：学习报告每周邮件推送",
      "优化：移动端适配改善",
      "修复：登录状态偶发失效问题",
    ],
  },
  {
    version: "v2.1.0",
    date: "2025 年 12 月 20 日",
    latest: false,
    changes: [
      "新增：听力闯关新增 3 套题库",
      "新增：好友系统开放内测",
      "修复：iOS 端音频播放问题",
      "修复：积分计算偶现偏差",
    ],
  },
];

export function ChangelogTimeline() {
  return (
    <div className="flex flex-1 bg-slate-50">
      {/* Left spacer for visual balance */}
      <div className="hidden w-[220px] shrink-0 lg:block" />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-8 px-4 py-6 md:px-8 lg:px-14 lg:py-10">
        {/* Title */}
        <div className="flex flex-col gap-3">
          <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            更新日志
          </h1>
          <p className="text-base text-slate-500">
            记录每一次迭代，让产品成长历程清晰可见。
          </p>
        </div>

        <div className="h-px w-full bg-slate-200" />

        {/* Timeline */}
        <div className="flex flex-col">
          {changelog.map((entry, i) => (
            <div key={entry.version} className="flex">
              {/* Timeline indicator */}
              <div className="flex w-10 flex-col items-center pt-1.5">
                <div
                  className={`h-3.5 w-3.5 rounded-full ${
                    entry.latest ? "bg-teal-600" : "bg-slate-500"
                  }`}
                />
                {i < changelog.length - 1 && (
                  <div className="w-0.5 flex-1 bg-slate-200" />
                )}
              </div>

              {/* Content */}
              <div className={`flex flex-col gap-3 pb-10 pl-5 ${i === changelog.length - 1 ? "pb-5" : ""}`}>
                <div className="flex flex-wrap items-center gap-2.5">
                  <span
                    className={`rounded-md border px-2.5 py-1 text-sm font-bold ${
                      entry.latest
                        ? "border-teal-600 text-teal-600"
                        : "border-slate-400 text-slate-500"
                    }`}
                  >
                    {entry.version}
                  </span>
                  <span className="text-[13px] text-slate-400">{entry.date}</span>
                  {entry.latest && (
                    <span className="rounded-full bg-green-100 px-2.5 py-0.5 text-xs text-green-600">
                      最新
                    </span>
                  )}
                </div>
                <div className="flex flex-col gap-1">
                  {entry.changes.map((change) => (
                    <p key={change} className="text-sm leading-[1.9] text-slate-600">
                      {change}
                    </p>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Right TOC */}
      <div className="hidden w-[220px] shrink-0 border-l border-slate-200 px-5 py-10 lg:block">
        <div className="flex flex-col gap-3">
          {changelog.map((entry) => (
            <div key={entry.version} className="flex items-center gap-2">
              <div
                className={`h-4 w-0.5 rounded-sm ${
                  entry.latest ? "bg-teal-600" : "bg-slate-300"
                }`}
              />
              <span
                className={`text-[13px] ${
                  entry.latest ? "font-medium text-teal-600" : "text-slate-500"
                }`}
              >
                {entry.version}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
