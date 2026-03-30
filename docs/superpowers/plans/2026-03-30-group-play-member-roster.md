# Group Play Member Roster Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Display group members with live completion indicators below the progress bar during group gameplay.

**Architecture:** Extend the `group_game_start` SSE event with a participants roster. Add a new `group_player_complete` SSE event for per-player completion. Store participants and completion state in the existing Zustand store. Render avatars in the top bar's player panel.

**Tech Stack:** Go/Goravel (backend SSE), React/Next.js + Zustand (frontend state), ShadCN Avatar component, Tailwind CSS

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `dx-api/app/services/api/group_game_service.go` | Modify | Extend `GroupGameStartEvent` struct + build participants in `StartGroupGame` |
| `dx-api/app/services/api/game_play_group_service.go` | Modify | Broadcast `group_player_complete` in `GroupPlayCompleteLevel` |
| `dx-web/src/features/web/groups/types/group.ts` | Modify | Add `participants` to `GroupGameStartEvent` type |
| `dx-web/src/features/web/groups/components/group-game-room.tsx` | Modify | Store participants in sessionStorage on game start |
| `dx-web/src/features/web/play-group/types/group-play.ts` | Modify | Add participant + player complete types |
| `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts` | Modify | Add participants + completedPlayerIds state/actions |
| `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts` | Modify | Add `group_player_complete` SSE listener |
| `dx-web/src/features/web/play-group/components/group-play-loading-screen.tsx` | Modify | Read participants from sessionStorage |
| `dx-web/src/features/web/play-group/components/group-play-top-bar.tsx` | Modify | Render member roster UI |
| `dx-web/src/features/web/play-group/components/group-play-shell.tsx` | Modify | Pass `playerId` to top bar |
| `docs/game-lsrw-group-rule.md` | Modify | Document new SSE event and roster |

---

### Task 1: Extend `GroupGameStartEvent` with participants (Backend)

**Files:**
- Modify: `dx-api/app/services/api/group_game_service.go:139-150` (struct), `dx-api/app/services/api/group_game_service.go:235-246` (broadcast)

- [ ] **Step 1: Add participant types and extend the event struct**

In `dx-api/app/services/api/group_game_service.go`, add these types before `GroupGameStartEvent` and extend the struct:

```go
// ParticipantMember represents a connected member in the game.
type ParticipantMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// SoloParticipants is the participants payload for group_solo mode.
type SoloParticipants struct {
	Mode    string              `json:"mode"`
	Members []ParticipantMember `json:"members"`
}

// TeamParticipantGroup is one team in the participants payload for group_team mode.
type TeamParticipantGroup struct {
	SubgroupID   string              `json:"subgroup_id"`
	SubgroupName string              `json:"subgroup_name"`
	Members      []ParticipantMember `json:"members"`
}

// TeamParticipants is the participants payload for group_team mode.
type TeamParticipants struct {
	Mode  string                 `json:"mode"`
	Teams []TeamParticipantGroup `json:"teams"`
}

// GroupGameStartEvent is the SSE payload for group_game_start.
type GroupGameStartEvent struct {
	GameGroupID    string  `json:"game_group_id"`
	GameID         string  `json:"game_id"`
	GameName       string  `json:"game_name"`
	GameMode       string  `json:"game_mode"`
	Degree         string  `json:"degree"`
	Pattern        *string `json:"pattern"`
	LevelTimeLimit int     `json:"level_time_limit"`
	LevelID        *string `json:"level_id"`
	LevelName      string  `json:"level_name"`
	Participants   any     `json:"participants"`
}
```

- [ ] **Step 2: Build participants roster in `StartGroupGame`**

In `StartGroupGame`, after `Set is_playing = true` (line 233) and before the `Broadcast` call (line 236), add the participant-building logic:

