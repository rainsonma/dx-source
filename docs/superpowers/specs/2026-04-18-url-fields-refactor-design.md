# URL-based Image & Audio Fields Refactor — Design

**Status:** Spec
**Date:** 2026-04-18
**Scope:** `dx-api` (Go/Goravel), `dx-web` (Next.js), `dx-mini` (WeChat mini program), `deploy/` (docker compose + nginx — no change)

## Goal

Replace image/audio foreign-key columns (`*_id` → UUID referencing the `images`/`audios` tables) with direct URL text columns (`*_url`). Remove the `images` and `audios` tables entirely. The database is a dev environment and will be reset after this change, so the refactor is done **in place** — no ALTER-TABLE migrations, no compat shims, no backfill.

## Rollout

Single coordinated change set (big-bang). No dual-writing, no temporary compatibility accepting both shapes. Justified by: dev DB, fresh reset, user instruction to edit existing files directly.

## Schema changes

### Columns renamed (UUID nullable → TEXT nullable)

| Table | Old column | New column |
|---|---|---|
| `users` | `avatar_id` | `avatar_url` |
| `adm_users` | `avatar_id` | `avatar_url` |
| `posts` | `image_id` | `image_url` |
| `games` | `cover_id` | `cover_url` |
| `game_categories` | `cover_id` | `cover_url` |
| `game_presses` | `cover_id` | `cover_url` |
| `game_groups` | `cover_id` | `cover_url` |
| `game_groups` | `invite_qrcode_id` | `invite_qrcode_url` |
| `content_items` | `uk_audio_id` | `uk_audio_url` |
| `content_items` | `us_audio_id` | `us_audio_url` |

Migrations are edited **in their existing `.go` files** — column type flips from `Uuid("x_id").Nullable()` to `Text("x_url").Nullable()`. Any existing `Index("x_id")` entry on these columns is dropped (URL columns are not filter/join keys).

### Tables deleted

- `images` — migration `20260322000017_create_images_table.go` removed
- `audios` — migration `20260322000018_create_audios_table.go` removed

Both entries removed from `bootstrap/migrations.go`.

## URL format & serving

### Storage convention

Files are stored on disk as today: `{STORAGE_PATH}/uploads/images/YYYY/MM/DD/{uuid}.{jpg|png}`. Audio files use `uploads/audios/...` if/when added (deferred — see Non-goals).

### URL format (stored in DB & returned to clients)

`/api/uploads/images/YYYY/MM/DD/{uuid}.{jpg|png}` — API-relative path. Served by the Go app directly from disk; no DB lookup.

### `POST /api/uploads/images` (rewritten)

- Request: unchanged — multipart `file` + form field `role` (required; validated against `consts.ImageRole*` set as today).
- Validations: size ≤ 2 MB, MIME ∈ {`image/jpeg`, `image/png`}, `role` ∈ allowed set.
- Behavior: generate `uuid.NewV7()`, construct `uploads/images/YYYY/MM/DD/{uuid}.{ext}`, `file.StoreAs(...)`. **No DB record is created.**
- Response body: `{"code":0,"message":"ok","data":{"url":"/api/uploads/images/YYYY/MM/DD/{uuid}.ext"}}` — only the URL field is returned.

### `GET /api/uploads/images/{year}/{month}/{day}/{filename}` (rewritten)

- Route replaces the single-segment `/uploads/images/{id}` pattern with four explicit segments.
- Path segment validation in the controller:
  - `year` must match `^\d{4}$`; `month` and `day` must match `^\d{1,2}$`.
  - `filename` must match `^[0-9a-f-]{36}\.(jpg|png)$`.
  - Any regex miss → 400 `CodeValidationError` with message `"invalid image path"`.
- Join `{STORAGE_PATH}/uploads/images/{year}/{month}/{day}/{filename}`, then `filepath.Clean` and assert the cleaned path is still prefixed by `{STORAGE_PATH}/uploads/images/` (defense in depth against traversal). Failure → 400 `CodeValidationError`.
- `os.Stat` miss → 404 with `consts.CodeImageNotFound`.
- Success: serve file with `Cache-Control: public, max-age=31536000, immutable` and `Content-Type` derived from extension: `jpg` → `image/jpeg`, `png` → `image/png` (no other extension can reach this branch because regex gates them).
- Replaces `services.GetImagePath(imageID)` (DB lookup) with pure filesystem logic in `services.ResolveImagePath(year, month, day, filename)`.

## Client-submitted URL validation

One shared helper `app/helpers/url_validate.go`:

```go
var uploadImageURLRegex = regexp.MustCompile(`^/api/uploads/images/\d{4}/\d{1,2}/\d{1,2}/[0-9a-f-]{36}\.(jpg|png)$`)

func IsUploadedImageURL(s string) bool { return uploadImageURLRegex.MatchString(s) }
```

Applied at every request validator where a client writes an image URL:

