package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000003CreateContentVocabEditsTable struct{}

func (r *M20260414000003CreateContentVocabEditsTable) Signature() string {
	return "20260414000003_create_content_vocab_edits_table"
}

func (r *M20260414000003CreateContentVocabEditsTable) Up() error {
	if facades.Schema().HasTable("content_vocab_edits") {
		return nil
	}
	return facades.Schema().Create("content_vocab_edits", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("content_vocab_id")
		table.Uuid("editor_user_id").Nullable()
		table.Text("edit_type").Default("")
		table.Json("before").Nullable()
		table.Json("after").Nullable()
		table.SoftDeletesTz()
		table.TimestampsTz()
		table.Index("content_vocab_id")
		table.Index("editor_user_id")
		table.Index("created_at")
	})
}

func (r *M20260414000003CreateContentVocabEditsTable) Down() error {
	return facades.Schema().DropIfExists("content_vocab_edits")
}
