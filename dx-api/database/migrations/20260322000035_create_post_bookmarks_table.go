package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000035CreatePostBookmarksTable struct{}

func (r *M20260322000035CreatePostBookmarksTable) Signature() string {
	return "20260322000035_create_post_bookmarks_table"
}

func (r *M20260322000035CreatePostBookmarksTable) Up() error {
	if !facades.Schema().HasTable("post_bookmarks") {
		return facades.Schema().Create("post_bookmarks", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("post_id")
			table.Text("user_id")
			table.TimestampTz("created_at").Nullable()
			table.Unique("post_id", "user_id")
			table.Index("post_id")
			table.Index("user_id")
		})
	}
	return nil
}

func (r *M20260322000035CreatePostBookmarksTable) Down() error {
	return facades.Schema().DropIfExists("post_bookmarks")
}
