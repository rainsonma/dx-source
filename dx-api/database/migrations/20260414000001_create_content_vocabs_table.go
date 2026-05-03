package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000001CreateContentVocabsTable struct{}

func (r *M20260414000001CreateContentVocabsTable) Signature() string {
	return "20260414000001_create_content_vocabs_table"
}

func (r *M20260414000001CreateContentVocabsTable) Up() error {
	if facades.Schema().HasTable("content_vocabs") {
		return nil
	}
	return facades.Schema().Create("content_vocabs", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Text("content").Default("")
		table.Text("content_key").Default("")
		table.Text("uk_phonetic").Nullable()
		table.Text("us_phonetic").Nullable()
		table.Text("uk_audio_url").Nullable()
		table.Text("us_audio_url").Nullable()
		table.Json("definition").Nullable()
		table.Text("explanation").Nullable()
		table.Boolean("is_verified").Default(false)
		table.Uuid("created_by").Nullable()
		table.Uuid("last_edited_by").Nullable()
		table.SoftDeletesTz()
		table.TimestampsTz()
		table.Index("created_by")
		table.Index("created_at")
	})
}

func (r *M20260414000001CreateContentVocabsTable) Down() error {
	return facades.Schema().DropIfExists("content_vocabs")
}
