package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000033_CreatePostCommentsTable struct{}

func (r *M20260322000033_CreatePostCommentsTable) Signature() string {
	return "20260322000033_create_post_comments_table"
}

func (r *M20260322000033_CreatePostCommentsTable) Up() error {
	if !facades.Schema().HasTable("post_comments") {
		return facades.Schema().Create("post_comments", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("post_id")
			table.String("user_id")
			table.Text("content").Default("")
			table.Integer("like_count").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000033_CreatePostCommentsTable) Down() error {
	return facades.Schema().DropIfExists("post_comments")
}
