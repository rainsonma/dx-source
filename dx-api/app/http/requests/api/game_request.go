package api

// ListGamesRequest holds query parameters for listing published games.
type ListGamesRequest struct {
	Cursor      string   `form:"cursor" json:"cursor"`
	Limit       int      `form:"limit" json:"limit"`
	CategoryIDs []string `form:"categoryIds" json:"categoryIds"`
	PressID     string   `form:"pressId" json:"pressId"`
	Mode        string   `form:"mode" json:"mode"`
}

// SearchGamesRequest holds query parameters for searching games.
type SearchGamesRequest struct {
	Query string `form:"q" json:"q"`
	Limit int    `form:"limit" json:"limit"`
}

// LevelContentRequest holds query parameters for fetching level content.
type LevelContentRequest struct {
	Degree string `form:"degree" json:"degree"`
}
