package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type CommunityController struct{}

func NewCommunityController() *CommunityController {
	return &CommunityController{}
}

// GetLeaderboard returns the leaderboard by type and period.
func (c *CommunityController) GetLeaderboard(ctx contractshttp.Context) contractshttp.Response {
	userID, _ := facades.Auth(ctx).Guard("user").ID()

	lbType := ctx.Request().Query("type", "exp")
	period := ctx.Request().Query("period", "all")

	if lbType != "exp" && lbType != "playtime" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "type must be exp or playtime")
	}
	if period != "all" && period != "day" && period != "week" && period != "month" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "period must be all, day, week, or month")
	}

	result, err := services.GetLeaderboard(lbType, period, userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get leaderboard")
	}

	return helpers.Success(ctx, result)
}

// GetDashboard returns aggregated dashboard data.
func (c *CommunityController) GetDashboard(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	data, err := services.GetDashboard(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get dashboard")
	}

	return helpers.Success(ctx, data)
}

// GetHeatmap returns daily activity counts for a year.
func (c *CommunityController) GetHeatmap(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	yearStr := ctx.Request().Query("year", "")
	year := 0
	if yearStr != "" {
		parsed, err := strconv.Atoi(yearStr)
		if err != nil || parsed < 2000 || parsed > 2100 {
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid year")
		}
		year = parsed
	}
	if year == 0 {
		year = currentYear()
	}

	data, err := services.GetHeatmap(userID, year)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get heatmap")
	}

	return helpers.Success(ctx, data)
}

// GetInviteData returns invite code, stats, and first page of referrals.
func (c *CommunityController) GetInviteData(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	data, err := services.GetInviteData(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get invite data")
	}

	return helpers.Success(ctx, data)
}

// GetReferrals returns paginated referral records.
func (c *CommunityController) GetReferrals(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	referrals, err := services.GetReferrals(userID, page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get referrals")
	}

	total, err := services.CountReferrals(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to count referrals")
	}

	return helpers.PaginatedOffset(ctx, referrals, total, page, pageSize)
}

// GetNotices returns active notices with cursor pagination.
func (c *CommunityController) GetNotices(ctx contractshttp.Context) contractshttp.Response {
	cursor, limit := helpers.ParseCursorParams(ctx, 20)

	items, nextCursor, hasMore, err := services.GetNotices(cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get notices")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// MarkNoticesRead updates the user's last_read_notice_at.
func (c *CommunityController) MarkNoticesRead(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	if err := services.MarkNoticesRead(userID); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to mark notices read")
	}

	return helpers.Success(ctx, nil)
}

// SubmitFeedback creates a feedback record.
func (c *CommunityController) SubmitFeedback(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.SubmitFeedbackRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Type == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "type is required")
	}
	if req.Description == "" || len(req.Description) > 200 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "description must be 1-200 characters")
	}

	result, err := services.SubmitFeedback(userID, req.Type, req.Description)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to submit feedback")
	}

	return helpers.Success(ctx, result)
}

// SubmitReport creates a game content report.
func (c *CommunityController) SubmitReport(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.SubmitReportRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.GameID == "" || req.GameLevelID == "" || req.ContentItemID == "" || req.Reason == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "game_id, game_level_id, content_item_id, and reason are required")
	}

	result, err := services.SubmitReport(userID, req.GameID, req.GameLevelID, req.ContentItemID, req.Reason, req.Note)
	if err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, constants.CodeRateLimited, "too many reports, please try again later")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to submit report")
	}

	return helpers.Success(ctx, result)
}

// GetRedeems returns the user's redemption records.
func (c *CommunityController) GetRedeems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	items, total, err := services.GetRedeems(userID, page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get redeems")
	}

	return helpers.PaginatedOffset(ctx, items, total, page, pageSize)
}

// RedeemCode processes a redemption code.
func (c *CommunityController) RedeemCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.RedeemCodeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Code == "" || len(req.Code) != 19 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid redeem code format")
	}

	result, err := services.RedeemCode(userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRedeemNotFound):
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeNotFound, "redeem code not found")
		case errors.Is(err, services.ErrRedeemAlreadyUsed):
			return helpers.Error(ctx, http.StatusConflict, constants.CodeValidationError, "redeem code already used")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to redeem code")
		}
	}

	return helpers.Success(ctx, result)
}

// GetContentSeeks returns the user's content seek records.
func (c *CommunityController) GetContentSeeks(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	items, err := services.GetContentSeeks(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get content seeks")
	}

	return helpers.Success(ctx, items)
}

// SubmitContentSeek creates or updates a content seek record.
func (c *CommunityController) SubmitContentSeek(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.SubmitContentSeekRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.CourseName == "" || len(req.CourseName) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "course_name must be 1-30 characters")
	}
	if req.Description == "" || len(req.Description) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "description must be 1-30 characters")
	}
	if req.DiskUrl == "" || len(req.DiskUrl) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "disk_url must be 1-30 characters")
	}

	result, err := services.SubmitContentSeek(userID, req.CourseName, req.Description, req.DiskUrl)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to submit content seek")
	}

	return helpers.Success(ctx, result)
}

// currentYear returns the current year.
func currentYear() int {
	return time.Now().Year()
}
