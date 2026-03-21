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

		// Public game routes
		gameController := &apicontrollers.GameController{}
		router.Prefix("/games").Group(func(games route.Router) {
			games.Get("/", gameController.List)
			games.Get("/search", gameController.Search)
			games.Get("/{id}", gameController.Detail)
		})
		router.Get("/game-categories", gameController.Categories)
		router.Get("/game-presses", gameController.Presses)

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
			// Protected game routes
			contentController := &apicontrollers.ContentController{}
			protected.Get("/games/recent", gameController.Recent)
			protected.Get("/games/{id}/levels/{levelId}/content", contentController.LevelContent)

			// User profile routes
			userController := &apicontrollers.UserController{}
			protected.Prefix("/user").Group(func(user route.Router) {
				user.Get("/profile", userController.GetProfile)
				user.Put("/profile", userController.UpdateProfile)
				user.Put("/avatar", userController.UpdateAvatar)
				user.Post("/email/send-code", userController.SendEmailCode)
				user.Put("/email", userController.ChangeEmail)
				user.Put("/password", userController.ChangePassword)
			})
		})
	})
}
