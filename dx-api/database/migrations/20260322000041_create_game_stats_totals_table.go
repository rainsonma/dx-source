package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000041CreateGameStatsTotalsTable struct{}

func (r *M20260322000041CreateGameStatsTotalsTable) Signature() string {
	return "20260322000041_create_game_stats_totals_table"
}

func (r *M20260322000041CreateGameStatsTotalsTable) Up() error {
	if !facades.Schema().HasTable("game_stats_totals") {
		return facades.Schema().Create("game_stats_totals", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("game_id")
			table.Integer("total_sessions").Default(0)
			table.Integer("total_exp").Default(0)
			table.Integer("highest_score").Default(0)
			table.Integer("total_scores").Default(0)
			table.Integer("total_play_time").Default(0)
			table.TimestampTz("first_played_at")
			table.TimestampTz("last_played_at")
			table.TimestampTz("first_completed_at").Nullable()
			table.TimestampTz("last_completed_at").Nullable()
			table.Integer("completion_count").Default(0)
			table.TimestampsTz()
			table.Unique("user_id", "game_id")
			table.Index("user_id")
			table.Index("game_id")
			table.Index("first_completed_at")
		})
	}
	return nil
}

func (r *M20260322000041CreateGameStatsTotalsTable) Down() error {
	return facades.Schema().DropIfExists("game_stats_totals")
}
