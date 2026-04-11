# Signup Grade & Invite Referral Fixes

**Date:** 2026-04-11
**Status:** Approved for implementation
**Scope:** dx-api (Goravel) + dx-web (Next.js 16)

## Problem Statement

Two bugs and two UI polish items in the signup/signin and invite flows:

1. **Empty `users.grade`** — after signup or email-code signin, new user rows land
   with `grade = ""` instead of `"free"`. A `consts.UserGradeFree = "free"`
   constant exists but is never referenced by the auth service.
2. **Empty `user_referrals`** — signing up via `http://localhost/invite/<code>`
   never produces a `user_referrals` row. The invite page calls a backend
   validate endpoint that does not exist, the `ref` cookie is never set, no
   request field carries the invite code, and no service code creates referral
   rows.
3. **UI polish — signup form:** the "信息补充（可跳过）" section header above
   the 账号 input is visual clutter and should be removed.
4. **UI polish — `/hall/invite`:** the 复制链接 button's icon never changes after
   a successful copy. Users want a short-lived check icon as feedback.

Additionally, when a user lands on `/auth/signup` because they followed an
invite link, the form should display a **"正在通过好友邀请注册"** badge so the
user knows a referrer will be credited.

## Non-Goals

- **No backfill** of existing `users` rows that already have `grade = ""`. The
  fix is forward-only. A separate migration can be written later if desired.
- **No changes to the `SignUpRequest` struct.** The invite code is read from
  the existing httpOnly `ref` cookie on the backend, not from the request body.
- **No changes to referral statuses/rewards logic.** New rows land with
  `status = "pending"`; the existing paid/rewarded transitions are untouched.
- **No changes to the signup form field set** beyond removing one section
  header and adding one conditional badge.

## Architecture Context

- **Same-origin cookie flow.** The nginx config in `deploy/nginx/nginx.dev.conf`
  proxies `/api/*` to the Go backend and `/` to Next.js from the same
  `localhost:80` origin. The browser automatically forwards the httpOnly `ref`
  cookie set by Next.js to `/api/auth/signup`. `dx-web/src/lib/api-client.ts`
  already uses `credentials: "include"`. This makes cookie-based invite-code
  propagation the least invasive fix — no form field, no request-struct change.
- **Code-level FK constraints.** The project uses PostgreSQL partitions and
  keeps FK checks in Go code. We follow this pattern: when creating the
  `user_referral` row, we look up the referrer via a plain query rather than
  relying on DB-level FK validation.
- **ID generation.** The project uses `uuid.Must(uuid.NewV7()).String()` (see
  `auth_service.go:74`). The new `user_referral` row uses the same pattern.
- **Service signatures already take `ctx`.** `SignUp` and `SignInByEmail` both
  accept `contractshttp.Context`, so the cookie can be read inside the service
  with no controller or request plumbing changes.

## Design Overview

Six implementation tasks, split across backend and frontend:

| # | Task | Surface |
|---|------|---------|
| T1 | Set `grade = "free"` on user creation | dx-api |
| T2 | Add `/api/invite/validate` public endpoint | dx-api |
| T3 | Record `user_referrals` row during signup & auto-register | dx-api |
| T4 | Remove "信息补充（可跳过）" row from signup form | dx-web |
| T5 | Swap Copy → Check icon for 2s on successful copy | dx-web |
| T6 | Show "正在通过好友邀请注册" badge when `ref` cookie present | dx-web |

## T1 — Set `grade = "free"` on user creation

### File: `dx-api/app/services/api/auth_service.go`

Add import:

```go
import "dx-api/app/consts"
```

In `SignUp()` (currently lines 73-80), set `Grade` in the struct literal:

```go
user := models.User{
    ID:         uuid.Must(uuid.NewV7()).String(),
    Grade:      consts.UserGradeFree,
    Username:   username,
    Email:      &emailStr,
    Password:   hashedPassword,
    IsActive:   true,
    InviteCode: helpers.GenerateInviteCode(8),
}
```

Apply the same change in the auto-register branch of `SignInByEmail()`
(currently lines 123-130).

