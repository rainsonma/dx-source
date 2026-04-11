# Signup Grade & Invite Referral Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix two backend bugs (empty `users.grade`, missing `user_referrals` rows from invite flow) and land three UI polish items on the signup/invite pages.

**Architecture:** Backend reads the existing httpOnly `ref` cookie via `ctx.Request().Cookie("ref")` during `SignUp`/`SignInByEmail`, looks up the referrer by `invite_code`, and writes a `user_referrals` row with `status = "pending"`. A new public `/api/invite/validate` endpoint lets the Next.js invite redirect page set the cookie reliably. Frontend adds a cookie-driven "正在通过好友邀请注册" badge on the signup form and a transient Check icon on the copy button.

**Tech Stack:** Go 1.x + Goravel v1.17.2 + GORM + PostgreSQL (dx-api); Next.js 16 + TypeScript + TailwindCSS v4 + lucide-react (dx-web); nginx same-origin proxy (deploy).

**Spec:** [`docs/superpowers/specs/2026-04-11-signup-grade-referral-fixes-design.md`](../specs/2026-04-11-signup-grade-referral-fixes-design.md)

---

## File Map

**Backend (dx-api):**

| File | Action | Responsibility |
|------|--------|---------------|
| `app/services/api/auth_service.go` | Modify | Set `Grade: consts.UserGradeFree` in `SignUp` + `SignInByEmail` auto-register; invoke `RecordReferralIfPresent` after user creation |
| `app/services/api/referral_service.go` | Modify | Add `ValidateInviteCode(code)` and `RecordReferralIfPresent(ctx, inviteeID)` |
| `app/services/api/referral_service_test.go` | Create | Behavior tests for new referral functions |
| `app/http/controllers/api/user_referral_controller.go` | Modify | Add `ValidateCode(ctx)` public handler |
| `routes/api.go` | Modify | Hoist `userReferralController`, register `GET /api/invite/validate` |

**Frontend (dx-web):**

| File | Action | Responsibility |
|------|--------|---------------|
| `src/features/web/auth/components/sign-up-form.tsx` | Modify | Remove 信息补充 row, accept `hasInviteRef` prop, render badge |
| `src/app/(web)/auth/signup/page.tsx` | Modify | Become async server component, read `ref` cookie, pass prop |
| `src/features/web/invite/components/invite-content.tsx` | Modify | Add `copied` state, swap Copy/Check icon for 2s after click |

No new files on the frontend. No schema changes. No config changes.

---

## Notes for the Implementer

- **Codebase conventions.** The project uses `facades.Orm().Query()` for GORM, `facades.Log().Warningf(...)` for structured logs, `uuid.Must(uuid.NewV7()).String()` for IDs, and the `helpers.Success`/`helpers.Error` envelope for all responses. Do not introduce new patterns.
- **Tests follow the project style.** Backend service tests in this project are deliberately lightweight — see `app/services/api/group_service_test.go` for existing `NotNil(Func)` smoke tests. Where pure logic allows (e.g., an empty-string short-circuit), add real behavior tests. Do not introduce a DB test harness — that's out of scope.
- **No frontend test framework.** dx-web does not have a configured unit test runner. Rely on `npm run lint` + `npm run build` + manual smoke tests.
- **Format/vet on every Go change.** Run `gofmt -w` and `go vet ./...` before committing. The user's settings already run `goimports` automatically on edit via Claude Code hooks, but verify anyway.
- **Never use `--no-verify` on commits.** If a hook fails, fix the root cause.
- **Stage files explicitly** (`git add <path>`), never `git add .` or `git add -A`.

---

