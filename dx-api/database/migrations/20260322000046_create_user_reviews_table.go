package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000046CreateUserReviewsTable struct{}

func (r *M20260322000046CreateUserReviewsTable) Signature() string {
	return "20260322000046_create_user_reviews_table"
}

func (r *M20260322000046CreateUserReviewsTable) Up() error {
	if !facades.Schema().HasTable("user_reviews") {
		return facades.Schema().Create("user_reviews", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("user_id")
			table.Text("content_item_id")
			table.Text("game_id")
			table.Text("game_level_id")
			table.TimestampTz("last_review_at").Nullable()
			table.TimestampTz("next_review_at").Nullable()
			table.Integer("review_count").Default(0)
			table.TimestampsTz()
			table.Unique("user_id", "content_item_id")
			table.Index("user_id")
			table.Index("content_item_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("next_review_at")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000046CreateUserReviewsTable) Down() error {
	return facades.Schema().DropIfExists("user_reviews")
}
