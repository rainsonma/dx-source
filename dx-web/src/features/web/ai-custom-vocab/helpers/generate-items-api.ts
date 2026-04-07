import { fetchWithProgress, type ProgressEvent } from "@/features/web/ai-custom-vocab/helpers/stream-progress";

type BatchResult =
  | { ok: true; processed: number; failed: number }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

type OnProgress = (event: ProgressEvent) => void;

export async function breakVocabMetadata(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom-vocab/break-metadata",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}

export async function generateVocabContentItems(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom-vocab/generate-content-items",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}
