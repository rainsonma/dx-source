# PK Specified Opponent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real player-vs-player PK with invitation flow, online presence tracking, and a PK room waiting page.

**Architecture:** Extend existing PK system with `pk_type`/`invitation_status` columns, add a global user SSE hub backed by Redis online set for presence, and build invitation + PK room frontend flows. Random/robot PK flow is completely unchanged.

**Tech Stack:** Go/Goravel, Next.js 16, PostgreSQL, Redis SET, SSE (EventSource), ShadCN UI (Tabs, Input, Button), Zustand

---

## File Structure

### New Files — Backend

| File | Responsibility |
|------|----------------|
| `dx-api/database/migrations/20260406000001_add_pk_type_columns.go` | Add `pk_type` + `invitation_status` columns to `game_pks` |
| `dx-api/app/helpers/redis_set.go` | Redis SET operations (SADD, SREM, SISMEMBER, SCARD) |
| `dx-api/app/helpers/sse_user_hub.go` | Global per-user SSE hub + Redis online set integration |
| `dx-api/app/http/controllers/api/user_sse_controller.go` | `GET /api/user/events` SSE endpoint |
| `dx-api/app/services/api/online_service.go` | Verify user exists + online + VIP |
| `dx-api/app/http/controllers/api/user_verify_controller.go` | `POST /api/users/verify-online` endpoint |
| `dx-api/app/http/requests/api/verify_online_request.go` | Request validation for verify-online |
| `dx-api/app/services/api/pk_invite_service.go` | Invitation business logic (invite, accept, decline, details) |
| `dx-api/app/http/controllers/api/pk_invite_controller.go` | Invitation + details endpoints |
| `dx-api/app/http/requests/api/pk_invite_request.go` | Request validation for invite |

### New Files — Frontend

| File | Responsibility |
|------|----------------|
| `dx-web/src/hooks/use-user-sse.ts` | Global SSE hook for `/api/user/events` |
| `dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx` | Slide-up invitation notification |
| `dx-web/src/features/web/play-pk/components/pk-invitation-provider.tsx` | Provider wiring SSE + popup state |
| `dx-web/src/features/web/play-pk/actions/invite.action.ts` | Invite/accept/decline/verify/details actions |
| `dx-web/src/app/(web)/hall/pk-room/[id]/page.tsx` | PK room waiting page |

### Modified Files

| File | Changes |
|------|---------|
| `dx-api/app/models/game_pk.go` | Add `PkType`, `InvitationStatus` fields |
| `dx-api/app/consts/error_code.go` | Add new error codes |
| `dx-api/app/services/api/errors.go` | Add new error sentinels |
| `dx-api/app/services/api/game_play_pk_service.go` | `NextPkLevel`: handle specified type |
| `dx-api/routes/api.go` | Register new routes |
| `dx-api/bootstrap/migrations.go` | Register new migration |
| `dx-web/src/features/web/play-core/components/game-mode-card.tsx` | Tabs, restyle difficulty, verify input |
| `dx-web/src/app/(web)/layout.tsx` | Add PkInvitationProvider |
| `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx` | Handle `pkId`+`sessionId` params |
| `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx` | Skip `startPkAction` when `pkId` provided |

---

## Task 1: Database Migration — Add pk_type + invitation_status

**Files:**
- Create: `dx-api/database/migrations/20260406000001_add_pk_type_columns.go`
- Modify: `dx-api/bootstrap/migrations.go`
- Modify: `dx-api/app/models/game_pk.go`

- [ ] **Step 1: Create migration file**

```go
// dx-api/database/migrations/20260406000001_add_pk_type_columns.go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260406000001AddPkTypeColumns struct{}

func (r *M20260406000001AddPkTypeColumns) Signature() string {
	return "20260406000001_add_pk_type_columns"
}

func (r *M20260406000001AddPkTypeColumns) Up() error {
	_, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_pks
		 ADD COLUMN pk_type text NOT NULL DEFAULT 'random',
		 ADD COLUMN invitation_status text`)
	return err
}

