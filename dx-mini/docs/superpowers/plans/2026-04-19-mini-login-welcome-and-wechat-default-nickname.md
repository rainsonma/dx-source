# dx-mini login welcome heading + dx-api WeChat default nickname — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `欢迎来到斗学` greeting above the tagline on the dx-mini login landing (logo stays alone in the upper hero zone, welcome+tagline group glued above the CTA), and auto-assign a `斗友_XXXXXX` default nickname to every new user registered via the WeChat Mini Program flow.

**Architecture:** Two coordinated changes driven by the WeChat registration journey. Backend: one new 2-line helper + one 2-line edit in `WechatMiniSignIn`, zero schema change. Frontend: DOM restructure (move `.brand-tagline` into a new `.greeting` wrapper with a new `.greeting-title` heading) plus CSS move (animation migrates from tagline to its new parent group). No migrations, no new dependencies, no touch to existing user rows.

**Tech Stack:**
- dx-api — Go 1.21+, Goravel, GORM, `crypto/rand` (already used), standard `testing` with race detector
- dx-mini — WeChat Mini Program (glass-easel + Skyline renderer), Vant Weapp 1.11.x, TypeScript strict

---

## File structure

```
dx-api/app/helpers/
├── random.go                       # MODIFY — append GenerateDefaultNickname()
└── random_test.go                  # CREATE — tests for the three random helpers

dx-api/app/services/api/
└── wechat_auth_service.go          # MODIFY — insert 2 lines inside the new-user branch

dx-mini/miniprogram/pages/login/
├── login.wxml                      # MODIFY — restructure: logo alone, new .greeting above CTA
└── login.wxss                      # MODIFY — slim .brand-tagline, add .greeting + .greeting-title
```

---

## Spec reference

This plan implements `dx-mini/docs/superpowers/specs/2026-04-19-mini-login-welcome-and-wechat-default-nickname-design.md` (commit `1f0729e`). Every task below maps to a concrete requirement in that spec.

---

## Task 1 — Backend: `GenerateDefaultNickname` helper with tests (TDD)

**Files:**
- Create: `dx-api/app/helpers/random_test.go`
- Modify: `dx-api/app/helpers/random.go`

**Why a new test file:** The `helpers` package currently has zero test coverage. Adding `random_test.go` gives us a regression net around the RNG primitives (`GenerateCode`, `GenerateInviteCode`) at the same time we TDD the new `GenerateDefaultNickname`.

- [ ] **Step 1.1: Write the failing test for the new helper + safety-net tests for the two existing ones**

Create `dx-api/app/helpers/random_test.go` with:

```go
package helpers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGenerateCode(t *testing.T) {
	for _, length := range []int{1, 4, 6, 16} {
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

	// Sanity: two back-to-back calls should differ. Not a strict uniqueness
	// claim — this catches a broken PRNG that returns constant output.
	// Collision odds for two legitimate calls: 1 in 10^6.
	a, b := GenerateDefaultNickname(), GenerateDefaultNickname()
	if a == b {
		t.Errorf("two consecutive calls returned the same nickname %q", a)
	}
}
```

- [ ] **Step 1.2: Run the tests to verify `TestGenerateDefaultNickname` fails to compile**

Run (from `dx-api/`):

```bash
go test -race ./app/helpers/...
```

Expected: compile error

```
./random_test.go:XX:XX: undefined: GenerateDefaultNickname
FAIL    dx-api/app/helpers [build failed]
```

(The other two tests `TestGenerateCode` and `TestGenerateInviteCode` can't run either because the package fails to build — that's fine, we expect compile-fail → green after Step 1.3.)

- [ ] **Step 1.3: Implement `GenerateDefaultNickname`**

In `dx-api/app/helpers/random.go`, append after the existing `GenerateInviteCode` function and before `randomString`:

```go
// GenerateDefaultNickname returns a default user nickname of the form
// "斗友_XXXXXX" where XXXXXX is 6 crypto-random digits. Uniqueness is not
// guaranteed by design: the users.nickname column has no unique constraint,
// and downstream display code accepts duplicate nicknames.
func GenerateDefaultNickname() string {
	return "斗友_" + GenerateCode(6)
}
```

