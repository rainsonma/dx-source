# Hall Me (个人中心) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a personal center page at `/hall/me` with hero banner, 6 info blocks, and edit modals for profile/email/password.

**Architecture:** Server component page fetches full user profile, passes data to client block components. Each editable block opens a Dialog modal that submits via Server Actions. Email change reuses existing Redis verification code pattern. Avatar upload reuses existing Uppy hook.

**Tech Stack:** Next.js 16 App Router, Prisma, Zod, Server Actions, shadcn/ui Dialog, Uppy, Redis, bcryptjs

---

### Task 1: Sidebar — Link 个人中心 to /hall/me

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:58`
- Modify: `src/features/web/auth/components/user-profile-menu.tsx:50`

**Step 1: Update sidebar nav href**

In `hall-sidebar.tsx`, change line 58:
```typescript
// Before
{ icon: Medal, label: "个人中心", href: "/hall" },
// After
{ icon: Medal, label: "个人中心", href: "/hall/me" },
```

**Step 2: Update profile menu href**

In `user-profile-menu.tsx`, update the 个人中心 menu item:
```typescript
// Before
{ label: "个人中心", icon: User, href: "/hall" },
// After
{ label: "个人中心", icon: User, href: "/hall/me" },
```

**Step 3: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx src/features/web/auth/components/user-profile-menu.tsx
git commit -m "feat: link 个人中心 to /hall/me in sidebar and profile menu"
```

---

### Task 2: Model Layer — Full Profile Query

**Files:**
- Modify: `src/models/user/user.query.ts`

**Step 1: Add getUserFullProfile function**

Append to `user.query.ts`:
```typescript
/** Get full user profile for the /hall/me personal center page */
export async function getUserFullProfile(userId: string) {
  const user = await db.user.findUnique({
    where: { id: userId },
    select: {
      id: true,
      username: true,
      nickname: true,
      email: true,
      phone: true,
      city: true,
      introduction: true,
      grade: true,
      vipDueAt: true,
      beans: true,
      exp: true,
      currentPlayStreak: true,
      maxPlayStreak: true,
      lastPlayedAt: true,
      inviteCode: true,
      createdAt: true,
      avatar: {
        select: { url: true },
      },
    },
  });

  if (!user) return null;

  return {
    ...user,
    grade: user.grade as UserGrade,
    avatarUrl: user.avatar?.url ?? null,
  };
}
```

**Step 2: Verify build**

```bash
npm run build
```

**Step 3: Commit**

```bash
git add src/models/user/user.query.ts
git commit -m "feat: add getUserFullProfile query for personal center"
```

---

### Task 3: Model Layer — Profile Mutations

**Files:**
- Modify: `src/models/user/user.mutation.ts`

**Step 1: Add updateUserProfile mutation**

```typescript
/** Update user profile fields (nickname, city, introduction) */
export async function updateUserProfile(
  userId: string,
  data: { nickname?: string | null; city?: string | null; introduction?: string | null }
) {
  return db.user.update({
    where: { id: userId },
    data,
    select: { id: true },
  });
}
```

**Step 2: Add updateUserAvatar mutation**

```typescript
/** Update user avatar by linking to an uploaded image */
export async function updateUserAvatar(userId: string, avatarId: string) {
  return db.user.update({
    where: { id: userId },
    data: { avatarId },
    select: { id: true },
  });
}
```

**Step 3: Add updateUserEmail mutation**

```typescript
/** Update user email address */
export async function updateUserEmail(userId: string, email: string) {
  return db.user.update({
    where: { id: userId },
    data: { email },
    select: { id: true },
  });
}
```

**Step 4: Add updateUserPassword mutation**

```typescript
/** Update user password (stores pre-hashed value) */
export async function updateUserPassword(userId: string, hashedPassword: string) {
  return db.user.update({
    where: { id: userId },
    data: { password: hashedPassword },
    select: { id: true },
  });
}
```

**Step 5: Verify build**

```bash
npm run build
```

**Step 6: Commit**

```bash
git add src/models/user/user.mutation.ts
git commit -m "feat: add profile/avatar/email/password mutations for personal center"
```

---

### Task 4: Types

