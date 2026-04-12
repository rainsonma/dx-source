package realtime

import "testing"

func TestUserTopic(t *testing.T) {
	if got := UserTopic("01HZA9"); got != "user:01HZA9" {
		t.Errorf("got %s", got)
	}
}

func TestUserKickTopic(t *testing.T) {
	if got := UserKickTopic("01HZA9"); got != "user:01HZA9:kick" {
		t.Errorf("got %s", got)
	}
}

func TestPkTopic(t *testing.T) {
	if got := PkTopic("pk_123"); got != "pk:pk_123" {
		t.Errorf("got %s", got)
	}
}

func TestGroupTopic(t *testing.T) {
	if got := GroupTopic("grp_xyz"); got != "group:grp_xyz" {
		t.Errorf("got %s", got)
	}
}

func TestGroupNotifyTopic(t *testing.T) {
	if got := GroupNotifyTopic("grp_xyz"); got != "group:grp_xyz:notify" {
		t.Errorf("got %s", got)
	}
}

func TestParseTopic(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKnd TopicKind
		wantID  string
		wantErr bool
	}{
		{"user valid", "user:alice", KindUser, "alice", false},
		{"user kick valid", "user:alice:kick", KindUserKick, "alice", false},
		{"pk valid", "pk:abc", KindPk, "abc", false},
		{"group valid", "group:xyz", KindGroup, "xyz", false},
		{"group notify valid", "group:xyz:notify", KindGroupNotify, "xyz", false},
		{"empty string", "", 0, "", true},
		{"single word", "user", 0, "", true},
		{"user with empty id", "user:", 0, "", true},
		{"unknown kind", "foo:bar", 0, "", true},
		{"pk with extra segment", "pk:abc:extra", 0, "", true},
		{"group with unknown suffix", "group:xyz:applications", 0, "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTopic(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("want error, got %+v", parsed)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected err: %v", err)
				return
			}
			if parsed.Kind != tc.wantKnd {
				t.Errorf("kind: got %d want %d", parsed.Kind, tc.wantKnd)
			}
			if parsed.ID != tc.wantID {
				t.Errorf("id: got %s want %s", parsed.ID, tc.wantID)
			}
		})
	}
}

func TestParseTopicRoundtrip(t *testing.T) {
	topics := []string{
		UserTopic("alice"),
		UserKickTopic("alice"),
		PkTopic("pk123"),
		GroupTopic("grp1"),
		GroupNotifyTopic("grp1"),
	}
	for _, topic := range topics {
		parsed, err := ParseTopic(topic)
		if err != nil {
			t.Errorf("parse %s: %v", topic, err)
			continue
		}
		if parsed.ID == "" {
			t.Errorf("empty id for %s", topic)
		}
	}
}
