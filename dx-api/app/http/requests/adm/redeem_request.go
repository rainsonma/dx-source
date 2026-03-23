package adm

// GenerateCodesRequest validates redeem code generation data.
type GenerateCodesRequest struct {
	Grade string `form:"grade" json:"grade"`
	Count int    `form:"count" json:"count"`
}
