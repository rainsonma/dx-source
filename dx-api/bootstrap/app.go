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
				&commands.ImportCourses{},
				&commands.ExpireStaleOrders{},
			}
		}).
		WithSchedule(func() []schedule.Event {
			return []schedule.Event{
				facades.Schedule().Command("app:reset-energy-beans").DailyAt("01:00").SkipIfStillRunning().Name("reset-energy-beans"),
				facades.Schedule().Command("app:update-play-streaks").DailyAt("02:00").SkipIfStillRunning().Name("update-play-streaks"),
				facades.Schedule().Command("app:expire-stale-orders").EveryFiveMinutes().SkipIfStillRunning().Name("expire-stale-orders"),
			}
		}).
		WithRules(Rules).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
