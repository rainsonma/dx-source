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

type ContentVocabController struct{}

func NewContentVocabController() *ContentVocabController {
	return &ContentVocabController{}
}

// GET /api/content-vocabs?content=<key>
func (c *ContentVocabController) GetByContent(ctx http.Context) http.Response {
	content := ctx.Request().Query("content", "")
	if content == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "content query param required")
	}
	v, err := services.GetContentVocabByContent(content)
	if err != nil {
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, err.Error())
	}
	return helpers.Success(ctx, v)
}

// POST /api/content-vocabs/{id}/complement
func (c *ContentVocabController) Complement(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.ComplementVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	vocabID := ctx.Request().Route("id")
	patch := services.VocabComplementPatch{
		Definition:  req.Definition,
		UkPhonetic:  req.UkPhonetic,
		UsPhonetic:  req.UsPhonetic,
		UkAudioURL:  req.UkAudioURL,
		UsAudioURL:  req.UsAudioURL,
		Explanation: req.Explanation,
	}
	v, err := services.ComplementContentVocab(userID, vocabID, patch)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, v)
}

// PUT /api/content-vocabs/{id}
func (c *ContentVocabController) Replace(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.ReplaceVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	vocabID := ctx.Request().Route("id")
	patch := services.VocabReplacePatch{
		Content:     req.Content,
		Definition:  req.Definition,
		UkPhonetic:  req.UkPhonetic,
		UsPhonetic:  req.UsPhonetic,
		UkAudioURL:  req.UkAudioURL,
		UsAudioURL:  req.UsAudioURL,
		Explanation: req.Explanation,
	}
	v, err := services.ReplaceContentVocab(userID, vocabID, patch)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, v)
}

// POST /api/content-vocabs/{id}/verify
func (c *ContentVocabController) Verify(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.VerifyVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	vocabID := ctx.Request().Route("id")
	v, err := services.VerifyContentVocab(userID, vocabID, req.Verified)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, v)
}

func mapVocabError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, services.ErrVocabNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeContentNotFound, "词条不存在")
	case errors.Is(err, services.ErrVocabNotEditable):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeForbidden, "无权编辑此词条")
	case errors.Is(err, services.ErrVocabAdminOnly):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeForbidden, "需要管理员权限")
	case errors.Is(err, services.ErrInvalidPosKey):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "definition 中包含无效词性")
	case errors.Is(err, services.ErrVocabContentEmpty), errors.Is(err, services.ErrVocabContentInvalid):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "词条内容无效")
	default:
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, err.Error())
	}
}
