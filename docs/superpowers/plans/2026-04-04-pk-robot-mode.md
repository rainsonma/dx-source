# PK Robot Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a PK button on the game details page that lets users compete against a robot opponent in real-time, reusing the group play UI pattern.

**Architecture:** New `game_pks` bridge table links PK matches to existing session tables via `game_pk_id`. A backend goroutine simulates the robot playing in real-time, broadcasting progress via a PK-scoped SSE hub. The frontend mirrors the group play shell with PK-specific adaptations (2-player podium, timeout countdown, pause/resume).

**Tech Stack:** Go/Goravel backend, Next.js 16 frontend, PostgreSQL, Zustand, SSE (Server-Sent Events)

**Spec:** `docs/superpowers/specs/2026-04-04-pk-robot-mode-design.md`

---

## File Structure

### Backend — New files

| File | Responsibility |
|------|----------------|
| `dx-api/app/consts/pk.go` | PK difficulty presets and constants |
| `dx-api/app/models/game_pk.go` | GamePk model struct |
| `dx-api/app/helpers/sse_pk_hub.go` | SSE hub scoped to PK matches |
| `dx-api/app/services/api/mock_user_service.go` | Find/create idle mock users |
| `dx-api/app/services/api/pk_winner_service.go` | Per-level winner determination (2 players) |
| `dx-api/app/services/api/game_play_pk_service.go` | PK session lifecycle + robot goroutine |
| `dx-api/app/http/requests/api/pk_request.go` | Request validation structs |
| `dx-api/app/http/controllers/api/game_play_pk_controller.go` | PK API endpoints |

### Backend — Modified files

| File | Change |
|------|--------|
| `dx-api/app/models/game_session_total.go` | Add `GamePkID *string` field |
| `dx-api/app/models/game_session_level.go` | Add `GamePkID *string` field |
| `dx-api/app/services/api/errors.go` | Add PK error sentinels |
| `dx-api/app/consts/error_code.go` | Add PK error codes |
| `dx-api/app/helpers/enum_rules.go` | Add `pk_difficulty` enum |
| `dx-api/routes/api.go` | Add PK route group |

### Frontend — New files

| File | Responsibility |
|------|----------------|
| `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx` | Route page |
| `dx-web/src/features/web/play-pk/types/pk-play.ts` | TypeScript types for PK events |
| `dx-web/src/features/web/play-pk/actions/session.action.ts` | Server actions for PK endpoints |
| `dx-web/src/hooks/use-pk-sse.ts` | SSE connection hook for PK |
| `dx-web/src/features/web/play-pk/hooks/use-pk-play-events.ts` | SSE event listeners |
| `dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts` | Zustand store |
| `dx-web/src/features/web/play-pk/components/pk-play-shell.tsx` | Main PK shell |
| `dx-web/src/features/web/play-pk/components/pk-play-top-bar.tsx` | Top bar with pause/exit |
| `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx` | Loading screen |
| `dx-web/src/features/web/play-pk/components/pk-play-waiting-screen.tsx` | Waiting for opponent |
| `dx-web/src/features/web/play-pk/components/pk-play-result-panel.tsx` | 2-player result podium |

### Frontend — Modified files

| File | Change |
|------|--------|
| `dx-web/src/features/web/games/components/hero-card.tsx` | Add PK button before 群组 |
| `dx-web/src/features/web/play-core/components/game-mode-card.tsx` | Add difficulty selector for PK mode |

---

## Task 1: Backend Constants & Enums

**Files:**
- Create: `dx-api/app/consts/pk.go`
- Modify: `dx-api/app/helpers/enum_rules.go`

- [ ] **Step 1: Create PK constants**

```go
// dx-api/app/consts/pk.go
package consts

const (
	PkDifficultyEasy   = "easy"
	PkDifficultyNormal = "normal"
	PkDifficultyHard   = "hard"
)

type PkDifficultyParams struct {
	AccuracyMin     float64
	AccuracyMax     float64
	MinDelayMs      int
	MaxDelayMs      int
	ComboBreakPct   float64
}

var PkDifficulties = map[string]PkDifficultyParams{
	PkDifficultyEasy:   {AccuracyMin: 0.50, AccuracyMax: 0.70, MinDelayMs: 3000, MaxDelayMs: 6000, ComboBreakPct: 0.50},
	PkDifficultyNormal: {AccuracyMin: 0.70, AccuracyMax: 0.85, MinDelayMs: 2000, MaxDelayMs: 4000, ComboBreakPct: 0.30},
	PkDifficultyHard:   {AccuracyMin: 0.85, AccuracyMax: 0.95, MinDelayMs: 1000, MaxDelayMs: 3000, ComboBreakPct: 0.10},
}

var PkDifficultySlugs = []string{PkDifficultyEasy, PkDifficultyNormal, PkDifficultyHard}

const (
	PkTimeoutDuration  = 5 * 60 // 5 minutes in seconds
	PkTimeoutWarning   = 30     // warning countdown in seconds
)
```

- [ ] **Step 2: Register enum in helpers**

In `dx-api/app/helpers/enum_rules.go`, add to the `enumValues` map:

```go
"pk_difficulty": consts.PkDifficultySlugs,
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build, no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/consts/pk.go dx-api/app/helpers/enum_rules.go
git commit -m "feat: add PK difficulty constants and enum"
```

---

## Task 2: Backend Model & Schema

**Files:**
- Create: `dx-api/app/models/game_pk.go`
- Modify: `dx-api/app/models/game_session_total.go`
- Modify: `dx-api/app/models/game_session_level.go`

- [ ] **Step 1: Create GamePk model**

```go
// dx-api/app/models/game_pk.go
package models

import "github.com/goravel/framework/database/orm"

type GamePk struct {
	orm.Timestamps
	ID              string  `gorm:"column:id;primaryKey" json:"id"`
	UserID          string  `gorm:"column:user_id" json:"user_id"`
	OpponentID      string  `gorm:"column:opponent_id" json:"opponent_id"`
	GameID          string  `gorm:"column:game_id" json:"game_id"`
	Degree          string  `gorm:"column:degree" json:"degree"`
	Pattern         *string `gorm:"column:pattern" json:"pattern"`
	RobotDifficulty string  `gorm:"column:robot_difficulty" json:"robot_difficulty"`
	CurrentLevelID  *string `gorm:"column:current_level_id" json:"current_level_id"`
	IsPlaying       bool    `gorm:"column:is_playing" json:"is_playing"`
	LastWinnerID    *string `gorm:"column:last_winner_id" json:"last_winner_id"`
}

func (g *GamePk) TableName() string {
	return "game_pks"
}
```

- [ ] **Step 2: Add GamePkID to GameSessionTotal**

In `dx-api/app/models/game_session_total.go`, add after the `GameSubgroupID` field:

```go
GamePkID           *string    `gorm:"column:game_pk_id" json:"game_pk_id"`
```

- [ ] **Step 3: Add GamePkID to GameSessionLevel**

In `dx-api/app/models/game_session_level.go`, add after the `GameSubgroupID` field:

```go
GamePkID           *string    `gorm:"column:game_pk_id" json:"game_pk_id"`
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/models/game_pk.go dx-api/app/models/game_session_total.go dx-api/app/models/game_session_level.go
git commit -m "feat: add GamePk model and game_pk_id to session tables"
```

---

## Task 3: Error Sentinels & Codes

**Files:**
- Modify: `dx-api/app/services/api/errors.go`
- Modify: `dx-api/app/consts/error_code.go`

- [ ] **Step 1: Add PK error sentinels**

In `dx-api/app/services/api/errors.go`, add to the `var` block:

```go
ErrPkNotFound         = errors.New("PK对战不存在")
ErrPkIsPlaying        = errors.New("PK对战进行中")
ErrPkNotPlaying       = errors.New("没有进行中的PK对战")
ErrNoMockUserAvail    = errors.New("没有可用的对手，请稍后再试")
```

- [ ] **Step 2: Add PK error codes**

In `dx-api/app/consts/error_code.go`, add:

```go
// 404xx: Not Found (add after CodeOrderNotFound)
CodePkNotFound = 40412

// 400xx: Validation (add after CodeInvalidProduct)
CodePkIsPlaying   = 40015
CodePkNotPlaying  = 40016
CodeNoMockUser    = 40017
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/errors.go dx-api/app/consts/error_code.go
git commit -m "feat: add PK error sentinels and error codes"
```

---

## Task 4: PK SSE Hub

**Files:**
- Create: `dx-api/app/helpers/sse_pk_hub.go`

- [ ] **Step 1: Create PkSSEHub**

