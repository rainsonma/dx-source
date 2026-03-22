package consts

import "time"

// ReviewIntervalDays defines the fixed spaced-repetition intervals in days,
// indexed by review count.
var ReviewIntervalDays = []int{1, 3, 7, 14, 30, 90}

// GetNextReviewAt computes the next review date from the current review count.
func GetNextReviewAt(reviewCount int) time.Time {
	index := reviewCount
	if index >= len(ReviewIntervalDays) {
		index = len(ReviewIntervalDays) - 1
	}
	days := ReviewIntervalDays[index]
	return time.Now().AddDate(0, 0, days)
}
