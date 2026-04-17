# Hall Greeting Const Design

## Summary

Replace the hardcoded greeting title/subtitle on the hall home page with server-driven content. Add a new const file `dx-api/app/consts/greeting.go` defining four time-banded titles (including emoji) and five subtitles per band. Expose the result through the existing `/api/hall/dashboard` response so the home page renders a correct greeting for the current hour with a randomly-picked subtitle on every load.

## Current State

- `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx:64-67` hardcodes:
  - title: `` `早上好，${displayName} 👋` ``
  - subtitle: `"继续你的学习之旅，今天也要加油！"`
- `GreetingTopBar` (`dx-web/src/features/web/hall/components/greeting-top-bar.tsx`) truncates title to 15 chars, subtitle to 20 chars.
- `/api/hall/dashboard` returns `DashboardData` from `dx-api/app/services/api/hall_service.go:14-20` (profile, masterStats, reviewStats, sessions, todayAnswers).

## Time Bands (Asia/Shanghai)

| Band | Hour range (inclusive) | Title |
|------|------------------------|-------|
| morning | 05–10 | `早上好 👋` |
| noon | 11–12 | `中午好 🍚` |
| afternoon | 13–17 | `下午好 ☕` |
| evening | 18–23 and 0–4 | `晚上好 🌙` |

Rule: titles are chosen by the hour component of the server-side `time.Time` after it is converted to `Asia/Shanghai`. Dashboard data is always fetched live, so the title reflects the clock at request time.

## Subtitles

All strings are ≤ 20 Unicode code points to stay within the truncation limit in `greeting-top-bar.tsx:23`.

**早上好**
1. `继续你的学习之旅，今天也要加油！`
2. `新的一天，一起来背几个单词吧！`
3. `早起的鸟儿有虫吃，冲呀！`
4. `今天也要笑着开始学习哦～`
5. `愿你的一天充满阳光和单词`

**中午好**
1. `吃饭前先来几道题热身吧！`
2. `午饭后，刷两道 quiz 如何？`
3. `中午能量满满，继续冲刺！`
4. `午休时间，来场英文小游戏吧`
5. `一顿好饭配一页单词，完美！`

**下午好**
1. `一杯咖啡配英语，下午更带劲`
2. `一起消灭那些顽固的生词吧！`
3. `午后微困？来段英语提提神！`
4. `坚持一下，今天的目标不远了`
5. `让英语给你的下午续点航`

**晚上好**
1. `结束今天前，再多学一点点`
2. `夜深人静，正适合练听力`
3. `月亮不睡你也别睡，单词等你`
4. `睡前温习，记忆更牢哦`
5. `今日份英语打卡，完成！`

## Backend

### 1. New file: `dx-api/app/consts/greeting.go`

Public types and function:

```go
// Greeting is a time-banded greeting for the hall dashboard.
type Greeting struct {
    Title    string `json:"title"`
    Subtitle string `json:"subtitle"`
}

// PickGreeting returns a Greeting whose Title matches the hour of t (interpreted
// in Asia/Shanghai) and whose Subtitle is a random entry from the band's pool.
// Uses math/rand/v2 top-level rand.IntN for subtitle selection — same pattern as
// dx-api/app/services/api/mock_user_service.go and game_play_pk_service.go.
func PickGreeting(t time.Time) Greeting
```

Internal shape: a private `greetingBand` struct holding `{ title string; subtitles []string }` and a private package-level slice `greetingBands` with the four bands in order morning → evening. The title/subtitle literals are declared directly inside that slice.

Hour selection helper:

```go
// bandFor returns the band index (0=morning, 1=noon, 2=afternoon, 3=evening)
// for the given hour (0–23). Unknown hours fall into evening.
func bandFor(hour int) int
```

Shanghai timezone is loaded once at package init via `time.LoadLocation("Asia/Shanghai")`. If loading fails (impossible in practice since the tzdata is compiled in for Goravel deploys), the init panics — consistent with how `dx-api` handles missing tzdata elsewhere. `PickGreeting` converts `t` to Shanghai internally; callers pass `time.Now()` directly.

### 2. New file: `dx-api/app/consts/greeting_test.go`

