package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000003AddGameSessionsRaw struct{}

func (r *M20260405000003AddGameSessionsRaw) Signature() string {
	return "20260405000003_add_game_sessions_raw"
}

func (r *M20260405000003AddGameSessionsRaw) Up() error {
	statements := []string{
		`CREATE UNIQUE INDEX idx_game_sessions_active_single ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, '')) WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_group ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_group_id) WHERE ended_at IS NULL AND game_group_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_pk ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_pk_id) WHERE ended_at IS NULL AND game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_group ON game_sessions (game_group_id) WHERE game_group_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_pk ON game_sessions (game_pk_id) WHERE game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_leaderboard ON game_sessions (user_id, last_played_at)`,
		`CREATE INDEX idx_game_sessions_user_game ON game_sessions (user_id, game_id)`,
		`CREATE INDEX idx_game_sessions_current_content_vocab_id
           ON game_sessions (current_content_vocab_id)
           WHERE current_content_vocab_id IS NOT NULL`,
		`ALTER TABLE game_sessions
           ADD CONSTRAINT game_sessions_current_content_xor
           CHECK (NOT (current_content_item_id IS NOT NULL
                  AND current_content_vocab_id IS NOT NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260405000003AddGameSessionsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_sessions DROP CONSTRAINT IF EXISTS game_sessions_current_content_xor",
		"DROP INDEX IF EXISTS idx_game_sessions_current_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_sessions_active_single",
		"DROP INDEX IF EXISTS idx_game_sessions_active_group",
		"DROP INDEX IF EXISTS idx_game_sessions_active_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_group",
		"DROP INDEX IF EXISTS idx_game_sessions_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_leaderboard",
		"DROP INDEX IF EXISTS idx_game_sessions_user_game",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
