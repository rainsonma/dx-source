"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Loader2, Play } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { groupApi } from "../actions/group.action";

const DEGREES = [
  { value: "practice", label: "练习" },
  { value: "beginner", label: "初级" },
  { value: "intermediate", label: "中级" },
  { value: "advanced", label: "高级" },
];

const PATTERNS = [
  { value: "listen", label: "听" },
  { value: "speak", label: "说" },
  { value: "read", label: "读" },
  { value: "write", label: "写" },
];

interface StartGameDialogProps {
  groupId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onStarted: () => void;
}

export function StartGameDialog({ groupId, open, onOpenChange, onStarted }: StartGameDialogProps) {
  const [degree, setDegree] = useState("intermediate");
  const [pattern, setPattern] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);

  async function handleStart() {
    setLoading(true);
    const res = await groupApi.startGame(groupId, degree, pattern);
    setLoading(false);
    if (res.code !== 0) {
      toast.error(res.message);
      return;
    }
    toast.success("游戏已开始");
    onOpenChange(false);
    onStarted();
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>开始游戏</DialogTitle>
          <DialogDescription>选择难度和模式后开始群组游戏</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <span className="text-xs font-medium text-muted-foreground">难度</span>
            <div className="flex flex-wrap gap-2">
              {DEGREES.map((d) => (
                <button
                  key={d.value}
                  type="button"
                  onClick={() => setDegree(d.value)}
                  className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                    degree === d.value
                      ? "bg-teal-600 text-white"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {d.label}
                </button>
              ))}
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <span className="text-xs font-medium text-muted-foreground">模式（可选）</span>
            <div className="flex flex-wrap gap-2">
              {PATTERNS.map((p) => (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => setPattern(pattern === p.value ? undefined : p.value)}
                  className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                    pattern === p.value
                      ? "bg-teal-600 text-white"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {p.label}
                </button>
              ))}
            </div>
          </div>
          <Button onClick={handleStart} disabled={loading} className="w-full">
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
            开始游戏
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
