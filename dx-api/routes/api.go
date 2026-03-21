package routes

import (
	"dx-api/app/facades"
	"dx-api/app/helpers"
	apicontrollers "dx-api/app/http/controllers/api"
	"dx-api/app/http/middleware"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
)

func Api() {
	r := facades.Route()

	authController := apicontrollers.NewAuthController()

	// All client API routes under /api prefix
	r.Prefix("/api").Group(func(router route.Router) {
		// Health check (public)
		router.Get("/health", func(ctx contractshttp.Context) contractshttp.Response {
			// Check DB
			dbOk := true
			if _, err := facades.Orm().Connection("postgres").Query().Exec("SELECT 1"); err != nil {
				dbOk = false
			}

			// Check Redis
			redisOk := true
			if err := helpers.RedisPing(); err != nil {
				redisOk = false
			}

			return helpers.Success(ctx, map[string]bool{
				"db":    dbOk,
				"redis": redisOk,
			})
		})

		// Auth routes (public, no JWT required)
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/signup/send-code", authController.SendSignUpCode)
			auth.Post("/signup", authController.SignUp)
			auth.Post("/signin/send-code", authController.SendSignInCode)
			auth.Post("/signin", authController.SignIn)
		})

		// Auth routes (protected, JWT required)
		router.Prefix("/auth").Middleware(middleware.JwtAuth()).Group(func(auth route.Router) {
			auth.Post("/refresh", authController.Refresh)
			auth.Get("/me", authController.Me)
			auth.Post("/logout", authController.Logout)
		})

		// Protected routes (user JWT required)
		router.Middleware(middleware.JwtAuth()).Group(func(protected route.Router) {
			// Protected endpoints will be added in Phases 2-9
		})
	})
}
