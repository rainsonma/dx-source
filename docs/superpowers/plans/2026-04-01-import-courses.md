# Import Courses Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import 47 game courses (2,971 levels, ~82K content items) from JSON files into the game system via a Go CLI command.

**Architecture:** Single Goravel console command `app:import-courses` that reads a directory of JSON course files, transforms the data (punctuation injection, phonetic normalization, structure colors), and batch-inserts into `games`, `game_levels`, and `content_items` tables. Pure transformation logic is separated into its own file for testability.

**Tech Stack:** Go, Goravel framework, GORM, PostgreSQL, UUID v7

**Spec:** `docs/superpowers/specs/2026-04-01-import-courses-design.md`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `app/console/commands/import_courses_transform.go` | JSON source types, JSONB target types, constants (punctuation map, color map), pure transformation functions |
| `app/console/commands/import_courses_transform_test.go` | Table-driven tests for all transformation functions |
| `app/console/commands/import_courses.go` | Command struct, Handle method, DB operations, progress output |
| `bootstrap/app.go` | Register the new command (modify) |

---

### Task 1: Transform Types, Constants & Function Stubs

**Files:**
- Create: `app/console/commands/import_courses_transform.go`

- [ ] **Step 1: Create the transform file with all types, constants, and stub functions**

```go
package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// --- Source JSON types ---

type CourseFile struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Sentences []CourseItem `json:"sentences"`
}

type CourseItem struct {
	ID                string              `json:"id"`
	Content           string              `json:"content"`
	Chinese           string              `json:"chinese"`
	Type              string              `json:"type"`
	SortOrder         int                 `json:"sortOrder"`
	WordDetails       []WordDetail        `json:"wordDetails"`
	SentenceStructure []SentenceStructure `json:"sentenceStructure"`
}

type WordDetail struct {
	Pos        *string  `json:"pos"`
	Word       string   `json:"word"`
	Phonetic   Phonetic `json:"phonetic"`
	Definition string   `json:"definition"`
}

type Phonetic struct {
	UK string `json:"uk"`
	US string `json:"us"`
}

type SentenceStructure struct {
	Start       int     `json:"start"`
	End         int     `json:"end"`
	Text        string  `json:"text"`
	Role        string  `json:"role"`
	Type        string  `json:"type"`
	Explanation *string `json:"explanation"`
}

// --- Target JSONB types ---

type ItemEntry struct {
	Position    int            `json:"position"`
	Item        string         `json:"item"`
	Phonetic    *PhoneticEntry `json:"phonetic"`
	Pos         *string        `json:"pos"`
	Translation string         `json:"translation"`
	Answer      bool           `json:"answer"`
}

type PhoneticEntry struct {
	UK string `json:"uk"`
	US string `json:"us"`
}

type StructureEntry struct {
	Start       int     `json:"start"`
	End         int     `json:"end"`
	Content     string  `json:"content"`
	Role        string  `json:"role"`
	RoleEn      string  `json:"role_en"`
	Explanation *string `json:"explanation"`
	Color       *string `json:"color"`
}

// --- Constants ---

var punctTranslations = map[rune]string{
	'.':  "句号",
	',':  "逗号",
	'!':  "感叹号",
	'?':  "问号",
	';':  "分号",
	':':  "冒号",
	'"':  "引号",
	'\'': "撇号",
	'-':  "连字符",
	'(':  "左括号",
	')':  "右括号",
}

var roleColors = map[string]string{
	"主语":  "#FFF3E0",
	"谓语":  "#FCE4EC",
	"宾语":  "#E3F2FD",
	"表语":  "#E8F5E9",
	"定语":  "#F3E5F5",
	"状语":  "#E0F7FA",
	"补语":  "#FFF9C4",
	"同位语": "#EFEBE9",
	"插入语": "#F1F8E9",
}

var typeNames = map[string]string{
	"word": "词汇", "phrase": "短语", "sentence": "句子", "block": "语段",
}

var folderPrefixRe = regexp.MustCompile(`^\d+_`)

// --- Function stubs (to be implemented in Task 3) ---

func cleanGameName(folderName string) string {
	panic("not implemented")
}

func wrapPhonetic(p string) string {
	panic("not implemented")
}

func isPunct(r rune) bool {
	panic("not implemented")
}

func transformItems(content string, details []WordDetail) (string, error) {
	panic("not implemented")
}

func transformStructure(structures []SentenceStructure) (*string, error) {
	panic("not implemented")
}

func computeDegrees(items []CourseItem) []string {
	panic("not implemented")
}

func generateGameDescription(levelNames []string, totalItems int, typeDist map[string]int) string {
	panic("not implemented")
}

func generateLevelDescription(items []CourseItem) string {
	panic("not implemented")
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: compiles with no errors (stubs use `panic` which is valid Go)

---

### Task 2: Transform Function Tests (TDD)

**Files:**
- Create: `app/console/commands/import_courses_transform_test.go`

- [ ] **Step 1: Write table-driven tests for all transform functions**

```go
package commands

