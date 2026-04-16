package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type CourseGameController struct{}

func NewCourseGameController() *CourseGameController {
	return &CourseGameController{}
}

// List returns the authenticated user's own games.
func (c *CourseGameController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	status := ctx.Request().Query("status", "")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	games, nextCursor, hasMore, err := services.ListUserGames(userID, status, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list games")
	}

	return helpers.Paginated(ctx, games, nextCursor, hasMore)
}

// Counts returns game counts by status for the authenticated user.
func (c *CourseGameController) Counts(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	counts, err := services.GetUserGameCounts(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get counts")
	}

	return helpers.Success(ctx, counts)
}

// Create creates a new course game.
func (c *CourseGameController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	var categoryID *string
	if req.GameCategoryID != "" {
		categoryID = &req.GameCategoryID
	}
	var pressID *string
	if req.GamePressID != "" {
		pressID = &req.GamePressID
	}

	gameID, err := services.CreateGame(userID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, map[string]string{"id": gameID})
}

// Update updates a course game's properties.
func (c *CourseGameController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	var req requests.UpdateGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	var categoryID *string
	if req.GameCategoryID != "" {
		categoryID = &req.GameCategoryID
	}
	var pressID *string
	if req.GamePressID != "" {
		pressID = &req.GamePressID
	}

	err = services.UpdateGame(userID, gameID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Delete deletes a course game.
func (c *CourseGameController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	if err := services.DeleteGame(userID, gameID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Publish publishes a course game after validating readiness.
func (c *CourseGameController) Publish(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	if err := services.PublishGame(userID, gameID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Withdraw withdraws a published game.
func (c *CourseGameController) Withdraw(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	if err := services.WithdrawGame(userID, gameID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Detail returns a course game with levels for editing.
func (c *CourseGameController) Detail(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	detail, err := services.GetCourseGameDetail(userID, gameID)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, detail)
}

// CreateLevel adds a level to a game.
func (c *CourseGameController) CreateLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	var req requests.CreateLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	levelID, err := services.CreateLevel(userID, gameID, req.Name, req.Description)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, map[string]string{"id": levelID})
}

// DeleteLevel removes a level from a game.
func (c *CourseGameController) DeleteLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	if gameID == "" || levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and level id are required")
	}

	if err := services.DeleteLevel(userID, gameID, levelID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// SaveMetadata saves content metadata in batch.
func (c *CourseGameController) SaveMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	if gameID == "" || levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and level id are required")
	}

	var req requests.SaveMetadataBatchRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	// Convert request entries to service entries
	entries := make([]services.MetadataEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = services.MetadataEntry{
			SourceData:  e.SourceData,
			Translation: e.Translation,
			SourceType:  e.SourceType,
		}
	}

	count, err := services.SaveMetadataBatch(userID, gameID, levelID, entries, req.SourceFrom)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, map[string]int{"count": count})
}

// ReorderMetadata reorders content metadata.
func (c *CourseGameController) ReorderMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	var req requests.ReorderMetadataRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.ReorderMetadata(userID, gameID, req.GameLevelID, req.MetaID, req.NewOrder); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// DeleteMetadata removes a single metadata entry and its content items.
func (c *CourseGameController) DeleteMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	metaID := ctx.Request().Route("metaId")
	if gameID == "" || levelID == "" || metaID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id, level id and meta id are required")
	}

	if err := services.DeleteMetadata(userID, gameID, levelID, metaID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// GetContentItems returns content items grouped by metadata for a level.
func (c *CourseGameController) GetContentItems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	if gameID == "" || levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and level id are required")
	}

	data, err := services.GetContentItemsByMeta(userID, gameID, levelID)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, data)
}

// InsertContentItem inserts a content item at a position.
func (c *CourseGameController) InsertContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	if gameID == "" || levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and level id are required")
	}

	var req requests.InsertContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	item, err := services.InsertContentItem(userID, gameID, levelID, req.ContentMetaID, req.Content, req.ContentType, req.Translation, req.ReferenceItemID, req.Direction)
	if err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, item)
}

// UpdateContentItemText updates content item text and translation.
func (c *CourseGameController) UpdateContentItemText(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	itemID := ctx.Request().Route("itemId")
	if gameID == "" || itemID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and item id are required")
	}

	var req requests.UpdateContentItemTextRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if err := services.UpdateContentItemText(userID, gameID, itemID, req.Content, req.Translation); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ReorderContentItems reorders a content item.
func (c *CourseGameController) ReorderContentItems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	var req requests.ReorderContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.ReorderContentItems(userID, gameID, req.GameLevelID, req.ItemID, req.NewOrder); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// DeleteContentItem removes a single content item.
func (c *CourseGameController) DeleteContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	itemID := ctx.Request().Route("itemId")
	if gameID == "" || levelID == "" || itemID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id, level id and item id are required")
	}

	if err := services.DeleteContentItem(userID, gameID, levelID, itemID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// DeleteAllLevelContent removes all content from a level.
func (c *CourseGameController) DeleteAllLevelContent(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	if gameID == "" || levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id and level id are required")
	}

	if err := services.DeleteAllLevelContent(userID, gameID, levelID); err != nil {
		return mapCourseGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// mapCourseGameError maps service errors to HTTP responses.
func mapCourseGameError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "游戏不存在或无权操作")
	case errors.Is(err, services.ErrGamePublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "已发布的游戏不可编辑，请先撤回")
	case errors.Is(err, services.ErrGameAlreadyPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏已经是发布状态")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "只有已发布的游戏可以撤回")
	case errors.Is(err, services.ErrNoGameLevels):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏至少需要一个关卡才能发布")
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	case errors.Is(err, services.ErrMetaNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeContentNotFound, "元数据不存在或无权操作")
	case errors.Is(err, services.ErrContentItemNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeContentNotFound, "练习单元不存在或无权操作")
	case errors.Is(err, services.ErrCapacityExceeded):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "超出关卡内容上限")
	case errors.Is(err, services.ErrItemLimitExceeded):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "每条元数据练习单元数量已达上限")
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrGameNameTaken):
		return helpers.Error(ctx, http.StatusConflict, consts.CodeValidationError, "游戏名称已存在，请换一个名称")
	default:
		// Pass through Chinese validation messages (e.g. publish: level has no content);
		// hide raw internal errors from clients.
		msg := err.Error()
		if len(msg) > 0 && msg[0] > 127 {
			// Starts with a multibyte char — likely a Chinese user-facing message
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, msg)
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}
