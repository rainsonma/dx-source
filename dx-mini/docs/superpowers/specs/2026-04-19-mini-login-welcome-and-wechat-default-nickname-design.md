# dx-mini login welcome heading + dx-api WeChat default nickname

**Date:** 2026-04-19
**Scope:**
- `dx-mini/miniprogram/pages/login/login.{wxml,wxss}` — layout tweak only
- `dx-api/app/helpers/random.go` + `dx-api/app/helpers/random_test.go` (new) + `dx-api/app/services/api/wechat_auth_service.go` — nickname default
**Goal:** Introduce a warm `欢迎来到斗学` greeting above the existing tagline, glue the greeting group directly above the 微信一键登录 button (logo stays alone in the upper zone), and auto-assign a `斗友_XXXXXX` default nickname to every new user created by the WeChat Mini Program login flow.
**Non-goal:** Any change to email/password registration, existing users' nicknames, the `wx.login → /api/auth/wechat-mini` contract, the `handleLogin` TS logic, the gradient background, or any other page.

---

## Background

The login landing screen was redesigned one commit ago (`e8cc516`). That redesign established a 5-stop pastel gradient, a 110rpx flat-teal brand mark, a two-line slate tagline, and a staggered `hero-rise` reveal. It did not include a greeting heading, and the brand mark + tagline currently sit together as a glued block around 38% of the mid-zone, leaving a large vertical gap above the CTA.

Two observations motivate this change:

1. The screen opens with a brand mark followed immediately by feature-pitch copy (`多种学习模式 · AI 定制内容 · …`). There is no human greeting. First-time WeChat users arrive on a screen that sells features but never says "welcome."
2. On the dx-api side, the WeChat auto-registration path at `services/api/wechat_auth_service.go:101-109` creates a `models.User{…}` without setting `Nickname`, so every downstream surface that displays a user name falls back to the auto-generated `wx_<openid-prefix>` username. That is technically fine (fallback code exists in 10+ services) but it leaks an internal identifier into the UI.

A single coordinated change — greeting on the landing, default nickname on registration — fixes both problems without touching anything else.

---

## Requirements

### Frontend (dx-mini landing)

1. A new centered heading `欢迎来到斗学` renders immediately above the existing two-line tagline.
2. The welcome heading + tagline together form a single visual group pinned **just above the 微信一键登录 button** — not floating in the middle of the page.
3. The `斗学` logo remains the only element centered in the upper hero zone; the welcome+tagline group leaves it.
4. The logo stays at its current vertical position (~37.5% of the mid-zone between the status bar and the CTA) — rhythm unchanged.
5. Motion rhythm remains the existing 3-beat stagger: logo `0ms` → welcome+tagline group `120ms` → button `240ms`. Total first-paint unchanged at 720ms.
6. Welcome heading typography: `40rpx`, `font-weight: 600`, color `#334155` (slate-700), `letter-spacing: 2rpx`, center-aligned.
7. Tagline typography unchanged (`28rpx`, `#475569`, two lines, 8rpx gap, center-aligned).
8. No new image, font, or npm asset. No `?.` / `??` in WXML. No change to `login.ts` or `login.json`.
9. No new `tsc --noEmit` errors introduced (the pass currently produces the well-known Component-`this` warnings only; this change should not add any).

### Backend (dx-api WeChat registration)

1. Every new `models.User` row created inside `WechatMiniSignIn` gets `Nickname` set to `斗友_XXXXXX`, where `XXXXXX` is 6 crypto-random digits.
2. The value is generated at registration time and persisted; it is NOT computed on every read.
3. Existing users (WeChat or otherwise) are **not** modified. No backfill, no migration.
4. `auth_service.go` (email/password registration paths) is **not** touched.
5. The downstream nil-fallback (`if user.Nickname != nil && *user.Nickname != ""`) continues to work for legacy users who predate this change.
6. Added code is formatted via `gofmt`/`goimports`, passes `go vet`, and does not introduce staticcheck findings. `go test -race ./app/helpers/...` passes.

---

## Design