```go
	// Build participants roster from connected SSE users
	connectedIDs := helpers.GroupSSEHub.ConnectedUserIDs(groupID)

	// Batch-load user names for connected users
	type userInfo struct {
		ID       string  `gorm:"column:id"`
		Username string  `gorm:"column:username"`
		Nickname *string `gorm:"column:nickname"`
	}
	var users []userInfo
	if len(connectedIDs) > 0 {
		facades.Orm().Query().Raw(
			"SELECT id, username, nickname FROM users WHERE id IN ?", connectedIDs,
		).Scan(&users)
	}
	userMap := make(map[string]string, len(users))
	for _, u := range users {
		name := u.Username
		if u.Nickname != nil && *u.Nickname != "" {
			name = *u.Nickname
		}
		userMap[u.ID] = name
	}

	var participants any
	if *group.GameMode == consts.GameModeTeam {
		// Build team participants: subgroups with their connected members
		type subgroupRow struct {
			ID   string `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		var subgroups []subgroupRow
		facades.Orm().Query().Raw(
			`SELECT id, name FROM game_subgroups WHERE game_group_id = ? ORDER BY "order" ASC`, groupID,
		).Scan(&subgroups)

		teams := make([]TeamParticipantGroup, 0, len(subgroups))
		for _, sg := range subgroups {
			type memberRow struct {
				UserID string `gorm:"column:user_id"`
			}
			var members []memberRow
			facades.Orm().Query().Raw(
				"SELECT user_id FROM game_subgroup_members WHERE game_subgroup_id = ?", sg.ID,
			).Scan(&members)

			teamMembers := make([]ParticipantMember, 0)
			for _, m := range members {
				if name, ok := userMap[m.UserID]; ok {
					teamMembers = append(teamMembers, ParticipantMember{UserID: m.UserID, UserName: name})
				}
			}
			if len(teamMembers) > 0 {
				teams = append(teams, TeamParticipantGroup{
					SubgroupID:   sg.ID,
					SubgroupName: sg.Name,
					Members:      teamMembers,
				})
			}
		}
		participants = TeamParticipants{Mode: consts.GameModeTeam, Teams: teams}
	} else {
		// Build solo participants: flat list of connected members
		members := make([]ParticipantMember, 0, len(connectedIDs))
		for _, uid := range connectedIDs {
			if name, ok := userMap[uid]; ok {
				members = append(members, ParticipantMember{UserID: uid, UserName: name})
			}
		}
		participants = SoloParticipants{Mode: consts.GameModeSolo, Members: members}
	}
```

Then update the broadcast call to include `Participants`:

```go
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_start", GroupGameStartEvent{
		GameGroupID:    groupID,
		GameID:         *group.CurrentGameID,
		GameName:       game.Name,
		GameMode:       *group.GameMode,
		Degree:         degree,
		Pattern:        pattern,
		LevelTimeLimit: group.LevelTimeLimit,
		LevelID:        levelID,
		LevelName:      startLevel.Name,
		Participants:   participants,
	})
```

- [ ] **Step 3: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/group_game_service.go
git commit -m "feat: include participants roster in group_game_start SSE event"
```

---

### Task 2: Add `group_player_complete` SSE event (Backend)

**Files:**
- Modify: `dx-api/app/services/api/game_play_group_service.go:362-383`

- [ ] **Step 1: Add the player complete event type**

In `dx-api/app/services/api/game_play_group_service.go`, add this type near the top (after the existing result types around line 45):

```go
// GroupPlayerCompleteEvent is the SSE payload for group_player_complete.
type GroupPlayerCompleteEvent struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	GameLevelID string `json:"game_level_id"`
}
```

- [ ] **Step 2: Broadcast `group_player_complete` after transaction commit**

In `GroupPlayCompleteLevel`, after the transaction commit (line 363) and before the group winner check (line 366), add:

```go
	// Broadcast individual player completion to group
	if session.GameGroupID != nil {
		// Resolve user name for broadcast
		var user models.User
		if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user); err == nil && user.ID != "" {
			userName := user.Username
			if user.Nickname != nil && *user.Nickname != "" {
				userName = *user.Nickname
			}
			helpers.GroupSSEHub.Broadcast(*session.GameGroupID, "group_player_complete", GroupPlayerCompleteEvent{
				UserID:      userID,
				UserName:    userName,
				GameLevelID: gameLevelID,
			})
		}
	}
```

The full block from commit to the existing winner check should now read:

```go
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Broadcast individual player completion to group
	if session.GameGroupID != nil {
		var user models.User
		if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&user); err == nil && user.ID != "" {
			userName := user.Username
			if user.Nickname != nil && *user.Nickname != "" {
				userName = *user.Nickname
			}
			helpers.GroupSSEHub.Broadcast(*session.GameGroupID, "group_player_complete", GroupPlayerCompleteEvent{
				UserID:      userID,
				UserName:    userName,
				GameLevelID: gameLevelID,
			})
		}
	}

	// 6. Check for group winner determination and broadcast SSE
	if session.GameGroupID != nil {
		result, winErr := CheckAndDetermineWinner(*session.GameGroupID, gameLevelID)
		// ... existing code unchanged ...
	}
```

- [ ] **Step 3: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/game_play_group_service.go
git commit -m "feat: broadcast group_player_complete SSE event on individual level completion"
```

---

### Task 3: Update frontend types

**Files:**
- Modify: `dx-web/src/features/web/groups/types/group.ts:36-46`
- Modify: `dx-web/src/features/web/play-group/types/group-play.ts`

- [ ] **Step 1: Add `participants` to `GroupGameStartEvent` in groups types**

In `dx-web/src/features/web/groups/types/group.ts`, update `GroupGameStartEvent` to include participants:

```typescript
export type ParticipantMember = {
  user_id: string;
  user_name: string;
};

export type SoloParticipants = {
  mode: "group_solo";
  members: ParticipantMember[];
};

export type TeamParticipants = {
  mode: "group_team";
  teams: {
    subgroup_id: string;
    subgroup_name: string;
    members: ParticipantMember[];
  }[];
};

export type Participants = SoloParticipants | TeamParticipants;

export type GroupGameStartEvent = {
  game_group_id: string;
  game_id: string;
  game_name: string;
  game_mode: "group_solo" | "group_team";
  degree: string;
  pattern: string | null;
  level_time_limit: number;
  level_id: string | null;
  level_name: string;
  participants: Participants;
};
```

- [ ] **Step 2: Add play-group types**

In `dx-web/src/features/web/play-group/types/group-play.ts`, add after the existing types:

```typescript
export type ParticipantMember = {
  user_id: string;
  user_name: string;
};

export type SoloParticipants = {
  mode: "group_solo";
  members: ParticipantMember[];
};

export type TeamParticipants = {
  mode: "group_team";
  teams: {
    subgroup_id: string;
    subgroup_name: string;
    members: ParticipantMember[];
  }[];
};

export type Participants = SoloParticipants | TeamParticipants;

export type GroupPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
};
```

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/groups/types/group.ts dx-web/src/features/web/play-group/types/group-play.ts
git commit -m "feat: add participant roster and player complete event types"
```

---

### Task 4: Store participants in sessionStorage on game start

**Files:**
- Modify: `dx-web/src/features/web/groups/components/group-game-room.tsx:63-67`

- [ ] **Step 1: Store participants before navigating**

In `group-game-room.tsx`, update the `onGameStart` handler inside `useGroupEvents` to save participants to sessionStorage before navigating:

```typescript
    onGameStart: (event) => {
      // Store participants for the play-group loading screen
      try {
        sessionStorage.setItem(
          `group-participants:${event.game_group_id}`,
          JSON.stringify(event.participants)
        );
      } catch {
        // sessionStorage may be unavailable; play-group will still work without roster
      }
      router.push(
        `/hall/play-group/${event.game_id}?groupId=${event.game_group_id}&degree=${event.degree}${event.pattern ? `&pattern=${event.pattern}` : ""}&levelTimeLimit=${event.level_time_limit}&gameMode=${event.game_mode}${event.level_id ? `&level=${event.level_id}` : ""}`
      );
    },
```

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/groups/components/group-game-room.tsx
git commit -m "feat: store participants in sessionStorage on group game start"
```

---

### Task 5: Add participants state to Zustand store

**Files:**
- Modify: `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts`

- [ ] **Step 1: Add state fields and actions**

In `use-group-play-store.ts`:

1. Add import for the new type at the top:

```typescript
import type { GroupLevelCompleteEvent, Participants } from "../types/group-play";
```

2. Add to `GroupPlayState` interface (after `groupResult`):

```typescript
  participants: Participants | null;
  completedPlayerIds: string[];
