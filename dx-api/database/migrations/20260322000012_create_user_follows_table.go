package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000012CreateUserFollowsTable struct{}

func (r *M20260322000012CreateUserFollowsTable) Signature() string {
	return "20260322000012_create_user_follows_table"
}

func (r *M20260322000012CreateUserFollowsTable) Up() error {
	if !facades.Schema().HasTable("user_follows") {
		return facades.Schema().Create("user_follows", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("follower_id")
			table.Uuid("following_id")
			table.TimestampTz("created_at").Nullable()
			table.Unique("follower_id", "following_id")
			table.Index("follower_id")
			table.Index("following_id")
		})
	}
	return nil
}

func (r *M20260322000012CreateUserFollowsTable) Down() error {
	return facades.Schema().DropIfExists("user_follows")
}
