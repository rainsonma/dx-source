# Suggestion Feedback (建议反馈) Modal Design

## Overview

A modal triggered from the Flag icon button in the hall top actions bar. Users submit feedback by selecting a type tag and writing a description. Duplicate type+description combinations increment a counter instead of creating new records.

## Constants

`src/consts/feedback-type.ts` — 5 feedback types:

| Key | Label |
|-----|-------|
| feature | 功能建议 |
| content | 内容纠错 |
| ux | 界面体验 |
| bug | Bug 报告 |
| other | 其它 |

## Model

**Feedback** — `feedbacks` table

| Field | Type | Constraints |
|-------|------|-------------|
| id | Char(26) | PK, ULID |
| userId | Char(26) | indexed |
| type | VarChar(20) | required |
| description | VarChar(600) | required, max 200 chars |
| count | Int | default 1 |
| createdAt | Timestamptz | default now() |
| updatedAt | Timestamptz | auto-updated |

Unique constraint on `(type, description)` for duplicate detection. No DB-level FK constraint.

## Architecture

**Approach**: Server Action + shadcn Dialog

### Files

| Layer | File |
|-------|------|
| Consts | `src/consts/feedback-type.ts` |
| Schema | `prisma/schema/feedback.prisma` |
| Model | `src/models/feedback/feedback.query.ts` |
| Model | `src/models/feedback/feedback.mutation.ts` |
| Zod | `src/features/web/hall/schemas/feedback.schema.ts` |
| Service | `src/features/web/hall/services/feedback.service.ts` |
| Action | `src/features/web/hall/actions/feedback.action.ts` |
| Component | `src/features/web/hall/components/feedback-modal.tsx` |
| Component | `src/features/web/hall/components/feedback-button.tsx` |

### Submit Flow

1. User selects type tag + fills description → calls `submitFeedbackAction({ type, description })`
2. Action: auth check → Zod validate → call service
3. Service: check if same type + description exists
   - **Exists**: increment count +1, return `{ duplicate: true }`
   - **New**: create record with ulid(), return `{ success: true }`
4. Client:
   - Duplicate → toast "已有相同建议，正在处理中..." (modal stays open)
   - Success → toast success message, close modal, reset form

### TopActions Change

Extract Flag icon button into `FeedbackButton` (client component) wrapping button + Dialog. Server component `TopActions` renders `<FeedbackButton />` in place of current static button.

## Validation

- `type`: required, must be one of the 5 type keys
- `description`: required, max 200 characters, with N/200 counter in textarea
