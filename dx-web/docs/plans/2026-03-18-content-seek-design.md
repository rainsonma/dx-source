# Content Seek (求课程) Modal Design

## Overview

A modal triggered from the "求课程" button in the hall top actions bar. Users submit course requests with a name, optional description, and optional cloud disk link. Duplicate course names increment a counter instead of creating new records.

## Model

**ContentSeek** — `content_seeks` table

| Field | Type | Constraints |
|-------|------|-------------|
| id | Char(26) | PK, ULID |
| userId | Char(26) | indexed |
| courseName | VarChar(90) | unique, max 30 chars |
| description | VarChar(90) | optional, max 30 chars |
| diskUrl | VarChar(90) | optional, max 30 chars |
| count | Int | default 1 |
| createdAt | Timestamptz | default now() |
| updatedAt | Timestamptz | auto-updated |

No DB-level FK constraint (project convention — code-level only).

## Architecture

**Approach**: Server Action + shadcn Dialog

### Files

| Layer | File |
|-------|------|
| Schema | `prisma/schema/content-seek.prisma` |
| Model | `src/models/content-seek/content-seek.query.ts` |
| Model | `src/models/content-seek/content-seek.mutation.ts` |
| Zod | `src/features/web/hall/schemas/content-seek.schema.ts` |
| Service | `src/features/web/hall/services/content-seek.service.ts` |
| Action | `src/features/web/hall/actions/content-seek.action.ts` |
| Component | `src/features/web/hall/components/content-seek-modal.tsx` |
| Component | `src/features/web/hall/components/content-seek-button.tsx` |

### Submit Flow

1. User fills form → calls `submitContentSeekAction({ courseName, description, diskUrl })`
2. Action: auth check → Zod validate → call service
3. Service: check if courseName exists
   - **Exists**: increment count +1, return `{ duplicate: true }`
   - **New**: create record with ulid(), return `{ success: true }`
4. Client:
   - Duplicate → toast "已有相同申请，正在处理中..." (modal stays open)
   - Success → toast success message, close modal, reset form

### TopActions Change

Extract "求课程" button into `ContentSeekButton` (client component) wrapping button + Dialog. Server component `TopActions` renders `<ContentSeekButton />` in place of current static button.

## Validation

All text inputs: max 30 Chinese characters. `courseName` is required; `description` and `diskUrl` are optional.
