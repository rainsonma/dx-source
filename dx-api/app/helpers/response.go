package helpers

import (
	nethttp "net/http"

	"dx-api/app/consts"

	"github.com/goravel/framework/contracts/http"
)

// Response envelope
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// PaginatedData for cursor-based pagination
type CursorPaginatedData struct {
	Items      any    `json:"items"`
	NextCursor string `json:"nextCursor"`
	HasMore    bool   `json:"hasMore"`
}

// PaginatedData for offset-based pagination
type OffsetPaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// Success returns a success response
func Success(ctx http.Context, data any) http.Response {
	return ctx.Response().Success().Json(Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Error returns an error response
func Error(ctx http.Context, httpStatus int, code int, message string) http.Response {
	return ctx.Response().Status(httpStatus).Json(Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Paginated returns a cursor-paginated success response
func Paginated(ctx http.Context, items any, nextCursor string, hasMore bool) http.Response {
	return Success(ctx, CursorPaginatedData{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

// PaginatedOffset returns an offset-paginated success response
func PaginatedOffset(ctx http.Context, items any, total int64, page int, pageSize int) http.Response {
	return Success(ctx, OffsetPaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Validate runs Goravel form request validation and returns an error response on failure.
// Returns nil when validation passes (struct populated, proceed with handler logic).
func Validate(ctx http.Context, req http.FormRequest) http.Response {
	errors, err := ctx.Request().ValidateRequest(req)
	if err != nil {
		return Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
	}
	if errors != nil {
		return Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, errors.One())
	}
	return nil
}
