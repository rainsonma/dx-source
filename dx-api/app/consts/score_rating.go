package consts

// ScoreRating holds the label for a score accuracy range.
type ScoreRating struct {
	Label string
}

// scoreRatingEntry is used internally for lookup.
type scoreRatingEntry struct {
	MinAccuracy float64
	Label       string
}

// scoreRatings defines rating tiers in descending order of accuracy.
var scoreRatings = []scoreRatingEntry{
	{MinAccuracy: 0.9, Label: "优秀"},
	{MinAccuracy: 0.7, Label: "良好"},
	{MinAccuracy: 0.6, Label: "及格"},
	{MinAccuracy: 0, Label: "继续加油"},
}

// GetScoreRating returns the score rating label based on accuracy (0-1).
func GetScoreRating(accuracy float64) ScoreRating {
	for _, r := range scoreRatings {
		if accuracy >= r.MinAccuracy {
			return ScoreRating{Label: r.Label}
		}
	}
	return ScoreRating{Label: "继续加油"}
}
