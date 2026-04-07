import type { GameMode } from "@/consts/game-mode";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type GenerateResult =
  | { ok: true; generated: string; sourceType: "vocab" }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

export async function generateVocab(
  difficulty: string,
  keywords: string[],
  gameMode: GameMode
): Promise<GenerateResult> {
  try {
    const res = await fetch(`${API_URL}/api/ai-custom-vocab/generate-vocab`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ difficulty, keywords, gameMode }),
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

    return { ok: true, generated: data.generated, sourceType: "vocab" };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
