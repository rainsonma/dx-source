import { CircleCheck } from "lucide-react";

interface PricingCardProps {
  name: string;
  price: string;
  period: string;
  features: string[];
  bgColor: string;
  borderColor?: string;
  ctaText: string;
  highlight?: boolean;
}

export function PricingCard({
  name,
  price,
  period,
  features,
  bgColor,
  borderColor,
  ctaText,
  highlight,
}: PricingCardProps) {
  return (
    <div
      className={`flex flex-1 flex-col gap-3 rounded-[14px] p-5 ${bgColor}`}
      style={borderColor ? { border: `1px solid ${borderColor}` } : undefined}
    >
      <span className="text-base font-semibold text-white">{name}</span>
      <div className="flex items-end gap-1">
        <span className="text-4xl font-extrabold text-white">{price}</span>
        {period && <span className="mb-1 text-sm text-white/70">{period}</span>}
      </div>
      <button
        className={`w-full rounded-[10px] py-2.5 text-center text-sm font-semibold ${
          highlight
            ? "bg-white text-teal-700"
            : "border border-white/25 bg-white/20 text-white"
        }`}
      >
        {ctaText}
      </button>
      <div className="h-px w-full bg-white/30" />
      <div className="flex flex-1 flex-col gap-1.5">
        {features.map((f) => (
          <div key={f} className="flex items-center gap-2">
            <CircleCheck className="h-4 w-4 shrink-0 text-white/70" />
            <span className="text-[13px] text-white/80">{f}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
