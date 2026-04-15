package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000001CreateGameMetasAndGameItemsTables struct{}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Signature() string {
	return "20260414000001_create_game_metas_and_game_items_tables"
}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Up() error {
	// 1. game_metas
	if !facades.Schema().HasTable("game_metas") {
		if err := facades.Schema().Create("game_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_meta_id")
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("game_id")
			table.Index("content_meta_id")
			table.Index("created_at")
			table.Index("game_level_id", "content_meta_id", "deleted_at").Name("idx_game_metas_level_meta")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_metas_level_order")
		}); err != nil {
			return err
		}
	}

	// 2. game_items
	if !facades.Schema().HasTable("game_items") {
		if err := facades.Schema().Create("game_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id")
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("game_id")
			table.Index("content_item_id")
			table.Index("created_at")
			table.Index("game_level_id", "content_item_id", "deleted_at").Name("idx_game_items_level_item")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_items_level_order")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Down() error {
	if err := facades.Schema().DropIfExists("game_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("game_metas")
}
