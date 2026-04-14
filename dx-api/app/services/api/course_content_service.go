package api

import (
	"encoding/json"
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
	ID            string          `json:"id"`
	ContentMetaID *string         `json:"contentMetaId"`
	Content       string          `json:"content"`
	ContentType   string          `json:"contentType"`
	Translation   *string         `json:"translation"`
	Items         json.RawMessage `json:"items"`
	Order         float64         `json:"order"`
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
	if err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Select("content_metas.*").
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		// Vocab modes: flat limit of MaxMetasPerLevel
		if len(existingMetas)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
	} else {
		// Word-sentence mode: existing ratio formula
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

		if float64(totalSentences)/float64(MaxSentences)+float64(totalVocabs)/float64(MaxVocab) > 1 {
			return 0, ErrCapacityExceeded
		}
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
		order := maxOrder + float64((i+1)*1000)
		meta := models.ContentMeta{
			ID:          id,
			GameLevelID: gameLevelID,
			SourceFrom:  sourceFrom,
			SourceType:  e.SourceType,
			SourceData:  e.SourceData,
			Translation: e.Translation,
			IsBreakDone: false,
			Order:       order,
		}
		if err := facades.Orm().Query().Create(&meta); err != nil {
			return 0, fmt.Errorf("failed to create content meta: %w", err)
		}

		gm := models.GameMeta{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        level.GameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: meta.ID,
			Order:         order,
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
	if _, err := getCourseGameOwned(userID, gameID); err != nil {
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

	// 1. Load game_metas for this level in order.
	var gameMetas []models.GameMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order("\"order\" ASC").
		Get(&gameMetas); err != nil {
		return nil, fmt.Errorf("failed to load game_metas: %w", err)
	}
	if len(gameMetas) == 0 {
		return []LevelContentData{}, nil
	}

	// 2. Load the referenced content_metas by ID.
	metaIDs := make([]string, 0, len(gameMetas))
	for _, gm := range gameMetas {
		metaIDs = append(metaIDs, gm.ContentMetaID)
	}
	var contentMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("id IN ?", metaIDs).
		Get(&contentMetas); err != nil {
		return nil, fmt.Errorf("failed to load content_metas: %w", err)
	}
	metaByID := make(map[string]models.ContentMeta, len(contentMetas))
	for _, cm := range contentMetas {
		metaByID[cm.ID] = cm
	}

	// 3. Load game_items for this level in order.
	var gameItems []models.GameItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order("\"order\" ASC").
		Get(&gameItems); err != nil {
		return nil, fmt.Errorf("failed to load game_items: %w", err)
	}

	// 4. Load the referenced content_items by ID.
	var items []models.ContentItem
	if len(gameItems) > 0 {
		itemIDs := make([]string, 0, len(gameItems))
		for _, gi := range gameItems {
			itemIDs = append(itemIDs, gi.ContentItemID)
		}
		if err := facades.Orm().Query().
			Where("id IN ?", itemIDs).
			Get(&items); err != nil {
			return nil, fmt.Errorf("failed to load content_items: %w", err)
		}
	}
	itemByID := make(map[string]models.ContentItem, len(items))
	for _, it := range items {
		itemByID[it.ID] = it
	}

	// 5. Build items-by-meta map, preserving game_items.order as the authoritative
	//    per-level order.
	itemsByMeta := make(map[string][]CourseContentItemData)
	for _, gi := range gameItems {
		it, ok := itemByID[gi.ContentItemID]
		if !ok {
			continue
		}
		metaKey := ""
		if it.ContentMetaID != nil {
			metaKey = *it.ContentMetaID
		}
		raw := json.RawMessage("null")
		if it.Items != nil {
			raw = json.RawMessage(*it.Items)
		}
		itemsByMeta[metaKey] = append(itemsByMeta[metaKey], CourseContentItemData{
			ID:            it.ID,
			ContentMetaID: it.ContentMetaID,
			Content:       it.Content,
			ContentType:   it.ContentType,
			Translation:   it.Translation,
			Items:         raw,
			Order:         gi.Order, // from game_items (authoritative)
		})
	}

	// 6. Assemble result in game_metas.order.
	result := make([]LevelContentData, 0, len(gameMetas))
	for _, gm := range gameMetas {
		cm, ok := metaByID[gm.ContentMetaID]
		if !ok {
			continue
		}
		result = append(result, LevelContentData{
			Meta: ContentMetaData{
				ID:          cm.ID,
				SourceFrom:  cm.SourceFrom,
				SourceType:  cm.SourceType,
				SourceData:  cm.SourceData,
				Translation: cm.Translation,
				IsBreakDone: cm.IsBreakDone,
				Order:       gm.Order, // from game_metas (authoritative)
			},
			Items: itemsByMeta[cm.ID],
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
		Where("game_level_id", gameLevelID).
		Where("content_meta_id", contentMetaID).
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
		GameLevelID:   gameLevelID,
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
		GameID:        level.GameID,
		GameLevelID:   gameLevelID,
		ContentItemID: item.ID,
		Order:         item.Order,
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

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete content item
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
			itemID,
		); err != nil {
			return fmt.Errorf("failed to delete content item: %w", err)
		}

		// Reset is_break_done when meta has no remaining active items
		if _, err := tx.Exec(
			`UPDATE content_metas SET is_break_done = false
			 WHERE id = (SELECT content_meta_id FROM content_items WHERE id = ?)
			   AND deleted_at IS NULL
			   AND NOT EXISTS (
			     SELECT 1 FROM content_items
			     WHERE content_meta_id = content_metas.id
			       AND deleted_at IS NULL
			   )`,
			itemID,
		); err != nil {
			return fmt.Errorf("failed to reset meta break status: %w", err)
		}
		return nil
	})
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

	// Delete content in transaction
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content items: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content metas: %w", err)
		}
		return nil
	})
}

