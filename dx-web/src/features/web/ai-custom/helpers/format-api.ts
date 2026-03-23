import type { SourceType } from "@/consts/source-type";
import { getToken } from "@/lib/api-client";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type FormatResult =
  | { ok: true; formatted: string; sourceTypes: SourceType[] }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

/** Call Go API to format raw text into structured learning content */
export async function formatMetadata(
  content: string,
  formatType: "sentence" | "vocab"
): Promise<FormatResult> {
  try {
    const token = getToken();
    const res = await fetch(`${API_URL}/api/ai-custom/format-metadata`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ content, formatType }),
    });

    const json = await res.json();

    if (!res.ok || json.code !== 0) {
      return {
        ok: false,
        message: json.message ?? "格式化失败",
        code: json.code === 40007 ? "INSUFFICIENT_BEANS" : undefined,
      };
    }

    const data = json.data;

    if (data.warning) {
      return { ok: false, message: data.warning };
    }

    return { ok: true, formatted: data.formatted, sourceTypes: data.sourceTypes };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
