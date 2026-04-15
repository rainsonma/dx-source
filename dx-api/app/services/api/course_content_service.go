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

// metaDedupKey is the identity tuple used for content_metas reuse.
// Translation is normalized: a nil/empty translation collapses to "".
type metaDedupKey struct {
	SourceType  string
	SourceData  string
	Translation string
}

func makeMetaDedupKey(e MetadataEntry) metaDedupKey {
	t := ""
	if e.Translation != nil {
		t = *e.Translation
	}
	return metaDedupKey{e.SourceType, e.SourceData, t}
}

// existingMetaRef is a content_metas row already owned by the user that
// can be reused on save.
type existingMetaRef struct {
	ID          string `gorm:"column:id"`
	SourceType  string `gorm:"column:source_type"`
	SourceData  string `gorm:"column:source_data"`
	Translation string `gorm:"column:translation"` // COALESCE'd to ""
	IsBreakDone bool   `gorm:"column:is_break_done"`
}

// findExistingMetasForBatch loads, in a single query, all content_metas rows
// owned by userID that match any (source_type, source_data) pair in the
// batch. Returns a map keyed on metaDedupKey; first match wins per key.
func findExistingMetasForBatch(userID string, entries []MetadataEntry) (map[metaDedupKey]existingMetaRef, error) {
	if len(entries) == 0 {
		return map[metaDedupKey]existingMetaRef{}, nil
	}

	typeSet := map[string]struct{}{}
	dataSet := map[string]struct{}{}
	for _, e := range entries {
		typeSet[e.SourceType] = struct{}{}
		dataSet[e.SourceData] = struct{}{}
	}
	sourceTypes := make([]string, 0, len(typeSet))
	for t := range typeSet {
		sourceTypes = append(sourceTypes, t)
	}
	sourceData := make([]string, 0, len(dataSet))
	for d := range dataSet {
		sourceData = append(sourceData, d)
	}

	var rows []existingMetaRef
	if err := facades.Orm().Query().Raw(
		`SELECT DISTINCT cm.id, cm.source_type, cm.source_data,
		        COALESCE(cm.translation, '') AS translation, cm.is_break_done
		   FROM content_metas cm
		   JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL
		   JOIN game_levels gl ON gl.id = gm.game_level_id AND gl.deleted_at IS NULL
		   JOIN games g ON g.id = gl.game_id AND g.deleted_at IS NULL
		  WHERE cm.deleted_at IS NULL
		    AND g.user_id = ?
		    AND cm.source_type IN ?
		    AND cm.source_data IN ?`,
		userID, sourceTypes, sourceData,
	).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query existing metas for dedup: %w", err)
	}

	out := make(map[metaDedupKey]existingMetaRef, len(rows))
	for _, r := range rows {
		key := metaDedupKey{r.SourceType, r.SourceData, r.Translation}
		if _, exists := out[key]; !exists {
			out[key] = r
		}
	}
	return out, nil
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

	// Check existing capacity by joining game_metas → content_metas for this level.
	// We need the source type (from content_metas) plus the junction order
	// (from game_metas) for the new meta order calculation below.
	type existingMetaRow struct {
		SourceType string  `gorm:"column:source_type"`
		GmOrder    float64 `gorm:"column:gm_order"`
	}
	var existing []existingMetaRow
	if err := facades.Orm().Query().Raw(
		`SELECT content_metas.source_type, gm."order" AS gm_order
		 FROM content_metas
		 JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL
		 WHERE gm.game_level_id = ? AND content_metas.deleted_at IS NULL`,
		gameLevelID,
	).Scan(&existing); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		// Vocab modes: flat limit of MaxMetasPerLevel
		if len(existing)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
	} else {
		// Word-sentence mode: existing ratio formula
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

	// Get max junction order for auto-increment of new metas.
	maxOrder := float64(0)
	for _, m := range existing {
		if m.GmOrder > maxOrder {
			maxOrder = m.GmOrder
		}
	}

	// Create metas in batch — dedup against the user's existing content.
	existingByKey, err := findExistingMetasForBatch(userID, entries)
	if err != nil {
		return 0, err
	}

	// State carried across the entry loop for items reuse.
	itemsByMetaCache := map[string][]string{} // metaID -> ordered content_item IDs
	var maxItemOrderInLevel *float64
	itemsAddedSoFar := 0

	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		for i, e := range entries {
			key := makeMetaDedupKey(e)

			var metaID string
			var isBreakDone bool
			if existing, ok := existingByKey[key]; ok {
				metaID = existing.ID
				isBreakDone = existing.IsBreakDone
			} else {
				metaID = uuid.Must(uuid.NewV7()).String()
				meta := models.ContentMeta{
					ID:          metaID,
					SourceFrom:  sourceFrom,
					SourceType:  e.SourceType,
					SourceData:  e.SourceData,
					Translation: e.Translation,
					IsBreakDone: false,
				}
				if err := tx.Create(&meta); err != nil {
					return fmt.Errorf("failed to create content meta: %w", err)
				}
				// Add to map so subsequent within-batch identical entries reuse this row.
				existingByKey[key] = existingMetaRef{
					ID:          metaID,
					SourceType:  e.SourceType,
					SourceData:  e.SourceData,
					Translation: key.Translation,
					IsBreakDone: false,
				}
				isBreakDone = false
			}

			// Always create a fresh game_metas junction row (allows in-level repetition).
			gm := models.GameMeta{
				ID:            uuid.Must(uuid.NewV7()).String(),
				GameID:        level.GameID,
				GameLevelID:   gameLevelID,
				ContentMetaID: metaID,
				Order:         maxOrder + float64((i+1)*1000),
			}
			if err := tx.Create(&gm); err != nil {
				return fmt.Errorf("failed to create game meta: %w", err)
			}

			// If we are reusing a meta that has already been broken down,
			// also create game_items rows in this level pointing at the
			// existing content_items.
			if isBreakDone {
				if err := reuseItemsIntoLevel(
					tx, metaID, gameLevelID, level.GameID,
					itemsByMetaCache, &maxItemOrderInLevel, &itemsAddedSoFar,
				); err != nil {
					return err
				}
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

	// Verify meta belongs to this game
	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`UPDATE game_metas SET "order" = ?
		   WHERE game_level_id = ? AND content_meta_id = ? AND deleted_at IS NULL`,
		newOrder, gameLevelID, metaID,
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

	// Check item limit per meta (via junction so the count is level-scoped).
	itemCount, err2 := facades.Orm().Query().Model(&models.ContentItem{}).
		Select("content_items.id").
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
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
	}

	if err := facades.Orm().Query().Create(&item); err != nil {
		return nil, fmt.Errorf("failed to create content item: %w", err)
	}

	gi := models.GameItem{
		ID:            uuid.Must(uuid.NewV7()).String(),
		GameID:        level.GameID,
		GameLevelID:   gameLevelID,
		ContentItemID: item.ID,
		Order:         order,
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
		`UPDATE game_items SET "order" = ?
		   WHERE game_level_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		newOrder, gameLevelID, itemID,
	); err != nil {
		return fmt.Errorf("failed to reorder content item: %w", err)
	}

	return nil
}

// DeleteContentItem removes a content item from one level. With reuse enabled,
// only the level's game_items junction row is soft-deleted; the underlying
// content_item is soft-deleted only when no other junction references it.
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

	// Load the underlying content_item up front so we know its content_meta_id
	// for the is_break_done reset below.
	var item models.ContentItem
	if err := facades.Orm().Query().Where("id", itemID).First(&item); err != nil {
		return fmt.Errorf("failed to load content item: %w", err)
	}
	if item.ID == "" {
		return ErrContentItemNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// 1. Soft-delete this level's game_items rows for this item (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_items SET deleted_at = NOW()
			  WHERE content_item_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			itemID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_item: %w", err)
		}

		// 2. Count live game_items across all levels for this content_item.
		n, err := tx.Table("game_items").
			Where("content_item_id", itemID).
			Where("deleted_at IS NULL").
			Count()
		if err != nil {
			return fmt.Errorf("failed to count game_items: %w", err)
		}
		if n == 0 {
			if _, err := tx.Exec(
				`UPDATE content_items SET deleted_at = NOW()
				  WHERE id = ? AND deleted_at IS NULL`,
				itemID,
			); err != nil {
				return fmt.Errorf("failed to soft-delete content_item: %w", err)
			}
		}

		// 3. Reset is_break_done if this LEVEL has no remaining game_items
		//    for the meta. (Existing per-level logic, preserved.)
		if item.ContentMetaID != nil {
			if _, err := tx.Exec(
				`UPDATE content_metas SET is_break_done = false
				  WHERE id = ?
				    AND deleted_at IS NULL
				    AND NOT EXISTS (
				      SELECT 1 FROM game_items gi
				      JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
				      WHERE ci.content_meta_id = content_metas.id
				        AND gi.game_level_id = ?
				        AND gi.deleted_at IS NULL
				    )`,
				*item.ContentMetaID, gameLevelID,
			); err != nil {
				return fmt.Errorf("failed to reset meta break status: %w", err)
			}
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

	// Delete content in transaction. Pre-reuse, game_items/game_metas are 1:1
	// with content_items/content_metas, so we drive the content cascade off the
	// junction rows that point at this level.
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (
			     SELECT content_item_id FROM game_items
			     WHERE game_level_id = ? AND deleted_at IS NULL
			   )`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (
			     SELECT content_meta_id FROM game_metas
			     WHERE game_level_id = ? AND deleted_at IS NULL
			   )`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content metas: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE game_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_items: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE game_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_metas: %w", err)
		}
		return nil
	})
}

// DeleteMetadata removes a metadata entry from one level. With reuse enabled,
// only the level's junction row(s) are soft-deleted; the underlying
// content_metas / content_items rows are soft-deleted only when no other
// junction references them.
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
		// 1. Collect content_item_ids referenced by this level for this meta.
		var itemRows []struct {
			ContentItemID string `gorm:"column:content_item_id"`
		}
		if err := tx.Raw(
			`SELECT gi.content_item_id
			   FROM game_items gi
			   JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
			  WHERE ci.content_meta_id = ?
			    AND gi.game_level_id = ?
			    AND gi.deleted_at IS NULL`,
			metaID, gameLevelID,
		).Scan(&itemRows); err != nil {
			return fmt.Errorf("failed to collect items for delete: %w", err)
		}

		// 2. Soft-delete the level-scoped game_items rows (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_items SET deleted_at = NOW()
			  WHERE game_level_id = ?
			    AND content_item_id IN (
			      SELECT id FROM content_items WHERE content_meta_id = ?
			    )
			    AND deleted_at IS NULL`,
			gameLevelID, metaID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_items: %w", err)
		}

		// 3. Soft-delete the level-scoped game_metas rows (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_metas SET deleted_at = NOW()
			  WHERE content_meta_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_metas: %w", err)
		}

		// 4. For each orphaned content_item_id, count remaining live junctions
		//    across ALL levels; if 0, soft-delete the content_item.
		seen := map[string]struct{}{}
		for _, r := range itemRows {
			if _, ok := seen[r.ContentItemID]; ok {
				continue
			}
			seen[r.ContentItemID] = struct{}{}
			n, err := tx.Table("game_items").
				Where("content_item_id", r.ContentItemID).
				Where("deleted_at IS NULL").
				Count()
			if err != nil {
				return fmt.Errorf("failed to count game_items: %w", err)
			}
			if n == 0 {
				if _, err := tx.Exec(
					`UPDATE content_items SET deleted_at = NOW()
					  WHERE id = ? AND deleted_at IS NULL`,
					r.ContentItemID,
				); err != nil {
					return fmt.Errorf("failed to soft-delete content_item: %w", err)
				}
			}
		}

		// 5. Count remaining live game_metas for this content_meta across ALL levels.
		n, err := tx.Table("game_metas").
			Where("content_meta_id", metaID).
			Where("deleted_at IS NULL").
			Count()
		if err != nil {
			return fmt.Errorf("failed to count game_metas: %w", err)
		}
		if n == 0 {
			if _, err := tx.Exec(
				`UPDATE content_metas SET deleted_at = NOW()
				  WHERE id = ? AND deleted_at IS NULL`,
				metaID,
			); err != nil {
				return fmt.Errorf("failed to soft-delete content_meta: %w", err)
			}
		}
		return nil
	})
}