### §1 Landing layout — WXML (`dx-mini/miniprogram/pages/login/login.wxml`)

```xml
<view class="login-page">
  <view class="login-bg" />
  <view class="login-content" style="padding-top: {{statusBarHeight}}px;">
    <view class="capsule-spacer" />
    <view class="spacer-top" />
    <view class="brand">
      <text class="brand-logo">斗学</text>
    </view>
    <view class="spacer-bottom" />
    <view class="greeting">
      <text class="greeting-title">欢迎来到斗学</text>
      <view class="brand-tagline">
        <text class="tagline-line">多种学习模式 · AI 定制内容 · 和朋友一起闯关</text>
        <text class="tagline-line">每天 10 分钟，英语悄悄流利了</text>
      </view>
    </view>
    <view class="cta-wrap">
      <van-button
        type="primary"
        color="#0d9488"
        block
        round
        loading="{{loading}}"
        bind:click="handleLogin"
        custom-style="height:96rpx;font-size:32rpx;font-weight:600;"
      >微信一键登录</van-button>
    </view>
  </view>
</view>
```

**What changed vs. the current file:**
- `.brand-tagline` is moved out of `.brand` and into a new sibling `.greeting`.
- `.brand` now contains only `.brand-logo`.
- `.greeting` is inserted between `.spacer-bottom` and `.cta-wrap`, placing welcome+tagline directly above the button.

**Why this DOM structure:**
The spacer-top/spacer-bottom flex ratio (0.6 / 1.0) is what determines where the logo sits. As long as `.brand` is the element being pushed by those two spacers, its vertical position is preserved. The welcome+tagline group lives outside the spacer chain, anchored to the CTA — so it doesn't displace the logo.

### §2 Landing layout — WXSS (`dx-mini/miniprogram/pages/login/login.wxss`)

**Modify** `.brand-tagline` — drop the outer padding and the animation (both move to `.greeting`):

```css
.brand-tagline {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
}
```

**Add** the new `.greeting` + `.greeting-title` rules anywhere after the `.brand-*` block (before or after `.cta-wrap` — both are fine, but putting them next to each other reads cleanest):

```css
.greeting {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
  padding: 0 60rpx 40rpx;
  box-sizing: border-box;
  animation: hero-rise 480ms ease-out 120ms both;
}

.greeting-title {
  font-size: 40rpx;
  font-weight: 600;
  color: #334155;
  line-height: 1.4;
  letter-spacing: 2rpx;
  text-align: center;
}
```

**Unchanged:** `page`, `.login-page`, `.login-bg`, `.login-bg::after`, `.login-content`, `.capsule-spacer`, `.spacer-top`, `.spacer-bottom`, `.brand`, `.brand-logo`, `.tagline-line`, `.cta-wrap`, `@keyframes hero-rise`.

**Why the animation moves to `.greeting`:** The CSS spec lets a parent animation sequence children via visibility/opacity on the group. By animating `.greeting` with `hero-rise 480ms ease-out 120ms both`, welcome + tagline fade/slide in together as one unit — which matches the "single visual group" requirement and keeps the 3-beat cadence.

**Color justification for `#334155`:** One step darker than the tagline's `#475569`. Over the palest gradient stop (`#ccfbf1`), contrast ratio is 9.2:1; over the pink 75% band (`#fce7f3`), contrast is 6.0:1. Both well above AA small-text (4.5:1) and AA large-text (3.0:1).

### §3 Backend default nickname — helper (`dx-api/app/helpers/random.go`)

Append:

```go
// GenerateDefaultNickname returns a default user nickname of the form
// "斗友_XXXXXX" where XXXXXX is 6 crypto-random digits. Uniqueness is not
// guaranteed by design: the users.nickname column has no unique constraint,
// and downstream display code accepts duplicate nicknames.
func GenerateDefaultNickname() string {
    return "斗友_" + GenerateCode(6)
}
```

Placement: immediately after the existing `GenerateInviteCode` function, before `randomString`.

