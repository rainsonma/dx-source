package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000025CreateFeedbacksTable struct{}

func (r *M20260322000025CreateFeedbacksTable) Signature() string {
	return "20260322000025_create_feedbacks_table"
}

func (r *M20260322000025CreateFeedbacksTable) Up() error {
	if !facades.Schema().HasTable("feedbacks") {
		return facades.Schema().Create("feedbacks", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("user_id")
			table.Text("type").Default("")
			table.Text("description").Default("")
			table.Integer("count").Default(0)
			table.TimestampsTz()
			table.Unique("type", "description")
			table.Index("user_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000025CreateFeedbacksTable) Down() error {
	return facades.Schema().DropIfExists("feedbacks")
}
