"use client";

import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Footer } from "@/components/in/footer";

export default function PurchaseLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();

  return (
    <div className="flex min-h-screen w-full flex-col items-center bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white">
      <header className="flex w-full max-w-[1200px] items-center px-4 pt-6 lg:px-8">
        <button
          type="button"
          onClick={() => router.back()}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-white/60 hover:bg-white/80"
        >
          <ArrowLeft className="h-5 w-5 text-slate-600" />
        </button>
      </header>
      <main className="flex w-full flex-1 flex-col items-center">
        {children}
      </main>
      <Footer />
    </div>
  );
}
