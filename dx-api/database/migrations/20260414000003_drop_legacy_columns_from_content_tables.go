package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000003DropLegacyColumnsFromContentTables struct{}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Signature() string {
	return "20260414000003_drop_legacy_columns_from_content_tables"
}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Up() error {
	stmts := []string{
		`ALTER TABLE content_metas DROP COLUMN IF EXISTS game_level_id`,
		`ALTER TABLE content_metas DROP COLUMN IF EXISTS "order"`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS game_level_id`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS "order"`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS is_active`,
	}
	for _, s := range stmts {
		if _, err := facades.Orm().Query().Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Down() error {
	stmts := []string{
		`ALTER TABLE content_metas ADD COLUMN IF NOT EXISTS game_level_id UUID`,
		`ALTER TABLE content_metas ADD COLUMN IF NOT EXISTS "order" DOUBLE PRECISION DEFAULT 0`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS game_level_id UUID`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS "order" DOUBLE PRECISION DEFAULT 0`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true`,
	}
	for _, s := range stmts {
		if _, err := facades.Orm().Query().Exec(s); err != nil {
			return err
		}
	}
	return nil
}
