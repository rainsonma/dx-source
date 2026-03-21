package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000026_CreateContentSeeksTable struct{}

func (r *M20260322000026_CreateContentSeeksTable) Signature() string {
	return "20260322000026_create_content_seeks_table"
}

func (r *M20260322000026_CreateContentSeeksTable) Up() error {
	if !facades.Schema().HasTable("content_seeks") {
		return facades.Schema().Create("content_seeks", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("course_name").Default("")
			table.Text("description").Default("")
			table.String("disk_url").Default("")
			table.Integer("count").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000026_CreateContentSeeksTable) Down() error {
	return facades.Schema().DropIfExists("content_seeks")
}
