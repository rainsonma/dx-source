# Group Feature Design Spec

## Overview

Learning group (学习群) feature for Douxue — allows users to create study groups, manage members and subgroups, and collaborate on English learning. This spec covers group CRUD, membership, applications, and subgroups. Game connection is deferred to a future task.

## Database & Models

### Model changes

- **Remove `role` field** from `GameGroupMember` and `GameSubgroupMember`
- Groups are controlled solely by the creator (`owner_id` on `GameGroup`)
- **Add `member_count` field** to `GameGroup` — denormalized counter, incremented/decremented in join/leave/kick transactions

### Migrations required

1. **Drop `role` column** from `game_group_members` and `game_subgroup_members` (new migration)
2. **Add `member_count` column** to `game_groups` (default 0, new migration)
3. **Create `game_group_applications` table** (new migration)

### New model: GameGroupApplication

```go
type GameGroupApplication struct {
    orm.Timestamps
    ID          string `gorm:"column:id;primaryKey" json:"id"`
    GameGroupID string `gorm:"column:game_group_id" json:"game_group_id"`
    UserID      string `gorm:"column:user_id" json:"user_id"`
    Status      string `gorm:"column:status" json:"status"` // pending, accepted, rejected
}
```

Table: `game_group_applications`
Indexes: `(game_group_id, user_id, status)`, `(game_group_id, status)`
ID generation: ULID (consistent with all other models)

### New constants: application status

File: `consts/group.go`

```go
const (
    ApplicationStatusPending  = "pending"
    ApplicationStatusAccepted = "accepted"
    ApplicationStatusRejected = "rejected"
)
```

### Existing models (unchanged except role removal)

| Model | Table | Key Fields |
|-------|-------|------------|
| GameGroup | game_groups | id, name, description, owner_id, cover_id, current_game_id, invite_code, is_active, member_count |
| GameSubgroup | game_subgroups | id, game_group_id, name, description, order |
| GameGroupMember | game_group_members | id, game_group_id, user_id |
| GameSubgroupMember | game_subgroup_members | id, game_subgroup_id, user_id |

### New error codes

```go
CodeGroupNotFound       = 40407
CodeGroupForbidden      = 40301
CodeAlreadyMember       = 40009
CodeAlreadyApplied      = 40010
CodeApplicationNotFound = 40408
```

### New error sentinels

```go
ErrGroupNotFound       = errors.New("group not found")
ErrNotGroupOwner       = errors.New("not group owner")
ErrAlreadyMember       = errors.New("already a group member")
ErrAlreadyApplied      = errors.New("already applied")
ErrApplicationNotFound = errors.New("application not found")
ErrNotGroupMember      = errors.New("not a group member")
ErrCannotLeaveOwned    = errors.New("owner cannot leave own group")
ErrSubgroupNotFound    = errors.New("subgroup not found")
```

## API Endpoints

All endpoints require user JWT auth. All under `/api/groups` prefix.

### GroupController (group CRUD + applications)

| Method | Path | Action | Auth |
|--------|------|--------|------|
| GET | `/groups` | List groups | member |
| POST | `/groups` | Create group | any user |
| GET | `/groups/{id}` | Group detail | member |
| PUT | `/groups/{id}` | Update group | owner only |
| DELETE | `/groups/{id}` | Delete group | owner only |
| POST | `/groups/{id}/apply` | Apply to join | any user |
| DELETE | `/groups/{id}/apply` | Cancel own application | applicant |
| GET | `/groups/{id}/applications` | List pending apps | owner only |
| PUT | `/groups/{id}/applications/{appId}` | Accept/reject | owner only |

### GroupMemberController

| Method | Path | Action | Auth |
|--------|------|--------|------|
| GET | `/groups/{id}/members` | List members | member |
| DELETE | `/groups/{id}/members/{userId}` | Kick member | owner only |
| POST | `/groups/{id}/leave` | Leave group | member (not owner) |
| POST | `/groups/join/{code}` | Join via invite code | any user |

### GroupSubgroupController

| Method | Path | Action | Auth |
|--------|------|--------|------|
| GET | `/groups/{id}/subgroups` | List subgroups | member |
| POST | `/groups/{id}/subgroups` | Create subgroup | owner only |
| PUT | `/groups/{id}/subgroups/{sid}` | Update subgroup | owner only |
| DELETE | `/groups/{id}/subgroups/{sid}` | Delete subgroup | owner only |
| GET | `/groups/{id}/subgroups/{sid}/members` | List subgroup members | member |
| POST | `/groups/{id}/subgroups/{sid}/members` | Assign members | owner only |
| DELETE | `/groups/{id}/subgroups/{sid}/members/{userId}` | Remove member | owner only |

### Query parameters

- `GET /groups?tab=all|created|joined&cursor=&limit=` — cursor pagination
- `GET /groups/{id}/members?cursor=&limit=` — cursor pagination
- `GET /groups/{id}/applications?cursor=&limit=` — cursor pagination

### Response format

Standard envelope: `{ "code": 0, "message": "ok", "data": {...} }`
Paginated: `{ "code": 0, "message": "ok", "data": { "items": [...], "nextCursor": "...", "hasMore": true } }`

## Business Rules