```

3. Add to `GroupPlayActions` interface:

```typescript
  setParticipants: (data: Participants) => void;
  addCompletedPlayer: (userId: string) => void;
```

4. Add to `initialState`:

```typescript
  participants: null,
  completedPlayerIds: [],
```

5. Add to the store's `initSession` action — include `participants` in the set call. Add this to the `initSession` data parameter type:

```typescript
    participants?: Participants | null;
```

And in the `initSession` set call, add:

```typescript
        participants: data.participants ?? null,
        completedPlayerIds: [],
```

6. Add the new action implementations:

```typescript
    setParticipants: (data) => set({ participants: data }),
    addCompletedPlayer: (userId) =>
      set((s) => ({
        completedPlayerIds: s.completedPlayerIds.includes(userId)
          ? s.completedPlayerIds
          : [...s.completedPlayerIds, userId],
      })),
```

7. Update `setGroupResult` to also reset completedPlayerIds:

```typescript
    setGroupResult: (result) => set({ groupPhase: "result", groupResult: result, completedPlayerIds: [] }),
```

8. Update `exitGame` to include the new fields in the reset:

The existing `exitGame: () => set({ ...initialState })` already resets everything since `initialState` includes `participants: null` and `completedPlayerIds: []`.

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-group/hooks/use-group-play-store.ts
git commit -m "feat: add participants and completedPlayerIds to group play store"
```

---

### Task 6: Read participants in loading screen

**Files:**
- Modify: `dx-web/src/features/web/play-group/components/group-play-loading-screen.tsx:106,211-228`

- [ ] **Step 1: Read from sessionStorage and pass to store**

In `group-play-loading-screen.tsx`:

1. Add import at top:

```typescript
import type { Participants } from "../types/group-play";
```

2. Add `setParticipants` to the store selector (alongside `initSession`):

```typescript
  const initGroupSession = useGroupPlayStore((s) => s.initSession);
  const setParticipants = useGroupPlayStore((s) => s.setParticipants);
  const initGameSession = useGameStore((s) => s.initSession);
```

3. Inside `loadGameData()`, after `setProgress(0)` and before Step 1 (start session), read participants from sessionStorage:

```typescript
        // Read participants roster from sessionStorage (stored by game room on game start)
        let participants: Participants | null = null;
        try {
          const key = `group-participants:${gameGroupId}`;
          const raw = sessionStorage.getItem(key);
          if (raw) {
            participants = JSON.parse(raw) as Participants;
            sessionStorage.removeItem(key);
          }
        } catch {
          // Proceed without participants if sessionStorage read fails
        }
```

4. After calling `initGroupSession(sessionInit)` (around line 228), add:

```typescript
        // Store participants in the group play store
        if (participants) {
          setParticipants(participants);
        }
```

Alternatively, pass participants into `sessionInit`:

```typescript
        const sessionInit = {
          sessionId: sessionResult.data.id,
          levelSessionId: levelResult.data.id,
          gameId,
          gameMode,
          degree,
          pattern,
          levelId: resolvedLevelId,
          contentItems: contentResult.data as ContentItem[],
          startFromIndex,
          gameGroupId,
          levelTimeLimit,
          ...(restored && { restored }),
          ...(participants && { participants }),
        };
```

This approach uses the `participants` field added to `initSession` in Task 5, keeping it atomic with the session initialization.

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-group/components/group-play-loading-screen.tsx
git commit -m "feat: read participants from sessionStorage during group play loading"
```

---

### Task 7: Add `group_player_complete` SSE listener

**Files:**
- Modify: `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts`

- [ ] **Step 1: Add the new event listener**

In `use-group-play-events.ts`:

1. Add the import:

```typescript
import type { GroupLevelCompleteEvent, GroupForceEndEvent, GroupNextLevelEvent, GroupPlayerCompleteEvent } from "../types/group-play";
```

2. Add to `GroupPlayEventHandlers`:

```typescript
  onPlayerComplete?: (event: GroupPlayerCompleteEvent) => void;
```

3. Add the event listener inside the `useEffect`, after the `group_next_level` listener:

```typescript
    eventSource.addEventListener("group_player_complete", (e) => {
      const data: GroupPlayerCompleteEvent = JSON.parse(e.data);
      handlersRef.current.onPlayerComplete?.(data);
    });