func (r *M20260406000001AddPkTypeColumns) Down() error {
	_, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_pks
		 DROP COLUMN IF EXISTS pk_type,
		 DROP COLUMN IF EXISTS invitation_status`)
	return err
}
```

- [ ] **Step 2: Register migration in bootstrap**

In `dx-api/bootstrap/migrations.go`, add to the end of the `Migrations()` slice:

```go
&migrations.M20260406000001AddPkTypeColumns{},
```

- [ ] **Step 3: Update GamePk model**

In `dx-api/app/models/game_pk.go`, add two fields to the `GamePk` struct:

```go
PkType           string  `gorm:"column:pk_type" json:"pk_type"`
InvitationStatus *string `gorm:"column:invitation_status" json:"invitation_status"`
```

- [ ] **Step 4: Run migration**

```bash
cd dx-api && go run . artisan migrate
```

Expected: Migration runs successfully, `game_pks` table has new columns.

- [ ] **Step 5: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add dx-api/database/migrations/20260406000001_add_pk_type_columns.go dx-api/bootstrap/migrations.go dx-api/app/models/game_pk.go
git commit -m "feat: add pk_type and invitation_status columns to game_pks"
```

---

## Task 2: Redis SET Helpers

**Files:**
- Create: `dx-api/app/helpers/redis_set.go`

- [ ] **Step 1: Create Redis SET helper functions**

```go
// dx-api/app/helpers/redis_set.go
package helpers

import "context"

// RedisSetAdd adds members to a Redis SET.
func RedisSetAdd(key string, members ...string) error {
	ctx := context.Background()
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return GetRedis().SAdd(ctx, key, args...).Err()
}

// RedisSetRemove removes members from a Redis SET.
func RedisSetRemove(key string, members ...string) error {
	ctx := context.Background()
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return GetRedis().SRem(ctx, key, args...).Err()
}

// RedisSetIsMember checks if a member exists in a Redis SET.
func RedisSetIsMember(key string, member string) (bool, error) {
	ctx := context.Background()
	return GetRedis().SIsMember(ctx, key, member).Result()
}

// RedisSetCard returns the number of members in a Redis SET.
func RedisSetCard(key string) (int64, error) {
	ctx := context.Background()
	return GetRedis().SCard(ctx, key).Result()
}
```

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/helpers/redis_set.go
git commit -m "feat: add Redis SET helper functions"
```

---

## Task 3: Global User SSE Hub

**Files:**
- Create: `dx-api/app/helpers/sse_user_hub.go`
- Create: `dx-api/app/http/controllers/api/user_sse_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Create UserSSEHub**

```go
// dx-api/app/helpers/sse_user_hub.go
package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const redisOnlineUsersKey = "online_users"

// UserSSEHub manages per-user SSE connections for global notifications.
type UserSSEHub struct {
	mu    sync.RWMutex
	conns map[string]*SSEConnection // userID -> conn
}

// UserHub is the global SSE hub instance for user-level events.
var UserHub = &UserSSEHub{
	conns: make(map[string]*SSEConnection),
}

// Register adds a connection for a user and marks them online in Redis.
func (h *UserSSEHub) Register(userID string, w http.ResponseWriter) *SSEConnection {
	conn := NewSSEConnection(w)

	h.mu.Lock()
	if old, ok := h.conns[userID]; ok {
		close(old.done)
	}
	h.conns[userID] = conn
	h.mu.Unlock()

	_ = RedisSetAdd(redisOnlineUsersKey, userID)

	return conn
}

// Unregister removes a connection and marks the user offline in Redis.
func (h *UserSSEHub) Unregister(userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[userID]
	if exists && current == conn {
		delete(h.conns, userID)
	}
	h.mu.Unlock()

	if exists && current == conn {
		_ = RedisSetRemove(redisOnlineUsersKey, userID)
	}
}

// SendToUser sends an SSE event to a specific user.
func (h *UserSSEHub) SendToUser(userID, event string, data any) {
	h.mu.RLock()
	conn, ok := h.conns[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	jsonBytes, _ := json.Marshal(data)
	fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event, jsonBytes)
	if conn.flusher != nil {
		conn.flusher.Flush()
	}
}

// IsOnline checks if a user is in the Redis online set.
func (h *UserSSEHub) IsOnline(userID string) bool {
	ok, _ := RedisSetIsMember(redisOnlineUsersKey, userID)
	return ok
}
```

- [ ] **Step 2: Create UserSSE controller**

```go
// dx-api/app/http/controllers/api/user_sse_controller.go
package api

import (
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
)

type UserSSEController struct{}

func NewUserSSEController() *UserSSEController {
	return &UserSSEController{}
}

// Events establishes a persistent SSE connection for user-level events.
func (c *UserSSEController) Events(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, 401, 0, "unauthorized")
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	conn := helpers.UserHub.Register(userID, w)
	defer helpers.UserHub.Unregister(userID, conn)

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

- [ ] **Step 3: Fix import — add `net/http` to controller**

The controller file needs the `net/http` import for `http.Flusher`:

```go
import (
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
)
```

- [ ] **Step 4: Register route**

In `dx-api/routes/api.go`, inside the `protected` group, add:

```go
// User SSE events
userSSEController := apicontrollers.NewUserSSEController()
protected.Get("/user/events", userSSEController.Events)
```

Place this BEFORE any prefix groups (same pattern as the PK events route).

- [ ] **Step 5: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/helpers/sse_user_hub.go dx-api/app/http/controllers/api/user_sse_controller.go dx-api/routes/api.go
git commit -m "feat: add global user SSE hub with Redis online presence"
```

---

## Task 4: Error Codes + Error Sentinels for Invitation

**Files:**
- Modify: `dx-api/app/consts/error_code.go`
- Modify: `dx-api/app/services/api/errors.go`

- [ ] **Step 1: Add error codes**

In `dx-api/app/consts/error_code.go`, add after `CodeNoMockUser = 40017`:

```go
CodeOpponentOffline    = 40018
CodeOpponentNotVip     = 40019
CodeCannotChallengeSelf = 40020
CodeInvitationNotPending = 40021
```

- [ ] **Step 2: Add error sentinels**

In `dx-api/app/services/api/errors.go`, add to the `var` block:

```go
ErrOpponentOffline      = errors.New("对方不在线")
ErrOpponentNotVip       = errors.New("对方会员已过期")
ErrCannotChallengeSelf  = errors.New("不能挑战自己")
ErrInvitationNotPending = errors.New("邀请状态已变更")
```

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/consts/error_code.go dx-api/app/services/api/errors.go
git commit -m "feat: add error codes and sentinels for PK invitation"
```

---

## Task 5: Verify Online Service + Endpoint

**Files:**
- Create: `dx-api/app/services/api/online_service.go`
- Create: `dx-api/app/http/requests/api/verify_online_request.go`
- Create: `dx-api/app/http/controllers/api/user_verify_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Create online service**

```go
// dx-api/app/services/api/online_service.go
package api

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type VerifyOnlineResult struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	IsOnline bool   `json:"is_online"`
	IsVip    bool   `json:"is_vip"`
}

// VerifyUserOnline checks if a user exists, is online, and has active VIP.
func VerifyUserOnline(callerID, username string) (*VerifyOnlineResult, error) {
	var user models.User
	if err := facades.Orm().Query().
		Select("id", "username", "nickname", "grade", "vip_due_at").
		Where("username", username).
		First(&user); err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return nil, ErrUserNotFound
	}

	if user.ID == callerID {
		return nil, ErrCannotChallengeSelf
	}

	if !helpers.UserHub.IsOnline(user.ID) {
		return &VerifyOnlineResult{
			UserID:   user.ID,
			Nickname: nickname(user),
			IsOnline: false,
			IsVip:    false,
		}, nil
	}

	vipActive := checkVipActive(user)

	return &VerifyOnlineResult{
		UserID:   user.ID,
		Nickname: nickname(user),
		IsOnline: true,
		IsVip:    vipActive,
	}, nil
}

func nickname(user models.User) string {
	if user.Nickname != nil && *user.Nickname != "" {
		return *user.Nickname
	}
	return user.Username
}
```

- [ ] **Step 2: Create request validation**

```go
// dx-api/app/http/requests/api/verify_online_request.go
package api

import "github.com/goravel/framework/contracts/http"

type VerifyOnlineRequest struct {
	Username string `form:"username" json:"username"`
}

func (r *VerifyOnlineRequest) Authorize(ctx http.Context) error { return nil }

func (r *VerifyOnlineRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:2|max_len:30",
	}
}

func (r *VerifyOnlineRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "trim",
	}
}

func (r *VerifyOnlineRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": "请输入用户名",
		"username.min_len":  "用户名至少2个字符",
		"username.max_len":  "用户名最长30个字符",
	}
}
```

- [ ] **Step 3: Create verify controller**

```go
// dx-api/app/http/controllers/api/user_verify_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type UserVerifyController struct{}

func NewUserVerifyController() *UserVerifyController {
	return &UserVerifyController{}
}

// VerifyOnline checks if a user exists, is online, and has active VIP.
func (c *UserVerifyController) VerifyOnline(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.VerifyOnlineRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.VerifyUserOnline(userID, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		case errors.Is(err, services.ErrCannotChallengeSelf):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeCannotChallengeSelf, "不能挑战自己")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
		}
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 4: Register route**

In `dx-api/routes/api.go`, inside the `protected` group, add:

```go
// User verify
userVerifyController := apicontrollers.NewUserVerifyController()
protected.Post("/users/verify-online", userVerifyController.VerifyOnline)
```

