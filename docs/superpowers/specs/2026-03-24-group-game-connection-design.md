# Group Game Connection Design Spec

## Overview

Connect a published course game to a learning group. The group owner can search games, select one, choose a play mode (单人/小组), and set it as the group's current game. This is stored on the `game_groups` table via `current_game_id` and a new `game_mode` column.

## Database

### Migration change (inline in existing migration, fresh migrate)

Add to `20260322000028_create_game_groups_table.go`:
```
table.Text("game_mode").Nullable()
```

### Model change

Add to `GameGroup`:
```go
GameMode *string `gorm:"column:game_mode" json:"game_mode"`
```

### Constants (append to `consts/group.go`)

```go
const (
    GameModeSolo = "solo"
    GameModeTeam = "team"
)
```

## API Endpoints

All under `/api/groups`, all require user JWT auth.

### GroupGameController (new controller)

| Method | Path | Action | Auth |
|--------|------|--------|------|
| GET | `/groups/{id}/games/search` | SearchGames | owner only |
| PUT | `/groups/{id}/game` | SetGame | owner only |
| DELETE | `/groups/{id}/game` | ClearGame | owner only |

### SearchGames

- Query: `?q=` (search term, optional)
- Searches `games` table where `status = 'published'`
- Filters by name ILIKE `%q%`
- Returns max 20 results
- Response: `[{ id, name, mode, category_name }]` (matches existing `GameSearchResultData` shape)

### SetGame

- Request body: `{ game_id: uuid, game_mode: "solo"|"team" }`
- Validates game exists and is published
- Validates game_mode is "solo" or "team"
- Updates `game_groups.current_game_id` and `game_groups.game_mode`
- Owner only

### ClearGame

- Sets `current_game_id = NULL` and `game_mode = NULL`
- Owner only

### Request validation

```go
type SetGroupGameRequest struct {
    GameID   string `form:"game_id" json:"game_id"`
    GameMode string `form:"game_mode" json:"game_mode"`
}
// Rules: game_id required, game_mode required|in:solo,team
// game_id validated by service layer (query DB to check existence + published status)
```

### Error sentinels

- Reuse existing `ErrGameNotFound` for missing game
- Add `ErrGameNotPublished` — selected game is not in published status
- `GroupGameController` needs its own `mapGroupGameError` function handling group + game errors

### Update GetGroupDetail response

Add to `GroupDetail` struct:
- `CurrentGameID *string` — current game ID (from model, already exists)
- `GameMode *string` — current game mode
- `CurrentGameName string` — resolved game name (empty if no game set)

The service fetches the game name when `current_game_id` is set, so the frontend doesn't need an extra request. If the game has been unpublished or deleted, return `current_game_name: ""` gracefully (don't error).

## Backend File Structure

```
dx-api/app/
├── http/controllers/api/group_game_controller.go    # SearchGames, SetGame, ClearGame
├── http/requests/api/group_game_request.go          # SetGroupGameRequest
├── services/api/group_game_service.go               # SearchGamesForGroup, SetGroupGame, ClearGroupGame
├── consts/group.go                                  # (append GameModeSolo, GameModeTeam)
├── models/game_group.go                             # (add GameMode field)
└── services/api/group_service.go                    # (update GetGroupDetail to return game info)
```

### Routes (add to routes/api.go inside groups prefix)

```go
groups.Get("/{id}/games/search", groupGameController.SearchGames)
groups.Put("/{id}/game", groupGameController.SetGame)
groups.Delete("/{id}/game", groupGameController.ClearGame)
```

## Frontend

### New component: `set-game-dialog.tsx`

Modal matching the .pen design:
- Header: teal icon + "设置群课程游戏" + close button
- Search input: debounced (300ms), calls searchGamesForGroup action
- Game list: scrollable, radio selection, selected item has teal background
- Mode selector: segmented control "单人" (user icon) | "小组" (users icon)
- Confirm button: "确认设置", calls setGame action
- After confirm: close, swrMutate to refresh group detail

Props: `open`, `onOpenChange`, `groupId`, `currentGameId?`, `currentGameMode?`

### Update `group-detail-content.tsx`

Add "当前课程游戏" section in left column (between stats and invite link):
- No game: "设置课程游戏" button
- Game set: game name, mode badge ("单人"/"小组"), "更换" button, "清除" button

### Update types

```typescript
export type GroupDetail = Group & {
  is_active: boolean;
  game_mode: string | null;
  current_game_id: string | null;
  current_game_name: string | null;
};
```

### Actions (append to group.action.ts)

```typescript
searchGamesForGroup(groupId, q?) → GET /api/groups/{id}/games/search?q=
setGame(groupId, gameId, gameMode) → PUT /api/groups/{id}/game
clearGame(groupId) → DELETE /api/groups/{id}/game
```

## Out of Scope

- Gameplay mechanics for solo vs team mode
- Starting a game session from the group
- Group cover image
- QR code generation
