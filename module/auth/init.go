// module/auth/init.go
package auth

import (
	"github.com/gin-gonic/gin"
	"vidora-api/contract"
	"vidora-api/shared/mod"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

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

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
