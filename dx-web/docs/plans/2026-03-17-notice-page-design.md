# Notice Page Design

## Overview

Rebuild the notice page from mock data to a real, functional system notice page. Simplify the Notice model, rename all "notification" references to "notice", add unread indicators, scrollable pagination, and a publish modal for admin user.

## Data Model Changes

### Notice model (simplified)

```prisma
model Notice {
  id        String   @id @db.Char(26)
  title     String
  content   String?
  icon      String?
  isActive  Boolean  @default(true) @map("is_active")
  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  @@map("notices")
}
```

**Removed:** `titleLabel`, `endLabels`, `url`, `isEnabled`
**Added:** `icon` (optional, Lucide icon name string, fallback `message-circle-more`), `isActive` (boolean, default true)

### User model addition

```prisma
lastReadNoticeAt DateTime? @map("last_read_notice_at") @db.Timestamptz
```

Nullable — `null` means user has never visited notices page; any active notice triggers the red dot.

### Migration

- Drop columns from `notices`: `title_label`, `end_labels`, `url`, `is_enabled`
- Add columns to `notices`: `icon` (nullable text), `is_active` (boolean, default true)
- Add `last_read_notice_at` (nullable timestamptz) to `users`

## Route & Naming Rename

### Route

`/hall/notifications` -> `/hall/notices`

### Files

- `src/app/(web)/hall/(main)/notifications/` -> `src/app/(web)/hall/(main)/notices/`
- `src/features/web/notification/` -> `src/features/web/notice/`
- `notifications-content.tsx` -> `notices-content.tsx`
- `notification-item.tsx` -> `notice-item.tsx`

### Sidebar

Label stays as "消息通知", href updates to `/hall/notices`.

## New File Structure

```
src/models/notice/
├── notice.query.ts            # fetchNotices (paginated), fetchLatestNoticeTime
└── notice.mutation.ts         # createNotice

src/features/web/notice/
├── components/
│   ├── notices-content.tsx    # Main list + publish button
│   ├── notice-item.tsx        # Single notice row
│   └── publish-notice-modal.tsx
├── hooks/
│   └── use-notice-list.ts    # Infinite scroll + publish
├── actions/
│   └── notice.action.ts      # Server actions
└── schemas/
    └── notice.schema.ts      # Zod validation for create form
```

Also update user model: add `updateLastReadNoticeAt` to existing user mutation file.

## Unread Indicator Flow

1. **Hall layout** (server-side): fetch user's `lastReadNoticeAt` and latest active notice's `createdAt`
2. Compute `hasUnreadNotices = latestNoticeTime != null && (lastReadNoticeAt == null || lastReadNoticeAt < latestNoticeTime)`
3. Pass `hasUnreadNotices` boolean to `HallSidebar` and `TopActions`
4. **Sidebar**: red dot on "消息通知" nav item
5. **Top bar**: red dot on Bell icon

## Notices Page Flow

1. **Page server component**: fetch first page of active notices, call `updateLastReadNoticeAt(userId)`
2. **NoticesContent** (client): receives `initialItems`, `initialCursor`
3. **useNoticeList** hook: cursor-based infinite scroll (same pattern as word lists)
4. **NoticeItem**: renders dynamic Lucide icon (from `icon` field, fallback `message-circle-more`), title, content, relative timestamp

## Publish Flow (rainson only)

1. **NoticesContent** checks `session.user.name === "rainson"` — shows "新通知" button if true
2. Opens `PublishNoticeModal` with fields: title (required), content (optional), icon (optional, placeholder "message-circle-more")
3. Submit -> `createNoticeAction()` -> `createNotice()` model
4. Success -> prepend to list, close modal, toast

## Dynamic Lucide Icon

Store icon name as string in DB. Use a lookup map of commonly used icons in `NoticeItem` component. Fallback to `MessageCircleMore` for unknown/missing names.

## Decisions

- **No tabs** — system notices only
- **No per-notice read tracking** — single `lastReadNoticeAt` on user
- **Red dot updates on navigation** (Approach A) — no polling or real-time
- **Page visit = seen** — `lastReadNoticeAt` updated on page load
- **Plain text input** for icon name in publish form
