import { apiClient } from "@/lib/api-client";
import { generateCodesSchema, redeemCodeSchema } from "@/features/web/redeem/schemas/redeem.schema";

/** Generate redeem codes (admin only) via Go API */
export async function generateCodesAction(input: { grade: string; quantity: string }) {
  try {
    const parsed = generateCodesSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const res = await apiClient.post<{ count: number }>("/api/admin/redeems/generate", {
      grade: parsed.data.grade,
      count: parsed.data.quantity,
    });

    if (res.code !== 0) {
      return { error: res.message || "生成兑换码失败" };
    }

    return { data: res.data };
  } catch {
    return { error: "生成兑换码失败" };
  }
}

/** Redeem a code for the current user via Go API */
export async function redeemCodeAction(input: { code: string }) {
  try {
    const parsed = redeemCodeSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "兑换码格式不正确" };
    }

    const res = await apiClient.post<{ grade: string }>("/api/redeems", { code: parsed.data.code });

    if (res.code !== 0) {
      return { error: res.message };
    }

    return { success: true as const, grade: res.data.grade };
  } catch {
    return { error: "兑换失败" };
  }
}

