package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000034CreatePostLikesTable struct{}

func (r *M20260322000034CreatePostLikesTable) Signature() string {
	return "20260322000034_create_post_likes_table"
}

func (r *M20260322000034CreatePostLikesTable) Up() error {
	if !facades.Schema().HasTable("post_likes") {
		return facades.Schema().Create("post_likes", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("post_id")
			table.String("user_id")
			table.TimestampTz("created_at").Nullable()
		})
	}
	return nil
}

func (r *M20260322000034CreatePostLikesTable) Down() error {
	return facades.Schema().DropIfExists("post_likes")
}