## Task 1: Fix empty `users.grade` on signup and auto-register

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go` (around lines 73-80 and 123-130)

**Why no test:** This task only changes a struct literal. No new function, no new branch. Existing `group_service_test.go`-style smoke tests wouldn't prove anything; the real verification is `go build ./...` + a manual smoke test of one signup.

- [ ] **Step 1: Add `consts` import**

Open `dx-api/app/services/api/auth_service.go`. The current imports block ends with `"dx-api/app/models"`. Add `"dx-api/app/consts"` as the first project-local import so the block reads:

```go
import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)
```

- [ ] **Step 2: Set `Grade` in `SignUp`**

In `SignUp` at lines 73-80, change the `user := models.User{...}` literal from:

```go
user := models.User{
	ID:         uuid.Must(uuid.NewV7()).String(),
	Username:   username,
	Email:      &emailStr,
	Password:   hashedPassword,
	IsActive:   true,
	InviteCode: helpers.GenerateInviteCode(8),
}
```

to:

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

- [ ] **Step 3: Set `Grade` in `SignInByEmail` auto-register**

In `SignInByEmail` at lines 123-130, change the `user = models.User{...}` literal from:

```go
user = models.User{
	ID:         uuid.Must(uuid.NewV7()).String(),
	Username:   username,
	Email:      &emailStr,
	Password:   hashedPw,
	IsActive:   true,
	InviteCode: helpers.GenerateInviteCode(8),
}
```

to:

```go
user = models.User{
	ID:         uuid.Must(uuid.NewV7()).String(),
	Grade:      consts.UserGradeFree,
	Username:   username,
	Email:      &emailStr,
	Password:   hashedPw,
	IsActive:   true,
	InviteCode: helpers.GenerateInviteCode(8),
}
```

- [ ] **Step 4: Build & vet**

Run from `dx-api/`:

```bash
go build ./...
go vet ./...
```

Expected: no output, exit code 0.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/auth_service.go
git commit -m "fix(auth): set default grade to free on user creation"
```

---

## Task 2: Add `ValidateInviteCode` service function (TDD)

**Files:**
- Modify: `dx-api/app/services/api/referral_service.go`
- Create: `dx-api/app/services/api/referral_service_test.go`

- [ ] **Step 1: Write the failing empty-code test**

Create `dx-api/app/services/api/referral_service_test.go` with:

```go
package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInviteCodeEmptyReturnsFalse(t *testing.T) {
	ok, err := ValidateInviteCode("")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestValidateInviteCodeFunctionExists(t *testing.T) {
	assert.NotNil(t, ValidateInviteCode)
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run from `dx-api/`:

```bash
go test ./app/services/api/ -run TestValidateInviteCode -v
```

Expected: compile error `undefined: ValidateInviteCode` or similar.

- [ ] **Step 3: Implement `ValidateInviteCode`**

Append to the end of `dx-api/app/services/api/referral_service.go` (after `derefStr`):

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

- [ ] **Step 4: Run the test and verify it passes**

Run from `dx-api/`:

```bash
go test ./app/services/api/ -run TestValidateInviteCode -v
```

Expected:

```
=== RUN   TestValidateInviteCodeEmptyReturnsFalse
--- PASS: TestValidateInviteCodeEmptyReturnsFalse
=== RUN   TestValidateInviteCodeFunctionExists
--- PASS: TestValidateInviteCodeFunctionExists
PASS
```

- [ ] **Step 5: Build & vet**

```bash
go build ./...
go vet ./...
```

Expected: no output, exit code 0.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/referral_service.go dx-api/app/services/api/referral_service_test.go
git commit -m "feat(referral): add ValidateInviteCode service"
```

---

## Task 3: Wire `/api/invite/validate` public endpoint

**Files:**
- Modify: `dx-api/app/http/controllers/api/user_referral_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add `ValidateCode` controller method**

Append to `dx-api/app/http/controllers/api/user_referral_controller.go` after `GetReferrals`:

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

- [ ] **Step 2: Hoist `userReferralController` in routes/api.go**

Open `dx-api/routes/api.go`. Currently:

- Line 17-21 declares `authController`, `emailController`, `uploadController` at the top of the `/api` group.
- Line 179 declares `userReferralController := apicontrollers.NewUserReferralController()` inside the protected middleware block.

Move the declaration so it sits alongside `authController` near the top of the `/api` group. Replace the block at lines 17-21:

```go
		authController := apicontrollers.NewAuthController()
		emailController := apicontrollers.NewEmailController()

		uploadController := apicontrollers.NewUploadController()
