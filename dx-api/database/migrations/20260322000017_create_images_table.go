package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000017_CreateImagesTable struct{}

func (r *M20260322000017_CreateImagesTable) Signature() string {
	return "20260322000017_create_images_table"
}

func (r *M20260322000017_CreateImagesTable) Up() error {
	if !facades.Schema().HasTable("images") {
		return facades.Schema().Create("images", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("adm_user_id").Nullable()
			table.String("user_id").Nullable()
			table.String("url").Default("")
			table.String("name").Default("")
			table.String("mime").Default("")
			table.Integer("size").Default(0)
			table.String("role").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000017_CreateImagesTable) Down() error {
	return facades.Schema().DropIfExists("images")
}
