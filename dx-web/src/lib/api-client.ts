const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

// Envelope response from dx-api
interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

// Cursor-paginated response
interface CursorPaginated<T> {
  items: T[];
  nextCursor: string;
  hasMore: boolean;
}

// Offset-paginated response
interface OffsetPaginated<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

// Base fetch wrapper
async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (res.status === 401) {
    const errorData: ApiResponse<null> = await res.clone().json().catch(() => ({ code: 0, message: "", data: null }));

    if (errorData.code === 40104 && typeof window !== "undefined") {
      alert("您的账号已在其他设备登录");
      window.location.href = "/auth/signin";
    } else if (typeof window !== "undefined") {
      window.location.href = "/auth/signin";
    }
    throw new Error("Unauthorized");
  }

  return res.json();
}

// Convenience methods
export const apiClient = {
  async get<T>(path: string): Promise<ApiResponse<T>> {
    return apiFetch<T>(path, { method: "GET" });
  },

  async post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return apiFetch<T>(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  async put<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return apiFetch<T>(path, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  async delete<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return apiFetch<T>(path, {
      method: "DELETE",
      body: body ? JSON.stringify(body) : undefined,
    });
  },
};

// Auth-specific API functions targeting the Go API
export const authApi = {
  /** Send a verification code for sign-up */
  async sendSignUpCode(email: string) {
    return apiClient.post<null>("/api/email/send-signup-code", { email });
  },
  /** Complete sign-up with email, code, and optional username/password */
  async signUp(data: {
    email: string;
    code: string;
    username?: string;
    password?: string;
  }) {
    return apiClient.post<{ access_token: string; refresh_token: string; user: { id: string; name: string } }>(
      "/api/auth/signup",
      data
    );
  },
  /** Send a verification code for email sign-in */
  async sendSignInCode(email: string) {
    return apiClient.post<null>("/api/email/send-signin-code", { email });
  },
  /** Sign in with email+code or account+password */
  async signIn(data: {
    email?: string;
    code?: string;
    account?: string;
    password?: string;
  }) {
    return apiClient.post<{ access_token: string; refresh_token: string; user: { id: string; name: string } }>(
      "/api/auth/signin",
      data
    );
  },
  /** Get current user info */
  async me() {
    return apiClient.get<{ id: string; name: string }>("/api/auth/me");
  },
  /** Log out (invalidate token server-side) */
  async logout() {
    try {
      await fetch(`${API_URL}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Ignore network errors on logout
    }
    clearAccessToken();
  },
};

// User profile API functions targeting the Go API
export const userApi = {
  /** Fetch the authenticated user's full profile */
  async getProfile() {
    return apiClient.get<unknown>("/api/user/profile");
  },
  /** Update profile fields (nickname, city, introduction) */
  async updateProfile(data: { nickname?: string; city?: string; introduction?: string }) {
    return apiClient.put<unknown>("/api/user/profile", data);
  },
  /** Set avatar from an uploaded image ID */
  async updateAvatar(imageId: string) {
    return apiClient.put<unknown>("/api/user/avatar", { image_id: imageId });
  },
  /** Send verification code for email change */
  async sendEmailCode(email: string) {
    return apiClient.post<null>("/api/email/send-change-code", { email });
  },
  /** Verify code and change email */
  async changeEmail(email: string, code: string) {
    return apiClient.put<unknown>("/api/user/email", { email, code });
  },
  /** Change password (current + new) */
  async changePassword(currentPassword: string, newPassword: string) {
    return apiClient.put<unknown>("/api/user/password", {
      current_password: currentPassword,
      new_password: newPassword,
    });
  },
};

// Game API functions targeting the Go API
export const gameApi = {
  /** List published games with cursor pagination and optional filters */
  async listGames(params?: {
    cursor?: string;
    limit?: number;
    categoryIds?: string[];
    pressId?: string;
    mode?: string;
  }) {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.limit) query.set("limit", String(params.limit));
    if (params?.pressId) query.set("pressId", params.pressId);
    if (params?.mode) query.set("mode", params.mode);
    if (params?.categoryIds?.length) {
      query.set("categoryIds", params.categoryIds.join(","));
    }
    const qs = query.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/games${qs ? `?${qs}` : ""}`);
  },
  /** Search published games by name */
  async searchGames(q: string, limit?: number) {
    const params = new URLSearchParams({ q });
    if (limit) params.set("limit", String(limit));
    return apiClient.get<unknown[]>(`/api/games/search?${params}`);
  },
  /** Get current user's recently played games */
  async getRecentGames() {
    return apiClient.get<unknown[]>("/api/games/played");
  },
  /** Get full game detail with levels */
  async getGameDetail(id: string) {
    return apiClient.get<unknown>(`/api/games/${id}`);
  },
  /** Get content items for a game level, filtered by degree */
  async getLevelContent(gameId: string, levelId: string, degree?: string) {
    const params = degree ? `?degree=${degree}` : "";
    return apiClient.get<unknown[]>(`/api/games/${gameId}/levels/${levelId}/content${params}`);
  },
  /** Get hierarchical game categories */
  async getCategories() {
    return apiClient.get<unknown[]>("/api/game-categories");
  },
  /** Get game publishers */
  async getPresses() {
    return apiClient.get<unknown[]>("/api/game-presses");
  },
};

// Session response types
interface SessionStartResponse {
  id: string;
  levelId: string;
}

interface ActiveSessionResponse {
  id: string;
  degree: string;
  pattern: string | null;
  currentLevelId: string;
}

interface SessionLevelResponse {
  id: string;
  currentContentItemId: string | null;
}

interface SessionRestoreResponse {
  sessionLevel: {
    score: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
    playTime: number;
  };
}

// Session/gameplay API functions targeting the Go API
export const sessionApi = {
  /** Start or resume a game session */
  async startSession(data: {
    game_id: string;
    degree?: string;
    level_id?: string;
    pattern?: string;
  }) {
    return apiClient.post<SessionStartResponse>("/api/play-single/start", data);
  },
  /** Check for an active session by degree+pattern */
  async checkActive(gameId: string, degree: string, pattern?: string | null) {
    const params = new URLSearchParams({ game_id: gameId, degree });
    if (pattern) params.set("pattern", pattern);
    return apiClient.get<ActiveSessionResponse | null>(`/api/play-single/active?${params}`);
  },
  /** Check for any active session for a game */
  async checkAnyActive(gameId: string) {
    return apiClient.get<ActiveSessionResponse | null>(`/api/play-single/any-active?game_id=${gameId}`);
  },
  /** Check for an active level session */
  async checkActiveLevel(
    gameId: string,
    degree: string,
    pattern: string | null,
    gameLevelId: string
  ) {
    const params = new URLSearchParams({ game_id: gameId, degree, game_level_id: gameLevelId });
    if (pattern) params.set("pattern", pattern);
    return apiClient.get<unknown>(`/api/play-single/active-level?${params}`);
  },
  /** End a session */
  async endSession(
    sessionId: string,
    data: {
      game_id: string;
      score: number;
      exp: number;
      max_combo: number;
      correct_count: number;
      wrong_count: number;
      skip_count: number;
      all_levels_completed: boolean;
    }
  ) {
    return apiClient.post<unknown>(`/api/play-single/${sessionId}/end`, data);
  },
  /** Force-complete a session */
  async forceComplete(sessionId: string) {
    return apiClient.post<unknown>(`/api/play-single/${sessionId}/force-complete`);
  },
  /** Start a level within a session */
  async startLevel(
    sessionId: string,
    data: { game_level_id: string; degree: string; pattern?: string }
  ) {
    return apiClient.post<SessionLevelResponse>(`/api/play-single/${sessionId}/levels/start`, data);
  },
  /** Complete a level */
  async completeLevel(
    sessionId: string,
    levelId: string,
    data: { score: number; max_combo: number; total_items: number }
  ) {
    return apiClient.post<unknown>(
      `/api/play-single/${sessionId}/levels/${levelId}/complete`,
      data
    );
  },
  /** Advance to next level */
  async advanceLevel(sessionId: string, levelId: string, nextLevelId: string) {
    return apiClient.post<unknown>(
      `/api/play-single/${sessionId}/levels/${levelId}/advance`,
      { next_level_id: nextLevelId }
    );
  },
  /** Restart a level */
  async restartLevel(sessionId: string, levelId: string) {
    return apiClient.post<unknown>(
      `/api/play-single/${sessionId}/levels/${levelId}/restart`
    );
  },
  /** Record an answer */
  async recordAnswer(
    sessionId: string,
    data: {
      game_session_level_id: string;
      game_level_id: string;
      content_item_id: string;
      is_correct: boolean;
      user_answer: string;
      source_answer: string;
      base_score: number;
      combo_score: number;
      score: number;
      max_combo: number;
      play_time: number;
      next_content_item_id: string | null;
      duration: number;
    }
  ) {
    return apiClient.post<unknown>(`/api/play-single/${sessionId}/answers`, data);
  },
  /** Record a skip */
  async recordSkip(
    sessionId: string,
    data: {
      game_level_id: string;
      play_time: number;
      next_content_item_id: string | null;
    }
  ) {
    return apiClient.post<unknown>(`/api/play-single/${sessionId}/skips`, data);
  },
  /** Sync playtime */
  async syncPlayTime(
    sessionId: string,
    data: { game_level_id: string; play_time: number }
  ) {
    return apiClient.post<unknown>(`/api/play-single/${sessionId}/sync-playtime`, data);
  },
  /** Restore session data */
  async restore(sessionId: string, gameLevelId: string) {
    return apiClient.get<SessionRestoreResponse>(
      `/api/play-single/${sessionId}/restore?game_level_id=${gameLevelId}`
    );
  },
  /** Update current content item */
  async updateContentItem(sessionId: string, contentItemId: string | null) {
    return apiClient.put<unknown>(`/api/play-single/${sessionId}/content-item`, {
      content_item_id: contentItemId,
    });
  },
};

// Tracking API functions (mastered / unknown / review)
export const trackingApi = {
  // Mastered
  async markMastered(data: { content_item_id: string; game_id: string; game_level_id: string }) {
    return apiClient.post<unknown>("/api/tracking/master", data);
  },
  async listMastered(cursor?: string, limit?: number) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/tracking/master${qs ? `?${qs}` : ""}`);
  },
  async masterStats() {
    return apiClient.get<unknown>("/api/tracking/master/stats");
  },
  async deleteMastered(id: string) {
    return apiClient.delete<unknown>(`/api/tracking/master/${id}`);
  },
  async bulkDeleteMastered(ids: string[]) {
    return apiClient.delete<unknown>("/api/tracking/master", { ids });
  },

  // Unknown
  async markUnknown(data: { content_item_id: string; game_id: string; game_level_id: string }) {
    return apiClient.post<unknown>("/api/tracking/unknown", data);
  },
  async listUnknown(cursor?: string, limit?: number) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/tracking/unknown${qs ? `?${qs}` : ""}`);
  },
  async unknownStats() {
    return apiClient.get<unknown>("/api/tracking/unknown/stats");
  },
  async deleteUnknown(id: string) {
    return apiClient.delete<unknown>(`/api/tracking/unknown/${id}`);
  },
  async bulkDeleteUnknown(ids: string[]) {
    return apiClient.delete<unknown>("/api/tracking/unknown", { ids });
  },

  // Review
  async markReview(data: { content_item_id: string; game_id: string; game_level_id: string }) {
    return apiClient.post<unknown>("/api/tracking/review", data);
  },
  async listReviews(cursor?: string, limit?: number) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    if (limit) params.set("limit", String(limit));
    const qs = params.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/tracking/review${qs ? `?${qs}` : ""}`);
  },
  async reviewStats() {
    return apiClient.get<unknown>("/api/tracking/review/stats");
  },
  async deleteReview(id: string) {
    return apiClient.delete<unknown>(`/api/tracking/review/${id}`);
  },
  async bulkDeleteReviews(ids: string[]) {
    return apiClient.delete<unknown>("/api/tracking/review", { ids });
  },
};

// Favorites API functions
export const favoriteApi = {
  async toggle(gameId: string) {
    return apiClient.post<{ favorited: boolean }>("/api/favorites/toggle", { game_id: gameId });
  },
  async list() {
    return apiClient.get<unknown[]>("/api/favorites");
  },
};

// Leaderboard API functions
export const leaderboardApi = {
  /** Get leaderboard by type (exp|playtime) and period (all|day|week|month) */
  async getLeaderboard(type: string, period: string) {
    return apiClient.get<unknown>(`/api/leaderboard?type=${type}&period=${period}`);
  },
};

// Hall API functions
export const hallApi = {
  /** Get aggregated dashboard data */
  async getDashboard() {
    return apiClient.get<unknown>("/api/hall/dashboard");
  },
  /** Get heatmap data for a given year */
  async getHeatmap(year: number) {
    return apiClient.get<unknown>(`/api/hall/heatmap?year=${year}`);
  },
};

// Invite & Referral API functions
export const inviteApi = {
  /** Get invite code, stats, and first page of referrals */
  async getInviteData() {
    return apiClient.get<unknown>("/api/invite");
  },
  /** Get paginated referral records */
  async getReferrals(page?: number, pageSize?: number) {
    const params = new URLSearchParams();
    if (page) params.set("page", String(page));
    if (pageSize) params.set("pageSize", String(pageSize));
    const qs = params.toString();
    return apiClient.get<OffsetPaginated<unknown>>(`/api/referrals${qs ? `?${qs}` : ""}`);
  },
};

// Notice API functions
export const noticeApi = {
  /** Get active notices (cursor paginated) */
  async getNotices(cursor?: string) {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    const qs = params.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/notices${qs ? `?${qs}` : ""}`);
  },
  /** Mark all notices as read */
  async markRead() {
    return apiClient.post<null>("/api/notices/mark-read");
  },
};