### Verification

- Unit/manual: create a user, assert `grade == "free"`.
- Existing tests that don't inspect `grade` are untouched.

## T2 — `/api/invite/validate` public endpoint

### File: `dx-api/app/services/api/referral_service.go`

Add:

```go
// ValidateInviteCode reports whether a non-empty invite_code matches an active user.
func ValidateInviteCode(code string) (bool, error) {
    if code == "" {
        return false, nil
    }
    var user models.User
    err := facades.Orm().Query().
        Select("id", "is_active").
        Where("invite_code", code).
        First(&user)
    if err != nil {
        return false, fmt.Errorf("failed to look up invite code: %w", err)
    }
    if user.ID == "" || !user.IsActive {
        return false, nil
    }
    return true, nil
}
```

### File: `dx-api/app/http/controllers/api/user_referral_controller.go`

Add a method. Matches the existing `helpers.Success` envelope and the shape
expected by `dx-web/src/app/(web)/invite/[code]/page.tsx:18`
(`{ code: 0, data: { valid: boolean } }`):

```go
// ValidateCode is a public endpoint that reports whether an invite_code is valid.
func (c *UserReferralController) ValidateCode(ctx contractshttp.Context) contractshttp.Response {
    code := ctx.Request().Query("code", "")
    ok, err := services.ValidateInviteCode(code)
    if err != nil {
        return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to validate invite")
    }
    return helpers.Success(ctx, map[string]any{"valid": ok})
}
```

### File: `dx-api/routes/api.go`

The `userReferralController` is currently instantiated inside the protected
group at line 179. Hoist it next to the other controllers declared at the top
of the `/api` group (alongside `authController` and `emailController`), then
register the public route inside the `/api` group before the protected block:

```go
userReferralController := apicontrollers.NewUserReferralController()

// ... existing public routes ...

router.Get("/invite/validate", userReferralController.ValidateCode)
```

Remove the duplicate instantiation that previously lived inside the protected
block — keep only the route registrations (`protected.Get("/invite", …)` and
`protected.Get("/referrals", …)`).

### Verification

- `curl "http://localhost/api/invite/validate?code=<existing-user-invite-code>"`
  returns `{ "code": 0, "data": { "valid": true } }`.
- `curl "http://localhost/api/invite/validate?code=garbage"` returns
  `{ "code": 0, "data": { "valid": false } }`.
- Empty `code` param also returns `valid: false`, not a 500.

## T3 — Record `user_referrals` row during signup & auto-register

### File: `dx-api/app/services/api/referral_service.go`

Add:

```go
// RecordReferralIfPresent creates a user_referrals row when the caller's request
// carries a valid `ref` cookie matching an active user. Never blocks signup:
// any lookup or create failure is returned so the caller can log it, but the
// caller does not surface the error to the API response.
func RecordReferralIfPresent(ctx contractshttp.Context, inviteeID string) error {
    refCode := ctx.Request().Cookie("ref")
    if refCode == "" {
        return nil
    }

    var referrer models.User
    err := facades.Orm().Query().
        Select("id", "is_active").
        Where("invite_code", refCode).
        First(&referrer)
    if err != nil {
        return fmt.Errorf("failed to look up referrer: %w", err)
    }
    if referrer.ID == "" || !referrer.IsActive {
        return nil // silently skip unknown or inactive referrers
    }
    if referrer.ID == inviteeID {
        return nil // self-referral guard
    }

    invitee := inviteeID
    record := models.UserReferral{
        ID:         uuid.Must(uuid.NewV7()).String(),
        ReferrerID: referrer.ID,
        InviteeID:  &invitee,
        Status:     consts.ReferralStatusPending,
    }
    if err := facades.Orm().Query().Create(&record); err != nil {
        return fmt.Errorf("failed to create user_referral: %w", err)
    }
    return nil
}
```

Imports to add (if not already present) at the top of the file:

```go
import (
    "github.com/google/uuid"
    contractshttp "github.com/goravel/framework/contracts/http"
    "dx-api/app/consts"
)
```