```go
// dx-api/app/helpers/sse_pk_hub.go
package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// PkSSEHub manages SSE connections for PK matches.
type PkSSEHub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*SSEConnection // pkID -> userID -> conn
}

var PkHub = &PkSSEHub{
	conns: make(map[string]map[string]*SSEConnection),
}

func (h *PkSSEHub) Register(pkID, userID string, w http.ResponseWriter) *SSEConnection {
	conn := NewSSEConnection(w)

	h.mu.Lock()
	if h.conns[pkID] == nil {
		h.conns[pkID] = make(map[string]*SSEConnection)
	}
	if old, ok := h.conns[pkID][userID]; ok {
		close(old.done)
	}
	h.conns[pkID][userID] = conn
	h.mu.Unlock()

	return conn
}

func (h *PkSSEHub) Unregister(pkID, userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[pkID][userID]
	if exists && current == conn {
		delete(h.conns[pkID], userID)
		if len(h.conns[pkID]) == 0 {
			delete(h.conns, pkID)
		}
	}
	h.mu.Unlock()
}

func (h *PkSSEHub) Broadcast(pkID, event string, data any) {
	jsonBytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	if pk, ok := h.conns[pkID]; ok {
		for _, conn := range pk {
			fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event, jsonBytes)
			if conn.flusher != nil {
				conn.flusher.Flush()
			}
		}
	}
}

func (h *PkSSEHub) IsConnected(pkID, userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if pk, ok := h.conns[pkID]; ok {
		_, exists := pk[userID]
		return exists
	}
	return false
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/helpers/sse_pk_hub.go
git commit -m "feat: add PK SSE hub for real-time PK events"
```

---

## Task 5: Mock User Service

**Files:**
- Create: `dx-api/app/services/api/mock_user_service.go`

- [ ] **Step 1: Create mock user service**

```go
// dx-api/app/services/api/mock_user_service.go
package api

import (
	"fmt"
	"math/rand"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	"golang.org/x/crypto/bcrypt"
)

var englishFirstNames = []string{
	"james", "emma", "oliver", "sophia", "liam", "ava", "noah", "mia",
	"lucas", "lily", "ethan", "chloe", "mason", "sarah", "logan", "emily",
	"jack", "grace", "henry", "alice", "leo", "ruby", "oscar", "ella",
	"charlie", "hannah", "max", "aria", "sam", "luna", "ben", "zoe",
}

var chineseNames = []string{
	"小明", "小红", "小华", "小丽", "小龙", "小凤", "小杰", "小雨",
	"小雪", "小云", "小星", "小月", "小天", "小海", "小风", "小林",
}

var chineseSurnames = []string{
	"wang", "li", "zhang", "liu", "chen", "yang", "huang", "wu",
}

// FindOrCreateMockUser returns an idle mock user or creates a new one.
func FindOrCreateMockUser() (*models.User, error) {
	var user models.User
	err := facades.Orm().Query().Raw(
		`SELECT u.* FROM users u
		 WHERE u.is_mock = true
		   AND NOT EXISTS (
		     SELECT 1 FROM game_pks gp
		     WHERE gp.opponent_id = u.id AND gp.is_playing = true
		   )
		 ORDER BY RANDOM() LIMIT 1`).Scan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to query mock users: %w", err)
	}
	if user.ID != "" {
		return &user, nil
	}
	return createMockUser()
}

func createMockUser() (*models.User, error) {
	username := generateMockUsername()
	nickname := generateMockNickname()

	hashed, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		ID:       uuid.Must(uuid.NewV7()).String(),
		Grade:    "month",
		Username: username,
		Nickname: &nickname,
		Password: string(hashed),
		IsActive: true,
		IsMock:   true,
		InviteCode: uuid.NewString()[:8],
	}

	if err := facades.Orm().Query().Create(&user); err != nil {
		return nil, fmt.Errorf("failed to create mock user: %w", err)
	}
	return &user, nil
}

func generateMockUsername() string {
	name := englishFirstNames[rand.Intn(len(englishFirstNames))]
	suffix := rand.Intn(9000) + 1000
	return fmt.Sprintf("%s%d", name, suffix)
}

func generateMockNickname() string {
	separators := []string{"-", "_", ""}
	sep := separators[rand.Intn(len(separators))]

	switch rand.Intn(4) {
	case 0:
		// Pure English: "Emma", "Jack"
		return englishFirstNames[rand.Intn(len(englishFirstNames))]
	case 1:
		// Pure Chinese: "小明", "小红"
		return chineseNames[rand.Intn(len(chineseNames))]
	case 2:
		// English + Chinese surname: "Emma_Li", "Jack-zhang"
		en := englishFirstNames[rand.Intn(len(englishFirstNames))]
		cn := chineseSurnames[rand.Intn(len(chineseSurnames))]
		return en + sep + cn
	default:
		// Chinese + English surname: "小明-wang", "小红_li"
		cn := chineseNames[rand.Intn(len(chineseNames))]
		en := chineseSurnames[rand.Intn(len(chineseSurnames))]
		return cn + sep + en
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/mock_user_service.go
git commit -m "feat: add mock user service for PK robot opponents"
```

---

## Task 6: PK Winner Service

**Files:**
- Create: `dx-api/app/services/api/pk_winner_service.go`

- [ ] **Step 1: Create PK winner service**

```go
// dx-api/app/services/api/pk_winner_service.go
package api

import (
	"fmt"
	"time"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

type PkWinnerResult struct {
	GameLevelID  string       `json:"game_level_id"`
	Winner       PkWinner     `json:"winner"`
	Participants []PkWinner   `json:"participants"`
}

type PkWinner struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

// DeterminePkWinner compares the two players' level scores for a given PK match and level.
func DeterminePkWinner(pkID, gameLevelID string) (*PkWinnerResult, error) {
	type scoreRow struct {
		UserID   string     `gorm:"column:user_id"`
		Nickname *string    `gorm:"column:nickname"`
		Score    int        `gorm:"column:score"`
		EndedAt  *time.Time `gorm:"column:ended_at"`
	}

	var rows []scoreRow
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, u.nickname, gsl.score, gsl.ended_at
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 JOIN users u ON u.id = gst.user_id
		 WHERE gsl.game_pk_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		 ORDER BY gsl.score DESC, gsl.ended_at ASC`,
		pkID, gameLevelID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query pk scores: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil
	}

	participants := make([]PkWinner, len(rows))
	for i, r := range rows {
		name := ""
		if r.Nickname != nil {
			name = *r.Nickname
		}
		participants[i] = PkWinner{
			UserID:   r.UserID,
			UserName: name,
			Score:    r.Score,
		}
	}

	winner := participants[0]

	// Update last_winner_id on game_pks
	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("last_winner_id", winner.UserID)

	return &PkWinnerResult{
		GameLevelID:  gameLevelID,
		Winner:       winner,
		Participants: participants,
	}, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/pk_winner_service.go
git commit -m "feat: add PK winner determination service"
```

---

## Task 7: PK Play Service (Core — Robot Goroutine)

**Files:**
- Create: `dx-api/app/services/api/game_play_pk_service.go`

- [ ] **Step 1: Create PK play service**

This is the largest file. It handles:
1. Starting a PK match (create `game_pks` record, select robot, create sessions)
2. Robot goroutine simulation
3. Level lifecycle (start, complete, next, end)
4. Pause/resume
5. Timeout handling

```go
// dx-api/app/services/api/game_play_pk_service.go
package api

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

// robotState tracks per-PK goroutine state for pause/cancel.
type robotState struct {
	cancel  context.CancelFunc
	pauseCh chan struct{}
	paused  bool
	mu      sync.Mutex
}

var (
	robotsMu sync.Mutex
	robots   = make(map[string]*robotState) // pkID -> state
)

func getRobot(pkID string) *robotState {
	robotsMu.Lock()
	defer robotsMu.Unlock()
	return robots[pkID]
}

func setRobot(pkID string, rs *robotState) {
	robotsMu.Lock()
	defer robotsMu.Unlock()
	robots[pkID] = rs
}

func removeRobot(pkID string) {
	robotsMu.Lock()
	defer robotsMu.Unlock()
	delete(robots, pkID)
}

// --- Public API ---

type PkStartResult struct {
	PkID       string `json:"pk_id"`
	SessionID  string `json:"session_id"`
	OpponentID string `json:"opponent_id"`
	OpponentName string `json:"opponent_name"`
}

