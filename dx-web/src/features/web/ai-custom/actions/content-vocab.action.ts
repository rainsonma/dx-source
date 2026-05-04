import { apiClient } from '@/lib/api-client';
import type {
  ContentVocabData,
  VocabInput,
  CreateVocabResult,
} from '@/lib/api-client';

export async function listMyVocabsAction(params?: { cursor?: string; search?: string; limit?: number }) {
  const query = new URLSearchParams();
  if (params?.cursor) query.set("cursor", params.cursor);
  if (params?.search) query.set("search", params.search);
  if (params?.limit) query.set("limit", String(params.limit));
  const qs = query.toString();
  return apiClient.get<{ items: ContentVocabData[]; nextCursor: string; hasMore: boolean }>(
    `/api/content-vocabs/mine${qs ? `?${qs}` : ""}`
  );
}

export async function createVocabAction(input: VocabInput) {
  return apiClient.post<CreateVocabResult>('/api/content-vocabs', input);
}

export async function updateVocabAction(id: string, input: VocabInput) {
  return apiClient.put<ContentVocabData>(`/api/content-vocabs/${id}`, input);
}

export async function deleteVocabAction(id: string) {
  return apiClient.delete<void>(`/api/content-vocabs/${id}`);
}

export async function generateVocabWordsAction(keywords: string[], difficulty: string) {
  return apiClient.post<{ words: string[] }>('/api/ai-custom/generate-vocab-words', { keywords, difficulty });
}

export async function createVocabsFromWordsAction(words: string[]) {
  return apiClient.post<CreateVocabResult[]>('/api/content-vocabs/from-words', { words });
}
