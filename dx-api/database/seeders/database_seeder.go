package seeders

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"github.com/goravel/framework/facades"
)

type DatabaseSeeder struct{}

func (s *DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

func (s *DatabaseSeeder) Run() error {
	return facades.Seeder().Call([]seeder.Seeder{
		&AdmUserSeeder{},
		&AdmPermitSeeder{},
		&AdmRoleSeeder{},
		&AdmMenuSeeder{},
		&GameCategorySeeder{},
		&GamePressSeeder{},
		&UserSeeder{},
		&UserBeanSeeder{},
	})
}
