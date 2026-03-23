# Notice Page Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild the notice page with real data — simplified Notice model, rename notification->notice, unread red dot indicators, scrollable pagination, and publish modal for admin user.

**Architecture:** Server-side unread check in hall layout, cursor-based infinite scroll on notice list, server actions for data fetching and mutation. Dynamic Lucide icon rendering via lookup map.

**Tech Stack:** Next.js 16 App Router, Prisma v7, shadcn/ui Dialog, Zod, sonner, lucide-react

---

### Task 1: Update Prisma Schema — Notice Model

**Files:**
- Modify: `prisma/schema/notice.prisma`

**Step 1: Replace the Notice model**

Replace the entire contents of `prisma/schema/notice.prisma` with:

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

This removes `titleLabel`, `endLabels`, `url`, `isEnabled` and adds `icon`, `isActive`.

**Step 2: Verify schema is valid**

Run: `npx prisma validate`
Expected: "The schemas are valid."

**Step 3: Commit**

```bash
git add prisma/schema/notice.prisma
git commit -m "refactor: simplify Notice model — remove unused fields, add icon and isActive"
```

---

### Task 2: Update Prisma Schema — User Model

**Files:**
- Modify: `prisma/schema/user.prisma`

**Step 1: Add lastReadNoticeAt field to User model**

Add the following line after the `payDueAt` field (line 16), before `createdAt`:

```prisma
  lastReadNoticeAt DateTime? @map("last_read_notice_at") @db.Timestamptz
```

**Step 2: Verify schema is valid**

Run: `npx prisma validate`
Expected: "The schemas are valid."

**Step 3: Commit**

```bash
git add prisma/schema/user.prisma
git commit -m "feat: add lastReadNoticeAt field to User model"
```

---

### Task 3: Create and Run Migration

**Files:**
- Creates: `prisma/migrations/<timestamp>_notice_simplify/migration.sql`

**Step 1: Generate migration**

Run: `npx prisma migrate dev --name notice_simplify`

This will:
- Drop columns `title_label`, `end_labels`, `url`, `is_enabled` from `notices`
- Add columns `icon` (TEXT, nullable), `is_active` (BOOLEAN, default true) to `notices`
- Add column `last_read_notice_at` (TIMESTAMPTZ, nullable) to `users`
- Regenerate Prisma client

Expected: "Your database is now in sync with your schema."

**Step 2: Verify Prisma client works**

Run: `npx prisma generate`
Expected: "Generated Prisma Client"

**Step 3: Commit**

```bash
git add prisma/migrations/ src/generated/
git commit -m "feat: migration — simplify notices table, add last_read_notice_at to users"
```

---

### Task 4: Create Notice Model Layer — Queries

**Files:**
- Create: `src/models/notice/notice.query.ts`

**Step 1: Create the query file**

```typescript
import "server-only";

import { db } from "@/lib/db";

const PAGE_SIZE = 20;

/** Paginated active notices, ordered newest first */
export async function getNotices(cursor?: string, limit = PAGE_SIZE) {
  const items = await db.notice.findMany({
    where: { isActive: true },
    orderBy: { createdAt: "desc" },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
    select: {
      id: true,
      title: true,
      content: true,
      icon: true,
      createdAt: true,
    },
  });

  const hasMore = items.length > limit;
  const list = hasMore ? items.slice(0, limit) : items;
  const nextCursor = hasMore ? list[list.length - 1].id : null;

  return { items: list, nextCursor };
}

/** Return type for a single notice item */
export type NoticeItem = Awaited<ReturnType<typeof getNotices>>["items"][number];

/** Get the createdAt of the most recent active notice (for unread check) */
export async function getLatestNoticeTime(): Promise<Date | null> {
  const notice = await db.notice.findFirst({
    where: { isActive: true },
    orderBy: { createdAt: "desc" },
    select: { createdAt: true },
  });

  return notice?.createdAt ?? null;
}
```

**Step 2: Commit**

```bash
git add src/models/notice/notice.query.ts
git commit -m "feat: add notice query model — paginated list and latest notice time"
```

---

### Task 5: Create Notice Model Layer — Mutations

**Files:**
- Create: `src/models/notice/notice.mutation.ts`

**Step 1: Create the mutation file**

