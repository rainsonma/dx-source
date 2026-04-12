"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { ReviewContent } from "@/features/web/user-review/components/review-content";
import type { ReviewItem, ReviewStats } from "@/features/web/user-review/actions/review.action";

export default function ReviewPage() {
  const menu = useHallMenuItem("/hall/review");
  const [items, setItems] = useState<ReviewItem[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [stats, setStats] = useState<ReviewStats>({ pending: 0, overdue: 0, reviewedToday: 0 });

  useEffect(() => {
    async function load() {
      const [dataRes, statsRes] = await Promise.all([
        apiClient.get<{ items: ReviewItem[]; nextCursor: string; hasMore: boolean }>(
          "/api/tracking/review"
        ),
        apiClient.get<ReviewStats>("/api/tracking/review/stats"),
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
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索复习内容..."
      />
      <ReviewContent
        initialItems={items}
        initialCursor={nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
