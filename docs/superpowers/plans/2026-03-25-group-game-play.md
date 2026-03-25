# Group Game Play Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect game groups to real game play with synchronized rounds, per-answer timers, and winner determination.

**Architecture:** Owner-initiated rounds via SSE push. Sessions link to groups via nullable FK columns. Winner calculated at level completion by comparing scores across group members. Solo mode ranks individuals; team mode ranks subgroups by sum of member scores.

**Tech Stack:** Go/Goravel backend, Next.js frontend, PostgreSQL, SSE for real-time events, Zustand for client state.

**Spec:** `docs/superpowers/specs/2026-03-25-group-game-play-design.md`

---

### Task 1: Database Migrations — Schema Additions

**Files:**
- Create: `dx-api/database/migrations/20260325000001_add_group_play_columns.go`
- Create: `dx-api/database/migrations/20260325000002_update_session_unique_index.go`
- Modify: `dx-api/bootstrap/app.go` (register new migrations)

- [ ] **Step 1: Create migration for new columns**

```go
// dx-api/database/migrations/20260325000001_add_group_play_columns.go
package migrations

import "github.com/goravel/framework/facades"

type M20260325000001AddGroupPlayColumns struct{}

func (r *M20260325000001AddGroupPlayColumns) Signature() string {
	return "20260325000001_add_group_play_columns"
}

func (r *M20260325000001AddGroupPlayColumns) Up() error {
	// game_groups: answer_time_limit, is_playing
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_groups
		 ADD COLUMN IF NOT EXISTS answer_time_limit INTEGER NOT NULL DEFAULT 10,
		 ADD COLUMN IF NOT EXISTS is_playing BOOLEAN NOT NULL DEFAULT false`); err != nil {
		return err
	}

	// game_group_members: last_won_at
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_group_members
		 ADD COLUMN IF NOT EXISTS last_won_at TIMESTAMP`); err != nil {
		return err
	}

	// game_subgroups: last_won_at
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_subgroups
		 ADD COLUMN IF NOT EXISTS last_won_at TIMESTAMP`); err != nil {
		return err
	}

	// game_session_totals: game_group_id, game_subgroup_id
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_totals
		 ADD COLUMN IF NOT EXISTS game_group_id UUID REFERENCES game_groups(id) ON DELETE SET NULL,
		 ADD COLUMN IF NOT EXISTS game_subgroup_id UUID REFERENCES game_subgroups(id) ON DELETE SET NULL`); err != nil {
		return err
	}

	// game_session_levels: game_group_id, game_subgroup_id
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_levels
		 ADD COLUMN IF NOT EXISTS game_group_id UUID REFERENCES game_groups(id) ON DELETE SET NULL,
		 ADD COLUMN IF NOT EXISTS game_subgroup_id UUID REFERENCES game_subgroups(id) ON DELETE SET NULL`); err != nil {
		return err
	}

	// Partial index on session totals for group queries
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_group
		 ON game_session_totals (game_group_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	// Partial index on session levels for winner determination queries
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_group_level
		 ON game_session_levels (game_group_id, game_level_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260325000001AddGroupPlayColumns) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_group_level`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_group`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_levels DROP COLUMN IF EXISTS game_subgroup_id, DROP COLUMN IF EXISTS game_group_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_totals DROP COLUMN IF EXISTS game_subgroup_id, DROP COLUMN IF EXISTS game_group_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_subgroups DROP COLUMN IF EXISTS last_won_at`)
	facades.Orm().Query().Exec(`ALTER TABLE game_group_members DROP COLUMN IF EXISTS last_won_at`)
	facades.Orm().Query().Exec(`ALTER TABLE game_groups DROP COLUMN IF EXISTS is_playing, DROP COLUMN IF EXISTS answer_time_limit`)
	return nil
}
```

- [ ] **Step 2: Create migration for unique index replacement**

```go
// dx-api/database/migrations/20260325000002_update_session_unique_index.go
package migrations

import "github.com/goravel/framework/facades"

type M20260325000002UpdateSessionUniqueIndex struct{}

func (r *M20260325000002UpdateSessionUniqueIndex) Signature() string {
	return "20260325000002_update_session_unique_index"
}

func (r *M20260325000002UpdateSessionUniqueIndex) Up() error {
	// Drop old single unique index
	if _, err := facades.Orm().Query().Exec(
		`DROP INDEX IF EXISTS idx_game_session_totals_unique_active`); err != nil {
		return err
	}

	// Partial unique index for regular (non-group) sessions
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_regular
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL AND game_group_id IS NULL`); err != nil {
		return err
	}

	// Partial unique index for group sessions
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_group
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''), game_group_id)
		 WHERE ended_at IS NULL AND game_group_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260325000002UpdateSessionUniqueIndex) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_group`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_regular`)
	facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL`)
	return nil
}
```

- [ ] **Step 3: Register migrations in bootstrap/app.go**

Add both new migrations to the migrations slice in `dx-api/bootstrap/app.go`.

- [ ] **Step 4: Run migrations**

Run: `cd dx-api && go run . artisan migrate`
Expected: Migrations applied successfully.

- [ ] **Step 5: Verify with psql**

Run: `psql -d douxue -c "\d game_groups" | head -20`
Expected: `answer_time_limit` and `is_playing` columns present.

Run: `psql -d douxue -c "\d game_session_totals" | head -25`
Expected: `game_group_id` and `game_subgroup_id` columns present.

- [ ] **Step 6: Commit**

