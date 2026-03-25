# Group Invite Link Design Spec

## Overview

Make the group invite link URL functional. A public page at `/g/{code}` shows group info and handles both authenticated and unauthenticated users. Invite link joins now require owner approval (same as the "加入" apply flow).

## Backend

### New public endpoint

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/groups/invite/{code}` | none | Get group info by invite code |

Response:
```json
{
  "id": "group-id",
  "name": "group name",
  "description": "...",
  "member_count": 128,
  "owner_name": "张老师"
}
```

Only returns active groups. Returns ErrGroupNotFound for invalid/inactive codes.

### Change `POST /api/groups/join/{code}` behavior

**Before:** directly creates member + increments member_count.
**After:** calls `ApplyToGroup(userID, groupID)` — creates pending application.

Response on success: `{ "group_id": "..." }`
Errors: ErrAlreadyMember (already a member), ErrAlreadyApplied (pending application exists), ErrGroupNotFound (invalid code).

### New service function

`GetGroupByInviteCode(code string) (*GroupInviteInfo, error)` — no auth required. Looks up group by invite_code where is_active=true, joins users table for owner_name.

### File changes

```
dx-api/app/services/api/group_service.go         # add GetGroupByInviteCode
dx-api/app/services/api/group_member_service.go   # change JoinByCode to apply instead of direct join
dx-api/app/http/controllers/api/group_member_controller.go  # update JoinByCode response
dx-api/routes/api.go                              # add public GET /api/groups/invite/{code}
```

## Frontend

### New page: `dx-web/src/app/(web)/g/[code]/page.tsx`

Server component under `(web)` layout (no sidebar). Delegates to `GroupInviteContent`.

### New component: `group-invite-content.tsx`

Client component with states:

**Loading:** spinner

**Not logged in:** group info card + "登录后加入" button → `/auth/signin?redirect=/g/{code}`

**Logged in, not a member, not applied:** group info card + "加入群组" button → calls join API → on success shows "申请已提交，等待群主审核" message

**Logged in, not a member, already applied:** group info card + disabled button showing "申请审核中..."

**Logged in, already a member:** redirect to `/hall/groups/{id}`

**Invalid code:** "邀请链接无效或群组已关闭" error state

### Auth detection

Check `getAccessToken()` from `@/lib/token`. If token exists, call join API to determine membership/application status.

### Update invite URL

In `group-detail-content.tsx`, change invite URL from `/hall/groups/join/{code}` to `/g/{code}`.

### New action

In `group-member.action.ts`:
```typescript
getGroupByInviteCode(code) → GET /api/groups/invite/{code}
```

### File changes

```
dx-web/src/app/(web)/g/[code]/page.tsx                          # new page
dx-web/src/features/web/groups/components/group-invite-content.tsx  # new component
dx-web/src/features/web/groups/actions/group-member.action.ts    # add getGroupByInviteCode
dx-web/src/features/web/groups/components/group-detail-content.tsx  # update invite URL
```

## Out of Scope

- QR code generation
- Group cover image on invite page
- Social sharing meta tags
