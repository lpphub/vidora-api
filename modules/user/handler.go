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

func (h *Handler) register(r *gin.RouterGroup) {
}
