type Item = {
  key: string;
  value: string;
  note?: string;
};

type Props = { items: Item[] };

export function DocKeyValue({ items }: Props) {
  return (
    <div className="rounded-[10px] border border-slate-200 bg-white">
      {items.map((item, i) => (
        <div
          key={i}
          className={`flex flex-col gap-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between ${
            i < items.length - 1 ? "border-b border-slate-100" : ""
          }`}
        >
          <span className="text-sm font-medium text-slate-700">
            {item.key}
          </span>
          <span className="text-sm text-slate-600">
            {item.value}
            {item.note && (
              <span className="ml-2 text-xs text-slate-400">{item.note}</span>
            )}
          </span>
        </div>
      ))}
    </div>
  );
}