import (
	"encoding/json"
	"testing"
)

func TestCleanGameName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"01_日常英语对话100句", "日常英语对话100句"},
		{"07_【DK】基础3000词", "DK基础3000词"},
		{"11_【新东方】100个句子记完4500个四级单词", "新东方100个句子记完4500个四级单词"},
		{"25_【朗文】9000高频词汇", "朗文9000高频词汇"},
		{"03_职场商务英语《会议话术》", "职场商务英语《会议话术》"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanGameName(tt.input)
			if got != tt.want {
				t.Errorf("cleanGameName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWrapPhonetic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ɪkˈskjuːs", "/ɪkˈskjuːs/"},
		{"", ""},
		{"aɪ", "/aɪ/"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := wrapPhonetic(tt.input)
			if got != tt.want {
				t.Errorf("wrapPhonetic(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestComputeDegrees(t *testing.T) {
	tests := []struct {
		name  string
		types []string
		want  []string
	}{
		{
			"only sentences",
			[]string{"sentence", "sentence"},
			[]string{"advanced"},
		},
		{
			"has phrase no word",
			[]string{"sentence", "phrase", "block"},
			[]string{"intermediate", "advanced"},
		},
		{
			"has word",
			[]string{"word", "phrase", "sentence"},
			[]string{"beginner", "intermediate", "advanced"},
		},
		{
			"only words",
			[]string{"word", "word"},
			[]string{"beginner", "intermediate", "advanced"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]CourseItem, len(tt.types))
			for i, typ := range tt.types {
				items[i] = CourseItem{Type: typ}
			}
			got := computeDegrees(items)
			if len(got) != len(tt.want) {
				t.Fatalf("computeDegrees() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("computeDegrees()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTransformItems(t *testing.T) {
	pos := "代词"

	t.Run("simple sentence with trailing punctuation", func(t *testing.T) {
		content := "I like food."
		details := []WordDetail{
			{Pos: &pos, Word: "I", Phonetic: Phonetic{UK: "aɪ", US: "aɪ"}, Definition: "我"},
			{Pos: nil, Word: "like", Phonetic: Phonetic{UK: "laɪk", US: "laɪk"}, Definition: "喜欢"},
			{Pos: nil, Word: "food", Phonetic: Phonetic{UK: "fuːd", US: "fuːd"}, Definition: "食物"},
		}

		jsonStr, err := transformItems(content, details)
		if err != nil {
			t.Fatal(err)
		}

		var items []ItemEntry
		if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
			t.Fatal(err)
		}

		// 3 words + 1 punctuation = 4 items
		if len(items) != 4 {
			t.Fatalf("expected 4 items, got %d", len(items))
		}

		// Check word item
		if items[0].Item != "I" || !items[0].Answer || items[0].Position != 1 {
			t.Errorf("item 0: got %+v", items[0])
		}
		if items[0].Phonetic.UK != "/aɪ/" {
			t.Errorf("phonetic not wrapped: %s", items[0].Phonetic.UK)
		}

		// Check punctuation
		if items[3].Item != "." || items[3].Answer || items[3].Position != 4 {
			t.Errorf("punctuation item: got %+v", items[3])
		}
		if items[3].Translation != "句号" {
			t.Errorf("punctuation translation: got %s", items[3].Translation)
		}
		if items[3].Phonetic != nil {
			t.Errorf("punctuation phonetic should be nil")
		}
	})

	t.Run("comma in middle", func(t *testing.T) {
		content := "Hello, world!"
		details := []WordDetail{
			{Word: "Hello", Phonetic: Phonetic{UK: "heˈləʊ", US: "heˈloʊ"}, Definition: "你好"},
			{Word: "world", Phonetic: Phonetic{UK: "wɜːld", US: "wɜːrld"}, Definition: "世界"},
		}

		jsonStr, err := transformItems(content, details)
		if err != nil {
			t.Fatal(err)
		}

		var items []ItemEntry
		json.Unmarshal([]byte(jsonStr), &items)

		// Hello + , + world + ! = 4 items
		if len(items) != 4 {
			t.Fatalf("expected 4 items, got %d: %+v", len(items), items)
		}
		if items[0].Item != "Hello" || items[1].Item != "," || items[2].Item != "world" || items[3].Item != "!" {
			t.Errorf("items = %s %s %s %s", items[0].Item, items[1].Item, items[2].Item, items[3].Item)
		}
	})

	t.Run("word with no punctuation", func(t *testing.T) {
		content := "Netherlands"
		details := []WordDetail{
			{Word: "Netherlands", Phonetic: Phonetic{UK: "ˈneðələndz", US: "ˈneðərləndz"}, Definition: "荷兰"},
		}

		jsonStr, err := transformItems(content, details)
		if err != nil {
			t.Fatal(err)
		}

		var items []ItemEntry
		json.Unmarshal([]byte(jsonStr), &items)

		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if !items[0].Answer {
			t.Error("word should have answer=true")
		}
	})

	t.Run("empty phonetic stays empty", func(t *testing.T) {
		content := "Friedel"
		details := []WordDetail{
			{Word: "Friedel", Phonetic: Phonetic{UK: "", US: ""}, Definition: "弗里德尔"},
		}

		jsonStr, _ := transformItems(content, details)

		var items []ItemEntry
		json.Unmarshal([]byte(jsonStr), &items)

		if items[0].Phonetic.UK != "" || items[0].Phonetic.US != "" {
			t.Errorf("empty phonetic should stay empty, got UK=%q US=%q", items[0].Phonetic.UK, items[0].Phonetic.US)
		}
	})

	t.Run("contraction not split", func(t *testing.T) {
		content := "don't stop."
		details := []WordDetail{
			{Word: "don't", Phonetic: Phonetic{UK: "dəʊnt", US: "doʊnt"}, Definition: "不要"},
			{Word: "stop", Phonetic: Phonetic{UK: "stɒp", US: "stɑːp"}, Definition: "停止"},
		}

		jsonStr, _ := transformItems(content, details)

		var items []ItemEntry
		json.Unmarshal([]byte(jsonStr), &items)

		// don't + stop + . = 3 items
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d: %+v", len(items), items)
		}
		if items[0].Item != "don't" {
			t.Errorf("contraction split incorrectly: %s", items[0].Item)
		}
	})
}

func TestTransformStructure(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result, err := transformStructure(nil)
		if err != nil {
			t.Fatal(err)
		}
		if result != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		result, err := transformStructure([]SentenceStructure{})
		if err != nil {
			t.Fatal(err)
		}
		if result != nil {
			t.Error("expected nil for empty slice")
		}
	})

	t.Run("transforms fields correctly", func(t *testing.T) {
		expl := "这是主语"
		input := []SentenceStructure{
			{Start: 0, End: 2, Text: "The cat", Role: "主语", Type: "subject", Explanation: &expl},
			{Start: 3, End: 3, Text: "sat", Role: "谓语", Type: "predicate", Explanation: nil},
		}

		result, err := transformStructure(input)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		var entries []StructureEntry
		json.Unmarshal([]byte(*result), &entries)

		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		// Check 0-based → 1-based
		if entries[0].Start != 1 || entries[0].End != 3 {
			t.Errorf("index shift: start=%d end=%d, want 1,3", entries[0].Start, entries[0].End)
		}

		// Check field renames
		if entries[0].Content != "The cat" {
			t.Errorf("content = %q, want 'The cat'", entries[0].Content)
		}
		if entries[0].RoleEn != "subject" {
			t.Errorf("role_en = %q, want 'subject'", entries[0].RoleEn)
		}

		// Check color
		if entries[0].Color == nil || *entries[0].Color != "#FFF3E0" {
			t.Errorf("主语 color = %v, want #FFF3E0", entries[0].Color)
		}
		if entries[1].Color == nil || *entries[1].Color != "#FCE4EC" {
			t.Errorf("谓语 color = %v, want #FCE4EC", entries[1].Color)
		}

		// Check nullable explanation
		if entries[0].Explanation == nil || *entries[0].Explanation != "这是主语" {
			t.Errorf("explanation = %v, want '这是主语'", entries[0].Explanation)
		}
		if entries[1].Explanation != nil {
			t.Errorf("nil explanation should stay nil")
		}
	})

	t.Run("标点符号 gets nil color", func(t *testing.T) {
		input := []SentenceStructure{
			{Start: 5, End: 5, Text: ".", Role: "标点符号", Type: "punctuation"},
		}

		result, _ := transformStructure(input)
		var entries []StructureEntry
		json.Unmarshal([]byte(*result), &entries)

		if entries[0].Color != nil {
			t.Errorf("标点符号 should have nil color, got %v", entries[0].Color)
		}
	})

	t.Run("unknown role gets fallback color", func(t *testing.T) {
		input := []SentenceStructure{
			{Start: 0, End: 0, Text: "wow", Role: "感叹语", Type: "interjection"},
		}

		result, _ := transformStructure(input)
		var entries []StructureEntry
		json.Unmarshal([]byte(*result), &entries)

		if entries[0].Color == nil || *entries[0].Color != "#F5F5F5" {
			t.Errorf("unknown role color = %v, want #F5F5F5", entries[0].Color)
		}
	})
}

func TestGenerateGameDescription(t *testing.T) {
	desc := generateGameDescription(
		[]string{"购物与询问", "餐饮与订餐", "旅行与住宿", "工作面试"},
		100,
		map[string]int{"sentence": 90, "phrase": 10},
	)
	if len([]rune(desc)) > 200 {
		t.Errorf("description too long: %d chars", len([]rune(desc)))
	}
	if desc == "" {
		t.Error("description should not be empty")
	}
}

func TestGenerateLevelDescription(t *testing.T) {
	items := []CourseItem{
		{Content: "Hello, how are you?", Type: "sentence", Chinese: "你好，你怎么样？"},
		{Content: "I'm fine, thanks.", Type: "sentence", Chinese: "我很好，谢谢。"},
	}
	desc := generateLevelDescription(items)
	if len([]rune(desc)) > 200 {
		t.Errorf("description too long: %d chars", len([]rune(desc)))
	}
	if desc == "" {
		t.Error("description should not be empty")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/console/commands/ -v -count=1`
Expected: FAIL — all tests panic with "not implemented"

---

### Task 3: Implement Transform Functions

**Files:**
- Modify: `app/console/commands/import_courses_transform.go`

- [ ] **Step 1: Replace all stub functions with real implementations**

Replace the stub section (`// --- Function stubs ---` onwards) with:

```go
// --- Transformation functions ---

func cleanGameName(folderName string) string {
	name := folderPrefixRe.ReplaceAllString(folderName, "")
	name = strings.ReplaceAll(name, "【", "")
	name = strings.ReplaceAll(name, "】", "")
	return name
}

func wrapPhonetic(p string) string {
	if p == "" {
		return ""
	}
	return "/" + p + "/"
}

func isPunct(r rune) bool {
	_, ok := punctTranslations[r]
	return ok
}

func transformItems(content string, details []WordDetail) (string, error) {
	tokens := strings.Fields(content)

	var items []ItemEntry
	pos := 1
	detailIdx := 0

	for _, token := range tokens {
		// Strip leading punctuation
		var leadPuncts []rune
		for len(token) > 0 {
			r, size := utf8.DecodeRuneInString(token)
			if isPunct(r) {
				leadPuncts = append(leadPuncts, r)
				token = token[size:]
			} else {
				break
			}
		}

		// Strip trailing punctuation
		var trailPuncts []rune
		for len(token) > 0 {
			r, size := utf8.DecodeLastRuneInString(token)
			if isPunct(r) {
				trailPuncts = append([]rune{r}, trailPuncts...)
				token = token[:len(token)-size]
			} else {
				break
			}
		}

		// Add leading punctuation items
		for _, r := range leadPuncts {
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Phonetic:    nil,
				Pos:         nil,
				Translation: punctTranslations[r],
				Answer:      false,
			})
			pos++
		}

		// Match word to next wordDetail
		if token != "" && detailIdx < len(details) {
			wd := details[detailIdx]
			items = append(items, ItemEntry{
				Position: pos,
				Item:     wd.Word,
				Phonetic: &PhoneticEntry{
					UK: wrapPhonetic(wd.Phonetic.UK),
					US: wrapPhonetic(wd.Phonetic.US),
				},
				Pos:         wd.Pos,
				Translation: wd.Definition,
				Answer:      true,
			})
			pos++
			detailIdx++
		}

		// Add trailing punctuation items
		for _, r := range trailPuncts {
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Phonetic:    nil,
				Pos:         nil,
				Translation: punctTranslations[r],
				Answer:      false,
			})
			pos++
		}
	}

	data, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("marshal items: %w", err)
	}
	return string(data), nil
}

func transformStructure(structures []SentenceStructure) (*string, error) {
	if len(structures) == 0 {
		return nil, nil
	}

	entries := make([]StructureEntry, len(structures))
	for i, s := range structures {
		var color *string
		if c, ok := roleColors[s.Role]; ok {
			color = &c
		} else if s.Role != "标点符号" {
			fallback := "#F5F5F5"
			color = &fallback
		}

		entries[i] = StructureEntry{
			Start:       s.Start + 1,
			End:         s.End + 1,
			Content:     s.Text,
			Role:        s.Role,
			RoleEn:      s.Type,
			Explanation: s.Explanation,
			Color:       color,
		}
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshal structure: %w", err)
	}
	str := string(data)
	return &str, nil
}

func computeDegrees(items []CourseItem) []string {
	hasWord := false
	hasBlockOrPhrase := false

	for _, item := range items {
		switch item.Type {
		case "word":
			hasWord = true
		case "block", "phrase":
			hasBlockOrPhrase = true
		}
	}

	if hasWord {
		return []string{"beginner", "intermediate", "advanced"}
	}
	if hasBlockOrPhrase {
		return []string{"intermediate", "advanced"}
	}
	return []string{"advanced"}
}

func generateGameDescription(levelNames []string, totalItems int, typeDist map[string]int) string {
	levelCount := len(levelNames)
	typeDesc := describeTypes(typeDist)

	samples := levelNames
	if len(samples) > 3 {
		samples = samples[:3]
	}
	sampleStr := strings.Join(samples, "、")

	desc := fmt.Sprintf("共%d个学习单元，%d个学习内容。%s，涵盖%s等主题。",
		levelCount, totalItems, typeDesc, sampleStr)

	runes := []rune(desc)
	if len(runes) > 200 {
		desc = string(runes[:197]) + "..."
	}
	return desc
}

func generateLevelDescription(items []CourseItem) string {
	if len(items) == 0 {
		return ""
	}

	typeDist := map[string]int{}
	for _, item := range items {
		typeDist[item.Type]++
	}

	primaryType := ""
	maxCount := 0
	for t, c := range typeDist {
		if c > maxCount {
			primaryType = t
			maxCount = c
		}
	}

	var samples []string
	for i, item := range items {
		if i >= 2 {
			break
		}
		s := item.Content
		runes := []rune(s)
		if len(runes) > 50 {
			s = string(runes[:47]) + "..."
		}
		samples = append(samples, "「"+s+"」")
	}

	desc := fmt.Sprintf("收录%d个%s，如%s等。",
		len(items), typeNames[primaryType], strings.Join(samples, ""))

	runes := []rune(desc)
	if len(runes) > 200 {
		desc = string(runes[:197]) + "..."
	}
	return desc
}

func describeTypes(dist map[string]int) string {
	total := 0
	for _, v := range dist {
		total += v
	}
	if total == 0 {
		return "学习内容"
	}

	maxType := ""
	maxCount := 0
	for t, c := range dist {
		if c > maxCount {
			maxType = t
			maxCount = c
		}
	}

	if float64(maxCount)/float64(total) > 0.7 {
		return fmt.Sprintf("以%s为主", typeNames[maxType])
	}

	var types []string
	for t, c := range dist {
		if c > 0 {
			types = append(types, typeNames[t])
		}
	}
	return fmt.Sprintf("包含%s等多种内容", strings.Join(types, "、"))
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/console/commands/ -v -count=1`
Expected: all tests PASS

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
git add app/console/commands/import_courses_transform.go app/console/commands/import_courses_transform_test.go
git commit -m "feat: add course import transformation functions with tests"
```

---

### Task 4: Command Implementation

**Files:**
- Create: `app/console/commands/import_courses.go`

- [ ] **Step 1: Create the command file with full Handle method**

```go
package commands

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type ImportCourses struct{}

func (c *ImportCourses) Signature() string {
	return "app:import-courses {path} {--force}"
}

func (c *ImportCourses) Description() string {
	return "Import course JSON files from a directory into the game system"
}

func (c *ImportCourses) Extend() command.Extend {
	return command.Extend{}
}

func (c *ImportCourses) Handle(ctx console.Context) error {
	start := time.Now()
	dirPath := ctx.Argument(0)
	force := ctx.OptionBool("force")

	if dirPath == "" {
		return fmt.Errorf("directory path is required")
	}

	query := facades.Orm().Query()

	// 1. Look up 实用英语 category
	var category models.GameCategory
	if err := query.Where("name", "实用英语").WhereNull("parent_id").First(&category); err != nil || category.ID == "" {
		return fmt.Errorf("category 实用英语 not found")
	}
	categoryID := category.ID

	// 2. Load top 1202 user IDs
	var users []models.User
	if err := query.Order("created_at asc").Limit(1202).Find(&users); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("no users found in database")
	}
	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	ctx.NewLine()
	ctx.Info(fmt.Sprintf("  Category: %s (%s)", category.Name, categoryID))
	ctx.Info(fmt.Sprintf("  Users loaded: %d", len(userIDs)))

	// 3. Read and sort directories
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var folders []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			folders = append(folders, e)
		}
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name() < folders[j].Name()
	})

	if len(folders) == 0 {
		return fmt.Errorf("no subdirectories found in %s", dirPath)
	}

	ctx.Info(fmt.Sprintf("  Games found: %d", len(folders)))
	ctx.NewLine()

	// 4. Handle --force cleanup
	if force {
		cleaned := forceCleanup(categoryID, folders)
		ctx.Info(fmt.Sprintf("  Force cleanup: removed %d existing games", cleaned))
		ctx.NewLine()
	}

	// 5. Process each folder
	succeeded, skipped, failed := 0, 0, 0

	for folderIdx, folder := range folders {
		gameName := cleanGameName(folder.Name())
		prefix := fmt.Sprintf("  [%d/%d] %s", folderIdx+1, len(folders), gameName)

		// Check existing
		if !force {
			var existing models.Game
			if err := query.Where("name", gameName).Where("game_category_id", categoryID).First(&existing); err == nil && existing.ID != "" {
				ctx.Warning(fmt.Sprintf("%s (skipped)", prefix))
				skipped++
				continue
			}
		}

		// Read JSON files
		folderPath := filepath.Join(dirPath, folder.Name())
		levels, totalItems, typeDist, err := parseLevels(folderPath)
		if err != nil {
			ctx.Error(fmt.Sprintf("%s ✗ %v", prefix, err))
			failed++
			continue
		}
		if len(levels) == 0 {
			continue
		}

		// Generate game description
		levelNames := make([]string, len(levels))
		for i, l := range levels {
			levelNames[i] = l.Title
		}
		gameDesc := generateGameDescription(levelNames, totalItems, typeDist)

		// Transaction per game
		tx, err := query.Begin()
		if err != nil {
			ctx.Error(fmt.Sprintf("%s ✗ begin tx: %v", prefix, err))
			failed++
			continue
		}

		gameID := uuid.Must(uuid.NewV7()).String()
		userID := userIDs[rand.IntN(len(userIDs))]

		if err := tx.Create(&models.Game{
			ID:             gameID,
			Name:           gameName,
			Description:    &gameDesc,
			UserID:         &userID,
			Mode:           "word-sentence",
			GameCategoryID: &categoryID,
			Order:          float64(folderIdx * 1000),
			IsActive:       true,
			Status:         "published",
		}); err != nil {
			_ = tx.Rollback()
			ctx.Error(fmt.Sprintf("%s ✗ create game: %v", prefix, err))
			failed++
			continue
		}

		if err := insertLevels(tx, gameID, levels); err != nil {
			_ = tx.Rollback()
			ctx.Error(fmt.Sprintf("%s ✗ %v", prefix, err))
			failed++
			continue
		}

		if err := tx.Commit(); err != nil {
			ctx.Error(fmt.Sprintf("%s ✗ commit: %v", prefix, err))
			failed++
			continue
		}

		ctx.Info(fmt.Sprintf("%s (%d levels, %d items) ✓", prefix, len(levels), totalItems))
		succeeded++
	}

	ctx.NewLine()
	ctx.Info(fmt.Sprintf("  Done: %d succeeded, %d skipped, %d failed. Time: %v",
		succeeded, skipped, failed, time.Since(start)))
	return nil
}

// parseLevels reads all JSON files in a folder and returns parsed levels.
func parseLevels(folderPath string) ([]CourseFile, int, map[string]int, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("read dir: %w", err)
	}

	var jsonFiles []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			jsonFiles = append(jsonFiles, f)
		}
	}
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i].Name() < jsonFiles[j].Name()
	})

	var levels []CourseFile
	totalItems := 0
	typeDist := map[string]int{}

	for _, jf := range jsonFiles {
		raw, err := os.ReadFile(filepath.Join(folderPath, jf.Name()))
		if err != nil {
			continue
		}
		var cf CourseFile
		if err := json.Unmarshal(raw, &cf); err != nil {
			continue
		}

		// Filter items with empty wordDetails
		var valid []CourseItem
		for _, item := range cf.Sentences {
			if len(item.WordDetails) > 0 {
				valid = append(valid, item)
			}
		}
		cf.Sentences = valid

		if len(valid) == 0 {
			continue
		}

		levels = append(levels, cf)
		totalItems += len(valid)
		for _, item := range valid {
			typeDist[item.Type]++
		}
	}

	return levels, totalItems, typeDist, nil
}

