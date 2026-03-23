import { apiClient } from "@/lib/api-client";
import type { HeatmapData } from "@/features/web/hall/types/heatmap";

export type HeatmapActionResult = {
  data: (HeatmapData & { accountYear: number }) | null;
  error?: string;
};

/** Fetch heatmap data for a specific year (client-side year switching) */
export async function fetchHeatmapDataAction(
  year: number
): Promise<HeatmapActionResult> {
  if (!Number.isInteger(year) || year < 2000 || year > 2100) {
    return { data: null, error: "无效年份" };
  }

  try {
    const res = await apiClient.get<HeatmapData & { accountYear: number }>(
      `/api/hall/heatmap?year=${year}`
    );

    if (res.code !== 0) {
      return { data: null, error: res.message };
    }

    return { data: res.data };
  } catch {
    return { data: null, error: "加载失败，请重试" };
  }
}
