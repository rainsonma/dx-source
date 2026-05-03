package api

import "testing"

func TestNormalizeVocabContent(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{"Fast", "fast"},
		{"  fast  ", "fast"},
		{"FAST  CAR", "fast car"},
		{"foo  bar  baz", "foo bar baz"},
		{"", ""},
		{"   ", ""},
	}
	for _, tc := range cases {
		if got := NormalizeVocabContent(tc.in); got != tc.out {
			t.Errorf("NormalizeVocabContent(%q) = %q, want %q", tc.in, got, tc.out)
		}
	}
}

func TestValidateVocabContent(t *testing.T) {
	good := []string{"fast", "fast car", "don't", "well-known", "abc123"}
	bad := []string{"", "   ", "fast.", "hello!", "你好", "a@b"}
	for _, s := range good {
		if err := ValidateVocabContent(s); err != nil {
			t.Errorf("ValidateVocabContent(%q) want nil, got %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ValidateVocabContent(s); err == nil {
			t.Errorf("ValidateVocabContent(%q) want error, got nil", s)
		}
	}
}

func TestValidateDefinition(t *testing.T) {
	good := []string{
		``,
		`[{"adj":"快的"}]`,
		`[{"adj":"快的"},{"v":"斋戒"}]`,
	}
	bad := []string{
		`[{"foo":"bar"}]`,         // unknown POS
		`[{"adj":"快的","v":"斋戒"}]`, // multi-key entry
		`not json`,
	}
	for _, s := range good {
		if err := ValidateDefinition(s); err != nil {
			t.Errorf("ValidateDefinition(%q) want nil, got %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ValidateDefinition(s); err == nil {
			t.Errorf("ValidateDefinition(%q) want error, got nil", s)
		}
	}
}

func TestMergeDefinition_AdditiveOnly(t *testing.T) {
	existing := `[{"adj":"快的"}]`
	new := []map[string]string{{"v": "斋戒"}, {"adj": "错误的"}}
	got, err := MergeDefinition(existing, new)
	if err != nil {
		t.Fatalf("MergeDefinition failed: %v", err)
	}
	want := `[{"adj":"快的"},{"v":"斋戒"}]`
	if got != want {
		t.Errorf("MergeDefinition got %q, want %q", got, want)
	}
}

func TestMergeDefinition_EmptyExisting(t *testing.T) {
	got, err := MergeDefinition("", []map[string]string{{"adj": "快的"}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != `[{"adj":"快的"}]` {
		t.Errorf("got %q", got)
	}
}
