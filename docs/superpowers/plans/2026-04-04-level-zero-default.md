# Level Zero Default Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Change default user level from Lv.1 to Lv.0 for new users (0 EXP), rebalance the exponential curve with base=100 and an introductory Lv.1 at 100 EXP.

**Architecture:** Two mirrored files (Go backend + TypeScript frontend) define the level progression table. Both use the same formula: Lv.0=0, Lv.1=100, Lv.2+=exponential(base=100, multiplier=1.05). No database changes, no API format changes.

**Tech Stack:** Go (Goravel), TypeScript (Next.js), table-driven tests

**Spec:** `docs/superpowers/specs/2026-04-04-level-zero-default-design.md`

---

### File Map

| Action | File | Purpose |
|--------|------|---------|
| Modify | `dx-api/app/consts/user_level.go` | Level constants, generation, lookup functions |
| Create | `dx-api/app/consts/user_level_test.go` | Table-driven tests for level functions |
| Modify | `dx-web/src/consts/user-level.ts` | Frontend mirror of level logic |

No other files need changes. UI components display `Lv.{level}` dynamically and will show `Lv.0` automatically.

---

### Task 1: Write failing Go test

**Files:**
- Create: `dx-api/app/consts/user_level_test.go`

- [ ] **Step 1: Write table-driven test**

```go
package consts

import (
	"testing"
)

func TestGetLevel(t *testing.T) {
	tests := []struct {
		name  string
		exp   int
		want  int
		isErr bool
	}{
		{"new user 0 exp", 0, 0, false},
		{"just below Lv.1", 99, 0, false},
		{"exactly Lv.1", 100, 1, false},
		{"just below Lv.2", 199, 1, false},
		{"exactly Lv.2", 200, 2, false},
		{"just below Lv.3", 304, 2, false},
		{"exactly Lv.3", 305, 3, false},
		{"negative exp", -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLevel(tt.exp)
			if tt.isErr {
				if err == nil {
					t.Fatalf("GetLevel(%d) expected error, got level %d", tt.exp, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetLevel(%d) unexpected error: %v", tt.exp, err)
			}
			if got != tt.want {
				t.Errorf("GetLevel(%d) = %d, want %d", tt.exp, got, tt.want)
			}
		})
	}
}

func TestGetExpForLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
		isErr bool
	}{
		{"level 0", 0, 0, false},
		{"level 1", 1, 100, false},
		{"level 2", 2, 200, false},
		{"level 3", 3, 305, false},
		{"max level", MaxLevel, -1, false}, // -1 means just check no error
		{"below min", -1, 0, true},
		{"above max", MaxLevel + 1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetExpForLevel(tt.level)
			if tt.isErr {
				if err == nil {
					t.Fatalf("GetExpForLevel(%d) expected error, got %d", tt.level, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetExpForLevel(%d) unexpected error: %v", tt.level, err)
			}
			if tt.want >= 0 && got != tt.want {
				t.Errorf("GetExpForLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

func TestLevelTableBoundaries(t *testing.T) {
	// Table has 101 entries (Lv.0 through Lv.100)
	if len(userLevels) != MaxLevel+1 {
		t.Errorf("userLevels length = %d, want %d", len(userLevels), MaxLevel+1)
	}

	// First entry is Lv.0 at 0 EXP
	if userLevels[0].Level != 0 || userLevels[0].ExpRequired != 0 {
		t.Errorf("userLevels[0] = %+v, want {Level:0 ExpRequired:0}", userLevels[0])
	}

	// Second entry is Lv.1 at 100 EXP
	if userLevels[1].Level != 1 || userLevels[1].ExpRequired != 100 {
		t.Errorf("userLevels[1] = %+v, want {Level:1 ExpRequired:100}", userLevels[1])
	}

	// EXP is strictly increasing
	for i := 1; i < len(userLevels); i++ {
		if userLevels[i].ExpRequired <= userLevels[i-1].ExpRequired {
			t.Errorf("EXP not increasing at level %d: %d <= %d",
				userLevels[i].Level, userLevels[i].ExpRequired, userLevels[i-1].ExpRequired)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/consts/ -run TestGetLevel -v`

Expected: FAIL — `GetLevel(0)` returns 1, test expects 0.

---

### Task 2: Update Go level constants and functions

**Files:**
- Modify: `dx-api/app/consts/user_level.go`

- [ ] **Step 3: Update constants**

Change `baseExp` from `1000` to `100` and add `introExp`:

```go
const (
	MaxLevel   = 100
	baseExp    = 100
	introExp   = 100
	multiplier = 1.05
)
```

- [ ] **Step 4: Update generateLevels**

Replace the function body:

```go
func generateLevels() []UserLevel {
	levels := make([]UserLevel, 0, MaxLevel+1)
	levels = append(levels, UserLevel{Level: 0, ExpRequired: 0})
	levels = append(levels, UserLevel{Level: 1, ExpRequired: introExp})

	cumulative := introExp
	for i := 2; i <= MaxLevel; i++ {
		cumulative += int(math.Floor(baseExp * math.Pow(multiplier, float64(i-2))))
		levels = append(levels, UserLevel{Level: i, ExpRequired: cumulative})
	}

	return levels
}
```