### File: `dx-api/app/services/api/auth_service.go`

In `SignUp()`, after the existing `Create(&user)` block succeeds, add:

```go
if refErr := RecordReferralIfPresent(ctx, user.ID); refErr != nil {
    facades.Log().Warningf("record referral failed: %v", refErr)
}
```

Apply the same block in the auto-register branch of `SignInByEmail()` right
after the create call there.

### Verification

- Signup with `ref` cookie set to a valid invite_code → new `user_referrals`
  row with `status = "pending"` and correct `referrer_id`/`invitee_id`.
- Signup with no `ref` cookie → no `user_referrals` row.
- Signup with a garbage `ref` cookie → no `user_referrals` row, warning log line,
  signup still returns 200 with a valid session token.
- Email-code signin that triggers auto-register → same outcomes as signup.

## T4 — Remove "信息补充（可跳过）" row

### File: `dx-web/src/features/web/auth/components/sign-up-form.tsx`

Delete the block currently on lines 134–142:

```tsx
<div className="h-px" />

{/* Section Header */}
<div className="flex items-center gap-1.5">
  <Info className="h-3.5 w-3.5 text-teal-600" />
  <span className="text-[13px] text-slate-700">
    信息补充（可跳过）
  </span>
</div>
```

Remove `Info` from the `lucide-react` import on line 4 since it is the only
usage in this file.

### Verification

- `npm run lint` passes (unused-import check).
- Signup form renders without the section header; 账号 input immediately
  follows the code input area.

## T5 — Copy → Check icon swap on `/hall/invite`

### File: `dx-web/src/features/web/invite/components/invite-content.tsx`

1. Add `Check` to the `lucide-react` import (lines 4–13).
2. Add a local state flag near the existing `useState` at line 37:

```tsx
const [copied, setCopied] = useState(false);
```

3. Replace `handleCopy` (lines 47–53):

```tsx
const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(inviteUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  } catch {
    // Silently fail if clipboard access is denied
  }
};
```

4. Replace the icon inside the button (line 88):

```tsx
{copied ? (
  <Check className="h-3.5 w-3.5" />
) : (
  <Copy className="h-3.5 w-3.5" />
)}
```

The button label "复制链接" stays the same; only the leading icon switches.

### Verification

- Click the button → icon becomes a check within one frame.
- After ~2 seconds the icon returns to the copy glyph.
- `npm run lint` passes.

## T6 — "正在通过好友邀请注册" badge on signup

### File: `dx-web/src/app/(web)/auth/signup/page.tsx`

Convert the page to an async server component that reads the `ref` cookie:

```tsx
import { cookies } from "next/headers";
import { SignUpForm } from "@/features/web/auth/components/sign-up-form";

export default async function SignUpPage() {
  const cookieStore = await cookies();
  const hasInviteRef = Boolean(cookieStore.get("ref")?.value);
  return <SignUpForm hasInviteRef={hasInviteRef} />;
}
```

### File: `dx-web/src/features/web/auth/components/sign-up-form.tsx`

1. Accept an optional `hasInviteRef` prop:

```tsx
type SignUpFormProps = {
  hasInviteRef?: boolean;
};

export function SignUpForm({ hasInviteRef }: SignUpFormProps) {
  // ...existing body unchanged except for the header block...
}
```

2. Add `Gift` to the `lucide-react` import.

3. Render the badge inside the header block (lines 39–46 area), directly below
   the subtitle, only when `hasInviteRef` is true:

```tsx
{/* Header */}
<div className="flex flex-col items-center gap-2">
  <h1 className="text-[32px] font-extrabold text-slate-900">
    创建斗学专属账号
  </h1>
  <p className="text-sm text-slate-400">
    进入斗学英语游戏世界，开启英语学习冒险之旅
  </p>
  {hasInviteRef && (
    <div className="flex items-center gap-1.5 rounded-full bg-teal-50 px-3 py-1 text-xs font-medium text-teal-700">
      <Gift className="h-3 w-3" />
      正在通过好友邀请注册
    </div>
  )}
</div>
```

