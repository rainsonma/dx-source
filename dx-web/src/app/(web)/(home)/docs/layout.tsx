import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { LandingHeader } from "@/components/in/landing-header";
import { Footer } from "@/components/in/footer";
import { DocsSidebar } from "@/features/web/docs/components/docs-sidebar";
import { DocsSidebarDrawer } from "@/features/web/docs/components/docs-sidebar-drawer";

export default async function DocsLayout({
  children,
}: {
  children: ReactNode;
}) {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_token")?.value;

  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader isLoggedIn={isLoggedIn} />
      <div className="h-px w-full bg-slate-200" />

      <div className="sticky top-0 z-10 flex h-12 items-center gap-2 border-b border-slate-200 bg-white px-4 lg:hidden">
        <DocsSidebarDrawer />
      </div>

      <div className="flex flex-1 bg-slate-50">
        <div className="mx-auto flex w-full max-w-[1280px]">
          <aside className="hidden w-[260px] shrink-0 border-r border-slate-200 bg-white px-5 py-6 lg:block">
            <DocsSidebar />
          </aside>
          <main className="flex flex-1 flex-col gap-8 px-4 py-6 md:px-8 lg:px-14 lg:py-10">
            {children}
          </main>
        </div>
      </div>

      <Footer />
    </div>
  );
}
