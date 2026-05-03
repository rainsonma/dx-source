'use server';

import { apiClient } from '@/lib/api-client';
import type {
  ContentVocabComplementPatch,
  ContentVocabData,
  ContentVocabReplacePatch,
} from '@/lib/api-client';

export async function getVocabByContentAction(content: string) {
  return apiClient.get<ContentVocabData | null>(
    `/api/content-vocabs?content=${encodeURIComponent(content)}`
  );
}

export async function complementVocabAction(id: string, patch: ContentVocabComplementPatch) {
  return apiClient.post<ContentVocabData>(`/api/content-vocabs/${id}/complement`, patch);
}

export async function replaceVocabAction(id: string, patch: ContentVocabReplacePatch) {
  return apiClient.put<ContentVocabData>(`/api/content-vocabs/${id}`, patch);
}

export async function verifyVocabAction(id: string, verified: boolean) {
  return apiClient.post<ContentVocabData>(`/api/content-vocabs/${id}/verify`, { verified });
}
