package api

import "github.com/goravel/framework/contracts/http"

// CreateGameRequest validates game creation data.
type CreateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *CreateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":           "required",
		"gameMode":       "required",
		"gameCategoryId": "required",
		"gamePressId":    "required",
	}
}

// UpdateGameRequest validates game update data.
type UpdateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *UpdateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

// CreateLevelRequest validates level creation data.
type CreateLevelRequest struct {
	Name        string  `form:"name" json:"name"`
	Description *string `form:"description" json:"description"`
}

func (r *CreateLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

// SaveMetadataBatchRequest validates batch metadata creation.
type SaveMetadataBatchRequest struct {
	GameLevelID string              `json:"gameLevelId"`
	SourceFrom  string              `json:"sourceFrom"`
	Entries     []MetadataEntryJSON `json:"entries"`
}

func (r *SaveMetadataBatchRequest) Authorize(ctx http.Context) error { return nil }
func (r *SaveMetadataBatchRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entries": "required|min_len:1",
	}
}

// MetadataEntryJSON represents a single entry in a batch metadata request.
type MetadataEntryJSON struct {
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	SourceType  string  `json:"sourceType"`
}

// ReorderMetadataRequest validates metadata reorder data.
type ReorderMetadataRequest struct {
	GameLevelID string  `json:"gameLevelId"`
	MetaID      string  `json:"metaId"`
	NewOrder    float64 `json:"newOrder"`
}

func (r *ReorderMetadataRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderMetadataRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"metaId": "required",
	}
}

// InsertContentItemRequest validates content item insertion data.
type InsertContentItemRequest struct {
	GameLevelID     string  `json:"gameLevelId"`
	ContentMetaID   string  `json:"contentMetaId"`
	Content         string  `json:"content"`
	ContentType     string  `json:"contentType"`
	Translation     *string `json:"translation"`
	ReferenceItemID string  `json:"referenceItemId"`
	Direction       string  `json:"direction"`
}

func (r *InsertContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *InsertContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"contentMetaId": "required",
	}
}

// UpdateContentItemTextRequest validates content item text update.
type UpdateContentItemTextRequest struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
}

// ReorderContentItemRequest validates content item reorder data.
type ReorderContentItemRequest struct {
	ItemID   string  `json:"itemId"`
	NewOrder float64 `json:"newOrder"`
}

func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"itemId": "required",
	}
}
