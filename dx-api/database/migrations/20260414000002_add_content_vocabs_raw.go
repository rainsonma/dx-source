package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000002AddContentVocabsRaw struct{}

func (r *M20260414000002AddContentVocabsRaw) Signature() string {
	return "20260414000002_add_content_vocabs_raw"
}

func (r *M20260414000002AddContentVocabsRaw) Up() error {
	statements := []string{
		`CREATE UNIQUE INDEX idx_content_vocabs_user_content_key_uq
           ON content_vocabs (user_id, content_key)
           WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_content_vocabs_content_key
           ON content_vocabs (content_key, deleted_at)`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000002AddContentVocabsRaw) Down() error {
	statements := []string{
		"DROP INDEX IF EXISTS idx_content_vocabs_content_key",
		"DROP INDEX IF EXISTS idx_content_vocabs_user_content_key_uq",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
