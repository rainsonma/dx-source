import Link from "next/link";
import { GraduationCap } from "lucide-react";
import { PLACEHOLDERS } from "@/features/com/legal/constants";

type FooterLink = { label: string; href: string };

const LEGAL_LINKS: FooterLink[] = [
  { label: "用户协议", href: "/wiki/account/user-agreement" },
  { label: "隐私政策", href: "/wiki/account/privacy-policy" },
  { label: "监护人同意书", href: "/wiki/account/guardian-consent" },
  { label: "产品服务协议", href: "/wiki/account/product-service" },
];

const PRODUCT_LINKS: FooterLink[] = [
  { label: "多种学习模式", href: "/wiki/learning-modes" },
  { label: "课程与游戏", href: "/wiki/courses-games" },
  { label: "AI 智能学习", href: "/wiki/ai" },
  { label: "词汇管理", href: "/wiki/vocabulary" },
  { label: "成长与激励", href: "/wiki/progress" },
];

const COMMUNITY_LINKS: FooterLink[] = [
  { label: "斗学社与好友", href: "/wiki/community" },
  { label: "学习小组", href: "/wiki/groups" },
  { label: "会员与能量豆", href: "/wiki/membership" },
  { label: "邀请与兑换", href: "/wiki/invites" },
  { label: "提交反馈", href: "/wiki/account/feedback" },
];

const linkColumns: { title: string; links: FooterLink[] }[] = [
  { title: "服务条款", links: LEGAL_LINKS },
  { title: "斗学产品", links: PRODUCT_LINKS },
  { title: "斗学社群", links: COMMUNITY_LINKS },
];

export function Footer() {
  return (
    <footer
      id="contact"
      className="scroll-mt-20 w-full border-t border-slate-200 bg-slate-50"
    >
      <div className="mx-auto flex max-w-[1280px] flex-col gap-12 px-[120px] pb-10 pt-[60px]">
        <div className="flex w-full flex-col gap-10 xl:flex-row xl:justify-between">
          {/* Brand */}
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2.5">
              <GraduationCap className="h-7 w-7 text-teal-600" />
              <span className="text-lg font-extrabold text-slate-900">斗学</span>
            </div>
            <p className="max-w-[280px] text-sm leading-[1.5] text-slate-500">
              玩游戏，学英语，AI 智能辅助，斗学重新定义英语学习体验，让进步自然发生...
            </p>
          </div>

          {/* Link columns */}
          <div className="grid grid-cols-1 gap-10 md:grid-cols-2 lg:grid-cols-3 xl:flex xl:gap-16">
            {linkColumns.map((col) => (
              <div key={col.title} className="flex flex-col gap-4">
                <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
                  {col.title}
                </h4>
                {col.links.map((l) => (
                  <Link
                    key={l.href}
                    href={l.href}
                    className="text-sm text-slate-500 hover:text-slate-700"
                  >
                    {l.label}
                  </Link>
                ))}
              </div>
            ))}
          </div>

          {/* Contact column */}
          <div className="flex flex-col items-start gap-4 xl:items-end">
            <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
              联系我们
            </h4>
            <div className="flex h-[140px] w-[140px] items-center justify-center rounded-lg bg-slate-100">
              <span className="text-xs text-slate-400">微信二维码</span>
            </div>
            <span className="text-xs text-slate-400">微信扫一扫联系小助手</span>
          </div>
        </div>

        <div className="h-px w-full bg-slate-200" />

        <div className="flex w-full flex-col items-center gap-2">
          <span className="text-[13px] text-slate-400">
            © 2026 douxue.fun 版权所有
          </span>
          <span className="text-[13px] text-slate-400">
            京 ICP 备 {PLACEHOLDERS.icpNumber} · 京公网安备 {PLACEHOLDERS.pscRecordNo}
          </span>
        </div>
      </div>
    </footer>
  );
}
