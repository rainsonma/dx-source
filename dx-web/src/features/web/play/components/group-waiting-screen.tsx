"use client";

import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";

interface GroupWaitingScreenProps {
  groupId: string;
}

export function GroupWaitingScreen({ groupId }: GroupWaitingScreenProps) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-6">
      <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
      <p className="text-lg font-medium text-muted-foreground">等待其他选手完成...</p>
      <Button variant="outline" asChild>
        <Link href={`/hall/groups/${groupId}`}>返回群组</Link>
      </Button>
    </div>
  );
}
