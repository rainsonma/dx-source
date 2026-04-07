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
			table.Unique("game_id", "game_level_id", "content_meta_id")
			table.Index("game_level_id")
			table.Index("content_meta_id")
		}); err != nil {
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
			table.Unique("game_id", "game_level_id", "content_item_id")
			table.Index("game_level_id")
			table.Index("content_item_id")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260407000001CreateGameJunctionTables) Down() error {
	_ = facades.Schema().DropIfExists("game_items")
	_ = facades.Schema().DropIfExists("game_metas")
	return nil
}
