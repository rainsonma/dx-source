package api

import (
	"errors"
	"net/http"
	"os"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
	"dx-api/app/facades"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// UploadImage handles POST /api/uploads/images — multipart file upload.
func (c *UploadController) UploadImage(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	// Get the uploaded file via Goravel's request
	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "file is required")
	}

	// Get role from form data
	role := ctx.Request().Input("role")
	if role == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "role is required")
	}

	// Validate file
	if err := services.ValidateUploadFile(file, role); err != nil {
		switch {
		case errors.Is(err, services.ErrFileTooLarge):
			return helpers.Error(ctx, http.StatusRequestEntityTooLarge, constants.CodeFileTooLarge, "file size exceeds 2MB limit")
		case errors.Is(err, services.ErrInvalidFileType):
			return helpers.Error(ctx, http.StatusUnsupportedMediaType, constants.CodeInvalidFileType, "only JPEG and PNG files are allowed")
		case errors.Is(err, services.ErrInvalidImageRole):
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidImageRole, "invalid image role")
		default:
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, err.Error())
		}
	}

	// Upload
	result, err := services.UploadImage(userID, file, role)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to upload image")
	}

	return helpers.Success(ctx, result)
}

// ServeImage handles GET /api/uploads/images/:id — serve an uploaded image.
func (c *UploadController) ServeImage(ctx contractshttp.Context) contractshttp.Response {
	imageID := ctx.Request().Route("id")
	if imageID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "image id is required")
	}

	absPath, contentType, err := services.GetImagePath(imageID)
	if err != nil {
		if errors.Is(err, services.ErrImageNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeImageNotFound, "image not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get image")
	}

	// Check file exists on disk
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return helpers.Error(ctx, http.StatusNotFound, constants.CodeImageNotFound, "image file not found")
	}

	// Serve the file with caching headers (images are immutable — ULID names never change)
	return ctx.Response().
		Header("Content-Type", contentType).
		Header("Cache-Control", "public, max-age=31536000, immutable").
		File(absPath)
}