```

with:

```go
		authController := apicontrollers.NewAuthController()
		emailController := apicontrollers.NewEmailController()
		userReferralController := apicontrollers.NewUserReferralController()

		uploadController := apicontrollers.NewUploadController()
```

Then delete the duplicate declaration at line 179. Replace:

```go
			// Invite & Referrals
			userReferralController := apicontrollers.NewUserReferralController()
			protected.Get("/invite", userReferralController.GetInviteData)
			protected.Get("/referrals", userReferralController.GetReferrals)
```

with:

```go
			// Invite & Referrals
			protected.Get("/invite", userReferralController.GetInviteData)
			protected.Get("/referrals", userReferralController.GetReferrals)
```

- [ ] **Step 3: Register the public `/api/invite/validate` route**

Still in `dx-api/routes/api.go`, add the new public route inside the `/api` prefix group but outside the protected `Middleware(middleware.JwtAuth())` block. Put it near the public auth/email routes (around line 71, after the `/email` group block). Add:

```go
		// Public invite code validation (no auth required)
		router.Get("/invite/validate", userReferralController.ValidateCode)
```

- [ ] **Step 4: Build & vet**

```bash
cd dx-api
go build ./...
go vet ./...
```

Expected: no output, exit code 0.

- [ ] **Step 5: Run existing tests**

```bash
go test -race ./...
```

Expected: all tests pass (this step catches any collateral breakage in existing code).

- [ ] **Step 6: Manual smoke test (human-run, not automated)**

With the server running via `air` or `go run .`:

```bash
# Empty code → valid=false
curl -s "http://localhost:3001/api/invite/validate?code=" | head -c 200

# Garbage code → valid=false
curl -s "http://localhost:3001/api/invite/validate?code=garbagecode" | head -c 200

# Any valid invite_code from the users table → valid=true
# First query one: psql -c "SELECT invite_code FROM users LIMIT 1;"
curl -s "http://localhost:3001/api/invite/validate?code=<realCode>" | head -c 200
```

Expected all three: HTTP 200 with `{"code":0,"message":"ok","data":{"valid":<bool>}}`.

If you can't run the server right now, note this step as pending and mark it complete when you verify post-merge.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/http/controllers/api/user_referral_controller.go dx-api/routes/api.go
git commit -m "feat(referral): add public /api/invite/validate endpoint"
```

---

## Task 4: Add `RecordReferralIfPresent` service function

**Files:**
- Modify: `dx-api/app/services/api/referral_service.go`
- Modify: `dx-api/app/services/api/referral_service_test.go`

- [ ] **Step 1: Write the failing function-exists test**

Append to `dx-api/app/services/api/referral_service_test.go`:

```go
func TestRecordReferralIfPresentFunctionExists(t *testing.T) {
	assert.NotNil(t, RecordReferralIfPresent)
}
```

- [ ] **Step 2: Run the test and verify it fails**

```bash
cd dx-api
go test ./app/services/api/ -run TestRecordReferralIfPresentFunctionExists -v
```

Expected: compile error `undefined: RecordReferralIfPresent`.

- [ ] **Step 3: Add required imports to `referral_service.go`**

Open `dx-api/app/services/api/referral_service.go`. The current import block is:

```go
import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)
```

Replace with:

```go
import (
	"fmt"

	"github.com/google/uuid"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/models"
)
```

- [ ] **Step 4: Implement `RecordReferralIfPresent`**

Append to `dx-api/app/services/api/referral_service.go` (after `ValidateInviteCode`):

```go
// RecordReferralIfPresent creates a user_referrals row when the request carries a
// `ref` cookie matching an active user. It never blocks signup: any lookup or
// create failure is wrapped and returned so the caller can log it, but callers
// must not propagate the error to the API response.
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
		return nil
	}
	if referrer.ID == inviteeID {
		return nil
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

- [ ] **Step 5: Run the test and verify it passes**

```bash
cd dx-api
go test ./app/services/api/ -run TestRecordReferralIfPresentFunctionExists -v
```

Expected: `PASS`.

- [ ] **Step 6: Build & vet**

```bash
go build ./...
go vet ./...
```

Expected: no output, exit code 0.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/referral_service.go dx-api/app/services/api/referral_service_test.go
git commit -m "feat(referral): add RecordReferralIfPresent service"
```

