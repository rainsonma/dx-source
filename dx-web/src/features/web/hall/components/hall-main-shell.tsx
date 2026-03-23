"use client";

import { GraduationCap } from "lucide-react";
import { HallSidebar, MobileSidebarTrigger } from "@/features/web/hall/components/hall-sidebar";

/** Client shell for the hall main layout — receives server-computed props */
export function HallMainShell({
  children,
  hasUnreadNotices,
}: {
  children: React.ReactNode;
  hasUnreadNotices?: boolean;
}) {
  return (
    <div className="flex h-screen w-full overflow-hidden bg-muted">
      {/* Mobile header */}
      <div className="fixed top-0 left-0 right-0 z-40 flex h-14 items-center gap-3 border-b bg-card px-4 md:hidden">
        <MobileSidebarTrigger hasUnreadNotices={hasUnreadNotices} />
        <div className="flex items-center gap-2">
          <GraduationCap className="h-6 w-6 text-teal-600" />
          <span className="text-base font-extrabold text-foreground">斗学</span>
        </div>
      </div>

      {/* Desktop sidebar */}
      <HallSidebar hasUnreadNotices={hasUnreadNotices} />

      {/* Main content */}
      <main className="flex-1 overflow-y-auto pt-14 md:pt-0">{children}</main>
    </div>
  );
}
