package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000026CreateContentSeeksTable struct{}

func (r *M20260322000026CreateContentSeeksTable) Signature() string {
	return "20260322000026_create_content_seeks_table"
}

func (r *M20260322000026CreateContentSeeksTable) Up() error {
	if !facades.Schema().HasTable("content_seeks") {
		return facades.Schema().Create("content_seeks", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("user_id")
			table.Text("course_name").Default("")
			table.Text("description").Default("")
			table.Text("disk_url").Default("")
			table.Integer("count").Default(0)
			table.TimestampsTz()
			table.Unique("course_name")
			table.Index("user_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000026CreateContentSeeksTable) Down() error {
	return facades.Schema().DropIfExists("content_seeks")
}
