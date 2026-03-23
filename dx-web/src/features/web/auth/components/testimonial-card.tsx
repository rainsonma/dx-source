interface TestimonialCardProps {
  name: string;
  tier: string;
  quote: string;
}

export function TestimonialCard({ name, tier, quote }: TestimonialCardProps) {
  return (
    <div className="flex flex-1 flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-6 shadow-[0_8px_24px_rgba(15,23,42,0.09)]">
      <div className="flex items-center gap-3">
        <div className="h-12 w-12 shrink-0 rounded-full bg-gradient-to-br from-slate-200 to-slate-300" />
        <div className="flex flex-col gap-0.5">
          <span className="text-lg font-bold text-slate-900">{name}</span>
          <span className="text-[13px] text-slate-500">{tier}</span>
        </div>
      </div>
      <p className="text-sm leading-[1.6] text-slate-700">{quote}</p>
    </div>
  );
}
