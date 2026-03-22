package consts

// API error codes.
const (
	// Success
	CodeSuccess = 0

	// 400xx: Validation
	CodeValidationError  = 40000
	CodeInvalidEmail     = 40001
	CodeInvalidPassword  = 40002
	CodeDuplicateEmail   = 40003
	CodeDuplicateUsername = 40004
	CodeInvalidCode      = 40005
	CodeCodeExpired      = 40006
	CodeInsufficientBeans = 40007
	CodeNicknameTaken     = 40008

	// 401xx: Auth
	CodeUnauthorized = 40100
	CodeTokenExpired = 40101
	CodeInvalidToken        = 40102
	CodeInvalidRefreshToken = 40103
	CodeSessionReplaced     = 40104

	// 403xx: Permission
	CodeForbidden = 40300

	// 404xx: Not Found
	CodeNotFound        = 40400
	CodeUserNotFound    = 40401
	CodeGameNotFound    = 40402
	CodeSessionNotFound = 40403
	CodeLevelNotFound   = 40404
	CodeContentNotFound = 40405
	CodeImageNotFound   = 40406

	// 429xx: Rate Limit
	CodeRateLimited = 42900

	// 413xx: Payload Too Large
	CodeFileTooLarge = 41300

	// 415xx: Unsupported Media Type
	CodeInvalidFileType  = 41500
	CodeInvalidImageRole = 41501

	// 500xx: Server Error
	CodeInternalError  = 50000
	CodeAIServiceError = 50001
	CodeEmailSendError = 50002
)