The complete file after this edit should look like:

```go
package helpers

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	digits       = "0123456789"
	alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// GenerateCode returns a random N-digit numeric string (e.g. "482916")
func GenerateCode(length int) string {
	return randomString(length, digits)
}

// GenerateInviteCode returns a random alphanumeric invite code
func GenerateInviteCode(length int) string {
	return randomString(length, alphanumeric)
}

// GenerateDefaultNickname returns a default user nickname of the form
// "斗友_XXXXXX" where XXXXXX is 6 crypto-random digits. Uniqueness is not
// guaranteed by design: the users.nickname column has no unique constraint,
// and downstream display code accepts duplicate nicknames.
func GenerateDefaultNickname() string {
	return "斗友_" + GenerateCode(6)
}

// randomString generates a cryptographically random string from the given charset
func randomString(length int, charset string) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}
```

- [ ] **Step 1.4: Run the tests to verify they pass**

Run (from `dx-api/`):

```bash
go test -race ./app/helpers/...
```

Expected output:

```
ok      dx-api/app/helpers      <time>s
```

All three tests (`TestGenerateCode`, `TestGenerateInviteCode`, `TestGenerateDefaultNickname`) pass.

- [ ] **Step 1.5: Run `gofmt` and `go vet` to guard against lint regressions**

Run (from `dx-api/`):

```bash
gofmt -l app/helpers/random.go app/helpers/random_test.go
go vet ./app/helpers/...
```

Expected: `gofmt -l` prints nothing (no unformatted files). `go vet` prints nothing (no findings).

- [ ] **Step 1.6: Commit**

```bash
git add dx-api/app/helpers/random.go dx-api/app/helpers/random_test.go
git commit -m "feat(api): add GenerateDefaultNickname helper with tests"
```

---

## Task 2 — Backend: wire default nickname into WeChat registration

**Files:**
- Modify: `dx-api/app/services/api/wechat_auth_service.go:100-109`

**Why no new test here:** `WechatMiniSignIn` depends on `facades.Orm()` and a live WeChat HTTP session, neither of which is bootstrapped in unit-test context. The correctness of the default nickname is fully proven by `TestGenerateDefaultNickname` from Task 1. The service change here is a 2-line struct-field assignment whose correctness is guaranteed by the Go type-checker (wrong type → compile error).

- [ ] **Step 2.1: Read the target file to confirm current state**

Run:

```bash
sed -n '95,115p' dx-api/app/services/api/wechat_auth_service.go
```

Expected output (lines 95-115):

```go
		pw := helpers.GenerateInviteCode(16)
		hashedPw, hashErr := helpers.HashPassword(pw)
		if hashErr != nil {
			return "", nil, fmt.Errorf("failed to hash password: %w", hashErr)
		}

		openID := session.OpenID
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Grade:      consts.UserGradeFree,
			Username:   username,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
			OpenID:     &openID,
		}
		if session.UnionID != "" {
			unionID := session.UnionID
			user.UnionID = &unionID
		}
```

If the output doesn't match exactly (e.g., line numbers have drifted, struct has extra fields), STOP and re-read the file to find the correct insertion point before proceeding.

- [ ] **Step 2.2: Insert the two new lines**

Apply this edit:

**Replace** (currently at lines 101-109 of `dx-api/app/services/api/wechat_auth_service.go`):

```go
		openID := session.OpenID
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Grade:      consts.UserGradeFree,
			Username:   username,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
			OpenID:     &openID,
		}
```

**With:**

```go
		openID := session.OpenID
		nickname := helpers.GenerateDefaultNickname()
		user = models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Grade:      consts.UserGradeFree,
			Username:   username,
			Nickname:   &nickname,
			Password:   hashedPw,
			IsActive:   true,
			InviteCode: helpers.GenerateInviteCode(8),
			OpenID:     &openID,
		}
```

