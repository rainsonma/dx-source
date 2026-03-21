package middleware

import (
	contractshttp "github.com/goravel/framework/contracts/http"
)

// AdmRbac checks admin RBAC permissions against the adm_permits table.
// This middleware must run AFTER AdmJwtAuth so the admin user is authenticated.
func AdmRbac() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		// TODO: Implement RBAC permission check
		// 1. Get admin user ID via facades.Auth(ctx).Guard("admin").ID()
		// 2. Load user's permits:
		//    - Direct permits: adm_user_permits -> adm_permits
		//    - Role-based permits: adm_user_roles -> adm_roles -> adm_role_permits -> adm_permits
		// 3. For each permit, check if:
		//    - Request method matches one of permit.http_methods (e.g., "GET,POST")
		//    - Request path matches one of permit.http_paths (support wildcard "*" at end)
		// 4. If any permit matches, allow the request through
		// 5. If no permit matches, return 403:
		//    ctx.Response().Json(403, helpers.Response{Code: 40300, Message: "forbidden"}).Abort()
		ctx.Request().Next()
	}
}