```

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-group/hooks/use-group-play-events.ts
git commit -m "feat: add group_player_complete SSE event listener"
```

---

### Task 8: Wire SSE handler in shell and pass playerId

**Files:**
- Modify: `dx-web/src/features/web/play-group/components/group-play-shell.tsx:82-86,126-143,239-253`

- [ ] **Step 1: Add `onPlayerComplete` handler and `playerId` prop**

In `group-play-shell.tsx`:

1. Add `addCompletedPlayer` to the store selectors (around line 82):

```typescript
  const addCompletedPlayer = useGroupPlayStore((s) => s.addCompletedPlayer);
```

2. Add `onPlayerComplete` to the `useGroupPlayEvents` call (around line 126):

```typescript
  useGroupPlayEvents(groupId, {
    onLevelComplete: (event) => {
      setGroupResult(event);
    },
    onForceEnd: (event) => {
      const lastResult = event.results[event.results.length - 1];
      if (lastResult) {
        setGroupResult(lastResult);
      }
      setPhase("result");
    },
    onNextLevel: (event) => {
      router.push(
        `/hall/play-group/${event.game_id}?groupId=${event.game_group_id}&degree=${event.degree}${event.pattern ? `&pattern=${event.pattern}` : ""}&levelTimeLimit=${event.level_time_limit}&gameMode=${gameMode}${event.level_id ? `&level=${event.level_id}` : ""}`
      );
    },
    onPlayerComplete: (event) => {
      addCompletedPlayer(event.user_id);
    },
  });
```

3. Pass `playerId` to `GroupPlayTopBar` (around line 239):

```tsx
      <GroupPlayTopBar
        player={player}
        playerId={player.id}
        levelName={levelName}
        levelTimeLimit={levelTimeLimit}
        onExit={() => showOverlay("exit")}
        onReset={() => showOverlay("reset")}
        onSettings={() => showOverlay("settings")}
        onReport={() => showOverlay("report")}
        onFullscreen={toggleFullscreen}
        isFullscreen={isFullscreen}
        onLevelTimeUp={() => {
          setPhase("result");
          completeAndWait();
        }}
      />
```

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: May show error for `playerId` prop not yet defined on `GroupPlayTopBar`. That's expected — Task 9 adds it.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-group/components/group-play-shell.tsx
git commit -m "feat: wire group_player_complete handler and pass playerId to top bar"
```

---

### Task 9: Render member roster in GroupPlayTopBar

**Files:**
- Modify: `dx-web/src/features/web/play-group/components/group-play-top-bar.tsx`

- [ ] **Step 1: Add imports and update props**

Add these imports at the top:

```typescript
import { Check } from "lucide-react";
import { AvatarBadge } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
```

Add `playerId` to `GroupPlayTopBarProps`:

```typescript
interface GroupPlayTopBarProps {
  player: { nickname: string; avatarUrl: string | null };
  playerId: string;
  levelName: string;
  levelTimeLimit: number;
  onExit: () => void;
  onReset: () => void;
  onSettings: () => void;
  onReport: () => void;
  onFullscreen: () => void;
  isFullscreen: boolean;
  onLevelTimeUp: () => void;
}
```

Add `playerId` to the destructured props.

- [ ] **Step 2: Read participants and completedPlayerIds from store**

Inside the component, add after the existing store selectors:

```typescript
  const participants = useGroupPlayStore((s) => s.participants);
  const completedPlayerIds = useGroupPlayStore((s) => s.completedPlayerIds);
