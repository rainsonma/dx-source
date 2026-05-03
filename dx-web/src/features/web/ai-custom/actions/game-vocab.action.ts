'use server';

import { apiClient } from '@/lib/api-client';
import type { AddedGameVocab, LevelVocabData } from '@/lib/api-client';

export async function listGameVocabsAction(gameId: string, levelId: string) {
  return apiClient.get<LevelVocabData[]>(
    `/api/course-games/${gameId}/levels/${levelId}/game-vocabs`
  );
}

export async function addGameVocabsAction(gameId: string, levelId: string, entries: string[]) {
  return apiClient.post<AddedGameVocab[]>(
    `/api/course-games/${gameId}/levels/${levelId}/game-vocabs`,
    { entries }
  );
}

export async function reorderGameVocabAction(gameId: string, gvId: string, newOrder: number) {
  return apiClient.put<void>(
    `/api/course-games/${gameId}/game-vocabs/${gvId}/reorder`,
    { newOrder }
  );
}

export async function deleteGameVocabAction(gameId: string, gvId: string) {
  return apiClient.delete<void>(`/api/course-games/${gameId}/game-vocabs/${gvId}`);
}
