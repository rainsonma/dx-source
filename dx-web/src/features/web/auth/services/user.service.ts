import { apiClient } from "@/lib/api-client";
import type { UserProfile } from "@/features/web/auth/types/user.types";

/** Fetch the current user's profile via Go API */
export async function fetchUserProfile(): Promise<UserProfile | null> {
  try {
    const res = await apiClient.get<UserProfile>("/api/user/profile");

    if (res.code !== 0 || !res.data) {
      return null;
    }

    return res.data;
  } catch {
    return null;
  }
}
