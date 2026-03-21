package config

import (
	"dx-api/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		// Cross-Origin Resource Sharing (CORS) Configuration
		//
		// Here you may configure your settings for cross-origin resource sharing
		// or "CORS". This determines what cross-origin operations may execute
		// in web browsers. You are free to adjust these settings as needed.
		//
		// To learn more: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
		"paths":                []string{"*"},
		"allowed_methods":      []string{"*"},
		"allowed_origins":      []string{"http://localhost:3000", config.Env("CORS_ALLOWED_ORIGINS", "").(string)},
		"allowed_headers":      []string{"Authorization", "Content-Type", "X-Requested-With", "Accept"},
		"exposed_headers":      []string{},
		"max_age":              0,
		"supports_credentials": true,
	})
}
