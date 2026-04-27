# dx-mini Sticky 搜索课程 Search Bar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a fixed-at-top, capsule-aware "搜索课程" launcher to the dx-mini home and 课程 (games) pages, plus a new dedicated search page that lists 最近搜索 + 搜索发现 chips and renders cursor-paginated results from `GET /api/games?q=...`. One new public dx-api endpoint, `GET /api/games/search-suggestions`, returns the popular-discovery chips (Redis-cached 1h).

**Architecture:** A single shared component (`dx-search-bar`) owns the WeChat-capsule alignment math (`wx.getMenuButtonBoundingClientRect`) so home / courses / search all share one source of truth. The home launcher hides on cold open and reveals via `onPageScroll` (Taobao-style). The courses launcher is always pinned and the existing van-tabs row becomes sticky directly under it. The new search page uses a real `<input>` with submit-only triggers (keyboard 搜索 + chip taps), persists last-10 history in `wx.storage` under `dx_recent_searches`, and reuses the existing courses-page game-card markup verbatim.

**Tech Stack:** Go (Goravel ORM, Redis, gorm raw SQL), WeChat Mini Program (TypeScript strict, glass-easel components, Vant Weapp 1.11), Lucide-static icons via the `dx-icon` component, `wx.storage` for local persistence.

**Reference spec:** `docs/superpowers/specs/2026-04-27-dx-mini-sticky-search-design.md`

---

## File Map

**dx-api (modify only):**
- `dx-api/app/services/api/game_service.go` — add `GetSearchSuggestions()` and 4 unexported constants
- `dx-api/app/services/api/game_service_test.go` — add function-existence test
- `dx-api/app/http/controllers/api/game_controller.go` — add `(*GameController).SearchSuggestions` method
- `dx-api/routes/api.go` — register `games.Get("/search-suggestions", gameController.SearchSuggestions)`

**dx-mini (create):**
- `dx-mini/miniprogram/components/dx-search-bar/index.json`
- `dx-mini/miniprogram/components/dx-search-bar/index.wxml`
- `dx-mini/miniprogram/components/dx-search-bar/index.wxss`
- `dx-mini/miniprogram/components/dx-search-bar/index.ts`
- `dx-mini/miniprogram/pages/games/search/search.json`
- `dx-mini/miniprogram/pages/games/search/search.wxml`
- `dx-mini/miniprogram/pages/games/search/search.wxss`
- `dx-mini/miniprogram/pages/games/search/search.ts`
- `dx-mini/miniprogram/pages/games/search/history.ts`

**dx-mini (modify):**
- `dx-mini/scripts/build-icons.mjs` — extend `ICONS` array
- `dx-mini/miniprogram/components/dx-icon/icons.ts` — regenerated
- `dx-mini/miniprogram/app.json` — register search page
- `dx-mini/miniprogram/pages/home/home.json` — add `dx-search-bar` to `usingComponents`
- `dx-mini/miniprogram/pages/home/home.wxml` — add pinned launcher
- `dx-mini/miniprogram/pages/home/home.ts` — add `compactRevealed` data, `onReady`, `onPageScroll`; redirect `goSearch`
- `dx-mini/miniprogram/pages/games/games.json` — add `dx-search-bar`
- `dx-mini/miniprogram/pages/games/games.wxml` — wrap tabs in `.sticky-tabs`, add launcher
- `dx-mini/miniprogram/pages/games/games.wxss` — adjust padding-top, add `.sticky-tabs`
- `dx-mini/miniprogram/pages/games/games.ts` — add `goSearch`

**Note on no test framework in dx-mini:** dx-mini ships zero unit tests today. Verification gates for mini changes are: (1) `npm run build:icons` (validates every literal `<dx-icon name="...">` is declared), (2) WeChat DevTools console must show no errors when each modified page opens, (3) manual smoke tests at the end. There is no `tsc` script in `package.json` — the WeChat DevTools project compiles TS internally and surfaces type errors in the Console panel.

---

## Phase 1 — dx-api new endpoint

### Task 1: Add `GetSearchSuggestions` skeleton + function-existence test

**Files:**
- Modify: `dx-api/app/services/api/game_service.go`
- Modify: `dx-api/app/services/api/game_service_test.go`

- [ ] **Step 1: Add the function-existence test**

Append to `dx-api/app/services/api/game_service_test.go`:

```go
// Pins the expected zero-arg signature at compile time so downstream
// callers (controllers) cannot drift from it silently.
func TestGetSearchSuggestionsFunctionExists(t *testing.T) {
	assert.NotNil(t, GetSearchSuggestions)
	var _ func() ([]string, error) = GetSearchSuggestions
}
```

- [ ] **Step 2: Run the test to verify it fails to compile**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test ./app/services/api/...
```
Expected: build error `undefined: GetSearchSuggestions`.

- [ ] **Step 3: Add a stub `GetSearchSuggestions`**

Append to `dx-api/app/services/api/game_service.go` (after the existing `SearchGames` function — append at the end of the file):

```go
// GetSearchSuggestions is a stub. Real implementation lands in the next task.
func GetSearchSuggestions() ([]string, error) {
	return []string{}, nil
}
```

- [ ] **Step 4: Re-run tests and verify they pass**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/services/api/...
```
Expected: all tests PASS, including `TestGetSearchSuggestionsFunctionExists`.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-api/app/services/api/game_service.go dx-api/app/services/api/game_service_test.go && \
  git commit -m "feat(api): add GetSearchSuggestions stub + signature test"
