"use client";

import { useRouter } from "next/navigation";
import { BatteryWarning } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type InsufficientBeansDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  required: number;
  available: number;
};

/** Alert dialog shown when user lacks enough energy beans for an AI operation */
export function InsufficientBeansDialog({
  open,
  onOpenChange,
  required,
  available,
}: InsufficientBeansDialogProps) {
  const router = useRouter();

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <BatteryWarning className="h-5 w-5 text-amber-500" />
            能量豆不足
          </AlertDialogTitle>
          <AlertDialogDescription>
            本次操作需要{" "}
            <span className="font-semibold text-foreground">{required}</span>{" "}
            能量豆，当前余额{" "}
            <span className="font-semibold text-foreground">{available}</span>
            ，还差{" "}
            <span className="font-semibold text-red-600">
              {required - available}
            </span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            className="bg-teal-600 hover:bg-teal-700"
            onClick={() => router.push("/recharge")}
          >
            去充值
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