**Files:**
- Create: `src/features/web/me/types/me.types.ts`

**Step 1: Create MeProfile type**

```typescript
import type { UserGrade } from "@/consts/user-grade";

/** Full user profile data for the personal center page */
export type MeProfile = {
  id: string;
  username: string;
  nickname: string | null;
  email: string | null;
  phone: string | null;
  city: string | null;
  introduction: string | null;
  grade: UserGrade;
  vipDueAt: Date | null;
  beans: number;
  exp: number;
  currentPlayStreak: number;
  maxPlayStreak: number;
  lastPlayedAt: Date | null;
  inviteCode: string;
  createdAt: Date;
  avatarUrl: string | null;
};
```

**Step 2: Commit**

```bash
git add src/features/web/me/types/me.types.ts
git commit -m "feat: add MeProfile type"
```

---

### Task 5: Zod Schemas

**Files:**
- Create: `src/features/web/me/schemas/me.schema.ts`

**Step 1: Create all validation schemas**

```typescript
import { z } from "zod";

/** Profile edit form schema (nickname, city, introduction) */
export const updateProfileSchema = z.object({
  nickname: z.string().max(30, "昵称最长30个字符").optional().or(z.literal("")),
  city: z.string().max(50, "城市最长50个字符").optional().or(z.literal("")),
  introduction: z.string().max(200, "简介最长200个字符").optional().or(z.literal("")),
});

/** Email change - send verification code */
export const sendEmailCodeSchema = z.object({
  email: z.string().min(1, "请输入邮箱").email("邮箱格式不正确"),
});

/** Email change - verify code and update */
export const updateEmailSchema = z.object({
  email: z.string().min(1, "请输入邮箱").email("邮箱格式不正确"),
  code: z.string().length(6, "验证码为6位数字").regex(/^\d{6}$/, "验证码为6位数字"),
});

/** Password change form schema */
export const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, "请输入当前密码"),
    newPassword: z
      .string()
      .min(8, "密码至少8个字符")
      .regex(/[a-z]/, "密码需包含小写字母")
      .regex(/[A-Z]/, "密码需包含大写字母")
      .regex(/\d/, "密码需包含数字"),
    confirmPassword: z.string().min(1, "请确认新密码"),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "两次输入的密码不一致",
    path: ["confirmPassword"],
  });

export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;
export type SendEmailCodeInput = z.infer<typeof sendEmailCodeSchema>;
export type UpdateEmailInput = z.infer<typeof updateEmailSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;
```

**Step 2: Commit**

```bash
git add src/features/web/me/schemas/me.schema.ts
git commit -m "feat: add Zod schemas for personal center forms"
```

---

### Task 6: Service Layer

**Files:**
- Create: `src/features/web/me/services/me.service.ts`

**Step 1: Create service with all business logic**