```

---

### Task 2: Implement `GetSearchSuggestions` body (cache + queries)

**Files:**
- Modify: `dx-api/app/services/api/game_service.go`

- [ ] **Step 1: Verify imports needed**

The function will use `encoding/json`, `time`, plus already-imported `fmt`, `dx-api/app/consts`, `dx-api/app/helpers`, and `github.com/goravel/framework/facades`. Open `dx-api/app/services/api/game_service.go` and confirm the existing imports include `"fmt"`, `"time"`, `"dx-api/app/consts"`, `"dx-api/app/models"`, and `"github.com/goravel/framework/facades"`. Add `"encoding/json"` and `"dx-api/app/helpers"` to the import block if missing.

- [ ] **Step 2: Add constants above the stub**

In `dx-api/app/services/api/game_service.go`, immediately above `func GetSearchSuggestions`, insert:

```go
const (
	searchSuggestionsCacheKey = "dx:search:suggestions"
	searchSuggestionsCacheTTL = time.Hour
	searchSuggestionsMaxItems = 12
	searchSuggestionsTopGames = 8
	searchSuggestionsTopCats  = 4
)
```

- [ ] **Step 3: Replace the stub body**

Replace the stub `GetSearchSuggestions` function with:

```go
// GetSearchSuggestions returns up to 12 search-term chips for the dx-mini
// search page. Cached in Redis for 1h. On cache miss: top published-game
// names by aggregated game_sessions count, plus top game-category names by
// number of published games. Strings deduped; total capped at 12.
func GetSearchSuggestions() ([]string, error) {
	if cached, err := helpers.RedisGet(searchSuggestionsCacheKey); err == nil && cached != "" {
		var out []string
		if jerr := json.Unmarshal([]byte(cached), &out); jerr == nil {
			return out, nil
		}
		// fall through and recompute on bad payload
	}

	type nameRow struct {
		Name string `gorm:"column:name"`
	}

	var topGames []nameRow
	if err := facades.Orm().Query().Raw(`
		SELECT g.name AS name
		FROM games g
		LEFT JOIN game_sessions gs ON gs.game_id = g.id
		WHERE g.status = ? AND g.is_active = TRUE AND g.is_private = FALSE
		GROUP BY g.id, g.name, g.created_at
		ORDER BY COUNT(gs.id) DESC, g.created_at DESC
		LIMIT ?
	`, consts.GameStatusPublished, searchSuggestionsTopGames).Scan(&topGames); err != nil {
		return nil, fmt.Errorf("failed to load top game names: %w", err)
	}

	var topCats []nameRow
	if err := facades.Orm().Query().Raw(`
		SELECT gc.name AS name
		FROM game_categories gc
		INNER JOIN games g ON g.game_category_id = gc.id
		WHERE g.status = ? AND g.is_active = TRUE AND g.is_private = FALSE
		GROUP BY gc.id, gc.name
		ORDER BY COUNT(g.id) DESC
		LIMIT ?
	`, consts.GameStatusPublished, searchSuggestionsTopCats).Scan(&topCats); err != nil {
		return nil, fmt.Errorf("failed to load top categories: %w", err)
	}

	seen := make(map[string]bool, searchSuggestionsMaxItems)
	out := make([]string, 0, searchSuggestionsMaxItems)
	push := func(s string) {
		if s == "" || seen[s] || len(out) >= searchSuggestionsMaxItems {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, r := range topGames {
		push(r.Name)
	}
	for _, r := range topCats {
		push(r.Name)
	}

	if buf, jerr := json.Marshal(out); jerr == nil {
		_ = helpers.RedisSet(searchSuggestionsCacheKey, string(buf), searchSuggestionsCacheTTL)
	}
	return out, nil
}
```

- [ ] **Step 4: Verify compilation**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```
Expected: no output (success).

- [ ] **Step 5: Run tests**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/services/api/...
```
Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-api/app/services/api/game_service.go && \
  git commit -m "feat(api): implement GetSearchSuggestions with 1h Redis cache"
```

---

### Task 3: Add `SearchSuggestions` controller method

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_controller.go`

- [ ] **Step 1: Add the controller method**

In `dx-api/app/http/controllers/api/game_controller.go`, immediately after the `Search` method (around line 60), insert:

```go
// SearchSuggestions returns popular search-term chips for the dx-mini
// search page. Public endpoint, no parameters, no validator needed.
func (c *GameController) SearchSuggestions(ctx contractshttp.Context) contractshttp.Response {
	suggestions, err := services.GetSearchSuggestions()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to load search suggestions")
	}
	return helpers.Success(ctx, suggestions)
}
```

- [ ] **Step 2: Verify compilation**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-api/app/http/controllers/api/game_controller.go && \
  git commit -m "feat(api): add SearchSuggestions controller method"
```

---

### Task 4: Register the new public route

**Files:**
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Register the route inside the existing `/games` group**

In `dx-api/routes/api.go`, locate the public games group at lines 49–52:

```go
router.Prefix("/games").Group(func(games route.Router) {
    games.Get("/", gameController.List)
    games.Get("/search", gameController.Search)
})
```

Replace it with:

```go
router.Prefix("/games").Group(func(games route.Router) {
    games.Get("/", gameController.List)
    games.Get("/search", gameController.Search)
    games.Get("/search-suggestions", gameController.SearchSuggestions)
})
```

- [ ] **Step 2: Verify compilation**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```
Expected: no output.

- [ ] **Step 3: Smoke test the endpoint**

In one terminal, start the server:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .
```

In another terminal:
```bash
curl -s http://localhost:3001/api/games/search-suggestions | head -c 400
```

Expected: a JSON envelope `{"code":0,"message":"ok","data":[...]}` where `data` is an array of strings (possibly empty if the local DB has no published games / sessions). Stop the server with Ctrl+C in the first terminal.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-api/routes/api.go && \
  git commit -m "feat(api): register GET /api/games/search-suggestions"
```

---

## Phase 2 — dx-mini icon assets

### Task 5: Add five new Lucide icons

**Files:**
- Modify: `dx-mini/scripts/build-icons.mjs`
- Modify: `dx-mini/miniprogram/components/dx-icon/icons.ts` (regenerated)

- [ ] **Step 1: Extend the `ICONS` array**

In `dx-mini/scripts/build-icons.mjs`, locate the closing bracket of the `ICONS` array at line 53. Replace the trailing `['arrow-right',    'arrow-right'],` and `]` lines with:

```js
  ['arrow-right',    'arrow-right'],
  ['arrow-left',     'arrow-left'],
  ['x',              'x'],
  ['circle-x',       'circle-x'],
  ['trash-2',        'trash-2'],
  ['search-x',       'search-x'],
]
```

- [ ] **Step 2: Verify lucide-static has every glyph (rare miss)**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && \
  for g in arrow-left x circle-x trash-2 search-x; do \
    test -f node_modules/lucide-static/icons/$g.svg && echo "$g: OK" || echo "$g: MISSING"; \
  done
```

