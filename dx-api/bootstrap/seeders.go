package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"dx-api/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
		&seeders.AdmUserSeeder{},
		&seeders.AdmPermitSeeder{},
		&seeders.AdmRoleSeeder{},
		&seeders.AdmMenuSeeder{},
		&seeders.GameCategorySeeder{},
		&seeders.GamePressSeeder{},
		&seeders.UserSeeder{},
		&seeders.UserBeanSeeder{},
	}
}
