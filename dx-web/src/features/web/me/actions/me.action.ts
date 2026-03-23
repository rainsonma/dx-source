import { apiClient } from "@/lib/api-client";
import {
  updateProfileSchema,
  sendEmailCodeSchema,
  updateEmailSchema,
  changePasswordSchema,
} from "@/features/web/me/schemas/me.schema";

export type ActionResult = {
  success?: boolean;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

/** Update profile (nickname, city, introduction) */
export async function updateProfileAction(
  _prev: ActionResult,
  formData: FormData
): Promise<ActionResult> {
  const raw = {
    nickname: formData.get("nickname"),
    city: formData.get("city"),
    introduction: formData.get("introduction"),
  };
  const parsed = updateProfileSchema.safeParse(raw);
  if (!parsed.success) {
    return { fieldErrors: parsed.error.flatten().fieldErrors };
  }

  const res = await apiClient.put("/api/user/profile", {
    nickname: parsed.data.nickname || "",
    city: parsed.data.city || "",
    introduction: parsed.data.introduction || "",
  });

  if (res.code !== 0) return { error: res.message };

  return { success: true };
}

/** Update avatar after upload */
export async function updateAvatarAction(imageId: string): Promise<ActionResult> {
  if (!imageId || imageId.length !== 36) {
    return { error: "无效的图片ID" };
  }

  const res = await apiClient.put("/api/user/avatar", { image_id: imageId });

  if (res.code !== 0) return { error: res.message };

  return { success: true };
}

/** Send verification code for email change */
export async function sendEmailCodeAction(
  _prev: ActionResult,
  formData: FormData
): Promise<ActionResult> {
  const raw = { email: formData.get("email") };
  const parsed = sendEmailCodeSchema.safeParse(raw);
  if (!parsed.success) {
    return { fieldErrors: parsed.error.flatten().fieldErrors };
  }

  const res = await apiClient.post("/api/email/send-change-code", { email: parsed.data.email });

  if (res.code !== 0) return { error: res.message };

  return { success: true };
}

/** Verify code and update email */
export async function updateEmailAction(
  _prev: ActionResult,
  formData: FormData
): Promise<ActionResult> {
  const raw = {
    email: formData.get("email"),
    code: formData.get("code"),
  };
  const parsed = updateEmailSchema.safeParse(raw);
  if (!parsed.success) {
    return { fieldErrors: parsed.error.flatten().fieldErrors };
  }

  const res = await apiClient.put("/api/user/email", { email: parsed.data.email, code: parsed.data.code });

  if (res.code !== 0) return { error: res.message };

  return { success: true };
}

/** Change password */
export async function changePasswordAction(
  _prev: ActionResult,
  formData: FormData
): Promise<ActionResult> {
  const raw = {
    currentPassword: formData.get("currentPassword"),
    newPassword: formData.get("newPassword"),
    confirmPassword: formData.get("confirmPassword"),
  };
  const parsed = changePasswordSchema.safeParse(raw);
  if (!parsed.success) {
    return { fieldErrors: parsed.error.flatten().fieldErrors };
  }

  const res = await apiClient.put("/api/user/password", {
    current_password: parsed.data.currentPassword,
    new_password: parsed.data.newPassword,
  });

  if (res.code !== 0) return { error: res.message };

  return { success: true };
}