Table-driven tests covering:
- `bandFor` for each hour 0–23 returning the expected index.
- `PickGreeting` title for representative hours in each band, plus UTC→Shanghai conversion (e.g. 00:00 UTC → 08:00 Shanghai → morning).
- Boundary hours: 4 → evening, 5 → morning, 10 → morning, 11 → noon, 12 → noon, 13 → afternoon, 17 → afternoon, 18 → evening.
- Property: `PickGreeting(t).Subtitle` always belongs to the subtitle pool of the band implied by `t`'s hour. Call many times per hour (e.g. 50) to exercise randomness.
- Sanity: 4 bands, every title non-empty, every band has exactly 5 subtitles, every subtitle non-empty and ≤ 20 runes.

### 3. Modified: `dx-api/app/services/api/hall_service.go`

Extend `DashboardData`:

```go
type DashboardData struct {
    Profile      DashboardProfile  `json:"profile"`
    MasterStats  MasterStats       `json:"masterStats"`
    ReviewStats  ReviewStats       `json:"reviewStats"`
    Sessions     []SessionProgress `json:"sessions"`
    TodayAnswers int               `json:"todayAnswers"`
    Greeting     consts.Greeting   `json:"greeting"` // new
}
```

In `GetDashboard` (just before the final `return`), add:

```go
greeting := consts.PickGreeting(time.Now())
```

and assign `Greeting: greeting` into the returned `DashboardData`. No other field, query, or behavior changes.

Imports added: `"dx-api/app/consts"`.

### 4. No controller or routing change

`GetDashboard` on `HallController` is already wired to `/api/hall/dashboard` and simply returns whatever the service builds — it does not need modification. The admin routes are not touched.

## Frontend

### 5. Modified: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`

- Extend the local `DashboardData` type with `greeting: { title: string; subtitle: string }`.
- Replace the hardcoded props on `<GreetingTopBar …/>` with values from `data.greeting`, composing the user name into the title:

```tsx
<GreetingTopBar
  title={data?.greeting
    ? `${data.greeting.title}，${displayName}`
    : `早上好，${displayName}`}
  subtitle={data?.greeting?.subtitle ?? "继续你的学习之旅，今天也要加油！"}
/>
```

The loading fallback preserves today's UX (shows `早上好，同学` with the requested subtitle) while the fetch is in flight, so the page never flashes empty strings.

### 6. No other frontend files change

`greeting-top-bar.tsx` already takes `title` and `subtitle` as props. `TopActions` and surrounding layout are unchanged.

## API Response Shape

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "profile": { "...": "..." },
    "masterStats": { "...": "..." },
    "reviewStats": { "...": "..." },
    "sessions": [],
    "todayAnswers": 0,
    "greeting": {
      "title": "下午好 ☕",
      "subtitle": "一杯咖啡配英语，下午更带劲"
    }
  }
}
```

`greeting.title` always includes the emoji. Adding a field is backward-compatible — existing clients ignore it.

## Files Changed

**New files:**
- `dx-api/app/consts/greeting.go`
- `dx-api/app/consts/greeting_test.go`

**Modified files:**
- `dx-api/app/services/api/hall_service.go` — extend `DashboardData`, populate `Greeting` in `GetDashboard`
- `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` — read `data.greeting`, compose title with `displayName`, loading fallback

## Verification

- `cd dx-api && go vet ./... && go test -race ./app/consts/...` — passes; new tests cover band boundaries, subtitle membership, length invariants.
- `cd dx-api && go build ./...` — compiles.
- `cd dx-web && npm run lint` — no new lint errors.
- Manual: hit `/api/hall/dashboard` at different hours (or by stubbing `time.Now()` in a local run) and confirm `greeting.title` changes with the hour and `greeting.subtitle` varies across refreshes.
- Manual: load `/hall` in the browser and confirm the top bar shows the expected title/subtitle with the logged-in user's name.

## Non-Goals

- No new API endpoint.
- No client-side time logic.
- No changes to any admin surface.
- No i18n — titles/subtitles stay Simplified Chinese, matching the rest of the hall UI.
- No caching — the service builds a fresh greeting every request; the dashboard endpoint is already personalized and uncached.