```typescript
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";

/** Create a new notice */
export async function createNotice(data: {
  title: string;
  content?: string;
  icon?: string;
}) {
  return db.notice.create({
    data: {
      id: ulid(),
      title: data.title,
      content: data.content ?? null,
      icon: data.icon ?? null,
    },
    select: {
      id: true,
      title: true,
      content: true,
      icon: true,
      createdAt: true,
    },
  });
}
```

**Step 2: Commit**

```bash
git add src/models/notice/notice.mutation.ts
git commit -m "feat: add notice mutation model — createNotice"
```

---

### Task 6: Add updateLastReadNoticeAt to User Mutations

**Files:**
- Modify: `src/models/user/user.mutation.ts`

**Step 1: Add the function at the end of the file**

Append after the existing `incrementUserExp` function:

```typescript
/** Update the user's last-read-notice timestamp to now */
export async function updateLastReadNoticeAt(userId: string) {
  return db.user.update({
    where: { id: userId },
    data: { lastReadNoticeAt: new Date() },
    select: { id: true },
  });
}
```

**Step 2: Commit**

```bash
git add src/models/user/user.mutation.ts
git commit -m "feat: add updateLastReadNoticeAt to user mutations"
```

---

### Task 7: Add lastReadNoticeAt to User Profile Query

**Files:**
- Modify: `src/models/user/user.query.ts`

**Step 1: Add lastReadNoticeAt to getUserProfile select**

In `getUserProfile` function, add `lastReadNoticeAt: true` to the `select` object (after `inviteCode: true`):

```typescript
select: {
  id: true,
  username: true,
  nickname: true,
  email: true,
  grade: true,
  exp: true,
  inviteCode: true,
  lastReadNoticeAt: true,
  avatar: {
    select: { url: true },
  },
},
```

**Step 2: Add lastReadNoticeAt to the return object**

In the return statement, add:

```typescript
lastReadNoticeAt: user.lastReadNoticeAt,
```

**Step 3: Commit**

```bash
git add src/models/user/user.query.ts
git commit -m "feat: include lastReadNoticeAt in user profile query"
```

---

### Task 8: Create Notice Zod Schema

**Files:**
- Create: `src/features/web/notice/schemas/notice.schema.ts`

**Step 1: Create the schema file**

```typescript
import { z } from "zod";

/** Validation schema for creating a new notice */
export const createNoticeSchema = z.object({
  title: z
    .string()
    .min(1, "标题不能为空")
    .max(200, "标题不能超过 200 个字符"),
  content: z
    .string()
    .max(2000, "内容不能超过 2000 个字符")
    .optional()
    .transform((v) => v || undefined),
  icon: z
    .string()
    .max(50, "图标名称不能超过 50 个字符")
    .optional()
    .transform((v) => v || undefined),
});

export type CreateNoticeInput = z.infer<typeof createNoticeSchema>;
```

**Step 2: Commit**

```bash
git add src/features/web/notice/schemas/notice.schema.ts
git commit -m "feat: add Zod validation schema for notice creation"
```

---

### Task 9: Create Notice Server Actions

**Files:**
- Create: `src/features/web/notice/actions/notice.action.ts`

**Step 1: Create the actions file**

```typescript
"use server";

import { auth } from "@/lib/auth";
import { getNotices } from "@/models/notice/notice.query";
import type { NoticeItem } from "@/models/notice/notice.query";
import { createNotice } from "@/models/notice/notice.mutation";
import { updateLastReadNoticeAt } from "@/models/user/user.mutation";
import { createNoticeSchema } from "@/features/web/notice/schemas/notice.schema";

/** Fetch next page of notices */
export async function fetchNoticesAction(cursor?: string): Promise<{
  items: NoticeItem[];
  nextCursor: string | null;
}> {
  return getNotices(cursor);
}

/** Create a new notice (rainson only) */
export async function createNoticeAction(
  input: { title: string; content?: string; icon?: string }
): Promise<{ data: NoticeItem } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.name || session.user.name !== "rainson") {
      return { error: "无权限" };
    }

    const parsed = createNoticeSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.errors[0].message };
    }

    const notice = await createNotice(parsed.data);
    return { data: notice };
  } catch {
    return { error: "创建失败" };
  }
}

/** Mark notices as read for the current user */
export async function markNoticesReadAction(): Promise<void> {
  const session = await auth();
  if (!session?.user?.id) return;
  await updateLastReadNoticeAt(session.user.id);
}
```

