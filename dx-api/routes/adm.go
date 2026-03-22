package routes

import (
	adm "dx-api/app/http/controllers/adm"
	"dx-api/app/http/middleware"

	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/route"
)

func Adm() {
	r := facades.Route()

	admAuthController := adm.NewAuthController()

	// All admin API routes under /adm prefix
	r.Prefix("/adm").Group(func(router route.Router) {
		// Admin auth routes (public, no JWT required)
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/login", admAuthController.Login)
			auth.Post("/refresh", admAuthController.Refresh)
			auth.Post("/logout", admAuthController.Logout)
		})

		// Admin auth routes (protected, admin JWT required)
		router.Prefix("/auth").Middleware(middleware.AdmJwtAuth()).Group(func(auth route.Router) {
			auth.Get("/me", admAuthController.Me)
		})

		// Protected admin routes (admin JWT + RBAC + operation log)
		router.Middleware(
			middleware.AdmJwtAuth(),
			middleware.AdmRbac(),
			middleware.AdmOperateLog(),
		).Group(func(protected route.Router) {
			communityController := adm.NewCommunityController()

			// Notices
			protected.Post("/notices", communityController.CreateNotice)
			protected.Put("/notices/{id}", communityController.UpdateNotice)
			protected.Delete("/notices/{id}", communityController.DeleteNotice)

			// Redeems
			protected.Post("/redeems/generate", communityController.GenerateCodes)
			protected.Get("/redeems", communityController.GetAllRedeems)
		})
	})
}
