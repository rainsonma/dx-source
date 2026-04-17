# Hall Greeting Const Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hardcoded hall home-page greeting with server-driven content picked by hour-of-day from a new Go const file, with a randomly chosen subtitle.

**Architecture:** New `dx-api/app/consts/greeting.go` owns the four time-banded titles (emoji included) and five subtitles per band, plus a `PickGreeting(t time.Time) Greeting` function. Subtitle randomness uses `math/rand/v2` top-level `rand.IntN`, matching existing dx-api patterns (`mock_user_service.go`, `game_play_pk_service.go`). The hall dashboard service calls `PickGreeting` once per request and returns it in a new `greeting` field on the existing `/api/hall/dashboard` response. The hall home page reads `data.greeting` and renders it through the existing `GreetingTopBar` — no new endpoints, no client-side time logic.

**Tech Stack:** Go 1.x (Goravel), GORM, Next.js 16 (App Router, TypeScript), TailwindCSS v4.

---

## File Structure

**Create:**
- `dx-api/app/consts/greeting.go` — `Greeting` struct, private band table, `bandFor(hour int) int`, `PickGreeting(t time.Time) Greeting`, package-init Shanghai timezone load.
- `dx-api/app/consts/greeting_test.go` — table-driven tests for `bandFor`, title-by-hour, UTC→Shanghai conversion, subtitle-membership invariant, subtitle rune-length invariant.

**Modify:**
- `dx-api/app/services/api/hall_service.go` — add `Greeting consts.Greeting` field to `DashboardData`; populate it in `GetDashboard`. Imports: add `dx-api/app/consts`.
- `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` — extend local `DashboardData` type; compose title via `${data.greeting.title}，${displayName}`; pass `data.greeting.subtitle`; keep current strings as loading fallbacks.

---

### Task 1: Create the greeting const file

**Files:**
- Create: `dx-api/app/consts/greeting.go`

- [ ] **Step 1: Write `dx-api/app/consts/greeting.go`**

```go
package consts

import (
	"math/rand/v2"
	"time"
)

// Greeting is a time-banded greeting for the hall dashboard.
type Greeting struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type greetingBand struct {
	title     string
	subtitles []string
}

var (
	shanghaiLocation *time.Location
	greetingBands    []greetingBand
)

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("consts: failed to load Asia/Shanghai timezone: " + err.Error())
	}
	shanghaiLocation = loc

	greetingBands = []greetingBand{
		// 0 — morning: 05–10
		{
			title: "早上好 👋",
			subtitles: []string{
				"继续你的学习之旅，今天也要加油！",
				"新的一天，一起来背几个单词吧！",
				"早起的鸟儿有虫吃，冲呀！",
				"今天也要笑着开始学习哦～",
				"愿你的一天充满阳光和单词",
			},
		},
		// 1 — noon: 11–12
		{
			title: "中午好 🍚",
			subtitles: []string{
				"吃饭前先来几道题热身吧！",
				"午饭后，刷两道 quiz 如何？",
				"中午能量满满，继续冲刺！",
				"午休时间，来场英文小游戏吧",
				"一顿好饭配一页单词，完美！",
			},
		},
		// 2 — afternoon: 13–17
		{
			title: "下午好 ☕",
			subtitles: []string{
				"一杯咖啡配英语，下午更带劲",
				"一起消灭那些顽固的生词吧！",
				"午后微困？来段英语提提神！",
				"坚持一下，今天的目标不远了",
				"让英语给你的下午续点航",
			},
		},
		// 3 — evening: 18–23 and 0–4
		{
			title: "晚上好 🌙",
			subtitles: []string{
				"结束今天前，再多学一点点",
				"夜深人静，正适合练听力",
				"月亮不睡你也别睡，单词等你",
				"睡前温习，记忆更牢哦",
				"今日份英语打卡，完成！",
			},
		},
	}
}

// bandFor returns the band index (0=morning, 1=noon, 2=afternoon, 3=evening)
// for the given hour (0–23). Out-of-range hours fall into evening.
func bandFor(hour int) int {
	switch {
	case hour >= 5 && hour <= 10:
		return 0
	case hour >= 11 && hour <= 12:
		return 1
	case hour >= 13 && hour <= 17:
		return 2
	default:
		return 3
	}
}

// PickGreeting returns a Greeting whose Title matches the hour of t
// (interpreted in Asia/Shanghai) and whose Subtitle is a random entry
// from the band's pool. Uses math/rand/v2 top-level rand.IntN for
// subtitle selection — same pattern as services/api/mock_user_service.go.
func PickGreeting(t time.Time) Greeting {
	hour := t.In(shanghaiLocation).Hour()
	band := greetingBands[bandFor(hour)]
	return Greeting{
		Title:    band.title,
		Subtitle: band.subtitles[rand.IntN(len(band.subtitles))],
	}
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd dx-api && go build ./app/consts/...`
Expected: no output, exit 0.

