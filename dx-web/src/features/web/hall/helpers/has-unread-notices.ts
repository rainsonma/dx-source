import { apiClient } from "@/lib/api-client";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";

type NoticeItem = {
  id: string;
  createdAt: string;
};

/** Check if the current user has unread notices */
export async function hasUnreadNotices(): Promise<boolean> {
  const [profile, noticeRes] = await Promise.all([
    fetchUserProfile(),
    apiClient.get<{ items: NoticeItem[]; nextCursor: string; hasMore: boolean }>(
      "/api/notices?limit=1"
    ),
  ]);

  if (!profile) return false;
  if (noticeRes.code !== 0) return false;

  const latestNotice = noticeRes.data.items?.[0];
  if (!latestNotice) return false;

  if (!profile.lastReadNoticeAt) return true;

  return new Date(profile.lastReadNoticeAt) < new Date(latestNotice.createdAt);
}