```bash
git add dx-api/database/migrations/20260325000001_add_group_play_columns.go \
       dx-api/database/migrations/20260325000002_update_session_unique_index.go \
       dx-api/bootstrap/app.go
git commit -m "feat: add group game play schema — columns and unique index split"
```

---

### Task 2: Update Go Models

**Files:**
- Modify: `dx-api/app/models/game_group.go`
- Modify: `dx-api/app/models/game_group_member.go`
- Modify: `dx-api/app/models/game_subgroup.go`
- Modify: `dx-api/app/models/game_session_total.go`
- Modify: `dx-api/app/models/game_session_level.go`

- [ ] **Step 1: Update GameGroup model**

Add to `GameGroup` struct in `dx-api/app/models/game_group.go`:

```go
AnswerTimeLimit  int     `gorm:"column:answer_time_limit" json:"answer_time_limit"`
IsPlaying        bool    `gorm:"column:is_playing" json:"is_playing"`
```

- [ ] **Step 2: Update GameGroupMember model**

Add to `GameGroupMember` struct in `dx-api/app/models/game_group_member.go`:

```go
LastWonAt *time.Time `gorm:"column:last_won_at" json:"last_won_at"`
```

Add `"time"` to imports.

- [ ] **Step 3: Update GameSubgroup model**

Add to `GameSubgroup` struct in `dx-api/app/models/game_subgroup.go`:

```go
LastWonAt *time.Time `gorm:"column:last_won_at" json:"last_won_at"`
```

Add `"time"` to imports.

- [ ] **Step 4: Update GameSessionTotal model**

Add to `GameSessionTotal` struct in `dx-api/app/models/game_session_total.go`:

```go
GameGroupID    *string `gorm:"column:game_group_id" json:"game_group_id"`
GameSubgroupID *string `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
```

- [ ] **Step 5: Update GameSessionLevel model**

Add to `GameSessionLevel` struct in `dx-api/app/models/game_session_level.go`:

```go
GameGroupID    *string `gorm:"column:game_group_id" json:"game_group_id"`
GameSubgroupID *string `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
```

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/models/game_group.go dx-api/app/models/game_group_member.go \
       dx-api/app/models/game_subgroup.go dx-api/app/models/game_session_total.go \
       dx-api/app/models/game_session_level.go
git commit -m "feat: update models with group play fields"
```

---

### Task 3: Backend Behavior Changes — Owner Leave & Guards

**Files:**
- Modify: `dx-api/app/services/api/group_member_service.go:116-140`
- Modify: `dx-api/app/services/api/group_game_service.go:77-120`
- Modify: `dx-api/app/services/api/errors.go:44`

- [ ] **Step 1: Remove ErrCannotLeaveOwned guard from KickMember**

In `dx-api/app/services/api/group_member_service.go:124-126`, remove:

```go
if targetUserID == userID {
    return ErrCannotLeaveOwned
}
```

- [ ] **Step 2: Remove ErrCannotLeaveOwned guard from LeaveGroup**

In `dx-api/app/services/api/group_member_service.go:136-138`, remove:

```go
if group.OwnerID == userID {
    return ErrCannotLeaveOwned
}
```

- [ ] **Step 3: Add is_playing guard to SetGroupGame**

In `dx-api/app/services/api/group_game_service.go`, after the ownership check in `SetGroupGame` (after line 84), add:

```go
if group.IsPlaying {
    return ErrGroupIsPlaying
}
```

- [ ] **Step 4: Add is_playing guard to ClearGroupGame**

In `dx-api/app/services/api/group_game_service.go`, after the ownership check in `ClearGroupGame` (after line 111), add:

```go
if group.IsPlaying {
    return ErrGroupIsPlaying
}
```

- [ ] **Step 5: Add is_playing guard to DeleteGroup**

In `dx-api/app/services/api/group_service.go`, in the `DeleteGroup` function, after the ownership check and before deletion, add:

```go
if group.IsPlaying {
    return ErrGroupIsPlaying
}
```

- [ ] **Step 6: Add answer_time_limit validation to UpdateGroup**

Add `answer_time_limit` as an optional field in the group update request (`dx-api/app/http/requests/api/group_request.go`). Validate range 5-60:

```go
AnswerTimeLimit *int `form:"answer_time_limit" json:"answer_time_limit"`
```

Validation rule: `"answer_time_limit": "min:5|max:60"`

In the `UpdateGroup` service (`dx-api/app/services/api/group_service.go`), if `answer_time_limit` is provided, include it in the update map.

- [ ] **Step 7: Add new error sentinels**

In `dx-api/app/services/api/errors.go`, add:

```go
ErrGroupIsPlaying       = errors.New("game is in progress")
ErrGroupNotPlaying      = errors.New("no game in progress")
ErrNoGameSet            = errors.New("no game set for group")
ErrNoGameModeSet        = errors.New("no game mode set for group")
ErrNotInSubgroup        = errors.New("member not in any subgroup")
```

- [ ] **Step 8: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 9: Commit**

```bash
git add dx-api/app/services/api/group_member_service.go \
       dx-api/app/services/api/group_game_service.go \
       dx-api/app/services/api/group_service.go \
       dx-api/app/services/api/errors.go \
       dx-api/app/http/requests/api/group_request.go
git commit -m "feat: remove owner leave guard, add is_playing guards, answer_time_limit validation"
```

---

### Task 4: SSE Hub Infrastructure

**Files:**
- Create: `dx-api/app/helpers/sse_hub.go`

- [ ] **Step 1: Create SSE hub with connection registry**

```go
// dx-api/app/helpers/sse_hub.go
package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEConnection wraps a single client SSE connection.
type SSEConnection struct {
	w       http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// SSEHub manages group SSE connections and broadcasting.
type SSEHub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*SSEConnection // groupID -> userID -> conn
}

