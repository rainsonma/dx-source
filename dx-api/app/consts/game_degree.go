package consts

// Game degree values.
const (
	GameDegreeBeginner     = "beginner"
	GameDegreeIntermediate = "intermediate"
	GameDegreeAdvanced     = "advanced"
)

// GameDegreeLabels maps each degree to its Chinese label.
var GameDegreeLabels = map[string]string{
	GameDegreeBeginner:     "初级",
	GameDegreeIntermediate: "中级",
	GameDegreeAdvanced:     "高级",
}

// DegreeContentTypes maps each degree to its allowed content types.
// A nil slice means all content types are allowed.
var DegreeContentTypes = map[string][]string{
	GameDegreeBeginner:     nil,
	GameDegreeIntermediate: {"block", "phrase", "sentence"},
	GameDegreeAdvanced:     {"sentence"},
}
