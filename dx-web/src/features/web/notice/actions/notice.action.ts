import { apiClient } from "@/lib/api-client";
import { createNoticeSchema, updateNoticeSchema } from "@/features/web/notice/schemas/notice.schema";

/** Notice item shape from Go API */
export type NoticeItem = {
  id: string;
  title: string;
  content: string | null;
  icon: string | null;
  createdAt: string;
};

/** Fetch next page of notices via Go API */
export async function fetchNoticesAction(cursor?: string): Promise<{
  items: NoticeItem[];
  nextCursor: string | null;
}> {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  const qs = params.toString();

  const res = await apiClient.get<{
    items: NoticeItem[];
    nextCursor: string;
    hasMore: boolean;
  }>(`/api/notices${qs ? `?${qs}` : ""}`);

  if (res.code !== 0) {
    return { items: [], nextCursor: null };
  }

  return {
    items: res.data.items ?? [],
    nextCursor: res.data.hasMore ? res.data.nextCursor : null,
  };
}

/** Create a new notice (admin only) via Go API */
export async function createNoticeAction(
  input: { title: string; content?: string; icon?: string }
): Promise<{ data: NoticeItem } | { error: string }> {
  try {
    const parsed = createNoticeSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0].message };
    }

    const res = await apiClient.post<NoticeItem>("/api/admin/notices", parsed.data);

    if (res.code !== 0) {
      return { error: res.message || "创建失败" };
    }

    return { data: res.data };
  } catch {
    return { error: "创建失败" };
  }
}

/** Update an existing notice (admin only) via Go API */
export async function updateNoticeAction(
  input: { id: string; title: string; content?: string; icon?: string }
): Promise<{ data: NoticeItem } | { error: string }> {
  try {
    const parsed = updateNoticeSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0].message };
    }

    const { id, ...data } = parsed.data;
    const res = await apiClient.put<NoticeItem>(`/api/admin/notices/${id}`, data);

    if (res.code !== 0) {
      return { error: res.message || "更新失败" };
    }

    return { data: res.data };
  } catch {
    return { error: "更新失败" };
  }
}

/** Delete a notice (admin only) via Go API */
export async function deleteNoticeAction(
  id: string
): Promise<{ success: true } | { error: string }> {
  try {
    const res = await apiClient.delete(`/api/admin/notices/${id}`);

    if (res.code !== 0) {
      return { error: res.message || "删除失败" };
    }

    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Mark notices as read for the current user via Go API */
export async function markNoticesReadAction(): Promise<void> {
  await apiClient.post("/api/notices/mark-read");
}
