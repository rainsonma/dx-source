# Private Games Feature

## Summary

Add an `is_private` boolean field to games, allowing creators to mark course games as private. Private games are only visible to their owner and hidden from all public-facing game lists and search.

## Backend (dx-api)

### Migration

Modify existing `20260322000016_create_games_table.go` — add `is_private` boolean column with default `false` and an index.

### Model

Add `IsPrivate bool` field to `Game` struct in `app/models/game.go`, following the `is_selective` pattern.

### Request Validation

Add optional `isPrivate` bool to both `CreateGameRequest` and `UpdateGameRequest` in `app/http/requests/api/course_game_request.go`.

### Controller

In `course_game_controller.go`, pass `isPrivate` from validated request to service in both `Create` and `Update` handlers.

### Service — Course Game (Owner CRUD)

In `course_game_service.go`:
- `CreateGame` — accept `isPrivate` param, persist on insert (default false).
- `UpdateGame` — accept `isPrivate` param, persist on update.
- `ListUserGames` — **no filter** on `is_private`. Owners always see their own games.

### Service — Public Games (Filtering)

In `game_service.go`, add `WHERE is_private = false` to all 4 public query paths:
- `ListPublishedGames` — browse games at `/hall/games`
- `SearchGames` — search games by name
- `GetPlayedGames` — user's played games list
- `GetGameDetail` — single game detail view (non-owner access)

### Editability

Follows existing pattern: published games cannot be edited. User must withdraw first, toggle `is_private`, then republish.

## Frontend (dx-web)

### ShadCN Switch

Verify the Switch component from shadcn/ui is installed; install if missing.

### Create Form

In `create-course-form.tsx`, add a `<Switch>` with label "私有" (default off/false).

### Edit Dialog

In `edit-game-dialog.tsx`, add a `<Switch>` with label "私有", pre-filled from current game data.

### Actions

In `course-game.action.ts`, add `isPrivate` to create and update action payloads and API request bodies.

## Behavior Summary

| Context | Private games visible? |
|---|---|
| Owner's `/hall/ai-custom` list | Yes |
| Public `/hall/games` browse | No |
| Game search | No |
| Played games list | No |
| Direct game detail (non-owner) | No |

## Defaults

- New games: `is_private = false` (public by default)
- Existing games: `false` (migration default)