Expected: each line ends with `OK`. If any prints `MISSING`, run `npm install` first; if still missing, that lucide-static version doesn't have that glyph and you must pick the nearest synonym (e.g., replace `search-x` with `circle-x` and update the spec). Do NOT proceed until all five print OK.

- [ ] **Step 3: Regenerate icons.ts**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: the script prints `Wrote 41 icons to miniprogram/components/dx-icon/icons.ts.` (was 36; we added 5).

- [ ] **Step 4: Verify the new entries are present**

```bash
grep -E '"(arrow-left|x|circle-x|trash-2|search-x)":' \
  /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/components/dx-icon/icons.ts | wc -l
```

Expected: `5`.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-mini/scripts/build-icons.mjs dx-mini/miniprogram/components/dx-icon/icons.ts && \
  git commit -m "chore(mini): add arrow-left, x, circle-x, trash-2, search-x icons"
```

---

## Phase 3 — dx-mini shared component `dx-search-bar`

### Task 6: Create the `dx-search-bar` component

**Files:**
- Create: `dx-mini/miniprogram/components/dx-search-bar/index.json`
- Create: `dx-mini/miniprogram/components/dx-search-bar/index.wxml`
- Create: `dx-mini/miniprogram/components/dx-search-bar/index.wxss`
- Create: `dx-mini/miniprogram/components/dx-search-bar/index.ts`

- [ ] **Step 1: Create `index.json`**

Write `dx-mini/miniprogram/components/dx-search-bar/index.json`:

```json
{
  "component": true,
  "usingComponents": {
    "dx-icon": "/components/dx-icon/index"
  }
}
```

- [ ] **Step 2: Create `index.wxml`**

Write `dx-mini/miniprogram/components/dx-search-bar/index.wxml`:

```xml
<view
  class="bar-host {{theme}} {{pinned ? 'pinned' : ''}} {{revealed ? 'revealed' : 'hidden'}}"
  style="--sb-h: {{statusBarHeight}}px; --row-h: {{rowHeight}}px; --pill-right: {{pillRight}}px;"
>
  <view class="bar-spacer"></view>
  <view class="bar-row">
    <view wx:if="{{showCancel}}" class="bar-cancel" bind:tap="onCancel">取消</view>

    <view wx:if="{{mode === 'launcher'}}" class="bar-pill" bind:tap="onTap">
      <dx-icon name="search" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
      <text class="bar-placeholder">{{placeholder}}</text>
    </view>

    <view wx:else class="bar-pill bar-pill-input">
      <slot></slot>
    </view>
  </view>
</view>
```

- [ ] **Step 3: Create `index.wxss`**

Write `dx-mini/miniprogram/components/dx-search-bar/index.wxss`:

```css
.bar-host {
  position: relative;
}

