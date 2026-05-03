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
	return facades.Schema().Create("game_sessions", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_id")
		table.Uuid("game_level_id")
		table.Text("degree").Default("")
		table.Text("pattern").Nullable()
		table.Uuid("current_content_item_id").Nullable()
		table.Uuid("current_content_vocab_id").Nullable()
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
		table.SoftDeletesTz()
		table.TimestampsTz()
	})
}

func (r *M20260405000002CreateGameSessionsTable) Down() error {
	return facades.Schema().DropIfExists("game_sessions")
}
