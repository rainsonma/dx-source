package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000041_CreateGameStatsTotalsTable struct{}

func (r *M20260322000041_CreateGameStatsTotalsTable) Signature() string {
	return "20260322000041_create_game_stats_totals_table"
}

func (r *M20260322000041_CreateGameStatsTotalsTable) Up() error {
	if !facades.Schema().HasTable("game_stats_totals") {
		return facades.Schema().Create("game_stats_totals", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_id")
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
		})
	}
	return nil
}

func (r *M20260322000041_CreateGameStatsTotalsTable) Down() error {
	return facades.Schema().DropIfExists("game_stats_totals")
}
