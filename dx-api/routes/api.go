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

	uploadController := apicontrollers.NewUploadController()

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

		// Public upload serving (no auth required)
		router.Get("/uploads/images/{id}", uploadController.ServeImage)

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
			// Upload routes
			protected.Post("/uploads/images", uploadController.UploadImage)

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

			// Game session routes
			sessionController := apicontrollers.NewGameSessionController()
			protected.Prefix("/sessions").Group(func(sessions route.Router) {
				sessions.Post("/start", sessionController.Start)
				sessions.Get("/active", sessionController.CheckActive)
				sessions.Get("/active-level", sessionController.CheckActiveLevel)
				sessions.Get("/any-active", sessionController.CheckAnyActive)

				sessions.Post("/{id}/end", sessionController.End)
				sessions.Post("/{id}/force-complete", sessionController.ForceComplete)
				sessions.Post("/{id}/levels/start", sessionController.StartLevel)
				sessions.Post("/{id}/levels/{levelId}/complete", sessionController.CompleteLevel)
				sessions.Post("/{id}/levels/{levelId}/advance", sessionController.AdvanceLevel)
				sessions.Post("/{id}/levels/{levelId}/restart", sessionController.RestartLevel)
				sessions.Post("/{id}/answers", sessionController.RecordAnswer)
				sessions.Post("/{id}/skips", sessionController.RecordSkip)
				sessions.Post("/{id}/sync-playtime", sessionController.SyncPlayTime)
				sessions.Get("/{id}/restore", sessionController.Restore)
				sessions.Put("/{id}/content-item", sessionController.UpdateContentItem)
			})

			// Tracking routes (mastered / unknown / review)
			trackingController := apicontrollers.NewTrackingController()
			protected.Prefix("/tracking").Group(func(tracking route.Router) {
				// Mastered
				tracking.Post("/master", trackingController.MarkMastered)
				tracking.Get("/master", trackingController.ListMastered)
				tracking.Get("/master/stats", trackingController.MasterStats)
				tracking.Delete("/master/{id}", trackingController.DeleteMastered)
				tracking.Delete("/master", trackingController.BulkDeleteMastered)

				// Unknown
				tracking.Post("/unknown", trackingController.MarkUnknown)
				tracking.Get("/unknown", trackingController.ListUnknown)
				tracking.Get("/unknown/stats", trackingController.UnknownStats)
				tracking.Delete("/unknown/{id}", trackingController.DeleteUnknown)
				tracking.Delete("/unknown", trackingController.BulkDeleteUnknown)

				// Review
				tracking.Post("/review", trackingController.MarkReview)
				tracking.Get("/review", trackingController.ListReviews)
				tracking.Get("/review/stats", trackingController.ReviewStats)
				tracking.Delete("/review/{id}", trackingController.DeleteReview)
				tracking.Delete("/review", trackingController.BulkDeleteReviews)
			})

			// Favorites routes
			protected.Post("/favorites/toggle", trackingController.ToggleFavorite)
			protected.Get("/favorites", trackingController.ListFavorites)

			// Community & Social routes
			communityController := apicontrollers.NewCommunityController()
			protected.Get("/leaderboard", communityController.GetLeaderboard)

			// Hall routes
			protected.Get("/hall/dashboard", communityController.GetDashboard)
			protected.Get("/hall/heatmap", communityController.GetHeatmap)

			// Invite & Referrals
			protected.Get("/invite", communityController.GetInviteData)
			protected.Get("/referrals", communityController.GetReferrals)

			// Notices
			protected.Get("/notices", communityController.GetNotices)
			protected.Post("/notices/mark-read", communityController.MarkNoticesRead)

			// Feedback & Reports
			protected.Post("/feedback", communityController.SubmitFeedback)
			protected.Post("/reports", communityController.SubmitReport)

			// Redeems
			protected.Get("/redeems", communityController.GetRedeems)
			protected.Post("/redeems", communityController.RedeemCode)

			// Content Seek
			protected.Get("/content-seek", communityController.GetContentSeeks)
			protected.Post("/content-seek", communityController.SubmitContentSeek)

			// AI custom content routes
			aiCustomController := apicontrollers.NewAiCustomController()
			protected.Prefix("/ai-custom").Group(func(ai route.Router) {
				ai.Post("/generate-metadata", aiCustomController.GenerateMetadata)
				ai.Post("/format-metadata", aiCustomController.FormatMetadata)
				ai.Post("/break-metadata", aiCustomController.BreakMetadata)
				ai.Post("/generate-content-items", aiCustomController.GenerateContentItems)
			})

			// Course game management routes
			courseGameController := apicontrollers.NewCourseGameController()
			protected.Prefix("/course-games").Group(func(cg route.Router) {
				cg.Get("/", courseGameController.List)
				cg.Get("/counts", courseGameController.Counts)
				cg.Post("/", courseGameController.Create)
				cg.Get("/{id}", courseGameController.Detail)
				cg.Put("/{id}", courseGameController.Update)
				cg.Delete("/{id}", courseGameController.Delete)
				cg.Post("/{id}/publish", courseGameController.Publish)
				cg.Post("/{id}/withdraw", courseGameController.Withdraw)
				cg.Post("/{id}/levels", courseGameController.CreateLevel)
				cg.Delete("/{id}/levels/{levelId}", courseGameController.DeleteLevel)
				cg.Post("/{id}/levels/{levelId}/metadata", courseGameController.SaveMetadata)
				cg.Put("/{id}/metadata/reorder", courseGameController.ReorderMetadata)
				cg.Get("/{id}/levels/{levelId}/content-items", courseGameController.GetContentItems)
				cg.Post("/{id}/levels/{levelId}/content-items", courseGameController.InsertContentItem)
				cg.Put("/{id}/content-items/{itemId}", courseGameController.UpdateContentItemText)
				cg.Put("/{id}/content-items/reorder", courseGameController.ReorderContentItems)
				cg.Delete("/{id}/content-items/{itemId}", courseGameController.DeleteContentItem)
				cg.Delete("/{id}/levels/{levelId}/content-items", courseGameController.DeleteAllLevelContent)
			})
		})
	})
}
