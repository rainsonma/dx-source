package api

import (
	"encoding/json"
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
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
// No dedup: every entry becomes a fresh content_metas row. Reordering is by (game_level_id, order).
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

	// Capacity check + max order pulled from content_metas directly.
	type existingMetaRow struct {
		SourceType string  `gorm:"column:source_type"`
		MetaOrder  float64 `gorm:"column:meta_order"`
	}
	var existing []existingMetaRow
	if err := facades.Orm().Query().Raw(
		`SELECT source_type, "order" AS meta_order
		   FROM content_metas
		  WHERE game_level_id = ? AND deleted_at IS NULL`,
		gameLevelID,
	).Scan(&existing); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		// (NOTE: vocab modes use content_vocabs / game_vocabs in Phase 5; this
		// branch keeps the legacy capacity check for safety in case word-sentence
		// metadata is somehow still saved into a vocab game during migration.)
		if len(existing)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
		batchSize := consts.VocabBatchSize(game.Mode)
		if batchSize > 0 && len(entries)%batchSize != 0 {
			return 0, ErrBatchSizeInvalid
		}
	} else {
		existingSentences := 0
		existingVocabs := 0
		for _, m := range existing {
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

	maxOrder := float64(0)
	for _, m := range existing {
		if m.MetaOrder > maxOrder {
			maxOrder = m.MetaOrder
		}
	}

	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		for i, e := range entries {
			meta := models.ContentMeta{
				ID:          uuid.Must(uuid.NewV7()).String(),
				GameID:      gameID,
				GameLevelID: gameLevelID,
				SourceFrom:  sourceFrom,
				SourceType:  e.SourceType,
				SourceData:  e.SourceData,
				Translation: e.Translation,
				IsBreakDone: false,
				Order:       maxOrder + float64((i+1)*1000),
			}
			if err := tx.Create(&meta); err != nil {
				return fmt.Errorf("failed to create content meta: %w", err)
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return len(entries), nil
}

// ReorderMetadata updates the order of a content metadata entry within a level.
func ReorderMetadata(userID, gameID, gameLevelID, metaID string, newOrder float64) error {
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
	if _, err := facades.Orm().Query().Exec(
		`UPDATE content_metas SET "order" = ?
		   WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
		newOrder, metaID, gameLevelID,
	); err != nil {
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

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	var contentMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&contentMetas); err != nil {
		return nil, fmt.Errorf("failed to load content_metas: %w", err)
	}
	if len(contentMetas) == 0 {
		return []LevelContentData{}, nil
	}

	var items []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load content_items: %w", err)
	}

	itemsByMeta := make(map[string][]CourseContentItemData)
	for _, it := range items {
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
			Order:         it.Order,
		})
	}

	result := make([]LevelContentData, 0, len(contentMetas))
	for _, cm := range contentMetas {
		result = append(result, LevelContentData{
			Meta: ContentMetaData{
				ID:          cm.ID,
				SourceFrom:  cm.SourceFrom,
				SourceType:  cm.SourceType,
				SourceData:  cm.SourceData,
				Translation: cm.Translation,
				IsBreakDone: cm.IsBreakDone,
				Order:       cm.Order,
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

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	if err := verifyMetaBelongsToGame(contentMetaID, gameID); err != nil {
		return nil, err
	}
	if referenceItemID != "" {
		if err := verifyItemBelongsToGame(referenceItemID, gameID); err != nil {
			return nil, err
		}
	}

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

	order, err := calculateInsertionOrder(gameLevelID, referenceItemID, direction)
	if err != nil {
		return nil, err
	}

	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:            id,
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Order:         order,
	}
	if err := facades.Orm().Query().Create(&item); err != nil {
		return nil, fmt.Errorf("failed to create content item: %w", err)
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

// ReorderContentItems updates the order of a content item within a level.
func ReorderContentItems(userID, gameID, gameLevelID, itemID string, newOrder float64) error {
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
	if _, err := facades.Orm().Query().Exec(
		`UPDATE content_items SET "order" = ?
		   WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
		newOrder, itemID, gameLevelID,
	); err != nil {
		return fmt.Errorf("failed to reorder content item: %w", err)
	}
	return nil
}

// DeleteContentItem soft-deletes a single item.
func DeleteContentItem(userID, gameID, gameLevelID, itemID string) error {
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

	var item models.ContentItem
	if err := facades.Orm().Query().Where("id", itemID).First(&item); err != nil {
		return fmt.Errorf("failed to load content item: %w", err)
	}
	if item.ID == "" {
		return ErrContentItemNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
			itemID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_item: %w", err)
		}

		// Reset is_break_done if the meta has no remaining live items in this level.
		if item.ContentMetaID != nil {
			if _, err := tx.Exec(
				`UPDATE content_metas SET is_break_done = false
				  WHERE id = ?
				    AND deleted_at IS NULL
				    AND NOT EXISTS (
				      SELECT 1 FROM content_items
				       WHERE content_meta_id = content_metas.id
				         AND game_level_id = ?
				         AND deleted_at IS NULL
				    )`,
				*item.ContentMetaID, gameLevelID,
			); err != nil {
				return fmt.Errorf("failed to reset meta break status: %w", err)
			}
		}
		return nil
	})
}

// DeleteAllLevelContent soft-deletes every content_meta and content_item in a level.
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

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_metas: %w", err)
		}
		return nil
	})
}

// DeleteMetadata soft-deletes a meta plus all its items in this level.
func DeleteMetadata(userID, gameID, gameLevelID, metaID string) error {
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
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE content_meta_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			  WHERE id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_meta: %w", err)
		}
		return nil
	})
}

// verifyMetaBelongsToGame checks that a content meta belongs to a game.
func verifyMetaBelongsToGame(metaID, gameID string) error {
	n, err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Where("id", metaID).
		Where("game_id", gameID).
		Count()
	if err != nil {
		return fmt.Errorf("failed to verify meta: %w", err)
	}
	if n == 0 {
		return ErrMetaNotFound
	}
	return nil
}

// verifyItemBelongsToGame checks that a content item belongs to a game.
func verifyItemBelongsToGame(itemID, gameID string) error {
	count, err := facades.Orm().Query().Model(&models.ContentItem{}).
		Where("id", itemID).
		Where("game_id", gameID).
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
func calculateInsertionOrder(gameLevelID, referenceItemID, direction string) (float64, error) {
	if referenceItemID == "" {
		var lastItem models.ContentItem
		if err := facades.Orm().Query().
			Where("game_level_id", gameLevelID).
			Order(`"order" DESC`).
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
		return lastItem.Order + 1000, nil
	}

	var refItem models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("id", referenceItemID).
		First(&refItem); err != nil {
		return 0, fmt.Errorf("failed to find reference item: %w", err)
	}
	if refItem.ID == "" {
		return 0, ErrContentItemNotFound
	}

	var items []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}

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

	if refIdx == len(items)-1 {
		return refItem.Order + 1000, nil
	}
	nextOrder := items[refIdx+1].Order
	return (refItem.Order + nextOrder) / 2, nil
}
