package api

import (
	"errors"
	nethttp "net/http"
	"strconv"

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

// GET /api/content-vocabs/mine
func (c *ContentVocabController) ListMine(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	cursor := ctx.Request().Query("cursor", "")
	search := ctx.Request().Query("search", "")
	limitStr := ctx.Request().Query("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	items, nextCursor, hasMore, err := services.ListUserVocabs(userID, cursor, search, limit)
	if err != nil {
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, err.Error())
	}
	return helpers.Success(ctx, map[string]any{
		"items":      items,
		"nextCursor": nextCursor,
		"hasMore":    hasMore,
	})
}

// POST /api/content-vocabs
func (c *ContentVocabController) Create(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.CreateVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	in := services.VocabInput{
		Content:     req.Content,
		Definition:  req.Definition,
		UkPhonetic:  req.UkPhonetic,
		UsPhonetic:  req.UsPhonetic,
		UkAudioURL:  req.UkAudioURL,
		UsAudioURL:  req.UsAudioURL,
		Explanation: req.Explanation,
	}
	result, err := services.CreateUserVocab(userID, in)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, result)
}

// POST /api/content-vocabs/batch
func (c *ContentVocabController) CreateBatch(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.CreateVocabsBatchRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	inputs := make([]services.VocabInput, 0, len(req.Inputs))
	for _, r := range req.Inputs {
		inputs = append(inputs, services.VocabInput{
			Content:     r.Content,
			Definition:  r.Definition,
			UkPhonetic:  r.UkPhonetic,
			UsPhonetic:  r.UsPhonetic,
			UkAudioURL:  r.UkAudioURL,
			UsAudioURL:  r.UsAudioURL,
			Explanation: r.Explanation,
		})
	}
	results, err := services.CreateUserVocabsBatch(userID, inputs)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, results)
}

// PUT /api/content-vocabs/{id}
func (c *ContentVocabController) Update(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.UpdateVocabRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	vocabID := ctx.Request().Route("id")
	in := services.VocabInput{
		Content:     req.Content,
		Definition:  req.Definition,
		UkPhonetic:  req.UkPhonetic,
		UsPhonetic:  req.UsPhonetic,
		UkAudioURL:  req.UkAudioURL,
		UsAudioURL:  req.UsAudioURL,
		Explanation: req.Explanation,
	}
	v, err := services.UpdateUserVocab(userID, vocabID, in)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, v)
}

// DELETE /api/content-vocabs/{id}
func (c *ContentVocabController) Delete(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	vocabID := ctx.Request().Route("id")
	if err := services.DeleteUserVocab(userID, vocabID); err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, nil)
}

// POST /api/content-vocabs/from-words
func (c *ContentVocabController) CreateFromWords(ctx http.Context) http.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.CreateVocabsFromWordsRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	if len(req.Words) == 0 || len(req.Words) > 50 {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "请提供1-50个词汇")
	}

	results, err := services.CreateVocabsFromWords(userID, req.Words)
	if err != nil {
		return mapVocabError(ctx, err)
	}
	return helpers.Success(ctx, results)
}

func mapVocabError(ctx http.Context, err error) http.Response {
	switch {
	case errors.Is(err, services.ErrVocabNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeContentNotFound, "词条不存在")
	case errors.Is(err, services.ErrDuplicateVocab):
		return helpers.Error(ctx, nethttp.StatusConflict, consts.CodeValidationError, "该词条已存在")
	case errors.Is(err, services.ErrInvalidPosKey):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "definition 中包含无效词性")
	case errors.Is(err, services.ErrVocabContentEmpty), errors.Is(err, services.ErrVocabContentInvalid):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "词条内容无效")
	default:
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, err.Error())
	}
}
