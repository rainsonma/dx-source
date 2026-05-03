package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000049CreateUserReviewsTable struct{}

func (r *M20260322000049CreateUserReviewsTable) Signature() string {
	return "20260322000049_create_user_reviews_table"
}

func (r *M20260322000049CreateUserReviewsTable) Up() error {
	if !facades.Schema().HasTable("user_reviews") {
		return facades.Schema().Create("user_reviews", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.TimestampTz("last_review_at").Nullable()
			table.TimestampTz("next_review_at").Nullable()
			table.Integer("review_count").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("next_review_at")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000049CreateUserReviewsTable) Down() error {
	return facades.Schema().DropIfExists("user_reviews")
}
