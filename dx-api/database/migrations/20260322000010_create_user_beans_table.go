package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000010CreateUserBeansTable struct{}

func (r *M20260322000010CreateUserBeansTable) Signature() string {
	return "20260322000010_create_user_beans_table"
}

func (r *M20260322000010CreateUserBeansTable) Up() error {
	if !facades.Schema().HasTable("user_beans") {
		return facades.Schema().Create("user_beans", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.Integer("beans").Default(0)
			table.Integer("origin").Default(0)
			table.Integer("result").Default(0)
			table.String("slug").Default("")
			table.String("reason").Default("")
			table.Text("data").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000010CreateUserBeansTable) Down() error {
	return facades.Schema().DropIfExists("user_beans")
}