// verifyMetaBelongsToGame checks that a content meta belongs to a game.
func verifyMetaBelongsToGame(metaID, gameID string) error {
	n, err := facades.Orm().Query().Model(&models.GameMeta{}).
		Where("content_meta_id", metaID).
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

// reuseItemsIntoLevel creates game_items rows in gameLevelID for every active
// content_item belonging to metaID. Item IDs are loaded once per metaID via
// itemsByMetaCache. The level's pre-save max game_items.order is loaded once
// via maxItemOrderInLevel. itemsAddedSoFar is incremented for every new
// game_items row to keep ordering monotonically increasing across multiple
// reused metas in the same batch.
func reuseItemsIntoLevel(
	tx orm.Query,
	metaID, gameLevelID, gameID string,
	itemsByMetaCache map[string][]string,
	maxItemOrderInLevel **float64,
	itemsAddedSoFar *int,
) error {
	itemIDs, ok := itemsByMetaCache[metaID]
	if !ok {
		var rows []struct {
			ID string `gorm:"column:id"`
		}
		if err := tx.Raw(
			`SELECT id FROM content_items
			  WHERE content_meta_id = ? AND deleted_at IS NULL
			  ORDER BY id`,
			metaID,
		).Scan(&rows); err != nil {
			return fmt.Errorf("failed to load content_items for reuse: %w", err)
		}
		itemIDs = make([]string, 0, len(rows))
		for _, r := range rows {
			itemIDs = append(itemIDs, r.ID)
		}
		itemsByMetaCache[metaID] = itemIDs
	}
	if len(itemIDs) == 0 {
		return nil
	}

	if *maxItemOrderInLevel == nil {
		var row struct {
			MaxOrder float64 `gorm:"column:max_order"`
		}
		if err := tx.Raw(
			`SELECT COALESCE(MAX("order"), 0) AS max_order
			   FROM game_items
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		).Scan(&row); err != nil {
			return fmt.Errorf("failed to load max game_items order: %w", err)
		}
		v := row.MaxOrder
		*maxItemOrderInLevel = &v
	}

	for j, contentItemID := range itemIDs {
		gi := models.GameItem{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentItemID: contentItemID,
			Order:         **maxItemOrderInLevel + float64((*itemsAddedSoFar+j+1)*1000),
		}
		if err := tx.Create(&gi); err != nil {
			return fmt.Errorf("failed to create game item: %w", err)
		}
	}
	*itemsAddedSoFar += len(itemIDs)
	return nil
}
