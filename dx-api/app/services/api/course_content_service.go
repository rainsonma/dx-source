package api

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

// MetadataEntry represents a single entry in a batch metadata creation request.
type MetadataEntry struct {
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	SourceType  string  `json:"sourceType"`
}

// CourseContentItemData represents a content item returned to the client.
type CourseContentItemData struct {
	ID            string  `json:"id"`
	ContentMetaID *string `json:"contentMetaId"`
	Content       string  `json:"content"`
	ContentType   string  `json:"contentType"`
	Translation   *string `json:"translation"`
	Items         *string `json:"items"`
	Order         float64 `json:"order"`
}

// LevelContentData groups content items by their metadata.
type LevelContentData struct {
	Meta  ContentMetaData         `json:"meta"`
	Items []CourseContentItemData `json:"items"`
}

// ContentMetaData represents content metadata returned to the client.
type ContentMetaData struct {
	ID          string  `json:"id"`
	SourceFrom  string  `json:"sourceFrom"`
	SourceType  string  `json:"sourceType"`
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	IsBreakDone bool    `json:"isBreakDone"`
	Order       float64 `json:"order"`
}

// SaveMetadataBatch creates content metadata entries in batch with capacity validation.
func SaveMetadataBatch(userID, gameID, gameLevelID string, entries []MetadataEntry, sourceFrom string) (int, error) {
	if err := requireVip(userID); err != nil {
		return 0, err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return 0, err
	}

	if game.Status == consts.GameStatusPublished {
		return 0, ErrGamePublished
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return 0, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return 0, ErrLevelNotFound
	}

	// Check existing capacity
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	existingSentences := 0
	existingVocabs := 0
	for _, m := range existingMetas {
		switch m.SourceType {
		case SourceTypeSentence:
			existingSentences++
		case SourceTypeVocab:
			existingVocabs++
		}
	}

	newSentences := 0
	newVocabs := 0
	for _, e := range entries {
		switch e.SourceType {
		case SourceTypeSentence:
			newSentences++
		case SourceTypeVocab:
			newVocabs++
		}
	}

	totalSentences := existingSentences + newSentences
	totalVocabs := existingVocabs + newVocabs

	// Check capacity using the same formula as the frontend
	if float64(totalSentences)/float64(MaxSentences)+float64(totalVocabs)/float64(MaxVocab) > 1 {
		return 0, ErrCapacityExceeded
	}

	// Get max order for auto-increment
	maxOrder := float64(0)
	if len(existingMetas) > 0 {
		for _, m := range existingMetas {
			if m.Order > maxOrder {
				maxOrder = m.Order
			}
		}
	}

	// Create metas in batch
	for i, e := range entries {
		id := uuid.Must(uuid.NewV7()).String()
		meta := models.ContentMeta{
			ID:          id,
			SourceFrom:  sourceFrom,
			SourceType:  e.SourceType,
			SourceData:  e.SourceData,
			Translation: e.Translation,
			IsBreakDone: false,
			Order:       maxOrder + float64((i+1)*1000),
		}
		if err := facades.Orm().Query().Create(&meta); err != nil {
			return 0, fmt.Errorf("failed to create content meta: %w", err)
		}

		gm := models.GameMeta{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: id,
		}
		if err := facades.Orm().Query().Create(&gm); err != nil {
			return 0, fmt.Errorf("failed to create game meta: %w", err)
		}
	}

	return len(entries), nil
}

// ReorderMetadata updates the order of a content metadata entry.
func ReorderMetadata(userID, gameID, metaID string, newOrder float64) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	// Verify meta belongs to this game
	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Model(&models.ContentMeta{}).Where("id", metaID).Update("order", newOrder); err != nil {
		return fmt.Errorf("failed to reorder metadata: %w", err)
	}

	return nil
}

