import { Loader2 } from "lucide-react";

type ProcessingOverlayProps = {
  done: number;
  total: number;
};

export function ProcessingOverlay({ done, total }: ProcessingOverlayProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="flex flex-col items-center gap-4 rounded-2xl bg-card px-10 py-8 shadow-2xl">
        <Loader2 className="h-10 w-10 animate-spin text-teal-600" />
        <p className="text-sm font-medium text-foreground">
          {done} / {total} 大模型正在处理，请耐心等待 ...
        </p>
      </div>
    </div>
  );
}
