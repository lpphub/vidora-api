// module/user/init.go
package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"vidora-api/shared/mod"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

// Module 用户模块
type Module struct {
	Service *Service
	handler *Handler
}

// New 创建用户模块
func New(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
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
