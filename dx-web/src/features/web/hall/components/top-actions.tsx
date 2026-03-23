"use client";

import Link from "next/link";
import { Bell } from "lucide-react";
import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { UserProfileMenu } from "@/features/web/auth/components/user-profile-menu";
import { ContentSeekButton } from "@/features/web/hall/components/content-seek-button";
import { FeedbackButton } from "@/features/web/hall/components/feedback-button";
import { GameSearchTrigger } from "@/features/web/hall/components/game-search-trigger";
import { ThemeToggleButton } from "@/features/web/hall/components/theme-toggle-button";
import type { UserProfile } from "@/features/web/auth/types/user.types";

export function TopActions({
  searchPlaceholder = "搜索课程游戏...",
}: {
  searchPlaceholder?: string;
} = {}) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [unread, setUnread] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        const [profileRes, noticeRes] = await Promise.all([
          apiClient.get<UserProfile>("/api/user/profile"),
          apiClient.get<{ items: { id: string; createdAt: string }[]; nextCursor: string; hasMore: boolean }>(
            "/api/notices?limit=1"
          ),
        ]);

        if (profileRes.code === 0 && profileRes.data) {
          setProfile(profileRes.data);

          if (noticeRes.code === 0) {
            const latestNotice = noticeRes.data.items?.[0];
            if (latestNotice) {
              if (!profileRes.data.lastReadNoticeAt) {
                setUnread(true);
              } else {
                setUnread(
                  new Date(profileRes.data.lastReadNoticeAt) < new Date(latestNotice.createdAt)
                );
              }
            }
          }
        }
      } catch {
        // Silently ignore errors — user may not be authenticated yet
      }
    }

    loadData();
  }, []);

  return (
    <div className="flex items-center gap-3">
      {/* Search — hidden below 1024px */}
      <div className="hidden lg:flex">
        <GameSearchTrigger placeholder={searchPlaceholder} />
      </div>

      {/* Request course — hidden below 1280px */}
      <div className="hidden xl:flex">
        <ContentSeekButton />
      </div>

      {/* Icon buttons — hidden below 1280px */}
      <div className="hidden xl:flex items-center gap-3">
        <FeedbackButton />
        <Link
          href="/hall/notices"
          className="relative flex h-10 w-10 items-center justify-center rounded-[10px] border border-border bg-card text-muted-foreground hover:bg-accent"
        >
          <Bell className="h-[18px] w-[18px]" />
          {unread && (
            <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-red-500" />
          )}
        </Link>
        <ThemeToggleButton />
      </div>

      {/* Profile */}
      {profile && <UserProfileMenu profile={profile} />}
    </div>
  );
}
