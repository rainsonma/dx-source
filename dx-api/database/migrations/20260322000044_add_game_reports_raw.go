package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000044AddGameReportsRaw struct{}

func (r *M20260322000044AddGameReportsRaw) Signature() string {
	return "20260322000044_add_game_reports_raw"
}

func (r *M20260322000044AddGameReportsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_game_reports_content_item_id
           ON game_reports (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_game_reports_content_vocab_id
           ON game_reports (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_reports_user_item_reason_uq
           ON game_reports (user_id, content_item_id, reason)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_game_reports_user_vocab_reason_uq
           ON game_reports (user_id, content_vocab_id, reason)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE game_reports
           ADD CONSTRAINT game_reports_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000044AddGameReportsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_reports DROP CONSTRAINT IF EXISTS game_reports_content_xor",
		"DROP INDEX IF EXISTS idx_game_reports_user_vocab_reason_uq",
		"DROP INDEX IF EXISTS idx_game_reports_user_item_reason_uq",
		"DROP INDEX IF EXISTS idx_game_reports_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_reports_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
