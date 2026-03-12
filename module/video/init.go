// module/video/init.go
package video

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"vidora-api/contract"
	"vidora-api/shared/mod"
)

// 确保实现接口
var _ mod.Module = (*Module)(nil)

// Module 视频模块
type Module struct {
	Service *Service
	handler *Handler
}

// Init 初始化视频模块
func Init(db *gorm.DB, tagSvc contract.TagBiz) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, tagSvc)
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