```typescript
import "server-only";

import bcrypt from "bcryptjs";

import { redis } from "@/lib/redis";
import { emailQueue } from "@/lib/queue";
import { generateVerificationCode } from "@/features/web/auth/helpers/code.helper";
import { getUserFullProfile } from "@/models/user/user.query";
import {
  updateUserProfile,
  updateUserAvatar,
  updateUserEmail,
  updateUserPassword,
} from "@/models/user/user.mutation";
import { findUserByEmail, findUserByAccount } from "@/models/user/user.query";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import type { UpdateProfileInput } from "@/features/web/me/schemas/me.schema";

const CODE_TTL = 300;
const RATE_TTL = 60;

function codeKey(email: string) {
  return `change-email:code:${email}`;
}

function rateKey(email: string) {
  return `change-email:rate:${email}`;
}

/** Fetch the full user profile for the personal center page */
export async function fetchMeProfile() {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  return getUserFullProfile(profile.id);
}

/** Update user profile fields (nickname, city, introduction) */
export async function editProfile(data: UpdateProfileInput) {
  const profile = await fetchUserProfile();
  if (!profile) return { error: "请先登录" };

  await updateUserProfile(profile.id, {
    nickname: data.nickname || null,
    city: data.city || null,
    introduction: data.introduction || null,
  });

  return { success: true };
}

/** Update user avatar after upload */
export async function editAvatar(imageId: string) {
  const profile = await fetchUserProfile();
  if (!profile) return { error: "请先登录" };

  await updateUserAvatar(profile.id, imageId);

  return { success: true };
}

/** Send verification code for email change */
export async function sendChangeEmailCode(email: string) {
  const profile = await fetchUserProfile();
  if (!profile) return { error: "请先登录" };

  const rateLimited = await redis.exists(rateKey(email));
  if (rateLimited) {
    return { error: "请稍后再试，每分钟只能发送一次" };
  }

  const existing = await findUserByEmail(email);
  if (existing && existing.id !== profile.id) {
    return { error: "该邮箱已被其他账号使用" };
  }

  const code = generateVerificationCode();
  await redis.setex(codeKey(email), CODE_TTL, code);
  await redis.setex(rateKey(email), RATE_TTL, "1");

  await emailQueue.add("send-verification-code", {
    to: email,
    subject: "斗学 修改邮箱验证码",
    html: `
      <div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
        <h2 style="color: #0d9488;">斗学 修改邮箱验证码</h2>
        <p>您的验证码是：</p>
        <div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #0d9488; padding: 16px 0;">
          ${code}
        </div>
        <p style="color: #94a3b8; font-size: 14px;">验证码有效期为5分钟，请尽快使用。</p>
      </div>
    `,
  });

  return { success: true };
}

/** Verify code and update email */
export async function editEmail(email: string, code: string) {
  const profile = await fetchUserProfile();
  if (!profile) return { error: "请先登录" };

  const storedCode = await redis.get(codeKey(email));
  if (!storedCode || storedCode !== code) {
    return { error: "验证码不正确或已过期" };
  }

  await redis.del(codeKey(email));

  const existing = await findUserByEmail(email);
  if (existing && existing.id !== profile.id) {
    return { error: "该邮箱已被其他账号使用" };
  }

  await updateUserEmail(profile.id, email);

  return { success: true };
}

/** Change user password (verify current, hash new) */
export async function editPassword(currentPassword: string, newPassword: string) {
  const profile = await fetchUserProfile();
  if (!profile) return { error: "请先登录" };

  const user = await findUserByAccount(profile.username);
  if (!user) return { error: "用户不存在" };

  const valid = await bcrypt.compare(currentPassword, user.password);
  if (!valid) {
    return { error: "当前密码不正确" };
  }

  const hashedPassword = await bcrypt.hash(newPassword, 12);
  await updateUserPassword(profile.id, hashedPassword);

  return { success: true };
}
```

**Step 2: Verify build**

```bash
npm run build
```

**Step 3: Commit**

```bash
git add src/features/web/me/services/me.service.ts
git commit -m "feat: add me service with profile/email/password/avatar logic"
```

---

### Task 7: Server Actions

**Files:**
- Create: `src/features/web/me/actions/me.action.ts`

**Step 1: Create all server actions**

```typescript
"use server";

import { revalidatePath } from "next/cache";

import {
  updateProfileSchema,
  sendEmailCodeSchema,
  updateEmailSchema,
  changePasswordSchema,
} from "@/features/web/me/schemas/me.schema";
import {
  editProfile,
  editAvatar,
  sendChangeEmailCode,
  editEmail,
  editPassword,
} from "@/features/web/me/services/me.service";

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

  const result = await editProfile(parsed.data);
  if (result.error) return { error: result.error };

  revalidatePath("/hall/me");
  return { success: true };
}

/** Update avatar after upload */
export async function updateAvatarAction(imageId: string): Promise<ActionResult> {
  const result = await editAvatar(imageId);
  if (result.error) return { error: result.error };

  revalidatePath("/hall/me");
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

  const result = await sendChangeEmailCode(parsed.data.email);
  if (result.error) return { error: result.error };

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

  const result = await editEmail(parsed.data.email, parsed.data.code);
  if (result.error) return { error: result.error };

  revalidatePath("/hall/me");
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

  const result = await editPassword(parsed.data.currentPassword, parsed.data.newPassword);
  if (result.error) return { error: result.error };

  revalidatePath("/hall/me");
  return { success: true };
}
```