---

### Task 2: Write greeting tests

**Files:**
- Create: `dx-api/app/consts/greeting_test.go`

- [ ] **Step 1: Write `dx-api/app/consts/greeting_test.go`**

```go
package consts

import (
	"testing"
	"time"
	"unicode/utf8"
)

func TestBandFor(t *testing.T) {
	tests := []struct {
		hour int
		want int
	}{
		{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3},
		{5, 0}, {6, 0}, {7, 0}, {8, 0}, {9, 0}, {10, 0},
		{11, 1}, {12, 1},
		{13, 2}, {14, 2}, {15, 2}, {16, 2}, {17, 2},
		{18, 3}, {19, 3}, {20, 3}, {21, 3}, {22, 3}, {23, 3},
	}
	for _, tc := range tests {
		if got := bandFor(tc.hour); got != tc.want {
			t.Errorf("bandFor(%d) = %d, want %d", tc.hour, got, tc.want)
		}
	}
}

func TestPickGreetingTitleByHour(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	tests := []struct {
		hour  int
		title string
	}{
		{4, "晚上好 🌙"},
		{5, "早上好 👋"},
		{10, "早上好 👋"},
		{11, "中午好 🍚"},
		{12, "中午好 🍚"},
		{13, "下午好 ☕"},
		{17, "下午好 ☕"},
		{18, "晚上好 🌙"},
		{23, "晚上好 🌙"},
		{0, "晚上好 🌙"},
	}
	for _, tc := range tests {
		tt := time.Date(2026, 4, 17, tc.hour, 0, 0, 0, shanghai)
		g := PickGreeting(tt)
		if g.Title != tc.title {
			t.Errorf("hour %d: title = %q, want %q", tc.hour, g.Title, tc.title)
		}
	}
}

func TestPickGreetingConvertsToShanghai(t *testing.T) {
	// 00:00 UTC → 08:00 Shanghai → morning
	utc := time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)
	g := PickGreeting(utc)
	if g.Title != "早上好 👋" {
		t.Errorf("UTC 00:00 (Shanghai 08:00): title = %q, want 早上好 👋", g.Title)
	}
}

func TestPickGreetingSubtitleInBand(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	bandHours := map[int][]int{
		0: {5, 6, 7, 8, 9, 10},
		1: {11, 12},
		2: {13, 14, 15, 16, 17},
		3: {18, 19, 20, 21, 22, 23, 0, 1, 2, 3, 4},
	}
	for bandIdx, hours := range bandHours {
		pool := greetingBands[bandIdx].subtitles
		for _, hour := range hours {
			tt := time.Date(2026, 4, 17, hour, 0, 0, 0, shanghai)
			for i := 0; i < 50; i++ {
				g := PickGreeting(tt)
				if !containsString(pool, g.Subtitle) {
					t.Errorf("hour %d iter %d: subtitle %q not in band %d pool",
						hour, i, g.Subtitle, bandIdx)
				}
			}
		}
	}
}

func TestGreetingBandInvariants(t *testing.T) {
	if len(greetingBands) != 4 {
		t.Fatalf("greetingBands length = %d, want 4", len(greetingBands))
	}
	for i, band := range greetingBands {
		if band.title == "" {
			t.Errorf("band %d has empty title", i)
		}
		if len(band.subtitles) != 5 {
			t.Errorf("band %d has %d subtitles, want 5", i, len(band.subtitles))
		}
		for j, s := range band.subtitles {
			if s == "" {
				t.Errorf("band %d subtitle %d is empty", i, j)
			}
			if runes := utf8.RuneCountInString(s); runes > 20 {
				t.Errorf("band %d subtitle %d = %q has %d runes, want ≤ 20",
					i, j, s, runes)
			}
		}
	}
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run the new tests**

Run: `cd dx-api && go test -race ./app/consts/...`
Expected: `ok  	dx-api/app/consts	...` — all tests pass, including pre-existing `user_level_test.go`.

- [ ] **Step 3: Run `go vet` on the consts package**

Run: `cd dx-api && go vet ./app/consts/...`
Expected: no output, exit 0.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/consts/greeting.go dx-api/app/consts/greeting_test.go
git commit -m "feat(api): add time-banded greeting const with randomized subtitles"
```