.bar-host.pinned {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 90;
  background: var(--bg-page, #ffffff);
  transition: opacity 200ms ease, transform 200ms ease;
}

.bar-host.pinned.hidden {
  opacity: 0;
  transform: translateY(-4px);
  pointer-events: none;
}

.bar-host.pinned.revealed {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}

.bar-spacer {
  height: var(--sb-h, 20px);
}

.bar-row {
  height: var(--row-h, 40px);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 var(--pill-right, 102px) 0 12px;
  box-sizing: border-box;
}

.bar-cancel {
  font-size: 14px;
  color: var(--text-primary, #1a1a1a);
  padding-right: 4px;
  flex-shrink: 0;
}

.bar-host.dark .bar-cancel {
  color: #f5f5f5;
}

.bar-pill {
  flex: 1;
  min-width: 0;
  height: 32px;
  border-radius: 16px;
  background: #ffffff;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
  box-sizing: border-box;
}

.bar-pill-input {
  padding: 0 8px 0 12px;
}

.bar-placeholder {
  font-size: 13px;
  color: #9ca3af;
}

.bar-host.dark .bar-pill {
  background: #2c2c2e;
}

.bar-host.dark .bar-placeholder {
  color: #6b7280;
}
```

- [ ] **Step 4: Create `index.ts`**

Write `dx-mini/miniprogram/components/dx-search-bar/index.ts`:

```ts
Component({
  options: {
    addGlobalClass: true,
    multipleSlots: false,
  },
  properties: {
    theme: { type: String, value: 'light' },
    pinned: { type: Boolean, value: true },
    revealed: { type: Boolean, value: true },
    placeholder: { type: String, value: '搜索课程' },
    mode: { type: String, value: 'launcher' },
    showCancel: { type: Boolean, value: false },
  },
  data: {
    statusBarHeight: 20,
    rowHeight: 40,
    pillRight: 102,
  },
  lifetimes: {
    attached() {
      const sys = wx.getSystemInfoSync()
      const cap = wx.getMenuButtonBoundingClientRect()
      const statusBarHeight = sys.statusBarHeight || 20
      const rowHeight = Math.max(40, (cap.bottom - statusBarHeight) + 8)
      const pillRight = Math.max(102, sys.windowWidth - cap.left + 8)
      this.setData({ statusBarHeight, rowHeight, pillRight })
    },
  },
  methods: {
    onTap() {
      this.triggerEvent('tap', {})
    },
    onCancel() {
      this.triggerEvent('cancel', {})
    },
  },
})
```

Notes for the implementer:
- No `?.` / `??` per project rules (use `||` and explicit checks).
- `Component({methods})` exhibits a known typed-`this` quirk in `miniprogram-api-typings@2.8.3` — the existing codebase tolerates these specific errors; **do not** add new tsc errors beyond that pattern.
- The `mode` property name here is component-local and never bound from page data via `{{mode}}` — pages pass it as a literal attribute (`mode="input"` or default `mode="launcher"`).

- [ ] **Step 5: Verify icon-declaration scan still passes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: `Wrote 41 icons to ...`. The scanner would error if any `<dx-icon name="...">` literal in the new WXML referenced an undeclared icon. We use `name="search"`, which is declared.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-mini/miniprogram/components/dx-search-bar && \
  git commit -m "feat(mini): add dx-search-bar shared component"
```

---

## Phase 4 — dx-mini new search page

### Task 7: Register the search page route

**Files:**
- Modify: `dx-mini/miniprogram/app.json`

- [ ] **Step 1: Add the page to the `pages` array**

In `dx-mini/miniprogram/app.json`, locate the line `"pages/games/favorites/favorites",` (line 8). Insert a new line directly below it:

```json
    "pages/games/search/search",
```

The full `pages` array (lines 2–26 after edit) should read:

```json
  "pages": [
    "pages/login/login",
    "pages/home/home",
    "pages/games/games",
    "pages/games/detail/detail",
    "pages/games/play/play",
    "pages/games/favorites/favorites",
    "pages/games/search/search",
    "pages/leaderboard/leaderboard",
    "pages/learn/learn",
    "pages/learn/mastered/mastered",
    "pages/learn/unknown/unknown",
    "pages/learn/review/review",
    "pages/me/me",
    "pages/me/profile-edit/profile-edit",
    "pages/me/notices/notices",
    "pages/me/groups/groups",
    "pages/me/groups-detail/groups-detail",
    "pages/me/invite/invite",
    "pages/me/redeem/redeem",
    "pages/me/purchase/purchase",
    "pages/me/study/study",
    "pages/me/tasks/tasks",
    "pages/me/community/community",
    "pages/me/feedback/feedback"
  ],
```

- [ ] **Step 2: Do NOT commit yet**

The page file needs to exist before this commit, otherwise WeChat DevTools will raise `pages/games/search/search not found`. Continue to Task 8 first.

---

### Task 8: Create the search page history helper

**Files:**
- Create: `dx-mini/miniprogram/pages/games/search/history.ts`

- [ ] **Step 1: Create the file**

Write `dx-mini/miniprogram/pages/games/search/history.ts`:

```ts
const KEY = 'dx_recent_searches'
const MAX = 10

export function loadHistory(): string[] {
  const raw = wx.getStorageSync(KEY)
  if (!Array.isArray(raw)) return []
  const arr = raw as unknown[]
  const out: string[] = []
  for (let i = 0; i < arr.length && out.length < MAX; i++) {
    const v = arr[i]
    if (typeof v === 'string' && v.length > 0) out.push(v)
  }
  return out
}

export function pushHistory(term: string): string[] {
  const trimmed = (term || '').trim()
  if (!trimmed) return loadHistory()
  const cur = loadHistory().filter((t) => t !== trimmed)
  const next = [trimmed, ...cur].slice(0, MAX)
  wx.setStorageSync(KEY, next)
  return next
}

export function clearHistory(): void {
  wx.removeStorageSync(KEY)
}
```

Notes for the implementer:
- No `?.` / `??` — `(term || '').trim()` and explicit `Array.isArray` check.
- Validates per-element type via `typeof v === 'string'` so a corrupt storage payload (e.g. someone manually wrote a number) doesn't crash the page.

---

### Task 9: Create the search page boilerplate (json + wxss + ts + wxml)

**Files:**
- Create: `dx-mini/miniprogram/pages/games/search/search.json`
- Create: `dx-mini/miniprogram/pages/games/search/search.wxss`
- Create: `dx-mini/miniprogram/pages/games/search/search.ts`
- Create: `dx-mini/miniprogram/pages/games/search/search.wxml`

- [ ] **Step 1: Create `search.json`**

Write `dx-mini/miniprogram/pages/games/search/search.json`:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-image": "@vant/weapp/image/index",
    "dx-icon": "/components/dx-icon/index",
    "dx-search-bar": "/components/dx-search-bar/index"
  }
}
```

- [ ] **Step 2: Create `search.wxss`**

Write `dx-mini/miniprogram/pages/games/search/search.wxss`:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  padding-bottom: 100rpx;
}

/* Body sits below the fixed search bar.
   The exact px value is set via inline style by the page using statusBarHeight. */
.body {
  /* fallback if inline style does not load */
  padding-top: calc(var(--status-bar-height, 20px) + 40px);
}

/* Slot content (wraps icon + input + clear-X inside the bar pill) */
.input-wrap {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  color: var(--text-primary, #1a1a1a);
  background: transparent;
  border: none;
  outline: none;
  height: 24px;
  line-height: 24px;
}

.search-input-placeholder {
  color: #9ca3af;
}

.page-container.dark .search-input {
  color: #f5f5f5;
}

.clear-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Section: 最近搜索 / 搜索发现 */

.section {
  padding: 16px 16px 0;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.page-container.dark .section-title {
  color: #f5f5f5;
}

.section-action {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  padding: 6px 12px;
  border-radius: 14px;
  background: var(--bg-card, #f3f4f6);
  border: 1px solid var(--border-color, transparent);
}

.page-container.dark .chip {
  background: #2c2c2e;
}

.chip-text {
  font-size: 13px;
  color: var(--text-primary, #1a1a1a);
}

.page-container.dark .chip-text {
  color: #f5f5f5;
}

.chip-discover {
  background: rgba(13, 148, 136, 0.08);
}

.page-container.dark .chip-discover {
  background: rgba(20, 184, 166, 0.16);
}

/* Idle hint */

.hint {
  display: flex;
  justify-content: center;
  padding: 24px;
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
}

/* Empty-results state */

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 64px 24px 0;
}

.empty-title {
  font-size: 14px;
  color: var(--text-primary, #1a1a1a);
}

.page-container.dark .empty-title {
  color: #f5f5f5;
}

.empty-sub {
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}

/* Game grid (mirrors pages/games/games.wxss to keep result cards consistent
   with the courses page). Keep these in lockstep manually if courses is restyled. */

.game-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
}

.game-card {
  width: calc(50% - 6px);
  background: var(--bg-card);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.game-cover {
  height: 120px;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-placeholder {
  width: 100%;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-page);
}

.game-info {
  padding: 10px;
}

.game-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
}

.meta-text,
.meta-dot {
  font-size: 11px;
  color: var(--text-secondary);
}
```

- [ ] **Step 3: Create `search.ts`**

Write `dx-mini/miniprogram/pages/games/search/search.ts`:

