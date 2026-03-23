import { apiClient } from "@/lib/api-client";

/** Fetch a specific page of referral records for the current user */
export async function fetchReferralPage(page: number) {
  try {
    const safePage = Math.max(1, Math.floor(page));

    const res = await apiClient.get<{
      items: any[];
      total: number;
      page: number;
      pageSize: number;
    }>(`/api/referrals?page=${safePage}&pageSize=15`);

    if (res.code !== 0) {
      return { error: res.message };
    }

    const totalPages = Math.ceil(res.data.total / res.data.pageSize);

    return { data: { referrals: res.data.items, totalPages } };
  } catch {
    return { error: "获取邀请记录失败" };
  }
}
