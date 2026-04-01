package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"dx-api/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}
