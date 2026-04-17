package consts

import (
	"slices"
	"testing"
	"time"
	"unicode/utf8"
)

func TestBandFor(t *testing.T) {
	tests := []struct {
		hour int
		want int
	}{
		{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3},
		{5, 0}, {6, 0}, {7, 0}, {8, 0}, {9, 0}, {10, 0},
		{11, 1}, {12, 1},
		{13, 2}, {14, 2}, {15, 2}, {16, 2}, {17, 2},
		{18, 3}, {19, 3}, {20, 3}, {21, 3}, {22, 3}, {23, 3},
	}
	for _, tc := range tests {
		if got := bandFor(tc.hour); got != tc.want {
			t.Errorf("bandFor(%d) = %d, want %d", tc.hour, got, tc.want)
		}
	}
}

func TestPickGreetingTitleByHour(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	tests := []struct {
		hour  int
		title string
	}{
		{4, "晚上好 🌙"},
		{5, "早上好 👋"},
		{10, "早上好 👋"},
		{11, "中午好 🍚"},
		{12, "中午好 🍚"},
		{13, "下午好 ☕"},
		{17, "下午好 ☕"},
		{18, "晚上好 🌙"},
		{23, "晚上好 🌙"},
		{0, "晚上好 🌙"},
	}
	for _, tc := range tests {
		tt := time.Date(2026, 4, 17, tc.hour, 0, 0, 0, shanghai)
		g := PickGreeting(tt)
		if g.Title != tc.title {
			t.Errorf("hour %d: title = %q, want %q", tc.hour, g.Title, tc.title)
		}
	}
}

func TestPickGreetingConvertsToShanghai(t *testing.T) {
	// 00:00 UTC → 08:00 Shanghai → morning
	utc := time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)
	g := PickGreeting(utc)
	if g.Title != "早上好 👋" {
		t.Errorf("UTC 00:00 (Shanghai 08:00): title = %q, want 早上好 👋", g.Title)
	}
}

func TestPickGreetingSubtitleInBand(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	bandHours := map[int][]int{
		0: {5, 6, 7, 8, 9, 10},
		1: {11, 12},
		2: {13, 14, 15, 16, 17},
		3: {18, 19, 20, 21, 22, 23, 0, 1, 2, 3, 4},
	}
	for bandIdx, hours := range bandHours {
		pool := greetingBands[bandIdx].subtitles
		for _, hour := range hours {
			tt := time.Date(2026, 4, 17, hour, 0, 0, 0, shanghai)
			for i := range 50 {
				g := PickGreeting(tt)
				if !slices.Contains(pool, g.Subtitle) {
					t.Errorf("hour %d iter %d: subtitle %q not in band %d pool",
						hour, i, g.Subtitle, bandIdx)
				}
			}
		}
	}
}

func TestGreetingBandInvariants(t *testing.T) {
	if len(greetingBands) != 4 {
		t.Fatalf("greetingBands length = %d, want 4", len(greetingBands))
	}
	for i, band := range greetingBands {
		if band.title == "" {
			t.Errorf("band %d has empty title", i)
		}
		if len(band.subtitles) != 5 {
			t.Errorf("band %d has %d subtitles, want 5", i, len(band.subtitles))
		}
		for j, s := range band.subtitles {
			if s == "" {
				t.Errorf("band %d subtitle %d is empty", i, j)
			}
			if runes := utf8.RuneCountInString(s); runes > 20 {
				t.Errorf("band %d subtitle %d = %q has %d runes, want ≤ 20",
					i, j, s, runes)
			}
		}
	}
}
