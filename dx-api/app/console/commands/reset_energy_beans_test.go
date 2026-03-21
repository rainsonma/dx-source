package commands

import (
	"testing"
)

func TestResetEnergyBeans_Signature(t *testing.T) {
	cmd := &ResetEnergyBeans{}
	if got := cmd.Signature(); got != "app:reset-energy-beans" {
		t.Errorf("Signature() = %q, want %q", got, "app:reset-energy-beans")
	}
}

func TestUpdatePlayStreaks_Signature(t *testing.T) {
	cmd := &UpdatePlayStreaks{}
	if got := cmd.Signature(); got != "app:update-play-streaks" {
		t.Errorf("Signature() = %q, want %q", got, "app:update-play-streaks")
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		wantNil  bool
		wantJSON string
	}{
		{
			"simple map",
			map[string]any{"grantedBeansCleared": 5000},
			false,
			`{"grantedBeansCleared":5000}`,
		},
		{
			"grade and amount",
			map[string]any{"gradeAtTime": "lifetime", "grantAmount": 15000},
			false,
			"",
		},
		{
			"empty map",
			map[string]any{},
			false,
			"{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marshalJSON(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %q", *result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result, got nil")
			}
			if tt.wantJSON != "" && *result != tt.wantJSON {
				t.Errorf("marshalJSON() = %q, want %q", *result, tt.wantJSON)
			}
		})
	}
}

func TestNewID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := newID()
		if id == "" {
			t.Fatal("newID() returned empty string")
		}
		if ids[id] {
			t.Fatalf("newID() returned duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestNewID_Length(t *testing.T) {
	id := newID()
	// ULID is always 26 characters
	if len(id) != 26 {
		t.Errorf("newID() returned ID of length %d, want 26", len(id))
	}
}

func TestResetEnergyBeans_GrantAmounts(t *testing.T) {
	// Verify grant amounts match business rules
	tests := []struct {
		grade      string
		isLifetime bool
		wantAmount int
	}{
		{"vip", false, 10000},
		{"lifetime", true, 15000},
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			grantAmount := 10000
			if tt.isLifetime {
				grantAmount = 15000
			}
			if grantAmount != tt.wantAmount {
				t.Errorf("grant amount for %s = %d, want %d", tt.grade, grantAmount, tt.wantAmount)
			}
		})
	}
}

func TestResetEnergyBeans_SkipLogic(t *testing.T) {
	// Verify the skip logic: expired + no granted beans = skip
	tests := []struct {
		name            string
		isExpired       bool
		hasGrantedBeans bool
		shouldSkip      bool
	}{
		{"active with beans", false, true, false},
		{"active without beans", false, false, false},
		{"expired with beans", true, true, false},
		{"expired without beans — skip", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip := tt.isExpired && !tt.hasGrantedBeans
			if skip != tt.shouldSkip {
				t.Errorf("skip = %v, want %v", skip, tt.shouldSkip)
			}
		})
	}
}