// StartPk creates a PK match, selects a robot, creates sessions for both players,
// and spawns the robot goroutine for the first level.
func StartPk(userID, gameID, degree string, pattern *string, levelID *string, difficulty string) (*PkStartResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	// Verify game exists and is published
	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).Where("status", consts.GameStatusPublished).First(&game); err != nil || game.ID == "" {
		return nil, ErrGameNotFound
	}

	// Find or create idle mock user
	mockUser, err := FindOrCreateMockUser()
	if err != nil {
		return nil, ErrNoMockUserAvail
	}

	// Resolve starting level
	var startLevel models.GameLevel
	if levelID != nil && *levelID != "" {
		if err := facades.Orm().Query().Where("id", *levelID).Where("game_id", gameID).Where("is_active", true).First(&startLevel); err != nil || startLevel.ID == "" {
			return nil, ErrLevelNotFound
		}
	} else {
		if err := facades.Orm().Query().Where("game_id", gameID).Where("is_active", true).Order("\"order\" ASC").First(&startLevel); err != nil || startLevel.ID == "" {
			return nil, ErrNoGameLevels
		}
	}

	// Count total levels
	var totalLevels int64
	facades.Orm().Query().Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count(&totalLevels)

	now := time.Now()
	pkID := uuid.Must(uuid.NewV7()).String()

	pk := models.GamePk{
		ID:              pkID,
		UserID:          userID,
		OpponentID:      mockUser.ID,
		GameID:          gameID,
		Degree:          degree,
		Pattern:         pattern,
		RobotDifficulty: difficulty,
		CurrentLevelID:  &startLevel.ID,
		IsPlaying:       true,
	}
	if err := facades.Orm().Query().Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create pk: %w", err)
	}

	// Create human's session
	humanSessionID := uuid.Must(uuid.NewV7()).String()
	humanSession := models.GameSessionTotal{
		ID:               humanSessionID,
		UserID:           userID,
		GameID:           gameID,
		Degree:           degree,
		Pattern:          pattern,
		CurrentLevelID:   &startLevel.ID,
		TotalLevelsCount: int(totalLevels),
		StartedAt:        now,
		LastPlayedAt:     now,
		GamePkID:         &pkID,
	}
	if err := facades.Orm().Query().Create(&humanSession); err != nil {
		return nil, fmt.Errorf("failed to create human session: %w", err)
	}

	// Upsert game stats for human
	upsertGameStats(userID, gameID, now)

	nickname := ""
	if mockUser.Nickname != nil {
		nickname = *mockUser.Nickname
	}

	return &PkStartResult{
		PkID:         pkID,
		SessionID:    humanSessionID,
		OpponentID:   mockUser.ID,
		OpponentName: nickname,
	}, nil
}

// StartPkLevel creates a level session for the human player and spawns the robot goroutine.
func StartPkLevel(userID, sessionID, gameLevelID, degree string, pattern *string) (*StartLevelResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	var session models.GameSessionTotal
	if err := facades.Orm().Query().Where("id", sessionID).Where("user_id", userID).First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}
	if session.GamePkID == nil {
		return nil, ErrForbidden
	}

	// Count content items for level
	totalItemsCount := countContentItems(session.GameID, gameLevelID, degree)

	now := time.Now()
	levelSessionID := uuid.Must(uuid.NewV7()).String()
	levelSession := models.GameSessionLevel{
		ID:                 levelSessionID,
		GameSessionTotalID: sessionID,
		GameLevelID:        gameLevelID,
		Degree:             degree,
		Pattern:            pattern,
		TotalItemsCount:    totalItemsCount,
		StartedAt:          now,
		LastPlayedAt:       now,
		GamePkID:           session.GamePkID,
	}
	if err := facades.Orm().Query().Create(&levelSession); err != nil {
		return nil, fmt.Errorf("failed to create level session: %w", err)
	}

	// Update session current level
	facades.Orm().Query().Model(&models.GameSessionTotal{}).
		Where("id", sessionID).
		Update("current_level_id", gameLevelID)

	// Spawn robot goroutine for this level
	go spawnRobotForLevel(*session.GamePkID, session.GameID, gameLevelID, degree)

	return &StartLevelResult{
		LevelSessionID: levelSessionID,
	}, nil
}

// CompletePkLevel completes the human's level and checks for winner.
func CompletePkLevel(userID, sessionID, gameLevelID string, score, maxCombo, totalItems int) (*CompleteLevelResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	var session models.GameSessionTotal
	if err := facades.Orm().Query().Where("id", sessionID).Where("user_id", userID).First(&session); err != nil || session.ID == "" {
		return nil, ErrSessionNotFound
	}
	if session.GamePkID == nil {
		return nil, ErrForbidden
	}

	pkID := *session.GamePkID

	// Find active level session
	var levelSession models.GameSessionLevel
	if err := facades.Orm().Query().
		Where("game_session_total_id", sessionID).
		Where("game_level_id", gameLevelID).
		Where("ended_at IS NULL").
		First(&levelSession); err != nil || levelSession.ID == "" {
		return nil, ErrSessionLevelNotFound
	}

	// Calculate accuracy and EXP
	accuracy := float64(0)
	if totalItems > 0 {
		accuracy = float64(levelSession.CorrectCount) / float64(totalItems)
	}
	expEarned := 0
	if accuracy >= consts.ExpAccuracyThreshold {
		expEarned = consts.LevelCompleteExp
	}

	now := time.Now()

	// Update level session
	facades.Orm().Query().Model(&models.GameSessionLevel{}).
		Where("id", levelSession.ID).
		Updates(map[string]any{
			"ended_at":  now,
			"score":     score,
			"exp":       expEarned,
			"max_combo": maxCombo,
		})

	// Update session total
	facades.Orm().Query().Model(&models.GameSessionTotal{}).
		Where("id", sessionID).
		Updates(map[string]any{
			"played_levels_count": session.PlayedLevelsCount + 1,
			"last_played_at":     now,
		})

	// Update stats
	updateLevelStats(userID, session.GameID, gameLevelID, score, expEarned, now)
	if expEarned > 0 {
		facades.Orm().Query().Exec("UPDATE users SET exp = exp + ? WHERE id = ?", expEarned, userID)
	}

	// Broadcast human completed
	var user models.User
	facades.Orm().Query().Where("id", userID).First(&user)
	userName := ""
	if user.Nickname != nil {
		userName = *user.Nickname
	}
	helpers.PkHub.Broadcast(pkID, "pk_player_complete", map[string]string{
		"user_id":       userID,
		"user_name":     userName,
		"game_level_id": gameLevelID,
	})

	// Check if robot also completed — determine winner
	result, _ := DeterminePkWinner(pkID, gameLevelID)
	if result != nil {
		helpers.PkHub.Broadcast(pkID, "pk_level_complete", result)

		// Cancel robot timeout goroutine since winner is determined
		if rs := getRobot(pkID); rs != nil {
			rs.cancel()
		}

		// Check if all levels completed
		var totalLevels int64
		facades.Orm().Query().Model(&models.GameLevel{}).
			Where("game_id", session.GameID).Where("is_active", true).Count(&totalLevels)
		if totalLevels > 0 && int64(session.PlayedLevelsCount+1) >= totalLevels {
			facades.Orm().Query().Model(&models.GamePk{}).
				Where("id", pkID).Update("is_playing", false)
			removeRobot(pkID)
		}
	}

	return &CompleteLevelResult{
		ExpEarned:      expEarned,
		Accuracy:       accuracy,
		MeetsThreshold: accuracy >= consts.ExpAccuracyThreshold,
	}, nil
}

// NextPkLevel advances to the next level.
func NextPkLevel(userID, pkID, currentLevelID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).Where("user_id", userID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}
	if !pk.IsPlaying {
		return ErrPkNotPlaying
	}

	var currentLevel models.GameLevel
	if err := facades.Orm().Query().Where("id", currentLevelID).First(&currentLevel); err != nil || currentLevel.ID == "" {
		return ErrLevelNotFound
	}

	var nextLevel models.GameLevel
	if err := facades.Orm().Query().
		Where("game_id", pk.GameID).
		Where("is_active", true).
		Where("\"order\" > ?", currentLevel.Order).
		Order("\"order\" ASC").
		First(&nextLevel); err != nil || nextLevel.ID == "" {
		return ErrLastLevel
	}

	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("current_level_id", nextLevel.ID)

	helpers.PkHub.Broadcast(pkID, "pk_next_level", map[string]any{
		"pk_id":      pkID,
		"game_id":    pk.GameID,
		"level_id":   nextLevel.ID,
		"level_name": nextLevel.Name,
		"degree":     pk.Degree,
		"pattern":    pk.Pattern,
	})

	return nil
}

// EndPk ends a PK match, cancels the robot, and frees resources.
func EndPk(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).Where("user_id", userID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}

	// Cancel robot goroutine
	if rs := getRobot(pkID); rs != nil {
		rs.cancel()
	}
	removeRobot(pkID)

	// End all active sessions
	now := time.Now()
	facades.Orm().Query().Exec(
		`UPDATE game_session_levels SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL`, now, pkID)
	facades.Orm().Query().Exec(
		`UPDATE game_session_totals SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL`, now, pkID)

	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("is_playing", false)

	helpers.PkHub.Broadcast(pkID, "pk_force_end", map[string]any{
		"pk_id": pkID,
	})

	return nil
}

// PausePkRobot pauses the robot goroutine.
func PausePkRobot(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).Where("user_id", userID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}

	rs := getRobot(pkID)
	if rs == nil {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if !rs.paused {
		rs.paused = true
		rs.pauseCh = make(chan struct{})
	}
	return nil
}

// ResumePkRobot resumes the robot goroutine.
func ResumePkRobot(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).Where("user_id", userID).First(&pk); err != nil || pk.ID == "" {
		return ErrPkNotFound
	}

	rs := getRobot(pkID)
	if rs == nil {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.paused {
		rs.paused = false
		close(rs.pauseCh)
	}
	return nil
}

