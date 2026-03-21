package middleware

import (
	contractshttp "github.com/goravel/framework/contracts/http"
)

// AdmOperateLog logs admin operations to the adm_operates table.
// This middleware must run AFTER AdmJwtAuth so the admin user is authenticated.
func AdmOperateLog() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		// TODO: Implement operation logging
		// 1. Get admin user ID via facades.Auth(ctx).Guard("admin").ID()
		// 2. After the request completes, record to adm_operates:
		//    - admin_user_id: authenticated admin's ID
		//    - path: ctx.Request().Path()
		//    - method: ctx.Request().Method()
		//    - ip: ctx.Request().Ip()
		//    - input: request body / parameters
		// 3. Use defer or post-Next() logic to capture the response status
		ctx.Request().Next()
	}
}
