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
			table.String("id")
			table.Primary("id")
			table.String("post_id")
			table.String("user_id")
			table.TimestampTz("created_at").Nullable()
		})
	}
	return nil
}

func (r *M20260322000035CreatePostBookmarksTable) Down() error {
	return facades.Schema().DropIfExists("post_bookmarks")
}