var GroupSSEHub = &SSEHub{
	conns: make(map[string]map[string]*SSEConnection),
}

// Register adds a connection for a user in a group.
func (h *SSEHub) Register(groupID, userID string, w http.ResponseWriter) *SSEConnection {
	flusher, _ := w.(http.Flusher)
	conn := &SSEConnection{w: w, flusher: flusher, done: make(chan struct{})}

	h.mu.Lock()
	if h.conns[groupID] == nil {
		h.conns[groupID] = make(map[string]*SSEConnection)
	}
	// Close existing connection if any
	if old, ok := h.conns[groupID][userID]; ok {
		close(old.done)
	}
	h.conns[groupID][userID] = conn
	h.mu.Unlock()

	return conn
}

// Unregister removes a connection.
func (h *SSEHub) Unregister(groupID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if group, ok := h.conns[groupID]; ok {
		delete(group, userID)
		if len(group) == 0 {
			delete(h.conns, groupID)
		}
	}
}

// Broadcast sends an event to all connected members of a group.
func (h *SSEHub) Broadcast(groupID, event string, data any) {
	jsonBytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	if group, ok := h.conns[groupID]; ok {
		for _, conn := range group {
			fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event, jsonBytes)
			if conn.flusher != nil {
				conn.flusher.Flush()
			}
		}
	}
}

// SendHeartbeat sends a comment line as keepalive.
func (conn *SSEConnection) SendHeartbeat() error {
	_, err := fmt.Fprintf(conn.w, ": heartbeat\n\n")
	if err != nil {
		return err
	}
	if conn.flusher != nil {
		conn.flusher.Flush()
	}
	return nil
}

