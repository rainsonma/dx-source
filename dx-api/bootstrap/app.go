package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"

	"dx-api/app/console/commands"
	"dx-api/config"
	"dx-api/routes"

	"github.com/goravel/framework/facades"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithSeeders(Seeders).
		WithRouting(func() {
			routes.Web()
			routes.Api()
			routes.Adm()
			routes.Grpc()
		}).
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.UpdatePlayStreaks{},
				&commands.ResetEnergyBeans{},
			}
		}).
		WithSchedule(func() []schedule.Event {
			return []schedule.Event{
				facades.Schedule().Command("app:reset-energy-beans").DailyAt("01:00").Name("reset-energy-beans"),
				facades.Schedule().Command("app:update-play-streaks").DailyAt("02:00").Name("update-play-streaks"),
			}
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
