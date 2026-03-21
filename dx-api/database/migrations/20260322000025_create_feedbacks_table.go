package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000025_CreateFeedbacksTable struct{}

func (r *M20260322000025_CreateFeedbacksTable) Signature() string {
	return "20260322000025_create_feedbacks_table"
}

func (r *M20260322000025_CreateFeedbacksTable) Up() error {
	if !facades.Schema().HasTable("feedbacks") {
		return facades.Schema().Create("feedbacks", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("type").Default("")
			table.Text("description").Default("")
			table.Integer("count").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000025_CreateFeedbacksTable) Down() error {
	return facades.Schema().DropIfExists("feedbacks")
}