- [ ] **Step 5: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/online_service.go dx-api/app/http/requests/api/verify_online_request.go dx-api/app/http/controllers/api/user_verify_controller.go dx-api/routes/api.go
git commit -m "feat: add verify-online endpoint for PK opponent validation"
```

---

## Task 6: PK Invitation Service

**Files:**
- Create: `dx-api/app/services/api/pk_invite_service.go`

- [ ] **Step 1: Create invitation service**

```go
// dx-api/app/services/api/pk_invite_service.go
package api

import (
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/oklog/ulid/v2"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type PkInviteResult struct {
	PkID        string `json:"pk_id"`
	SessionID   string `json:"session_id"`
	GameLevelID string `json:"game_level_id"`
}

type PkAcceptResult struct {
	SessionID   string  `json:"session_id"`
	GameID      string  `json:"game_id"`
	GameLevelID string  `json:"game_level_id"`
	Degree      string  `json:"degree"`
	Pattern     *string `json:"pattern"`
}

type PkDetailsResult struct {
	PkID             string  `json:"pk_id"`
	GameID           string  `json:"game_id"`
	GameName         string  `json:"game_name"`
	GameMode         string  `json:"game_mode"`
	LevelID          string  `json:"level_id"`
	LevelName        string  `json:"level_name"`
	Degree           string  `json:"degree"`
	Pattern          *string `json:"pattern"`
	InitiatorID      string  `json:"initiator_id"`
	InitiatorName    string  `json:"initiator_name"`
	OpponentID       string  `json:"opponent_id"`
	OpponentName     string  `json:"opponent_name"`
	InvitationStatus string  `json:"invitation_status"`
}

// InvitePk creates a specified PK and sends an invitation to the opponent.
func InvitePk(userID, gameID, gameLevelID, degree string, pattern *string, opponentID string) (*PkInviteResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	// Verify opponent is still online and VIP
	if !helpers.UserHub.IsOnline(opponentID) {
		return nil, ErrOpponentOffline
	}
	opponentVip, err := IsVipActive(opponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check opponent VIP: %w", err)
	}
	if !opponentVip {
		return nil, ErrOpponentNotVip
	}

	// Verify game exists and is published
	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
	if game.ID == "" {
		return nil, ErrGameNotFound
	}
	if game.Status != "published" {
		return nil, ErrGameNotPublished
	}

	// Verify level exists
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	// Get opponent name for SSE event
	var opponent models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", opponentID).First(&opponent); err != nil {
		return nil, fmt.Errorf("failed to find opponent: %w", err)
	}

	// Get initiator name for SSE event
	var initiator models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&initiator); err != nil {
		return nil, fmt.Errorf("failed to find initiator: %w", err)
	}

	pkID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
	statusPending := "pending"

	pk := models.GamePk{
		ID:               pkID,
		UserID:           userID,
		OpponentID:       opponentID,
		GameID:           gameID,
		GameLevelID:      gameLevelID,
		Degree:           degree,
		Pattern:          pattern,
		RobotDifficulty:  "",
		IsPlaying:        true,
		PkType:           "specified",
		InvitationStatus: &statusPending,
	}
	if err := facades.Orm().Query().Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create PK: %w", err)
	}

	// Create initiator's session
	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
	now := time.Now()
	totalItems, _ := countLevelItems(gameLevelID, degree)
	session := models.GameSession{
		ID:              sessionID,
		UserID:          userID,
		GameID:          gameID,
		GameLevelID:     gameLevelID,
		Degree:          degree,
		Pattern:         pattern,
		GamePkID:        &pkID,
		StartedAt:       &now,
		TotalItemsCount: totalItems,
	}
	if err := facades.Orm().Query().Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Push invitation SSE event to opponent
	helpers.UserHub.SendToUser(opponentID, "pk_invitation", map[string]string{
		"pk_id":          pkID,
		"game_id":        gameID,
		"game_name":      game.Name,
		"game_mode":      game.Mode,
		"level_name":     level.Name,
		"initiator_id":   userID,
		"initiator_name": nickname(initiator),
	})

	return &PkInviteResult{
		PkID:        pkID,
		SessionID:   sessionID,
		GameLevelID: gameLevelID,
	}, nil
}

// AcceptPkInvite accepts an invitation and creates the opponent's session.
func AcceptPkInvite(userID, pkID string) (*PkAcceptResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return nil, fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return nil, ErrPkNotFound
	}
	if pk.OpponentID != userID {
		return nil, ErrForbidden
	}
	if pk.InvitationStatus == nil || *pk.InvitationStatus != "pending" {
		return nil, ErrInvitationNotPending
	}

	// Update invitation status
	statusAccepted := "accepted"
	if _, err := facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).
		Update("invitation_status", statusAccepted); err != nil {
		return nil, fmt.Errorf("failed to update invitation status: %w", err)
	}

	// Create opponent's session
	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
	now := time.Now()
	totalItems, _ := countLevelItems(pk.GameLevelID, pk.Degree)
	session := models.GameSession{
		ID:              sessionID,
		UserID:          userID,
		GameID:          pk.GameID,
		GameLevelID:     pk.GameLevelID,
		Degree:          pk.Degree,
		Pattern:         pk.Pattern,
		GamePkID:        &pkID,
		StartedAt:       &now,
		TotalItemsCount: totalItems,
	}
	if err := facades.Orm().Query().Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create opponent session: %w", err)
	}

	// Get opponent name for SSE event
	var opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&opponent)

	// Broadcast accepted event to PK room via PkHub
	helpers.PkHub.Broadcast(pkID, "pk_invitation_accepted", map[string]string{
		"pk_id":         pkID,
		"opponent_id":   userID,
		"opponent_name": nickname(opponent),
	})

	return &PkAcceptResult{
		SessionID:   sessionID,
		GameID:      pk.GameID,
		GameLevelID: pk.GameLevelID,
		Degree:      pk.Degree,
		Pattern:     pk.Pattern,
	}, nil
}

// DeclinePkInvite declines an invitation and ends the initiator's session.
func DeclinePkInvite(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.OpponentID != userID {
		return ErrForbidden
	}
	if pk.InvitationStatus == nil || *pk.InvitationStatus != "pending" {
		return ErrInvitationNotPending
	}

	statusDeclined := "declined"
	now := time.Now()

	// Update PK status
	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).
		Updates(map[string]interface{}{
			"invitation_status": statusDeclined,
			"is_playing":        false,
		})

	// End initiator's session
	facades.Orm().Query().Exec(
		"UPDATE game_sessions SET ended_at = ? WHERE game_pk_id = ? AND user_id = ? AND ended_at IS NULL",
		now, pkID, pk.UserID)

	// Broadcast declined event
	helpers.PkHub.Broadcast(pkID, "pk_invitation_declined", map[string]string{
		"pk_id": pkID,
	})

	return nil
}

