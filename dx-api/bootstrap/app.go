package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"

	"dx-api/app/console/commands"
	"github.com/goravel/framework/facades"
	"dx-api/app/jobs"
	"dx-api/config"
	"dx-api/routes"
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
		WithJobs(func() []queue.Job {
			return []queue.Job{
				&jobs.SendEmailJob{},
			}
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
