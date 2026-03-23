import { apiClient } from "@/lib/api-client";
import { contentSeekSchema } from "@/features/web/hall/schemas/content-seek.schema";

/** Submit a course request for the current user */
export async function submitContentSeekAction(input: {
  courseName: string;
  description: string;
  diskUrl: string;
}) {
  try {
    const parsed = contentSeekSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const res = await apiClient.post<{ duplicate: boolean }>("/api/content-seek", {
      course_name: parsed.data.courseName,
      description: parsed.data.description,
      disk_url: parsed.data.diskUrl,
    });

    if (res.code !== 0) {
      return { error: res.message };
    }

    return res.data.duplicate
      ? { duplicate: true as const }
      : { success: true as const };
  } catch {
    return { error: "提交失败" };
  }
}