| Endpoint | Validator | Field |
|---|---|---|
| `PUT /api/user/avatar` | `UpdateAvatarRequest` in `user_request.go` | `avatar_url` |
| `POST /api/posts` | `CreatePostRequest` in `post_request.go` | `image_url` (optional) |
| `PUT /api/posts/{id}` | `UpdatePostRequest` in `post_request.go` | `image_url` (optional) |
| `POST /api/course-games` | `course_game_request.go` | `cover_url` (optional) |
| `PUT /api/course-games/{id}` | `course_game_request.go` | `cover_url` (optional) |

Invalid URL → HTTP 400, `CodeValidationError`.

**Not client-writable (no validator needed):**
- `content_items.uk_audio_url`/`us_audio_url` — externally seeded, trusted.
- `game_categories.cover_url`, `game_presses.cover_url` — admin-only (public controllers are list-only; 27 lines each, confirmed read-only).
- `game_groups.cover_url`, `game_groups.invite_qrcode_url` — no client write endpoint. `invite_qrcode_url` is server-generated via `helpers.GenerateGroupQRCode`.

## File deletions & rewrites (dx-api)

### Deleted

- `app/models/image.go`
- `app/models/audio.go`
- `app/helpers/image_url.go` (the `ImageServeURL(id)` helper — obsolete)
- `database/migrations/20260322000017_create_images_table.go`
- `database/migrations/20260322000018_create_audios_table.go`

### Registrations removed

- `bootstrap/migrations.go`: drop `&migrations.M20260322000017CreateImagesTable{}` and `&migrations.M20260322000018CreateAudiosTable{}`.

### Rewritten

- `app/services/api/upload_service.go`:
  - `UploadImageResult` struct keeps only `URL string`.
  - `UploadImage` returns `&UploadImageResult{URL: "/api/uploads/images/..."}`.
  - `GetImagePath` → `ResolveImagePath(year, month, day, filename)`; no ORM import needed.
  - `models`/`helpers` imports removed where no longer used.

- `app/http/controllers/api/upload_controller.go`:
  - `ServeImage` reads four path params, validates with regex, calls `ResolveImagePath`, serves file.

- `app/helpers/qrcode.go::GenerateGroupQRCode`:
  - Drops `models.Image` creation.
  - Returns URL string `/api/uploads/images/YYYY/MM/DD/{uuid}.png` (instead of image ID).
  - Caller `group_service.go:102` writes the URL into `invite_qrcode_url`.

- `routes/api.go`:
  - Replace `router.Get("/uploads/images/{id}", uploadController.ServeImage)` with `router.Get("/uploads/images/{year}/{month}/{day}/{filename}", uploadController.ServeImage)`.

## Service-layer simplification (dx-api)

All ID→URL resolution via `ImageServeURL` + the `images`-table batch loads disappear. Services just read the new `*_url` column and put it in the response payload.

| File | Change |
|---|---|
| `services/api/game_service.go` | Drop `coverIDs` batching + `images` join (lines 119-120, 186-187, 383-385). Read `g.CoverURL` directly. |
| `services/api/course_game_service.go` | Same as above (lines 113-114, 150-151, 191, 228, 638-640, 677). `CoverID` struct field → `CoverURL`; `"cover_id"` update key → `"cover_url"`. |
| `services/api/favorite_service.go` | Drop `coverIDs` batching (lines 86-87, 145-146). Use `g.CoverURL`. |
| `services/api/post_service.go` | In `buildPostItems`/`buildPostItem` (lines 313-441): drop image-ID batch + `ImageServeURL` calls; use `post.ImageURL` and `user.AvatarURL`. Raw SQL `SELECT ... author_avatar_id` → `author_avatar_url`. |
| `services/api/post_comment_service.go` | Drop avatar-ID resolution; use `user.AvatarURL`. |
| `services/api/user_service.go` | `GetProfile`/`UpdateAvatar`: read/write `avatar_url`. `UpdateAvatar` writes the validated URL string directly. |
| `services/api/content_service.go` | Drop audio batch (lines 56-75, currently buggy — queries `images` for audio IDs). Copy `UkAudioURL`/`UsAudioURL` directly from the row. |
| `services/api/group_service.go` | Line 102: update `invite_qrcode_url` with URL returned by `GenerateGroupQRCode`. Lines 256-257: drop `ImageServeURL` call; use `group.InviteQrcodeURL`. |
| `services/api/leaderboard_service.go` | Drop avatar-ID resolution; use `user.AvatarURL`. |
| `services/api/hall_service.go` | Same. |

## Request/response JSON field names

All camelCase keys flip consistently: `avatarId` → `avatarUrl`, `imageId` → `imageUrl`, `coverId` → `coverUrl`, `ukAudioId`/`usAudioId` → `ukAudioUrl`/`usAudioUrl`. Go struct `json:""` tags updated in lockstep with GORM tags.

## Frontend (dx-web) changes

15 files touched. Categories:

### Upload pipeline type surface

- `src/lib/api-client.ts` — `uploadApi.uploadImage()` response type narrows to `{url: string}`. `userApi.updateAvatar(url: string)` accepts a URL (not image ID); body sent as `{avatar_url: url}`.
- `src/features/com/images/types/image.types.ts` — `UploadedImage` becomes `{url: string}`. Drop `id` and `name`.
- `src/features/com/images/hooks/use-image-uploader.ts` — on Uppy upload success, extract `body.data.url` only; callback emits the URL string.
- `src/features/com/images/components/image-uploader.tsx` — `onImageChange` parameter type becomes `(url: string) => void`; remove any reference to `img.id`.