// GetPkDetails returns PK information for the room page.
func GetPkDetails(userID, pkID string) (*PkDetailsResult, error) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return nil, fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return nil, ErrPkNotFound
	}
	if pk.UserID != userID && pk.OpponentID != userID {
		return nil, ErrForbidden
	}

	var game models.Game
	facades.Orm().Query().Select("id", "name", "mode").Where("id", pk.GameID).First(&game)

	var level models.GameLevel
	facades.Orm().Query().Select("id", "name").Where("id", pk.GameLevelID).First(&level)

	var initiator, opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", pk.UserID).First(&initiator)
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", pk.OpponentID).First(&opponent)

	status := ""
	if pk.InvitationStatus != nil {
		status = *pk.InvitationStatus
	}

	return &PkDetailsResult{
		PkID:             pk.ID,
		GameID:           pk.GameID,
		GameName:         game.Name,
		GameMode:         game.Mode,
		LevelID:          pk.GameLevelID,
		LevelName:        level.Name,
		Degree:           pk.Degree,
		Pattern:          pk.Pattern,
		InitiatorID:      pk.UserID,
		InitiatorName:    nickname(initiator),
		OpponentID:       pk.OpponentID,
		OpponentName:     nickname(opponent),
		InvitationStatus: status,
	}, nil
}
```

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/pk_invite_service.go
git commit -m "feat: add PK invitation service (invite, accept, decline, details)"
```

---

## Task 7: PK Invitation Controller + Routes

**Files:**
- Create: `dx-api/app/http/requests/api/pk_invite_request.go`
- Create: `dx-api/app/http/controllers/api/pk_invite_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Create invite request validation**

```go
// dx-api/app/http/requests/api/pk_invite_request.go
package api

import (
	"dx-api/app/helpers"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type PkInviteRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	OpponentID  string  `form:"opponent_id" json:"opponent_id"`
}

func (r *PkInviteRequest) Authorize(ctx http.Context) error { return nil }

func (r *PkInviteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
		"opponent_id":   "required|uuid",
	}
}

func (r *PkInviteRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":  "trim",
		"pattern": "trim",
	}
}

func (r *PkInviteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
		"opponent_id.required":   "请指定对手",
		"opponent_id.uuid":       "无效的对手ID",
	}
}

func (r *PkInviteRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", "intermediate")
	}
	return nil
}
```

- [ ] **Step 2: Create invite controller**

```go
// dx-api/app/http/controllers/api/pk_invite_controller.go
package api

import (
	"errors"
	"fmt"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type PkInviteController struct{}

func NewPkInviteController() *PkInviteController {
	return &PkInviteController{}
}

// Invite creates a specified PK and sends an invitation.
func (c *PkInviteController) Invite(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.PkInviteRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.InvitePk(userID, req.GameID, req.GameLevelID, req.Degree, req.Pattern, req.OpponentID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Accept accepts a PK invitation.
func (c *PkInviteController) Accept(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	result, err := services.AcceptPkInvite(userID, pkID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Decline declines a PK invitation.
func (c *PkInviteController) Decline(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.DeclinePkInvite(userID, pkID); err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Details returns PK info for the room page.
func (c *PkInviteController) Details(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	result, err := services.GetPkDetails(userID, pkID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

func mapInviteError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrOpponentOffline):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeOpponentOffline, "对方不在线")
	case errors.Is(err, services.ErrOpponentNotVip):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeOpponentNotVip, "对方会员已过期")
	case errors.Is(err, services.ErrCannotChallengeSelf):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeCannotChallengeSelf, "不能挑战自己")
	case errors.Is(err, services.ErrPkNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePkNotFound, "PK对战不存在")
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "forbidden")
	case errors.Is(err, services.ErrInvitationNotPending):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvitationNotPending, "邀请状态已变更")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏未发布")
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	default:
		fmt.Printf("[PK Invite] unhandled error: %v\n", err)
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
	}
}
```

- [ ] **Step 3: Register routes**

In `dx-api/routes/api.go`, inside the `protected` group, add after the existing play-pk routes:

```go
// PK invitation routes
pkInviteController := apicontrollers.NewPkInviteController()
protected.Get("/play-pk/{id}/details", pkInviteController.Details)
protected.Prefix("/play-pk/invite").Group(func(inv route.Router) {
	inv.Post("/", pkInviteController.Invite)
	inv.Post("/{id}/accept", pkInviteController.Accept)
	inv.Post("/{id}/decline", pkInviteController.Decline)
})
```

- [ ] **Step 4: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/requests/api/pk_invite_request.go dx-api/app/http/controllers/api/pk_invite_controller.go dx-api/routes/api.go
git commit -m "feat: add PK invitation controller and routes"
```

---

## Task 8: Modify NextPkLevel for Specified PK

**Files:**
- Modify: `dx-api/app/services/api/game_play_pk_service.go`

- [ ] **Step 1: Update NextPkLevel to handle specified type**

In `dx-api/app/services/api/game_play_pk_service.go`, find the `NextPkLevel` function. Currently it calls `StartPk(...)` which spawns a robot. Add a branch for specified PK:

After the line that fetches the PK and finds the next level, but before calling `StartPk`, add:

```go
// For specified PK, create a new PK directly without robot
if pk.PkType == "specified" {
	return nextSpecifiedPkLevel(userID, pk, *nextLevelID)
}
```

Then add the helper function in the same file:

```go
// nextSpecifiedPkLevel creates a new specified PK for the next level (no robot, no re-invitation).
func nextSpecifiedPkLevel(userID string, oldPk models.GamePk, nextLevelID string) (*PkStartResult, error) {
	pkID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
	statusAccepted := "accepted"

	pk := models.GamePk{
		ID:               pkID,
		UserID:           oldPk.UserID,
		OpponentID:       oldPk.OpponentID,
		GameID:           oldPk.GameID,
		GameLevelID:      nextLevelID,
		Degree:           oldPk.Degree,
		Pattern:          oldPk.Pattern,
		RobotDifficulty:  "",
		IsPlaying:        true,
		PkType:           "specified",
		InvitationStatus: &statusAccepted,
	}
	if err := facades.Orm().Query().Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create next PK: %w", err)
	}

	// Create sessions for both players
	now := time.Now()
	totalItems, _ := countLevelItems(nextLevelID, oldPk.Degree)

	for _, uid := range []string{oldPk.UserID, oldPk.OpponentID} {
		sid := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
		session := models.GameSession{
			ID:              sid,
			UserID:          uid,
			GameID:          oldPk.GameID,
			GameLevelID:     nextLevelID,
			Degree:          oldPk.Degree,
			Pattern:         oldPk.Pattern,
			GamePkID:        &pkID,
			StartedAt:       &now,
			TotalItemsCount: totalItems,
		}
		if err := facades.Orm().Query().Create(&session); err != nil {
			return nil, fmt.Errorf("failed to create session for %s: %w", uid, err)
		}
	}

	var opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", oldPk.OpponentID).First(&opponent)

	return &PkStartResult{
		PkID:         pkID,
		SessionID:    "", // Both players will get their session from restore
		GameLevelID:  nextLevelID,
		OpponentID:   oldPk.OpponentID,
		OpponentName: nickname(opponent),
	}, nil
}
```