**Step 2: Commit**

```bash
git add src/features/web/notice/actions/notice.action.ts
git commit -m "feat: add notice server actions — fetch, create, markRead"
```

---

### Task 10: Create useNoticeList Hook

**Files:**
- Create: `src/features/web/notice/hooks/use-notice-list.ts`

**Step 1: Create the hook file**

```typescript
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import type { NoticeItem } from "@/models/notice/notice.query";
import {
  fetchNoticesAction,
  createNoticeAction,
} from "@/features/web/notice/actions/notice.action";

interface UseNoticeListParams {
  initialItems: NoticeItem[];
  initialCursor: string | null;
}

/** Manages infinite scroll and publish for notice list */
export function useNoticeList({ initialItems, initialCursor }: UseNoticeListParams) {
  const [items, setItems] = useState(initialItems);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialCursor !== null);
  const [isLoading, setIsLoading] = useState(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  /** Sync state when server re-renders with new initial data */
  useEffect(() => {
    setItems(initialItems);
    setCursor(initialCursor);
    setHasMore(initialCursor !== null);
  }, [initialItems, initialCursor]);

  /** Load next page of notices */
  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || !cursor) return;
    setIsLoading(true);
    try {
      const result = await fetchNoticesAction(cursor);
      setItems((prev) => [...prev, ...result.items]);
      setCursor(result.nextCursor);
      setHasMore(result.nextCursor !== null);
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, hasMore, cursor]);

  /** IntersectionObserver for infinite scroll */
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading) loadMore();
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoading, loadMore]);

  /** Publish a new notice and prepend to list */
  const publishNotice = useCallback(
    async (input: { title: string; content?: string; icon?: string }) => {
      const result = await createNoticeAction(input);
      if ("error" in result) {
        toast.error(result.error);
        return false;
      }
      setItems((prev) => [result.data, ...prev]);
      toast.success("通知已发布");
      return true;
    },
    []
  );

  return {
    items,
    isLoading,
    hasMore,
    sentinelRef,
    publishNotice,
  };
}
```

**Step 2: Commit**

```bash
git add src/features/web/notice/hooks/use-notice-list.ts
git commit -m "feat: add useNoticeList hook — infinite scroll and publish"
```

---

### Task 11: Create Lucide Icon Map Helper

**Files:**
- Create: `src/features/web/notice/helpers/notice-icon.ts`

**Step 1: Create the icon map helper**

```typescript
import type { LucideIcon } from "lucide-react";
import {
  MessageCircleMore,
  Swords,
  Bell,
  Megaphone,
  Trophy,
  Gift,
  Rocket,
  Star,
  Shield,
  BookOpen,
  Calendar,
  UserPlus,
  Heart,
  Zap,
  PartyPopper,
  Info,
  AlertTriangle,
  CheckCircle2,
  Sparkles,
  Crown,
} from "lucide-react";

/** Map of supported Lucide icon names to components */
const iconMap: Record<string, LucideIcon> = {
  "message-circle-more": MessageCircleMore,
  swords: Swords,
  bell: Bell,
  megaphone: Megaphone,
  trophy: Trophy,
  gift: Gift,
  rocket: Rocket,
  star: Star,
  shield: Shield,
  "book-open": BookOpen,
  calendar: Calendar,
  "user-plus": UserPlus,
  heart: Heart,
  zap: Zap,
  "party-popper": PartyPopper,
  info: Info,
  "alert-triangle": AlertTriangle,
  "check-circle-2": CheckCircle2,
  sparkles: Sparkles,
  crown: Crown,
};

/** Resolve a Lucide icon name string to its component, fallback to MessageCircleMore */
export function resolveNoticeIcon(name?: string | null): LucideIcon {
  if (!name) return MessageCircleMore;
  return iconMap[name] ?? MessageCircleMore;
}
```

**Step 2: Commit**

```bash
git add src/features/web/notice/helpers/notice-icon.ts
git commit -m "feat: add Lucide icon map helper for dynamic notice icons"
```

---

### Task 12: Create NoticeItem Component

**Files:**
- Create: `src/features/web/notice/components/notice-item.tsx`

