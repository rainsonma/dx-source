package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000003CreateGameVocabsTable struct{}

func (r *M20260414000003CreateGameVocabsTable) Signature() string {
	return "20260414000003_create_game_vocabs_table"
}

func (r *M20260414000003CreateGameVocabsTable) Up() error {
	if facades.Schema().HasTable("game_vocabs") {
		return nil
	}
	return facades.Schema().Create("game_vocabs", func(table schema.Blueprint) {
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
	})
}

func (r *M20260414000003CreateGameVocabsTable) Down() error {
	return facades.Schema().DropIfExists("game_vocabs")
}
