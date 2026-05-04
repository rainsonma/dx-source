package consts

import "testing"

func TestIsValidPos(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"n", true},
		{"v", true},
		{"adj", true},
		{"adv", true},
		{"prep", true},
		{"conj", true},
		{"pron", true},
		{"art", true},
		{"num", true},
		{"int", true},
		{"aux", true},
		{"det", true},
		{"verb", false},
		{"adjective", false},
		{"", false},
		{"N", false},
		{"phr", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := IsValidPos(tc.input); got != tc.want {
				t.Errorf("IsValidPos(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestAllPosCount(t *testing.T) {
	if len(AllPos) != 12 {
		t.Errorf("expected 12 POS keys, got %d", len(AllPos))
	}
	if len(PosLabels) != len(AllPos) {
		t.Errorf("PosLabels (%d) and AllPos (%d) must be the same size", len(PosLabels), len(AllPos))
	}
}