**Step 1: Create the component**

```typescript
import type { NoticeItem as NoticeItemType } from "@/models/notice/notice.query";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";

interface NoticeItemProps {
  notice: NoticeItemType;
}

/** Renders a single notice row with dynamic icon */
export function NoticeItem({ notice }: NoticeItemProps) {
  const Icon = resolveNoticeIcon(notice.icon);

  return (
    <div className="flex gap-3.5 border-b border-slate-200 px-4 py-4 last:border-b-0 md:px-5">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px] bg-teal-50">
        <Icon className="h-[18px] w-[18px] text-teal-600" />
      </div>
      <div className="flex flex-1 flex-col gap-1.5">
        <span className="text-sm font-semibold text-slate-900">
          {notice.title}
        </span>
        {notice.content && (
          <span className="text-[13px] leading-[1.5] text-slate-500">
            {notice.content}
          </span>
        )}
        <span className="text-xs text-slate-400">
          {formatRelativeTime(notice.createdAt)}
        </span>
      </div>
    </div>
  );
}
```

**Step 2: Create the relative time helper**

Create `src/features/web/notice/helpers/notice-time.ts`:

```typescript
/** Format a date to a relative time string in Chinese */
export function formatRelativeTime(date: Date | string): string {
  const now = new Date();
  const d = new Date(date);
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);

  if (diffMin < 1) return "刚刚";
  if (diffMin < 60) return `${diffMin} 分钟前`;
  if (diffHour < 24) return `${diffHour} 小时前`;
  if (diffDay < 30) return `${diffDay} 天前`;

  return d.toLocaleDateString("zh-CN");
}
```

**Step 3: Commit**

```bash
git add src/features/web/notice/components/notice-item.tsx src/features/web/notice/helpers/notice-time.ts
git commit -m "feat: add NoticeItem component with dynamic icon and relative time"
```

---

### Task 13: Create PublishNoticeModal Component

**Files:**
- Create: `src/features/web/notice/components/publish-notice-modal.tsx`

**Step 1: Create the modal component**

```typescript
"use client";

import { useState } from "react";
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
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

interface PublishNoticeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onPublish: (input: { title: string; content?: string; icon?: string }) => Promise<boolean>;
}

/** Modal form for publishing a new notice */
export function PublishNoticeModal({ open, onOpenChange, onPublish }: PublishNoticeModalProps) {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [icon, setIcon] = useState("");
  const [submitting, setSubmitting] = useState(false);

  /** Reset form fields */
  function resetForm() {
    setTitle("");
    setContent("");
    setIcon("");
  }

  async function handleSubmit() {
    if (!title.trim()) return;
    setSubmitting(true);
    const ok = await onPublish({
      title: title.trim(),
      content: content.trim() || undefined,
      icon: icon.trim() || undefined,
    });
    setSubmitting(false);
    if (ok) {
      resetForm();
      onOpenChange(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>发布新通知</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="notice-title">标题 *</Label>
            <Input
              id="notice-title"
              placeholder="通知标题"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={submitting}
              maxLength={200}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="notice-content">内容</Label>
            <Textarea
              id="notice-content"
              placeholder="通知内容（可选）"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              disabled={submitting}
              maxLength={2000}
              rows={4}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="notice-icon">图标</Label>
            <Input
              id="notice-icon"
              placeholder="message-circle-more"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
              disabled={submitting}
              maxLength={50}
            />
            <span className="text-xs text-slate-400">
              Lucide 图标名称，留空使用默认图标
            </span>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={submitting || !title.trim()}
          >
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            发布
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Commit**

```bash
git add src/features/web/notice/components/publish-notice-modal.tsx
git commit -m "feat: add PublishNoticeModal component"
```

---

### Task 14: Create NoticesContent Component

**Files:**
- Create: `src/features/web/notice/components/notices-content.tsx`

**Step 1: Create the component**

```typescript
"use client";

import { useState } from "react";
import { Plus, Loader2, Megaphone } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { NoticeItem as NoticeItemType } from "@/models/notice/notice.query";
import { useNoticeList } from "@/features/web/notice/hooks/use-notice-list";
import { NoticeItem } from "@/features/web/notice/components/notice-item";
import { PublishNoticeModal } from "@/features/web/notice/components/publish-notice-modal";

