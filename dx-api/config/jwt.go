package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("jwt", map[string]any{
		"secret": config.Env("JWT_SECRET", ""),
		"ttl":    config.Env("JWT_TTL", 10),
	})
	config.Add("refresh_token", map[string]any{
		"ttl":           config.Env("REFRESH_TOKEN_TTL", 10080),
		"cookie_secure": config.Env("REFRESH_COOKIE_SECURE", true),
	})
}