**Step 2: Verify build**

```bash
npm run build
```

**Step 3: Commit**

```bash
git add src/features/web/me/actions/me.action.ts
git commit -m "feat: add server actions for profile/email/password/avatar updates"
```

---

### Task 8: Page Route + Hero Component

**Files:**
- Create: `src/app/(web)/hall/(main)/me/page.tsx`
- Create: `src/features/web/me/components/me-hero.tsx`
- Create: `src/features/web/me/components/avatar-uploader.tsx`

**Step 1: Create avatar-uploader.tsx**

```typescript
"use client";

import { useRef } from "react";
import { Camera } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useImageUploader } from "@/features/com/images/hooks/use-image-uploader";
import { IMAGE_ROLES } from "@/consts/image-role";
import { updateAvatarAction } from "@/features/web/me/actions/me.action";

const avatarColors = [
  "#ef4444", "#f97316", "#f59e0b", "#eab308", "#84cc16",
  "#22c55e", "#14b8a6", "#06b6d4", "#0ea5e9", "#3b82f6",
  "#6366f1", "#8b5cf6", "#a855f7", "#d946ef", "#ec4899",
];

function getAvatarColor(id: string) {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0;
  }
  return avatarColors[Math.abs(hash) % avatarColors.length];
}

interface AvatarUploaderProps {
  userId: string;
  avatarUrl: string | null;
  displayName: string;
}

/** Clickable avatar with Uppy upload overlay */
export function AvatarUploader({ userId, avatarUrl, displayName }: AvatarUploaderProps) {
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const avatarBg = getAvatarColor(userId);

  const { upload, isUploading, inputRef, handleFileChange } = useImageUploader({
    role: IMAGE_ROLES.USER_AVATAR,
    onUploadComplete: async (image) => {
      await updateAvatarAction(image.id);
    },
  });

  return (
    <div className="relative cursor-pointer" onClick={upload}>
      <Avatar className="h-20 w-20">
        {avatarUrl && <AvatarImage src={avatarUrl} alt={displayName} />}
        <AvatarFallback
          className="text-2xl font-bold"
          style={{ backgroundColor: avatarBg, color: "#fff" }}
        >
          {fallbackChar}
        </AvatarFallback>
      </Avatar>

      <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition-opacity hover:opacity-100">
        <Camera className="h-6 w-6 text-white" />
      </div>

      {isUploading && (
        <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/50">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent" />
        </div>
      )}

      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png"
        className="hidden"
        onChange={handleFileChange}
      />
    </div>
  );
}
```

**Step 2: Create me-hero.tsx**

```typescript
"use client";

import { Zap, Flame, Coins } from "lucide-react";

import { getLevel } from "@/consts/user-level";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import { AvatarUploader } from "@/features/web/me/components/avatar-uploader";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Hero banner with avatar, display name, level, grade, and key stats */
export function MeHero({ profile }: { profile: MeProfile }) {
  const displayName = profile.nickname ?? profile.username;
  const level = getLevel(profile.exp);
  const gradeLabel = USER_GRADE_LABELS[profile.grade];

  return (
    <div className="flex flex-col items-center gap-4 rounded-2xl border border-slate-200 bg-white p-8 md:flex-row md:items-center md:gap-6">
      <AvatarUploader
        userId={profile.id}
        avatarUrl={profile.avatarUrl}
        displayName={displayName}
      />

      <div className="flex flex-1 flex-col items-center gap-2 md:items-start">
        <div className="flex items-center gap-2">
          <h2 className="text-xl font-bold text-slate-900">{displayName}</h2>
          <span className="rounded-full bg-indigo-100 px-2 py-0.5 text-[11px] font-bold text-indigo-600">
            Lv.{level}
          </span>
          <span className="rounded bg-slate-200 px-1.5 py-0.5 text-[11px] font-semibold text-slate-500">
            {gradeLabel}
          </span>
        </div>
        <p className="text-sm text-slate-400">@{profile.username}</p>

        <div className="mt-2 flex gap-6">
          <div className="flex items-center gap-1.5">
            <Zap className="h-4 w-4 text-teal-600" />
            <span className="text-sm font-semibold text-slate-700">{profile.exp.toLocaleString()}</span>
            <span className="text-xs text-slate-400">经验</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Flame className="h-4 w-4 text-orange-500" />
            <span className="text-sm font-semibold text-slate-700">{profile.currentPlayStreak}</span>
            <span className="text-xs text-slate-400">天连续</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Coins className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-semibold text-slate-700">{profile.beans.toLocaleString()}</span>
            <span className="text-xs text-slate-400">能量豆</span>
          </div>
        </div>
      </div>
    </div>
  );
}
```

