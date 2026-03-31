# Group Member & Subgroup Limits Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce a maximum of 50 members and 10 subgroups per game group, returning Chinese error messages that flow into frontend toasts.

**Architecture:** Backend-only enforcement at the Go service layer. Three service functions get limit checks. Two new error sentinels, two new error codes, two new constants. The existing frontend `toast.error(res.message)` pattern handles display without code changes.

**Tech Stack:** Go/Goravel backend, PostgreSQL (code-level FK constraints, no DB-level constraints)

---

### Task 1: Add group limit constants

**Files:**
- Modify: `dx-api/app/consts/group.go`

- [ ] **Step 1: Add limit constants**

Add after the existing `GameModeTeam` constant block:

```go
const (
	MaxGroupMembers   = 50
	MaxGroupSubgroups = 10
)
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/consts/group.go
git commit -m "feat: add MaxGroupMembers and MaxGroupSubgroups constants"
```

---

### Task 2: Add error codes and sentinels

**Files:**
- Modify: `dx-api/app/consts/error_code.go`
- Modify: `dx-api/app/services/api/errors.go`

- [ ] **Step 1: Add error codes**

In `dx-api/app/consts/error_code.go`, add after `CodeAlreadyApplied = 40010`:

```go
CodeGroupMembersFull   = 40011
CodeGroupSubgroupsFull = 40012
```

- [ ] **Step 2: Add error sentinels**

In `dx-api/app/services/api/errors.go`, add after the `ErrNotGroupMemberForAction` line:

```go
ErrGroupMembersFull   = errors.New("当前群组已满员")
ErrGroupSubgroupsFull = errors.New("每群最多 10 个小组")
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/consts/error_code.go dx-api/app/services/api/errors.go
git commit -m "feat: add error codes and sentinels for group limits"
```

---

### Task 3: Enforce member limit in ApplyToGroup

**Files:**
- Modify: `dx-api/app/services/api/group_application_service.go`

- [ ] **Step 1: Add member limit check**

In the `ApplyToGroup` function, after the group fetch (line 25 `if err := facades.Orm()...First(&group); ...`) and before the existing member check (line 30), add:

```go
if group.MemberCount >= consts.MaxGroupMembers {
	return "", ErrGroupMembersFull
}
```

Also add `"dx-api/app/consts"` to the import block if not already present.

This single check covers both:
- Direct apply: `POST /api/groups/{id}/apply`
- Join by code: `POST /api/groups/join/{code}` (which calls `ApplyToGroup` internally)

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/group_application_service.go
git commit -m "feat: enforce 50-member limit in ApplyToGroup"
```

---

### Task 4: Enforce member limit in HandleApplication (safety net)

**Files:**
- Modify: `dx-api/app/services/api/group_application_service.go`

- [ ] **Step 1: Add safety-net check on accept**

In the `HandleApplication` function, inside the `if action == "accept"` block (line 159), before the transaction (line 160), add:

```go
if group.MemberCount >= consts.MaxGroupMembers {
	return ErrGroupMembersFull
}
```

This prevents acceptance when the group filled between application submission and owner approval.

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/group_application_service.go
git commit -m "feat: enforce member limit safety net in HandleApplication"
```

---

### Task 5: Enforce subgroup limit in CreateSubgroup

**Files:**
- Modify: `dx-api/app/services/api/group_subgroup_service.go`

- [ ] **Step 1: Add subgroup count check**

In the `CreateSubgroup` function, after the `VerifyGroupOwnership` call (line 43) and before the max order query (line 47), add:

```go
type countRow struct {
	Count int64 `gorm:"column:count"`
}
var cnt countRow
if err := facades.Orm().Query().Raw(
	`SELECT COUNT(*) AS count FROM game_subgroups WHERE game_group_id = ?`, groupID,
).Scan(&cnt); err != nil {
	return "", fmt.Errorf("failed to count subgroups: %w", err)
}
if cnt.Count >= int64(consts.MaxGroupSubgroups) {
	return "", ErrGroupSubgroupsFull
}
```

Also add `"dx-api/app/consts"` to the import block.

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/group_subgroup_service.go
git commit -m "feat: enforce 10-subgroup limit in CreateSubgroup"
```

---

### Task 6: Add error mapping in controllers

**Files:**
- Modify: `dx-api/app/http/controllers/api/group_controller.go`

- [ ] **Step 1: Add limit error cases to mapGroupError**

In the `mapGroupError` function, add two new cases before the `default` case (before line 228):

```go
case errors.Is(err, services.ErrGroupMembersFull):
	return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupMembersFull, "当前群组已满员")
case errors.Is(err, services.ErrGroupSubgroupsFull):
	return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupSubgroupsFull, "每群最多 10 个小组")
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/http/controllers/api/group_controller.go
git commit -m "feat: map group limit errors to HTTP responses"
```

---

### Task 7: Verify no lint issues

**Files:** None (verification only)

- [ ] **Step 1: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: No issues.

- [ ] **Step 2: Run build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds.

- [ ] **Step 3: Run frontend lint**

Run: `cd dx-web && npm run lint`
Expected: No new lint errors introduced. (No frontend files were changed, so this should be clean.)

---

### Task 8: Update group rules documentation

**Files:**
- Modify: `docs/game-lsrw-group-rule.md`

- [ ] **Step 1: Add Group Limits section**

After the "Creating a Group" subsection (after line 18, before "### Adding Members"), add:

```markdown
### Group Limits

| Resource | Maximum | Error Message |
|----------|---------|---------------|
| Members per group | 50 | 当前群组已满员 |
| Subgroups per group | 10 | 每群最多 10 个小组 |

- **Member limit** is enforced when applying to join (direct apply or invite code) and when the owner accepts an application
- **Subgroup limit** is enforced when the owner creates a new subgroup
- When the limit is reached, the API returns the error message above, which is displayed as a toast notification
```

- [ ] **Step 2: Commit**

```bash
git add docs/game-lsrw-group-rule.md
git commit -m "docs: document group member and subgroup limits in group rules"
```