Two lines added:
1. `nickname := helpers.GenerateDefaultNickname()` — one line above the struct literal
2. `Nickname:   &nickname,` — inside the struct literal, between `Username` and `Password`

No other lines change. The `if session.UnionID != ""` block below is untouched.

- [ ] **Step 2.3: Verify the build succeeds**

Run (from `dx-api/`):

```bash
go build ./...
```

Expected: no output, exit code 0. If you see `undefined: helpers.GenerateDefaultNickname`, Task 1 wasn't completed or wasn't saved — fix Task 1 first.

- [ ] **Step 2.4: Verify existing service tests still pass**

Run (from `dx-api/`):

```bash
go test -race ./app/services/api/...
```

Expected: `ok  dx-api/app/services/api  <time>s`. The existing tests (`TestFetchWechatSession_Success`, `TestFetchWechatSession_WechatError`, `TestGenerateWxUsername`) should all still pass — they don't exercise `WechatMiniSignIn` directly, so they're unaffected.

- [ ] **Step 2.5: Run `gofmt` and `go vet` on the edited file**

Run (from `dx-api/`):

```bash
gofmt -l app/services/api/wechat_auth_service.go
go vet ./app/services/api/...
```

Expected: both print nothing.

- [ ] **Step 2.6: Commit**

```bash
git add dx-api/app/services/api/wechat_auth_service.go
git commit -m "feat(api): assign default nickname on WeChat mini registration"
```

---

## Task 3 — Frontend: landing layout (WXML + WXSS)

**Files:**
- Modify: `dx-mini/miniprogram/pages/login/login.wxml`
- Modify: `dx-mini/miniprogram/pages/login/login.wxss`

**Why WXML and WXSS in one task:** The DOM restructure and the CSS rules are atomically coupled — committing one without the other leaves the page in a visibly broken intermediate state (styles targeting non-existent classes, or tagline rendering unstyled). Keeping them in one commit preserves git bisect usefulness for this feature.

**Why no unit test:** dx-mini has no WXML/WXSS test framework. Verification is manual-in-simulator (Step 3.4). The spec's acceptance criteria #7–#11 define the visual checks.

- [ ] **Step 3.1: Read the two target files to confirm current state**

Run:

```bash
cat dx-mini/miniprogram/pages/login/login.wxml
```

Expected output (27 lines, ending with `</view>`):

```xml
<view class="login-page">
  <view class="login-bg" />
  <view class="login-content" style="padding-top: {{statusBarHeight}}px;">
    <view class="capsule-spacer" />
    <view class="spacer-top" />
    <view class="brand">
      <text class="brand-logo">斗学</text>
      <view class="brand-tagline">
        <text class="tagline-line">多种学习模式 · AI 定制内容 · 和朋友一起闯关</text>
        <text class="tagline-line">每天 10 分钟，英语悄悄流利了</text>
      </view>
    </view>
    <view class="spacer-bottom" />
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

If the file doesn't match this exactly, STOP and re-read — the spec assumes this is the starting state.

- [ ] **Step 3.2: Replace the WXML with the new structure**

Overwrite `dx-mini/miniprogram/pages/login/login.wxml` with:

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

**Three structural changes:**
1. `.brand-tagline` block removed from inside `.brand`; `.brand` now contains only the logo.
2. New `.greeting` wrapper inserted between `.spacer-bottom` and `.cta-wrap`.
3. `.greeting` contains: a new `<text class="greeting-title">欢迎来到斗学</text>` heading, followed by the existing `.brand-tagline` block moved here.

No attribute on `<van-button>` changes. No new prop-bindings, no new event handlers, no new data dependencies in `login.ts`.

- [ ] **Step 3.3: Replace the WXSS**

Overwrite `dx-mini/miniprogram/pages/login/login.wxss` with:

```css
/* Override the app-level page background so our gradient layer is the only
   surface visible above the WeChat shell. */
page {
  background: transparent;
}

/* === Layout scaffolding === */