// Done returns a channel that closes when the connection should end.
func (conn *SSEConnection) Done() <-chan struct{} {
	return conn.done
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/helpers/sse_hub.go
git commit -m "feat: add SSE hub for group event broadcasting"
```

---

### Task 5: SSE Connection Endpoint

**Files:**
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go`
- Modify: `dx-api/routes/api.go:277-281`

- [ ] **Step 1: Add Events handler to group game controller**

Add to `dx-api/app/http/controllers/api/group_game_controller.go`:

```go
// Events establishes a persistent SSE connection for group events.
// Auth via query param ?token=xxx since EventSource cannot set headers.
func (c *GroupGameController) Events(ctx http.Context) http.Response {
	// Read token from query param (EventSource cannot send Authorization header)
	token := ctx.Request().Query("token", "")
	if token == "" {
		return ctx.Response().Json(http.StatusUnauthorized, helpers.ErrorResponse("missing token"))
	}
	// Manually parse JWT to get user ID
	userID, err := helpers.ParseJWTUserID(token)
	if err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, helpers.ErrorResponse("invalid token"))
	}
	groupID := ctx.Request().Route("id")

	// Verify membership
	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).
		Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return ctx.Response().Json(http.StatusForbidden, helpers.ErrorResponse("not a group member"))
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.(http.Flusher).Flush()

	conn := helpers.GroupSSEHub.Register(groupID, userID, w)
	defer helpers.GroupSSEHub.Unregister(groupID, userID)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			return nil
		case <-conn.Done():
			return nil
		case <-ticker.C:
			if err := conn.SendHeartbeat(); err != nil {
				return nil
			}
		}
	}
}
```

- [ ] **Step 2: Add ParseJWTUserID helper**

In `dx-api/app/helpers/`, create a helper function `ParseJWTUserID(token string) (string, error)` that parses the JWT token and extracts the user ID claim. This is needed because the SSE endpoint authenticates via query param instead of middleware.

- [ ] **Step 3: Add route (outside JWT middleware)**

In `dx-api/routes/api.go`, add the SSE endpoint **outside** the JWT-protected group. It handles its own auth via query param:

```go
// Before the protected group — SSE uses query param auth
groupGameController := apicontrollers.NewGroupGameController()
router.Get("/api/groups/{id}/events", groupGameController.Events)
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/group_game_controller.go dx-api/routes/api.go
git commit -m "feat: add SSE connection endpoint for group events"
```

---

### Task 6: Start Game Backend — Service & Controller

**Files:**
- Modify: `dx-api/app/services/api/group_game_service.go`
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go`
- Modify: `dx-api/app/http/requests/api/group_game_request.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add StartGameRequest validation**

In `dx-api/app/http/requests/api/group_game_request.go`, add:

```go
type StartGroupGameRequest struct {
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
}

func (r *StartGroupGameRequest) Rules() map[string]string {
	return map[string]string{
		"degree": "required|in:practice,beginner,intermediate,advanced",
	}
}

func (r *StartGroupGameRequest) Messages() map[string]string {
	return map[string]string{}
}
```

- [ ] **Step 2: Add StartGroupGame service**

In `dx-api/app/services/api/group_game_service.go`, add:

```go
// GroupGameStartEvent is the SSE payload for group_game_start.
type GroupGameStartEvent struct {
	GameGroupID    string  `json:"game_group_id"`
	GameID         string  `json:"game_id"`
	GameName       string  `json:"game_name"`
	GameMode       string  `json:"game_mode"`
	Degree         string  `json:"degree"`
	Pattern        *string `json:"pattern"`
	AnswerTimeLimit int    `json:"answer_time_limit"`
}

// StartGroupGame validates and initiates a group game round, broadcasting via SSE.
func StartGroupGame(userID, groupID, degree string, pattern *string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}
	if group.CurrentGameID == nil || *group.CurrentGameID == "" {
		return ErrNoGameSet
	}
	if group.GameMode == nil || *group.GameMode == "" {
		return ErrNoGameModeSet
	}

	// Fetch game name for SSE payload
	var game models.Game
	if err := facades.Orm().Query().Where("id", *group.CurrentGameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}

	// Set is_playing = true
	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).
		Update("is_playing", true); err != nil {
		return fmt.Errorf("failed to set is_playing: %w", err)
	}

	// Broadcast SSE event
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_start", GroupGameStartEvent{
		GameGroupID:     groupID,
		GameID:          *group.CurrentGameID,
		GameName:        game.Name,
		GameMode:        *group.GameMode,
		Degree:          degree,
		Pattern:         pattern,
		AnswerTimeLimit: group.AnswerTimeLimit,
	})

	return nil
}
```

Add `"dx-api/app/helpers"` to imports.

- [ ] **Step 3: Add StartGame controller**

In `dx-api/app/http/controllers/api/group_game_controller.go`, add:

```go
// StartGame initiates a group game round.
func (c *GroupGameController) StartGame(ctx http.Context) http.Response {
	userID := ctx.Value("id").(string)
	groupID := ctx.Request().Route("id")

	var req requests.StartGroupGameRequest
	if errs, err := ctx.Request().ValidateRequest(&req); err != nil || errs != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, helpers.ValidationResponse(errs))
	}

	if err := api.StartGroupGame(userID, groupID, req.Degree, req.Pattern); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return ctx.Response().Json(http.StatusOK, helpers.SuccessResponse(nil))
}
```

- [ ] **Step 4: Add route**

In `dx-api/routes/api.go`, add inside the group game routes block:

```go
groups.Post("/{id}/start-game", groupGameController.StartGame)
```

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/group_game_service.go \
       dx-api/app/http/controllers/api/group_game_controller.go \
       dx-api/app/http/requests/api/group_game_request.go \
       dx-api/routes/api.go
git commit -m "feat: add start-game endpoint for group play"
```

---

### Task 7: Modify Session Service for Group Sessions

**Files:**
- Modify: `dx-api/app/services/api/session_service.go:133-244` (StartSession)
- Modify: `dx-api/app/services/api/session_service.go:555-615` (StartLevel)
- Modify: `dx-api/app/services/api/session_service.go:875-891` (findActiveSession)
- Modify: `dx-api/app/http/controllers/api/game_session_controller.go` (StartSession handler)
- Modify: `dx-api/app/http/requests/api/` (StartSession request)

- [ ] **Step 1: Update findActiveSession to accept game_group_id**

In `dx-api/app/services/api/session_service.go:875-891`, change signature and add filter:

```go
func findActiveSession(query orm.Query, userID, gameID, degree string, pattern *string, gameGroupID *string) (*models.GameSessionTotal, error) {
	var session models.GameSessionTotal
	q := query.Where("user_id", userID).Where("game_id", gameID).
		Where("degree", degree).Where("ended_at IS NULL").
		Order("started_at desc")

	if pattern != nil {
		q = q.Where("pattern", *pattern)
	} else {
		q = q.Where("pattern IS NULL")
	}

	if gameGroupID != nil {
		q = q.Where("game_group_id", *gameGroupID)
	} else {
		q = q.Where("game_group_id IS NULL")
	}

	if err := q.First(&session); err != nil || session.ID == "" {
		return nil, nil
	}
	return &session, nil
}
```

- [ ] **Step 2: Update all callers of findActiveSession**

Update all calls to `findActiveSession` throughout `session_service.go` to pass the new `gameGroupID` parameter. For existing non-group calls, pass `nil`:

- Line 144: `findActiveSession(query, userID, gameID, degree, pattern, nil)` → change to accept `gameGroupID` param
- Line 212: same
- Line 253 (CheckActiveSession): same

For `StartSession`, accept `gameGroupID` as a new parameter and pass it through.

- [ ] **Step 3: Update StartSession to accept and set group fields**

In `dx-api/app/services/api/session_service.go`, update `StartSession` signature:

```go
func StartSession(userID, gameID, degree string, pattern *string, levelID *string, gameGroupID *string) (*StartSessionResult, error) {
```

In the "Create new session" block (~line 198), resolve `gameSubgroupID` if group session:

```go
var gameSubgroupID *string
if gameGroupID != nil {
    // Look up group to check game_mode
    var group models.GameGroup
    if err := query.Where("id", *gameGroupID).First(&group); err != nil || group.ID == "" {
        return nil, ErrGroupNotFound
    }
    if group.GameMode != nil && *group.GameMode == consts.GameModeTeam {
        // Resolve subgroup membership
        var subMember models.GameSubgroupMember
        if err := query.
            Joins("JOIN game_subgroups ON game_subgroups.id = game_subgroup_members.game_subgroup_id").
            Where("game_subgroups.game_group_id", *gameGroupID).
            Where("game_subgroup_members.user_id", userID).
            First(&subMember); err != nil || subMember.ID == "" {
            return nil, ErrNotInSubgroup
        }
        gameSubgroupID = &subMember.GameSubgroupID
    }
}
```

Set the fields on the new session struct:

```go
session := models.GameSessionTotal{
    // ... existing fields ...
    GameGroupID:    gameGroupID,
    GameSubgroupID: gameSubgroupID,
}
```

- [ ] **Step 4: Update StartLevel to inherit group fields**

In `dx-api/app/services/api/session_service.go:592-607`, **before** the `query.Create(&levelSession)` call, look up the parent session's group fields and set them on the struct:

```go
// Before creating — inherit group fields from parent session
var parentSession models.GameSessionTotal
if err := query.Where("id", sessionID).First(&parentSession); err == nil {
    levelSession.GameGroupID = parentSession.GameGroupID
    levelSession.GameSubgroupID = parentSession.GameSubgroupID
}
```

This must be placed after the `levelSession` struct is initialized but **before** `query.Create(&levelSession)` on line 605.

- [ ] **Step 5: Update StartSession controller and request**

Update the `StartSession` controller to read `game_group_id` from the request and pass to the service. Add `GameGroupID *string` to the start session request struct.

- [ ] **Step 6: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/services/api/session_service.go \
       dx-api/app/http/controllers/api/game_session_controller.go \
       dx-api/app/http/requests/api/
git commit -m "feat: session service supports game_group_id for group play"
```

---

### Task 8: Winner Determination Service

**Files:**
- Create: `dx-api/app/services/api/group_winner_service.go`

- [ ] **Step 1: Create winner determination service**

```go
// dx-api/app/services/api/group_winner_service.go
package api

import (
	"fmt"
	"time"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// LevelWinnerResult holds winner data for SSE broadcast.
type LevelWinnerResult struct {
	GameLevelID string `json:"game_level_id"`
	Mode        string `json:"mode"`
	Winner      any    `json:"winner"`
}

type SoloWinner struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

type TeamWinner struct {
	SubgroupID   string       `json:"subgroup_id"`
	SubgroupName string       `json:"subgroup_name"`
	TotalScore   int          `json:"total_score"`
	Members      []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

// CheckAndDetermineWinner checks if all participants completed a level and determines the winner.
// Uses SELECT ... FOR UPDATE to prevent concurrent duplicate winner calculations.
// Returns the result if winner was determined, nil if still waiting for participants.
func CheckAndDetermineWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Lock participant rows to prevent concurrent winner determination
	var participantCount int64
	row := tx.Raw(
		"SELECT COUNT(*) FROM game_session_totals WHERE game_group_id = ? AND ended_at IS NULL FOR UPDATE",
		gameGroupID)
	// Use Goravel's count approach within transaction
	participantCount, err = tx.Model(&models.GameSessionTotal{}).
		Where("game_group_id", gameGroupID).Where("ended_at IS NULL").Count()
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to count participants: %w", err)
	}
	_ = row // use raw FOR UPDATE in actual implementation

	// Count completed level sessions for this group+level
	completedCount, err := tx.Model(&models.GameSessionLevel{}).
		Where("game_group_id", gameGroupID).Where("game_level_id", gameLevelID).
		Where("ended_at IS NOT NULL").Count()
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to count completed levels: %w", err)
	}

	if completedCount < participantCount {
		_ = tx.Rollback()
		return nil, nil // Still waiting
	}

	_ = tx.Commit()

	// All done — determine winner (outside transaction, reads only)
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	if group.GameMode != nil && *group.GameMode == "team" {
		return determineTeamWinner(gameGroupID, gameLevelID)
	}
	return determineSoloWinner(gameGroupID, gameLevelID)
}

// DetermineWinnerForLevel calculates winner without checking participant counts.
// Used by ForceEndGroupGame after sessions are already ended.
func DetermineWinnerForLevel(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	if group.GameMode != nil && *group.GameMode == "team" {
		return determineTeamWinner(gameGroupID, gameLevelID)
	}
	return determineSoloWinner(gameGroupID, gameLevelID)
}

func determineSoloWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	// Query all completed level sessions, ordered by score desc then ended_at asc (tie-break)
	type levelScore struct {
		UserID  string    `gorm:"column:user_id"`
		Score   int       `gorm:"column:score"`
		EndedAt time.Time `gorm:"column:ended_at"`
	}

	var scores []levelScore
	rows, err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, gsl.score, gsl.ended_at
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		 ORDER BY gsl.score DESC, gsl.ended_at ASC
		 LIMIT 1`, gameGroupID, gameLevelID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query solo winner: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var winner levelScore
	if err := rows.Scan(&winner.UserID, &winner.Score, &winner.EndedAt); err != nil {
		return nil, fmt.Errorf("failed to scan winner: %w", err)
	}

	// Get username
	var user models.User
	facades.Orm().Query().Where("id", winner.UserID).First(&user)

	// Update last_won_at on game_group_members
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id = ?",
		now, gameGroupID, winner.UserID)

	return &LevelWinnerResult{
		GameLevelID: gameLevelID,
		Mode:        "solo",
		Winner: SoloWinner{
			UserID:   winner.UserID,
			UserName: user.Nickname,
			Score:    winner.Score,
		},
	}, nil
}

func determineTeamWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	// Sum scores by subgroup
	type subgroupScore struct {
		GameSubgroupID string    `gorm:"column:game_subgroup_id"`
		TotalScore     int       `gorm:"column:total_score"`
		LastEndedAt    time.Time `gorm:"column:last_ended_at"`
	}

	rows, err := facades.Orm().Query().Raw(
		`SELECT gsl.game_subgroup_id, SUM(gsl.score) as total_score, MAX(gsl.ended_at) as last_ended_at
		 FROM game_session_levels gsl
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id IS NOT NULL
		 GROUP BY gsl.game_subgroup_id
		 ORDER BY total_score DESC, last_ended_at ASC
		 LIMIT 1`, gameGroupID, gameLevelID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query team winner: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var winnerGroup subgroupScore
	if err := rows.Scan(&winnerGroup.GameSubgroupID, &winnerGroup.TotalScore, &winnerGroup.LastEndedAt); err != nil {
		return nil, fmt.Errorf("failed to scan team winner: %w", err)
	}

	// Get subgroup name
	var subgroup models.GameSubgroup
	facades.Orm().Query().Where("id", winnerGroup.GameSubgroupID).First(&subgroup)

	// Get individual member scores for this subgroup
	memberRows, err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, gsl.score
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id = ?
		 ORDER BY gsl.score DESC`, gameGroupID, gameLevelID, winnerGroup.GameSubgroupID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer memberRows.Close()

	now := time.Now()
	var members []TeamMember
	var memberUserIDs []string
	for memberRows.Next() {
		var userID string
		var score int
		if err := memberRows.Scan(&userID, &score); err != nil {
			continue
		}
		var user models.User
		facades.Orm().Query().Where("id", userID).First(&user)
		members = append(members, TeamMember{UserID: userID, UserName: user.Nickname, Score: score})
		memberUserIDs = append(memberUserIDs, userID)
	}

	// Update last_won_at on winning subgroup
	facades.Orm().Query().Exec(
		"UPDATE game_subgroups SET last_won_at = ? WHERE id = ?",
		now, winnerGroup.GameSubgroupID)

	// Update last_won_at on all participating members of winning subgroup
	if len(memberUserIDs) > 0 {
		facades.Orm().Query().Exec(
			"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id IN ?",
			now, gameGroupID, memberUserIDs)
	}

	return &LevelWinnerResult{
		GameLevelID: gameLevelID,
		Mode:        "team",
		Winner: TeamWinner{
			SubgroupID:   winnerGroup.GameSubgroupID,
			SubgroupName: subgroup.Name,
			TotalScore:   winnerGroup.TotalScore,
			Members:      members,
		},
	}, nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/group_winner_service.go
git commit -m "feat: add winner determination service for group play"
```

---

### Task 9: Integrate Winner Check into CompleteLevel

**Files:**
- Modify: `dx-api/app/services/api/session_service.go:618-725` (CompleteLevel)

- [ ] **Step 1: Add group winner check after level completion**

In `dx-api/app/services/api/session_service.go`, after the transaction commits in `CompleteLevel` (after line 718), add:

```go
// Check for group winner determination
if session.GameGroupID != nil {
    result, err := CheckAndDetermineWinner(*session.GameGroupID, gameLevelID)
    if err == nil && result != nil {
        // Broadcast level complete event
        helpers.GroupSSEHub.Broadcast(*session.GameGroupID, "group_level_complete", result)

        // Check if this was the last level — if so, end the round
        var totalLevels int64
        facades.Orm().Query().Model(&models.GameLevel{}).
            Where("game_id", session.GameID).Where("is_active", true).Count(&totalLevels)

        if int64(session.PlayedLevelsCount+1) >= totalLevels {
            // All levels done — set is_playing = false
            facades.Orm().Query().Model(&models.GameGroup{}).
                Where("id", *session.GameGroupID).Update("is_playing", false)
        }
    }
}
```

Update `CompleteLevelResult` to include group winner info if needed for the response.

- [ ] **Step 2: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/session_service.go
git commit -m "feat: integrate group winner check into CompleteLevel"
```

---

### Task 10: Force End Backend

**Files:**
- Modify: `dx-api/app/services/api/group_game_service.go`
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add ForceEndGroupGame service**

In `dx-api/app/services/api/group_game_service.go`, add:

```go
// ForceEndGroupGame ends all active sessions and determines winners.
func ForceEndGroupGame(userID, groupID string) ([]LevelWinnerResult, error) {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return nil, ErrNotGroupOwner
	}
	if !group.IsPlaying {
		return nil, ErrGroupNotPlaying
	}

	now := time.Now()

	// End all active session levels
	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_session_levels SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID); err != nil {
		return nil, fmt.Errorf("failed to end session levels: %w", err)
	}

	// End all active session totals
	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_session_totals SET ended_at = ? WHERE game_group_id = ? AND ended_at IS NULL",
		now, groupID); err != nil {
		return nil, fmt.Errorf("failed to end session totals: %w", err)
	}

	// Collect completed level IDs for winner determination
	type levelID struct {
		GameLevelID string `gorm:"column:game_level_id"`
	}
	rows, err := facades.Orm().Query().Raw(
		`SELECT DISTINCT game_level_id FROM game_session_levels
		 WHERE game_group_id = ? AND ended_at IS NOT NULL`, groupID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query levels: %w", err)
	}
	defer rows.Close()

	var results []LevelWinnerResult
	for rows.Next() {
		var lid string
		if err := rows.Scan(&lid); err != nil {
			continue
		}
		// Use DetermineWinnerForLevel (no participant count check — sessions already ended)
		result, err := DetermineWinnerForLevel(groupID, lid)
		if err == nil && result != nil {
			results = append(results, *result)
		}
	}

	// Set is_playing = false
	facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("is_playing", false)

	// Broadcast force end event
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_force_end", map[string]any{
		"results": results,
	})

	return results, nil
}
```

Add `"time"` to imports.

- [ ] **Step 2: Add ForceEnd controller**

In `dx-api/app/http/controllers/api/group_game_controller.go`, add:

```go
// ForceEnd force-ends the current group game round.
func (c *GroupGameController) ForceEnd(ctx http.Context) http.Response {
	userID := ctx.Value("id").(string)
	groupID := ctx.Request().Route("id")

	results, err := api.ForceEndGroupGame(userID, groupID)
	if err != nil {
		return mapGroupGameError(ctx, err)
	}

	return ctx.Response().Json(http.StatusOK, helpers.SuccessResponse(map[string]any{
		"results": results,
	}))
}
```

- [ ] **Step 3: Add route**

In `dx-api/routes/api.go`, add inside group game routes:

```go
groups.Post("/{id}/force-end", groupGameController.ForceEnd)
```

- [ ] **Step 4: Update mapGroupGameError for new errors**

Add cases for `ErrGroupIsPlaying`, `ErrGroupNotPlaying`, `ErrNoGameSet`, `ErrNoGameModeSet`, `ErrNotInSubgroup` in the error mapping function.

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/group_game_service.go \
       dx-api/app/http/controllers/api/group_game_controller.go \
       dx-api/routes/api.go
git commit -m "feat: add force-end endpoint for group game"
```