// Feedback API functions
export const feedbackApi = {
  /** Submit feedback */
  async submit(data: { type: string; description: string }) {
    return apiClient.post<unknown>("/api/feedback", data);
  },
};

// Report API functions
export const reportApi = {
  /** Submit a game content report */
  async submit(data: {
    game_id: string;
    game_level_id: string;
    content_item_id: string;
    reason: string;
    note?: string;
  }) {
    return apiClient.post<unknown>("/api/reports", data);
  },
};

// Redeem API functions
export const redeemApi = {
  /** Get user's redemption records */
  async getRedeems(page?: number) {
    const params = new URLSearchParams();
    if (page) params.set("page", String(page));
    const qs = params.toString();
    return apiClient.get<OffsetPaginated<unknown>>(`/api/redeems${qs ? `?${qs}` : ""}`);
  },
  /** Redeem a code */
  async redeemCode(code: string) {
    return apiClient.post<unknown>("/api/redeems", { code });
  },
};

// Content Seek API functions
export const contentSeekApi = {
  /** Get user's content seek records */
  async getSeeks() {
    return apiClient.get<unknown[]>("/api/content-seek");
  },
  /** Submit a content seek request */
  async submit(data: { course_name: string; description: string; disk_url: string }) {
    return apiClient.post<unknown>("/api/content-seek", data);
  },
};