1. **Create group** — creator auto-added as member, invite_code generated (8-char alphanumeric, matching existing convention), member_count initialized to 1
2. **Delete group** — cascades: delete all members, subgroups, subgroup members, applications
3. **Kick / leave** — also removes user from any subgroups within that group; decrements member_count
4. **Owner cannot leave** — must delete the group instead
5. **Join by code** — directly becomes member (no approval needed); increments member_count
6. **Invite links** — frontend concern; link embeds the invite code, frontend extracts code and calls `POST /groups/join/{code}`
7. **Apply** — creates pending application; rejects if already member or already has pending application
8. **Re-apply after rejection** — allowed; old rejected row is deleted and a new pending application is created
9. **Cancel application** — user can cancel their own pending application (deletes the row)
10. **Accept application** — creates member record, updates application status to accepted, increments member_count
11. **Reject application** — updates application status to rejected
12. **Assign subgroup members** — request body: `{ "user_ids": ["id1", "id2"] }`; all must be group members; users already in subgroup are silently skipped; transactional
13. **Delete subgroup** — cascades: delete all subgroup members
14. **tab=all** — returns all groups where the user is a member (union of created + joined, since creator is auto-added as member)
15. **Frontend role cleanup** — remove all mock role labels ("群主", "管理员", "成员") from UI; distinguish owner by comparing user_id === group.owner_id

## Backend File Structure

```
dx-api/app/
├── http/controllers/api/
│   ├── group_controller.go           # Group CRUD + applications
│   ├── group_member_controller.go    # Members: list, kick, leave, join
│   └── group_subgroup_controller.go  # Subgroups CRUD + member assignment
├── http/requests/api/
│   ├── group_request.go              # CreateGroup, UpdateGroup, HandleApplication
│   ├── group_member_request.go       # JoinByCode
│   └── group_subgroup_request.go     # CreateSubgroup, UpdateSubgroup, AssignMembers
├── services/api/
│   ├── group_service.go              # Group CRUD + invite code
│   ├── group_member_service.go       # Membership operations
│   ├── group_application_service.go  # Application flow
│   └── group_subgroup_service.go     # Subgroup CRUD + member assignment
├── models/
│   ├── game_group.go                 # (existing, unchanged)
│   ├── game_subgroup.go              # (existing, unchanged)
│   ├── game_group_member.go          # (remove role field)
│   ├── game_subgroup_member.go       # (remove role field)
│   └── game_group_application.go     # (new)
└── consts/
    └── error_code.go                 # (add new codes)
```

## Frontend File Structure

```
dx-web/src/features/web/groups/
├── types/
│   └── group.ts                      # Group, GroupMember, Subgroup, Application types
├── schemas/
│   └── group.schema.ts               # Zod schemas for forms
├── actions/
│   ├── group.action.ts               # Group CRUD + applications
│   ├── group-member.action.ts        # List, kick, leave, join
│   └── group-subgroup.action.ts      # Subgroup CRUD + assign members
├── hooks/
│   ├── use-groups.ts                 # List page state + tab switching
│   ├── use-group-detail.ts           # Detail page state
│   └── use-group-members.ts          # Member + subgroup member state
└── components/
    ├── group-list-content.tsx         # Refactor: real data, tabs, pagination
    ├── group-detail-content.tsx       # Refactor: real data, 4-column layout
    ├── create-group-dialog.tsx        # Create group modal
    ├── create-subgroup-dialog.tsx     # Create subgroup modal
    ├── group-card.tsx                 # Reusable group card component
    ├── member-list.tsx                # Member list with kick/leave buttons
    ├── subgroup-list.tsx              # Subgroup list with create button
    ├── subgroup-member-list.tsx       # Subgroup members with remove button
    └── application-list.tsx           # Pending applications (accept/reject)
```

## New UI Elements (beyond current mock)

1. **Leave group button** — group detail header (non-owners)
2. **Kick button** — per member in member list (owner view)
3. **Remove from subgroup button** — per subgroup member (owner view)
4. **Application list panel** — owner sees pending applications
5. **Accept / Reject buttons** — in application items
6. **Delete group button** — group detail (owner, with confirmation)
7. **Edit group button** — group detail (owner)

## Data Flow

```
Pages (thin) → Components → Hooks → Actions → apiClient → /api/groups/*
```

- `group-list-content.tsx` uses `useGroups()` for tab switching + cursor pagination
- `group-detail-content.tsx` uses `useGroupDetail()` for group info + `useGroupMembers()` for member/subgroup state
- Dialogs use Zod schemas for client validation, then call actions

## Implementation Order

### Phase 1: Backend foundation (TDD)
1. Create migrations: drop role columns, add member_count, create game_group_applications table
2. Update models: remove role from GameGroupMember/GameSubgroupMember, add MemberCount to GameGroup
3. Create GameGroupApplication model + consts/group.go (application status constants)
4. Add error codes and sentinels
5. Implement group_service.go + tests
6. Implement group_application_service.go + tests
7. Implement group_member_service.go + tests
8. Implement group_subgroup_service.go + tests
9. Implement controllers + request validation
10. Register routes

### Phase 2: Frontend integration
1. Define types and Zod schemas
2. Implement actions (API calls)
3. Implement hooks
4. Refactor group-list-content.tsx with real data
5. Extract group-card.tsx
6. Build create-group-dialog.tsx
7. Refactor group-detail-content.tsx with real data
8. Build member-list.tsx with kick/leave
9. Build subgroup-list.tsx + create-subgroup-dialog.tsx
10. Build subgroup-member-list.tsx
11. Build application-list.tsx

## Out of Scope (deferred)

- Game connection (current_game_id, start game modal)
- Group cover image
- QR code generation
- Group search (public discovery)