```

- [ ] **Step 3: Add the member roster section in the JSX**

Between the progress bar `div` (lines 173-180) and the stats `div` (lines 183-208), add the member roster:

```tsx
        {/* Member roster */}
        {participants && (
          <div className="border-t border-border px-3 py-2 max-h-24 overflow-y-auto">
            {participants.mode === "group_solo" ? (
              <div className="flex flex-wrap gap-1.5">
                {participants.members.map((m) => {
                  const isCompleted = completedPlayerIds.includes(m.user_id);
                  const isMe = m.user_id === playerId;
                  const color = getAvatarColor(m.user_id);
                  return (
                    <Avatar
                      key={m.user_id}
                      size="sm"
                      className={isMe ? "ring-2 ring-teal-500" : ""}
                      style={{ backgroundColor: color }}
                    >
                      <AvatarFallback
                        className="text-white text-[10px] font-bold"
                        style={{ backgroundColor: color }}
                      >
                        {m.user_name[0]?.toUpperCase()}
                      </AvatarFallback>
                      {isCompleted && (
                        <AvatarBadge className="bg-green-500">
                          <Check className="h-2 w-2 text-white" />
                        </AvatarBadge>
                      )}
                    </Avatar>
                  );
                })}
              </div>
            ) : (
              <div className="space-y-1.5">
                {participants.teams.map((team) => (
                  <div key={team.subgroup_id}>
                    <p className="text-[10px] font-medium text-muted-foreground mb-1">
                      {team.subgroup_name}
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {team.members.map((m) => {
                        const isCompleted = completedPlayerIds.includes(m.user_id);
                        const isMe = m.user_id === playerId;
                        const color = getAvatarColor(m.user_id);
                        return (
                          <Avatar
                            key={m.user_id}
                            size="sm"
                            className={isMe ? "ring-2 ring-teal-500" : ""}
                            style={{ backgroundColor: color }}
                          >
                            <AvatarFallback
                              className="text-white text-[10px] font-bold"
                              style={{ backgroundColor: color }}
                            >
                              {m.user_name[0]?.toUpperCase()}
                            </AvatarFallback>
                            {isCompleted && (
                              <AvatarBadge className="bg-green-500">
                                <Check className="h-2 w-2 text-white" />
                              </AvatarBadge>
                            )}
                          </Avatar>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
```

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 5: Verify lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/play-group/components/group-play-top-bar.tsx
git commit -m "feat: render member roster with completion indicators in group play top bar"
```

---

### Task 10: Update game rules documentation

**Files:**
- Modify: `docs/game-lsrw-group-rule.md`

- [ ] **Step 1: Add `group_player_complete` to the SSE Events table**

In `docs/game-lsrw-group-rule.md`, update the SSE Events table (around line 453-462) to include the new event. Add a row after `group_level_complete`:

```markdown
| `group_player_complete` | Individual player completes a level | `{ user_id, user_name, game_level_id }` |
```

- [ ] **Step 2: Document the participants payload in `group_game_start`**

In the "On Success" section under "Starting the Game" (around line 160-173), update the JSON example to include the `participants` field:

```json
   {
     "game_group_id": "uuid",
     "game_id": "uuid",
     "game_name": "游戏名称",
     "game_mode": "group_solo",
     "degree": "intermediate",
     "pattern": "write",
     "level_time_limit": 10,
     "level_id": "uuid",
     "level_name": "Level 1",
     "participants": {
       "mode": "group_solo",
       "members": [
         { "user_id": "uuid", "user_name": "张三" },
         { "user_id": "uuid", "user_name": "李四" }
       ]
     }
   }
```

- [ ] **Step 3: Document the member roster in the Top Bar section**

In the "Top Bar" table (around line 236-243), update the "Player panel" row description:

```markdown
| Player panel | Expandable: avatar, score, combo, progress bar, **member roster with completion indicators**, stats |
```

- [ ] **Step 4: Add a "Member Roster" subsection under "Playing Phase"**

After the existing "Top Bar" section, add:

```markdown
### Member Roster (During Play)

- Displayed below the progress bar in the player panel
- **Solo mode**: Flat row of member avatars (ShadCN `Avatar size="sm"` with deterministic color from user ID)
- **Team mode**: Subgroup name labels with member avatars grouped underneath
- **Completion indicator**: Green checkmark badge (`AvatarBadge`) appears on a player's avatar when they complete the level
- **Current player highlight**: Ring-2 teal border on the current user's avatar
- **Data source**: Participant roster embedded in `group_game_start` SSE event, stored via sessionStorage across navigation
- **Live updates**: `group_player_complete` SSE event triggers badge appearance in real-time
- **Reset**: Completion indicators reset on level transition (`group_next_level`)
```

- [ ] **Step 5: Commit**

```bash
git add docs/game-lsrw-group-rule.md
git commit -m "docs: document member roster, group_player_complete event, and participants payload"
```
