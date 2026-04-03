package config

import (
	"os"
	"strings"

	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("jwt", map[string]any{
		"secret":      config.Env("JWT_SECRET", ""),
		"ttl":         config.Env("JWT_TTL", 60),
		"refresh_ttl": config.Env("JWT_REFRESH_TTL", 20160),
	})

	// Read directly from OS env to avoid Goravel/Viper IsSet bug
	// where Docker env_file vars are invisible to viper.IsSet().
	cookieSecure := true
	if v := os.Getenv("JWT_COOKIE_SECURE"); strings.EqualFold(v, "false") || v == "0" {
		cookieSecure = false
	}
	config.Add("jwt_cookie", map[string]any{
		"secure": cookieSecure,
	})
}
