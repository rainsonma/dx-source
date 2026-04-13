"use client";

import { Menu } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { DocsSidebar } from "./docs-sidebar";

export function DocsSidebarDrawer() {
  return (
    <Sheet>
      <SheetTrigger className="flex items-center gap-1.5 text-sm font-medium text-slate-700 hover:text-slate-900">
        <Menu className="h-4 w-4" aria-hidden="true" />
        目录
      </SheetTrigger>
      <SheetContent side="left" className="w-[280px] overflow-y-auto p-5">
        <SheetHeader>
          <SheetTitle className="text-left text-sm font-extrabold">
            斗学文档
          </SheetTitle>
        </SheetHeader>
        <div className="mt-4">
          <DocsSidebar />
        </div>
      </SheetContent>
    </Sheet>
  );
}
