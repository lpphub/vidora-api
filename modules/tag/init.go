package tag

import (
	"vidora-api/modules/core"
	"vidora-api/modules/tag/handler"
	"vidora-api/modules/tag/repository"
	"vidora-api/modules/tag/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	Service *service.Service
	handler *handler.Handler
}

func New(db *gorm.DB) *Module {
	repo := repository.NewRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	svc := service.NewService(repo)
	groupSvc := service.NewGroupService(groupRepo, repo)
	h := handler.NewHandler(svc, groupSvc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Register(r)
}
