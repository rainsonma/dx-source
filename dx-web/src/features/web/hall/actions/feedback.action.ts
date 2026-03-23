import { apiClient } from "@/lib/api-client";
import { feedbackSchema } from "@/features/web/hall/schemas/feedback.schema";

/** Submit feedback for the current user */
export async function submitFeedbackAction(input: {
  type: string;
  description: string;
}) {
  try {
    const parsed = feedbackSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const res = await apiClient.post<{ duplicate: boolean }>("/api/feedback", parsed.data);

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
