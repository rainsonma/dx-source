package api

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/models"
)

// AddedGameVocab is one item in the AddVocabsToLevel response.
type AddedGameVocab struct {
	GameVocabID    string `json:"gameVocabId"`
	ContentVocabID string `json:"contentVocabId"`
	Content        string `json:"content"`
}

// LevelVocabData is one row in GetLevelVocabs.
type LevelVocabData struct {
	GameVocabID string            `json:"gameVocabId"`
	Order       float64           `json:"order"`
	Vocab       *ContentVocabData `json:"vocab"`
}

// AddVocabsToLevel places vocabs from the user's pool into a level.
// vocabIDs must all belong to userID; ownership is verified before insert.
func AddVocabsToLevel(userID, gameID, levelID string, vocabIDs []string) ([]AddedGameVocab, error) {
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
	if !consts.IsVocabMode(game.Mode) {
		return nil, ErrForbidden
	}

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", levelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	if len(vocabIDs) == 0 {
		return []AddedGameVocab{}, nil
	}

	// Ownership verification — load all referenced vocabs in one query.
	var vocabs []models.ContentVocab
	if err := facades.Orm().Query().Where("id IN ?", vocabIDs).Get(&vocabs); err != nil {
		return nil, fmt.Errorf("failed to load vocabs: %w", err)
	}
	vocabByID := make(map[string]models.ContentVocab, len(vocabs))
	for i := range vocabs {
		vocabByID[vocabs[i].ID] = vocabs[i]
	}
	for _, id := range vocabIDs {
		v, ok := vocabByID[id]
		if !ok {
			return nil, ErrContentItemNotFound
		}
		if v.UserID != userID {
			return nil, ErrForbidden
		}
	}

	// Capacity + batch-size (existing rules)
	existingCount, err := facades.Orm().Query().Model(&models.GameVocab{}).Where("game_level_id", levelID).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count existing vocabs: %w", err)
	}
	if existingCount+int64(len(vocabIDs)) > int64(consts.MaxMetasPerLevel) {
		return nil, ErrCapacityExceeded
	}
	batchSize := consts.VocabBatchSize(game.Mode)
	if batchSize > 0 && (existingCount+int64(len(vocabIDs)))%int64(batchSize) != 0 {
		return nil, ErrBatchSizeInvalid
	}

	// Find max order
	type ordRow struct {
		MaxOrder float64 `gorm:"column:max_order"`
	}
	var maxRow ordRow
	if err := facades.Orm().Query().Raw(
		`SELECT COALESCE(MAX("order"), 0) AS max_order FROM game_vocabs WHERE game_level_id = ? AND deleted_at IS NULL`,
		levelID,
	).Scan(&maxRow); err != nil {
		return nil, fmt.Errorf("failed to load max order: %w", err)
	}
	maxOrder := maxRow.MaxOrder

	added := make([]AddedGameVocab, 0, len(vocabIDs))
	for i, id := range vocabIDs {
		v := vocabByID[id]
		gv := models.GameVocab{
			ID:             uuid.Must(uuid.NewV7()).String(),
			GameID:         gameID,
			GameLevelID:    levelID,
			ContentVocabID: id,
			Order:          maxOrder + float64((i+1)*1000),
		}
		if err := facades.Orm().Query().Create(&gv); err != nil {
			return nil, fmt.Errorf("create game_vocab: %w", err)
		}
		added = append(added, AddedGameVocab{
			GameVocabID:    gv.ID,
			ContentVocabID: id,
			Content:        v.Content,
		})
	}

	return added, nil
}

// GetLevelVocabs returns all game_vocabs in a level joined with their canonical rows.
func GetLevelVocabs(userID, gameID, gameLevelID string) ([]LevelVocabData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if _, err := getCourseGameOwned(userID, gameID); err != nil {
		return nil, err
	}

	var gvs []models.GameVocab
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&gvs); err != nil {
		return nil, fmt.Errorf("failed to load game_vocabs: %w", err)
	}
	if len(gvs) == 0 {
		return []LevelVocabData{}, nil
	}

	vocabIDs := make([]string, 0, len(gvs))
	for _, gv := range gvs {
		vocabIDs = append(vocabIDs, gv.ContentVocabID)
	}
	var vocabs []models.ContentVocab
	if err := facades.Orm().Query().Where("id IN ?", vocabIDs).Get(&vocabs); err != nil {
		return nil, fmt.Errorf("failed to load content_vocabs: %w", err)
	}
	vocabByID := make(map[string]models.ContentVocab, len(vocabs))
	for _, v := range vocabs {
		vocabByID[v.ID] = v
	}

	result := make([]LevelVocabData, 0, len(gvs))
	for _, gv := range gvs {
		row := LevelVocabData{
			GameVocabID: gv.ID,
			Order:       gv.Order,
		}
		if v, ok := vocabByID[gv.ContentVocabID]; ok {
			row.Vocab = vocabToData(&v)
		}
		result = append(result, row)
	}
	return result, nil
}

// ReorderGameVocab updates the placement order.
func ReorderGameVocab(userID, gameID, gameVocabID string, newOrder float64) error {
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

	// Verify ownership: gv.game_id must match
	var gv models.GameVocab
	if err := facades.Orm().Query().Where("id", gameVocabID).First(&gv); err != nil || gv.ID == "" {
		return ErrContentItemNotFound
	}
	if gv.GameID != gameID {
		return ErrForbidden
	}
	if _, err := facades.Orm().Query().Exec(
		`UPDATE game_vocabs SET "order" = ?
		   WHERE id = ? AND deleted_at IS NULL`,
		newOrder, gameVocabID,
	); err != nil {
		return fmt.Errorf("failed to reorder game_vocab: %w", err)
	}
	return nil
}

// DeleteGameVocab soft-deletes a placement row only; canonical row stays.
func DeleteGameVocab(userID, gameID, gameVocabID string) error {
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
	var gv models.GameVocab
	if err := facades.Orm().Query().Where("id", gameVocabID).First(&gv); err != nil || gv.ID == "" {
		return ErrContentItemNotFound
	}
	if gv.GameID != gameID {
		return ErrForbidden
	}
	if _, err := facades.Orm().Query().Exec(
		`UPDATE game_vocabs SET deleted_at = NOW()
		   WHERE id = ? AND deleted_at IS NULL`,
		gameVocabID,
	); err != nil {
		return fmt.Errorf("failed to delete game_vocab: %w", err)
	}
	return nil
}
