package video

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vidora-api/infra/storage"
	"vidora-api/module/video/handler"
	"vidora-api/module/video/repository"
	"vidora-api/module/video/service"
	"vidora-api/server/core"
)

var _ mod.Module = (*Module)(nil)

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