interface NoticesContentProps {
  initialItems: NoticeItemType[];
  initialCursor: string | null;
  username: string | null;
}

/** Main notice list with infinite scroll and optional publish button */
export function NoticesContent({ initialItems, initialCursor, username }: NoticesContentProps) {
  const { items, isLoading, hasMore, sentinelRef, publishNotice } = useNoticeList({
    initialItems,
    initialCursor,
  });
  const [showPublish, setShowPublish] = useState(false);

  const canPublish = username === "rainson";

  return (
    <>
      {/* Header row with optional publish button */}
      <div className="flex items-center justify-between">
        <span className="text-[13px] text-slate-400">
          共 {items.length} 条通知{hasMore ? "+" : ""}
        </span>
        {canPublish && (
          <Button
            size="sm"
            onClick={() => setShowPublish(true)}
            className="gap-1.5"
          >
            <Plus className="h-4 w-4" />
            新通知
          </Button>
        )}
      </div>

      {/* Notice list */}
      {items.length > 0 ? (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          {items.map((notice) => (
            <NoticeItem key={notice.id} notice={notice} />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center gap-3 py-16 text-slate-400">
          <Megaphone className="h-10 w-10" />
          <span className="text-sm">暂无通知</span>
        </div>
      )}

      {/* Loading spinner */}
      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}

      {/* Infinite scroll sentinel */}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      {/* Publish modal */}
      {canPublish && (
        <PublishNoticeModal
          open={showPublish}
          onOpenChange={setShowPublish}
          onPublish={publishNotice}
        />
      )}
    </>
  );
}
```

**Step 2: Commit**

```bash
git add src/features/web/notice/components/notices-content.tsx
git commit -m "feat: add NoticesContent component with infinite scroll and publish"
```

---

### Task 15: Rename Route and Create Notice Page

**Files:**
- Delete: `src/app/(web)/hall/(main)/notifications/page.tsx`
- Delete: `src/features/web/notification/` (entire directory)
- Create: `src/app/(web)/hall/(main)/notices/page.tsx`

**Step 1: Delete old notification files**

```bash
rm -rf src/app/(web)/hall/(main)/notifications
rm -rf src/features/web/notification
```

**Step 2: Create the new notices page**

Create `src/app/(web)/hall/(main)/notices/page.tsx`:

```typescript
import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { NoticesContent } from "@/features/web/notice/components/notices-content";
import { getNotices } from "@/models/notice/notice.query";
import { markNoticesReadAction } from "@/features/web/notice/actions/notice.action";

export default async function NoticesPage() {
  const session = await auth();
  const userId = session?.user?.id;
  const username = session?.user?.name ?? null;

  const data = await getNotices();

  // Mark notices as read on page visit
  if (userId) {
    markNoticesReadAction();
  }

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="消息通知"
        subtitle="查看系统通知和公告"
      />
      <NoticesContent
        initialItems={data.items}
        initialCursor={data.nextCursor}
        username={username}
      />
    </div>
  );
}
```

Note: `markNoticesReadAction()` is intentionally not awaited — fire-and-forget to avoid blocking page render.

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: rename notifications route to notices, wire real data"
```

---

### Task 16: Update Sidebar — Route and Red Dot

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx`

**Step 1: Update the href in navSections**

Change line 42 from:
```typescript
{ icon: Bell, label: "消息通知", href: "/hall/notifications" },
```
to:
```typescript
{ icon: Bell, label: "消息通知", href: "/hall/notices" },
```

**Step 2: Add hasUnreadNotices prop to SidebarContent**

Update the `SidebarContent` function signature:

```typescript
function SidebarContent({ onNavigate, hasUnreadNotices }: { onNavigate?: () => void; hasUnreadNotices?: boolean }) {
```

**Step 3: Add red dot to NavItem for notices**

Add a `showDot` prop to the nav items data and NavItem component. Update the navSections type and NavItem:

Update `NavItem` to accept an optional `showDot` prop:

```typescript
function NavItem({
  icon: Icon,
  label,
  href,
  active,
  showDot,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  href: string;
  active: boolean;
  showDot?: boolean;
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className={`flex w-full items-center gap-3 rounded-[10px] px-3.5 py-2.5 ${
        active
          ? "bg-teal-600/10 font-semibold text-teal-600"
          : "text-slate-500 hover:bg-slate-50"
      }`}
    >
      <Icon className="h-[18px] w-[18px]" />
      <span className="text-[13px]">{label}</span>
      {showDot && (
        <span className="ml-auto h-2 w-2 rounded-full bg-red-500" />
      )}
    </Link>
  );
}
```

**Step 4: Pass showDot when rendering NavItems**

In the `SidebarContent` component, update the NavItem rendering inside the map:

```typescript
<NavItem
  key={item.label}
  icon={item.icon}
  label={item.label}
  href={item.href}
  active={pathname === item.href}
  showDot={item.href === "/hall/notices" && hasUnreadNotices}
  onClick={onNavigate}
/>
```

**Step 5: Update HallSidebar and MobileSidebarTrigger to accept and pass hasUnreadNotices**

```typescript
export function HallSidebar({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <aside className="hidden md:flex h-full w-[260px] shrink-0 flex-col border-r border-slate-200 bg-white px-5 py-6">
      <SidebarContent hasUnreadNotices={hasUnreadNotices} />
    </aside>
  );
}

export function MobileSidebarTrigger({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <button
          type="button"
          className="flex h-9 w-9 items-center justify-center rounded-lg text-slate-600 hover:bg-slate-100"
        >
          <Menu className="h-5 w-5" />
        </button>
      </SheetTrigger>
      <SheetContent side="left" className="w-[260px] p-5" showCloseButton={false}>
        <SidebarContent hasUnreadNotices={hasUnreadNotices} />
      </SheetContent>
    </Sheet>
  );
}
```

**Step 6: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx
git commit -m "feat: update sidebar — notices route, red dot indicator"
```

---

### Task 17: Update TopActions — Bell Red Dot

**Files:**
- Modify: `src/features/web/hall/components/top-actions.tsx`

**Step 1: Add hasUnreadNotices prop**

Update the function signature to accept `hasUnreadNotices`:

```typescript
export async function TopActions({
  searchPlaceholder = "搜索课程...",
  hasUnreadNotices,
}: {
  searchPlaceholder?: string;
  hasUnreadNotices?: boolean;
} = {}) {
```

**Step 2: Add red dot to the Bell button**

Replace the Bell button (around line 30-32) with:

```typescript
<button className="relative flex h-10 w-10 items-center justify-center rounded-[10px] border border-slate-200 bg-white text-slate-500 hover:bg-slate-50">
  <Bell className="h-[18px] w-[18px]" />
  {hasUnreadNotices && (
    <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-red-500" />
  )}
</button>
```

**Step 3: Commit**

```bash
git add src/features/web/hall/components/top-actions.tsx
git commit -m "feat: add red dot to Bell icon in top actions"
```

---

### Task 18: Update PageTopBar to Pass hasUnreadNotices

**Files:**
- Modify: `src/features/web/hall/components/page-top-bar.tsx`

**Step 1: Add hasUnreadNotices prop and forward to TopActions**

```typescript
import { TopActions } from "@/features/web/hall/components/top-actions";

export async function PageTopBar({
  title,
  subtitle,
  searchPlaceholder,
  hasUnreadNotices,
}: {
  title: string;
  subtitle: string;
  searchPlaceholder?: string;
  hasUnreadNotices?: boolean;
}) {
  return (
    <div className="flex w-full items-center justify-between">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold text-slate-900">{title}</h1>
        <p className="text-sm text-slate-400">{subtitle}</p>
      </div>
      <TopActions searchPlaceholder={searchPlaceholder} hasUnreadNotices={hasUnreadNotices} />
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add src/features/web/hall/components/page-top-bar.tsx
git commit -m "feat: forward hasUnreadNotices through PageTopBar to TopActions"
```

---

### Task 19: Wire Unread Check into Hall Layout

**Files:**
- Modify: `src/app/(web)/hall/(main)/layout.tsx`

**Step 1: Convert to server+client architecture**

The hall main layout is currently a `"use client"` component. We need to pass `hasUnreadNotices` from the server. Create a wrapper that computes the value server-side and passes it as a prop.

Create `src/features/web/hall/helpers/has-unread-notices.ts`:

```typescript
import "server-only";

import { getLatestNoticeTime } from "@/models/notice/notice.query";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";

/** Check if the current user has unread notices */
export async function hasUnreadNotices(): Promise<boolean> {
  const [profile, latestNoticeTime] = await Promise.all([
    fetchUserProfile(),
    getLatestNoticeTime(),
  ]);

  if (!latestNoticeTime) return false;
  if (!profile) return false;
  if (!profile.lastReadNoticeAt) return true;

  return profile.lastReadNoticeAt < latestNoticeTime;
}
```

**Step 2: Update the hall `(main)` layout to pass the value**

Replace `src/app/(web)/hall/(main)/layout.tsx` entirely:

```typescript
import { hasUnreadNotices } from "@/features/web/hall/helpers/has-unread-notices";
import { HallMainShell } from "@/features/web/hall/components/hall-main-shell";

export default async function HallMainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const unread = await hasUnreadNotices();

  return <HallMainShell hasUnreadNotices={unread}>{children}</HallMainShell>;
}
```

**Step 3: Extract existing layout UI into a client component**

Create `src/features/web/hall/components/hall-main-shell.tsx`:

```typescript
"use client";

import { GraduationCap } from "lucide-react";
import { HallSidebar, MobileSidebarTrigger } from "@/features/web/hall/components/hall-sidebar";

/** Client shell for the hall main layout — receives server-computed props */
export function HallMainShell({
  children,
  hasUnreadNotices,
}: {
  children: React.ReactNode;
  hasUnreadNotices?: boolean;
}) {
  return (
    <div className="flex h-screen w-full overflow-hidden bg-slate-50">
      {/* Mobile header */}
      <div className="fixed top-0 left-0 right-0 z-40 flex h-14 items-center gap-3 border-b bg-white px-4 md:hidden">
        <MobileSidebarTrigger hasUnreadNotices={hasUnreadNotices} />
        <div className="flex items-center gap-2">
          <GraduationCap className="h-6 w-6 text-teal-600" />
          <span className="text-base font-extrabold text-slate-900">斗学</span>
        </div>
      </div>

      {/* Desktop sidebar */}
      <HallSidebar hasUnreadNotices={hasUnreadNotices} />

      {/* Main content */}
      <main className="flex-1 overflow-y-auto pt-14 md:pt-0">{children}</main>
    </div>
  );
}
```

**Step 4: Commit**

```bash
git add src/app/(web)/hall/(main)/layout.tsx src/features/web/hall/helpers/has-unread-notices.ts src/features/web/hall/components/hall-main-shell.tsx
git commit -m "feat: wire unread notice check into hall layout, extract client shell"
```

---

### Task 20: Update All Pages Using PageTopBar to Pass hasUnreadNotices

**Files:**
- Modify: all page files under `src/app/(web)/hall/(main)/` that use `PageTopBar`

Since `PageTopBar` now accepts an optional `hasUnreadNotices` prop but it defaults to `undefined`, existing pages will continue to work without changes — the Bell dot just won't show on those pages unless they pass it. This is acceptable because:
- The sidebar red dot (from layout) is always visible
- The Bell dot in TopActions is a nice-to-have per-page

No changes needed for this task. The sidebar red dot is the primary indicator.

**Step 1: Verify build**

Run: `npm run build`
Expected: Build succeeds with no errors.

**Step 2: Commit (if any fixes needed)**

---

### Task 21: Verify Everything Works

**Step 1: Run the dev server**

Run: `npm run dev`

**Step 2: Manual verification checklist**

- [ ] Visit `/hall/notices` — page loads with empty state or real notices
- [ ] Old route `/hall/notifications` returns 404
- [ ] Sidebar shows "消息通知" linking to `/hall/notices`
- [ ] Red dot appears on sidebar "消息通知" when unread notices exist
- [ ] Red dot appears on Bell icon in top bar when unread notices exist
- [ ] Visiting `/hall/notices` clears the red dot on next navigation
- [ ] Infinite scroll loads more notices when scrolling down
- [ ] Log in as `rainson` — "新通知" button appears
- [ ] Click "新通知" — modal opens with title, content, icon fields
- [ ] Submit a notice — it appears at top of list, toast shows success
- [ ] Log in as other user — "新通知" button is hidden

**Step 3: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: address verification issues"
```