// OnPkDisconnect is called when the human's SSE connection drops.
func OnPkDisconnect(pkID string) {
	if rs := getRobot(pkID); rs != nil {
		rs.cancel()
	}
	removeRobot(pkID)

	now := time.Now()
	facades.Orm().Query().Exec(
		`UPDATE game_session_levels SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL`, now, pkID)
	facades.Orm().Query().Exec(
		`UPDATE game_session_totals SET ended_at = ? WHERE game_pk_id = ? AND ended_at IS NULL`, now, pkID)

	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("is_playing", false)
}

// --- Robot Goroutine ---

func spawnRobotForLevel(pkID, gameID, gameLevelID, degree string) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	rs := &robotState{cancel: cancel, pauseCh: make(chan struct{})}
	// Close the initial pauseCh so the robot doesn't start paused
	close(rs.pauseCh)
	setRobot(pkID, rs)

	difficulty, ok := consts.PkDifficulties[pk.RobotDifficulty]
	if !ok {
		difficulty = consts.PkDifficulties[consts.PkDifficultyNormal]
	}

	// Create robot's session total (if not exists for this PK)
	var robotSession models.GameSessionTotal
	facades.Orm().Query().
		Where("game_pk_id", pkID).
		Where("user_id", pk.OpponentID).
		Where("ended_at IS NULL").
		First(&robotSession)

	now := time.Now()
	if robotSession.ID == "" {
		var totalLevels int64
		facades.Orm().Query().Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count(&totalLevels)

		robotSession = models.GameSessionTotal{
			ID:               uuid.Must(uuid.NewV7()).String(),
			UserID:           pk.OpponentID,
			GameID:           gameID,
			Degree:           pk.Degree,
			Pattern:          pk.Pattern,
			CurrentLevelID:   &gameLevelID,
			TotalLevelsCount: int(totalLevels),
			StartedAt:        now,
			LastPlayedAt:     now,
			GamePkID:         &pkID,
		}
		facades.Orm().Query().Create(&robotSession)
		upsertGameStats(pk.OpponentID, gameID, now)
	}

	// Create robot's level session
	totalItemsCount := countContentItems(gameID, gameLevelID, degree)
	robotLevelSession := models.GameSessionLevel{
		ID:                 uuid.Must(uuid.NewV7()).String(),
		GameSessionTotalID: robotSession.ID,
		GameLevelID:        gameLevelID,
		Degree:             degree,
		Pattern:            pk.Pattern,
		TotalItemsCount:    totalItemsCount,
		StartedAt:          now,
		LastPlayedAt:       now,
		GamePkID:           &pkID,
	}
	facades.Orm().Query().Create(&robotLevelSession)

	// Fetch content items for this level
	var contentItems []models.ContentItem
	query := facades.Orm().Query().Where("game_level_id", gameLevelID).Where("is_active", true)
	degreeFilter := contentDegreeFilter(degree)
	if degreeFilter != "" {
		query = query.WhereIn("content_type", degreeFilter)
	}
	query.Order("\"order\" ASC").Find(&contentItems)

	// Get robot's display name
	var robotUser models.User
	facades.Orm().Query().Where("id", pk.OpponentID).First(&robotUser)
	robotName := ""
	if robotUser.Nickname != nil {
		robotName = *robotUser.Nickname
	}

	// Run simulation
	runRobotSimulation(ctx, rs, pkID, pk.OpponentID, robotName, robotSession.ID,
		robotLevelSession.ID, gameLevelID, contentItems, difficulty)
}

func runRobotSimulation(
	ctx context.Context,
	rs *robotState,
	pkID, robotUserID, robotName, sessionID, levelSessionID, gameLevelID string,
	contentItems []models.ContentItem,
	difficulty consts.PkDifficultyParams,
) {
	// Roll accuracy for this round
	accuracy := difficulty.AccuracyMin + rand.Float64()*(difficulty.AccuracyMax-difficulty.AccuracyMin)

	combo := 0
	score := 0
	maxCombo := 0
	correctCount := 0
	wrongCount := 0

	for i, item := range contentItems {
		// Check pause
		rs.mu.Lock()
		pauseCh := rs.pauseCh
		rs.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-pauseCh:
			// Unpaused or was never paused, continue
		}

		// Random delay
		delayMs := difficulty.MinDelayMs + rand.Intn(difficulty.MaxDelayMs-difficulty.MinDelayMs+1)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delayMs) * time.Millisecond):
		}

		// Decide correct/wrong
		isCorrect := rand.Float64() < accuracy
		// Combo break chance
		if isCorrect && combo >= 3 && rand.Float64() < difficulty.ComboBreakPct {
			isCorrect = false
		}

		baseScore := 0
		comboScore := 0
		if isCorrect {
			combo++
			if combo > maxCombo {
				maxCombo = combo
			}
			correctCount++
			baseScore = consts.CorrectAnswer
			comboScore = helpers.CalcComboBonus(combo)
			score += baseScore + comboScore
		} else {
			combo = 0
			wrongCount++
		}

		// Write game record
		recordID := uuid.Must(uuid.NewV7()).String()
		facades.Orm().Query().Exec(
			`INSERT INTO game_records (id, user_id, game_session_total_id, game_session_level_id, game_level_id, content_item_id, is_correct, source_answer, user_answer, base_score, combo_score, duration, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (game_session_level_id, content_item_id) DO UPDATE SET
			   is_correct = EXCLUDED.is_correct, base_score = EXCLUDED.base_score, combo_score = EXCLUDED.combo_score, updated_at = EXCLUDED.updated_at`,
			recordID, robotUserID, sessionID, levelSessionID, gameLevelID, item.ID,
			isCorrect, item.Content, item.Content, baseScore, comboScore, delayMs/1000,
			time.Now(), time.Now())

		// Update level session stats
		facades.Orm().Query().Model(&models.GameSessionLevel{}).Where("id", levelSessionID).Updates(map[string]any{
			"score":             score,
			"max_combo":         maxCombo,
			"correct_count":     correctCount,
			"wrong_count":       wrongCount,
			"played_items_count": i + 1,
			"last_played_at":    time.Now(),
		})

		// Broadcast action
		action := "score"
		if !isCorrect {
			action = "skip"
		}
		helpers.PkHub.Broadcast(pkID, "pk_player_action", map[string]any{
			"user_id":      robotUserID,
			"user_name":    robotName,
			"action":       action,
			"combo_streak": combo,
		})
	}

	// Robot completed level
	now := time.Now()
	expEarned := 0
	if len(contentItems) > 0 {
		acc := float64(correctCount) / float64(len(contentItems))
		if acc >= consts.ExpAccuracyThreshold {
			expEarned = consts.LevelCompleteExp
		}
	}

	facades.Orm().Query().Model(&models.GameSessionLevel{}).Where("id", levelSessionID).Updates(map[string]any{
		"ended_at":  now,
		"score":     score,
		"exp":       expEarned,
		"max_combo": maxCombo,
	})

	facades.Orm().Query().Model(&models.GameSessionTotal{}).Where("id", sessionID).Updates(map[string]any{
		"played_levels_count": facades.Orm().Query().Raw("played_levels_count + 1"),
		"last_played_at":      now,
	})
	// Increment played_levels_count via raw SQL to avoid race
	facades.Orm().Query().Exec("UPDATE game_session_totals SET played_levels_count = played_levels_count + 1, last_played_at = ? WHERE id = ?", now, sessionID)

	updateLevelStats(robotUserID, "", gameLevelID, score, expEarned, now)
	if expEarned > 0 {
		facades.Orm().Query().Exec("UPDATE users SET exp = exp + ? WHERE id = ?", expEarned, robotUserID)
	}

	helpers.PkHub.Broadcast(pkID, "pk_player_complete", map[string]string{
		"user_id":       robotUserID,
		"user_name":     robotName,
		"game_level_id": gameLevelID,
	})

	// Check if human already completed — determine winner
	result, _ := DeterminePkWinner(pkID, gameLevelID)
	if result != nil {
		helpers.PkHub.Broadcast(pkID, "pk_level_complete", result)
		return
	}

	// Wait for human with timeout
	warningDuration := time.Duration(consts.PkTimeoutDuration-consts.PkTimeoutWarning) * time.Second
	select {
	case <-ctx.Done():
		return
	case <-time.After(warningDuration):
	}

	// Broadcast timeout warning
	helpers.PkHub.Broadcast(pkID, "pk_timeout_warning", map[string]int{
		"countdown": consts.PkTimeoutWarning,
	})

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(consts.PkTimeoutWarning) * time.Second):
	}

	// Timeout — auto-end human's level, robot wins
	facades.Orm().Query().Exec(
		`UPDATE game_session_levels SET ended_at = ? WHERE game_pk_id = ? AND game_level_id = ? AND ended_at IS NULL`,
		time.Now(), pkID, gameLevelID)

	result, _ = DeterminePkWinner(pkID, gameLevelID)
	if result != nil {
		helpers.PkHub.Broadcast(pkID, "pk_level_complete", result)
	}
	helpers.PkHub.Broadcast(pkID, "pk_timeout", map[string]string{"game_level_id": gameLevelID})
}

// --- Helpers ---

