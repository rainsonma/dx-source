import Link from "next/link";
import { GraduationCap } from "lucide-react";

export function AuthHeader() {
  return (
    <header className="flex h-20 w-full items-center px-20">
      <Link href="/" className="flex items-center gap-2.5">
        <GraduationCap className="h-9 w-9 text-teal-600" />
        <span className="text-[22px] font-semibold text-slate-900">斗学</span>
      </Link>
    </header>
  );
}
