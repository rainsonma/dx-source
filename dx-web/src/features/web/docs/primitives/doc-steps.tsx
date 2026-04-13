type Step = { title: string; desc: string };

type Props = { steps: Step[] };

export function DocSteps({ steps }: Props) {
  return (
    <div className="flex flex-col gap-3">
      {steps.map((step, i) => (
        <div key={i} className="flex gap-3">
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-teal-600">
            <span className="text-xs font-bold text-white">{i + 1}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm font-semibold text-slate-900">
              {step.title}
            </span>
            <span className="text-[13px] text-slate-500">{step.desc}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
