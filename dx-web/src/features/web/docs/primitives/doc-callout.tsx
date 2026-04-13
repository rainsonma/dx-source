import {
  AlertTriangle,
  CheckCircle2,
  Info,
  Lightbulb,
  type LucideIcon,
} from "lucide-react";
import type { ReactNode } from "react";

type Variant = "info" | "tip" | "warning" | "success";

const VARIANTS: Record<
  Variant,
  {
    icon: LucideIcon;
    boxClass: string;
    titleClass: string;
    bodyClass: string;
  }
> = {
  info: {
    icon: Info,
    boxClass: "border-teal-200 bg-teal-50",
    titleClass: "text-teal-700",
    bodyClass: "text-teal-700",
  },
  tip: {
    icon: Lightbulb,
    boxClass: "border-amber-200 bg-amber-50",
    titleClass: "text-amber-700",
    bodyClass: "text-amber-700",
  },
  warning: {
    icon: AlertTriangle,
    boxClass: "border-rose-200 bg-rose-50",
    titleClass: "text-rose-700",
    bodyClass: "text-rose-700",
  },
  success: {
    icon: CheckCircle2,
    boxClass: "border-emerald-200 bg-emerald-50",
    titleClass: "text-emerald-700",
    bodyClass: "text-emerald-700",
  },
};

type Props = {
  variant?: Variant;
  title?: string;
  children: ReactNode;
};

export function DocCallout({ variant = "info", title, children }: Props) {
  const { icon: Icon, boxClass, titleClass, bodyClass } = VARIANTS[variant];
  return (
    <div className={`flex gap-3 rounded-lg border p-4 ${boxClass}`}>
      <Icon className={`h-5 w-5 shrink-0 ${titleClass}`} />
      <div className="flex flex-col gap-1">
        {title && (
          <span className={`text-[13px] font-semibold ${titleClass}`}>
            {title}
          </span>
        )}
        <div className={`text-sm leading-snug ${bodyClass}`}>{children}</div>
      </div>
    </div>
  );
}