**Why a thin wrapper, not inline:** Three reasons — (1) the string `斗友_` is a product-level constant that could migrate to a `consts/` file later without disturbing callers; (2) centralizing the definition makes the unit test trivial; (3) CLAUDE.md's house rule is "always encapsulate logic in functions or helpers when possible."

### §4 Backend default nickname — service (`dx-api/app/services/api/wechat_auth_service.go`)

Modify the user-creation block at lines 100-109. Insert two lines:

```go
openID := session.OpenID
nickname := helpers.GenerateDefaultNickname()   // new
user = models.User{
    ID:         uuid.Must(uuid.NewV7()).String(),
    Grade:      consts.UserGradeFree,
    Username:   username,
    Nickname:   &nickname,                       // new
    Password:   hashedPw,
    IsActive:   true,
    InviteCode: helpers.GenerateInviteCode(8),
    OpenID:     &openID,
}
```

Everything else in the file is unchanged — including the subsequent UnionID block, the Create call, the referral call, and `issueSession`.

**Why address-of a local variable:** The `Nickname` field is `*string`. Taking `&nickname` of a function-local variable is the idiomatic pattern already used in this same function for `openID` (line 100) and `unionID` (line 112).

### §5 Tests

**New file** — `dx-api/app/helpers/random_test.go`:

```go
package helpers

import (
    "strings"
    "testing"
    "unicode/utf8"
)

func TestGenerateCode(t *testing.T) {
    tests := []int{1, 4, 6, 16}
    for _, length := range tests {
        got := GenerateCode(length)
        if len(got) != length {
            t.Errorf("GenerateCode(%d) len = %d, want %d", length, len(got), length)
        }
        for _, r := range got {
            if r < '0' || r > '9' {
                t.Errorf("GenerateCode(%d) = %q, contains non-digit %q", length, got, r)
                break
            }
        }
    }
}

func TestGenerateInviteCode(t *testing.T) {
    got := GenerateInviteCode(8)
    if len(got) != 8 {
        t.Errorf("GenerateInviteCode(8) len = %d, want 8", len(got))
    }
    for _, r := range got {
        ok := (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
        if !ok {
            t.Errorf("GenerateInviteCode(8) = %q, contains non-alphanumeric %q", got, r)
            break
        }
    }
}

func TestGenerateDefaultNickname(t *testing.T) {
    got := GenerateDefaultNickname()

    if !strings.HasPrefix(got, "斗友_") {
        t.Fatalf("GenerateDefaultNickname() = %q, want prefix %q", got, "斗友_")
    }

    suffix := strings.TrimPrefix(got, "斗友_")
    if utf8.RuneCountInString(suffix) != 6 {
        t.Errorf("suffix rune count = %d, want 6 (got %q)", utf8.RuneCountInString(suffix), suffix)
    }
    for _, r := range suffix {
        if r < '0' || r > '9' {
            t.Errorf("suffix %q contains non-digit %q", suffix, r)
            break
        }
    }

    // Sanity: two back-to-back calls should differ (not a strict uniqueness
    // claim — this catches a broken PRNG that returns constant output).
    a, b := GenerateDefaultNickname(), GenerateDefaultNickname()
    if a == b {
        t.Errorf("two consecutive calls returned the same nickname %q", a)
    }
}
```

No changes to `wechat_auth_service_test.go`. The file currently covers `fetchWechatSession` and `generateWxUsername`; adding a test for `WechatMiniSignIn` would require bootstrapping Goravel's `facades.Orm()`, which is out of scope. Correctness of the nickname default is fully covered by `TestGenerateDefaultNickname` plus the type-checker + compile step.

---

## Files touched

| File | Change |
|---|---|
| `dx-mini/miniprogram/pages/login/login.wxml` | Move `.brand-tagline` out of `.brand`, wrap it (with new `.greeting-title`) inside a new `.greeting` sibling placed between `.spacer-bottom` and `.cta-wrap`. |
| `dx-mini/miniprogram/pages/login/login.wxss` | Drop `padding` + `animation` from `.brand-tagline`. Add `.greeting` and `.greeting-title` rules. |
| `dx-api/app/helpers/random.go` | Append `GenerateDefaultNickname()`. |
| `dx-api/app/helpers/random_test.go` | New file — table-driven tests for `GenerateCode`, `GenerateInviteCode`, `GenerateDefaultNickname`. |
| `dx-api/app/services/api/wechat_auth_service.go` | Insert `nickname := helpers.GenerateDefaultNickname()` just before the struct literal at line 101, and add `Nickname: &nickname` inside the struct. Both lines live inside the new-user branch that starts at line 84. |

