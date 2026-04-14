package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000002BackfillJunctionTables struct{}

func (r *M20260414000002BackfillJunctionTables) Signature() string {
	return "20260414000002_backfill_junction_tables"
}

func (r *M20260414000002BackfillJunctionTables) Up() error {
	// Backfill game_metas from content_metas. Uses INNER JOIN to game_levels so
	// any orphaned content_metas (whose game_level_id references a deleted level)
	// are skipped. ON CONFLICT DO NOTHING makes this safely re-runnable if the
	// migration aborts mid-way.
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, "order", created_at, updated_at)
		SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm."order", cm.created_at, cm.updated_at
		FROM content_metas cm
		JOIN game_levels gl ON gl.id = cm.game_level_id AND gl.deleted_at IS NULL
		WHERE cm.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// Backfill game_items from content_items (~1.22M rows).
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_items (id, game_id, game_level_id, content_item_id, "order", created_at, updated_at)
		SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci."order", ci.created_at, ci.updated_at
		FROM content_items ci
		JOIN game_levels gl ON gl.id = ci.game_level_id AND gl.deleted_at IS NULL
		WHERE ci.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	return nil
}

func (r *M20260414000002BackfillJunctionTables) Down() error {
	// Safe: the backfilled rows are the only data in these tables during this
	// migration step, so a full delete rolls back the entire step.
	if _, err := facades.Orm().Query().Exec(`DELETE FROM game_items`); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(`DELETE FROM game_metas`); err != nil {
		return err
	}
	return nil
}