```ts
import { api, PaginatedData } from '../../../utils/api'
import { loadHistory, pushHistory, clearHistory } from './history'

type Mode = 'idle' | 'loading' | 'results' | 'empty'

interface GameCardData {
  id: string
  name: string
  description: string | null
  mode: string
  coverUrl: string | null
  categoryName: string | null
  levelCount: number
  author: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark'; userId: string } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    autoFocus: true,
    query: '',
    mode: 'idle' as Mode,
    recents: [] as string[],
    suggestions: [] as string[],
    suggestionsLoading: false,
    games: [] as GameCardData[],
    nextCursor: '',
    hasMore: false,
    loadingMore: false,
  },

  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight,
      recents: loadHistory(),
    })
    this.loadSuggestions()
  },

  onShow() {
    this.setData({ theme: app.globalData.theme })
  },

  onReachBottom() {
    if (this.data.mode !== 'results') return
    if (!this.data.hasMore || this.data.loadingMore) return
    this.loadMore()
  },

  async loadSuggestions() {
    this.setData({ suggestionsLoading: true })
    try {
      const items = await api.get<string[]>('/api/games/search-suggestions')
      this.setData({
        suggestions: Array.isArray(items) ? items : [],
        suggestionsLoading: false,
      })
    } catch {
      // Suggestions are non-critical; fall back to empty list silently.
      this.setData({ suggestions: [], suggestionsLoading: false })
    }
  },

  onInput(e: WechatMiniprogram.Input) {
    const value = (e.detail as { value: string }).value
    const next: { query: string; mode?: Mode } = { query: value }
    if (value.trim() === '' && (this.data.mode === 'results' || this.data.mode === 'empty')) {
      next.mode = 'idle'
    }
    this.setData(next)
  },

  onClear() {
    this.setData({ query: '', mode: 'idle', autoFocus: false })
    // Re-trigger focus on next tick.
    setTimeout(() => {
      this.setData({ autoFocus: true })
    }, 0)
  },

  onSubmit(e: WechatMiniprogram.Input) {
    const raw = (e.detail as { value: string }).value
    const term = (raw || this.data.query || '').trim()
    if (!term) return
    this.runSearch(term)
  },

  onChipTap(e: WechatMiniprogram.TouchEvent) {
    const term = String(e.currentTarget.dataset['term'] || '').trim()
    if (!term) return
    this.runSearch(term)
  },

  async runSearch(term: string) {
    const recents = pushHistory(term)
    this.setData({
      query: term,
      mode: 'loading',
      recents,
      games: [],
      nextCursor: '',
      hasMore: false,
    })
    try {
      const qs = `q=${encodeURIComponent(term)}&limit=20`
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: res.items,
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        mode: res.items.length > 0 ? 'results' : 'empty',
      })
    } catch {
      this.setData({ mode: 'idle' })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  async loadMore() {
    if (this.data.loadingMore) return
    this.setData({ loadingMore: true })
    try {
      const qs = `q=${encodeURIComponent(this.data.query)}&limit=20&cursor=${encodeURIComponent(this.data.nextCursor)}`
      const res = await api.get<PaginatedData<GameCardData>>(`/api/games?${qs}`)
      this.setData({
        games: [...this.data.games, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
        loadingMore: false,
      })
    } catch {
      this.setData({ loadingMore: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  onClearHistoryTap() {
    wx.showModal({
      title: '清除最近搜索?',
      confirmText: '清除',
      cancelText: '取消',
      success: (res) => {
        if (res.confirm) {
          clearHistory()
          this.setData({ recents: [] })
        }
      },
    })
  },

  onCancel() {
    const pages = getCurrentPages()
    if (pages.length > 1) {
      wx.navigateBack({ delta: 1 })
    } else {
      wx.switchTab({ url: '/pages/games/games' })
    }
  },

  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = String(e.currentTarget.dataset['id'] || '')
    if (!id) return
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
})
```

Notes for the implementer:
- No `?.` / `??` anywhere. `(raw || this.data.query || '').trim()` is the canonical pattern.
- `setTimeout` is provided by the WeChat runtime (no Node typings needed).
- The `WechatMiniprogram.Input` type has `e.detail.value` typed as string; the explicit cast `(e.detail as { value: string }).value` matches the project's existing pattern (see `pages/me/feedback/feedback.ts` for similar inputs).

- [ ] **Step 4: Create `search.wxml`**

Write `dx-mini/miniprogram/pages/games/search/search.wxml`:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <dx-search-bar
      theme="{{theme}}"
      pinned="{{true}}"
      revealed="{{true}}"
      mode="input"
      show-cancel="{{true}}"
      bind:cancel="onCancel"
    >
      <view class="input-wrap">
        <dx-icon name="search" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
        <input
          class="search-input"
          type="text"
          confirm-type="search"
          focus="{{autoFocus}}"
          value="{{query}}"
          placeholder="搜索课程"
          placeholder-class="search-input-placeholder"
          bindinput="onInput"
          bindconfirm="onSubmit"
        />
        <view wx:if="{{query.length > 0}}" class="clear-btn" bind:tap="onClear">
          <dx-icon name="circle-x" size="16px" color="{{theme === 'dark' ? '#6b7280' : '#9ca3af'}}" />
        </view>
      </view>
    </dx-search-bar>

    <view class="body" style="padding-top: calc({{statusBarHeight}}px + 40px);">

      <block wx:if="{{mode === 'idle'}}">

        <view wx:if="{{recents.length > 0}}" class="section">
          <view class="section-head">
            <text class="section-title">最近搜索</text>
            <view class="section-action" bind:tap="onClearHistoryTap">
              <dx-icon name="trash-2" size="16px" color="{{theme === 'dark' ? '#9ca3af' : '#6b7280'}}" />
            </view>
          </view>
          <view class="chip-row">
            <view
              wx:for="{{recents}}"
              wx:key="*this"
              class="chip"
              data-term="{{item}}"
              bind:tap="onChipTap"
            >
              <text class="chip-text">{{item}}</text>
            </view>
          </view>
        </view>

        <view wx:if="{{suggestions.length > 0}}" class="section">
          <view class="section-head">
            <text class="section-title">搜索发现</text>
          </view>
          <view class="chip-row">
            <view
              wx:for="{{suggestions}}"
              wx:key="*this"
              class="chip chip-discover"
              data-term="{{item}}"
              bind:tap="onChipTap"
            >
              <text class="chip-text">{{item}}</text>
            </view>
          </view>
        </view>

        <view wx:if="{{recents.length === 0 && suggestions.length === 0 && !suggestionsLoading}}" class="hint">
          <text>搜一搜你感兴趣的课程</text>
        </view>
        <view wx:if="{{suggestionsLoading}}" class="hint">
          <van-loading size="20px" color="#0d9488" />
        </view>
      </block>

      <view wx:if="{{mode === 'loading'}}" class="hint">
        <van-loading size="20px" color="#0d9488" />
      </view>

      <block wx:if="{{mode === 'results'}}">
        <view class="game-grid">
          <view
            wx:for="{{games}}"
            wx:key="id"
            class="game-card"
            data-id="{{item.id}}"
            bind:tap="goDetail"
          >
            <view class="game-cover">
              <van-image
                wx:if="{{item.coverUrl}}"
                src="{{item.coverUrl}}"
                width="100%"
                height="120px"
                fit="cover"
                radius="8px 8px 0 0"
              />
              <view wx:else class="cover-placeholder">
                <dx-icon name="book-open" size="28px" color="#9ca3af" />
              </view>
            </view>
            <view class="game-info">
              <text class="game-name">{{item.name}}</text>
              <view class="game-meta">
                <text class="meta-text">{{item.levelCount}}关</text>
                <text class="meta-dot">·</text>
                <text class="meta-text">{{item.mode}}</text>
              </view>
            </view>
          </view>
        </view>
        <view wx:if="{{loadingMore}}" class="hint">
          <van-loading size="20px" color="#0d9488" />
        </view>
      </block>

      <view wx:if="{{mode === 'empty'}}" class="empty-state">
        <dx-icon name="search-x" size="48px" color="{{theme === 'dark' ? '#374151' : '#d1d5db'}}" />
        <text class="empty-title">没找到相关课程</text>
        <text class="empty-sub">换个关键词试试</text>
      </view>

    </view>
  </view>
