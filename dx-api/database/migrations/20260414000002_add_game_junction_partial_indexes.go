package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000002AddGameJunctionPartialIndexes struct{}

func (r *M20260414000002AddGameJunctionPartialIndexes) Signature() string {
	return "20260414000002_add_game_junction_partial_indexes"
}

func (r *M20260414000002AddGameJunctionPartialIndexes) Up() error {
	indexes := []string{
		`CREATE UNIQUE INDEX idx_game_metas_level_meta_unique ON game_metas (game_level_id, content_meta_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_game_metas_level_order ON game_metas (game_level_id, deleted_at, "order")`,
		`CREATE UNIQUE INDEX idx_game_items_level_item_unique ON game_items (game_level_id, content_item_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_game_items_level_order ON game_items (game_level_id, deleted_at, "order")`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000002AddGameJunctionPartialIndexes) Down() error {
	indexes := []string{
		`DROP INDEX IF EXISTS idx_game_items_level_order`,
		`DROP INDEX IF EXISTS idx_game_items_level_item_unique`,
		`DROP INDEX IF EXISTS idx_game_metas_level_order`,
		`DROP INDEX IF EXISTS idx_game_metas_level_meta_unique`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
