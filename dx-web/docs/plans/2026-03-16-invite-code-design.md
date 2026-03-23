# Invite Code System Design

## Overview

Add a permanent invite code to each user, enabling referral tracking and QR code sharing. When someone visits an invite link, they're redirected to sign-up with the referrer stored in a cookie. On sign-up, a `UserReferral` record is created to track the relationship and future commission rewards.

## Data Model Changes

### User model — add field

```prisma
inviteCode String @unique @map("invite_code") @db.VarChar(8)
```

- 8-character alphanumeric (a-z, A-Z, 0-9)
- Generated at sign-up time for every user
- Used as the invite URL path: `/invite/{inviteCode}`

### UserReferral model — update

Remove:
- `code` (redundant — `User.inviteCode` is the lookup key)
- `rewardDays` (replaced by monetary commission)

Add:
- `rewardAmount Decimal` — commission earned when invitee pays

Final shape:
```prisma
model UserReferral {
  id           String    @id @db.Char(26)
  referrerId   String    @map("referrer_id") @db.Char(26)
  inviteeId    String?   @map("invitee_id") @db.Char(26)
  status       String    @default("pending") @db.VarChar(20)
  rewardAmount Decimal   @default(0) @map("reward_amount") @db.Decimal(10, 2)
  rewardedAt   DateTime? @map("rewarded_at") @db.Timestamptz
  createdAt    DateTime  @default(now()) @map("created_at") @db.Timestamptz
  updatedAt    DateTime  @updatedAt @map("updated_at") @db.Timestamptz

  referrer User  @relation("Referrals", fields: [referrerId], references: [id], onDelete: Restrict, onUpdate: Restrict)
  invitee  User? @relation("ReferredBy", fields: [inviteeId], references: [id], onDelete: Restrict, onUpdate: Restrict)

  @@index([referrerId])
  @@index([inviteeId])
  @@index([status])
  @@index([createdAt])
  @@map("user_referrals")
}
```

### Referral status values

`pending` → `paid` → `rewarded`

- **pending**: invitee signed up, hasn't paid
- **paid**: invitee paid subscription, reward eligible
- **rewarded**: referrer received commission

Constants defined in `src/consts/referral-status.ts`.

## Invite Code Flow

### Generation

At sign-up (both `signUp` and `emailSignIn` auto-register paths), generate an 8-char alphanumeric code and save to `User.inviteCode`.

### Invite URL

`{BASE_URL}/invite/{inviteCode}`

### Redirect flow

1. User visits `/invite/abc12345`
2. Server component validates the code exists (looks up `User` by `inviteCode`)
3. Sets a `ref` cookie (value = inviteCode, 7-day expiry, httpOnly)
4. Redirects to `/auth/signup`
5. If code is invalid, redirects to `/auth/signup` without setting cookie

### Sign-up integration

1. Sign-up service reads the `ref` cookie
2. Looks up referrer by `inviteCode`
3. Creates user + creates `UserReferral` record (status: `pending`, referrerId = referrer, inviteeId = new user)
4. Clears the `ref` cookie
5. If cookie is missing or referrer not found, sign-up proceeds normally without referral

## QR Code on Invite Page

- **Library**: `easyqrcodejs` (client-side, browser rendering)
- **Location**: Invite page at `/hall/invite`
- **Content**: QR code encodes `{BASE_URL}/invite/{inviteCode}`
- **Hook**: `use-invite-qrcode.ts` takes invite URL and renders to a container ref
- **Copy button**: Copies the same invite URL to clipboard

## File Structure

```
prisma/schema/
├── user.prisma                          # Add inviteCode field
├── user-referral.prisma                 # Remove code, rewardDays; add rewardAmount

src/consts/
├── referral-status.ts                   # pending, paid, rewarded

src/models/user/
├── user.query.ts                        # Add findUserByInviteCode
├── user.mutation.ts                     # Update createUser to include inviteCode

src/models/user-referral/                # New model directory
├── user-referral.query.ts
├── user-referral.mutation.ts

src/features/web/invite/
├── helpers/
│   └── invite-code.helper.ts           # generateInviteCode (8-char alphanumeric)
├── hooks/
│   └── use-invite-qrcode.ts            # easyqrcodejs rendering hook
├── components/
│   └── invite-content.tsx              # Update: wire real data, real QR code

src/features/web/auth/
├── services/
│   ├── signup.service.ts               # Add referral logic (read cookie, create UserReferral)
│   └── signin.service.ts              # Add referral logic for emailSignIn auto-register

src/app/(web)/invite/[code]/
├── page.tsx                            # Validate code, set cookie, redirect to /auth/signup
```

No new API routes — everything through server components and server actions.