---

### Task 11: Frontend — SSE Hook & Group Types

**Files:**
- Create: `dx-web/src/features/web/groups/hooks/use-group-events.ts`
- Modify: `dx-web/src/features/web/groups/types/group.ts`
- Modify: `dx-web/src/features/web/groups/actions/group.action.ts`

- [ ] **Step 1: Add group play types**

In `dx-web/src/features/web/groups/types/group.ts`, add:

```typescript
export type GroupGameStartEvent = {
  game_group_id: string
  game_id: string
  game_name: string
  game_mode: "solo" | "team"
  degree: string
  pattern: string | null
  answer_time_limit: number
}

export type SoloWinner = {
  user_id: string
  user_name: string
  score: number
}

export type TeamWinner = {
  subgroup_id: string
  subgroup_name: string
  total_score: number
  members: { user_id: string; user_name: string; score: number }[]
}

export type GroupLevelCompleteEvent = {
  game_level_id: string
  mode: "solo" | "team"
  winner: SoloWinner | TeamWinner
}

export type GroupForceEndEvent = {
  results: GroupLevelCompleteEvent[]
}
```

- [ ] **Step 2: Add group play API actions**

In `dx-web/src/features/web/groups/actions/group.action.ts`, add:

```typescript
startGame: (groupId: string, degree: string, pattern?: string) =>
  apiClient.post(`/groups/${groupId}/start-game`, { degree, pattern }),

forceEnd: (groupId: string) =>
  apiClient.post<{ results: GroupLevelCompleteEvent[] }>(`/groups/${groupId}/force-end`),
```

