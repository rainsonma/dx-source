"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { UnknownContent } from "@/features/web/user-unknown/components/unknown-content";
import type { UnknownItem, UnknownStats } from "@/features/web/user-unknown/actions/unknown.action";

export default function UnknownPage() {
  const [items, setItems] = useState<UnknownItem[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [stats, setStats] = useState<UnknownStats>({ total: 0, today: 0, lastThreeDays: 0 });

  useEffect(() => {
    async function load() {
      const [dataRes, statsRes] = await Promise.all([
        apiClient.get<{ items: UnknownItem[]; nextCursor: string; hasMore: boolean }>(
          "/api/tracking/unknown"
        ),
        apiClient.get<UnknownStats>("/api/tracking/unknown/stats"),
      ]);

      if (dataRes.code === 0) {
        setItems(dataRes.data.items ?? []);
        setNextCursor(dataRes.data.hasMore ? dataRes.data.nextCursor : null);
      }
      if (statsRes.code === 0) setStats(statsRes.data);
    }

    load();
  }, []);

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="生词本"
        subtitle="记录你遇到的新单词和生词"
        searchPlaceholder="搜索生词..."
      />
      <UnknownContent
        initialItems={items}
        initialCursor={nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
