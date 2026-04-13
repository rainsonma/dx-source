import { ChevronRight } from "lucide-react";
import { Fragment } from "react";

type Node = {
  label: string;
  desc?: string;
};

type Props = { nodes: Node[] };

export function DocFlow({ nodes }: Props) {
  return (
    <div className="flex flex-col gap-3 rounded-[10px] border border-slate-200 bg-slate-50 p-4 sm:flex-row sm:items-stretch">
      {nodes.map((node, i) => (
        <Fragment key={i}>
          <div className="flex flex-1 flex-col gap-1 rounded-md bg-white px-3 py-3 text-center">
            <span className="text-sm font-semibold text-slate-900">
              {node.label}
            </span>
            {node.desc && (
              <span className="text-xs text-slate-500">{node.desc}</span>
            )}
          </div>
          {i < nodes.length - 1 && (
            <div className="flex items-center justify-center sm:px-1">
              <ChevronRight
                className="h-5 w-5 rotate-90 text-slate-400 sm:rotate-0"
                aria-hidden="true"
              />
            </div>
          )}
        </Fragment>
      ))}
    </div>
  );
}
