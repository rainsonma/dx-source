package api

// SubmitContentSeekRequest validates a content seek submission.
type SubmitContentSeekRequest struct {
	CourseName  string `form:"course_name" json:"course_name"`
	Description string `form:"description" json:"description"`
	DiskUrl     string `form:"disk_url" json:"disk_url"`
}
