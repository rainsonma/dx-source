"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  GraduationCap,
  LayoutDashboard,
  Gamepad2,
  Star,
  Bell,
  Sparkles,
  MessageSquare,
  Users,
  BookOpen,
  Swords,
  Trophy,
  Medal,
  MessageCircle,
  Gift,
  ChevronRight,
  ArrowUpCircle,
  RotateCcw,
  CheckCircle2,
  Ticket,
  Menu,
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet";

const navSections = [
  {
    items: [
      { icon: LayoutDashboard, label: "我的主页", href: "/hall" },
      { icon: Gamepad2, label: "课程游戏", href: "/hall/games" },
      { icon: Gamepad2, label: "我的游戏", href: "/hall/games/mine" },
      { icon: Star, label: "我的收藏", href: "/hall/favorites" },
      { icon: Swords, label: "每日挑战", href: "/hall" },
      { icon: Users, label: "学习群组", href: "/hall/groups" },
      { icon: Bell, label: "消息通知", href: "/hall/notices" },
    ],
  },
  {
    items: [
      { icon: Sparkles, label: "AI 随心配", href: "/hall/ai-custom" },
      { icon: MessageSquare, label: "AI 随心练", href: "/hall/ai-practice" },
    ],
  },
  {
    items: [
      { icon: BookOpen, label: "生词本", href: "/hall/unknown" },
      { icon: RotateCcw, label: "复习本", href: "/hall/review" },
      { icon: CheckCircle2, label: "已掌握", href: "/hall/mastered" },

      { icon: Trophy, label: "排行榜", href: "/hall/leaderboard" },
      { icon: Medal, label: "个人中心", href: "/hall/me" },
    ],
  },
  {
    items: [{ icon: MessageCircle, label: "斗学社", href: "/hall/community" }],
  },
];

function NavItem({
  icon: Icon,
  label,
  href,
  active,
  showDot,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  href: string;
  active: boolean;
  showDot?: boolean;
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className={`flex w-full items-center gap-3 rounded-[10px] px-3.5 py-2.5 ${
        active
          ? "bg-teal-600/10 font-semibold text-teal-600"
          : "text-muted-foreground hover:bg-accent"
      }`}
    >
      <Icon className="h-[18px] w-[18px]" />
      <span className="text-[13px]">{label}</span>
      {showDot && (
        <span className="ml-auto h-2 w-2 rounded-full bg-red-500" />
      )}
    </Link>
  );
}

/** CTA card configurations for the sidebar bottom section. */
const ctaItems = [
  {
    icon: Gift,
    label: "推广有奖",
    subtitle: "推广、邀请、赚佣金",
    href: "/hall/invite",
    iconGradient: "from-orange-400 to-red-500",
    badge: { text: "HOT", gradient: "from-orange-400 to-red-500" },
  },
  {
    icon: ArrowUpCircle,
    label: "续费升级",
    subtitle: "选择会员套餐",
    href: "/auth/membership",
    iconGradient: "from-amber-300 to-yellow-500",
    badge: { text: "VIP", gradient: "from-amber-300 to-yellow-500" },
  },
  {
    icon: Ticket,
    label: "兑换码",
    subtitle: "兑换码兑换会员",
    href: "/hall/redeem",
    iconGradient: "from-violet-400 to-purple-600",
  },
];

function CtaCard({
  icon: Icon,
  label,
  subtitle,
  href,
  iconGradient,
  badge,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  subtitle: string;
  href: string;
  iconGradient: string;
  badge?: { text: string; gradient: string };
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className="flex w-full items-center justify-between rounded-[10px] border border-border px-3.5 py-3 hover:bg-accent"
    >
      <div className="flex items-center gap-3">
        <div
          className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${iconGradient}`}
        >
          <Icon className="h-4 w-4 text-white" />
        </div>
        <div className="flex flex-col gap-0.5">
          <div className="flex items-center gap-1.5">
            <span className="text-[13px] font-medium text-foreground">
              {label}
            </span>
            {badge && (
              <span
                className={`rounded-full bg-gradient-to-r px-1.5 py-0.5 text-[10px] font-semibold text-white ${badge.gradient}`}
              >
                {badge.text}
              </span>
            )}
          </div>
          <span className="text-[11px] text-muted-foreground">{subtitle}</span>
        </div>
      </div>
      <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
    </Link>
  );
}

function SidebarContent({ onNavigate, hasUnreadNotices }: { onNavigate?: () => void; hasUnreadNotices?: boolean }) {
  const pathname = usePathname();

  return (
    <div className="flex h-full flex-col">
      {/* Header — pinned top */}
      <div className="shrink-0">
        <div className="flex w-full items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <GraduationCap className="h-7 w-7 text-teal-600" />
            <span className="text-lg font-extrabold text-foreground">斗学</span>
          </Link>
          <Swords className="h-[18px] w-[18px] text-muted-foreground" />
        </div>
      </div>

      {/* Nav — scrollable middle */}
      <nav className="mt-8 flex-1 overflow-y-auto min-h-0 flex flex-col gap-1">
        {navSections.map((section) => (
          <div key={section.items[0].label} className="flex flex-col gap-1">
            {section !== navSections[0] && (
              <div className="my-1 h-px w-full bg-border" />
            )}
            {section.items.map((item) => (
              <NavItem
                key={item.label}
                icon={item.icon}
                label={item.label}
                href={item.href}
                active={pathname === item.href}
                showDot={item.href === "/hall/notices" && hasUnreadNotices}
                onClick={onNavigate}
              />
            ))}
          </div>
        ))}
      </nav>

      {/* Bottom CTAs — pinned bottom */}
      <div className="shrink-0 mt-4 flex flex-col gap-2">
        {ctaItems.map((item) => (
          <CtaCard key={item.label} {...item} onClick={onNavigate} />
        ))}
      </div>
    </div>
  );
}

export function HallSidebar({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <aside className="hidden md:flex h-full w-[260px] shrink-0 flex-col border-r border-border bg-card px-5 py-6">
      <SidebarContent hasUnreadNotices={hasUnreadNotices} />
    </aside>
  );
}

export function MobileSidebarTrigger({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <button
          type="button"
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground hover:bg-accent"
        >
          <Menu className="h-5 w-5" />
        </button>
      </SheetTrigger>
      <SheetContent side="left" className="w-[260px] p-5" showCloseButton={false}>
        <SidebarContent hasUnreadNotices={hasUnreadNotices} />
      </SheetContent>
    </Sheet>
  );
}
