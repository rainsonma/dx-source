package api

import "errors"

var (
	ErrPostNotFound    = errors.New("帖子不存在")
	ErrPostNotOwner    = errors.New("无权操作此帖子")
	ErrCommentNotFound = errors.New("评论不存在")
	ErrCommentNotOwner = errors.New("无权操作此评论")
	ErrNestedReply     = errors.New("不能回复评论的回复")
	ErrSelfFollow      = errors.New("不能关注自己")
)
