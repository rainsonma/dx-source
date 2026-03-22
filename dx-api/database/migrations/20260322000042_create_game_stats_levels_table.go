package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000042CreateGameStatsLevelsTable struct{}

func (r *M20260322000042CreateGameStatsLevelsTable) Signature() string {
	return "20260322000042_create_game_stats_levels_table"
}

func (r *M20260322000042CreateGameStatsLevelsTable) Up() error {
	if !facades.Schema().HasTable("game_stats_levels") {
		return facades.Schema().Create("game_stats_levels", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_level_id")
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

func (r *M20260322000042CreateGameStatsLevelsTable) Down() error {
	return facades.Schema().DropIfExists("game_stats_levels")
}