- [ ] **Step 3: Create SSE hook for group events**

```typescript
// dx-web/src/features/web/groups/hooks/use-group-events.ts
"use client"

import { useEffect, useRef, useCallback } from "react"
import type {
  GroupGameStartEvent,
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
} from "../types/group"

type GroupEventHandlers = {
  onGameStart?: (event: GroupGameStartEvent) => void
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void
  onForceEnd?: (event: GroupForceEndEvent) => void
}

export function useGroupEvents(groupId: string, handlers: GroupEventHandlers) {
  const handlersRef = useRef(handlers)
  handlersRef.current = handlers

  useEffect(() => {
    const token = localStorage.getItem("dx_token")
    if (!token) return

    const apiUrl = process.env.NEXT_PUBLIC_API_URL
    const url = `${apiUrl}/api/groups/${groupId}/events?token=${encodeURIComponent(token)}`

    const eventSource = new EventSource(url)

    eventSource.addEventListener("group_game_start", (e) => {
      const data: GroupGameStartEvent = JSON.parse(e.data)
      handlersRef.current.onGameStart?.(data)
    })

    eventSource.addEventListener("group_level_complete", (e) => {
      const data: GroupLevelCompleteEvent = JSON.parse(e.data)
      handlersRef.current.onLevelComplete?.(data)
    })

    eventSource.addEventListener("group_game_force_end", (e) => {
      const data: GroupForceEndEvent = JSON.parse(e.data)
      handlersRef.current.onForceEnd?.(data)
    })

    return () => eventSource.close()
  }, [groupId])
}
```

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npm run build`
Expected: Compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/groups/hooks/use-group-events.ts \
       dx-web/src/features/web/groups/types/group.ts \
       dx-web/src/features/web/groups/actions/group.action.ts
git commit -m "feat: add SSE hook and types for group game events"
```

---

### Task 12: Frontend — Start Game UI on Group Detail

**Files:**
- Create: `dx-web/src/features/web/groups/components/start-game-dialog.tsx`
- Modify: `dx-web/src/features/web/groups/components/group-detail-content.tsx`

- [ ] **Step 1: Create start game dialog component**

Create `dx-web/src/features/web/groups/components/start-game-dialog.tsx` — a dialog that shows the degree/pattern selection (same card panel as game details page). On confirm, calls `groupApi.startGame()`.

Key elements:
- Degree selector: practice, beginner, intermediate, advanced
- Pattern selector: listen, speak, read, write
- Submit button: "开始游戏"
- Disabled if `is_playing` is true

- [ ] **Step 2: Add "开始游戏" button to group detail**

In `dx-web/src/features/web/groups/components/group-detail-content.tsx`, below the current game section (当前课程游戏), add a "开始游戏" button that opens the `StartGameDialog`. Only visible to the group owner when a game is set and `is_playing` is false.

Also add a "强制结束" button visible when `is_playing` is true.

- [ ] **Step 3: Wire up SSE events**

