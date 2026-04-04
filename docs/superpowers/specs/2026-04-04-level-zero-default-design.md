# Level Zero Default Design

## Summary

Change the user level system so new users (0 EXP) start at Lv.0 instead of Lv.1, and rebalance the exponential curve with a lower base cost for faster early progression.

## Current Behavior

- Level range: 1–100 (100 entries)
- Lv.1 requires 0 EXP (new users are immediately Lv.1)
- Formula: `baseExp=1000, multiplier=1.05`
- Lv.2 requires 1,000 EXP (100 level completions)

## New Behavior

- Level range: 0–100 (101 entries)
- Lv.0 requires 0 EXP (new users start at Lv.0)
- Lv.1 requires 100 EXP (10 level completions — quick early win)
- Lv.2+ uses exponential formula: `baseExp=100, multiplier=1.05`
- Formula for Lv.N (N >= 2): `cumExp = 100 + Σ floor(100 × 1.05^(i-2))` for i = 2..N

## Level Progression Table (Key Milestones)

| Level | EXP Required | Increment | Casual (~100 EXP/day) | Active (~300 EXP/day) |
|-------|-------------|-----------|----------------------|----------------------|
| 0 | 0 | — | — | — |
| 1 | 100 | 100 | 1 day | < 1 day |
| 2 | 200 | 100 | 2 days | < 1 day |
| 3 | 305 | 105 | 3 days | 1 day |
| 5 | 530 | ~115 | ~5 days | ~2 days |
| 10 | 1,199 | ~147 | ~12 days | ~4 days |
| 20 | ~3,150 | ~240 | ~1 month | ~11 days |
| 50 | ~19,900 | ~860 | ~6.6 months | ~2.2 months |
| 100 | ~233,600 | ~11,700 | ~6.4 years | ~2.1 years |

EXP rate assumptions: 10 EXP per level completion at 60%+ accuracy.

## Code Changes

### Backend: `dx-api/app/consts/user_level.go`

1. Change `baseExp` from `1000` to `100`
2. Update `generateLevels()`:
   - Start with `{Level: 0, ExpRequired: 0}` instead of `{Level: 1, ExpRequired: 0}`
   - Add `{Level: 1, ExpRequired: 100}` as introductory threshold
   - Exponential loop starts at `i = 2` with `cumulative = 100`
3. Update `GetLevel()`: fallback return changes from `1` to `0`
4. Update `GetExpForLevel()`: accept `level >= 0` instead of `level >= 1`

### Frontend: `dx-web/src/consts/user-level.ts`

Mirror all backend changes:

1. Change `BASE_EXP` from `1_000` to `100`
2. Update `generateLevels()` to match backend
3. Update `getLevel()`: fallback return from `1` to `0`
4. Update `getExpForLevel()`: accept `level >= 0`

### No Changes Required

- **UI components**: Already display `Lv.{level}` — will naturally show `Lv.0`
- **Profile API** (`user_service.go`): Level is computed from EXP via `GetLevel()`
- **Leaderboard**: Ranks by EXP value, not level number
- **Game session logic**: Tracks EXP earned, not user level
- **Database**: `users.exp` field is unchanged, no migration needed
- **Seeders**: Mock users have 0 EXP, will correctly be Lv.0

## Edge Cases

- `GetExpForLevel(0)` / `getExpForLevel(0)` must return `0` (currently errors)
- Progress bar at Lv.0: `(0 - 0) / (100 - 0) = 0%` — correct
- Progress bar at Lv.100: existing `level < 100` guard still works (MaxLevel unchanged)
- Existing users with EXP see a level increase (positive UX), not a regression
- No lint issues — only constants and loop bounds change in 2 files

## Risks

- **Low**: Existing users see higher level numbers after the change. This is a net positive.
- **None**: No database migration, no API format change, no breaking changes.
