package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000011_CreateUserFavoritesTable struct{}

func (r *M20260322000011_CreateUserFavoritesTable) Signature() string {
	return "20260322000011_create_user_favorites_table"
}

func (r *M20260322000011_CreateUserFavoritesTable) Up() error {
	if !facades.Schema().HasTable("user_favorites") {
		return facades.Schema().Create("user_favorites", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_id")
			table.TimestampTz("created_at").Nullable()
		})
	}
	return nil
}

func (r *M20260322000011_CreateUserFavoritesTable) Down() error {
	return facades.Schema().DropIfExists("user_favorites")
}