func countContentItems(gameID, gameLevelID, degree string) int {
	query := facades.Orm().Query().Model(&models.ContentItem{}).Where("game_level_id", gameLevelID).Where("is_active", true)
	filter := contentDegreeFilter(degree)
	if filter != "" {
		query = query.WhereIn("content_type", filter)
	}
	var count int64
	query.Count(&count)
	return int(count)
}

func contentDegreeFilter(degree string) string {
	switch degree {
	case consts.GameDegreeIntermediate:
		return "'block','phrase','sentence'"
	case consts.GameDegreeAdvanced:
		return "'sentence'"
	default:
		return ""
	}
}

func upsertGameStats(userID, gameID string, now time.Time) {
	facades.Orm().Query().Exec(
		`INSERT INTO game_stats_totals (id, user_id, game_id, total_sessions, first_played_at, last_played_at, created_at, updated_at)
		 VALUES (?, ?, ?, 1, ?, ?, ?, ?)
		 ON CONFLICT (user_id, game_id) DO UPDATE SET
		   total_sessions = game_stats_totals.total_sessions + 1,
		   last_played_at = EXCLUDED.last_played_at,
		   updated_at = EXCLUDED.updated_at`,
		uuid.Must(uuid.NewV7()).String(), userID, gameID, now, now, now, now)
}

func updateLevelStats(userID, gameID, gameLevelID string, score, exp int, now time.Time) {
	facades.Orm().Query().Exec(
		`INSERT INTO game_stats_levels (id, user_id, game_level_id, highest_score, total_scores, total_play_time, first_played_at, last_played_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
		 ON CONFLICT (user_id, game_level_id) DO UPDATE SET
		   highest_score = GREATEST(game_stats_levels.highest_score, EXCLUDED.highest_score),
		   total_scores = game_stats_levels.total_scores + EXCLUDED.total_scores,
		   last_played_at = EXCLUDED.last_played_at,
		   updated_at = EXCLUDED.updated_at`,
		uuid.Must(uuid.NewV7()).String(), userID, gameLevelID, score, score, now, now, now, now)
}
```

**Note:** The `requireVip`, `CalcComboBonus`, `StartLevelResult`, and `CompleteLevelResult` types are already defined in the existing single/group play services. If `CalcComboBonus` doesn't exist as a helper, create it in `dx-api/app/helpers/scoring.go`:

```go
func CalcComboBonus(combo int) int {
	bonus := 0
	normalized := ((combo - 1) % consts.ComboCycleLength) + 1
	for _, t := range consts.ComboThresholds {
		if normalized >= t.Streak {
			bonus = t.Bonus
		}
	}
	return bonus
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build. Fix any import issues.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/game_play_pk_service.go
git commit -m "feat: add PK play service with robot goroutine simulation"
```

---

## Task 8: PK Request Validation

**Files:**
- Create: `dx-api/app/http/requests/api/pk_request.go`

- [ ] **Step 1: Create request structs**

```go
// dx-api/app/http/requests/api/pk_request.go
package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

type PkStartRequest struct {
	GameID     string  `form:"game_id" json:"game_id"`
	Degree     string  `form:"degree" json:"degree"`
	Pattern    *string `form:"pattern" json:"pattern"`
	LevelID    *string `form:"level_id" json:"level_id"`
	Difficulty string  `form:"difficulty" json:"difficulty"`
}

func (r *PkStartRequest) Authorize(ctx http.Context) error { return nil }

func (r *PkStartRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":    "required|uuid",
		"degree":     helpers.InEnum("degree"),
		"pattern":    helpers.InEnum("pattern"),
		"level_id":   "uuid",
		"difficulty": "required|" + helpers.InEnum("pk_difficulty"),
	}
}

func (r *PkStartRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{"degree": "trim", "pattern": "trim", "difficulty": "trim"}
}

func (r *PkStartRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":    "请选择游戏",
		"game_id.uuid":        "无效的游戏ID",
		"degree.in":           "无效的难度级别",
		"pattern.in":          "无效的练习模式",
		"level_id.uuid":       "无效的关卡ID",
		"difficulty.required": "请选择对手难度",
		"difficulty.in":       "无效的对手难度",
	}
}

func (r *PkStartRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

type PkNextLevelRequest struct {
	CurrentLevelID string `form:"current_level_id" json:"current_level_id"`
}

func (r *PkNextLevelRequest) Authorize(ctx http.Context) error { return nil }

func (r *PkNextLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{"current_level_id": "required|uuid"}
}

func (r *PkNextLevelRequest) Filters(ctx http.Context) map[string]string { return nil }

func (r *PkNextLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"current_level_id.required": "缺少当前关卡ID",
		"current_level_id.uuid":    "无效的关卡ID",
	}
}

func (r *PkNextLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/http/requests/api/pk_request.go
git commit -m "feat: add PK request validation structs"
```

---

## Task 9: PK Controller

**Files:**
- Create: `dx-api/app/http/controllers/api/game_play_pk_controller.go`

- [ ] **Step 1: Create PK controller**

```go
// dx-api/app/http/controllers/api/game_play_pk_controller.go
package api

import (
	"errors"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type GamePlayPkController struct{}

func NewGamePlayPkController() *GamePlayPkController {
	return &GamePlayPkController{}
}

func (c *GamePlayPkController) Start(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.PkStartRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartPk(userID, req.GameID, req.Degree, req.Pattern, req.LevelID, req.Difficulty)
	if err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, result)
}

func (c *GamePlayPkController) StartLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.StartLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartPkLevel(userID, sessionID, req.GameLevelID, req.Degree, req.Pattern)
	if err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, result)
}

func (c *GamePlayPkController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")

	var req requests.CompleteLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CompletePkLevel(userID, sessionID, levelID, req.Score, req.MaxCombo, req.TotalItems)
	if err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, result)
}

func (c *GamePlayPkController) RecordAnswer(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RecordAnswerRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.PkRecordAnswer(userID, req); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) RecordSkip(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RecordSkipRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.PkRecordSkip(userID, req); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) SyncPlayTime(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.SyncPlayTimeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.PkSyncPlayTime(userID, sessionID, req.GameLevelID, req.PlayTime); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) Restore(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	data, err := services.PkRestoreSessionData(userID, sessionID)
	if err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, data)
}

func (c *GamePlayPkController) UpdateContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	contentItemID := ctx.Request().Input("content_item_id")

	if err := services.PkUpdateContentItem(userID, sessionID, contentItemID); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) End(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.EndPk(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) NextLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	var req requests.PkNextLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.NextPkLevel(userID, pkID, req.CurrentLevelID); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) Pause(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.PausePkRobot(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) Resume(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.ResumePkRobot(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func (c *GamePlayPkController) Events(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	conn := helpers.PkHub.Register(pkID, userID, w)
	defer func() {
		helpers.PkHub.Unregister(pkID, userID, conn)
		services.OnPkDisconnect(pkID)
	}()

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

func mapPkError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrPkNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePkNotFound, err.Error())
	case errors.Is(err, services.ErrPkNotPlaying):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodePkNotPlaying, err.Error())
	case errors.Is(err, services.ErrNoMockUserAvail):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeNoMockUser, err.Error())
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, err.Error())
	case errors.Is(err, services.ErrNoGameLevels):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, err.Error())
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, err.Error())
	case errors.Is(err, services.ErrSessionNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, err.Error())
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, err.Error())
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, err.Error())
	case errors.Is(err, services.ErrRateLimited):
		return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, err.Error())
	case errors.Is(err, services.ErrLastLevel):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, err.Error())
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
	}
}
```

**Note:** The `PkRecordAnswer`, `PkRecordSkip`, `PkSyncPlayTime`, `PkRestoreSessionData`, and `PkUpdateContentItem` functions need to be added to the PK service. These are thin wrappers that verify the session belongs to the user and has a `game_pk_id`, then delegate to the existing single-play recording logic. Add them to `game_play_pk_service.go` following the same pattern as the group play service's equivalent methods.

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Fix any compilation errors from missing service methods.

- [ ] **Step 3: Add missing PK service methods**

Add these thin wrappers to `dx-api/app/services/api/game_play_pk_service.go` (they follow the same pattern as group play service):

```go
// PkRecordAnswer records an answer for the human player in a PK session.
func PkRecordAnswer(userID string, input requests.RecordAnswerRequest) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return RecordAnswer(userID, input)
}

// PkRecordSkip records a skip for the human player in a PK session.
func PkRecordSkip(userID string, input requests.RecordSkipRequest) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	return RecordSkip(userID, input)
}

// PkSyncPlayTime syncs playtime for PK session.
func PkSyncPlayTime(userID, sessionID, gameLevelID string, playTime int) error {
	return SyncPlayTime(userID, sessionID, gameLevelID, playTime)
}

// PkRestoreSessionData restores PK session data.
func PkRestoreSessionData(userID, sessionID string) (*SessionRestoreData, error) {
	return RestoreSessionData(userID, sessionID, "")
}

