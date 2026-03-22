package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000001CreateUsersTable struct{}

func (r *M20260322000001CreateUsersTable) Signature() string {
	return "20260322000001_create_users_table"
}

func (r *M20260322000001CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("grade").Default("")
			table.String("username").Default("")
			table.String("nickname").Nullable()
			table.String("email").Nullable()
			table.String("phone").Nullable()
			table.String("password").Default("")
			table.String("avatar_id").Nullable()
			table.String("city").Nullable()
			table.String("introduction").Nullable()
			table.Boolean("is_active").Default(true)
			table.Integer("beans").Default(0)
			table.Integer("granted_beans").Default(0)
			table.Integer("exp").Default(0)
			table.String("invite_code").Default("")
			table.Integer("current_play_streak").Default(0)
			table.Integer("max_play_streak").Default(0)
			table.TimestampTz("last_played_at").Nullable()
			table.TimestampTz("vip_due_at").Nullable()
			table.TimestampTz("last_read_notice_at").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000001CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
