// module/transcode/init.go
package transcode

import (
	"github.com/gin-gonic/gin"
	"vidora-api/shared/mod"
	"gorm.io/gorm"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

// Module 转码模块
type Module struct {
	Service *Service
	handler *Handler
}

// Init 初始化转码模块
func Init(db *gorm.DB) *Module {
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