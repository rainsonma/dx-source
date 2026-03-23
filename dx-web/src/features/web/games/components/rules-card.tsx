import { Info } from "lucide-react";

export function RulesCard({ rules }: { rules: string[] }) {
  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-6">
      <div className="flex items-center gap-2">
        <Info className="h-[18px] w-[18px] text-teal-600" />
        <h3 className="text-base font-bold text-foreground">游戏规则</h3>
      </div>
      <div className="flex flex-col gap-3">
        {rules.map((rule) => (
          <div key={rule} className="flex items-start gap-2.5">
            <div className="mt-[7px] h-1.5 w-1.5 shrink-0 rounded-full bg-teal-600" />
            <span className="text-[13px] leading-[1.5] text-muted-foreground">
              {rule}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
