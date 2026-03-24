"use client";

import { useState, useEffect } from "react";
import { Loader2 } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createGroupSchema } from "../schemas/group.schema";

interface EditGroupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  name: string;
  description: string | null;
  onSave: (name: string, description?: string) => Promise<boolean>;
}

export function EditGroupDialog({ open, onOpenChange, name: initialName, description: initialDesc, onSave }: EditGroupDialogProps) {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDesc ?? "");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (open) {
      setName(initialName);
      setDescription(initialDesc ?? "");
      setErrors({});
    }
  }, [open, initialName, initialDesc]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrors({});

    const result = createGroupSchema.safeParse({ name, description: description || undefined });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = issue.path[0] as string;
        if (!fieldErrors[key]) fieldErrors[key] = issue.message;
      }
      setErrors(fieldErrors);
      return;
    }

    setPending(true);
    try {
      const ok = await onSave(name, description || undefined);
      if (ok) {
        onOpenChange(false);
      }
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑群组</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">群名称</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="输入群名称"
              maxLength={50}
            />
            {errors.name && <p className="text-xs text-red-500">{errors.name}</p>}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">群描述（选填）</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="介绍一下这个群"
              maxLength={200}
              rows={3}
              className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
            {errors.description && <p className="text-xs text-red-500">{errors.description}</p>}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={pending} className="bg-teal-600 hover:bg-teal-700">
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
