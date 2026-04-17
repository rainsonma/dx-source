"use client";

import { useState } from "react";
import Link from "next/link";
import { Menu, SquareArrowRightEnter } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { HashNavLink } from "@/components/in/hash-nav-link";

type NavLink =
  | { kind: "route"; href: string; label: string }
  | { kind: "hash"; hash: string; label: string };

const navLinks: NavLink[] = [
  { kind: "route", href: "/wiki", label: "Wiki" },
  { kind: "hash", hash: "features", label: "Features" },
  { kind: "hash", hash: "faq", label: "常见问题" },
  { kind: "hash", hash: "contact", label: "联系我们" },
];

interface MobileNavProps {
  isLoggedIn?: boolean;
}

export function MobileNav({ isLoggedIn }: MobileNavProps) {
  const [open, setOpen] = useState(false);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger className="flex h-10 w-10 items-center justify-center rounded-lg border border-slate-300 lg:hidden">
        <Menu className="h-5 w-5 text-slate-600" />
      </SheetTrigger>
      <SheetContent side="right" className="w-72">
        <SheetHeader>
          <SheetTitle className="text-left text-lg">菜单</SheetTitle>
        </SheetHeader>
        <nav className="flex flex-col gap-1 px-4">
          {navLinks.map((link) => {
            const className =
              "rounded-md px-3 py-2.5 text-[15px] font-medium text-slate-600 hover:bg-slate-100 hover:text-slate-900";
            if (link.kind === "hash") {
              return (
                <HashNavLink
                  key={link.label}
                  hash={link.hash}
                  className={className}
                  onNavigate={() => setOpen(false)}
                >
                  {link.label}
                </HashNavLink>
              );
            }
            return (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setOpen(false)}
                className={className}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>
        <div className="mt-auto flex flex-col gap-3 border-t border-slate-200 p-4">
          {isLoggedIn ? (
            <Link
              href="/hall"
              onClick={() => setOpen(false)}
              className="flex items-center justify-center gap-2 rounded-lg bg-teal-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-teal-700"
            >
              进入学习大厅
              <SquareArrowRightEnter className="h-4 w-4" />
            </Link>
          ) : (
            <>
              <Link
                href="/auth/signin"
                onClick={() => setOpen(false)}
                className="rounded-lg border border-slate-300 px-6 py-2.5 text-center text-sm font-medium text-slate-900 hover:bg-slate-50"
              >
                登录
              </Link>
              <Link
                href="/auth/signup"
                onClick={() => setOpen(false)}
                className="rounded-lg bg-teal-600 px-6 py-2.5 text-center text-sm font-semibold text-white hover:bg-teal-700"
              >
                注册
              </Link>
            </>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
