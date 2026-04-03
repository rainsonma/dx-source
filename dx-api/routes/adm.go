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
			noticeController := adm.NewNoticeController()
			redeemController := adm.NewRedeemController()

			// Notices
			protected.Post("/notices", noticeController.CreateNotice)
			protected.Put("/notices/{id}", noticeController.UpdateNotice)
			protected.Delete("/notices/{id}", noticeController.DeleteNotice)

			// Redeems
			protected.Post("/redeems/generate", redeemController.GenerateCodes)
			protected.Get("/redeems", redeemController.GetAllRedeems)
		})
	})
}
