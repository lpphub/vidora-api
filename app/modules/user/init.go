// modules/user/init.go
package user

import (
	"vidora-api/app/modules/core"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 确保实现接口
var _ core.Module = (*Module)(nil)

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

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.register(r)
}
