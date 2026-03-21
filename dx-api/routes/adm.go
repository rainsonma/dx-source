package routes

import (
	"dx-api/app/facades"
	"dx-api/app/http/middleware"

	"github.com/goravel/framework/contracts/route"
)

func Adm() {
	r := facades.Route()

	// All admin API routes under /adm prefix
	r.Prefix("/adm").Group(func(router route.Router) {
		// Admin auth routes (public, no JWT required)
		router.Prefix("/auth").Group(func(auth route.Router) {
			// Admin auth endpoints will be added in Phase 1
		})

		// Protected admin routes (admin JWT + RBAC + operation log)
		router.Middleware(
			middleware.AdmJwtAuth(),
			middleware.AdmRbac(),
			middleware.AdmOperateLog(),
		).Group(func(protected route.Router) {
			// Admin CRUD endpoints will be added later
		})
	})
}
