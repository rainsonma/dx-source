package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000001CreateContentVocabsAndGameVocabsTables struct{}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Signature() string {
	return "20260414000001_create_content_vocabs_and_game_vocabs_tables"
}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Up() error {
	if !facades.Schema().HasTable("content_vocabs") {
		if err := facades.Schema().Create("content_vocabs", func(table schema.Blueprint) {
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
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("game_vocabs") {
		if err := facades.Schema().Create("game_vocabs", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_vocab_id")
			table.Double("order").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("game_id")
			table.Index("content_vocab_id")
			table.Index("created_at")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_vocabs_level_order")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Down() error {
	if err := facades.Schema().DropIfExists("game_vocabs"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("content_vocabs")
}
