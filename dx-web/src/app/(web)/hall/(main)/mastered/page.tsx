"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { MasterContent } from "@/features/web/user-master/components/master-content";
import type { MasterItem, MasterStats } from "@/features/web/user-master/actions/master.action";

export default function MasterPage() {
  const [items, setItems] = useState<MasterItem[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [stats, setStats] = useState<MasterStats>({ total: 0, thisWeek: 0, thisMonth: 0 });

  useEffect(() => {
    async function load() {
      const [dataRes, statsRes] = await Promise.all([
        apiClient.get<{ items: MasterItem[]; nextCursor: string; hasMore: boolean }>(
          "/api/tracking/master"
        ),
        apiClient.get<MasterStats>("/api/tracking/master/stats"),
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
        title="已掌握"
        subtitle="你已经掌握的词汇和知识点"
        searchPlaceholder="搜索已掌握内容..."
      />
      <MasterContent
        initialItems={items}
        initialCursor={nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