---

### Task 3: Wire greeting into dashboard response

**Files:**
- Modify: `dx-api/app/services/api/hall_service.go:3-15` (imports), `:14-20` (`DashboardData` struct), `:131-137` (return in `GetDashboard`).

- [ ] **Step 1: Update imports**

Replace the current import block at `dx-api/app/services/api/hall_service.go:3-11`:

```go
import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)
```

with:

```go
import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)
```

- [ ] **Step 2: Extend `DashboardData`**

Replace the struct at `dx-api/app/services/api/hall_service.go:14-20`:

```go
// DashboardData aggregates user stats for the hall dashboard.
type DashboardData struct {
	Profile      DashboardProfile  `json:"profile"`
	MasterStats  MasterStats       `json:"masterStats"`
	ReviewStats  ReviewStats       `json:"reviewStats"`
	Sessions     []SessionProgress `json:"sessions"`
	TodayAnswers int               `json:"todayAnswers"`
}
```

with:

```go
// DashboardData aggregates user stats for the hall dashboard.
type DashboardData struct {
	Profile      DashboardProfile  `json:"profile"`
	MasterStats  MasterStats       `json:"masterStats"`
	ReviewStats  ReviewStats       `json:"reviewStats"`
	Sessions     []SessionProgress `json:"sessions"`
	TodayAnswers int               `json:"todayAnswers"`
	Greeting     consts.Greeting   `json:"greeting"`
}
```

- [ ] **Step 3: Populate the greeting in `GetDashboard`**

Replace the final `return` block at `dx-api/app/services/api/hall_service.go:131-137`:

```go
	return &DashboardData{
		Profile:      profile,
		MasterStats:  *masterStats,
		ReviewStats:  *reviewStats,
		Sessions:     sessions,
		TodayAnswers: todayAnswers,
	}, nil
}
```

with:

```go
	return &DashboardData{
		Profile:      profile,
		MasterStats:  *masterStats,
		ReviewStats:  *reviewStats,
		Sessions:     sessions,
		TodayAnswers: todayAnswers,
		Greeting:     consts.PickGreeting(time.Now()),
	}, nil
}
```

- [ ] **Step 4: Build and vet the API**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 5: Run the full test suite with race detector**

Run: `cd dx-api && go test -race ./...`
Expected: all packages pass (`ok` for each). No pre-existing test should start failing.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/hall_service.go
git commit -m "feat(api): return greeting in /api/hall/dashboard"
```

---

### Task 4: Render server-driven greeting on the hall home page

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx:14-34` (`DashboardData` type), `:60-67` (greeting render).

