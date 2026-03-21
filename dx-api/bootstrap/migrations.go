package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
	}
}
