# Hall Me (个人中心) Design

## Overview

Add a personal center page at `/hall/me` where users can view and edit their profile, account, and security settings. Display-only blocks show membership, learning stats, and invite code with contextual links.

## Page Layout

PageTopBar ("个人中心" / "管理你的个人资料和账号信息") → Hero Banner → 6 stacked blocks.

### Hero Banner

- Large clickable avatar (Uppy upload on click)
- Nickname or username as display name
- Level badge (Lv.X) + grade label (e.g. 月度会员)
- Key stats row: EXP, play streak, beans

### Blocks

| # | Block | Fields | Editable | Action |
|---|-------|--------|----------|--------|
| 1 | 个人资料 | nickname, city, introduction | Yes (Dialog) | 编辑 button |
| 2 | 账号信息 | username (read-only), email (verification code), phone (disabled) | Partial (Dialog) | 编辑 button |
| 3 | 安全设置 | password (masked) | Yes (Dialog) | 编辑 button |
| 4 | 会员信息 | grade, vipDueAt, beans | No | 升级会员 → /auth/membership |
| 5 | 学习数据 | level, EXP progress, currentPlayStreak, maxPlayStreak, lastPlayedAt | No | 查看排行 → /hall/leaderboard |
| 6 | 邀请推广 | inviteCode (copy button) | No | 去推广 → /hall/invite |

## Edit Modals

### Profile Dialog (Block 1)

| Field | Type | Validation |
|-------|------|------------|
| nickname | text | max 30 chars, optional |
| city | text | max 50 chars, optional |
| introduction | textarea | max 200 chars, optional |

### Email Dialog (Block 2)

| Field | Type | Validation |
|-------|------|------------|
| username | text (read-only) | — |
| email | text | valid email format |
| code | text (6 digits) | after 获取验证码 button |
| phone | text (disabled) | 暂未开放 hint |

Email change: enter new email → send verification code (Redis, 5 min TTL) → enter code → submit.

### Password Dialog (Block 3)

| Field | Type | Validation |
|-------|------|------------|
| currentPassword | password | required |
| newPassword | password | min 8, uppercase + lowercase + number |
| confirmPassword | password | must match newPassword |

### Avatar Upload (Hero)

Click avatar → Uppy upload dialog (existing pattern) → on success call updateAvatar action → revalidatePath.

## File Structure

### Route

- `src/app/(web)/hall/(main)/me/page.tsx`

### Feature Module

```
src/features/web/me/
├── components/
│   ├── me-hero.tsx                 # Hero banner
│   ├── profile-block.tsx           # 个人资料
│   ├── account-block.tsx           # 账号信息
│   ├── security-block.tsx          # 安全设置
│   ├── membership-block.tsx        # 会员信息
│   ├── stats-block.tsx             # 学习数据
│   ├── invite-block.tsx            # 邀请推广
│   ├── edit-profile-dialog.tsx     # Profile edit modal
│   ├── edit-email-dialog.tsx       # Email edit modal
│   ├── change-password-dialog.tsx  # Password change modal
│   └── avatar-uploader.tsx         # Avatar upload wrapper
├── schemas/
│   └── me.schema.ts                # Zod schemas
├── actions/
│   └── me.action.ts                # Server actions
├── services/
│   └── me.service.ts               # Business logic
└── types/
    └── me.types.ts                 # MeProfile type
```

### Model Layer Changes

- `user.query.ts` — add `getUserFullProfile()` with all fields for /hall/me
- `user.mutation.ts` — add `updateUserProfile()`, `updateUserEmail()`, `updateUserAvatar()`, `updateUserPassword()`

### Sidebar Change

- `hall-sidebar.tsx` line 58 — update 个人中心 href from `/hall` to `/hall/me`

## Data Flow

```
page.tsx (server) → me.service.ts → user.query.ts → DB

Dialog (client) → server action → me.service.ts → user.mutation.ts → DB
                                                 → revalidatePath("/hall/me")
```

## Decisions

- Username is read-only (cannot be changed)
- Email change requires verification code (same Redis pattern as sign-in)
- Phone field displayed but disabled ("暂未开放") — handled later
- Avatar uses existing Uppy image uploader
- Password change requires current password + new password + confirm
- Display-only blocks have contextual links (升级会员, 查看排行, 去推广)