**Step 3: Create page.tsx**

```typescript
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { MeHero } from "@/features/web/me/components/me-hero";
import { ProfileBlock } from "@/features/web/me/components/profile-block";
import { AccountBlock } from "@/features/web/me/components/account-block";
import { SecurityBlock } from "@/features/web/me/components/security-block";
import { MembershipBlock } from "@/features/web/me/components/membership-block";
import { StatsBlock } from "@/features/web/me/components/stats-block";
import { InviteBlock } from "@/features/web/me/components/invite-block";
import { fetchMeProfile } from "@/features/web/me/services/me.service";

export default async function MePage() {
  const profile = await fetchMeProfile();
  if (!profile) return null;

  return (
    <div className="flex min-h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="个人中心"
        subtitle="管理你的个人资料和账号信息"
      />
      <MeHero profile={profile} />
      <div className="flex flex-col gap-5">
        <ProfileBlock profile={profile} />
        <AccountBlock profile={profile} />
        <SecurityBlock />
        <MembershipBlock profile={profile} />
        <StatsBlock profile={profile} />
        <InviteBlock inviteCode={profile.inviteCode} />
      </div>
    </div>
  );
}
```

**Step 4: Verify build**

```bash
npm run build
```
Expected: May fail because block components don't exist yet. That's OK — continue to next tasks.

**Step 5: Commit**

```bash
git add src/app/(web)/hall/(main)/me/page.tsx src/features/web/me/components/me-hero.tsx src/features/web/me/components/avatar-uploader.tsx
git commit -m "feat: add /hall/me page route with hero banner and avatar upload"
```

---

### Task 9: Profile Block + Edit Dialog

**Files:**
- Create: `src/features/web/me/components/profile-block.tsx`
- Create: `src/features/web/me/components/edit-profile-dialog.tsx`

**Step 1: Create edit-profile-dialog.tsx**

```typescript
"use client";

import { useActionState, useEffect } from "react";
import { Loader2 } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { updateProfileAction } from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface EditProfileDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nickname: string | null;
  city: string | null;
  introduction: string | null;
}

/** Dialog for editing profile fields (nickname, city, introduction) */
export function EditProfileDialog({
  open,
  onOpenChange,
  nickname,
  city,
  introduction,
}: EditProfileDialogProps) {
  const [state, formAction, pending] = useActionState(updateProfileAction, initialState);

  useEffect(() => {
    if (state.success) {
      onOpenChange(false);
    }
  }, [state.success, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑个人资料</DialogTitle>
        </DialogHeader>

        <form action={formAction} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">昵称</label>
            <Input name="nickname" defaultValue={nickname ?? ""} placeholder="设置昵称" maxLength={30} />
            {state.fieldErrors?.nickname && (
              <p className="text-xs text-red-500">{state.fieldErrors.nickname[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">城市</label>
            <Input name="city" defaultValue={city ?? ""} placeholder="所在城市" maxLength={50} />
            {state.fieldErrors?.city && (
              <p className="text-xs text-red-500">{state.fieldErrors.city[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">简介</label>
            <textarea
              name="introduction"
              defaultValue={introduction ?? ""}
              placeholder="介绍一下自己吧"
              maxLength={200}
              rows={3}
              className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
            {state.fieldErrors?.introduction && (
              <p className="text-xs text-red-500">{state.fieldErrors.introduction[0]}</p>
            )}
          </div>

          {state.error && <p className="text-sm text-red-500">{state.error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={pending}>
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Create profile-block.tsx**

```typescript
"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";

