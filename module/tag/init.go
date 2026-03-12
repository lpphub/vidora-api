// module/tag/init.go
package tag

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"vidora-api/shared/mod"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

// Module 标签模块
type Module struct {
	Service *Service
	Repo    *Repository
	handler *Handler
}

// Init 初始化标签模块
func Init(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Repo:    repo,
		handler: h,
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.Routes(r)
}