**Not touched:** `login.ts`, `login.json`, `app.wxss`, `app.ts`, `app.json`, `project.config.json`, `tsconfig.json`, any other page, any other service, any controller, any middleware, any migration, any seeder, `auth_service.go`, any model. No new npm or Go package dependencies.

---

## Risks & mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Nickname collision between two users | very low (1M slots, ~0 overlap at current scale) | Column has no unique constraint; downstream code tolerates duplicates. Accepted. |
| Existing WeChat users lose nickname display | nil | Only the register branch is modified; existing users are untouched. Fallback path unchanged. |
| WXML lint / Skyline renderer rejects new `.greeting` wrapper | nil | Simple `<view>` + `<text>`; same primitives used throughout the file. |
| Animation regression because `.brand-tagline` loses its `animation` property | nil | Tagline's parent `.greeting` now carries the animation; children inherit the opacity/transform transition through the group. |
| Logo shifts vertical position | nil | `.brand` still sits inside the same spacer-top/spacer-bottom flex chain with the same ratios. The welcome+tagline block is outside the chain. |
| `GenerateDefaultNickname` called with a stale package import | low | `helpers` is already imported in `wechat_auth_service.go` (line 15); no new import needed. |
| Goravel ORM rejects the `Nickname` write | nil | Column is nullable TEXT, no constraints beyond the non-unique index. `*string` assignment is how other optional string fields (`OpenID`, `UnionID`) are already set in the same struct literal. |
| `斗友` triggers WeChat content filters | low | The prefix is generic/positive Chinese ("fellow student" idiom). No WeChat content-policy conflict. If a future issue surfaces, the prefix is isolated in one helper function and trivial to swap. |

---

## Verification (acceptance criteria)

**dx-api:**

1. `go build ./...` from `dx-api/` succeeds with no errors.
2. `go vet ./...` reports zero findings on the changed files.
3. `gofmt -l app/helpers/random.go app/helpers/random_test.go app/services/api/wechat_auth_service.go` prints nothing.
4. `go test -race ./app/helpers/...` passes — all three tests green.
5. `go test -race ./app/services/api/...` passes — existing tests untouched.
6. Running dx-api locally and hitting `POST /api/auth/wechat-mini` with a fresh OpenID produces a user row whose `nickname` column matches `^斗友_[0-9]{6}$` in PostgreSQL.

**dx-mini:**

7. In WeChat DevTools simulator (iPhone 14 Pro, iPhone SE, Pixel 6), opening `pages/login/login` shows:
   - Logo `斗学` in teal at ~37.5% of the mid-zone (same vertical position as before the change).
   - Welcome heading `欢迎来到斗学` in slate-700 `#334155` at 40rpx, immediately followed by the two-line tagline, both center-aligned.
   - Welcome + tagline group sits directly above the 微信一键登录 button with a ~40rpx visual gap between the tagline's last line and the button's top edge.
   - Three-beat hero-rise animation: logo first, welcome+tagline group second (as one unit at 120ms), button third at 240ms.
8. iPhone SE (320×568) — no content overlap; the welcome+tagline group compresses gracefully.
9. Real device verification via 预览 → 小程序助手 (per the `project_wechat_devtools_realdevice_bug.md` note, 真机调试 is broken on the current DevTools build; use 预览).
10. Completing a WeChat login on a fresh account shows the new nickname `斗友_XXXXXX` in the `/me` profile screen (or wherever the profile nickname is rendered).
11. No `?.` or `??` operators introduced. No new `console.log`. No new files in dx-mini outside the two listed (`login.wxml`, `login.wxss`).