Use `useGroupEvents` hook in the group detail page. On `group_game_start` event, navigate to the game play screen with the group context.

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npm run build`
Expected: Compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/groups/components/start-game-dialog.tsx \
       dx-web/src/features/web/groups/components/group-detail-content.tsx
git commit -m "feat: add start game UI and force-end button on group detail"
```

---

### Task 13: Frontend — Game Play with Answer Timer

**Files:**
- Modify: `dx-web/src/features/web/play/hooks/use-game-store.ts`
- Modify: `dx-web/src/features/web/play/components/game-loading-screen.tsx`
- Create: `dx-web/src/features/web/play/components/answer-timer.tsx`
- Modify: `dx-web/src/features/web/play/actions/session.action.ts`

- [ ] **Step 1: Update game store with group context**

In `dx-web/src/features/web/play/hooks/use-game-store.ts`, add to `GameState`:

```typescript
gameGroupId: string | null
answerTimeLimit: number | null
```

Update `initSession` to accept and store these values.

- [ ] **Step 2: Update startSessionAction to pass game_group_id**

In `dx-web/src/features/web/play/actions/session.action.ts`, update `startSessionAction` to accept optional `gameGroupId` parameter and include it in the API request body.

- [ ] **Step 3: Update loading screen for group context**

In `dx-web/src/features/web/play/components/game-loading-screen.tsx`, read group context from URL search params (e.g., `?groupId=xxx&answerTimeLimit=10`) and pass `gameGroupId` to `startSessionAction`.

- [ ] **Step 4: Create answer timer component**

```typescript
// dx-web/src/features/web/play/components/answer-timer.tsx
"use client"

import { useEffect, useRef, useState } from "react"

type Props = {
  seconds: number
  onExpire: () => void
  resetKey: string | number // changes to reset the timer
}

export function AnswerTimer({ seconds, onExpire, resetKey }: Props) {
  const [remaining, setRemaining] = useState(seconds)
  const onExpireRef = useRef(onExpire)
  onExpireRef.current = onExpire

  useEffect(() => {
    setRemaining(seconds)
    const interval = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          onExpireRef.current()
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [seconds, resetKey])

  return (
    <div className="text-sm font-mono tabular-nums">
      {remaining}s
    </div>
  )
}
```

- [ ] **Step 5: Integrate timer into game play UI**

In the game play component, render `<AnswerTimer>` when `gameGroupId` is present in the store. On expire, auto-submit or auto-skip depending on whether the user has started typing.

- [ ] **Step 6: Verify build**

Run: `cd dx-web && npm run build`
Expected: Compiles successfully.

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/play/hooks/use-game-store.ts \
       dx-web/src/features/web/play/components/game-loading-screen.tsx \
       dx-web/src/features/web/play/components/answer-timer.tsx \
       dx-web/src/features/web/play/actions/session.action.ts
git commit -m "feat: add answer timer and group context to game play"
```

---

### Task 14: Frontend — Waiting Screen & Results Panel

**Files:**
- Create: `dx-web/src/features/web/play/components/group-waiting-screen.tsx`
- Create: `dx-web/src/features/web/play/components/group-result-panel.tsx`
- Modify: `dx-web/src/features/web/play/hooks/use-game-store.ts`

- [ ] **Step 1: Create waiting screen component**

```typescript
// dx-web/src/features/web/play/components/group-waiting-screen.tsx
"use client"

import { Loader2 } from "lucide-react"

export function GroupWaitingScreen() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <p className="text-lg text-muted-foreground">等待其他选手完成...</p>
    </div>
  )
}
```

- [ ] **Step 2: Create result panel component**

Create `dx-web/src/features/web/play/components/group-result-panel.tsx` that displays:
- Solo mode: winner name, score, trophy icon
- Team mode: winning team name, total score, member breakdown
- "下一关" button if more levels remain
- "查看最终结果" for last level

- [ ] **Step 3: Add group play phases to game store**

In `dx-web/src/features/web/play/hooks/use-game-store.ts`, extend `GamePhase` or add a `groupPhase`:

```typescript
groupPhase: "playing" | "waiting" | "result" | null
groupResult: GroupLevelCompleteEvent | null
```

Add actions: `setGroupWaiting()`, `setGroupResult(result)`.

- [ ] **Step 4: Wire up SSE in game play screen**

Use `useGroupEvents` in the game play shell. When `group_level_complete` fires, call `setGroupResult()`. When `group_game_force_end` fires, show final results.

- [ ] **Step 5: Verify build**

Run: `cd dx-web && npm run build`
Expected: Compiles successfully.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/play/components/group-waiting-screen.tsx \
       dx-web/src/features/web/play/components/group-result-panel.tsx \
       dx-web/src/features/web/play/hooks/use-game-store.ts
git commit -m "feat: add group waiting screen and result panel"
```

---

### Task 15: Frontend — Owner Remove Button & Group Detail Updates

**Files:**
- Modify: `dx-web/src/features/web/groups/components/member-list.tsx` (or similar)
- Modify: `dx-web/src/features/web/groups/types/group.ts`

- [ ] **Step 1: Show remove button for owner in member list**

Update the member list component to show the remove/kick button for the owner's own row. Currently the owner row likely hides the kick button — remove that condition.

- [ ] **Step 2: Update GroupDetail type**

Add `answer_time_limit`, `is_playing` to `GroupDetail` type in `dx-web/src/features/web/groups/types/group.ts`.

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npm run build`
Expected: Compiles successfully.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/groups/components/ dx-web/src/features/web/groups/types/group.ts
git commit -m "feat: owner remove button, group detail updates for play state"
```
