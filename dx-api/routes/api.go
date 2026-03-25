package routes

import (
	"dx-api/app/helpers"
	admcontrollers "dx-api/app/http/controllers/adm"
	apicontrollers "dx-api/app/http/controllers/api"
	"dx-api/app/http/middleware"

	"github.com/goravel/framework/facades"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
)

func Api() {
	r := facades.Route()

	authController := apicontrollers.NewAuthController()
	emailController := apicontrollers.NewEmailController()

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
		gameCategoryController := apicontrollers.NewGameCategoryController()
		router.Get("/game-categories", gameCategoryController.Categories)
		gamePressController := apicontrollers.NewGamePressController()
		router.Get("/game-presses", gamePressController.Presses)

		// Public upload serving (no auth required)
		router.Get("/uploads/images/{id}", uploadController.ServeImage)

		// Auth routes (public, no JWT required)
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/signup", authController.SignUp)
			auth.Post("/signin", authController.SignIn)
			auth.Post("/refresh", authController.Refresh)
			auth.Post("/logout", authController.Logout)
		})

		// Email verification code routes (public)
		router.Prefix("/email").Group(func(email route.Router) {
			email.Post("/send-signup-code", emailController.SendSignUpCode)
			email.Post("/send-signin-code", emailController.SendSignInCode)
		})

		// Public group invite info (no auth required)
		publicGroupMemberController := apicontrollers.NewGroupMemberController()
		router.Get("/groups/invite/{code}", publicGroupMemberController.GetInviteInfo)

		// Auth routes (protected, JWT required)
		router.Prefix("/auth").Middleware(middleware.JwtAuth()).Group(func(auth route.Router) {
			auth.Get("/me", authController.Me)
		})

		// Protected routes (user JWT required)
		router.Middleware(middleware.JwtAuth()).Group(func(protected route.Router) {
			// Upload routes
			protected.Post("/uploads/images", uploadController.UploadImage)

			// Protected game routes
			contentController := &apicontrollers.ContentController{}
			gameStatsController := apicontrollers.NewGameStatsController()
			userFavoriteController := apicontrollers.NewUserFavoriteController()
			protected.Get("/games/played", gameController.Played)
			protected.Get("/games/{id}/stats", gameStatsController.Stats)
			protected.Get("/games/{id}/favorited", userFavoriteController.Favorited)
			protected.Get("/games/{id}/levels/{levelId}/content", contentController.LevelContent)

			// User profile routes
			userController := &apicontrollers.UserController{}
			protected.Prefix("/user").Group(func(user route.Router) {
				user.Get("/profile", userController.GetProfile)
				user.Put("/profile", userController.UpdateProfile)
				user.Put("/avatar", userController.UpdateAvatar)
				user.Put("/email", userController.ChangeEmail)
				user.Put("/password", userController.ChangePassword)
			})

			// Email verification code route (protected)
			protected.Post("/email/send-change-code", emailController.SendChangeCode)

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
			userMasterController := apicontrollers.NewUserMasterController()
			userUnknownController := apicontrollers.NewUserUnknownController()
			userReviewController := apicontrollers.NewUserReviewController()
			protected.Prefix("/tracking").Group(func(tracking route.Router) {
				// Mastered
				tracking.Post("/master", userMasterController.MarkMastered)
				tracking.Get("/master", userMasterController.ListMastered)
				tracking.Get("/master/stats", userMasterController.MasterStats)
				tracking.Delete("/master/{id}", userMasterController.DeleteMastered)
				tracking.Delete("/master", userMasterController.BulkDeleteMastered)

				// Unknown
				tracking.Post("/unknown", userUnknownController.MarkUnknown)
				tracking.Get("/unknown", userUnknownController.ListUnknown)
				tracking.Get("/unknown/stats", userUnknownController.UnknownStats)
				tracking.Delete("/unknown/{id}", userUnknownController.DeleteUnknown)
				tracking.Delete("/unknown", userUnknownController.BulkDeleteUnknown)

				// Review
				tracking.Post("/review", userReviewController.MarkReview)
				tracking.Get("/review", userReviewController.ListReviews)
				tracking.Get("/review/stats", userReviewController.ReviewStats)
				tracking.Delete("/review/{id}", userReviewController.DeleteReview)
				tracking.Delete("/review", userReviewController.BulkDeleteReviews)
			})

			// Favorites routes
			protected.Post("/favorites/toggle", userFavoriteController.ToggleFavorite)
			protected.Get("/favorites", userFavoriteController.ListFavorites)

			// Leaderboard
			leaderboardController := apicontrollers.NewLeaderboardController()
			protected.Get("/leaderboard", leaderboardController.GetLeaderboard)

			// Hall routes
			hallController := apicontrollers.NewHallController()
			protected.Get("/hall/dashboard", hallController.GetDashboard)
			protected.Get("/hall/heatmap", hallController.GetHeatmap)

			// Invite & Referrals
			userReferralController := apicontrollers.NewUserReferralController()
			protected.Get("/invite", userReferralController.GetInviteData)
			protected.Get("/referrals", userReferralController.GetReferrals)

			// Notices
			noticeController := apicontrollers.NewNoticeController()
			protected.Get("/notices", noticeController.GetNotices)
			protected.Post("/notices/mark-read", noticeController.MarkNoticesRead)

			// Feedback & Reports
			feedbackController := apicontrollers.NewFeedbackController()
			protected.Post("/feedback", feedbackController.SubmitFeedback)
			gameReportController := apicontrollers.NewGameReportController()
			protected.Post("/reports", gameReportController.SubmitReport)

			// Redeems
			userRedeemController := apicontrollers.NewUserRedeemController()
			protected.Get("/redeems", userRedeemController.GetRedeems)
			protected.Post("/redeems", userRedeemController.RedeemCode)

			// Content Seek
			contentSeekController := apicontrollers.NewContentSeekController()
			protected.Get("/content-seek", contentSeekController.GetContentSeeks)
			protected.Post("/content-seek", contentSeekController.SubmitContentSeek)

			// AI custom content routes
			aiCustomController := apicontrollers.NewAiCustomController()
			protected.Prefix("/ai-custom").Group(func(ai route.Router) {
				ai.Post("/generate-metadata", aiCustomController.GenerateMetadata)
				ai.Post("/format-metadata", aiCustomController.FormatMetadata)
				ai.Post("/break-metadata", aiCustomController.BreakMetadata)
				ai.Post("/generate-content-items", aiCustomController.GenerateContentItems)
			})

			// User-facing admin routes (user JWT + admin check)
			protected.Middleware(middleware.AdminGuard()).Group(func(admin route.Router) {
				admNoticeController := admcontrollers.NewNoticeController()
				admRedeemController := admcontrollers.NewRedeemController()

				// Notice management
				admin.Post("/admin/notices", admNoticeController.CreateNotice)
				admin.Put("/admin/notices/{id}", admNoticeController.UpdateNotice)
				admin.Delete("/admin/notices/{id}", admNoticeController.DeleteNotice)

				// Redeem management
				admin.Post("/admin/redeems/generate", admRedeemController.GenerateCodes)
				admin.Get("/admin/redeems", admRedeemController.GetAllRedeems)
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

			// Group routes
			groupController := apicontrollers.NewGroupController()
			groupMemberController := apicontrollers.NewGroupMemberController()
			groupSubgroupController := apicontrollers.NewGroupSubgroupController()

			protected.Post("/groups/join/{code}", groupMemberController.JoinByCode)
			protected.Prefix("/groups").Group(func(groups route.Router) {
				groups.Get("/", groupController.List)
				groups.Post("/", groupController.Create)
				groups.Get("/{id}", groupController.Detail)
				groups.Put("/{id}", groupController.Update)
				groups.Delete("/{id}", groupController.Delete)

				// Applications
				groups.Post("/{id}/apply", groupController.Apply)
				groups.Delete("/{id}/apply", groupController.CancelApply)
				groups.Get("/{id}/applications", groupController.ListApplications)
				groups.Put("/{id}/applications/{appId}", groupController.HandleApplication)

				// Members
				groups.Get("/{id}/members", groupMemberController.List)
				groups.Delete("/{id}/members/{userId}", groupMemberController.Kick)
				groups.Post("/{id}/leave", groupMemberController.Leave)

				// Subgroups
				groups.Get("/{id}/subgroups", groupSubgroupController.List)
				groups.Post("/{id}/subgroups", groupSubgroupController.Create)
				groups.Put("/{id}/subgroups/{sid}", groupSubgroupController.Update)
				groups.Delete("/{id}/subgroups/{sid}", groupSubgroupController.Delete)
				groups.Get("/{id}/subgroups/{sid}/members", groupSubgroupController.ListMembers)
				groups.Post("/{id}/subgroups/{sid}/members", groupSubgroupController.Assign)
				groups.Delete("/{id}/subgroups/{sid}/members/{userId}", groupSubgroupController.RemoveMember)

				// Group game
				groupGameController := apicontrollers.NewGroupGameController()
				groups.Get("/{id}/games/search", groupGameController.SearchGames)
				groups.Put("/{id}/game", groupGameController.SetGame)
				groups.Delete("/{id}/game", groupGameController.ClearGame)
			})
		})
	})
}
