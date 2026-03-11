package helper

import (
	"strconv"
	"vidora-api/server/middleware"
	"vidora-api/shared/errs"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/base"
	"github.com/lpphub/goweb/pkg/logging"
)

// ---------------- JSON / Query ----------------

// MustBindJSON 绑定 JSON 并返回 bool，失败返回 false
func MustBindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		logging.Errorw(c.Request.Context(), err)
		base.Fail(c, errs.ErrInvalidParam)
		return false
	}
	return true
}

// MustBindQuery 绑定 Query 并返回 bool，失败返回 false
func MustBindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		logging.Errorw(c.Request.Context(), err)
		base.Fail(c, errs.ErrInvalidParam)
		return false
	}
	return true
}

// ---------------- UserID ----------------

// MustGetUserID 获取用户 ID，失败返回 0 + false
func MustGetUserID(c *gin.Context) (uint, bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		base.Fail(c, errs.ErrNoToken)
		return 0, false
	}
	return userID, true
}

// ---------------- Param ----------------

// MustParseUintParam 解析 uint 参数，失败返回 0 + false
func MustParseUintParam(c *gin.Context, param string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(param), 10, 32)
	if err != nil {
		logging.Errorw(c, err)
		base.Fail(c, errs.ErrInvalidParam)
		return 0, false
	}
	return uint(id), true
}

// ---------------- Service Result 处理 ----------------

func Respond(c *gin.Context, err error, data ...any) {
	base.Respond(c, err, data...)
}
