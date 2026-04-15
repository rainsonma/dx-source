package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260415000001RelaxGameJunctionIndexes struct{}

func (r *M20260415000001RelaxGameJunctionIndexes) Signature() string {
	return "20260415000001_relax_game_junction_indexes"
}

func (r *M20260415000001RelaxGameJunctionIndexes) Up() error {
	stmts := []string{
		// Step 1: Create the new non-unique 3-column indexes BEFORE dropping
		//         the old unique ones. Columns are indexed throughout, and the
		//         old uniqueness is still enforced while the new indexes build.
		`CREATE INDEX IF NOT EXISTS idx_game_metas_level_meta
		   ON game_metas (game_level_id, content_meta_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_game_items_level_item
		   ON game_items (game_level_id, content_item_id, deleted_at)`,
		// Step 2: Drop the old partial unique indexes. Columns remain covered
		//         by the new indexes from step 1.
		`DROP INDEX IF EXISTS idx_game_metas_level_meta_unique`,
		`DROP INDEX IF EXISTS idx_game_items_level_item_unique`,
	}
	for _, sql := range stmts {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260415000001RelaxGameJunctionIndexes) Down() error {
	stmts := []string{
		// Recreate the original partial unique indexes. This will fail if the
		// application has created multi-junction M:N data post-Up — expected,
		// Down is for dev-time rollback before any M:N data exists.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_metas_level_meta_unique
		   ON game_metas (game_level_id, content_meta_id)
		   WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_items_level_item_unique
		   ON game_items (game_level_id, content_item_id)
		   WHERE deleted_at IS NULL`,
		// Drop the new non-unique indexes.
		`DROP INDEX IF EXISTS idx_game_metas_level_meta`,
		`DROP INDEX IF EXISTS idx_game_items_level_item`,
	}
	for _, sql := range stmts {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
