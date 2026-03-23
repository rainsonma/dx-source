const stats: {
  value: string;
  label: string;
  gradient: string;
}[] = [
  {
    value: "500K+",
    label: "活跃学习者",
    gradient: "from-teal-400 to-violet-500",
  },
  {
    value: "10M+",
    label: "课程已完成",
    gradient: "from-violet-500 to-teal-400",
  },
  {
    value: "92%",
    label: "留存率",
    gradient: "from-green-400 to-teal-600",
  },
  {
    value: "4.9",
    label: "应用商店评分",
    gradient: "from-yellow-400 to-orange-500",
  },
];

function StatCard({ stat }: { stat: (typeof stats)[number] }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-2xl border border-slate-200 bg-white px-8 py-10 text-center shadow-[0_4px_16px_rgba(15,23,42,0.03)]">
      <span
        className={`bg-gradient-to-r ${stat.gradient} bg-clip-text text-5xl font-extrabold text-transparent`}
      >
        {stat.value}
      </span>
      <span className="text-[15px] font-medium text-slate-500">
        {stat.label}
      </span>
    </div>
  );
}

export function StatsSection() {
  return (
    <section className="flex w-full flex-col items-center gap-[60px] bg-gradient-to-b from-orange-50 to-violet-50 px-[120px] py-[100px]">
      <div className="flex flex-col items-center gap-4">
        <span className="text-sm font-semibold tracking-wide text-emerald-600">
          数据说话
        </span>
        <h2 className="text-4xl font-extrabold tracking-tight text-slate-900">
          学习者爱上这款游戏
        </h2>
      </div>
      <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <StatCard key={stat.label} stat={stat} />
        ))}
      </div>
    </section>
  );
}
