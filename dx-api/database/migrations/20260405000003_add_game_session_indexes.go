package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000003AddGameSessionIndexes struct{}

func (r *M20260405000003AddGameSessionIndexes) Signature() string {
	return "20260405000003_add_game_session_indexes"
}

func (r *M20260405000003AddGameSessionIndexes) Up() error {
	indexes := []string{
		`CREATE UNIQUE INDEX idx_game_sessions_active_single ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, '')) WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_group ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_group_id) WHERE ended_at IS NULL AND game_group_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_pk ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_pk_id) WHERE ended_at IS NULL AND game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_group ON game_sessions (game_group_id) WHERE game_group_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_pk ON game_sessions (game_pk_id) WHERE game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_leaderboard ON game_sessions (user_id, last_played_at)`,
		`CREATE INDEX idx_game_sessions_user_game ON game_sessions (user_id, game_id)`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260405000003AddGameSessionIndexes) Down() error {
	indexes := []string{
		"DROP INDEX IF EXISTS idx_game_sessions_active_single",
		"DROP INDEX IF EXISTS idx_game_sessions_active_group",
		"DROP INDEX IF EXISTS idx_game_sessions_active_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_group",
		"DROP INDEX IF EXISTS idx_game_sessions_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_leaderboard",
		"DROP INDEX IF EXISTS idx_game_sessions_user_game",
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
