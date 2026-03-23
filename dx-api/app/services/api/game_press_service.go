package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// PressData represents a game publisher.
type PressData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListPresses returns all game publishers ordered by their display order.
func ListPresses() ([]PressData, error) {
	var presses []models.GamePress
	if err := facades.Orm().Query().
		Order("\"order\" ASC").
		Get(&presses); err != nil {
		return nil, fmt.Errorf("failed to list presses: %w", err)
	}

	result := make([]PressData, 0, len(presses))
	for _, p := range presses {
		result = append(result, PressData{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return result, nil
}
