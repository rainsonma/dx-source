package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000043CreateGameReportsTable struct{}

func (r *M20260322000043CreateGameReportsTable) Signature() string {
	return "20260322000043_create_game_reports_table"
}

func (r *M20260322000043CreateGameReportsTable) Up() error {
	if !facades.Schema().HasTable("game_reports") {
		return facades.Schema().Create("game_reports", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_id")
			table.String("game_level_id")
			table.String("content_item_id")
			table.String("reason").Default("")
			table.Text("note").Nullable()
			table.Integer("count").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000043CreateGameReportsTable) Down() error {
	return facades.Schema().DropIfExists("game_reports")
}
