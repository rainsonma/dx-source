"use client";

import { useRouter } from "next/navigation";
import { authApi } from "@/lib/api-client";
import {
  Zap,
  Coins,
  Crown,
  Ticket,
  Gift,
  FileText,
  LogOut,
  ChevronDown,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getLevel } from "@/consts/user-level";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { UserProfile } from "@/features/web/auth/types/user.types";

const avatarColors = [
  "#ef4444", "#f97316", "#f59e0b", "#eab308", "#84cc16",
  "#22c55e", "#14b8a6", "#06b6d4", "#0ea5e9", "#3b82f6",
  "#6366f1", "#8b5cf6", "#a855f7", "#d946ef", "#ec4899",
];

function getAvatarColor(id: string) {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0;
  }
  return avatarColors[Math.abs(hash) % avatarColors.length];
}

const menuItems = [
  {
    group: [
      { label: "升级会员", icon: Crown, href: "/purchase/membership" },
      { label: "兑换会员", icon: Ticket, href: "/hall/redeem" },
      { label: "推广邀请", icon: Gift, href: "/hall/invite" },
    ],
  },
  {
    group: [{ label: "帮助文档", icon: FileText, href: "/docs" }],
  },
];

export function UserProfileMenu({ profile }: { profile: UserProfile }) {
  const router = useRouter();
  const displayName = profile.nickname ?? profile.username;
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const gradeLabel = USER_GRADE_LABELS[profile.grade];
  const level = getLevel(profile.exp);
  const avatarBg = getAvatarColor(profile.id);

  function handleNavigate(href: string) {
    router.push(href);
  }

  async function handleSignOut() {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout API errors — clear local state regardless
    }
    window.location.href = "/";
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="flex items-center gap-2.5 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2">
          <div className="h-6 w-px bg-border" />
          <Avatar>
            {profile.avatarUrl && (
              <AvatarImage src={profile.avatarUrl} alt={displayName} />
            )}
            <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>{fallbackChar}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col items-start gap-0.5">
            <span className="text-sm font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-1.5">
              <span className="rounded-full bg-indigo-100 px-2 py-0.5 text-[10px] font-bold text-indigo-600">
                Lv.{level}
              </span>
              <span className="rounded bg-border px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground">
                {gradeLabel}
              </span>
            </div>
          </div>
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent className="w-56" align="end">
        <DropdownMenuLabel className="flex items-center gap-3 px-3 py-2.5">
          <Avatar size="lg">
            {profile.avatarUrl && (
              <AvatarImage src={profile.avatarUrl} alt={displayName} />
            )}
            <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>{fallbackChar}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <div className="flex items-center gap-1.5">
              <span className="text-sm font-semibold">{displayName}</span>
              <span className="rounded-full bg-indigo-100 px-1.5 py-0.5 text-[10px] font-bold text-indigo-600">
                Lv.{level}
              </span>
            </div>
            <span className="text-xs text-muted-foreground">
              @{profile.username}
            </span>
          </div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
            <span className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-teal-500" />
              经验值
            </span>
            <span className="rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-semibold text-indigo-600">
              {profile.exp.toLocaleString()}
            </span>
          </DropdownMenuItem>
          <DropdownMenuItem className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
            <span className="flex items-center gap-2">
              <Coins className="h-4 w-4 text-amber-500" />
              能量豆
            </span>
            <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-600">
              {profile.beans.toLocaleString()}
            </span>
          </DropdownMenuItem>
        </DropdownMenuGroup>

        {menuItems.map((section, sectionIdx) => (
          <div key={sectionIdx}>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              {section.group.map((item) => (
                <DropdownMenuItem
                  key={item.href}
                  onClick={() => handleNavigate(item.href)}
                  className="cursor-pointer"
                >
                  <item.icon className="h-4 w-4" />
                  {item.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuGroup>
          </div>
        ))}

        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          onClick={handleSignOut}
          className="cursor-pointer"
        >
          <LogOut className="h-4 w-4" />
          安全退出
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
