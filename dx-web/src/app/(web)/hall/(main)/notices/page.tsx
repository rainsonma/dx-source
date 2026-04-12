"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { NoticesContent } from "@/features/web/notice/components/notices-content";
import { MarkNoticesRead } from "@/features/web/notice/components/mark-notices-read";

type NoticeItem = {
  id: string;
  title: string;
  content: string | null;
  icon: string | null;
  createdAt: string;
};

export default function NoticesPage() {
  const menu = useHallMenuItem("/hall/notices");
  const [username, setUsername] = useState<string | null>(null);
  const [items, setItems] = useState<NoticeItem[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      const [profileRes, noticesRes] = await Promise.all([
        apiClient.get<{ username: string }>("/api/user/profile"),
        apiClient.get<{ items: NoticeItem[]; nextCursor: string; hasMore: boolean }>("/api/notices"),
      ]);

      if (profileRes.code === 0) setUsername(profileRes.data.username);

      if (noticesRes.code === 0) {
        setItems(noticesRes.data.items ?? []);
        setNextCursor(noticesRes.data.hasMore ? noticesRes.data.nextCursor : null);
      }
    }

    load();
  }, []);

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
      <NoticesContent
        initialItems={items}
        initialCursor={nextCursor}
        username={username}
      />
      <MarkNoticesRead />
    </div>
  );
}
