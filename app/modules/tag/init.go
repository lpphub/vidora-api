package tag

import (
	"vidora-api/app/modules/core"
	"vidora-api/app/modules/tag/handler"
	repository2 "vidora-api/app/modules/tag/repository"
	service2 "vidora-api/app/modules/tag/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	Service *service2.Service
	handler *handler.Handler
}

func New(db *gorm.DB) *Module {
	repo := repository2.NewRepository(db)
	groupRepo := repository2.NewGroupRepository(db)
	svc := service2.NewService(repo)
	groupSvc := service2.NewGroupService(groupRepo, repo)
	h := handler.NewHandler(svc, groupSvc)

	return &Module{
		Service: svc,
		handler: h,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.handler.Register(r)
}