- [ ] **Step 5: Update GetLevel fallback**

Change the fallback return from `1` to `0`:

```go
func GetLevel(exp int) (int, error) {
	if exp < 0 {
		return 0, fmt.Errorf("failed to get level: exp must be non-negative, got %d", exp)
	}
	for i := len(userLevels) - 1; i >= 0; i-- {
		if exp >= userLevels[i].ExpRequired {
			return userLevels[i].Level, nil
		}
	}
	return 0, nil
}
```

- [ ] **Step 6: Update GetExpForLevel to accept level 0**

Change bounds check from `level < 1` to `level < 0`, and index from `level-1` to `level`:

```go
func GetExpForLevel(level int) (int, error) {
	if level < 0 || level > MaxLevel {
		return 0, fmt.Errorf("failed to get exp for level: level must be between 0 and %d, got %d", MaxLevel, level)
	}
	return userLevels[level].ExpRequired, nil
}
```

- [ ] **Step 7: Run all Go tests**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/consts/ -v`

Expected: All PASS.

- [ ] **Step 8: Run go vet**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./app/consts/`

Expected: No issues.

- [ ] **Step 9: Run go build to verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`

Expected: Clean build, no errors.

---

### Task 3: Update TypeScript level constants and functions

**Files:**
- Modify: `dx-web/src/consts/user-level.ts`

- [ ] **Step 10: Update constants**

Change `BASE_EXP` from `1_000` to `100` and add `INTRO_EXP`:

```typescript
const BASE_EXP = 100;
const INTRO_EXP = 100;
const MULTIPLIER = 1.05;
```

- [ ] **Step 11: Update generateLevels**

Replace the function body:

```typescript
function generateLevels(): UserLevel[] {
  const levels: UserLevel[] = [
    { level: 0, expRequired: 0 },
    { level: 1, expRequired: INTRO_EXP },
  ];
  let cumulative = INTRO_EXP;

  for (let i = 2; i <= MAX_LEVEL; i++) {
    cumulative += Math.floor(BASE_EXP * Math.pow(MULTIPLIER, i - 2));
    levels.push({ level: i, expRequired: cumulative });
  }

  return levels;
}
```

- [ ] **Step 12: Update getLevel fallback**

Change the fallback return from `1` to `0`:

```typescript
export function getLevel(exp: number): number {
  if (exp < 0) {
    throw new Error("exp must be non-negative");
  }
  for (let i = USER_LEVELS.length - 1; i >= 0; i--) {
    if (exp >= USER_LEVELS[i].expRequired) {
      return USER_LEVELS[i].level;
    }
  }
  return 0;
}
```

- [ ] **Step 13: Update getExpForLevel to accept level 0**

Change bounds check from `level < 1` to `level < 0`, and index from `level - 1` to `level`:

```typescript
export function getExpForLevel(level: number): number {
  if (!Number.isInteger(level) || level < 0 || level > MAX_LEVEL) {
    throw new Error(`Level must be an integer between 0 and ${MAX_LEVEL}`);
  }
  return USER_LEVELS[level].expRequired;
}
```

- [ ] **Step 14: Run frontend lint**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx next lint`

Expected: No errors.

---

### Task 4: Full verification and commit

- [ ] **Step 15: Run full Go test suite**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...`

Expected: All tests pass.

- [ ] **Step 16: Verify frontend build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx next build`

Expected: Clean build, no errors.

- [ ] **Step 17: Print level table for verification**

Run a quick Go script to print the full table:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/consts/ -run TestPrintLevelTable -v
```

Add this temporary test to see the table (remove before final commit):

```go
func TestPrintLevelTable(t *testing.T) {
	t.Log("Level | EXP Required | Increment")
	t.Log("------|-------------|----------")
	for i, ul := range userLevels {
		increment := 0
		if i > 0 {
			increment = ul.ExpRequired - userLevels[i-1].ExpRequired
		}
		t.Logf("Lv.%-3d | %12d | %d", ul.Level, ul.ExpRequired, increment)
	}
}
```

- [ ] **Step 18: Remove temp test, commit**

Remove `TestPrintLevelTable` from the test file, then commit:

```bash
git add dx-api/app/consts/user_level.go dx-api/app/consts/user_level_test.go dx-web/src/consts/user-level.ts
git commit -m "feat: default new users to Lv.0 and rebalance level curve

- Level range changed from 1-100 to 0-100 (101 entries)
- New users (0 EXP) start at Lv.0 instead of Lv.1
- Lv.1 threshold lowered to 100 EXP (10 level completions)
- Exponential base reduced from 1000 to 100 for gentler progression
- Add table-driven tests for GetLevel and GetExpForLevel"
```
