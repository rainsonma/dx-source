package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000032CreatePostsTable struct{}

func (r *M20260322000032CreatePostsTable) Signature() string {
	return "20260322000032_create_posts_table"
}

func (r *M20260322000032CreatePostsTable) Up() error {
	if !facades.Schema().HasTable("posts") {
		return facades.Schema().Create("posts", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Text("content").Default("")
			table.Text("image_url").Nullable()
			table.Column("tags", "text[]").Nullable()
			table.Integer("like_count").Default(0)
			table.Integer("comment_count").Default(0)
			table.Integer("share_count").Default(0)
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("is_active")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000032CreatePostsTable) Down() error {
	return facades.Schema().DropIfExists("posts")
}
