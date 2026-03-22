package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000038CreateGameSessionTotalsTable struct{}

func (r *M20260322000038CreateGameSessionTotalsTable) Signature() string {
	return "20260322000038_create_game_session_totals_table"
}

func (r *M20260322000038CreateGameSessionTotalsTable) Up() error {
	if !facades.Schema().HasTable("game_session_totals") {
		return facades.Schema().Create("game_session_totals", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_id")
			table.String("degree").Default("")
			table.String("pattern").Nullable()
			table.String("current_level_id").Default("")
			table.String("current_content_item_id").Nullable()
			table.TimestampTz("started_at")
			table.TimestampTz("last_played_at")
			table.TimestampTz("ended_at").Nullable()
			table.Integer("score").Default(0)
			table.Integer("exp").Default(0)
			table.Integer("max_combo").Default(0)
			table.Integer("correct_count").Default(0)
			table.Integer("wrong_count").Default(0)
			table.Integer("skip_count").Default(0)
			table.Integer("play_time").Default(0)
			table.Integer("total_levels_count").Default(0)
			table.Integer("played_levels_count").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000038CreateGameSessionTotalsTable) Down() error {
	return facades.Schema().DropIfExists("game_session_totals")
}
