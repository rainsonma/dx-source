package commands

import (
	"fmt"
	"time"

	services "dx-api/app/services/api"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type ExpireStaleOrders struct{}

func (c *ExpireStaleOrders) Signature() string {
	return "app:expire-stale-orders"
}

func (c *ExpireStaleOrders) Description() string {
	return "Expire unpaid orders past their 30-minute window"
}

func (c *ExpireStaleOrders) Extend() command.Extend {
	return command.Extend{}
}

func (c *ExpireStaleOrders) Handle(ctx console.Context) error {
	start := time.Now()

	count, err := services.ExpireStaleOrders()
	if err != nil {
		ctx.Error(fmt.Sprintf("[expire-stale-orders] failed: %v", err))
		return err
	}

	elapsed := time.Since(start)
	ctx.Info(fmt.Sprintf("[expire-stale-orders] done in %s — expired: %d", elapsed, count))
	return nil
}
