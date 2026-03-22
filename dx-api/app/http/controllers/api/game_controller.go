package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type GameController struct{}

// List returns published games with cursor pagination and optional filters.
func (c *GameController) List(ctx contractshttp.Context) contractshttp.Response {
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	categoryIDsRaw := ctx.Request().Query("categoryIds", "")
	var categoryIDs []string
	if categoryIDsRaw != "" {
		categoryIDs = strings.Split(categoryIDsRaw, ",")
	}

	pressID := ctx.Request().Query("pressId", "")
	mode := ctx.Request().Query("mode", "")

	games, nextCursor, hasMore, err := services.ListPublishedGames(cursor, limit, categoryIDs, pressID, mode)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to list games")
	}

	return helpers.Paginated(ctx, games, nextCursor, hasMore)
}

// Search returns published games matching a name query.
func (c *GameController) Search(ctx contractshttp.Context) contractshttp.Response {
	query := ctx.Request().Query("q", "")
	if strings.TrimSpace(query) == "" {
		return helpers.Success(ctx, []any{})
	}

	limitStr := ctx.Request().Query("limit", "")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	games, err := services.SearchGames(strings.TrimSpace(query), limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to search games")
	}

	return helpers.Success(ctx, games)
}

// Recent returns the authenticated user's recently played games.
func (c *GameController) Recent(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	games, err := services.GetRecentGames(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get recent games")
	}

	return helpers.Success(ctx, games)
}

// Detail returns full game detail with levels.
func (c *GameController) Detail(ctx contractshttp.Context) contractshttp.Response {
	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "game id is required")
	}

	detail, err := services.GetGameDetail(gameID)
	if err != nil {
		if errors.Is(err, services.ErrGameNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeGameNotFound, "game not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get game detail")
	}

	return helpers.Success(ctx, detail)
}

// Categories returns all enabled categories in hierarchical order.
func (c *GameController) Categories(ctx contractshttp.Context) contractshttp.Response {
	categories, err := services.ListCategories()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to list categories")
	}

	return helpers.Success(ctx, categories)
}

// Presses returns all game publishers.
func (c *GameController) Presses(ctx contractshttp.Context) contractshttp.Response {
	presses, err := services.ListPresses()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to list presses")
	}

	return helpers.Success(ctx, presses)
}
