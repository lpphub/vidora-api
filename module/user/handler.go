// module/user/handler.go
package user

import "github.com/gin-gonic/gin"

// Handler 用户处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建用户处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	// User 模块不直接暴露 HTTP 接口
	// 用户相关操作通过 Auth 模块进行
}