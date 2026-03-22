package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000012CreateUserFollowsTable struct{}

func (r *M20260322000012CreateUserFollowsTable) Signature() string {
	return "20260322000012_create_user_follows_table"
}

func (r *M20260322000012CreateUserFollowsTable) Up() error {
	if !facades.Schema().HasTable("user_follows") {
		return facades.Schema().Create("user_follows", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("follower_id")
			table.String("following_id")
			table.TimestampTz("created_at").Nullable()
		})
	}
	return nil
}

func (r *M20260322000012CreateUserFollowsTable) Down() error {
	return facades.Schema().DropIfExists("user_follows")
}