// PkUpdateContentItem updates the current content item for a PK session.
func PkUpdateContentItem(userID, sessionID, contentItemID string) error {
	return UpdateCurrentContentItem(userID, sessionID, contentItemID)
}
```

**Note:** These reference existing functions from `game_play_single_service.go`. Check that `RecordAnswer`, `RecordSkip`, `SyncPlayTime`, `RestoreSessionData`, and `UpdateCurrentContentItem` accept the same parameter types. Adjust as needed to match the existing signatures.

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/controllers/api/game_play_pk_controller.go dx-api/app/services/api/game_play_pk_service.go
git commit -m "feat: add PK controller with all endpoints and SSE"
```

---

## Task 10: Backend Routes

**Files:**
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add PK routes**

In `dx-api/routes/api.go`, add after the group play routes block (after line 271):

```go
// PK game play routes
playPkController := apicontrollers.NewGamePlayPkController()
protected.Get("/play-pk/{id}/events", playPkController.Events)
protected.Prefix("/play-pk").Group(func(pk route.Router) {
	pk.Post("/start", playPkController.Start)
	pk.Post("/{id}/levels/start", playPkController.StartLevel)
	pk.Post("/{id}/levels/{levelId}/complete", playPkController.CompleteLevel)
	pk.Post("/{id}/answers", playPkController.RecordAnswer)
	pk.Post("/{id}/skips", playPkController.RecordSkip)
	pk.Post("/{id}/sync-playtime", playPkController.SyncPlayTime)
	pk.Get("/{id}/restore", playPkController.Restore)
	pk.Put("/{id}/content-item", playPkController.UpdateContentItem)
	pk.Post("/{id}/end", playPkController.End)
	pk.Post("/{id}/next-level", playPkController.NextLevel)
	pk.Post("/{id}/pause", playPkController.Pause)
	pk.Post("/{id}/resume", playPkController.Resume)
})
```

**Note:** The SSE events endpoint is registered outside the prefix group to ensure it works with the long-lived connection pattern.

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add dx-api/routes/api.go
git commit -m "feat: add PK play routes"
```

---

## Task 11: Frontend Types

**Files:**
- Create: `dx-web/src/features/web/play-pk/types/pk-play.ts`

- [ ] **Step 1: Create PK type definitions**

```typescript
// dx-web/src/features/web/play-pk/types/pk-play.ts

export type PkWinner = {
  user_id: string;
  user_name: string;
  score: number;
};

export type PkLevelCompleteEvent = {
  game_level_id: string;
  winner: PkWinner;
  participants: PkWinner[];
};

export type PkForceEndEvent = {
  pk_id: string;
};

export type PkNextLevelEvent = {
  pk_id: string;
  game_id: string;
  level_id: string;
  level_name: string;
  degree: string;
  pattern: string | null;
};

export type PkPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
};

export type PkPlayerActionEvent = {
  user_id: string;
  user_name: string;
  action: "score" | "skip" | "combo";
  combo_streak?: number;
};

export type PkTimeoutWarningEvent = {
  countdown: number;
};

export type PkTimeoutEvent = {
  game_level_id: string;
};
```

- [ ] **Step 2: Commit**

```bash
git add dx-web/src/features/web/play-pk/types/pk-play.ts
git commit -m "feat: add PK play TypeScript types"
```

---

## Task 12: Frontend Actions

**Files:**
- Create: `dx-web/src/features/web/play-pk/actions/session.action.ts`

- [ ] **Step 1: Create PK session actions**

Follow the exact pattern from `dx-web/src/features/web/play-group/actions/session.action.ts`, replacing `/api/play-group/` with `/api/play-pk/` and adding the `difficulty` parameter to `startSessionAction`. Also add `pauseAction`, `resumeAction`, `endPkAction`, and `nextLevelAction`.

```typescript
// dx-web/src/features/web/play-pk/actions/session.action.ts

import { apiClient } from "@/lib/api-client";

export async function startPkAction(
  gameId: string,
  degree: string,
  pattern: string | null,
  levelId: string | null,
  difficulty: string
) {
  try {
    const res = await apiClient.post<{
      pk_id: string;
      session_id: string;
      opponent_id: string;
      opponent_name: string;
    }>("/api/play-pk/start", {
      game_id: gameId,
      degree,
      pattern,
      level_id: levelId,
      difficulty,
    });
    if (res.code !== 0) return { data: null, error: res.message || "无法开始PK" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "无法开始PK" };
  }
}

export async function startLevelAction(
  sessionId: string,
  gameLevelId: string,
  degree: string,
  pattern: string | null
) {
  try {
    const res = await apiClient.post<{ id: string; currentContentItemId?: string | null }>(
      `/api/play-pk/${sessionId}/levels/start`,
      { game_level_id: gameLevelId, degree, pattern }
    );
    if (res.code !== 0) return { data: null, error: res.message || "开始关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "开始关卡失败" };
  }
}

export async function completeLevelAction(
  sessionId: string,
  gameLevelId: string,
  data: { score: number; maxCombo: number; totalItems: number }
) {
  try {
    const res = await apiClient.post<unknown>(
      `/api/play-pk/${sessionId}/levels/${gameLevelId}/complete`,
      { score: data.score, max_combo: data.maxCombo, total_items: data.totalItems }
    );
    if (res.code !== 0) return { data: null, error: res.message || "完成关卡失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "完成关卡失败" };
  }
}

export async function recordAnswerAction(data: {
  gameSessionTotalId: string;
  gameSessionLevelId: string;
  gameLevelId: string;
  contentItemId: string;
  isCorrect: boolean;
  userAnswer: string;
  sourceAnswer: string;
  baseScore: number;
  comboScore: number;
  score: number;
  maxCombo: number;
  playTime: number;
  nextContentItemId: string | null;
  duration: number;
}) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${data.gameSessionTotalId}/answers`, {
      game_session_level_id: data.gameSessionLevelId,
      game_level_id: data.gameLevelId,
      content_item_id: data.contentItemId,
      is_correct: data.isCorrect,
      user_answer: data.userAnswer,
      source_answer: data.sourceAnswer,
      base_score: data.baseScore,
      combo_score: data.comboScore,
      score: data.score,
      max_combo: data.maxCombo,
      play_time: data.playTime,
      next_content_item_id: data.nextContentItemId,
      duration: data.duration,
    });
    return { data: null, error: null };
  } catch {
    return { data: null, error: "记录失败" };
  }
}

export async function recordSkipAction(data: {
  gameSessionTotalId: string;
  gameLevelId: string;
  playTime: number;
  nextContentItemId: string | null;
}) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${data.gameSessionTotalId}/skips`, {
      game_level_id: data.gameLevelId,
      play_time: data.playTime,
      next_content_item_id: data.nextContentItemId,
    });
    return { data: null, error: null };
  } catch {
    return { data: null, error: "跳过失败" };
  }
}

