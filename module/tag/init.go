// module/tag/init.go
package tag

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"vidora-api/server/core"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

// Module 标签模块
type Module struct {
	Service *Service
	Repo    *Repository
	handler *Handler
}

// New 创建标签模块
func New(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Repo:    repo,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.register(r)
}
