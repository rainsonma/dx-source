package constants

import (
	"testing"
	"time"
)

func TestGetNextReviewAt(t *testing.T) {
	tests := []struct {
		reviewCount  int
		expectedDays int
	}{
		{0, 1},
		{1, 3},
		{2, 7},
		{3, 14},
		{4, 30},
		{5, 90},
		{6, 90},  // clamped to last
		{99, 90}, // clamped to last
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			before := time.Now()
			result := GetNextReviewAt(tt.reviewCount)
			expected := before.AddDate(0, 0, tt.expectedDays)

			// Allow 1 second tolerance
			diff := result.Sub(expected)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("reviewCount=%d: got %v, expected ~%v (diff %v)",
					tt.reviewCount, result, expected, diff)
			}
		})
	}
}
