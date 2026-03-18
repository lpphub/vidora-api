// module/auth/init.go
package auth

import (
	"vidora-api/contract"
	"vidora-api/server/core"

	"github.com/gin-gonic/gin"
)

// 确保实现接口
var _ core.Module = (*Module)(nil)

// Module 认证模块
type Module struct {
	Service *Service
	handler *Handler
}

// New 创建认证模块
func New(userSvc contract.UserBiz) *Module {
	svc := NewService(userSvc)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.register(r)
}
