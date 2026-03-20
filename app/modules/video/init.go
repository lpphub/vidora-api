package video

import (
	"vidora-api/app/infra/storage"
	"vidora-api/app/modules/core"
	"vidora-api/app/modules/video/handler"
	"vidora-api/app/modules/video/repository"
	"vidora-api/app/modules/video/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	UploadService *service.UploadService
	uploadH       *handler.UploadHandler
}

func New(db *gorm.DB, rdb *redis.Client, _ interface{}, st storage.Client) *Module {
	uploadRepo := repository.NewUploadRepository(db, rdb, st)
	uploadSvc := service.NewUploadService(uploadRepo, st)
	uploadH := handler.NewUploadHandler(uploadSvc)

	return &Module{
		UploadService: uploadSvc,
		uploadH:       uploadH,
	}
}

func (m *Module) Routes(r *gin.RouterGroup) {
	m.uploadH.Register(r)
}
