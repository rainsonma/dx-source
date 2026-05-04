package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000005AddGameRecordsRaw struct{}

func (r *M20260405000005AddGameRecordsRaw) Signature() string {
	return "20260405000005_add_game_records_raw"
}

func (r *M20260405000005AddGameRecordsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_game_records_content_item_id
           ON game_records (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_game_records_content_vocab_id
           ON game_records (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_records_session_item_uq
           ON game_records (game_session_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_game_records_session_vocab_uq
           ON game_records (game_session_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE game_records
           ADD CONSTRAINT game_records_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260405000005AddGameRecordsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_records DROP CONSTRAINT IF EXISTS game_records_content_xor",
		"DROP INDEX IF EXISTS idx_game_records_session_vocab_uq",
		"DROP INDEX IF EXISTS idx_game_records_session_item_uq",
		"DROP INDEX IF EXISTS idx_game_records_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_records_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
