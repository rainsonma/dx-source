package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000002CreateAdmUsersTable struct{}

func (r *M20260322000002CreateAdmUsersTable) Signature() string {
	return "20260322000002_create_adm_users_table"
}

func (r *M20260322000002CreateAdmUsersTable) Up() error {
	if !facades.Schema().HasTable("adm_users") {
		return facades.Schema().Create("adm_users", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("username").Default("")
			table.String("nickname").Nullable()
			table.String("password").Default("")
			table.String("avatar_id").Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000002CreateAdmUsersTable) Down() error {
	return facades.Schema().DropIfExists("adm_users")
}
