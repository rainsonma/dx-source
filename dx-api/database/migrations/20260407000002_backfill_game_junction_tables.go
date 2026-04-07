package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260407000002BackfillGameJunctionTables struct{}

func (r *M20260407000002BackfillGameJunctionTables) Signature() string {
	return "20260407000002_backfill_game_junction_tables"
}

func (r *M20260407000002BackfillGameJunctionTables) Up() error {
	// 1. Backfill game_metas from existing content_metas
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, created_at)
		SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm.created_at
		FROM content_metas cm
		JOIN game_levels gl ON gl.id = cm.game_level_id
		WHERE cm.game_level_id IS NOT NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// 2. Backfill game_items from existing content_items
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_items (id, game_id, game_level_id, content_item_id, created_at)
		SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci.created_at
		FROM content_items ci
		JOIN game_levels gl ON gl.id = ci.game_level_id
		WHERE ci.game_level_id IS NOT NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// 3. Make game_level_id nullable on content tables
	if _, err := facades.Orm().Query().Exec(
		"ALTER TABLE content_metas ALTER COLUMN game_level_id DROP NOT NULL",
	); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(
		"ALTER TABLE content_items ALTER COLUMN game_level_id DROP NOT NULL",
	); err != nil {
		return err
	}

	return nil
}

func (r *M20260407000002BackfillGameJunctionTables) Down() error {
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_metas ALTER COLUMN game_level_id SET NOT NULL",
	)
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_items ALTER COLUMN game_level_id SET NOT NULL",
	)
	_, _ = facades.Orm().Query().Exec("DELETE FROM game_items")
	_, _ = facades.Orm().Query().Exec("DELETE FROM game_metas")
	return nil
}
