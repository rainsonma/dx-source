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
	wechatAuthController := apicontrollers.NewWechatAuthController()
	emailController := apicontrollers.NewEmailController()
	userReferralController := apicontrollers.NewUserReferralController()

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

		// Public game routes (list + search only; detail is protected below)
		gameController := &apicontrollers.GameController{}
		router.Prefix("/games").Group(func(games route.Router) {
			games.Get("/", gameController.List)
			games.Get("/search", gameController.Search)
		})
		gameCategoryController := apicontrollers.NewGameCategoryController()
		router.Get("/game-categories", gameCategoryController.Categories)
		gamePressController := apicontrollers.NewGamePressController()
		router.Get("/game-presses", gamePressController.Presses)

		// Public upload serving (no auth required)
		router.Get("/uploads/images/{year}/{month}/{day}/{filename}", uploadController.ServeImage)

		// Auth routes (public, no JWT required)
		router.Prefix("/auth").Group(func(auth route.Router) {
			auth.Post("/signup", authController.SignUp)
			auth.Post("/signin", authController.SignIn)
			auth.Post("/logout", authController.Logout)
			auth.Post("/wechat-mini", wechatAuthController.MiniSignIn)
		})

		// Email verification code routes (public)
		router.Prefix("/email").Group(func(email route.Router) {
			email.Post("/send-signup-code", emailController.SendSignUpCode)
			email.Post("/send-signin-code", emailController.SendSignInCode)
		})

		// Public invite code validation (no auth required)
		router.Get("/invite/validate", userReferralController.ValidateCode)

		// Public group invite info (no auth required)
		publicGroupMemberController := apicontrollers.NewGroupMemberController()
		router.Get("/groups/invite/{code}", publicGroupMemberController.GetInviteInfo)

		// Public payment callbacks (no JWT required)
		paymentController := apicontrollers.NewPaymentController()
		router.Post("/payments/callback/wechat", paymentController.WechatCallback)
		router.Post("/payments/callback/alipay", paymentController.AlipayCallback)

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
			protected.Get("/games/{id}", gameController.Detail)
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

			// Game session routes (solo play)
			playSingleController := apicontrollers.NewGamePlaySingleController()
			protected.Prefix("/play-single").Group(func(sessions route.Router) {
				sessions.Post("/start", playSingleController.Start)
				sessions.Get("/active", playSingleController.CheckActive)
				sessions.Get("/any-active", playSingleController.CheckAnyActive)

				sessions.Post("/{id}/end", playSingleController.End)
				sessions.Post("/{id}/force-complete", playSingleController.ForceComplete)
				sessions.Post("/{id}/levels/{levelId}/complete", playSingleController.CompleteLevel)
				sessions.Post("/{id}/levels/{levelId}/restart", playSingleController.RestartLevel)
				sessions.Post("/{id}/answers", playSingleController.RecordAnswer)
				sessions.Post("/{id}/skips", playSingleController.RecordSkip)
				sessions.Post("/{id}/sync-playtime", playSingleController.SyncPlayTime)
				sessions.Get("/{id}/restore", playSingleController.Restore)
				sessions.Put("/{id}/content-item", playSingleController.UpdateContentItem)
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
			protected.Get("/hall/menus", hallController.GetMenus)

			// Invite & Referrals
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

			// Orders
			orderController := apicontrollers.NewOrderController()
			protected.Post("/orders/membership", orderController.CreateMembershipOrder)
			protected.Post("/orders/beans", orderController.CreateBeansOrder)
			protected.Get("/orders/{id}", paymentController.GetOrder)

			// Content Seek
			contentSeekController := apicontrollers.NewContentSeekController()
			protected.Get("/content-seek", contentSeekController.GetContentSeeks)
			protected.Post("/content-seek", contentSeekController.SubmitContentSeek)

			// AI custom content routes
			aiCustomController := apicontrollers.NewAiCustomController()
			aiCustomVocabController := apicontrollers.NewAiCustomVocabController()
			protected.Prefix("/ai-custom").Group(func(ai route.Router) {
				// Word-sentence endpoints
				ai.Post("/generate-metadata", aiCustomController.GenerateMetadata)
				ai.Post("/format-metadata", aiCustomController.FormatMetadata)
				ai.Post("/break-metadata", aiCustomController.BreakMetadata)
				ai.Post("/generate-content-items", aiCustomController.GenerateContentItems)
				// Vocab endpoints
				ai.Post("/generate-vocab", aiCustomVocabController.GenerateVocab)
				ai.Post("/format-vocab", aiCustomVocabController.FormatVocab)
				ai.Post("/break-vocab-metadata", aiCustomVocabController.BreakMetadata)
				ai.Post("/generate-vocab-content-items", aiCustomVocabController.GenerateContentItems)
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
				cg.Delete("/{id}/levels/{levelId}/metadata/{metaId}", courseGameController.DeleteMetadata)
				cg.Get("/{id}/levels/{levelId}/content-items", courseGameController.GetContentItems)
				cg.Post("/{id}/levels/{levelId}/content-items", courseGameController.InsertContentItem)
				cg.Put("/{id}/content-items/{itemId}", courseGameController.UpdateContentItemText)
				cg.Put("/{id}/content-items/reorder", courseGameController.ReorderContentItems)
				cg.Delete("/{id}/levels/{levelId}/content-items/{itemId}", courseGameController.DeleteContentItem)
				cg.Delete("/{id}/levels/{levelId}/content-items", courseGameController.DeleteAllLevelContent)
			})

			// Group game play routes
			playGroupController := apicontrollers.NewGamePlayGroupController()
			protected.Prefix("/play-group").Group(func(gp route.Router) {
				gp.Post("/start", playGroupController.Start)
				gp.Post("/{id}/levels/{levelId}/complete", playGroupController.CompleteLevel)
				gp.Post("/{id}/answers", playGroupController.RecordAnswer)
				gp.Post("/{id}/sync-playtime", playGroupController.SyncPlayTime)
				gp.Get("/{id}/restore", playGroupController.Restore)
				gp.Put("/{id}/content-item", playGroupController.UpdateContentItem)
			})

			// WebSocket
			wsController := apicontrollers.NewWSController()
			protected.Get("/ws", wsController.Handle)

			// User verify
			userVerifyController := apicontrollers.NewUserVerifyController()
			protected.Post("/users/verify-online", userVerifyController.VerifyOnline)

			// PK game play routes
			playPkController := apicontrollers.NewGamePlayPkController()
			protected.Prefix("/play-pk").Group(func(pk route.Router) {
				pk.Post("/start", playPkController.Start)
				pk.Post("/{id}/levels/{levelId}/complete", playPkController.CompleteLevel)
				pk.Post("/{id}/answers", playPkController.RecordAnswer)
				pk.Post("/{id}/sync-playtime", playPkController.SyncPlayTime)
				pk.Get("/{id}/restore", playPkController.Restore)
				pk.Put("/{id}/content-item", playPkController.UpdateContentItem)
				pk.Post("/{id}/end", playPkController.End)
				pk.Post("/{id}/next-level", playPkController.NextLevel)
				pk.Post("/{id}/pause", playPkController.Pause)
				pk.Post("/{id}/resume", playPkController.Resume)
			})

			// PK invitation routes
			pkInviteController := apicontrollers.NewPkInviteController()
			protected.Get("/play-pk/{id}/details", pkInviteController.Details)
			protected.Prefix("/play-pk/invite").Group(func(inv route.Router) {
				inv.Post("/", pkInviteController.Invite)
				inv.Post("/{id}/accept", pkInviteController.Accept)
				inv.Post("/{id}/decline", pkInviteController.Decline)
			})

			// Community
			postController := apicontrollers.NewPostController()
			postCommentController := apicontrollers.NewPostCommentController()
			postInteractController := apicontrollers.NewPostInteractController()
			followController := apicontrollers.NewFollowController()

			protected.Post("/posts", postController.Create)
			protected.Get("/posts", postController.List)
			protected.Get("/posts/{id}", postController.Show)
			protected.Put("/posts/{id}", postController.Update)
			protected.Delete("/posts/{id}", postController.Delete)

			protected.Post("/posts/{id}/comments", postCommentController.Create)
			protected.Get("/posts/{id}/comments", postCommentController.List)
			protected.Put("/posts/{id}/comments/{commentId}", postCommentController.Update)
			protected.Delete("/posts/{id}/comments/{commentId}", postCommentController.Delete)

			protected.Post("/posts/{id}/like", postInteractController.ToggleLike)
			protected.Post("/posts/{id}/bookmark", postInteractController.ToggleBookmark)

			protected.Post("/users/{id}/follow", followController.ToggleFollow)

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
				groups.Post("/{id}/dismiss", groupController.Dismiss)

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
				groups.Post("/{id}/start-game", groupGameController.StartGame)
				groups.Post("/{id}/force-end", groupGameController.ForceEnd)
				groups.Post("/{id}/next-level", groupGameController.NextLevel)
				groups.Get("/{id}/room-members", groupGameController.RoomMembers)
			})
		})
	})
}
