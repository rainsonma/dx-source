package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000018CreateAudiosTable struct{}

func (r *M20260322000018CreateAudiosTable) Signature() string {
	return "20260322000018_create_audios_table"
}

func (r *M20260322000018CreateAudiosTable) Up() error {
	if !facades.Schema().HasTable("audios") {
		return facades.Schema().Create("audios", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("adm_user_id").Nullable()
			table.Uuid("user_id").Nullable()
			table.Text("url").Default("")
			table.Text("name").Default("")
			table.Text("mime").Default("")
			table.Integer("size").Default(0)
			table.Integer("duration").Default(0)
			table.Text("role").Default("")
			table.TimestampsTz()
			table.Index("adm_user_id")
			table.Index("user_id")
			table.Index("url")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000018CreateAudiosTable) Down() error {
	return facades.Schema().DropIfExists("audios")
}