// Upload API functions
export const uploadApi = {
  /** Upload an image file via FormData */
  async uploadImage(file: File, role: string) {
    const token = getAccessToken();
    const formData = new FormData();
    formData.append("file", file);
    formData.append("role", role);

    const doUpload = async (authToken: string | null) => {
      return fetch(`${API_URL}/api/uploads/images`, {
        method: "POST",
        headers: authToken ? { Authorization: `Bearer ${authToken}` } : {},
        body: formData,
        credentials: "include",
      });
    };

    let res = await doUpload(token);

    if (res.status === 401) {
      const errorData = await res.clone().json().catch(() => ({ code: 0 }));
      if (errorData.code === 40104) {
        clearAccessToken();
        if (typeof window !== "undefined") {
          alert("您的账号已在其他设备登录");
          window.location.href = "/auth/signin";
        }
        throw new Error("Session replaced");
      }
      try {
        const newToken = await refreshAccessToken();
        res = await doUpload(newToken);
      } catch {
        clearAccessToken();
        if (typeof window !== "undefined") {
          window.location.href = "/auth/signin";
        }
        throw new Error("Unauthorized");
      }
    }

    if (res.status === 401) {
      clearAccessToken();
      if (typeof window !== "undefined") {
        window.location.href = "/auth/signin";
      }
      throw new Error("Unauthorized");
    }

    const data: ApiResponse<{ id: string; url: string; name: string }> = await res.json();
    return data;
  },
};