import { EditProfileDialog } from "@/features/web/me/components/edit-profile-dialog";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Profile info block (nickname, city, introduction) with edit dialog */
export function ProfileBlock({ profile }: { profile: MeProfile }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">个人资料</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <InfoItem label="昵称" value={profile.nickname ?? "未设置"} />
        <InfoItem label="城市" value={profile.city ?? "未设置"} />
        <div className="md:col-span-2">
          <InfoItem label="简介" value={profile.introduction ?? "未设置"} />
        </div>
      </div>

      <EditProfileDialog
        open={open}
        onOpenChange={setOpen}
        nickname={profile.nickname}
        city={profile.city}
        introduction={profile.introduction}
      />
    </div>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  const isEmpty = value === "未设置";
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs text-slate-400">{label}</span>
      <span className={`text-sm ${isEmpty ? "text-slate-300" : "text-slate-700"}`}>{value}</span>
    </div>
  );
}
```

**Step 3: Commit**

```bash
git add src/features/web/me/components/profile-block.tsx src/features/web/me/components/edit-profile-dialog.tsx
git commit -m "feat: add profile block with edit dialog"
```

---

### Task 10: Account Block + Edit Email Dialog

**Files:**
- Create: `src/features/web/me/components/account-block.tsx`
- Create: `src/features/web/me/components/edit-email-dialog.tsx`

**Step 1: Create edit-email-dialog.tsx**

```typescript
"use client";

import { useState, useActionState, useEffect } from "react";
import { Loader2 } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  sendEmailCodeAction,
  updateEmailAction,
} from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface EditEmailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentEmail: string | null;
}

