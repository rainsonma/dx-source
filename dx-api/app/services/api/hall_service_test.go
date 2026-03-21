package api

import (
	"testing"
	"time"
)

func TestMondayOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantDay  time.Weekday
		wantHour int
	}{
		{
			"Monday stays Monday",
			time.Date(2026, 3, 16, 15, 30, 0, 0, time.UTC), // Monday
			time.Monday,
			0,
		},
		{
			"Tuesday goes to Monday",
			time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC), // Tuesday
			time.Monday,
			0,
		},
		{
			"Wednesday goes to Monday",
			time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC), // Wednesday
			time.Monday,
			0,
		},
		{
			"Sunday goes to Monday of same week",
			time.Date(2026, 3, 22, 23, 59, 0, 0, time.UTC), // Sunday
			time.Monday,
			0,
		},
		{
			"Saturday goes to Monday",
			time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC), // Saturday
			time.Monday,
			0,
		},
		{
			"Friday goes to Monday",
			time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC), // Friday
			time.Monday,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mondayOfWeek(tt.input)

			if result.Weekday() != tt.wantDay {
				t.Errorf("weekday = %s, want %s", result.Weekday(), tt.wantDay)
			}
			if result.Hour() != tt.wantHour || result.Minute() != 0 || result.Second() != 0 {
				t.Errorf("time = %s, want 00:00:00", result.Format("15:04:05"))
			}
		})
	}
}

func TestMondayOfWeek_SameWeekConsistency(t *testing.T) {
	// All days in the same week should return the same Monday
	monday := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		result := mondayOfWeek(day)
		expected := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)

		if !result.Equal(expected) {
			t.Errorf("day %s: mondayOfWeek = %s, want %s",
				day.Weekday(), result.Format("2006-01-02"), expected.Format("2006-01-02"))
		}
	}
}

func TestMondayOfWeek_PreservesLocation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	input := time.Date(2026, 3, 18, 10, 0, 0, 0, loc)
	result := mondayOfWeek(input)

	if result.Location() != loc {
		t.Errorf("location = %v, want %v", result.Location(), loc)
	}
}

func TestMondayOfWeek_YearBoundary(t *testing.T) {
	// Jan 1, 2026 is a Thursday — Monday should be Dec 29, 2025
	input := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	result := mondayOfWeek(input)

	expected := time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("year boundary: got %s, want %s",
			result.Format("2006-01-02"), expected.Format("2006-01-02"))
	}
}