// insertLevels creates game_levels and content_items within an existing transaction.
func insertLevels(tx interface {
	Create(value any) error
}, gameID string, levels []CourseFile) error {
	for levelIdx, level := range levels {
		levelID := uuid.Must(uuid.NewV7()).String()
		degrees := computeDegrees(level.Sentences)
		levelDesc := generateLevelDescription(level.Sentences)

		if err := tx.Create(&models.GameLevel{
			ID:           levelID,
			GameID:       gameID,
			Name:         level.Title,
			Description:  &levelDesc,
			Order:        float64(levelIdx * 1000),
			PassingScore: 0,
			Degrees:      degrees,
			IsActive:     true,
		}); err != nil {
			return fmt.Errorf("create level %q: %w", level.Title, err)
		}

		// Create content items in batches
		batch := make([]models.ContentItem, 0, 100)
		for _, item := range level.Sentences {
			itemsJSON, err := transformItems(item.Content, item.WordDetails)
			if err != nil {
				return fmt.Errorf("transform items for %q: %w", item.Content, err)
			}

			structureJSON, err := transformStructure(item.SentenceStructure)
			if err != nil {
				return fmt.Errorf("transform structure for %q: %w", item.Content, err)
			}

			translation := item.Chinese

			batch = append(batch, models.ContentItem{
				ID:          uuid.Must(uuid.NewV7()).String(),
				GameLevelID: levelID,
				Content:     item.Content,
				ContentType: item.Type,
				Translation: &translation,
				Items:       &itemsJSON,
				Structure:   structureJSON,
				Order:       float64(item.SortOrder * 1000),
				IsActive:    true,
			})

			if len(batch) >= 100 {
				if err := tx.Create(&batch); err != nil {
					return fmt.Errorf("batch insert items: %w", err)
				}
				batch = batch[:0]
			}
		}

		// Flush remaining
		if len(batch) > 0 {
			if err := tx.Create(&batch); err != nil {
				return fmt.Errorf("batch insert remaining items: %w", err)
			}
		}
	}
	return nil
}

