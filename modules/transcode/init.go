// module/transcode/init.go
package transcode

import (
	"vidora-api/server/core"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 确保实现接口
var _ core.Module = (*Module)(nil)

// Module 转码模块
type Module struct {
	Service *Service
	handler *Handler
}

// New 创建转码模块
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
