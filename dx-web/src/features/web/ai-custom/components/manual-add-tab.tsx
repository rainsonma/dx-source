import { BookOpen, CircleAlert, Copy, Trash2, Type } from "lucide-react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";

type ManualAddTabProps = {
  value: string;
  onChange: (value: string) => void;
  error?: string;
};

export function ManualAddTab({ value, onChange, error }: ManualAddTabProps) {
  return (
    <div className="flex flex-col gap-3 px-6 py-3">
      <div className="flex flex-col gap-2">
        <div className="flex items-start gap-2 rounded-xl bg-amber-50 px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-amber-600">学习说明：</span>
          <span className="text-amber-500">
            为了获得最佳学习体验，建议同学们科学合理的设置关卡内练习单元数量，贪多无益！每个学习关卡最多不要超过 200 个练习单元 （约 20 个语句），100 个左右最为理想，以获得最佳学习状态与学习效果。
          </span>
        </div>
        <div className="flex items-start gap-2 rounded-xl bg-muted px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-muted-foreground">格式说明：</span>
          <span className="text-muted-foreground">
            鉴于个体的学习娱乐习惯不同，我们尊重每个人的喜爱与偏好。为保证最佳学习效果，请严格按照每行填写一个句子、短语或单词；如需自定义中文释义，英文一行、中文下一行，依次交替。
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
              <BookOpen className="h-3.5 w-3.5" />
              语句输入格式示例
            </button>
          </HoverCardTrigger>
          <HoverCardContent align="start" className="w-80">
            <div className="flex flex-col gap-2">
              <p className="text-xs font-semibold text-foreground">语句输入格式示例</p>
              <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                <p>Cats can see clearly in the dark.</p>
                <p>Sunflowers always face the sun.</p>
                <p>Dolphins are one of the smartest animals.</p>
              </div>
              <p className="text-xs font-semibold text-foreground">附带中文释义格式示例</p>
              <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                <p>Cats can see clearly in the dark.</p>
                <p className="text-muted-foreground">猫在黑暗中能看得很清楚。</p>
                <p>Sunflowers always face the sun.</p>
                <p className="text-muted-foreground">向日葵总是朝向太阳。</p>
              </div>
            </div>
          </HoverCardContent>
        </HoverCard>
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
                <p>dolphin</p>
                <p>sunflower</p>
                <p>polar bear</p>
              </div>
              <p className="text-xs font-semibold text-foreground">附带中文释义格式示例</p>
              <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                <p>dolphin</p>
                <p className="text-muted-foreground">海豚</p>
                <p>sunflower</p>
                <p className="text-muted-foreground">向日葵</p>
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
        placeholder="输入学习内容，每行一条..."
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
