const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type FormatResult =
  | { ok: true; formatted: string }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

// formatVocabWords cleans a paste into pure English words/phrases (one per
// line). For the AI 词汇库 manual tab — strips Chinese, drops translations,
// rejects sentence-shaped input.
export async function formatVocabWords(content: string): Promise<FormatResult> {
  try {
    const res = await fetch(`${API_URL}/api/ai-custom/format-vocab-words`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ content }),
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

    return { ok: true, formatted: data.formatted };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