- [ ] **Step 1: Extend the local `DashboardData` type**

Replace the type at `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx:14-34`:

```ts
type DashboardData = {
  profile: {
    username: string;
    nickname: string | null;
    exp: number;
    currentPlayStreak: number;
  };
  masterStats: { total: number; thisWeek: number };
  reviewStats: { pending: number };
  sessions: {
    gameId: string;
    gameName: string;
    gameMode: string;
    completedLevels: number;
    totalLevels: number;
    score: number;
    exp: number;
    lastPlayedAt: Date;
  }[];
  todayAnswers: number;
};
```

with:

```ts
type DashboardData = {
  profile: {
    username: string;
    nickname: string | null;
    exp: number;
    currentPlayStreak: number;
  };
  masterStats: { total: number; thisWeek: number };
  reviewStats: { pending: number };
  sessions: {
    gameId: string;
    gameName: string;
    gameMode: string;
    completedLevels: number;
    totalLevels: number;
    score: number;
    exp: number;
    lastPlayedAt: Date;
  }[];
  todayAnswers: number;
  greeting: {
    title: string;
    subtitle: string;
  };
};
```

- [ ] **Step 2: Render the server greeting**

Replace the `<GreetingTopBar … />` block at `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx:64-67`:

```tsx
      <GreetingTopBar
        title={`早上好，${displayName} 👋`}
        subtitle="继续你的学习之旅，今天也要加油！"
      />
```

with:

```tsx
      <GreetingTopBar
        title={
          data?.greeting
            ? `${data.greeting.title}，${displayName}`
            : `早上好，${displayName}`
        }
        subtitle={
          data?.greeting?.subtitle ?? "继续你的学习之旅，今天也要加油！"
        }
      />
```

- [ ] **Step 3: Lint the web app**

Run: `cd dx-web && npm run lint`
Expected: no new warnings or errors. If ESLint flags formatting, fix the specific lines and rerun until clean.

- [ ] **Step 4: Type-check via build (smoke)**

Run: `cd dx-web && npx tsc --noEmit`
Expected: no type errors.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx
git commit -m "feat(web): render server-driven greeting on hall home"
```

---

### Task 5: Manual end-to-end verification

**Files:** (none — runtime check only)

- [ ] **Step 1: Start the backend**

Run in one shell: `cd dx-api && go run .`
Expected: server starts, logs `http://localhost:3001` (or the configured port) with no panic. The `init()` in `consts/greeting.go` must not fail to load the Shanghai timezone.

- [ ] **Step 2: Hit the dashboard endpoint directly**

Run in another shell (substitute a real user JWT you already have):

```bash
curl -s -H "Authorization: Bearer $DX_USER_JWT" http://localhost:3001/api/hall/dashboard | jq '.data.greeting'
```

Expected: JSON object with `title` (one of `早上好 👋 / 中午好 🍚 / 下午好 ☕ / 晚上好 🌙`, matching the current Shanghai hour) and `subtitle` (from that band's pool). Repeat the curl a few times — the subtitle should vary.

- [ ] **Step 3: Start the web app and view the hall**

Run in another shell: `cd dx-web && npm run dev`
Open `http://localhost:3000/hall` after logging in.
Expected: top bar shows `{title}，{displayName}` and a subtitle from the current band. Refreshing the page rerolls the subtitle.

- [ ] **Step 4: Confirm loading fallback**

In DevTools, throttle the `/api/hall/dashboard` request to a slow profile (e.g. 3G) and refresh. Expected: while the request is in flight the top bar shows `早上好，{displayName}` with `继续你的学习之旅，今天也要加油！`, then swaps to the real greeting once the response arrives. No flash of empty text.

- [ ] **Step 5: Report completion**

Summarize: backend const added, dashboard endpoint now returns `greeting`, home page renders it, all tests + `go vet` + `npm run lint` clean, manual verification done. No existing behavior changed.
