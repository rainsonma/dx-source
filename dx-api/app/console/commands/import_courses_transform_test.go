package commands

import (
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func TestCleanGameName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strip numeric prefix", "01_日常英语对话100句", "日常英语对话100句"},
		{"strip prefix and brackets", "07_【DK】基础3000词", "DK基础3000词"},
		{"long prefix and brackets", "11_【新东方】100个句子记完4500个四级单词", "新东方100个句子记完4500个四级单词"},
		{"no prefix", "日常英语", "日常英语"},
		{"prefix only digits underscore", "99_abc", "abc"},
		{"brackets without prefix", "【测试】内容", "测试内容"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanGameName(tt.input)
			if got != tt.want {
				t.Errorf("cleanGameName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWrapPhonetic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal phonetic", "ɪkˈskjuːs", "/ɪkˈskjuːs/"},
		{"empty stays empty", "", ""},
		{"simple", "hello", "/hello/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapPhonetic(tt.input)
			if got != tt.want {
				t.Errorf("wrapPhonetic(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsPunct(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"period", '.', true},
		{"comma", ',', true},
		{"exclamation", '!', true},
		{"question", '?', true},
		{"letter", 'a', false},
		{"digit", '1', false},
		{"apostrophe", '\'', true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPunct(tt.r)
			if got != tt.want {
				t.Errorf("isPunct(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestComputeDegrees(t *testing.T) {
	tests := []struct {
		name  string
		items []CourseItem
		want  []string
	}{
		{
			"only sentence",
			[]CourseItem{{Type: "sentence"}, {Type: "sentence"}},
			[]string{"advanced"},
		},
		{
			"has phrase no word",
			[]CourseItem{{Type: "phrase"}, {Type: "sentence"}},
			[]string{"intermediate", "advanced"},
		},
		{
			"has block no word",
			[]CourseItem{{Type: "block"}, {Type: "sentence"}},
			[]string{"intermediate", "advanced"},
		},
		{
			"has word",
			[]CourseItem{{Type: "word"}, {Type: "sentence"}},
			[]string{"beginner", "intermediate", "advanced"},
		},
		{
			"all types",
			[]CourseItem{{Type: "word"}, {Type: "phrase"}, {Type: "block"}, {Type: "sentence"}},
			[]string{"beginner", "intermediate", "advanced"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeDegrees(tt.items)
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
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		content    string
		details    []WordDetail
		wantCount  int
		checkItems func(t *testing.T, items []ItemEntry)
	}{
		{
			name:    "sentence with trailing period",
			content: "I like food.",
			details: []WordDetail{
				{Word: "I", Pos: strPtr("pron"), Phonetic: Phonetic{UK: "aɪ", US: "aɪ"}, Definition: "我"},
				{Word: "like", Pos: strPtr("v"), Phonetic: Phonetic{UK: "laɪk", US: "laɪk"}, Definition: "喜欢"},
				{Word: "food", Pos: strPtr("n"), Phonetic: Phonetic{UK: "fuːd", US: "fuːd"}, Definition: "食物"},
			},
			wantCount: 4,
			checkItems: func(t *testing.T, items []ItemEntry) {
				// word items
				if items[0].Item != "I" || !items[0].Answer {
					t.Errorf("item[0] = %+v", items[0])
				}
				if items[0].Phonetic.UK != "/aɪ/" {
					t.Errorf("item[0].Phonetic.UK = %q, want /aɪ/", items[0].Phonetic.UK)
				}
				if items[1].Item != "like" || !items[1].Answer {
					t.Errorf("item[1] = %+v", items[1])
				}
				// period
				if items[3].Item != "." || items[3].Answer {
					t.Errorf("item[3] = %+v", items[3])
				}
				if items[3].Definition != "句号" {
					t.Errorf("item[3].Definition = %q", items[3].Definition)
				}
				if items[3].Phonetic != nil {
					t.Errorf("punct should have nil phonetic")
				}
				if items[3].Pos != nil {
					t.Errorf("punct should have nil pos")
				}
				// positions sequential
				for i, it := range items {
					if it.Position != i+1 {
						t.Errorf("item[%d].Position = %d, want %d", i, it.Position, i+1)
					}
				}
			},
		},
		{
			name:    "hello comma world exclamation",
			content: "Hello, world!",
			details: []WordDetail{
				{Word: "Hello", Pos: strPtr("interj"), Phonetic: Phonetic{UK: "həˈləʊ", US: "həˈloʊ"}, Definition: "你好"},
				{Word: "world", Pos: strPtr("n"), Phonetic: Phonetic{UK: "wɜːld", US: "wɜːrld"}, Definition: "世界"},
			},
			wantCount: 4,
			checkItems: func(t *testing.T, items []ItemEntry) {
				if items[0].Item != "Hello" || !items[0].Answer {
					t.Errorf("item[0] = %+v", items[0])
				}
				if items[1].Item != "," || items[1].Answer {
					t.Errorf("item[1] = %+v", items[1])
				}
				if items[1].Definition != "逗号" {
					t.Errorf("item[1].Definition = %q", items[1].Definition)
				}
				if items[2].Item != "world" || !items[2].Answer {
					t.Errorf("item[2] = %+v", items[2])
				}
				if items[3].Item != "!" || items[3].Answer {
					t.Errorf("item[3] = %+v", items[3])
				}
			},
		},
		{
			name:    "single word no punctuation",
			content: "Netherlands",
			details: []WordDetail{
				{Word: "Netherlands", Pos: strPtr("n"), Phonetic: Phonetic{UK: "", US: ""}, Definition: "荷兰"},
			},
			wantCount: 1,
			checkItems: func(t *testing.T, items []ItemEntry) {
				if items[0].Item != "Netherlands" || !items[0].Answer {
					t.Errorf("item[0] = %+v", items[0])
				}
				// empty phonetic stays empty, not wrapped
				if items[0].Phonetic.UK != "" {
					t.Errorf("empty phonetic should stay empty, got %q", items[0].Phonetic.UK)
				}
			},
		},
		{
			name:    "contraction not split",
			content: "don't stop.",
			details: []WordDetail{
				{Word: "don't", Pos: strPtr("v"), Phonetic: Phonetic{UK: "dəʊnt", US: "doʊnt"}, Definition: "不要"},
				{Word: "stop", Pos: strPtr("v"), Phonetic: Phonetic{UK: "stɒp", US: "stɑːp"}, Definition: "停止"},
			},
			wantCount: 3,
			checkItems: func(t *testing.T, items []ItemEntry) {
				if items[0].Item != "don't" || !items[0].Answer {
					t.Errorf("item[0] = %+v", items[0])
				}
				if items[1].Item != "stop" || !items[1].Answer {
					t.Errorf("item[1] = %+v", items[1])
				}
				if items[2].Item != "." || items[2].Answer {
					t.Errorf("item[2] = %+v", items[2])
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformItems(tt.content, tt.details)
			if err != nil {
				t.Fatalf("transformItems() error = %v", err)
			}
			var items []ItemEntry
			if err := json.Unmarshal([]byte(result), &items); err != nil {
				t.Fatalf("unmarshal result: %v", err)
			}
			if len(items) != tt.wantCount {
				t.Fatalf("got %d items, want %d; items=%+v", len(items), tt.wantCount, items)
			}
			if tt.checkItems != nil {
				tt.checkItems(t, items)
			}
		})
	}
}

func TestTransformStructure(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got, err := transformStructure(nil)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if got != nil {
			t.Errorf("want nil, got %v", *got)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got, err := transformStructure([]SentenceStructure{})
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if got != nil {
			t.Errorf("want nil, got %v", *got)
		}
	})

	t.Run("normal input", func(t *testing.T) {
		input := []SentenceStructure{
			{Start: 0, End: 1, Text: "I", Role: json.RawMessage(`"主语"`), Type: json.RawMessage(`"subject"`), Explanation: json.RawMessage(`"代词作主语"`)},
			{Start: 2, End: 5, Text: "like", Role: json.RawMessage(`"谓语"`), Type: json.RawMessage(`"predicate"`), Explanation: json.RawMessage("null")},
			{Start: 6, End: 10, Text: "food", Role: json.RawMessage(`"宾语"`), Type: json.RawMessage(`"object"`), Explanation: json.RawMessage("null")},
			{Start: 10, End: 11, Text: ".", Role: json.RawMessage(`"标点符号"`), Type: json.RawMessage(`"punctuation"`), Explanation: json.RawMessage("null")},
		}
		result, err := transformStructure(input)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if result == nil {
			t.Fatal("want non-nil result")
		}

		var entries []StructureEntry
		if err := json.Unmarshal([]byte(*result), &entries); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(entries) != 4 {
			t.Fatalf("got %d entries, want 4", len(entries))
		}
		// start/end shifted +1
		if entries[0].Start != 1 || entries[0].End != 2 {
			t.Errorf("entry[0] start/end = %d/%d, want 1/2", entries[0].Start, entries[0].End)
		}
		// text -> content
		if entries[0].Content != "I" {
			t.Errorf("entry[0].Content = %q, want %q", entries[0].Content, "I")
		}
		// role preserved
		if entries[0].Role != "主语" {
			t.Errorf("entry[0].Role = %q", entries[0].Role)
		}
		// type -> role_en
		if entries[0].RoleEN != "subject" {
			t.Errorf("entry[0].RoleEN = %q", entries[0].RoleEN)
		}
		// 主语 color
		if entries[0].Color == nil || *entries[0].Color != "#FFF3E0" {
			t.Errorf("entry[0].Color = %v", entries[0].Color)
		}
		// 谓语 color
		if entries[1].Color == nil || *entries[1].Color != "#FCE4EC" {
			t.Errorf("entry[1].Color = %v", entries[1].Color)
		}
		// 标点符号 nil color
		if entries[3].Color != nil {
			t.Errorf("punctuation color should be nil, got %v", *entries[3].Color)
		}
		// explanation
		if entries[0].Explanation == nil || *entries[0].Explanation != "代词作主语" {
			t.Errorf("entry[0].Explanation = %v", entries[0].Explanation)
		}
		if entries[1].Explanation != nil {
			t.Errorf("entry[1].Explanation should be nil")
		}
	})

	t.Run("unknown role gets fallback color", func(t *testing.T) {
		input := []SentenceStructure{
			{Start: 0, End: 3, Text: "abc", Role: json.RawMessage(`"未知角色"`), Type: json.RawMessage(`"unknown"`)},
		}
		result, err := transformStructure(input)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		var entries []StructureEntry
		if err := json.Unmarshal([]byte(*result), &entries); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if entries[0].Color == nil || *entries[0].Color != "#F5F5F5" {
			t.Errorf("unknown role color = %v, want #F5F5F5", entries[0].Color)
		}
	})
}

func TestGenerateGameDescription(t *testing.T) {
	tests := []struct {
		name       string
		levelNames []string
		totalItems int
		typeDist   map[string]int
	}{
		{
			"basic",
			[]string{"Unit 1", "Unit 2", "Unit 3", "Unit 4"},
			100,
			map[string]int{"word": 80, "sentence": 20},
		},
		{
			"single level",
			[]string{"Lesson A"},
			10,
			map[string]int{"sentence": 10},
		},
		{
			"mixed types",
			[]string{"L1", "L2", "L3"},
			50,
			map[string]int{"word": 15, "phrase": 15, "sentence": 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateGameDescription(tt.levelNames, tt.totalItems, tt.typeDist)
			if got == "" {
				t.Error("description should not be empty")
			}
			if utf8.RuneCountInString(got) > 200 {
				t.Errorf("description too long: %d runes", utf8.RuneCountInString(got))
			}
		})
	}
}

func TestGenerateLevelDescription(t *testing.T) {
	tests := []struct {
		name  string
		items []CourseItem
	}{
		{
			"basic items",
			[]CourseItem{
				{Content: "Hello world", Type: "sentence"},
				{Content: "Good morning", Type: "sentence"},
				{Content: "Thank you", Type: "phrase"},
			},
		},
		{
			"single item",
			[]CourseItem{
				{Content: "apple", Type: "word"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateLevelDescription(tt.items)
			if got == "" {
				t.Error("description should not be empty")
			}
			if utf8.RuneCountInString(got) > 200 {
				t.Errorf("description too long: %d runes", utf8.RuneCountInString(got))
			}
		})
	}
}