// forceCleanup deletes previously imported games and their children.
func forceCleanup(categoryID string, folders []os.DirEntry) int {
	query := facades.Orm().Query()

	nameSet := make(map[string]bool)
	for _, f := range folders {
		nameSet[cleanGameName(f.Name())] = true
	}

	var games []models.Game
	query.Where("game_category_id", categoryID).Find(&games)

	deleted := 0
	for _, g := range games {
		if !nameSet[g.Name] {
			continue
		}

		var levels []models.GameLevel
		query.Where("game_id", g.ID).Find(&levels)

		for _, l := range levels {
			query.Where("game_level_id", l.ID).Delete(&models.ContentItem{})
		}
		query.Where("game_id", g.ID).Delete(&models.GameLevel{})
		query.Where("id", g.ID).Delete(&models.Game{})
		deleted++
	}
	return deleted
}

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Run unit tests still pass**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./app/console/commands/ -v -count=1`
Expected: all tests PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
git add app/console/commands/import_courses.go
git commit -m "feat: add import-courses command with DB import logic"
```

---

### Task 5: Register Command

**Files:**
- Modify: `bootstrap/app.go`

- [ ] **Step 1: Add ImportCourses to the commands registration**

In `bootstrap/app.go`, find the `WithCommands` block:

```go
return []console.Command{
    &commands.UpdatePlayStreaks{},
    &commands.ResetEnergyBeans{},
}
```

