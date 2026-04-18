package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000001CreateUsersTable struct{}

func (r *M20260322000001CreateUsersTable) Signature() string {
	return "20260322000001_create_users_table"
}

func (r *M20260322000001CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("openid").Nullable()
			table.Text("unionid").Nullable()
			table.Text("grade").Default("")
			table.Text("username").Default("")
			table.Text("nickname").Nullable()
			table.Text("email").Nullable()
			table.Text("phone").Nullable()
			table.Text("password").Default("")
			table.Text("avatar_url").Nullable()
			table.Text("city").Nullable()
			table.Text("introduction").Nullable()
			table.Text("invite_code").Default("")
			table.Integer("beans").Default(0)
			table.Integer("granted_beans").Default(0)
			table.Integer("exp").Default(0)
			table.Integer("current_play_streak").Default(0)
			table.Integer("max_play_streak").Default(0)
			table.TimestampTz("last_played_at").Nullable()
			table.TimestampTz("vip_due_at").Nullable()
			table.TimestampTz("last_read_notice_at").Nullable()
			table.Boolean("is_active").Default(true)
			table.Boolean("is_mock").Default(false)
			table.TimestampsTz()
			table.Unique("username")
			table.Unique("email")
			table.Unique("phone")
			table.Unique("invite_code")
			table.Unique("openid")
			table.Index("nickname")
			table.Index("created_at")
			table.Index("last_played_at")
		})
	}
	return nil
}

func (r *M20260322000001CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
