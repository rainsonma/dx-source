import type { SourceType } from "@/consts/source-type";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type GenerateResult =
  | { ok: true; generated: string; sourceType: SourceType }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

/** Call Go API to generate a story from keywords using AI */
export async function generateStory(
  difficulty: string,
  keywords: string[]
): Promise<GenerateResult> {
  try {
    const res = await fetch(`${API_URL}/api/ai-custom/generate-metadata`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ difficulty, keywords }),
    });

    const json = await res.json();

    if (!res.ok || json.code !== 0) {
      return {
        ok: false,
        message: json.message ?? "生成失败",
        code: json.code === 40007 ? "INSUFFICIENT_BEANS" : undefined,
      };
    }

    const data = json.data;

    if (data.warning) {
      return { ok: false, message: data.warning };
    }

    return { ok: true, generated: data.generated, sourceType: data.sourceType };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
