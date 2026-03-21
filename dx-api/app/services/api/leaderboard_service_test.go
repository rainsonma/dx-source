package api

import (
	"testing"
)

func TestWindowStartSQL(t *testing.T) {
	tests := []struct {
		period string
		want   string
	}{
		{"day", "DATE_TRUNC('day', NOW())"},
		{"week", "DATE_TRUNC('week', NOW())"},
		{"month", "DATE_TRUNC('month', NOW())"},
		{"unknown", "DATE_TRUNC('day', NOW())"},
		{"", "DATE_TRUNC('day', NOW())"},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			got := windowStartSQL(tt.period)
			if got != tt.want {
				t.Errorf("windowStartSQL(%q) = %q, want %q", tt.period, got, tt.want)
			}
		})
	}
}

func TestSafeUID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
	}{
		{"empty returns placeholder", "", "___none___"},
		{"non-empty returns as-is", "user123", "user123"},
		{"whitespace returned as-is", " ", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeUID(tt.input)
			if got != tt.want {
				t.Errorf("safeUID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildLeaderboardResult(t *testing.T) {
	t.Run("empty rows", func(t *testing.T) {
		result := buildLeaderboardResult(nil, "user1")
		if len(result.Entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result.Entries))
		}
		if result.MyRank != nil {
			t.Error("expected nil myRank for empty rows")
		}
	})

	t.Run("user in top 100", func(t *testing.T) {
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", Value: 1000, Rank: 1},
			{ID: "user2", Username: "bob", Value: 800, Rank: 2},
			{ID: "user3", Username: "charlie", Value: 600, Rank: 3},
		}
		result := buildLeaderboardResult(rows, "user2")

		if len(result.Entries) != 3 {
			t.Errorf("expected 3 entries, got %d", len(result.Entries))
		}
		if result.MyRank == nil {
			t.Fatal("expected non-nil myRank")
		}
		if result.MyRank.ID != "user2" {
			t.Errorf("myRank.ID = %q, want %q", result.MyRank.ID, "user2")
		}
		if result.MyRank.Rank != 2 {
			t.Errorf("myRank.Rank = %d, want 2", result.MyRank.Rank)
		}
		if result.MyRank.Value != 800 {
			t.Errorf("myRank.Value = %d, want 800", result.MyRank.Value)
		}
	})

	t.Run("user not in results", func(t *testing.T) {
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", Value: 1000, Rank: 1},
		}
		result := buildLeaderboardResult(rows, "unknown")

		if len(result.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(result.Entries))
		}
		if result.MyRank != nil {
			t.Error("expected nil myRank for unknown user")
		}
	})

	t.Run("user beyond rank 100", func(t *testing.T) {
		rows := []leaderboardRow{
			{ID: "top1", Username: "top", Value: 9999, Rank: 1},
			{ID: "me", Username: "me", Value: 10, Rank: 150},
		}
		result := buildLeaderboardResult(rows, "me")

		// Only rank <= 100 should be in entries
		if len(result.Entries) != 1 {
			t.Errorf("expected 1 entry (top 100 only), got %d", len(result.Entries))
		}
		if result.Entries[0].ID != "top1" {
			t.Errorf("entry should be top1, got %s", result.Entries[0].ID)
		}

		// But myRank should still be populated
		if result.MyRank == nil {
			t.Fatal("expected myRank even beyond rank 100")
		}
		if result.MyRank.Rank != 150 {
			t.Errorf("myRank.Rank = %d, want 150", result.MyRank.Rank)
		}
	})

	t.Run("no userID provided", func(t *testing.T) {
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", Value: 100, Rank: 1},
		}
		result := buildLeaderboardResult(rows, "")

		if result.MyRank != nil {
			t.Error("expected nil myRank when no userID")
		}
	})

	t.Run("avatar URL resolved", func(t *testing.T) {
		avatarID := "img123"
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", AvatarID: &avatarID, Value: 100, Rank: 1},
			{ID: "user2", Username: "bob", AvatarID: nil, Value: 50, Rank: 2},
		}
		result := buildLeaderboardResult(rows, "user1")

		if result.Entries[0].AvatarURL == nil {
			t.Error("expected avatar URL for user with avatarID")
		}
		if result.Entries[1].AvatarURL != nil {
			t.Error("expected nil avatar URL for user without avatarID")
		}
	})

	t.Run("empty avatarID string", func(t *testing.T) {
		emptyID := ""
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", AvatarID: &emptyID, Value: 100, Rank: 1},
		}
		result := buildLeaderboardResult(rows, "")

		if result.Entries[0].AvatarURL != nil {
			t.Error("expected nil avatar URL for empty avatarID string")
		}
	})

	t.Run("nickname preserved", func(t *testing.T) {
		nick := "AliceNick"
		rows := []leaderboardRow{
			{ID: "user1", Username: "alice", Nickname: &nick, Value: 100, Rank: 1},
		}
		result := buildLeaderboardResult(rows, "")

		if result.Entries[0].Nickname == nil || *result.Entries[0].Nickname != nick {
			t.Errorf("expected nickname %q, got %v", nick, result.Entries[0].Nickname)
		}
	})
}
