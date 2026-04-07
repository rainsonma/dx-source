package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260407000001CreateGameJunctionTables struct{}

func (r *M20260407000001CreateGameJunctionTables) Signature() string {
	return "20260407000001_create_game_junction_tables"
}

func (r *M20260407000001CreateGameJunctionTables) Up() error {
	// 1. Create game_metas table
	if !facades.Schema().HasTable("game_metas") {
		if err := facades.Schema().Create("game_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_meta_id")
			table.TimestampTz("created_at").UseCurrent()
			table.Index("game_level_id")
			table.Index("content_meta_id")
		}); err != nil {
			return err
		}
		// Composite unique index
		if _, err := facades.Orm().Query().Exec(
			"CREATE UNIQUE INDEX idx_game_metas_unique ON game_metas (game_id, game_level_id, content_meta_id)",
		); err != nil {
			return err
		}
	}

	// 2. Create game_items table
	if !facades.Schema().HasTable("game_items") {
		if err := facades.Schema().Create("game_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id")
			table.TimestampTz("created_at").UseCurrent()
			table.Index("game_level_id")
			table.Index("content_item_id")
		}); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Exec(
			"CREATE UNIQUE INDEX idx_game_items_unique ON game_items (game_id, game_level_id, content_item_id)",
		); err != nil {
			return err
		}
	}

	// 3. Backfill game_metas from existing content_metas
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

	// 4. Backfill game_items from existing content_items
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

	// 5. Make game_level_id nullable on content tables
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

func (r *M20260407000001CreateGameJunctionTables) Down() error {
	// Restore NOT NULL (only safe if no NULLs exist yet)
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_metas ALTER COLUMN game_level_id SET NOT NULL",
	)
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_items ALTER COLUMN game_level_id SET NOT NULL",
	)
	_ = facades.Schema().DropIfExists("game_items")
	_ = facades.Schema().DropIfExists("game_metas")
	return nil
}