export async function restoreSessionDataAction(sessionId: string) {
  try {
    const res = await apiClient.get<{
      sessionLevel?: {
        score: number;
        maxCombo: number;
        correctCount: number;
        wrongCount: number;
        skipCount: number;
        playTime: number;
      };
    }>(`/api/play-pk/${sessionId}/restore`);
    if (res.code !== 0) return { data: null, error: res.message || "恢复会话数据失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "恢复会话数据失败" };
  }
}

export async function endPkAction(pkId: string) {
  try {
    const res = await apiClient.post<unknown>(`/api/play-pk/${pkId}/end`);
    if (res.code !== 0) return { data: null, error: res.message || "结束PK失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "结束PK失败" };
  }
}

export async function nextLevelAction(pkId: string, currentLevelId: string) {
  try {
    const res = await apiClient.post<unknown>(`/api/play-pk/${pkId}/next-level`, {
      current_level_id: currentLevelId,
    });
    if (res.code !== 0) return { data: null, error: res.message || "下一关失败" };
    return { data: res.data, error: null };
  } catch {
    return { data: null, error: "下一关失败" };
  }
}

export async function pausePkAction(pkId: string) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${pkId}/pause`);
  } catch {
    // Fire-and-forget
  }
}

export async function resumePkAction(pkId: string) {
  try {
    await apiClient.post<unknown>(`/api/play-pk/${pkId}/resume`);
  } catch {
    // Fire-and-forget
  }
}

export async function fetchLevelContentAction(
  gameId: string,
  levelId: string,
  degree?: string
) {
  try {
    const params = degree ? `?degree=${degree}` : "";
    const res = await apiClient.get<Record<string, unknown>[]>(
      `/api/games/${gameId}/levels/${levelId}/content${params}`
    );
    if (res.code !== 0) return { data: null, error: res.message || "加载内容失败" };
    return { data: res.data ?? [], error: null };
  } catch {
    return { data: null, error: "加载内容失败" };
  }
}

export async function markAsReviewAction(data: {
  contentItemId: string;
  gameId: string;
  gameLevelId: string;
}) {
  try {
    await apiClient.post<unknown>("/api/tracking/review", {
      content_item_id: data.contentItemId,
      game_id: data.gameId,
      game_level_id: data.gameLevelId,
    });
  } catch {
    // Fire-and-forget
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-web/src/features/web/play-pk/actions/session.action.ts
git commit -m "feat: add PK session actions"
```

---

## Task 13: Frontend SSE Hooks

**Files:**
- Create: `dx-web/src/hooks/use-pk-sse.ts`
- Create: `dx-web/src/features/web/play-pk/hooks/use-pk-play-events.ts`

- [ ] **Step 1: Create PK SSE hook**

```typescript
// dx-web/src/hooks/use-pk-sse.ts
"use client";

import { useEffect, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export function usePkSSE(
  pkId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  listenersRef.current = listeners;

  useEffect(() => {
    if (!pkId) return;

    const url = `${API_URL}/api/play-pk/${pkId}/events`;
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
  }, [pkId]);
}
```

- [ ] **Step 2: Create PK play events hook**

```typescript
// dx-web/src/features/web/play-pk/hooks/use-pk-play-events.ts
"use client";

import { useRef, useMemo, useEffect } from "react";
import { usePkSSE } from "@/hooks/use-pk-sse";
import type {
  PkLevelCompleteEvent,
  PkForceEndEvent,
  PkNextLevelEvent,
  PkPlayerCompleteEvent,
  PkPlayerActionEvent,
  PkTimeoutWarningEvent,
  PkTimeoutEvent,
} from "../types/pk-play";

type PkPlayEventHandlers = {
  onLevelComplete?: (event: PkLevelCompleteEvent) => void;
  onForceEnd?: (event: PkForceEndEvent) => void;
  onNextLevel?: (event: PkNextLevelEvent) => void;
  onPlayerComplete?: (event: PkPlayerCompleteEvent) => void;
  onPlayerAction?: (event: PkPlayerActionEvent) => void;
  onTimeoutWarning?: (event: PkTimeoutWarningEvent) => void;
  onTimeout?: (event: PkTimeoutEvent) => void;
};

export function usePkPlayEvents(
  pkId: string | null,
  handlers: PkPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  const listeners = useMemo(() => ({
    pk_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as PkLevelCompleteEvent),
    pk_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as PkForceEndEvent),
    pk_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as PkNextLevelEvent),
    pk_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as PkPlayerCompleteEvent),
    pk_player_action: (data: unknown) =>
      handlersRef.current.onPlayerAction?.(data as PkPlayerActionEvent),
    pk_timeout_warning: (data: unknown) =>
      handlersRef.current.onTimeoutWarning?.(data as PkTimeoutWarningEvent),
    pk_timeout: (data: unknown) =>
      handlersRef.current.onTimeout?.(data as PkTimeoutEvent),
  }), []);

  usePkSSE(pkId, listeners);
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/hooks/use-pk-sse.ts dx-web/src/features/web/play-pk/hooks/use-pk-play-events.ts
git commit -m "feat: add PK SSE hooks"
```

---

## Task 14: Frontend Zustand Store

**Files:**
- Create: `dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts`

- [ ] **Step 1: Create PK play store**

Follow the `useGroupPlayStore` pattern but adapted for PK (no team mode, add timeout state, opponent info).

```typescript
// dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts
"use client";

import { create } from "zustand";
import type { ContentItem } from "@/features/web/play-core/types/content";
import type { PkLevelCompleteEvent, PkPlayerActionEvent } from "../types/pk-play";

type PkPhase = "playing" | "waiting" | "result" | null;

interface PkPlayState {
  // Session
  pkId: string | null;
  sessionId: string | null;
  levelSessionId: string | null;
  gameId: string | null;
  gameMode: string | null;
  degree: string | null;
  pattern: string | null;
  levelId: string | null;
  difficulty: string | null;

  // Content
  contentItems: ContentItem[] | null;
  currentIndex: number;
  startFromIndex: number;

  // Scoring
  score: number;
  combo: { current: number; maxCombo: number };
  correctCount: number;
  wrongCount: number;
  skipCount: number;
  playTime: number;

  // PK-specific
  opponentId: string | null;
  opponentName: string | null;
  pkPhase: PkPhase;
  pkResult: PkLevelCompleteEvent | null;
  opponentCompleted: boolean;
  lastOpponentAction: PkPlayerActionEvent | null;
  timeoutCountdown: number | null;

  // Actions
  initSession: (data: {
    pkId: string;
    sessionId: string;
    levelSessionId: string;
    gameId: string;
    gameMode: string;
    degree: string;
    pattern: string | null;
    levelId: string;
    difficulty: string;
    opponentId: string;
    opponentName: string;
    contentItems: ContentItem[];
    startFromIndex?: number;
    restore?: {
      score: number;
      maxCombo: number;
      correctCount: number;
      wrongCount: number;
      skipCount: number;
      playTime: number;
    };
  }) => void;
  setPkWaiting: () => void;
  setPkResult: (result: PkLevelCompleteEvent) => void;
  setOpponentCompleted: () => void;
  setLastOpponentAction: (action: PkPlayerActionEvent) => void;
  setTimeoutCountdown: (seconds: number | null) => void;
  exitGame: () => void;
}

const initialState = {
  pkId: null,
  sessionId: null,
  levelSessionId: null,
  gameId: null,
  gameMode: null,
  degree: null,
  pattern: null,
  levelId: null,
  difficulty: null,
  contentItems: null,
  currentIndex: 0,
  startFromIndex: 0,
  score: 0,
  combo: { current: 0, maxCombo: 0 },
  correctCount: 0,
  wrongCount: 0,
  skipCount: 0,
  playTime: 0,
  opponentId: null,
  opponentName: null,
  pkPhase: null as PkPhase,
  pkResult: null,
  opponentCompleted: false,
  lastOpponentAction: null,
  timeoutCountdown: null,
};

export const usePkPlayStore = create<PkPlayState>((set) => ({
  ...initialState,

  initSession: (data) =>
    set({
      pkId: data.pkId,
      sessionId: data.sessionId,
      levelSessionId: data.levelSessionId,
      gameId: data.gameId,
      gameMode: data.gameMode,
      degree: data.degree,
      pattern: data.pattern,
      levelId: data.levelId,
      difficulty: data.difficulty,
      opponentId: data.opponentId,
      opponentName: data.opponentName,
      contentItems: data.contentItems,
      startFromIndex: data.startFromIndex ?? 0,
      currentIndex: data.startFromIndex ?? 0,
      pkPhase: "playing",
      pkResult: null,
      opponentCompleted: false,
      lastOpponentAction: null,
      timeoutCountdown: null,
      ...(data.restore
        ? {
            score: data.restore.score,
            combo: { current: 0, maxCombo: data.restore.maxCombo },
            correctCount: data.restore.correctCount,
            wrongCount: data.restore.wrongCount,
            skipCount: data.restore.skipCount,
            playTime: data.restore.playTime,
          }
        : {
            score: 0,
            combo: { current: 0, maxCombo: 0 },
            correctCount: 0,
            wrongCount: 0,
            skipCount: 0,
            playTime: 0,
          }),
    }),

  setPkWaiting: () => set({ pkPhase: "waiting" }),

  setPkResult: (result) => set({ pkPhase: "result", pkResult: result, timeoutCountdown: null }),

  setOpponentCompleted: () => set({ opponentCompleted: true }),

  setLastOpponentAction: (action) => set({ lastOpponentAction: action }),

  setTimeoutCountdown: (seconds) => set({ timeoutCountdown: seconds }),

  exitGame: () => set(initialState),
}));
```

- [ ] **Step 2: Commit**

```bash
git add dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts
git commit -m "feat: add PK play Zustand store"
```

---

## Task 15: Modify HeroCard — Add PK Button

**Files:**
- Modify: `dx-web/src/features/web/games/components/hero-card.tsx`

- [ ] **Step 1: Add PK button and onPkStart prop**

Add `Swords` icon import and `onPkStart` prop. Insert PK button between the start button and 群组 link.

In the props interface, add:

```typescript
onPkStart?: () => void;
```

In the imports, add `Swords` from lucide-react.

In the action buttons div, add PK button after the start game button and before the 群组 Link:

```tsx
<button
  type="button"
  onClick={onPkStart}
  className="flex items-center gap-2 rounded-[10px] border border-border bg-card px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent"
>
  <Swords className="h-4 w-4" />
  PK
</button>
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/games/components/hero-card.tsx
git commit -m "feat: add PK button to game detail hero card"
```

---

## Task 16: Modify GameModeCard — Add Difficulty Selector

**Files:**
- Modify: `dx-web/src/features/web/play-core/components/game-mode-card.tsx`

- [ ] **Step 1: Add PK mode support**

Add a `mode` prop (`"single" | "pk"`, default `"single"`) to `GameModeCardProps`. When `mode === "pk"`:
- Show a difficulty selector (Easy / Normal / Hard) below the degree options
- Change the start button to navigate to `/hall/play-pk/{gameId}` with `difficulty` param
- Hide resume/restart buttons (PK always starts fresh)

Add difficulty options array:

```typescript
const difficultyOptions: { value: string; title: string; desc: string; icon: LucideIcon; iconColor: string; iconBg: string }[] = [
  { value: "easy", title: "简单", desc: "对手较弱，适合新手", icon: Zap, iconColor: "text-emerald-500", iconBg: "bg-emerald-500/[0.08]" },
  { value: "normal", title: "普通", desc: "旗鼓相当，适度挑战", icon: Flame, iconColor: "text-amber-500", iconBg: "bg-amber-500/[0.08]" },
  { value: "hard", title: "困难", desc: "强力对手，极限挑战", icon: Trophy, iconColor: "text-red-500", iconBg: "bg-red-500/[0.08]" },
];
```

Add state: `const [selectedDifficulty, setSelectedDifficulty] = useState("normal");`

Add PK start handler:

```typescript
function handlePkStart() {
  startTransition(() => {
    const params = new URLSearchParams({ degree: selectedDegree, difficulty: selectedDifficulty });
    if (isWordSentence) params.set("pattern", selectedPattern);
    if (levelId) params.set("level", levelId);
    router.push(`/hall/play-pk/${gameId}?${params}`);
  });
}
```

When `mode === "pk"`, render the difficulty selector section after the degree options and use `handlePkStart` for the start button.

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-mode-card.tsx
git commit -m "feat: add difficulty selector to game mode card for PK"
```

---

## Task 17: PK Shell Components

**Files:**
- Create: `dx-web/src/features/web/play-pk/components/pk-play-shell.tsx`
- Create: `dx-web/src/features/web/play-pk/components/pk-play-top-bar.tsx`
- Create: `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx`
- Create: `dx-web/src/features/web/play-pk/components/pk-play-waiting-screen.tsx`
- Create: `dx-web/src/features/web/play-pk/components/pk-play-result-panel.tsx`

These components follow the exact patterns from the group play equivalents. The key differences are:

1. **PkPlayShell** — Uses `usePkPlayStore` instead of `useGroupPlayStore`, listens via `usePkPlayEvents`, adds timeout countdown modal, pause sends POST to `/api/play-pk/{pkId}/pause`
2. **PkPlayTopBar** — Same as `GroupPlayTopBar` but without the level time limit countdown and member roster (only 2 players). Adds pause/resume toggle.
3. **PkPlayLoadingScreen** — Calls `startPkAction` instead of `startSessionAction`, shows opponent name/avatar
4. **PkPlayWaitingScreen** — Shows "等待对手完成..." with opponent info
5. **PkPlayResultPanel** — Simplified podium with exactly 2 players (no 3rd place), shows next level / end buttons

- [ ] **Step 1: Create PkPlayShell**

Mirror `group-play-shell.tsx` structure. Replace group store/events/actions with PK equivalents. Add timeout countdown modal that shows when `timeoutCountdown` is not null. On pause, call `pausePkAction(pkId)`. On resume, call `resumePkAction(pkId)`.

- [ ] **Step 2: Create PkPlayTopBar**

Mirror `group-play-top-bar.tsx`. Remove group member roster. Add pause toggle button. Show opponent's live score/combo from `lastOpponentAction`.

- [ ] **Step 3: Create PkPlayLoadingScreen**

Mirror `group-play-loading-screen.tsx`. Call `startPkAction()` instead of `startSessionAction()`. Initialize `pkPlayStore` with opponent info from the start response.

- [ ] **Step 4: Create PkPlayWaitingScreen**

Simplified waiting screen showing "等待对手完成..." with opponent avatar and name. Show opponent's last action.

- [ ] **Step 5: Create PkPlayResultPanel**

Simplified 2-player podium. Left: 2nd place, Right: 1st place (with crown). Show scores. "下一关" button calls `nextLevelAction(pkId, currentLevelId)`. "结束" button calls `endPkAction(pkId)` and navigates to game details.

- [ ] **Step 6: Verify lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/play-pk/components/
git commit -m "feat: add PK play shell components"
```

---

## Task 18: PK Route Page

**Files:**
- Create: `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx`

- [ ] **Step 1: Create route page**

Mirror the `play-group/[id]/page.tsx` pattern. Read `degree`, `pattern`, `difficulty` from search params. Fetch game data and user profile. Render `PkPlayShell`.

```typescript
// dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx
"use client";

import { useEffect, useState } from "react";
import { use } from "react";
import { notFound } from "next/navigation";
import { useSearchParams } from "next/navigation";
import { apiClient } from "@/lib/api-client";
import { PkPlayShell } from "@/features/web/play-pk/components/pk-play-shell";

export default function PkPlayPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const searchParams = useSearchParams();

  const degree = searchParams.get("degree");
  const pattern = searchParams.get("pattern");
  const level = searchParams.get("level");
  const difficulty = searchParams.get("difficulty") ?? "normal";

  type GameData = {
    id: string;
    name: string;
    mode: string;
    levels: { id: string; name: string; order: number }[];
  };

  type ApiGameData = {
    id: string;
    name: string;
    mode: string;
    levels?: { id: string; name: string; order: number }[];
  };

  type ApiProfileData = {
    id?: string;
    nickname?: string | null;
    username?: string;
    avatarUrl?: string | null;
  };

  const [game, setGame] = useState<GameData | null>(null);
  const [player, setPlayer] = useState<{
    id: string;
    nickname: string;
    avatarUrl: string | null;
  }>({ id: "", nickname: "我", avatarUrl: null });
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    async function load() {
      const [gameRes, profileRes] = await Promise.all([
        apiClient.get<ApiGameData>(`/api/games/${id}`),
        apiClient.get<ApiProfileData>("/api/user/profile"),
      ]);

      if (gameRes.code !== 0 || !gameRes.data) {
        setLoaded(true);
        return;
      }

      const g = gameRes.data;
      setGame({
        id: g.id,
        name: g.name,
        mode: g.mode,
        levels: (g.levels ?? []).map((l) => ({
          id: l.id,
          name: l.name,
          order: l.order,
        })),
      });

      if (profileRes.code === 0 && profileRes.data) {
        setPlayer({
          id: profileRes.data.id ?? "",
          nickname: profileRes.data.nickname || profileRes.data.username || "我",
          avatarUrl: profileRes.data.avatarUrl ?? null,
        });
      }

      setLoaded(true);
    }

    load();
  }, [id]);

  if (!loaded) return null;
  if (!game || !degree) {
    notFound();
    return null;
  }

  const targetLevelId = level ?? game.levels[0]?.id ?? "";

  return (
    <PkPlayShell
      game={game}
      player={player}
      degree={degree}
      pattern={pattern}
      levelId={targetLevelId}
      difficulty={difficulty}
    />
  );
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/play-pk/
git commit -m "feat: add PK play route page"
```

---

## Task 19: Database Migration

**Files:**
- Run SQL directly against PostgreSQL

- [ ] **Step 1: Create game_pks table and add columns**

```sql
-- Create game_pks table
CREATE TABLE game_pks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    opponent_id UUID NOT NULL,
    game_id UUID NOT NULL,
    degree VARCHAR(50) NOT NULL,
    pattern VARCHAR(50),
    robot_difficulty VARCHAR(50) NOT NULL,
    current_level_id UUID,
    is_playing BOOLEAN NOT NULL DEFAULT false,
    last_winner_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_pks_user_id ON game_pks (user_id);
CREATE INDEX idx_game_pks_opponent_id ON game_pks (opponent_id);
CREATE INDEX idx_game_pks_game_id ON game_pks (game_id);
CREATE INDEX idx_game_pks_is_playing ON game_pks (is_playing);

-- Add game_pk_id to session tables
ALTER TABLE game_session_totals ADD COLUMN game_pk_id UUID;
CREATE INDEX idx_game_session_totals_game_pk_id ON game_session_totals (game_pk_id);

ALTER TABLE game_session_levels ADD COLUMN game_pk_id UUID;
CREATE INDEX idx_game_session_levels_game_pk_id ON game_session_levels (game_pk_id);
```

- [ ] **Step 2: Verify the migration**

Run: `psql -d douxue -c "\d game_pks"`
Expected: Table structure matches the schema.

Run: `psql -d douxue -c "\d game_session_totals" | grep game_pk_id`
Expected: Column exists.

- [ ] **Step 3: Commit migration notes**

No migration file needed if using direct SQL. Document in commit message.

```bash
git commit --allow-empty -m "chore: document game_pks table and game_pk_id column migration"
```

---

## Task 20: Integration Verification

- [ ] **Step 1: Backend build check**

Run: `cd dx-api && go build ./...`
Expected: Clean build.

Run: `cd dx-api && go vet ./...`
Expected: No issues.

- [ ] **Step 2: Frontend lint check**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 3: Frontend build check**

Run: `cd dx-web && npm run build`
Expected: Clean build.

- [ ] **Step 4: Manual smoke test**

1. Start backend: `cd dx-api && go run .`
2. Start frontend: `cd dx-web && npm run dev`
3. Navigate to a game details page
4. Verify PK button appears between "开始游戏" and "群组"
5. Click PK → verify GameModeCard shows difficulty selector
6. Select degree + difficulty → verify navigation to `/hall/play-pk/{id}`
7. Verify loading screen appears, session created
8. Verify game plays normally
9. Verify robot's score/combo updates appear via SSE
10. Verify result panel shows 2-player podium with winner
11. Verify "下一关" and "结束" buttons work
12. Verify pause/resume works (robot stops/continues)