.login-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: linear-gradient(
    to bottom,
    #ccfbf1 0%,
    #dbeafe 25%,
    #ede9fe 50%,
    #fce7f3 75%,
    #ffffff 100%
  );
}

.login-bg::after {
  content: '';
  position: absolute;
  top: 36%;
  left: 50%;
  width: 600rpx;
  height: 600rpx;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, rgba(94, 234, 212, 0.35), transparent 70%);
  pointer-events: none;
}

.login-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-bottom: env(safe-area-inset-bottom);
  box-sizing: border-box;
}

.capsule-spacer {
  height: 88rpx;
  flex: 0 0 auto;
}

.spacer-top    { flex: 0.6 1 0; }
.spacer-bottom { flex: 1 1 0; }

/* === Brand === */

.brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32rpx;
}

.brand-logo {
  font-size: 110rpx;
  font-weight: 800;
  letter-spacing: 12rpx;
  color: #0d9488;
  animation: hero-rise 480ms ease-out both;
}

.brand-tagline {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
}

.tagline-line {
  font-size: 28rpx;
  line-height: 1.6;
  color: #475569;
  text-align: center;
}

/* === Greeting (welcome heading + tagline, glued above CTA) === */

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

/* === CTA === */

.cta-wrap {
  width: 100%;
  padding: 0 60rpx 80rpx;
  box-sizing: border-box;
  animation: hero-rise 480ms ease-out 240ms both;
}

/* === First-paint reveal === */

@keyframes hero-rise {
  from { opacity: 0; transform: translateY(16rpx); }
  to   { opacity: 1; transform: translateY(0);     }
}
```

**Two CSS changes vs. the current file:**
1. `.brand-tagline` — drop the `padding: 0 60rpx;` line and the `animation: hero-rise 480ms ease-out 120ms both;` line. The selector remains, only those two properties are gone.
2. Add the `.greeting` and `.greeting-title` rules (the new "Greeting" section). The `.greeting` rule carries the 120ms-delayed `hero-rise` animation that used to sit on `.brand-tagline`.

All other rules (`page`, `.login-page`, `.login-bg`, `.login-bg::after`, `.login-content`, `.capsule-spacer`, `.spacer-top`, `.spacer-bottom`, `.brand`, `.brand-logo`, `.tagline-line`, `.cta-wrap`, `@keyframes hero-rise`) are byte-identical to the current version.

- [ ] **Step 3.4: Verify the mini program compiles (no TS errors introduced)**

Run (from `dx-mini/`):

```bash
npx tsc --noEmit -p tsconfig.json 2>&1 | head -80
```

Expected: any output matches the pre-existing `miniprogram-api-typings` v2.8.3 `this`-in-`Component({methods})` warnings only (see CLAUDE.md / the dx-mini conventions). No **new** errors. If a new error appears that references `pages/login/`, something upstream is wrong — stop and investigate. The login page file we touched (`login.wxml`, `login.wxss`) are not TS files, so they shouldn't cause a TS error on their own.

- [ ] **Step 3.5: Manual visual verification in WeChat DevTools**

Open the dx-mini project in WeChat Developer Tools at `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini`. In the simulator:

1. Navigate to `pages/login/login`.
2. Confirm the layout renders top-to-bottom as:
   - Status bar area (gradient flows under it)
   - Capsule menu in the top right
   - Empty upper zone
   - **Logo `斗学`** in teal at ~37.5% of the mid-zone (visually identical position to the pre-change build)
   - Empty mid-to-lower zone
   - **Welcome heading `欢迎来到斗学`** in dark slate
   - **Tagline** — two lines in lighter slate, immediately below the heading
   - **微信一键登录 button** with a ~40rpx gap above it
3. Test on three device presets: **iPhone 14 Pro**, **iPhone SE (1st gen, 320×568)**, **Pixel 6**. On iPhone SE, the welcome+tagline group must not overlap the button.
4. Reload the page (`Cmd+R` in DevTools) and watch the animation: logo fades in first (0ms), welcome+tagline group fades in as ONE unit (120ms), button fades in last (240ms). Three beats, not four.
5. Tap the 微信一键登录 button (DevTools auto-mocks `wx.login`). The existing click handler fires — no regression on the login flow.

If any of the five checks above fails, STOP and fix before committing. Common fixes:
- Logo shifted up/down: verify `.spacer-top: 0.6 1 0;` and `.spacer-bottom: 1 1 0;` are intact in the CSS.
- Welcome+tagline not glued to button: verify `.greeting` has `padding: 0 60rpx 40rpx` and sits as a direct child of `.login-content` between `.spacer-bottom` and `.cta-wrap`.
- Only welcome OR only tagline animates (not both together): verify `.brand-tagline` has no `animation` property — the animation must live ONLY on the parent `.greeting`.

- [ ] **Step 3.6: Commit**

```bash
git add dx-mini/miniprogram/pages/login/login.wxml dx-mini/miniprogram/pages/login/login.wxss
git commit -m "feat(mini): add 欢迎来到斗学 greeting above login CTA"
```

---

## Task 4 — End-to-end smoke (integration verification)

**Files touched:** none (verification-only).

**Why this task:** All three previous tasks passed their own unit-level checks, but the spec includes one integration-level acceptance criterion (spec §11, #10): "Completing a WeChat login on a fresh account shows the new nickname `斗友_XXXXXX` in the `/me` profile screen." This task executes that check.

- [ ] **Step 4.1: Start dx-api locally**

In one terminal (from `dx-api/`):

```bash
air
```

(or `go run .` if `air` isn't installed). Verify logs show the server bound to port 3001.

- [ ] **Step 4.2: Point dx-mini at the local dx-api**

In the WeChat DevTools console for the dx-mini project:

```js
require('./utils/config').setDevApiBaseUrl('http://<your-lan-ip>')
```

(Replace `<your-lan-ip>` with the machine's LAN IP — e.g., `http://192.168.1.50`. DevTools simulator can't reach `localhost` directly on all macOS versions; LAN IP is safer per the dx-mini README.)