/** Dialog for changing email with verification code */
export function EditEmailDialog({ open, onOpenChange, currentEmail }: EditEmailDialogProps) {
  const [email, setEmail] = useState(currentEmail ?? "");
  const [codeSent, setCodeSent] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const [codeState, codeFormAction, codePending] = useActionState(sendEmailCodeAction, initialState);
  const [emailState, emailFormAction, emailPending] = useActionState(updateEmailAction, initialState);

  useEffect(() => {
    if (codeState.success) {
      setCodeSent(true);
      setCountdown(60);
    }
  }, [codeState.success]);

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => setCountdown((c) => c - 1), 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  useEffect(() => {
    if (emailState.success) {
      onOpenChange(false);
      setCodeSent(false);
      setCountdown(0);
    }
  }, [emailState.success, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>修改邮箱</DialogTitle>
        </DialogHeader>

        {!codeSent ? (
          <form action={codeFormAction} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-slate-700">新邮箱</label>
              <Input
                name="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="请输入新邮箱"
              />
              {codeState.fieldErrors?.email && (
                <p className="text-xs text-red-500">{codeState.fieldErrors.email[0]}</p>
              )}
            </div>

            {codeState.error && <p className="text-sm text-red-500">{codeState.error}</p>}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={codePending}>
                {codePending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                获取验证码
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <form action={emailFormAction} className="flex flex-col gap-4">
            <input type="hidden" name="email" value={email} />

            <p className="text-sm text-slate-500">
              验证码已发送至 <span className="font-medium text-slate-700">{email}</span>
            </p>

            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-slate-700">验证码</label>
              <Input name="code" placeholder="请输入6位验证码" maxLength={6} />
              {emailState.fieldErrors?.code && (
                <p className="text-xs text-red-500">{emailState.fieldErrors.code[0]}</p>
              )}
            </div>

            {emailState.error && <p className="text-sm text-red-500">{emailState.error}</p>}

            <div className="flex items-center justify-between">
              <button
                type="button"
                disabled={countdown > 0}
                onClick={() => setCodeSent(false)}
                className="text-sm text-teal-600 hover:text-teal-700 disabled:text-slate-400"
              >
                {countdown > 0 ? `${countdown}s 后重新获取` : "重新获取"}
              </button>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={emailPending}>
                {emailPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                确认修改
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Create account-block.tsx**

```typescript
"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";

import { EditEmailDialog } from "@/features/web/me/components/edit-email-dialog";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Account info block (username, email, phone) with email edit dialog */
export function AccountBlock({ profile }: { profile: MeProfile }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">账号信息</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <InfoItem label="用户名" value={profile.username} />
        <InfoItem label="邮箱" value={profile.email ?? "未设置"} />
        <InfoItem label="手机号" value={profile.phone ?? "暂未开放"} muted />
      </div>

      <EditEmailDialog
        open={open}
        onOpenChange={setOpen}
        currentEmail={profile.email}
      />
    </div>
  );
}

function InfoItem({ label, value, muted }: { label: string; value: string; muted?: boolean }) {
  const isEmpty = value === "未设置" || value === "暂未开放";
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs text-slate-400">{label}</span>
      <span className={`text-sm ${isEmpty || muted ? "text-slate-300" : "text-slate-700"}`}>{value}</span>
    </div>
  );
}
```

**Step 3: Commit**

```bash
git add src/features/web/me/components/account-block.tsx src/features/web/me/components/edit-email-dialog.tsx
git commit -m "feat: add account block with email change dialog"
```

---

### Task 11: Security Block + Change Password Dialog

**Files:**
- Create: `src/features/web/me/components/security-block.tsx`
- Create: `src/features/web/me/components/change-password-dialog.tsx`

**Step 1: Create change-password-dialog.tsx**

```typescript
"use client";

import { useActionState, useEffect } from "react";
import { Loader2 } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { changePasswordAction } from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** Dialog for changing password (current + new + confirm) */
export function ChangePasswordDialog({ open, onOpenChange }: ChangePasswordDialogProps) {
  const [state, formAction, pending] = useActionState(changePasswordAction, initialState);

  useEffect(() => {
    if (state.success) {
      onOpenChange(false);
    }
  }, [state.success, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>修改密码</DialogTitle>
        </DialogHeader>

        <form action={formAction} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">当前密码</label>
            <Input name="currentPassword" type="password" placeholder="请输入当前密码" />
            {state.fieldErrors?.currentPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.currentPassword[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">新密码</label>
            <Input name="newPassword" type="password" placeholder="至少8位，含大小写字母和数字" />
            {state.fieldErrors?.newPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.newPassword[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-slate-700">确认新密码</label>
            <Input name="confirmPassword" type="password" placeholder="再次输入新密码" />
            {state.fieldErrors?.confirmPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.confirmPassword[0]}</p>
            )}
          </div>

          {state.error && <p className="text-sm text-red-500">{state.error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={pending}>
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认修改
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Create security-block.tsx**

```typescript
"use client";

import { useState } from "react";
import { Pencil, Lock } from "lucide-react";

import { ChangePasswordDialog } from "@/features/web/me/components/change-password-dialog";

/** Security settings block (password) with change password dialog */
export function SecurityBlock() {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">安全设置</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="flex items-center gap-2">
        <Lock className="h-4 w-4 text-slate-400" />
        <span className="text-xs text-slate-400">密码</span>
        <span className="text-sm text-slate-700">••••••••</span>
      </div>

      <ChangePasswordDialog open={open} onOpenChange={setOpen} />
    </div>
  );
}
```

**Step 3: Commit**

```bash
git add src/features/web/me/components/security-block.tsx src/features/web/me/components/change-password-dialog.tsx
git commit -m "feat: add security block with change password dialog"
```

---

### Task 12: Display-Only Blocks (Membership, Stats, Invite)

**Files:**
- Create: `src/features/web/me/components/membership-block.tsx`
- Create: `src/features/web/me/components/stats-block.tsx`
- Create: `src/features/web/me/components/invite-block.tsx`

**Step 1: Create membership-block.tsx**

```typescript
import Link from "next/link";
import { Crown, ChevronRight } from "lucide-react";

import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Membership info block (grade, vipDueAt, beans) with upgrade link */
export function MembershipBlock({ profile }: { profile: MeProfile }) {
  const gradeLabel = USER_GRADE_LABELS[profile.grade];
  const dueDate = profile.vipDueAt
    ? new Date(profile.vipDueAt).toLocaleDateString("zh-CN")
    : "—";

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">会员信息</h3>
        <Link
          href="/auth/membership"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          升级会员
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">会员等级</span>
          <div className="flex items-center gap-1.5">
            <Crown className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-medium text-slate-700">{gradeLabel}</span>
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">到期时间</span>
          <span className="text-sm text-slate-700">{dueDate}</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">能量豆</span>
          <span className="text-sm font-medium text-slate-700">{profile.beans.toLocaleString()}</span>
        </div>
      </div>
    </div>
  );
}
```

**Step 2: Create stats-block.tsx**

```typescript
import Link from "next/link";
import { ChevronRight } from "lucide-react";

import { getLevel, getExpForLevel } from "@/consts/user-level";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Learning stats block (level, EXP progress, streaks) with leaderboard link */
export function StatsBlock({ profile }: { profile: MeProfile }) {
  const level = getLevel(profile.exp);
  const currentLevelExp = getExpForLevel(level);
  const nextLevelExp = level < 100 ? getExpForLevel(level + 1) : currentLevelExp;
  const progress = nextLevelExp > currentLevelExp
    ? ((profile.exp - currentLevelExp) / (nextLevelExp - currentLevelExp)) * 100
    : 100;

  const lastPlayed = profile.lastPlayedAt
    ? new Date(profile.lastPlayedAt).toLocaleDateString("zh-CN")
    : "—";

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">学习数据</h3>
        <Link
          href="/hall/leaderboard"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          查看排行
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="flex flex-col gap-2 md:col-span-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400">等级 Lv.{level}</span>
            <span className="text-xs text-slate-400">
              {profile.exp.toLocaleString()} / {nextLevelExp.toLocaleString()} EXP
            </span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
            <div
              className="h-full rounded-full bg-teal-500 transition-all"
              style={{ width: `${Math.min(progress, 100)}%` }}
            />
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">当前连续</span>
          <span className="text-sm font-medium text-slate-700">{profile.currentPlayStreak} 天</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">最高连续</span>
          <span className="text-sm font-medium text-slate-700">{profile.maxPlayStreak} 天</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-slate-400">上次学习</span>
          <span className="text-sm text-slate-700">{lastPlayed}</span>
        </div>
      </div>
    </div>
  );
}
```

**Step 3: Create invite-block.tsx**

```typescript
"use client";

import { useState } from "react";
import Link from "next/link";
import { ChevronRight, Copy, Check } from "lucide-react";

/** Invite code block with copy button and referral link */
export function InviteBlock({ inviteCode }: { inviteCode: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(inviteCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">邀请推广</h3>
        <Link
          href="/hall/invite"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          去推广
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="flex flex-col gap-1">
        <span className="text-xs text-slate-400">邀请码</span>
        <div className="flex items-center gap-3">
          <span className="rounded-lg bg-slate-50 px-4 py-2 font-mono text-lg font-bold tracking-widest text-slate-900">
            {inviteCode}
          </span>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1 text-sm text-teal-600 hover:text-teal-700"
          >
            {copied ? (
              <>
                <Check className="h-4 w-4" />
                已复制
              </>
            ) : (
              <>
                <Copy className="h-4 w-4" />
                复制
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
```

**Step 4: Verify build**

```bash
npm run build
```

**Step 5: Commit**

```bash
git add src/features/web/me/components/membership-block.tsx src/features/web/me/components/stats-block.tsx src/features/web/me/components/invite-block.tsx
git commit -m "feat: add display-only blocks (membership, stats, invite)"
```

---

### Task 13: Final Build Verification & Dev Test

**Step 1: Run full build**

```bash
npm run build
```

Fix any build errors.

**Step 2: Run dev server and test manually**

```bash
npm run dev
```

Navigate to `http://localhost:3000/hall/me` and verify:
- PageTopBar renders correctly
- Hero shows avatar, name, level, grade, stats
- All 6 blocks render with correct data
- Edit profile dialog opens and saves
- Avatar upload works
- Edit email sends code
- Change password validates correctly
- Membership links to /auth/membership
- Leaderboard links to /hall/leaderboard
- Invite code copies to clipboard

**Step 3: Run lint**

```bash
npm run lint
```

Fix any lint issues.

**Step 4: Final commit if any fixes**

```bash
git add -A
git commit -m "fix: resolve build/lint issues for personal center"
```
