package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000050AddUserReviewsRaw struct{}

func (r *M20260322000050AddUserReviewsRaw) Signature() string {
	return "20260322000050_add_user_reviews_raw"
}

func (r *M20260322000050AddUserReviewsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_user_reviews_content_item_id
           ON user_reviews (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_user_reviews_content_vocab_id
           ON user_reviews (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_user_reviews_user_item_uq
           ON user_reviews (user_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_user_reviews_user_vocab_uq
           ON user_reviews (user_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE user_reviews
           ADD CONSTRAINT user_reviews_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000050AddUserReviewsRaw) Down() error {
	statements := []string{
		"ALTER TABLE user_reviews DROP CONSTRAINT IF EXISTS user_reviews_content_xor",
		"DROP INDEX IF EXISTS idx_user_reviews_user_vocab_uq",
		"DROP INDEX IF EXISTS idx_user_reviews_user_item_uq",
		"DROP INDEX IF EXISTS idx_user_reviews_content_vocab_id",
		"DROP INDEX IF EXISTS idx_user_reviews_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
