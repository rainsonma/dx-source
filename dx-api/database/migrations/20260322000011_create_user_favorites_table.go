package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000011CreateUserFavoritesTable struct{}

func (r *M20260322000011CreateUserFavoritesTable) Signature() string {
	return "20260322000011_create_user_favorites_table"
}

func (r *M20260322000011CreateUserFavoritesTable) Up() error {
	if !facades.Schema().HasTable("user_favorites") {
		return facades.Schema().Create("user_favorites", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("user_id")
			table.Text("game_id")
			table.TimestampTz("created_at").Nullable()
			table.Unique("user_id", "game_id")
			table.Index("user_id")
			table.Index("game_id")
		})
	}
	return nil
}

func (r *M20260322000011CreateUserFavoritesTable) Down() error {
	return facades.Schema().DropIfExists("user_favorites")
}
