package constants

// Game degree values.
const (
	GameDegreePractice     = "practice"
	GameDegreeBeginner     = "beginner"
	GameDegreeIntermediate = "intermediate"
	GameDegreeAdvanced     = "advanced"
)

// GameDegreeLabels maps each degree to its Chinese label.
var GameDegreeLabels = map[string]string{
	GameDegreePractice:     "练习",
	GameDegreeBeginner:     "初级",
	GameDegreeIntermediate: "中级",
	GameDegreeAdvanced:     "高级",
}

// DegreeContentTypes maps each degree to its allowed content types.
// A nil slice means all content types are allowed.
var DegreeContentTypes = map[string][]string{
	GameDegreePractice:     nil,
	GameDegreeBeginner:     nil,
	GameDegreeIntermediate: {"block", "phrase", "sentence"},
	GameDegreeAdvanced:     {"sentence"},
}
