package config

import "github.com/goravel/framework/facades"

func init() {
	config := facades.Config()
	config.Add("wechat", map[string]any{
		"mini_app_id":     config.Env("WECHAT_MINI_APP_ID", ""),
		"mini_app_secret": config.Env("WECHAT_MINI_APP_SECRET", ""),
	})
}
