import { Type, CircleAlert, Copy, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";

type VocabManualTabProps = {
  value: string;
  onChange: (value: string) => void;
  error?: string;
  maxPairs: number;
  batchSize: number;
};

export function VocabManualTab({ value, onChange, error, maxPairs, batchSize }: VocabManualTabProps) {
  return (
    <div className="flex flex-col gap-3 px-6 py-3">
      <div className="flex flex-col gap-2">
        <div className="flex items-start gap-2 rounded-xl bg-amber-50 px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-amber-600">输入说明：</span>
          <span className="text-amber-500">
            请输入英文-中文词汇对，英文一行、中文释义下一行，依次交替。每次最多添加 {maxPairs} 对词汇{batchSize > 0 ? `，数量须为 ${batchSize} 的倍数` : ""}。
          </span>
        </div>
        <div className="flex items-start gap-2 rounded-xl bg-muted px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-muted-foreground">格式说明：</span>
          <span className="text-muted-foreground">
            严格按照英文一行、中文释义下一行的格式输入，每对词汇占两行，不要混合在同一行。
          </span>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <HoverCard openDelay={200}>
          <HoverCardTrigger asChild>
            <button
              type="button"
              className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent"
            >
              <Type className="h-3.5 w-3.5" />
              词汇输入格式示例
            </button>
          </HoverCardTrigger>
          <HoverCardContent align="start" className="w-80">
            <div className="flex flex-col gap-2">
              <p className="text-xs font-semibold text-foreground">词汇输入格式示例</p>
              <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                <p>apple</p>
                <p className="text-muted-foreground">苹果</p>
                <p>banana</p>
                <p className="text-muted-foreground">香蕉</p>
                <p>polar bear</p>
                <p className="text-muted-foreground">北极熊</p>
              </div>
            </div>
          </HoverCardContent>
        </HoverCard>
        <div className="ml-auto flex items-center gap-2">
          <button
            type="button"
            disabled={!value}
            onClick={async () => {
              await navigator.clipboard.writeText(value);
              toast.success("已复制到剪贴板");
            }}
            className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
          >
            <Copy className="h-3.5 w-3.5" />
            复制
          </button>
          <button
            type="button"
            disabled={!value}
            onClick={() => onChange("")}
            className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
            清空
          </button>
        </div>
      </div>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="输入词汇，英文一行、中文下一行..."
        className={`h-[280px] resize-none overflow-y-auto rounded-xl bg-muted px-4 py-3.5 text-[13px] leading-[1.8] text-foreground shadow-none focus-visible:ring-1 ${error ? "border-red-400 focus-visible:ring-red-400" : "border-border focus-visible:ring-teal-500"}`}
      />
      {error && (
        <p className="flex items-center gap-1.5 text-xs text-red-500">
          <CircleAlert className="h-3.5 w-3.5 shrink-0" />
          {error}
        </p>
      )}
    </div>
  );
}