// Course game management API functions
export const courseGameApi = {
  /** List user's own games with optional status filter */
  async listGames(params?: { status?: string; cursor?: string; limit?: number }) {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.limit) query.set("limit", String(params.limit));
    const qs = query.toString();
    return apiClient.get<CursorPaginated<unknown>>(`/api/course-games${qs ? `?${qs}` : ""}`);
  },
  /** Get course game detail with levels */
  async getDetail(id: string) {
    return apiClient.get<unknown>(`/api/course-games/${id}`);
  },
  /** Create a new course game */
  async create(data: {
    name: string;
    description?: string;
    gameMode: string;
    gameCategoryId: string;
    gamePressId: string;
    coverId?: string;
  }) {
    return apiClient.post<{ id: string }>("/api/course-games", data);
  },
  /** Update a course game */
  async update(id: string, data: {
    name: string;
    description?: string;
    gameMode: string;
    gameCategoryId: string;
    gamePressId: string;
    coverId?: string;
  }) {
    return apiClient.put<null>(`/api/course-games/${id}`, data);
  },
  /** Delete a course game */
  async deleteGame(id: string) {
    return apiClient.delete<null>(`/api/course-games/${id}`);
  },
  /** Publish a course game */
  async publish(id: string) {
    return apiClient.post<null>(`/api/course-games/${id}/publish`);
  },
  /** Withdraw a published course game */
  async withdraw(id: string) {
    return apiClient.post<null>(`/api/course-games/${id}/withdraw`);
  },
  /** Add a level to a game */
  async createLevel(gameId: string, data: { name: string; description?: string }) {
    return apiClient.post<{ id: string }>(`/api/course-games/${gameId}/levels`, data);
  },
  /** Delete a level from a game */
  async deleteLevel(gameId: string, levelId: string) {
    return apiClient.delete<null>(`/api/course-games/${gameId}/levels/${levelId}`);
  },
  /** Save content metadata in batch */
  async saveMetadata(gameId: string, levelId: string, data: {
    sourceFrom: string;
    entries: Array<{ sourceData: string; translation?: string; sourceType: string }>;
  }) {
    return apiClient.post<{ count: number }>(`/api/course-games/${gameId}/levels/${levelId}/metadata`, data);
  },
  /** Reorder content metadata */
  async reorderMetadata(gameId: string, data: { metaId: string; newOrder: number }) {
    return apiClient.put<null>(`/api/course-games/${gameId}/metadata/reorder`, data);
  },
  /** Get content items grouped by metadata for a level */
  async getContentItems(gameId: string, levelId: string) {
    return apiClient.get<unknown[]>(`/api/course-games/${gameId}/levels/${levelId}/content-items`);
  },
  /** Insert a content item at a position */
  async insertContentItem(gameId: string, levelId: string, data: {
    contentMetaId: string;
    content: string;
    contentType: string;
    translation?: string | null;
    referenceItemId: string;
    direction: string;
  }) {
    return apiClient.post<unknown>(`/api/course-games/${gameId}/levels/${levelId}/content-items`, data);
  },
  /** Update content item text and translation */
  async updateContentItemText(gameId: string, itemId: string, data: {
    content: string;
    translation?: string | null;
  }) {
    return apiClient.put<null>(`/api/course-games/${gameId}/content-items/${itemId}`, data);
  },
  /** Reorder content items */
  async reorderContentItems(gameId: string, data: { itemId: string; newOrder: number }) {
    return apiClient.put<null>(`/api/course-games/${gameId}/content-items/reorder`, data);
  },
  /** Delete a single content item */
  async deleteContentItem(gameId: string, itemId: string) {
    return apiClient.delete<null>(`/api/course-games/${gameId}/content-items/${itemId}`);
  },
  /** Delete all content from a level */
  async deleteAllLevelContent(gameId: string, levelId: string) {
    return apiClient.delete<null>(`/api/course-games/${gameId}/levels/${levelId}/content-items`);
  },
};

// Order API functions
export const orderApi = {
  async createMembershipOrder(data: { grade: string; paymentMethod: string }) {
    return apiClient.post<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string;
      expiresAt: string;
      createdAt: string;
    }>("/api/orders/membership", data);
  },
  async createBeansOrder(data: { package: string; paymentMethod: string }) {
    return apiClient.post<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string;
      expiresAt: string;
      createdAt: string;
    }>("/api/orders/beans", data);
  },
  async getOrder(id: string) {
    return apiClient.get<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string | null;
      expiresAt: string;
      createdAt: string;
    }>(`/api/orders/${id}`);
  },
};

export type { ApiResponse, CursorPaginated, OffsetPaginated };
