# Group Member & Subgroup Limits — Design Spec

## Overview

Add number limits on game group members and subgroups. Each group can have at most 50 members and at most 10 subgroups. When limits are exceeded, the backend returns Chinese error messages that flow directly into frontend toast notifications.

## Limits

| Resource | Max | Error Message |
|----------|-----|---------------|
| Members per group | 50 | 当前群组已满员 |
| Subgroups per group | 10 | 每群最多 10 个小组 |

## Approach

Backend-only enforcement (Approach A). The existing frontend error handling pattern (`toast.error(res.message)`) already displays backend error messages as toasts — no frontend code changes needed.

## Backend Changes

### Constants

**`dx-api/app/consts/group.go`** — Add:

```go
MaxGroupMembers   = 50
MaxGroupSubgroups = 10
```

### Error Sentinels

**`dx-api/app/services/api/errors.go`** — Add:

```go
ErrGroupMembersFull   = errors.New("当前群组已满员")
ErrGroupSubgroupsFull = errors.New("每群最多 10 个小组")
```

### Error Codes

**`dx-api/app/consts/error_code.go`** — Add:

```go
CodeGroupMembersFull   = 40011
CodeGroupSubgroupsFull = 40012
```

### Enforcement Points

**1. `ApplyToGroup()` in `group_application_service.go`**

After fetching the group and before creating the application, check:

```go
if group.MemberCount >= consts.MaxGroupMembers {
    return "", ErrGroupMembersFull
}
```

This blocks both:
- Direct apply: `POST /api/groups/{id}/apply`
- Join by code: `POST /api/groups/join/{code}` (calls `ApplyToGroup()` internally)

**2. `HandleApplication()` in `group_application_service.go`**

Before accepting (inside `action == "accept"` branch), re-check the limit as a safety net (group could have filled between application and acceptance):

```go
if group.MemberCount >= consts.MaxGroupMembers {
    return ErrGroupMembersFull
}
```

**3. `CreateSubgroup()` in `group_subgroup_service.go`**

After ownership verification, count existing subgroups:

```sql
SELECT COUNT(*) FROM game_subgroups WHERE game_group_id = ?
```

If count >= `consts.MaxGroupSubgroups`, return `ErrGroupSubgroupsFull`.

### Controller Error Mapping

**`mapGroupError()` in `group_controller.go`** — Add two new cases:

```go
case errors.Is(err, services.ErrGroupMembersFull):
    return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupMembersFull, "当前群组已满员")
case errors.Is(err, services.ErrGroupSubgroupsFull):
    return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupSubgroupsFull, "每群最多 10 个小组")
```

## Frontend Changes

None required. Existing error handling in these components already displays `res.message` as toast:

- `group-list-content.tsx` — apply flow → `toast.error(res.message)`
- `group-invite-content.tsx` — join-by-code flow → `toast.error(res.message)`
- `group-detail-content.tsx` — create subgroup flow → `toast.error(res.message)`

## Documentation

Update `docs/game-lsrw-group-rule.md` with a **Group Limits** section documenting the 50-member and 10-subgroup caps.

## Non-Goals

- No UI changes to show current count vs. limit (e.g., "45/50")
- No database-level constraints (code-level FK strategy for PostgreSQL partitions)
- No frontend pre-checks (backend is single source of truth)