</van-config-provider>
```

- [ ] **Step 5: Verify icon scan + WeChat compile**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: `Wrote 41 icons to ...` with no error. The scanner verifies every literal `<dx-icon name="...">` in `search.wxml` (`search`, `circle-x`, `trash-2`, `search-x`, `book-open`) is declared in `ICONS`. All five are present.

Then open WeChat DevTools, ensure the project at `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/` compiles without errors in the Console panel. The simulator should launch with no `pages/games/search/search not found` error.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-mini/miniprogram/app.json \
          dx-mini/miniprogram/pages/games/search && \
  git commit -m "feat(mini): add dedicated 搜索课程 search page"
```

---

## Phase 5 — dx-mini home page wiring

### Task 10: Wire `dx-search-bar` into the home page (Taobao reveal)

**Files:**
- Modify: `dx-mini/miniprogram/pages/home/home.json`
- Modify: `dx-mini/miniprogram/pages/home/home.wxml`
- Modify: `dx-mini/miniprogram/pages/home/home.ts`

- [ ] **Step 1: Register the component**

In `dx-mini/miniprogram/pages/home/home.json`, replace the file contents with:

```json
{
  "navigationStyle": "custom",
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-skeleton": "@vant/weapp/skeleton/index",
    "dx-icon": "/components/dx-icon/index",
    "dx-search-bar": "/components/dx-search-bar/index",
    "home-why-different": "/components/home/why-different/index",
    "home-features": "/components/home/features/index",
    "home-ai-features": "/components/home/ai-features/index",
    "home-learning-loop": "/components/home/learning-loop/index",
    "home-community": "/components/home/community/index",
    "home-membership": "/components/home/membership/index"
  }
}
```

(The only change is one new key `"dx-search-bar"`.)

- [ ] **Step 2: Add the launcher to home.wxml**

In `dx-mini/miniprogram/pages/home/home.wxml`, locate line 2 (the opening `<view class="page-container ...">`). Insert a new line **immediately after** line 2 (i.e., as the first child of `.page-container`, before the `<!-- Teal header band -->` comment):

```xml
    <dx-search-bar
      theme="{{theme}}"
      pinned="{{true}}"
      revealed="{{compactRevealed}}"
      bind:tap="goSearch"
    />
```

The full top of `home.wxml` should now read:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">
    <dx-search-bar
      theme="{{theme}}"
      pinned="{{true}}"
      revealed="{{compactRevealed}}"
      bind:tap="goSearch"
    />

    <!-- Teal header band -->
    <view class="teal-wrap">
      ...
```

The existing `.teal-wrap` (greeting, badges, in-hero search-box) stays exactly as it is. The in-hero `<view class="search-box" bind:tap="goSearch">` will continue to work — it now navigates to the new search page because `goSearch` body is updated below.

- [ ] **Step 3: Add reveal-on-scroll logic to home.ts**

In `dx-mini/miniprogram/pages/home/home.ts`:

(a) Inside the `data: { ... }` block (currently lines 48–58), append two new fields. After the existing `vipDueAt: '' as string,` line, add:

```ts
    compactRevealed: false,
    heroBottomPx: 0,
```

(b) Add an `onReady` lifecycle method directly after the existing `onLoad()` method (around line 63). Insert:

```ts
  onReady() {
    wx.createSelectorQuery()
      .in(this)
      .select('.search-row')
      .boundingClientRect((rect) => {
        if (rect && typeof rect.bottom === 'number') {
          this.setData({ heroBottomPx: rect.bottom })
        }
      })
      .exec()
  },
```

(c) Add `onPageScroll` directly after `onReady`:

```ts
  onPageScroll(e: WechatMiniprogram.Page.IPageScrollOption) {
    const threshold = this.data.heroBottomPx
    if (threshold <= 0) return
    const shouldReveal = e.scrollTop >= threshold
    if (shouldReveal !== this.data.compactRevealed) {
      this.setData({ compactRevealed: shouldReveal })
    }
  },
```

(d) Replace the existing `goSearch` method body (line 97). Change:

```ts
  goSearch() { wx.navigateTo({ url: '/pages/games/games' }) },
```

To:

```ts
  goSearch() { wx.navigateTo({ url: '/pages/games/search/search' }) },
