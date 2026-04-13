import type { LucideIcon } from "lucide-react";

type Item = {
  icon: LucideIcon;
  iconColor: string;
  title: string;
  desc: string;
};

type Props = {
  columns?: 2 | 3 | 4;
  items: Item[];
};

export function DocFeatureGrid({ columns = 3, items }: Props) {
  const colsClass =
    columns === 2
      ? "lg:grid-cols-2"
      : columns === 4
        ? "lg:grid-cols-4"
        : "lg:grid-cols-3";
  return (
    <div className={`grid grid-cols-1 gap-4 sm:grid-cols-2 ${colsClass}`}>
      {items.map((item, i) => {
        const Icon = item.icon;
        return (
          <div
            key={i}
            className="flex flex-col gap-2.5 rounded-[10px] border border-slate-200 bg-white p-5"
          >
            <Icon className={`h-7 w-7 ${item.iconColor}`} aria-hidden="true" />
            <span className="text-[15px] font-semibold text-slate-900">
              {item.title}
            </span>
            <span className="text-[13px] leading-snug text-slate-500">
              {item.desc}
            </span>
          </div>
        );
      })}
    </div>
  );
}
