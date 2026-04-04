package consts

// API error codes.
const (
	// Success
	CodeSuccess = 0

	// 400xx: Validation
	CodeValidationError    = 40000
	CodeInvalidEmail       = 40001
	CodeInvalidPassword    = 40002
	CodeDuplicateEmail     = 40003
	CodeDuplicateUsername  = 40004
	CodeInvalidCode        = 40005
	CodeCodeExpired        = 40006
	CodeInsufficientBeans  = 40007
	CodeNicknameTaken      = 40008
	CodeAlreadyMember      = 40009
	CodeAlreadyApplied     = 40010
	CodeGroupMembersFull   = 40011
	CodeGroupSubgroupsFull = 40012
	CodeOrderNotPending    = 40013
	CodeInvalidProduct     = 40014
	CodePkIsPlaying        = 40015
	CodePkNotPlaying       = 40016
	CodeNoMockUser         = 40017

	// 401xx: Auth
	CodeUnauthorized    = 40100
	CodeInvalidToken    = 40102
	CodeSessionReplaced = 40104

	// 403xx: Permission
	CodeForbidden      = 40300
	CodeGroupForbidden = 40301
	CodeVipRequired    = 40302

	// 404xx: Not Found
	CodeNotFound            = 40400
	CodeUserNotFound        = 40401
	CodeGameNotFound        = 40402
	CodeSessionNotFound     = 40403
	CodeLevelNotFound       = 40404
	CodeContentNotFound     = 40405
	CodeImageNotFound       = 40406
	CodeGroupNotFound       = 40407
	CodeApplicationNotFound = 40408
	CodePostNotFound        = 40409
	CodeCommentNotFound     = 40410
	CodeOrderNotFound       = 40411
	CodePkNotFound          = 40412

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
