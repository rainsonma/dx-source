package api

import (
	"errors"
	"net/http"
	"os"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// UploadImage handles POST /api/uploads/images — multipart file upload.
func (c *UploadController) UploadImage(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UploadImageRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	// Get the uploaded file via Goravel's request
	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "file is required")
	}

	// Validate file
	if err := services.ValidateUploadFile(file, req.Role); err != nil {
		switch {
		case errors.Is(err, services.ErrFileTooLarge):
			return helpers.Error(ctx, http.StatusRequestEntityTooLarge, consts.CodeFileTooLarge, "文件大小不能超过2MB")
		case errors.Is(err, services.ErrInvalidFileType):
			return helpers.Error(ctx, http.StatusUnsupportedMediaType, consts.CodeInvalidFileType, "仅支持JPEG和PNG格式")
		case errors.Is(err, services.ErrInvalidImageRole):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidImageRole, "无效的图片类型")
		default:
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, err.Error())
		}
	}

	// Upload
	result, err := services.UploadImage(userID, file, req.Role)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to upload image")
	}

	return helpers.Success(ctx, result)
}

// ServeImage handles GET /api/uploads/images/:id — serve an uploaded image.
func (c *UploadController) ServeImage(ctx contractshttp.Context) contractshttp.Response {
	imageID := ctx.Request().Route("id")
	if imageID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "image id is required")
	}

	absPath, contentType, err := services.GetImagePath(imageID)
	if err != nil {
		if errors.Is(err, services.ErrImageNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeImageNotFound, "图片不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get image")
	}

	// Check file exists on disk
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeImageNotFound, "图片文件不存在")
	}

	// Serve the file with caching headers (images are immutable — UUID names never change)
	return ctx.Response().
		Header("Content-Type", contentType).
		Header("Cache-Control", "public, max-age=31536000, immutable").
		File(absPath)
}
