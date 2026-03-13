package errs

import (
	"net/http"

	"github.com/lpphub/goweb/base"
)

var (
	// 系统错误
	ErrServerError = base.NewErrorWithStatus(500, "server internal error", http.StatusInternalServerError)

	// 通用错误
	ErrNoToken      = base.NewErrorWithStatus(1000, "no token", http.StatusUnauthorized)
	ErrInvalidToken = base.NewErrorWithStatus(1001, "invalid token", http.StatusUnauthorized)
	ErrInvalidParam = base.NewError(1100, "参数错误")

	// 业务错误
	ErrUserExists      = base.NewError(2101, "用户已存在")
	ErrUserNotFound    = base.NewError(2102, "用户不存在")
	ErrInvalidPassword = base.NewError(2103, "密码错误")
	ErrLoginFailed     = base.NewError(2104, "登录失败")
	ErrUserDisabled    = base.NewError(2105, "用户已禁用")

	// 视频错误
	ErrVideoNotFound      = base.NewError(2201, "视频不存在")
	ErrVideoNotPublished  = base.NewError(2202, "视频未发布")
	ErrCategoryNotFound   = base.NewError(2203, "分类不存在")
	ErrCategoryExists     = base.NewError(2204, "分类已存在")
	ErrTagNotFound        = base.NewError(2205, "标签不存在")
	ErrTagExists          = base.NewError(2206, "标签已存在")
	ErrTranscodeNotFound  = base.NewError(2207, "转码任务不存在")
	ErrTranscodeNotFailed = base.NewError(2208, "只有失败的任务可以重试")

	ErrUploadSessionNotFound = base.NewError(2301, "上传会话不存在或已过期")
	ErrChunksIncomplete      = base.NewError(2302, "分片不完整")
	ErrChunkUploadFailed     = base.NewError(2303, "分片上传失败")

	ErrObjectNotFound  = base.NewError(2401, "文件不存在")
	ErrUnsupportedType = base.NewError(2402, "不支持的存储类型")
)
