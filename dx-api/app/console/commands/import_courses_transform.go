package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Source JSON types (parsed from course files)
// ---------------------------------------------------------------------------

type CourseFile struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Sentences []CourseItem `json:"sentences"`
}

type CourseItem struct {
	ID                 string               `json:"id"`
	Content            string               `json:"content"`
	Chinese            string               `json:"chinese"`
	Type               string               `json:"type"`
	SortOrder          int                  `json:"sortOrder"`
	WordDetails        []WordDetail         `json:"wordDetails"`
	SentenceStructure  []SentenceStructure  `json:"sentenceStructure"`
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
	Start       int              `json:"start"`
	End         int              `json:"end"`
	Text        string           `json:"text"`
	Role        string           `json:"role"`
	Type        string           `json:"type"`
	Explanation json.RawMessage  `json:"explanation"`
}

// explanationString extracts a string from Explanation which may be a JSON string or array of strings.
func (s *SentenceStructure) explanationString() *string {
	if len(s.Explanation) == 0 || string(s.Explanation) == "null" {
		return nil
	}

	// Try string first
	var str string
	if err := json.Unmarshal(s.Explanation, &str); err == nil {
		return &str
	}

	// Try array of strings
	var arr []string
	if err := json.Unmarshal(s.Explanation, &arr); err == nil {
		joined := strings.Join(arr, "")
		return &joined
	}

	return nil
}

// ---------------------------------------------------------------------------
// Target JSONB types (stored in database)
// ---------------------------------------------------------------------------

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
	RoleEN      string  `json:"role_en"`
	Explanation *string `json:"explanation"`
	Color       *string `json:"color"`
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

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
	"word":     "词汇",
	"phrase":   "短语",
	"sentence": "句子",
	"block":    "语段",
}

var folderPrefixRe = regexp.MustCompile(`^\d+_`)

// ---------------------------------------------------------------------------
// Functions
// ---------------------------------------------------------------------------

func cleanGameName(folderName string) string {
	s := folderPrefixRe.ReplaceAllString(folderName, "")
	s = strings.ReplaceAll(s, "【", "")
	s = strings.ReplaceAll(s, "】", "")
	return s
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
	di := 0 // index into details

	for _, tok := range tokens {
		// strip leading punctuation
		var leading []rune
		runes := []rune(tok)
		for len(runes) > 0 && isPunct(runes[0]) {
			leading = append(leading, runes[0])
			runes = runes[1:]
		}

		// strip trailing punctuation, but preserve apostrophes inside words (contractions)
		var trailing []rune
		for len(runes) > 0 && isPunct(runes[len(runes)-1]) {
			r := runes[len(runes)-1]
			// apostrophe inside a word is part of the word (e.g. don't)
			if r == '\'' && len(runes) > 1 {
				break
			}
			trailing = append([]rune{r}, trailing...)
			runes = runes[:len(runes)-1]
		}

		// add leading punct items
		for _, r := range leading {
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Translation: punctTranslations[r],
				Answer:      false,
			})
			pos++
		}

		// add word item if remaining
		word := string(runes)
		if word != "" && di < len(details) {
			d := details[di]
			di++
			items = append(items, ItemEntry{
				Position: pos,
				Item:     word,
				Phonetic: &PhoneticEntry{
					UK: wrapPhonetic(d.Phonetic.UK),
					US: wrapPhonetic(d.Phonetic.US),
				},
				Pos:         d.Pos,
				Translation: d.Definition,
				Answer:      true,
			})
			pos++
		}

		// add trailing punct items
		for _, r := range trailing {
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Translation: punctTranslations[r],
				Answer:      false,
			})
			pos++
		}
	}

	b, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal items: %w", err)
	}
	return string(b), nil
}

func transformStructure(structures []SentenceStructure) (*string, error) {
	if len(structures) == 0 {
		return nil, nil
	}

	entries := make([]StructureEntry, 0, len(structures))
	for _, s := range structures {
		entry := StructureEntry{
			Start:       s.Start + 1,
			End:         s.End + 1,
			Content:     s.Text,
			Role:        s.Role,
			RoleEN:      s.Type,
			Explanation: s.explanationString(),
		}

		if s.Role == "标点符号" {
			entry.Color = nil
		} else if c, ok := roleColors[s.Role]; ok {
			entry.Color = &c
		} else {
			fallback := "#F5F5F5"
			entry.Color = &fallback
		}

		entries = append(entries, entry)
	}

	b, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal structure: %w", err)
	}
	result := string(b)
	return &result, nil
}

func computeDegrees(items []CourseItem) []string {
	hasWord, hasMiddle := false, false
	for _, it := range items {
		switch it.Type {
		case "word":
			hasWord = true
		case "phrase", "block":
			hasMiddle = true
		}
	}

	if hasWord {
		return []string{"beginner", "intermediate", "advanced"}
	}
	if hasMiddle {
		return []string{"intermediate", "advanced"}
	}
	return []string{"advanced"}
}

func describeTypes(typeDist map[string]int) string {
	total := 0
	for _, n := range typeDist {
		total += n
	}
	if total == 0 {
		return ""
	}

	// find dominant type
	var dominant string
	var dominantN int
	for tp, n := range typeDist {
		if n > dominantN {
			dominant = tp
			dominantN = n
		}
	}

	if float64(dominantN)/float64(total) > 0.7 {
		name := typeNames[dominant]
		if name == "" {
			name = dominant
		}
		return "以" + name + "为主"
	}

	var names []string
	for tp := range typeDist {
		name := typeNames[tp]
		if name == "" {
			name = tp
		}
		names = append(names, name)
	}
	return "包含" + strings.Join(names, "、") + "等多种内容"
}

func generateGameDescription(levelNames []string, totalItems int, typeDist map[string]int) string {
	levelCount := len(levelNames)
	typeDesc := describeTypes(typeDist)

	sampleCount := 3
	if len(levelNames) < sampleCount {
		sampleCount = len(levelNames)
	}
	samples := strings.Join(levelNames[:sampleCount], "、")

	desc := fmt.Sprintf("共%d个学习单元，%d个学习内容。%s，涵盖%s等主题。", levelCount, totalItems, typeDesc, samples)
	return truncateRunes(desc, 200)
}

func generateLevelDescription(items []CourseItem) string {
	if len(items) == 0 {
		return ""
	}

	// find primary type
	dist := map[string]int{}
	for _, it := range items {
		dist[it.Type]++
	}
	var primary string
	var primaryN int
	for tp, n := range dist {
		if n > primaryN {
			primary = tp
			primaryN = n
		}
	}
	typeName := typeNames[primary]
	if typeName == "" {
		typeName = primary
	}

	sampleCount := 2
	if len(items) < sampleCount {
		sampleCount = len(items)
	}
	var samples []string
	for i := 0; i < sampleCount; i++ {
		s := items[i].Content
		if utf8.RuneCountInString(s) > 50 {
			s = string([]rune(s)[:50])
		}
		samples = append(samples, "「"+s+"」")
	}

	desc := fmt.Sprintf("收录%d个%s，如%s等。", len(items), typeName, strings.Join(samples, ""))
	return truncateRunes(desc, 200)
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}
