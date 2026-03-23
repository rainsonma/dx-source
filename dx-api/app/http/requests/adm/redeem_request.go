package adm

import "github.com/goravel/framework/contracts/http"

// Count is string because Goravel validation operates on raw parsed data
// where JSON numbers become float64 — the in rule cannot compare float64.
type GenerateCodesRequest struct {
	Grade string `form:"grade" json:"grade"`
	Count string `form:"count" json:"count"`
}

func (r *GenerateCodesRequest) Authorize(ctx http.Context) error { return nil }
func (r *GenerateCodesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"grade": "required|in:month,season,year,lifetime",
		"count": "required|in:10,50,100,500",
	}
}
func (r *GenerateCodesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"grade.required": "Grade is required",
		"grade.in":       "Grade must be one of: month, season, year, lifetime",
		"count.required": "Count is required",
		"count.in":       "Count must be one of: 10, 50, 100, 500",
	}
}