Add the new command:

```go
return []console.Command{
    &commands.UpdatePlayStreaks{},
    &commands.ResetEnergyBeans{},
    &commands.ImportCourses{},
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Verify command is registered**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan list`
Expected: `app:import-courses` appears in the command list

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
git add bootstrap/app.go
git commit -m "feat: register import-courses command"
```

---

### Task 6: Integration Test — Run the Import

- [ ] **Step 1: Ensure database is running and seeded**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan migrate && go run . artisan db:seed`
Expected: migrations and seeds complete (game_categories and users exist)

- [ ] **Step 2: Run the import command**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan app:import-courses "/Users/rainsen/Programs/Projects/dx-courses-copy/实用英语"`
Expected:
```
  Category: 实用英语 (...)
  Users loaded: 1202
  Games found: 47

  [1/47] 100篇新闻英语听力-BBC新闻 (...) ✓
  ...
  [47/47] 朗文9000高频词汇 (...) ✓

  Done: 47 succeeded, 0 skipped, 0 failed. Time: ...
```

- [ ] **Step 3: Verify data in database**

Run spot checks via `psql` or similar:
```sql
-- Game count under 实用英语
SELECT COUNT(*) FROM games WHERE game_category_id = (SELECT id FROM game_categories WHERE name = '实用英语' AND parent_id IS NULL);
-- Expected: 47

