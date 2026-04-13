import { Check, X } from "lucide-react";
import type { ReactNode } from "react";

type Cell = boolean | string | ReactNode;

type Row = {
  label: string;
  values: Cell[];
};

type Props = {
  columns: string[];
  rows: Row[];
  labelHeader?: string;
};

export function DocCompareTable({
  columns,
  rows,
  labelHeader = "功能 / 权益",
}: Props) {
  return (
    <div className="overflow-x-auto rounded-[10px] border border-slate-200">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200 bg-slate-50">
            <th className="px-4 py-3 text-left font-semibold text-slate-700">
              {labelHeader}
            </th>
            {columns.map((col) => (
              <th
                key={col}
                className="px-4 py-3 text-center font-semibold text-slate-700"
              >
                {col}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={i}
              className="border-b border-slate-100 last:border-b-0"
            >
              <td className="px-4 py-3 font-medium text-slate-700">
                {row.label}
              </td>
              {row.values.map((v, j) => (
                <td
                  key={j}
                  className="px-4 py-3 text-center text-slate-600"
                >
                  {typeof v === "boolean" ? (
                    v ? (
                      <Check
                        className="mx-auto h-4 w-4 text-teal-600"
                        aria-hidden="true"
                      />
                    ) : (
                      <X
                        className="mx-auto h-4 w-4 text-slate-300"
                        aria-hidden="true"
                      />
                    )
                  ) : (
                    v
                  )}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
