package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260405000002CreateGameSessionsTable struct{}

func (r *M20260405000002CreateGameSessionsTable) Signature() string {
	return "20260405000002_create_game_sessions_table"
}

func (r *M20260405000002CreateGameSessionsTable) Up() error {
	// Drop old tables in dependency order
	for _, table := range []string{
		"game_stats_levels",
		"game_stats_totals",
		"game_records",
		"game_session_levels",
		"game_session_totals",
		"game_pks",
	} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}

	// Create game_sessions table
	if err := facades.Schema().Create("game_sessions", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_id")
		table.Uuid("game_level_id")
		table.Text("degree").Default("")
		table.Text("pattern").Nullable()
		table.Uuid("current_content_item_id").Nullable()
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
		table.Integer("total_items_count").Default(0)
		table.Integer("played_items_count").Default(0)
		table.Uuid("game_group_id").Nullable()
		table.Uuid("game_subgroup_id").Nullable()
		table.Uuid("game_pk_id").Nullable()
		table.TimestampsTz()
	}); err != nil {
		return err
	}

	// Recreate game_records table with game_session_id
	if err := facades.Schema().Create("game_records", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_session_id")
		table.Uuid("game_level_id")
		table.Uuid("content_item_id")
		table.Boolean("is_correct").Default(false)
		table.Text("source_answer").Default("")
		table.Text("user_answer").Default("")
		table.Integer("base_score").Default(0)
		table.Integer("combo_score").Default(0)
		table.Integer("duration").Default(0)
		table.TimestampsTz()
		table.Unique("game_session_id", "content_item_id")
		table.Index("user_id")
		table.Index("game_session_id")
		table.Index("game_level_id")
		table.Index("content_item_id")
		table.Index("is_correct")
	}); err != nil {
		return err
	}

	// Recreate game_pks table with game_level_id
	if err := facades.Schema().Create("game_pks", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("opponent_id")
		table.Uuid("game_id")
		table.Uuid("game_level_id")
		table.Text("degree").Default("")
		table.Text("pattern").Nullable()
		table.Text("robot_difficulty").Default("normal")
		table.Boolean("is_playing").Default(false)
		table.Uuid("last_winner_id").Nullable()
		table.TimestampsTz()
		table.Index("user_id")
		table.Index("opponent_id")
		table.Index("game_id")
		table.Index("is_playing")
	}); err != nil {
		return err
	}

	// Recreate game_pks unique active index
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX idx_game_pks_unique_active ON game_pks (user_id, game_id) WHERE is_playing = true`); err != nil {
		return err
	}

	return nil
}

func (r *M20260405000002CreateGameSessionsTable) Down() error {
	facades.Schema().DropIfExists("game_records")
	facades.Schema().DropIfExists("game_pks")
	facades.Schema().DropIfExists("game_sessions")
	return nil
}