-- Level count
SELECT COUNT(*) FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE game_category_id = (SELECT id FROM game_categories WHERE name = '实用英语' AND parent_id IS NULL));
-- Expected: ~2971

-- Content item count
SELECT COUNT(*) FROM content_items WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE game_category_id = (SELECT id FROM game_categories WHERE name = '实用英语' AND parent_id IS NULL)));
-- Expected: ~82000 (minus items with empty wordDetails)

-- Verify items JSONB has phonetic wrapping
SELECT items FROM content_items WHERE items IS NOT NULL LIMIT 1;
-- Expected: phonetic fields wrapped with /

-- Verify structure JSONB has 1-based indexing and colors
SELECT structure FROM content_items WHERE structure IS NOT NULL LIMIT 1;
-- Expected: start/end are 1-based, color fields present

-- Verify degrees
SELECT name, degrees FROM game_levels LIMIT 5;
-- Expected: varying degree arrays based on content types
```

- [ ] **Step 4: Test --force flag (re-import)**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan app:import-courses "/Users/rainsen/Programs/Projects/dx-courses-copy/实用英语" --force`
Expected: cleans up and reimports all 47 games

- [ ] **Step 5: Test idempotency (run without --force)**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan app:import-courses "/Users/rainsen/Programs/Projects/dx-courses-copy/实用英语"`
Expected: all 47 games show "(skipped)"

- [ ] **Step 6: Final commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add -A
git commit -m "feat: import 47 实用英语 courses (2971 levels, ~82K items) via CLI command"
```
