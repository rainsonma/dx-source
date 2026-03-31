# Group Winner Crown Badge — Design Spec

## Overview

Show a small amber crown badge on the winner member(s) in the 群成员 list and on the winner subgroup in the 群小组 list on the group detail page. "Winner" means the member(s)/subgroup whose `last_won_at` is the most recent (maximum timestamp) within the group.

## Scope

- Backend: Modify two existing SQL queries to add a computed `is_last_winner` boolean
- Frontend: Add `is_last_winner` to two types, render crown icon in two list components
- Docs: Update `game-lsrw-group-rule.md`
- No database migrations, no new endpoints, no new components

## Backend Changes

### `ListGroupMembers` (group_member_service.go)

Add to the existing SELECT query:

```sql
CASE WHEN m.last_won_at IS NOT NULL
     AND m.last_won_at = (SELECT MAX(m2.last_won_at) FROM game_group_members m2 WHERE m2.game_group_id = m.game_group_id)
THEN true ELSE false END AS is_last_winner
```

Add to the `MemberItem` response struct:

```go
IsLastWinner bool `json:"is_last_winner"`
```

### `ListSubgroups` (group_subgroup_service.go)

Add to the existing SELECT query:

```sql
CASE WHEN s.last_won_at IS NOT NULL
     AND s.last_won_at = (SELECT MAX(s2.last_won_at) FROM game_subgroups s2 WHERE s2.game_group_id = s.game_group_id)
THEN true ELSE false END AS is_last_winner
```

Add to the `SubgroupItem` response struct:

```go
IsLastWinner bool `json:"is_last_winner"`
```

### Winner logic

- Solo mode: One member wins a level, their `last_won_at` is updated. That member gets `is_last_winner = true` (highest timestamp).
- Team mode: Winning subgroup's `last_won_at` AND all its members' `last_won_at` are updated to the same timestamp. The subgroup gets `is_last_winner = true` in the subgroup list, and its members get `is_last_winner = true` in the member list.
- If no member/subgroup has ever won (`last_won_at` is NULL for all), no crowns are shown.

## Frontend Changes

### Types (features/web/groups/types/group.ts)

```typescript
// Add to GroupMember:
is_last_winner: boolean;

// Add to Subgroup:
is_last_winner: boolean;
```

### Member List (features/web/groups/components/member-list.tsx)

When `member.is_last_winner` is true, render inline after the member name:

```tsx
{member.is_last_winner && (
  <Crown className="h-3.5 w-3.5 text-amber-400" />
)}
```

Import `Crown` from `lucide-react` (already a project dependency, used in play-group result panel).

### Subgroup List (features/web/groups/components/subgroup-list.tsx)

When `subgroup.is_last_winner` is true, render inline after the subgroup name:

```tsx
{subgroup.is_last_winner && (
  <Crown className="h-3.5 w-3.5 text-amber-400" />
)}
```

## Documentation

Update `docs/game-lsrw-group-rule.md`:
- Add `is_last_winner` computed field to the Database Schema section for both `game_group_members` and `game_subgroups`
- Add a brief "Winner Crown Badge" section describing the display behavior

## Verification

- `go build ./...` passes with no errors
- `npm run lint` passes with no new warnings or errors
- Existing functionality unaffected (no model changes, no migration, no endpoint signature changes)
