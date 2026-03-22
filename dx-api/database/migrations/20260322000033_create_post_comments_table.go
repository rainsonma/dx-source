package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000033CreatePostCommentsTable struct{}

func (r *M20260322000033CreatePostCommentsTable) Signature() string {
	return "20260322000033_create_post_comments_table"
}

func (r *M20260322000033CreatePostCommentsTable) Up() error {
	if !facades.Schema().HasTable("post_comments") {
		return facades.Schema().Create("post_comments", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("post_id")
			table.Text("user_id")
			table.Text("content").Default("")
			table.Integer("like_count").Default(0)
			table.TimestampsTz()
			table.Index("post_id")
			table.Index("user_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000033CreatePostCommentsTable) Down() error {
	return facades.Schema().DropIfExists("post_comments")
}
