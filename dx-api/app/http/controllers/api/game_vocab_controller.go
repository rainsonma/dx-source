package api

import (
	"errors"
	nethttp "net/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	apiReq "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type GameVocabController struct{}

func NewGameVocabController() *GameVocabController {
	return &GameVocabController{}
}

// POST /api/course-games/{id}/levels/{levelId}/game-vocabs
func (c *GameVocabController) Add(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.AddGameVocabsRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")

	added, err := services.AddVocabsToLevel(userID, gameID, levelID, req.VocabIDs)
	if err != nil {
		return mapGameVocabError(ctx, err)
	}
	return helpers.Success(ctx, added)
}

// GET /api/course-games/{id}/levels/{levelId}/game-vocabs
func (c *GameVocabController) List(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	rows, err := services.GetLevelVocabs(userID, gameID, levelID)
	if err != nil {
		return mapGameVocabError(ctx, err)
	}
	return helpers.Success(ctx, rows)
}

// PUT /api/course-games/{id}/game-vocabs/{gvId}/reorder
func (c *GameVocabController) Reorder(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.ReorderGameVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	gameID := ctx.Request().Route("id")
	gvID := ctx.Request().Route("gvId")
	if err := services.ReorderGameVocab(userID, gameID, gvID, req.NewOrder); err != nil {
		return mapGameVocabError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

// DELETE /api/course-games/{id}/game-vocabs/{gvId}
func (c *GameVocabController) Delete(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	gameID := ctx.Request().Route("id")
	gvID := ctx.Request().Route("gvId")
	if err := services.DeleteGameVocab(userID, gameID, gvID); err != nil {
		return mapGameVocabError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

func mapGameVocabError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, services.ErrGamePublished):
		return helpers.Error(ctx, nethttp.StatusConflict, consts.CodeValidationError, "已发布的游戏不可编辑，请先撤回")
	case errors.Is(err, services.ErrCapacityExceeded):
		return helpers.Error(ctx, nethttp.StatusUnprocessableEntity, consts.CodeValidationError, "容量已满")
	case errors.Is(err, services.ErrBatchSizeInvalid):
		return helpers.Error(ctx, nethttp.StatusUnprocessableEntity, consts.CodeValidationError, "数量必须是批次大小的倍数")
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeForbidden, "无权操作")
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrVocabContentEmpty), errors.Is(err, services.ErrVocabContentInvalid):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "词条内容无效（仅允许字母/数字/空格/' /-）")
	default:
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, err.Error())
	}
}