// GetContentItemsByMeta returns content items grouped by their metadata for a given level.
func GetContentItemsByMeta(userID, gameID, gameLevelID string) ([]LevelContentData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	_, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return nil, err
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	// Load metas ordered
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		return nil, fmt.Errorf("failed to load metas: %w", err)
	}

	if len(metas) == 0 {
		return []LevelContentData{}, nil
	}

	// Load all items for this level
	var items []models.ContentItem
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load items: %w", err)
	}

	// Group items by meta ID
	itemsByMeta := make(map[string][]CourseContentItemData)
	for _, item := range items {
		metaID := ""
		if item.ContentMetaID != nil {
			metaID = *item.ContentMetaID
		}
		itemsByMeta[metaID] = append(itemsByMeta[metaID], CourseContentItemData{
			ID:            item.ID,
			ContentMetaID: item.ContentMetaID,
			Content:       item.Content,
			ContentType:   item.ContentType,
			Translation:   item.Translation,
			Items:         item.Items,
			Order:         item.Order,
		})
	}

	// Build grouped result
	result := make([]LevelContentData, 0, len(metas))
	for _, meta := range metas {
		metaData := ContentMetaData{
			ID:          meta.ID,
			SourceFrom:  meta.SourceFrom,
			SourceType:  meta.SourceType,
			SourceData:  meta.SourceData,
			Translation: meta.Translation,
			IsBreakDone: meta.IsBreakDone,
			Order:       meta.Order,
		}

		metaItems := itemsByMeta[meta.ID]
		if metaItems == nil {
			metaItems = []CourseContentItemData{}
		}

		result = append(result, LevelContentData{
			Meta:  metaData,
			Items: metaItems,
		})
	}

	return result, nil
}

// InsertContentItem inserts a content item at a calculated position.
func InsertContentItem(userID, gameID, gameLevelID, contentMetaID string, content, contentType string, translation *string, referenceItemID, direction string) (*CourseContentItemData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status == consts.GameStatusPublished {
		return nil, ErrGamePublished
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	// Verify meta belongs to game
	if err := verifyMetaBelongsToGame(contentMetaID, gameID); err != nil {
		return nil, err
	}

	// Verify reference item belongs to game
	if referenceItemID != "" {
		if err := verifyItemBelongsToGame(referenceItemID, gameID); err != nil {
			return nil, err
		}
	}

	// Check item limit per meta
	itemCount, err2 := facades.Orm().Query().Model(&models.ContentItem{}).
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.content_meta_id", contentMetaID).
		Count()
	if err2 != nil {
		return nil, fmt.Errorf("failed to count items: %w", err2)
	}
	if itemCount >= int64(MaxItemsPerMeta) {
		return nil, ErrItemLimitExceeded
	}

	// Calculate insertion order
	order, err := calculateInsertionOrder(gameLevelID, referenceItemID, direction)
	if err != nil {
		return nil, err
	}

	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:            id,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Order:         order,
		IsActive:      true,
	}

	if err := facades.Orm().Query().Create(&item); err != nil {
		return nil, fmt.Errorf("failed to create content item: %w", err)
	}

	gi := models.GameItem{
		ID:            uuid.Must(uuid.NewV7()).String(),
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentItemID: id,
	}
	if err := facades.Orm().Query().Create(&gi); err != nil {
		return nil, fmt.Errorf("failed to create game item: %w", err)
	}

	return &CourseContentItemData{
		ID:            id,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Items:         nil,
		Order:         order,
	}, nil
}