```

- [ ] **Step 4: Verify icon scan still passes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: `Wrote 41 icons to ...`.

- [ ] **Step 5: Open WeChat DevTools and verify home page**

In WeChat DevTools, switch to the simulator's home page. Expected:
- Top of viewport shows only the WeChat capsule, no visible search pill (because `compactRevealed=false` on cold open).
- Scroll down past the in-hero `搜索课程` pill — the pinned launcher fades + slides in (200ms).
- Scroll back up past the threshold — it fades + slides out.
- Tapping either the pinned launcher or the in-hero pill navigates to the new search page (keyboard pops up).
- Console panel shows no errors.

If the launcher overlaps the capsule on a non-iPhone device, inspect `bar-host.pinned` in DevTools' wxml inspector to verify `--row-h` and `--pill-right` are reasonable values (>= 40 and >= 80 respectively).

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-mini/miniprogram/pages/home && \
  git commit -m "feat(mini): pin search launcher to home with scroll-driven reveal"
```

---

## Phase 6 — dx-mini courses page wiring

### Task 11: Pin search bar + sticky tabs on the courses page

**Files:**
- Modify: `dx-mini/miniprogram/pages/games/games.json`
- Modify: `dx-mini/miniprogram/pages/games/games.wxml`
- Modify: `dx-mini/miniprogram/pages/games/games.wxss`
- Modify: `dx-mini/miniprogram/pages/games/games.ts`

- [ ] **Step 1: Register `dx-search-bar` in games.json**

In `dx-mini/miniprogram/pages/games/games.json`, replace the contents with:

```json
{
  "navigationStyle": "custom",
  "enablePullDownRefresh": true,
  "usingComponents": {
    "van-config-provider": "@vant/weapp/config-provider/index",
    "van-tabs": "@vant/weapp/tabs/index",
    "van-tab": "@vant/weapp/tab/index",
    "dx-icon": "/components/dx-icon/index",
    "van-loading": "@vant/weapp/loading/index",
    "van-empty": "@vant/weapp/empty/index",
    "van-image": "@vant/weapp/image/index",
    "dx-search-bar": "/components/dx-search-bar/index"
  }
}
```

(The only change is the new `"dx-search-bar"` entry.)

- [ ] **Step 2: Update games.wxml**

In `dx-mini/miniprogram/pages/games/games.wxml`, replace the entire file contents with:

```xml
<van-config-provider theme="{{theme}}">
  <view class="page-container {{theme}}" style="--status-bar-height: {{statusBarHeight}}px">

    <dx-search-bar
      theme="{{theme}}"
      pinned="{{true}}"
      revealed="{{true}}"
      bind:tap="goSearch"
    />

    <view class="sticky-tabs" style="top: calc({{statusBarHeight}}px + 40px);">
      <van-tabs
        active="{{activeCategoryId}}"
        bind:click="onCategoryChange"
        color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        background="{{theme === 'dark' ? '#1c1c1e' : '#ffffff'}}"
        title-active-color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
        scrollable
      >
        <van-tab
          wx:for="{{categories}}"
          wx:key="id"
          title="{{item.name}}"
          name="{{item.id}}"
        />
      </van-tabs>
    </view>

    <!-- Favorites button -->
    <view class="fav-bar" bind:tap="goFavorites">
      <dx-icon name="star" size="16px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" />
      <text class="fav-text">收藏的课程</text>
    </view>

    <!-- Game grid -->
    <view class="game-grid">
      <view
        wx:for="{{games}}"
        wx:key="id"
        class="game-card"
        data-id="{{item.id}}"
        bind:tap="goDetail"
      >
        <view class="game-cover">
          <van-image
            wx:if="{{item.coverUrl}}"
            src="{{item.coverUrl}}"
            width="100%"
            height="120px"
            fit="cover"
            radius="8px 8px 0 0"
          />
          <view wx:else class="cover-placeholder">
            <dx-icon name="book-open" size="28px" color="#9ca3af" />
          </view>
        </view>
        <view class="game-info">
          <text class="game-name">{{item.name}}</text>
          <view class="game-meta">
            <text class="meta-text">{{item.levelCount}}关</text>
            <text class="meta-dot">·</text>
            <text class="meta-text">{{item.mode}}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Loading / empty -->
    <view wx:if="{{loading}}" class="load-more">
      <van-loading size="20px" color="#0d9488" />
    </view>
    <van-empty wx:if="{{!loading && games.length === 0}}" description="暂无课程" />
  </view>
</van-config-provider>
```

Diff summary: added the `<dx-search-bar>` line as the first child of `.page-container`, wrapped the existing `<van-tabs>` block in a `<view class="sticky-tabs" style="top: calc({{statusBarHeight}}px + 40px);">` element. Everything else is byte-identical.

- [ ] **Step 3: Update games.wxss**

In `dx-mini/miniprogram/pages/games/games.wxss`, locate lines 1–6 (the `.page-container` rule). Replace the rule with:

```css
.page-container {
  min-height: 100vh;
  background: var(--bg-page);
  /* status-bar + 40px launcher + 88rpx tabs row */
  padding-top: calc(var(--status-bar-height, 20px) + 40px + 88rpx);
  padding-bottom: 100rpx;
}

.sticky-tabs {
  position: fixed;
  /* `top` is set inline by games.wxml: calc(<statusBarHeight>px + 40px) */
  left: 0;
  right: 0;
  z-index: 80;
  background: var(--bg-page);
}
```

The remaining rules (`.fav-bar`, `.game-grid`, etc.) are unchanged.

- [ ] **Step 4: Add `goSearch` to games.ts**

In `dx-mini/miniprogram/pages/games/games.ts`, locate the closing `})` of the `Page({ ... })` call (line 80). Immediately before that closing `)`, append a `goSearch` method. The end of the file should read:

```ts
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
  goFavorites() {
    wx.navigateTo({ url: '/pages/games/favorites/favorites' })
  },
  goSearch() {
    wx.navigateTo({ url: '/pages/games/search/search' })
  },
})
```

