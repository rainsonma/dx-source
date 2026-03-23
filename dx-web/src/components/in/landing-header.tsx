"use client";

import Link from "next/link";
import { GraduationCap, SquareArrowRightEnter } from "lucide-react";
import { useEffect, useState } from "react";
import { getAccessToken } from "@/lib/token";
import { MobileNav } from "@/components/in/mobile-nav";

export function LandingHeader() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    setIsLoggedIn(!!getAccessToken());
  }, []);

  return (
    <header className="flex h-20 w-full items-center justify-between px-5 lg:px-20">
      <Link href="/" className="flex items-center gap-2.5">
        <GraduationCap className="h-9 w-9 text-teal-600" />
        <span className="text-[22px] font-semibold text-slate-900">斗学</span>
      </Link>
      <nav className="hidden items-center gap-9 lg:flex">
        <Link
          href="/docs"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          文档
        </Link>
        <Link
          href="/features"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          功能
        </Link>
        <Link
          href="/changelog"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          更新日志
        </Link>
        <Link
          href="#faq"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          常见问题
        </Link>
        <Link
          href="#contact"
          className="text-[15px] font-medium text-slate-500 hover:text-slate-700"
        >
          联系我们
        </Link>
      </nav>
      <div className="flex items-center gap-3">
        {isLoggedIn ? (
          <Link
            href="/hall"
            className="hidden items-center gap-2 rounded-lg bg-teal-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-teal-700 lg:inline-flex"
          >
            进入学习大厅
            <SquareArrowRightEnter className="h-4 w-4" />
          </Link>
        ) : (
          <>
            <Link
              href="/auth/signin"
              className="hidden rounded-lg border border-slate-300 px-6 py-2.5 text-sm font-medium text-slate-900 hover:bg-slate-50 lg:inline-flex"
            >
              登录
            </Link>
            <Link
              href="/auth/signup"
              className="hidden rounded-lg bg-teal-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-teal-700 lg:inline-flex"
            >
              注册
            </Link>
          </>
        )}
        <MobileNav isLoggedIn={isLoggedIn} />
      </div>
    </header>
  );
}