Also: 详情 → 本地设置 → 勾选「不校验合法域名…」

- [ ] **Step 4.3: Register a fresh WeChat user**

In DevTools, open a clean session (clear storage via 详情 → 清缓存 → 全部清除 if needed). Navigate to `pages/login/login` and tap **微信一键登录**. The flow should:

1. POST to `http://<lan-ip>:3001/api/auth/wechat-mini` (dev returns a stub/test OpenID since DevTools doesn't speak real WeChat API).
2. `wx.reLaunch` to `/pages/home/home`.

- [ ] **Step 4.4: Verify the DB row has the default nickname**

In a third terminal:

```bash
psql -d douxue -c "SELECT id, username, nickname, openid FROM users ORDER BY created_at DESC LIMIT 1;"
```

Expected columns:
- `username` matches `wx_<8-char-openid-prefix>` (optionally with a `_XXXX` suffix if there was a collision)
- `nickname` matches the regex `^斗友_[0-9]{6}$` — for example `斗友_482916`
- `openid` is non-null

If `nickname` is `NULL`, Task 2 wasn't applied correctly — revisit the service edit.

- [ ] **Step 4.5: Verify the nickname surfaces in the mini UI**

Navigate to the `/me` (or equivalent profile) page in the mini program. Confirm the displayed nickname reads `斗友_XXXXXX` (six digits), not the `wx_...` username fallback.

If the `/me` page still shows the username, check whether the mini's profile view reads `user.nickname` directly or via a fallback helper — the fallback takes `user.username` when nickname is empty. With nickname now populated on fresh registrations, it should display as-is.

- [ ] **Step 4.6: Clean up the test user (optional)**

If you want to keep the DB tidy:

```bash
psql -d douxue -c "DELETE FROM users WHERE openid LIKE 'test_%' OR created_at > now() - interval '10 minutes';"
```

Only run if you're certain the target row is test-only — review `SELECT` output first.

- [ ] **Step 4.7: No commit for this task**

Task 4 is verification-only. If any step fails, fix the earlier task it points to, push a new commit there, and re-run from Step 4.3.

---

## Final checks (run once, after all four tasks complete)

- [ ] **Final 1: All dx-api tests green with race detector**

```bash
cd dx-api && go test -race ./...
```

Expected: all packages `ok`. If any new failures appear, they're regressions caused by this plan — fix before declaring done.

- [ ] **Final 2: Full dx-api build + vet + gofmt**

```bash
cd dx-api && go build ./... && go vet ./... && gofmt -l .
```

Expected: no output.

- [ ] **Final 3: dx-mini TS compile (no new errors)**

```bash
cd dx-mini && npx tsc --noEmit -p tsconfig.json 2>&1 | grep -E "error TS" | wc -l
```

Expected: the count matches the pre-change baseline (the known `this`-in-`Component({methods})` warnings per CLAUDE.md). Record the number before starting and compare.

- [ ] **Final 4: git log confirms 3 commits on top of the spec commit**

```bash
git log --oneline -5
```

Expected:

```
<sha> feat(mini): add 欢迎来到斗学 greeting above login CTA
<sha> feat(api): assign default nickname on WeChat mini registration
<sha> feat(api): add GenerateDefaultNickname helper with tests
1f0729e docs(mini): spec for login welcome heading + WeChat default nickname
<sha> ... (older commits)
```

Four feature-adjacent commits total, well-scoped, bisect-friendly.

---

## Self-review (executed while writing this plan)

- **Spec coverage:**
  - Frontend Req #1 (welcome heading above tagline) → Task 3 Step 3.2 WXML
  - Frontend Req #2 (group pinned above CTA) → Task 3 Step 3.2 + 3.3 (`.greeting` between `.spacer-bottom` and `.cta-wrap`, `padding-bottom: 40rpx`)
  - Frontend Req #3 (logo alone in upper zone) → Task 3 Step 3.2 (`.brand` contains only `.brand-logo`)
  - Frontend Req #4 (logo vertical position unchanged) → Task 3 Step 3.3 (`.spacer-top`/`.spacer-bottom` ratios unchanged) + Step 3.5 visual check
  - Frontend Req #5 (3-beat motion) → Task 3 Step 3.3 (`.greeting` animation at 120ms, `.cta-wrap` at 240ms) + Step 3.5 visual check
  - Frontend Req #6 (welcome heading typography) → Task 3 Step 3.3 `.greeting-title` rule
  - Frontend Req #7 (tagline typography unchanged) → Task 3 Step 3.3 (`.tagline-line` rule copied verbatim)
  - Frontend Req #8 (no new assets, no `?.`/`??`) → Task 3 touches only WXML+WXSS; neither allows those operators, so N/A by construction
  - Frontend Req #9 (no new tsc errors) → Final 3
  - Backend Req #1 (nickname set on new WeChat users) → Task 2 Step 2.2
  - Backend Req #2 (generated at registration, persisted) → Task 2 Step 2.2 (local var → struct field → Create call at existing line 115)
  - Backend Req #3 (no backfill, no migration) → confirmed by non-inclusion — no task touches migrations or existing rows
  - Backend Req #4 (email/password not touched) → confirmed by non-inclusion — no task touches `auth_service.go`
  - Backend Req #5 (fallback still works) → not actively tested but preserved: the spec fallback pattern `if user.Nickname != nil && *user.Nickname != ""` is untouched in downstream services
  - Backend Req #6 (gofmt/vet/tests clean) → Task 1 Step 1.5, Task 2 Step 2.4–2.5, Final 1–2

- **Placeholder scan:** no TBDs, no "implement later", no "appropriate error handling," no "similar to Task N". Every code block is runnable as-is. Every command has an expected output.

- **Type consistency:** helper `GenerateDefaultNickname()` returns `string` (no error). Caller takes its return into a local `nickname` and assigns `&nickname` to `user.Nickname` (type `*string`, confirmed in User model). Test `TestGenerateDefaultNickname` asserts on `string` return. All consistent.
