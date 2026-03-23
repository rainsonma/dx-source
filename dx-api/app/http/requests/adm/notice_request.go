package adm

// CreateNoticeRequest validates notice creation data.
type CreateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

// UpdateNoticeRequest validates notice update data.
type UpdateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}
