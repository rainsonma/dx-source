import { getToken } from "@/lib/api-client";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

export type ProgressEvent = {
  done: number;
  total: number;
  status: "ok" | "failed";
  processed?: number;
  failed?: number;
  complete?: boolean;
};

type StreamResult =
  | { ok: true; processed: number; failed: number }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

/** Fetch a Go API SSE endpoint with progress reporting */
export async function fetchWithProgress(
  path: string,
  body: object,
  signal: AbortSignal | undefined,
  onProgress: (event: ProgressEvent) => void
): Promise<StreamResult> {
  try {
    const token = getToken();
    const res = await fetch(`${API_URL}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(body),
      signal,
    });

    if (!res.ok) {
      // Try to read error from envelope response
      try {
        const json = await res.json();
        return {
          ok: false,
          message: json.message ?? "请求失败",
          code: json.code === 40007 ? "INSUFFICIENT_BEANS" : undefined,
        };
      } catch {
        return { ok: false, message: "请求失败" };
      }
    }

    const contentType = res.headers.get("content-type") ?? "";

    // If server returned JSON (e.g. empty metas case via envelope), handle directly
    if (contentType.includes("application/json")) {
      const json = await res.json();
      const data = json.data ?? json;
      return { ok: true, processed: data.processed ?? 0, failed: data.failed ?? 0 };
    }

    // Read SSE stream
    const reader = res.body!.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let finalResult: StreamResult = { ok: false, message: "流中断" };

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      const lines = buffer.split("\n");
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        // Handle error events
        if (line.startsWith("event: error")) {
          // Next data line is the error message
          continue;
        }
        if (line.startsWith("event: done")) {
          continue;
        }
        if (!line.startsWith("data: ")) continue;
        const json = line.slice(6);
        if (!json) continue;

        try {
          const event = JSON.parse(json) as ProgressEvent;
          onProgress(event);

          if (event.complete) {
            finalResult = {
              ok: true,
              processed: event.processed ?? 0,
              failed: event.failed ?? 0,
            };
          }
        } catch {
          // Error event data is plain text, not JSON
          finalResult = { ok: false, message: json };
        }
      }
    }

    return finalResult;
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      return { ok: false, message: "已取消" };
    }
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
