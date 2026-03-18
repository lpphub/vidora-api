package tag

import (
	"vidora-api/modules/core"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	Service *Service
	handler *Handler
}

func New(db *gorm.DB) *Module {
	repo := NewRepository(db)
	groupRepo := NewGroupRepository(db)
	svc := NewService(repo)
	groupSvc := NewGroupService(groupRepo, repo)
	h := NewHandler(svc, groupSvc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.register(r)
}