// UpdateContentItemText updates the text and translation of a content item.
func UpdateContentItemText(userID, gameID, itemID, content string, translation *string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	if err := verifyItemBelongsToGame(itemID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Model(&models.ContentItem{}).Where("id", itemID).Update(map[string]any{
		"content":     content,
		"translation": translation,
	}); err != nil {
		return fmt.Errorf("failed to update content item: %w", err)
	}

	return nil
}

// ReorderContentItems updates the order of a content item.
func ReorderContentItems(userID, gameID, itemID string, newOrder float64) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	if err := verifyItemBelongsToGame(itemID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Model(&models.ContentItem{}).Where("id", itemID).Update("order", newOrder); err != nil {
		return fmt.Errorf("failed to reorder content item: %w", err)
	}

	return nil
}

// DeleteContentItem removes a single content item.
func DeleteContentItem(userID, gameID, itemID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	if err := verifyItemBelongsToGame(itemID, gameID); err != nil {
		return err
	}

	// Delete junction row
	if _, err := facades.Orm().Query().
		Where("content_item_id", itemID).Where("game_id", gameID).
		Delete(&models.GameItem{}); err != nil {
		return fmt.Errorf("failed to delete game item: %w", err)
	}

	// Delete content item if no other games reference it
	var remaining int64
	remaining, _ = facades.Orm().Query().Model(&models.GameItem{}).
		Where("content_item_id", itemID).Count()
	if remaining == 0 {
		if _, err := facades.Orm().Query().Where("id", itemID).Delete(&models.ContentItem{}); err != nil {
			return fmt.Errorf("failed to delete content item: %w", err)
		}
	}

	return nil
}

// DeleteAllLevelContent removes all content items and metas from a level.
func DeleteAllLevelContent(userID, gameID, gameLevelID string) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	// Delete junction rows then orphaned content in transaction
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Delete junction rows for this level
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game items: %w", err)
		}
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
			return fmt.Errorf("failed to delete game metas: %w", err)
		}

		// Delete orphaned content (not referenced by any game)
		if _, err := tx.Exec(
			"DELETE FROM content_items WHERE id NOT IN (SELECT content_item_id FROM game_items)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec(
			"DELETE FROM content_metas WHERE id NOT IN (SELECT content_meta_id FROM game_metas)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}

		return nil
	})
}

// verifyMetaBelongsToGame checks that a content meta belongs to a game via its level.
func verifyMetaBelongsToGame(metaID, gameID string) error {
	var gm models.GameMeta
	if err := facades.Orm().Query().
		Where("content_meta_id", metaID).
		Where("game_id", gameID).
		First(&gm); err != nil || gm.ID == "" {
		return ErrMetaNotFound
	}
	return nil
}

// verifyItemBelongsToGame checks that a content item belongs to a game via its level.
func verifyItemBelongsToGame(itemID, gameID string) error {
	var gi models.GameItem
	if err := facades.Orm().Query().
		Where("content_item_id", itemID).
		Where("game_id", gameID).
		First(&gi); err != nil || gi.ID == "" {
		return ErrContentItemNotFound
	}
	return nil
}

// calculateInsertionOrder computes the order for a new item relative to a reference item.
func calculateInsertionOrder(gameLevelID, referenceItemID, direction string) (float64, error) {
	if referenceItemID == "" {
		var lastItem models.ContentItem
		if err := facades.Orm().Query().
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
			Where("gi.game_level_id", gameLevelID).
			Order("content_items.\"order\" DESC").
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
		return lastItem.Order + 1000, nil
	}

	// Find reference item
	var refItem models.ContentItem
	if err := facades.Orm().Query().Where("id", referenceItemID).First(&refItem); err != nil {
		return 0, fmt.Errorf("failed to find reference item: %w", err)
	}
	if refItem.ID == "" {
		return 0, ErrContentItemNotFound
	}

	var items []models.ContentItem
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}

	// Find reference index
	refIdx := -1
	for i, item := range items {
		if item.ID == referenceItemID {
			refIdx = i
			break
		}
	}
	if refIdx == -1 {
		return refItem.Order + 1000, nil
	}

	if direction == "above" || direction == "before" {
		if refIdx == 0 {
			return refItem.Order / 2, nil
		}
		prevOrder := items[refIdx-1].Order
		return (prevOrder + refItem.Order) / 2, nil
	}

	// "after" (default)
	if refIdx == len(items)-1 {
		return refItem.Order + 1000, nil
	}
	nextOrder := items[refIdx+1].Order
	return (refItem.Order + nextOrder) / 2, nil
}