### Notes

- The badge is cosmetic; the authoritative referral creation happens in T3.
  If the badge ever shows but the cookie is missing server-side during the POST
  (extremely unlikely — same-origin, same request cycle), the worst case is a
  missing referral row, not a bad write.
- `hasInviteRef` is optional to preserve backwards compatibility with any
  direct imports; missing prop defaults to `undefined` which falsy-guards
  cleanly.

### Verification

- Visit `/invite/<validCode>` → redirected to `/auth/signup` → badge visible.
- Visit `/auth/signup` directly → badge hidden.
- `npm run lint` passes.

## Error Handling & Resilience

| Situation | Behavior |
|-----------|----------|
| `ref` cookie missing | Signup proceeds; no referral row; no log |
| `ref` cookie set but invite_code not found | Signup proceeds; no referral row; warning log |
| Referrer is inactive | Signup proceeds; no referral row; no log |
| Referrer and invitee are the same user | Signup proceeds; no referral row; no log |
| `facades.Orm().Query().Create(&record)` fails | Signup proceeds; warning log with error wrap |
| `/api/invite/validate?code=` called with empty `code` | Returns `{valid: false}`, not an error |
| `/api/invite/validate` DB error | Returns 500 `CodeInternalError` |

The net guarantee: **no signup ever 500s because of the referral path**, and
new users always get `grade = "free"`.

## Testing Strategy

### Automated

- **Backend:** `cd dx-api && go build ./... && go vet ./... && go test -race ./...`
  Existing auth tests continue to pass; they do not assert on `grade`, so
  setting `"free"` is additive.
- **Frontend:** `cd dx-web && npm run lint && npm run build`
  Verifies TypeScript, ESLint, and Next.js compile.

### Manual

1. **Grade backfill:** signup via `/auth/signup` → SQL
   `SELECT grade FROM users WHERE email = '…'` returns `free`.
2. **Valid invite flow:**
   a. Visit `http://localhost/invite/<existingInviteCode>`.
   b. DevTools → Application → Cookies → confirm `ref` exists.
   c. Land on `/auth/signup`; confirm "正在通过好友邀请注册" badge visible.
   d. Complete signup.
   e. SQL `SELECT * FROM user_referrals WHERE referrer_id = '<referrer>'` shows
      a new `pending` row with the new user as `invitee_id`.
3. **Invalid invite flow:**
   a. Visit `http://localhost/invite/garbagecode`.
   b. Confirm no `ref` cookie set.
   c. Land on `/auth/signup`; confirm badge hidden.
   d. Complete signup; confirm no new `user_referrals` row.
4. **Auto-register via signin:** set a valid `ref` cookie manually, then use
   email-code signin with a brand-new email. Confirm a `user_referrals` row
   is created and the new user has `grade = "free"`.
5. **`/hall/invite` copy UX:** click 复制链接, confirm icon swaps for ~2s and
   reverts; clipboard contents are the invite URL.
6. **Signup form cleanup:** confirm "信息补充（可跳过）" row is gone and the
   账号 field follows the code area directly.

## Open Decisions (resolved)

- **Cookie vs. URL query param?** Cookie — already wired in the invite
  redirect page, matches same-origin nginx setup, zero request-struct changes.
- **Backfill existing users?** No — forward-only fix.
- **Fix auto-register path too?** Yes — same funnel as signup.
- **Clear cookie after consumption?** No — natural 7-day expiry is fine.
- **Block signup on referral error?** No — log and continue.

## Files Touched

**dx-api:**
- `app/services/api/auth_service.go` (T1, T3)
- `app/services/api/referral_service.go` (T2, T3)
- `app/http/controllers/api/user_referral_controller.go` (T2)
- `routes/api.go` (T2)

**dx-web:**
- `src/app/(web)/auth/signup/page.tsx` (T6)
- `src/features/web/auth/components/sign-up-form.tsx` (T4, T6)
- `src/features/web/invite/components/invite-content.tsx` (T5)

No new files. No schema changes. No configuration changes.