Note: The `nickname` function is defined in `online_service.go` (same package), so it's accessible here.

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/game_play_pk_service.go
git commit -m "feat: handle specified PK type in NextPkLevel (no robot spawn)"
```

---

## Task 9: Frontend — Invite Actions

**Files:**
- Create: `dx-web/src/features/web/play-pk/actions/invite.action.ts`

- [ ] **Step 1: Create invite action file**

```typescript
// dx-web/src/features/web/play-pk/actions/invite.action.ts
import { apiClient } from "@/lib/api-client";

export async function verifyOpponentAction(username: string) {
  try {
    const res = await apiClient.post<{
      user_id: string;
      nickname: string;
      is_online: boolean;
      is_vip: boolean;
    }>("/api/users/verify-online", { username });
    if (res.code !== 0) return { data: null, error: res.message || "验证失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "验证失败" };
  }
}

export async function invitePkAction(data: {
  gameId: string;
  gameLevelId: string;
  degree: string;
  pattern: string | null;
  opponentId: string;
}) {
  try {
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      game_level_id: string;
    }>("/api/play-pk/invite", {
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
      degree: data.degree,
      pattern: data.pattern,
      opponent_id: data.opponentId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "邀请失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "邀请失败" };
  }
}

export async function acceptPkInviteAction(pkId: string) {
  try {
    const res = await apiClient.post<{
      session_id: string;
      game_id: string;
      game_level_id: string;
      degree: string;
      pattern: string | null;
    }>(`/api/play-pk/invite/${pkId}/accept`);
    if (res.code !== 0) return { data: null, error: res.message || "接受失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "接受失败" };
  }
}

export async function declinePkInviteAction(pkId: string) {
  try {
    const res = await apiClient.post<unknown>(`/api/play-pk/invite/${pkId}/decline`);
    if (res.code !== 0) return { data: null, error: res.message || "拒绝失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "拒绝失败" };
  }
}

export async function fetchPkDetailsAction(pkId: string) {
  try {
    const res = await apiClient.get<{
      pk_id: string;
      game_id: string;
      game_name: string;
      game_mode: string;
      level_id: string;
      level_name: string;
      degree: string;
      pattern: string | null;
      initiator_id: string;
      initiator_name: string;
      opponent_id: string;
      opponent_name: string;
      invitation_status: string;
    }>(`/api/play-pk/${pkId}/details`);
    if (res.code !== 0) return { data: null, error: res.message || "获取详情失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "获取详情失败" };
  }
}
```

- [ ] **Step 2: Verify lint**

```bash
cd dx-web && npx eslint src/features/web/play-pk/actions/invite.action.ts
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-pk/actions/invite.action.ts
git commit -m "feat: add PK invite frontend actions"
```

---

## Task 10: Frontend — Global User SSE Hook

**Files:**
- Create: `dx-web/src/hooks/use-user-sse.ts`

- [ ] **Step 1: Create user SSE hook**

```typescript
// dx-web/src/hooks/use-user-sse.ts
"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function useUserSSE(
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  listenersRef.current = listeners;

  useEffect(() => {
    const url = `${API_URL}/api/user/events`;
    const eventSource = new EventSource(url, { withCredentials: true });

    for (const eventName of Object.keys(listenersRef.current)) {
      eventSource.addEventListener(eventName, (e: MessageEvent) => {
        try {
          const data: unknown = JSON.parse(e.data);
          listenersRef.current[eventName]?.(data);
        } catch {
          // Discard malformed SSE messages
        }
      });
    }

    return () => {
      eventSource.close();
    };
  }, []);
}
```

- [ ] **Step 2: Verify lint**

```bash
cd dx-web && npx eslint src/hooks/use-user-sse.ts
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/hooks/use-user-sse.ts
git commit -m "feat: add global user SSE hook"
```

---

## Task 11: Frontend — PK Invitation Popup + Provider

**Files:**
- Create: `dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx`
- Create: `dx-web/src/features/web/play-pk/components/pk-invitation-provider.tsx`
- Modify: `dx-web/src/app/(web)/layout.tsx`

- [ ] **Step 1: Create invitation popup component**

```tsx
// dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx
"use client";

import { useEffect, useState } from "react";
import { Swords, X } from "lucide-react";

interface PkInvitationPopupProps {
  pkId: string;
  gameName: string;
  levelName: string;
  initiatorName: string;
  onAccept: () => void;
  onDecline: () => void;
}

export function PkInvitationPopup({
  pkId,
  gameName,
  levelName,
  initiatorName,
  onAccept,
  onDecline,
}: PkInvitationPopupProps) {
  const [timeLeft, setTimeLeft] = useState(30);

  useEffect(() => {
    if (timeLeft <= 0) {
      onDecline();
      return;
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, onDecline]);

  return (
    <div className="fixed bottom-6 right-6 z-50 animate-in slide-in-from-bottom-4 fade-in duration-300">
      <div className="flex w-80 flex-col gap-3 rounded-2xl border border-border bg-card p-4 shadow-lg">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Swords className="h-4 w-4 text-teal-600" />
            <span className="text-sm font-bold text-foreground">PK 邀请</span>
          </div>
          <button
            type="button"
            onClick={onDecline}
            className="flex h-6 w-6 items-center justify-center rounded-md hover:bg-muted"
          >
            <X className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>

        {/* Content */}
        <div className="flex flex-col gap-1">
          <p className="text-sm text-foreground">
            <span className="font-semibold">{initiatorName}</span> 邀请你进行 PK 对战
          </p>
          <p className="text-xs text-muted-foreground">
            {gameName} · {levelName}
          </p>
        </div>

        {/* Timer + Actions */}
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">{timeLeft}s</span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={onDecline}
              className="rounded-lg border border-border px-4 py-1.5 text-sm font-medium text-muted-foreground hover:bg-muted"
            >
              拒绝
            </button>
            <button
              type="button"
              onClick={onAccept}
              className="rounded-lg bg-teal-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-teal-700"
            >
              接受
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create invitation provider**

```tsx
// dx-web/src/features/web/play-pk/components/pk-invitation-provider.tsx
"use client";

import { useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useUserSSE } from "@/hooks/use-user-sse";
import { acceptPkInviteAction, declinePkInviteAction } from "../actions/invite.action";
import { PkInvitationPopup } from "./pk-invitation-popup";

interface PkInvitation {
  pk_id: string;
  game_id: string;
  game_name: string;
  game_mode: string;
  level_name: string;
  initiator_id: string;
  initiator_name: string;
}

export function PkInvitationProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [invitation, setInvitation] = useState<PkInvitation | null>(null);

  useUserSSE({
    pk_invitation: (data) => {
      const inv = data as PkInvitation;
      setInvitation(inv);
    },
  });

  const handleAccept = useCallback(async () => {
    if (!invitation) return;
    const result = await acceptPkInviteAction(invitation.pk_id);
    setInvitation(null);
    if (result.data) {
      const params = new URLSearchParams({
        degree: result.data.degree,
        level: result.data.game_level_id,
        pkId: invitation.pk_id,
        sessionId: result.data.session_id,
      });
      if (result.data.pattern) params.set("pattern", result.data.pattern);
      router.push(`/hall/pk-room/${invitation.pk_id}?${params}`);
    }
  }, [invitation, router]);

  const handleDecline = useCallback(async () => {
    if (!invitation) return;
    await declinePkInviteAction(invitation.pk_id);
    setInvitation(null);
  }, [invitation]);

  return (
    <>
      {children}
      {invitation && (
        <PkInvitationPopup
          pkId={invitation.pk_id}
          gameName={invitation.game_name}
          levelName={invitation.level_name}
          initiatorName={invitation.initiator_name}
          onAccept={handleAccept}
          onDecline={handleDecline}
        />
      )}
    </>
  );
}
```

- [ ] **Step 3: Add provider to web layout**

In `dx-web/src/app/(web)/layout.tsx`, wrap children with the provider:

```tsx
import { PkInvitationProvider } from "@/features/web/play-pk/components/pk-invitation-provider";

export default function WebLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <PkInvitationProvider>{children}</PkInvitationProvider>;
}
```

- [ ] **Step 4: Verify lint**

```bash
cd dx-web && npx eslint src/features/web/play-pk/components/pk-invitation-popup.tsx src/features/web/play-pk/components/pk-invitation-provider.tsx src/app/\(web\)/layout.tsx
```

Expected: No errors.

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx dx-web/src/features/web/play-pk/components/pk-invitation-provider.tsx dx-web/src/app/\(web\)/layout.tsx
git commit -m "feat: add PK invitation popup and global provider"
```

---

## Task 12: Frontend — GameModeCard UI Overhaul

**Files:**
- Modify: `dx-web/src/features/web/play-core/components/game-mode-card.tsx`

- [ ] **Step 1: Add new imports**

Add to the existing imports in `game-mode-card.tsx`:

```typescript
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Search, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { verifyOpponentAction, invitePkAction } from "@/features/web/play-pk/actions/invite.action";
```

- [ ] **Step 2: Add new state variables**

After the existing `selectedPkLevel` state, add:

```typescript
const [pkTab, setPkTab] = useState<"random" | "specified">("random");
const [specifiedUsername, setSpecifiedUsername] = useState("");
const [verifyResult, setVerifyResult] = useState<{
  userId: string;
  nickname: string;
  isOnline: boolean;
  isVip: boolean;
} | null>(null);
const [verifyError, setVerifyError] = useState<string | null>(null);
const [isVerifying, setIsVerifying] = useState(false);
```

- [ ] **Step 3: Add tab reset handler**

After the state declarations, add:

```typescript
function handleTabChange(value: string) {
  setPkTab(value as "random" | "specified");
  if (value === "random") {
    setSpecifiedUsername("");
    setVerifyResult(null);
    setVerifyError(null);
  } else {
    setSelectedDifficulty("normal");
  }
}

async function handleVerify() {
  if (!specifiedUsername.trim()) return;
  setIsVerifying(true);
  setVerifyResult(null);
  setVerifyError(null);
  const res = await verifyOpponentAction(specifiedUsername.trim());
  setIsVerifying(false);
  if (res.error) {
    setVerifyError(res.error);
    return;
  }
  if (res.data) {
    if (!res.data.is_online) {
      setVerifyError("对方不在线");
    } else if (!res.data.is_vip) {
      setVerifyError("对方会员已过期");
    } else {
      setVerifyResult({
        userId: res.data.user_id,
        nickname: res.data.nickname,
        isOnline: res.data.is_online,
        isVip: res.data.is_vip,
      });
    }
  }
}
```

- [ ] **Step 4: Update handlePkStart to handle both tabs**

Replace the existing `handlePkStart` function:

```typescript
function handlePkStart() {
  startTransition(async () => {
    if (pkTab === "specified") {
      if (!verifyResult) return;
      const pkLevel = selectedPkLevel || levels?.[0]?.id;
      const res = await invitePkAction({
        gameId,
        gameLevelId: pkLevel || "",
        degree: selectedDegree,
        pattern: isWordSentence ? selectedPattern : null,
        opponentId: verifyResult.userId,
      });
      if (res.error || !res.data) return;
      const params = new URLSearchParams({
        sessionId: res.data.session_id,
      });
      router.push(`/hall/pk-room/${res.data.pk_id}?${params}`);
    } else {
      const params = new URLSearchParams({ degree: selectedDegree, difficulty: selectedDifficulty });
      if (isWordSentence) params.set("pattern", selectedPattern);
      const pkLevel = selectedPkLevel || levels?.[0]?.id;
      if (pkLevel) params.set("level", pkLevel);
      router.push(`/hall/play-pk/${gameId}?${params}`);
    }
  });
}
```

- [ ] **Step 5: Replace the PK difficulty + level section**

Replace the entire `{/* Difficulty options (PK mode only) */}` block (from `{isPk && (` to the matching `)}`) with:

```tsx
{isPk && (
  <>
    <Tabs value={pkTab} onValueChange={handleTabChange} className="mt-4">
      <TabsList className="w-full">
        <TabsTrigger value="random" className="flex-1">随机对手</TabsTrigger>
        <TabsTrigger value="specified" className="flex-1">指定对手</TabsTrigger>
      </TabsList>

      <TabsContent value="random" className="flex flex-col gap-3 pt-3">
        {/* Difficulty selector — left-right style */}
        <div className="flex items-center gap-3">
          <Flame className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="shrink-0 text-[13px] font-medium text-foreground">对手强度</span>
          <Select value={selectedDifficulty} onValueChange={setSelectedDifficulty}>
            <SelectTrigger className="h-9 flex-1 text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {difficultyOptions.map(({ value, label }) => (
                <SelectItem key={value} value={value}>{label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {/* Starting level selector */}
        {levels && levels.length > 0 && (
          <div className="flex items-center gap-3">
            <Gamepad2 className="h-4 w-4 shrink-0 text-muted-foreground" />
            <span className="shrink-0 text-[13px] font-medium text-foreground">起始关卡</span>
            <Select value={selectedPkLevel} onValueChange={setSelectedPkLevel}>
              <SelectTrigger className="h-9 flex-1 text-sm">
                <SelectValue placeholder="选择关卡" />
              </SelectTrigger>
              <SelectContent>
                {levels.map((level) => (
                  <SelectItem key={level.id} value={level.id}>{level.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </TabsContent>

      <TabsContent value="specified" className="flex flex-col gap-3 pt-3">
        {/* Username input + verify button */}
        <div className="flex items-center gap-2">
          <Input
            value={specifiedUsername}
            onChange={(e) => {
              setSpecifiedUsername(e.target.value);
              setVerifyResult(null);
              setVerifyError(null);
            }}
            placeholder="输入对手用户名"
            className="h-9 flex-1 text-sm"
          />
          <button
            type="button"
            onClick={handleVerify}
            disabled={isVerifying || !specifiedUsername.trim()}
            className="flex h-9 shrink-0 items-center gap-1.5 rounded-md bg-teal-600 px-3 text-sm font-medium text-white disabled:opacity-50"
          >
            {isVerifying ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Search className="h-3.5 w-3.5" />
            )}
            验证
          </button>
        </div>
        {/* Verify result */}
        {verifyResult && (
          <div className="flex items-center gap-2 text-sm text-emerald-600">
            <CheckCircle2 className="h-4 w-4" />
            <span>{verifyResult.nickname} · 在线</span>
          </div>
        )}
        {verifyError && (
          <div className="flex items-center gap-2 text-sm text-red-500">
            <XCircle className="h-4 w-4" />
            <span>{verifyError}</span>
          </div>
        )}
        {/* Starting level selector */}
        {levels && levels.length > 0 && (
          <div className="flex items-center gap-3">
            <Gamepad2 className="h-4 w-4 shrink-0 text-muted-foreground" />
            <span className="shrink-0 text-[13px] font-medium text-foreground">起始关卡</span>
            <Select value={selectedPkLevel} onValueChange={setSelectedPkLevel}>
              <SelectTrigger className="h-9 flex-1 text-sm">
                <SelectValue placeholder="选择关卡" />
              </SelectTrigger>
              <SelectContent>
                {levels.map((level) => (
                  <SelectItem key={level.id} value={level.id}>{level.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </TabsContent>
    </Tabs>

    <div className="h-px bg-border my-5" />
  </>
)}
```

- [ ] **Step 6: Update the PK start button disabled state**

In the action buttons section, update the PK start button to be disabled when specified tab has no verified opponent:

Replace:
```tsx
disabled={isPending}
```

With:
```tsx
disabled={isPending || (pkTab === "specified" && !verifyResult)}
```

- [ ] **Step 7: Verify lint**

```bash
cd dx-web && npx eslint src/features/web/play-core/components/game-mode-card.tsx
```

Expected: No errors.

- [ ] **Step 8: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds.

- [ ] **Step 9: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-mode-card.tsx
git commit -m "feat: add Tabs (随机对手/指定对手) and restyle difficulty to left-right layout"
```

---

## Task 13: Frontend — PK Room Page

**Files:**
- Create: `dx-web/src/app/(web)/hall/pk-room/[id]/page.tsx`

- [ ] **Step 1: Create PK room page**

```tsx
// dx-web/src/app/(web)/hall/pk-room/[id]/page.tsx
"use client";

import { useEffect, useState, useCallback, use } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Swords, Loader2, ArrowLeft, User } from "lucide-react";
import { usePkSSE } from "@/hooks/use-pk-sse";
import { fetchPkDetailsAction } from "@/features/web/play-pk/actions/invite.action";
import { endPkAction } from "@/features/web/play-pk/actions/session.action";
import { getAvatarColor } from "@/lib/avatar";

type PkDetails = {
  pk_id: string;
  game_id: string;
  game_name: string;
  game_mode: string;
  level_id: string;
  level_name: string;
  degree: string;
  pattern: string | null;
  initiator_id: string;
  initiator_name: string;
  opponent_id: string;
  opponent_name: string;
  invitation_status: string;
};

export default function PkRoomPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id: pkId } = use(params);
  const router = useRouter();
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("sessionId") ?? "";

  const [details, setDetails] = useState<PkDetails | null>(null);
  const [status, setStatus] = useState<"waiting" | "accepted" | "declined" | "timeout">("waiting");
  const [timeLeft, setTimeLeft] = useState(30);
  const [countdown, setCountdown] = useState<number | null>(null);

  // Fetch PK details
  useEffect(() => {
    fetchPkDetailsAction(pkId).then((res) => {
      if (res.data) {
        setDetails(res.data);
        if (res.data.invitation_status === "accepted") {
          setStatus("accepted");
          setCountdown(1);
        }
      }
    });
  }, [pkId]);

  // 30s waiting timeout
  useEffect(() => {
    if (status !== "waiting") return;
    if (timeLeft <= 0) {
      setStatus("timeout");
      endPkAction(pkId);
      return;
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, status, pkId]);

  // Countdown after accepted → navigate to play
  useEffect(() => {
    if (countdown === null || !details) return;
    if (countdown <= 0) {
      const params = new URLSearchParams({
        degree: details.degree,
        level: details.level_id,
        pkId: details.pk_id,
        sessionId,
      });
      if (details.pattern) params.set("pattern", details.pattern);
      router.push(`/hall/play-pk/${details.game_id}?${params}`);
      return;
    }
    const timer = setTimeout(() => setCountdown((c) => (c ?? 1) - 1), 1000);
    return () => clearTimeout(timer);
  }, [countdown, details, sessionId, router]);

  // SSE listeners
  usePkSSE(status === "waiting" ? pkId : null, {
    pk_invitation_accepted: (data: unknown) => {
      const event = data as { opponent_name: string };
      setStatus("accepted");
      if (details) {
        setDetails({ ...details, opponent_name: event.opponent_name, invitation_status: "accepted" });
      }
      setCountdown(1);
    },
    pk_invitation_declined: () => {
      setStatus("declined");
    },
  });

  const handleCancel = useCallback(async () => {
    await endPkAction(pkId);
    router.back();
  }, [pkId, router]);

  const handleBack = useCallback(() => {
    router.back();
  }, [router]);

  if (!details) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
      </div>
    );
  }

  const initiatorColor = getAvatarColor(details.initiator_name);
  const opponentColor = getAvatarColor(details.opponent_name);

  return (
    <div className="flex h-screen w-full flex-col items-center justify-center gap-8 bg-[radial-gradient(ellipse_at_center,#1E1B4B,#0F0A2E)] px-4">
      {/* Game info */}
      <div className="flex flex-col items-center gap-2">
        <Swords className="h-8 w-8 text-teal-400" />
        <h1 className="text-2xl font-bold text-white">{details.game_name}</h1>
        <p className="text-sm text-slate-400">{details.level_name} · PK 对战</p>
      </div>

      {/* Player slots */}
      <div className="flex items-center gap-8">
        {/* Initiator */}
        <div className="flex flex-col items-center gap-3">
          <div
            className="flex h-16 w-16 items-center justify-center rounded-full text-xl font-bold text-white"
            style={{ backgroundColor: initiatorColor }}
          >
            {details.initiator_name[0]}
          </div>
          <span className="text-sm font-medium text-white">{details.initiator_name}</span>
        </div>

        <span className="text-2xl font-bold text-slate-500">VS</span>

        {/* Opponent */}
        <div className="flex flex-col items-center gap-3">
          {status === "accepted" || status === "waiting" && details.invitation_status === "accepted" ? (
            <div
              className="flex h-16 w-16 items-center justify-center rounded-full text-xl font-bold text-white"
              style={{ backgroundColor: opponentColor }}
            >
              {details.opponent_name[0]}
            </div>
          ) : (
            <div className="flex h-16 w-16 items-center justify-center rounded-full border-2 border-dashed border-slate-600">
              <User className="h-6 w-6 text-slate-600" />
            </div>
          )}
          <span className="text-sm font-medium text-slate-400">
            {status === "accepted" ? details.opponent_name : "等待对手..."}
          </span>
        </div>
      </div>

      {/* Status messages */}
      {status === "waiting" && (
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="h-5 w-5 animate-spin text-teal-400" />
          <p className="text-sm text-slate-400">等待对方接受邀请... {timeLeft}s</p>
          <button
            type="button"
            onClick={handleCancel}
            className="mt-2 flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            取消
          </button>
        </div>
      )}

      {status === "accepted" && countdown !== null && (
        <p className="text-lg font-bold text-teal-400">
          {countdown > 0 ? `${countdown}s 后开始...` : "开始!"}
        </p>
      )}

      {status === "declined" && (
        <div className="flex flex-col items-center gap-3">
          <p className="text-sm font-medium text-red-400">对方已拒绝</p>
          <button
            type="button"
            onClick={handleBack}
            className="flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </button>
        </div>
      )}

      {status === "timeout" && (
        <div className="flex flex-col items-center gap-3">
          <p className="text-sm font-medium text-amber-400">对方未响应</p>
          <button
            type="button"
            onClick={handleBack}
            className="flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5 text-sm font-medium text-white hover:bg-white/15"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </button>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Verify lint**

```bash
cd dx-web && npx eslint src/app/\(web\)/hall/pk-room/\[id\]/page.tsx
```

Expected: No errors.

- [ ] **Step 3: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/pk-room/\[id\]/page.tsx
git commit -m "feat: add PK room waiting page with SSE + countdown"
```

---

## Task 14: Frontend — Play Page Modifications for Specified PK

**Files:**
- Modify: `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx`
- Modify: `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx`

- [ ] **Step 1: Update play-pk page to pass pkId and sessionId**

In `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx`, extract new params and pass to shell:

After the existing param extraction:
```typescript
const pkId = searchParams.get("pkId");
const sessionId = searchParams.get("sessionId");
```

Update the `PkPlayShell` component to receive these:
```tsx
<PkPlayShell
  game={game}
  player={player}
  degree={degree}
  pattern={pattern}
  levelId={targetLevelId}
  difficulty={difficulty}
  pkId={pkId}
  sessionId={sessionId}
/>
```

- [ ] **Step 2: Update PkPlayShell props**

In `dx-web/src/features/web/play-pk/components/pk-play-shell.tsx`, add `pkId` and `sessionId` to the interface and pass them to the loading screen.

Add to the interface:
```typescript
pkId?: string | null;
sessionId?: string | null;
```

Pass to `PkPlayLoadingScreen`:
```tsx
<PkPlayLoadingScreen
  // ...existing props
  existingPkId={pkId}
  existingSessionId={sessionId}
/>
```

- [ ] **Step 3: Update loading screen to skip startPkAction when pkId provided**

In `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx`, add new props:

```typescript
existingPkId?: string | null;
existingSessionId?: string | null;
```

In the `loadGameData` function, modify Step 1 to branch:

```typescript
// Step 1: Start PK or use existing session
let pkData: {
  pk_id: string;
  session_id: string;
  game_level_id: string;
  opponent_id: string;
  opponent_name: string;
  robot_completed: boolean;
};

if (existingPkId && existingSessionId) {
  // Specified PK — session already created during invite/accept
  pkData = {
    pk_id: existingPkId,
    session_id: existingSessionId,
    game_level_id: levelId,
    opponent_id: "",
    opponent_name: "",
    robot_completed: false,
  };
  // Fetch PK details for opponent info
  const details = await fetchPkDetailsAction(existingPkId);
  if (cancelled) return;
  if (details.data) {
    pkData.opponent_id = details.data.opponent_id;
    pkData.opponent_name = details.data.opponent_name;
  }
} else {
  // Random PK — existing flow
  const pkResult = await startPkAction(gameId, levelId, degree, pattern, difficulty);
  if (cancelled) return;
  if (pkResult.error || !pkResult.data) {
    setError(pkResult.error ?? "无法开始PK");
    return;
  }
  pkData = pkResult.data;
}
setProgress(33);
```

Then use `pkData` instead of `pkResult.data` throughout the rest of the function.

Add import at the top:
```typescript
import { fetchPkDetailsAction } from "../actions/invite.action";
```

- [ ] **Step 4: Verify lint**

```bash
cd dx-web && npx eslint src/app/\(web\)/hall/play-pk/\[id\]/page.tsx src/features/web/play-pk/components/pk-play-loading-screen.tsx src/features/web/play-pk/components/pk-play-shell.tsx
```

Expected: No errors.

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/play-pk/\[id\]/page.tsx dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx dx-web/src/features/web/play-pk/components/pk-play-shell.tsx
git commit -m "feat: support specified PK in play page (skip startPkAction when pkId provided)"
```

---

## Task 15: Full Integration Verification

**Files:** None (verification only)

- [ ] **Step 1: Verify Go backend builds**

```bash
cd dx-api && go build ./...
```

Expected: Build succeeds.

- [ ] **Step 2: Verify Go vet**

```bash
cd dx-api && go vet ./...
```

Expected: No issues.

- [ ] **Step 3: Verify frontend lint**

```bash
cd dx-web && npm run lint
```

Expected: No errors.

- [ ] **Step 4: Verify frontend build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds.

- [ ] **Step 5: Commit any remaining fixes**

If any lint or build fixes were needed, commit them:

```bash
git add -A
git commit -m "fix: resolve lint and build issues from PK specified opponent feature"
```