---

## Task 5: Wire `RecordReferralIfPresent` into signup + auto-register

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go` (two call sites)

- [ ] **Step 1: Call `RecordReferralIfPresent` after `SignUp` user creation**

Open `dx-api/app/services/api/auth_service.go`. In `SignUp`, the current block at lines 82-84 is:

```go
	if err := facades.Orm().Query().Create(&user); err != nil {
		return "", nil, fmt.Errorf("failed to create user: %w", err)
	}
```

Directly after that block (before the `token, err := issueSession(...)` call), add:

```go
	if refErr := RecordReferralIfPresent(ctx, user.ID); refErr != nil {
		facades.Log().Warningf("record referral failed: %v", refErr)
	}
```

- [ ] **Step 2: Call `RecordReferralIfPresent` after `SignInByEmail` auto-register**

In `SignInByEmail`, the current auto-register branch at lines 132-134 is:

```go
		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	}
```

Change to:

```go
		if createErr := facades.Orm().Query().Create(&user); createErr != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", createErr)
		}

		if refErr := RecordReferralIfPresent(ctx, user.ID); refErr != nil {
			facades.Log().Warningf("record referral failed: %v", refErr)
		}
	}
```

The `RecordReferralIfPresent` call must sit **inside** the `if err != nil || user.ID == ""` block (the auto-register branch), not after it — only auto-registered users should trigger referral creation, not existing users who are just signing in.

- [ ] **Step 3: Build & vet**

```bash
cd dx-api
go build ./...
go vet ./...
```

Expected: no output, exit code 0.

- [ ] **Step 4: Run all tests**

```bash
go test -race ./...
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/auth_service.go
git commit -m "feat(auth): record referral on signup and auto-register"
```

---

## Task 6: Remove "信息补充（可跳过）" row from signup form

**Files:**
- Modify: `dx-web/src/features/web/auth/components/sign-up-form.tsx`

- [ ] **Step 1: Delete the spacer + section header block**

Open `dx-web/src/features/web/auth/components/sign-up-form.tsx`. Find the block at lines 134-142:

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

Delete the entire block (all 9 lines including the blank line). The 账号 input block should now sit directly after the 验证码 "Code sent hint" conditional block.

- [ ] **Step 2: Remove unused `Info` import**

Line 4 currently reads:

```tsx
import { Info, Eye, EyeOff, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";
```

Change to:

```tsx
import { Eye, EyeOff, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";
```

- [ ] **Step 3: Run lint**

```bash
cd dx-web
npm run lint
```

Expected: exit code 0, no errors. (If lint flags any unused imports you missed, address them before continuing.)

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/auth/components/sign-up-form.tsx
git commit -m "refactor(signup): remove 信息补充 section header"
```

---

## Task 7: Swap copy → check icon on `/hall/invite`

**Files:**
- Modify: `dx-web/src/features/web/invite/components/invite-content.tsx`

- [ ] **Step 1: Add `Check` to the lucide-react import**

Open `dx-web/src/features/web/invite/components/invite-content.tsx`. Lines 4-13 currently read:

```tsx
import {
  Copy,
  Users,
  DollarSign,
  UserCheck,
  TrendingUp,
  ScrollText,
  Info,
  Share2,
} from "lucide-react";
```

Change to:

```tsx
import {
  Copy,
  Check,
  Users,
  DollarSign,
  UserCheck,
  TrendingUp,
  ScrollText,
  Info,
  Share2,
} from "lucide-react";
```

- [ ] **Step 2: Add `copied` state**

Line 37 currently reads:

```tsx
  const [shareOpen, setShareOpen] = useState(false);
```

Add a new line immediately after:

```tsx
  const [shareOpen, setShareOpen] = useState(false);
  const [copied, setCopied] = useState(false);
```

- [ ] **Step 3: Update `handleCopy` to flip state for 2s**

Lines 47-53 currently read:

```tsx
  /** Copy invite URL to clipboard */
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(inviteUrl);
    } catch {
      // Silently fail if clipboard access is denied
    }
  };
```

Change to:

```tsx
  /** Copy invite URL to clipboard and flash a check icon for 2 seconds */
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

- [ ] **Step 4: Swap the icon conditionally**

Lines 83-90 currently read:

```tsx
            <button
              type="button"
              onClick={handleCopy}
              className="flex items-center justify-center gap-1.5 rounded-[10px] bg-white px-4 py-2 text-[13px] font-semibold text-teal-700"
            >
              <Copy className="h-3.5 w-3.5" />
              复制链接
            </button>
```

Change to:

```tsx
            <button
              type="button"
              onClick={handleCopy}
              className="flex items-center justify-center gap-1.5 rounded-[10px] bg-white px-4 py-2 text-[13px] font-semibold text-teal-700"
            >
              {copied ? (
                <Check className="h-3.5 w-3.5" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
              复制链接
            </button>
```

- [ ] **Step 5: Run lint**

```bash
cd dx-web
npm run lint
```

Expected: exit code 0, no errors.

- [ ] **Step 6: Manual smoke test**

With the dev server running (`npm run dev`), log in and visit `/hall/invite`. Click 复制链接 and confirm:
- The icon switches to a check mark immediately.
- The icon reverts to the copy glyph ~2 seconds later.
- The clipboard contains the invite URL (paste into a new tab to verify).

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/invite/components/invite-content.tsx
git commit -m "feat(invite): show check icon for 2s after copy link"
```

---

## Task 8: Show "正在通过好友邀请注册" badge on signup form

**Files:**
- Modify: `dx-web/src/app/(web)/auth/signup/page.tsx`
- Modify: `dx-web/src/features/web/auth/components/sign-up-form.tsx`

- [ ] **Step 1: Make the signup page async and read the `ref` cookie**

Replace the entire contents of `dx-web/src/app/(web)/auth/signup/page.tsx` with:

```tsx
import { cookies } from "next/headers";

import { SignUpForm } from "@/features/web/auth/components/sign-up-form";

export default async function SignUpPage() {
  const cookieStore = await cookies();
  const hasInviteRef = Boolean(cookieStore.get("ref")?.value);
  return <SignUpForm hasInviteRef={hasInviteRef} />;
}
```

- [ ] **Step 2: Add `Gift` to the lucide-react import**

Open `dx-web/src/features/web/auth/components/sign-up-form.tsx`. After Task 6 the import on line 4 should read:

```tsx
import { Eye, EyeOff, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";
```

Change to:

```tsx
import { Eye, EyeOff, Gift, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";
```

- [ ] **Step 3: Accept the `hasInviteRef` prop**

The component currently starts at line 9 with:

```tsx
export function SignUpForm() {
```

Change to:

```tsx
type SignUpFormProps = {
  hasInviteRef?: boolean;
};

export function SignUpForm({ hasInviteRef }: SignUpFormProps) {
```

- [ ] **Step 4: Render the badge inside the header block**

The header block currently at lines 38-46 reads:

```tsx
      {/* Header */}
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-[32px] font-extrabold text-slate-900">
          创建斗学专属账号
        </h1>
        <p className="text-sm text-slate-400">
          进入斗学英语游戏世界，开启英语学习冒险之旅
        </p>
      </div>
```

Change to:

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

- [ ] **Step 5: Run lint**

```bash
cd dx-web
npm run lint
```

Expected: exit code 0, no errors.

- [ ] **Step 6: Run production build**

```bash
cd dx-web
npm run build
```

Expected: build succeeds with no type errors. Next.js 16 is strict about async server components — this step catches any issues with the `await cookies()` call or the new prop type.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/app/\(web\)/auth/signup/page.tsx dx-web/src/features/web/auth/components/sign-up-form.tsx
git commit -m "feat(signup): show invite referral badge when ref cookie present"
```

---

## Task 9: End-to-end verification pass

No code changes. This task gates the PR.

- [ ] **Step 1: Backend — full build, vet, and race-detector tests**

```bash
cd dx-api
go build ./...
go vet ./...
go test -race ./...
```

Expected: all three exit 0. Note any new warnings or failures — a failure here means a previous task committed broken code.

- [ ] **Step 2: Frontend — lint and production build**

```bash
cd dx-web
npm run lint
npm run build
```

Expected: both exit 0 with no errors.

- [ ] **Step 3: Manual smoke test — grade field**

With the stack running (`docker compose -f deploy/docker-compose.dev.yml up` or equivalent local setup):

1. Visit `http://localhost/auth/signup`.
2. Complete signup with a fresh email.
3. Run:
   ```bash
   psql -h localhost -U postgres -d dxdb -c \
     "SELECT email, grade FROM users WHERE email='<the-email-you-used>';"
   ```
4. Confirm `grade = 'free'`.

- [ ] **Step 4: Manual smoke test — happy-path invite flow**

1. With any existing user's invite_code (query the DB): `SELECT invite_code FROM users WHERE username='<someone>';`
2. Visit `http://localhost/invite/<that-code>` in a private/incognito window.
3. Expect redirect to `/auth/signup`.
4. DevTools → Application → Cookies → confirm `ref` is set to the code.
5. Confirm the signup header shows the **"正在通过好友邀请注册"** badge with a gift icon.
6. Complete signup with a fresh email.
7. Run:
   ```bash
   psql -h localhost -U postgres -d dxdb -c \
     "SELECT referrer_id, invitee_id, status FROM user_referrals ORDER BY created_at DESC LIMIT 1;"
   ```
8. Confirm a row exists with `status='pending'` and the correct referrer/invitee IDs.

- [ ] **Step 5: Manual smoke test — invalid invite flow**

1. In a fresh incognito window, visit `http://localhost/invite/garbagecodexyz`.
2. Expect redirect to `/auth/signup` and NO badge shown.
3. DevTools → Application → Cookies → confirm `ref` is NOT set.
4. Complete signup with a fresh email.
5. Confirm no new `user_referrals` row was added.

- [ ] **Step 6: Manual smoke test — signup form cleanup**

Visit `/auth/signup` in a non-invite session. Confirm:
- "信息补充（可跳过）" label and icon are GONE.
- The 账号 input sits directly beneath the 验证码 area.
- The form still submits correctly with and without username/password.

- [ ] **Step 7: Manual smoke test — copy icon on /hall/invite**

1. Log into any account.
2. Visit `/hall/invite`.
3. Click the 复制链接 button.
4. Confirm icon swaps to a check for ~2 seconds, then reverts.
5. Paste into a new browser tab — URL should match the `inviteUrl` displayed in the banner.

- [ ] **Step 8: Manual smoke test — auto-register with invite cookie**

1. Visit `http://localhost/invite/<validCode>` in a fresh incognito window.
2. Navigate to `/auth/signin` instead of `/auth/signup`.
3. Sign in with a brand-new email + the 6-digit code flow.
4. Confirm the new user has `grade = 'free'` AND a `user_referrals` row was created with the expected referrer.

- [ ] **Step 9: Regression check — existing flows still work**

Run through these previously-working flows to confirm nothing broke:
- Signin with an existing account (username + password) → redirects to `/hall`, no referral row, no log warning.
- Signin with an existing account via email + code → same.
- Visit `/hall/invite` → page loads, stats render, table populates.
- Logout from anywhere → cookie cleared, redirect to signin.

If any regression surfaces, open a fresh debugging session; do not try to patch on top of this branch in a rush.

---

## Done criteria

- All 9 tasks checked off
- No lint or vet errors anywhere
- `go test -race ./...` passes
- All manual smoke tests green
- Nine commits on the branch, each self-describing
- Spec file reference is unchanged and still accurate

---

## Rollback notes

If a production issue surfaces post-merge, each task is its own commit, so individual reverts are clean:

- `git revert <commit-sha>` for any single task
- Reverting Task 3 (the route registration) by itself removes the public endpoint but leaves the service function in place — safe
- Reverting Task 5 by itself leaves `RecordReferralIfPresent` defined but uncalled — safe
- Reverting Task 1 resets grade to empty on new users — correct rollback