// DeleteMetadata removes a single metadata entry and its associated content items.
func DeleteMetadata(userID, gameID, metaID string) error {
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

	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete content items belonging to this meta
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE content_meta_id = ? AND deleted_at IS NULL",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete content items for meta: %w", err)
		}

		// Soft-delete the content meta
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete content meta: %w", err)
		}

		return nil
	})
}

// verifyMetaBelongsToGame checks that a content meta belongs to a game via its level.
func verifyMetaBelongsToGame(metaID, gameID string) error {
	count, err := facades.Orm().Query().Model(&models.GameMeta{}).
		Join("JOIN game_levels gl ON gl.id = game_metas.game_level_id AND gl.deleted_at IS NULL").
		Where("game_metas.content_meta_id", metaID).
		Where("gl.game_id", gameID).
		Count()
	if err != nil {
		return fmt.Errorf("failed to verify meta: %w", err)
	}
	if count == 0 {
		return ErrMetaNotFound
	}
	return nil
}

// verifyItemBelongsToGame checks that a content item belongs to a game via its level.
func verifyItemBelongsToGame(itemID, gameID string) error {
	count, err := facades.Orm().Query().Model(&models.GameItem{}).
		Join("JOIN game_levels gl ON gl.id = game_items.game_level_id AND gl.deleted_at IS NULL").
		Where("game_items.content_item_id", itemID).
		Where("gl.game_id", gameID).
		Count()
	if err != nil {
		return fmt.Errorf("failed to verify item: %w", err)
	}
	if count == 0 {
		return ErrContentItemNotFound
	}
	return nil
}

// calculateInsertionOrder computes the order for a new item relative to a reference item.
// referenceItemID is a content_item ID; lookups go through the game_items junction.
func calculateInsertionOrder(gameLevelID, referenceItemID, direction string) (float64, error) {
	if referenceItemID == "" {
		var lastItem models.GameItem
		if err := facades.Orm().Query().
			Where("game_level_id", gameLevelID).
			Order("\"order\" DESC").
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
		return lastItem.Order + 1000, nil
	}

	// Find reference item via the junction (referenceItemID is a content_item ID).
	var refItem models.GameItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("content_item_id", referenceItemID).
		First(&refItem); err != nil {
		return 0, fmt.Errorf("failed to find reference item: %w", err)
	}
	if refItem.ID == "" {
		return 0, ErrContentItemNotFound
	}

	var items []models.GameItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order("\"order\" ASC").
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}

	// Find reference index
	refIdx := -1
	for i, item := range items {
		if item.ContentItemID == referenceItemID {
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
