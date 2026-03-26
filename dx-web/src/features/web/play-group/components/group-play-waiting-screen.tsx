"use client";

import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";

interface GroupPlayWaitingScreenProps {
  groupId: string;
}

export function GroupPlayWaitingScreen({ groupId }: GroupPlayWaitingScreenProps) {
  return (
    <div className="flex min-h-full flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <div className="flex flex-col items-center gap-3 py-4">
          <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
          <p className="text-base font-medium text-muted-foreground">正在计算结果...</p>
        </div>

        <Button variant="outline" className="w-full" asChild>
          <Link href={`/hall/groups/${groupId}`}>返回群组</Link>
        </Button>
      </div>
    </div>
  );
}