### Consumers

- `src/features/web/me/actions/me.action.ts` — `updateAvatarAction(url: string)` → `{avatar_url: url}`.
- `src/features/web/me/types/me.types.ts` — `avatarId` → `avatarUrl`.
- `src/features/web/community/schemas/post.schema.ts` — `image_id` → `image_url`.
- `src/features/web/community/actions/post.action.ts` — `image_id` → `image_url` in request payloads and types.
- `src/features/web/ai-custom/schemas/course-game.schema.ts` — `coverId` → `coverUrl`.
- `src/features/web/ai-custom/actions/course-game.action.ts` — payload field + types.
- `src/features/web/ai-custom/hooks/use-create-course-game.ts`, `use-update-course-game.ts` — `coverId` → `coverUrl`.
- `src/features/web/ai-custom/components/create-course-form.tsx`, `edit-game-dialog.tsx`, `course-detail-content.tsx`, `game-hero-card.tsx` — field rename + any `img.id` → url plumbing.
- `src/features/web/docs/topics/account/profile-edit.tsx` — documentation text change ("通过 imageUrl 绑定").

### What stays the same

The existing Uppy upload-to-`/api/uploads/images` flow continues to work. Only the response shape changes. Components that previously displayed an image by URL via `{url}` remain unchanged.

## Mini program (dx-mini) changes

One file: `miniprogram/pages/me/profile-edit/profile-edit.ts`.

- In `chooseAvatar()`: after `wx.uploadFile` success, parse `{data: {url}}` (drop `.id`).
- Request to `/api/user/avatar`: body becomes `{avatar_url: body.data.url}`.
- Local state update: `avatarUrl = body.data.url`.

Response types in the file (the profile `avatarUrl` field) unchanged — backend already returns URLs.

## Deploy / nginx

**No changes.** The nginx config already proxies `/api/*` → `dx-api`. The rewritten serve route is still under `/api/uploads/images/...`, so the same proxy rule covers it. No new volumes, no new static-file rules.

## Non-goals (deferred)

- **Audio upload endpoint.** Currently no `POST /api/uploads/audios` exists; adding one is out of scope. Audio URLs remain seeded by external processes.
- **Orphaned file cleanup.** With no DB record tracking uploads, we can't identify orphans. Not addressed here; accepted as a dev-DB trade-off.
- **CDN / absolute URLs.** Deferred — relative URLs work for dev + single-host prod. A future env-var-driven prefix (`PUBLIC_BASE_URL`) is a separate concern.

## Security considerations

- **Path traversal at serve time:** regex-validated segments + `filepath.Clean` prefix check in `ResolveImagePath` — both defenses run on every request.
- **Unauthorized URL storage:** all client-writable URL fields pass through `IsUploadedImageURL` prefix validation — clients cannot inject external URLs (no tracker pixels, no `javascript:` schemes).
- **Upload auth unchanged:** `POST /api/uploads/images` still requires user JWT. Role validation kept (per Q4).
- **Filename entropy:** UUIDv7 keeps time-ordered randomness; the 36-char lowercase-hex+hyphen regex rejects anything else.

## Testing

### Go (`go test -race ./...`)

- `tests/feature/upload_test.go` — extend or create:
  - `POST /api/uploads/images` with valid jpg → response `{url: "/api/uploads/images/..."}` only.
  - Too large / bad MIME / bad role → correct error codes unchanged.
- `tests/feature/upload_serve_test.go` — new:
  - Valid path serves file with correct content type + cache header.
  - `..` in segments → 400.
  - Non-numeric date → 400.
  - Unknown file → 404.
- `tests/feature/url_validate_test.go` — new, table-driven:
  - Matrix of valid and invalid URLs for `IsUploadedImageURL`.
- Any existing service tests that asserted `ImageServeURL` shape get updated to the new field reads.

### TypeScript (`npm run lint`)

- ESLint passes across dx-web; TypeScript compile clean.

### Build checks

- `go build ./...` — clean.
- `go vet ./...`, `gofmt`/`goimports` — clean.
- `npm run build` (dx-web) — clean.

### Manual verification

- Upload avatar via dx-web → see new URL in payload → display avatar → refresh → still displays.
- Upload post image via dx-web → view post → image shows.
- Course-game cover upload round-trip.
- WeChat mini program avatar upload round-trip.
- Group creation → QR code generated → `invite_qrcode_url` populated → image loads.

## Correctness guarantees

- The existing `content_service.go` bug (batch-loading `models.Image` for audio IDs) is fixed as a byproduct — we stop using the `images` table altogether.
- Every `*_id` → `*_url` rename is paired across: migration, model GORM/json tags, every service read, every request validator, every frontend schema/type.
- No code path remains that references `models.Image` or `models.Audio` after cleanup.

## Open questions

None. All clarifications resolved in brainstorming (URL format = API-relative path, upload response = `{url}`, validation = strict prefix match, role retained, audio upload deferred, approach = single big-bang).
