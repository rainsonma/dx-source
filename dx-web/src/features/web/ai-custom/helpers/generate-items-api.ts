import { fetchWithProgress, type ProgressEvent } from "@/features/web/ai-custom/helpers/stream-progress";

type BatchResult =
  | { ok: true; processed: number; failed: number }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

type OnProgress = (event: ProgressEvent) => void;

/** Call Go API to break content metas into learning units via SSE */
export async function breakMetadata(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom/break-metadata",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}

/** Call Go API to generate word-level details via SSE */
export async function generateContentItems(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom/generate-content-items",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}