- [ ] **Step 5: Verify icon scan still passes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
```

Expected: `Wrote 41 icons to ...`.

- [ ] **Step 6: Verify in WeChat DevTools**

Switch the simulator to the courses tab. Expected:
- Pinned launcher visible immediately at top, alongside the WeChat capsule.
- Tabs row pinned directly below the launcher.
- Favorites bar + game grid scroll under both fixed layers (no content peeks above the fixed bars).
- Pull-to-refresh still works (existing `onPullDownRefresh` is untouched).
- Tapping the launcher navigates to `/pages/games/search/search`.
- No console errors.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
  git add dx-mini/miniprogram/pages/games/games.json \
          dx-mini/miniprogram/pages/games/games.wxml \
          dx-mini/miniprogram/pages/games/games.wxss \
          dx-mini/miniprogram/pages/games/games.ts && \
  git commit -m "feat(mini): pin search launcher + sticky tabs on courses page"
```

---

## Phase 7 — End-to-end smoke test

### Task 12: Final manual QA in WeChat DevTools

**Files:** none (verification only).

This task documents the manual smoke test from the spec's "Acceptance / Smoke Test Plan" section. It does not produce a commit. If any step fails, file a follow-up task and stop.

- [ ] **Step 1: Cold-open home (iPhone notch profile)**

  In WeChat DevTools, set simulator to `iPhone 14 Pro`. Reload the project. Switch to the home tab. Expected: only WeChat capsule visible at top, no search pill in viewport.

- [ ] **Step 2: Verify scroll-reveal**

  Scroll down in the simulator. As soon as the in-hero `搜索课程` pill leaves the viewport, the pinned launcher should fade + slide in over ~200ms. Scroll back up; it should fade + slide out at the same threshold.

- [ ] **Step 3: Verify Android profile**

  Switch simulator to `Android 6.7"`. Reload. Confirm same reveal behavior; capsule alignment looks correct (no overlap, no >12px gap).

- [ ] **Step 4: Verify courses page**

  Switch to the 课程 tab. Pinned launcher visible immediately; tabs pinned beneath it. Pull down — refresh fires; the fixed layers stay where they are. Switch tabs — the active tab indicator moves; grid below scrolls under both fixed layers.

- [ ] **Step 5: Tap into search page**

  Tap any of: home pinned launcher, home in-hero pill, courses pinned launcher. Each navigates to `/pages/games/search/search`. Keyboard pops up automatically.

- [ ] **Step 6: Suggestion idle state**

  On first visit, 最近搜索 section is hidden (history empty). 搜索发现 chips render from API. If neither populates, the hint `搜一搜你感兴趣的课程` shows.

- [ ] **Step 7: Tap a 搜索发现 chip**

  Tap any chip. Input fills with the chip text; mode flips to `loading`; result grid appears (or empty-state if the catalog has no matching games for that term).

- [ ] **Step 8: Type + submit**

  Tap × to clear input → mode returns to `idle`. Type `vocab`, tap the keyboard's 搜索 button. Mode flips to `loading` then `results`. Scroll to the bottom → next page loads (`onReachBottom`). Tap a card → navigates to `/pages/games/detail/detail?id=<id>`.

- [ ] **Step 9: Recent search persistence**

  Navigate back to the search page (via courses tab → pinned launcher). 最近搜索 section is now visible with `vocab` chip on top.

- [ ] **Step 10: Clear history**

  Tap the trash icon next to 最近搜索. Confirm modal `清除最近搜索?`. Confirm. Section disappears.

- [ ] **Step 11: Empty / whitespace submit**

  Clear the input. Type three spaces. Tap 搜索. Expected: no API call (verify in WeChat DevTools' Network panel — no new request to `/api/games`); mode stays `idle`; input keeps focus.

- [ ] **Step 12: Network failure**

  In WeChat DevTools, set Network to `Offline`. Type `error-test` → 搜索. Expected: toast `加载失败`; mode falls back to `idle`. Restore Network to `WIFI` or default.

- [ ] **Step 13: Cancel / back navigation**

  On the search page, tap 取消 (top left). Returns to the entry page (home or courses, whichever you came from).

- [ ] **Step 14: Dark mode**

  In the simulator, change Settings → 深色模式 → ON. All three pages re-render with dark surfaces; pill, chip, and input text/placeholder remain readable.

- [ ] **Step 15: dx-api endpoint smoke test**

  Confirm the new endpoint is healthy:
  ```bash
  curl -s http://localhost:3001/api/games/search-suggestions | head -c 400
  ```
  Expected: `{"code":0,"message":"ok","data":[...]}`. Re-run within 1h — same payload (cache hit); first call after `helpers.RedisDel(\"dx:search:suggestions\")` re-computes.

- [ ] **Step 16: Final test sweep**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
  cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npm run build:icons
  ```
  Expected: dx-api tests all PASS; icons build prints `Wrote 41 icons to ...`.

- [ ] **Step 17: Memory note** (only if a real-device run was attempted and failed)

  Per `feedback_dx_mini_no_optional_chaining.md` and `project_wechat_devtools_realdevice_bug.md`: if 真机调试 misbehaves, fall back to 预览 + 小程序助手. Don't introduce `?.` / `??` to "fix" any TypeScript issue surfaced.

---

## Notes for the implementer

- **Order matters.** Phase 1 → 2 → 3 → 4 → 5 → 6 → 7 sequentially; each phase's commit lands working software. If you must reorder, Phase 4 (search page) must land before Phase 5 / Phase 6 because both depend on the route being registered AND the page existing.
- **No `?.` / `??`** in any TS or WXML file you touch. Use `||`, explicit null/undefined checks, or `Array.isArray`.
- **No `console.log`** in any code that ships to production. If you need temporary logging while debugging, remove it before the commit.
- **Don't restyle the courses page card.** Keep the inline copy in `search.wxss` byte-identical to `games.wxss` for the `.game-grid`, `.game-card`, `.game-cover`, `.cover-placeholder`, `.game-info`, `.game-name`, `.game-meta`, `.meta-text`, and `.meta-dot` selectors. Drift will be visible.
- **Theme primary colors:** light `#0d9488`, dark `#14b8a6`, dark surface `#1c1c1e`. Pill background light `#ffffff`, dark `#2c2c2e`.
- **WeChat capsule math** is brittle on Android emulators — `wx.getMenuButtonBoundingClientRect()` can return `0` for some fields if called before the page is fully mounted. The component uses `lifetimes.attached` (after the page mounts), and our CSS uses sane fallbacks (`--row-h: 40px`, `--pill-right: 102px`) so a zero result still renders something usable.
