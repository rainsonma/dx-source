package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000048AddUserUnknownsRaw struct{}

func (r *M20260322000048AddUserUnknownsRaw) Signature() string {
	return "20260322000048_add_user_unknowns_raw"
}

func (r *M20260322000048AddUserUnknownsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_user_unknowns_content_item_id
           ON user_unknowns (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_user_unknowns_content_vocab_id
           ON user_unknowns (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_user_unknowns_user_item_uq
           ON user_unknowns (user_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_user_unknowns_user_vocab_uq
           ON user_unknowns (user_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE user_unknowns
           ADD CONSTRAINT user_unknowns_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000048AddUserUnknownsRaw) Down() error {
	statements := []string{
		"ALTER TABLE user_unknowns DROP CONSTRAINT IF EXISTS user_unknowns_content_xor",
		"DROP INDEX IF EXISTS idx_user_unknowns_user_vocab_uq",
		"DROP INDEX IF EXISTS idx_user_unknowns_user_item_uq",
		"DROP INDEX IF EXISTS idx_user_unknowns_content_vocab_id",
		"DROP INDEX IF EXISTS idx_user_unknowns_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
