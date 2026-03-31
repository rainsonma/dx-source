import { apiClient } from "@/lib/api-client";
import type { CursorPaginated } from "@/lib/api-client";
import type { Post, Comment, CommentWithReplies, FeedTab } from "../types/post";

export const postApi = {
  async list(tab: FeedTab, cursor?: string, limit?: number) {
    const params = new URLSearchParams();
    params.set("tab", tab);
    if (cursor) params.set("cursor", cursor);
    if (limit) params.set("limit", String(limit));
    return apiClient.get<CursorPaginated<Post>>(`/api/posts?${params}`);
  },

  async detail(id: string) {
    return apiClient.get<Post>(`/api/posts/${id}`);
  },

  async create(data: { content: string; image_id?: string; tags?: string[] }) {
    return apiClient.post<Post>("/api/posts", data);
  },

  async update(id: string, data: { content: string; image_id?: string; tags?: string[] }) {
    return apiClient.put<null>(`/api/posts/${id}`, data);
  },

  async delete(id: string) {
    return apiClient.delete<null>(`/api/posts/${id}`);
  },

  async toggleLike(id: string) {
    return apiClient.post<{ liked: boolean; like_count: number }>(`/api/posts/${id}/like`);
  },

  async toggleBookmark(id: string) {
    return apiClient.post<{ bookmarked: boolean }>(`/api/posts/${id}/bookmark`);
  },

  async listComments(postId: string, cursor?: string, limit?: number) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return apiClient.get<CursorPaginated<CommentWithReplies>>(
      `/api/posts/${postId}/comments${qs ? `?${qs}` : ""}`
    );
  },

  async createComment(postId: string, data: { content: string; parent_id?: string }) {
    return apiClient.post<Comment>(`/api/posts/${postId}/comments`, data);
  },

  async updateComment(postId: string, commentId: string, content: string) {
    return apiClient.put<null>(`/api/posts/${postId}/comments/${commentId}`, { content });
  },

  async deleteComment(postId: string, commentId: string) {
    return apiClient.delete<null>(`/api/posts/${postId}/comments/${commentId}`);
  },

  async toggleFollow(userId: string) {
    return apiClient.post<{ followed: boolean }>(`/api/users/${userId}/follow`);
  },
};
