package config

import (
	"dx-api/app/facades"
	"github.com/goravel/framework/contracts/cache"
	redisfacades "github.com/goravel/redis/facades"
)

func init() {
	config := facades.Config()
	config.Add("cache", map[string]any{
		"default": config.Env("CACHE_DRIVER", "redis"),
		"stores": map[string]any{
			"memory": map[string]any{
				"driver": "memory",
			},
			"redis": map[string]any{
				"driver":     "custom",
				"connection": "default",
				"via": func() (cache.Driver, error) {
					return redisfacades.Cache("redis")
				},
			},
		},
		"prefix": config.GetString("APP_NAME", "douxue") + "_cache",
	})
}
