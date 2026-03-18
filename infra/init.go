package infra

import (
	"fmt"
	"log"

	"github.com/lpphub/goweb/pkg/logging"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vidora-api/infra/storage"
	tagModel "vidora-api/modules/tag/model"
	"vidora-api/modules/user"
	vModel "vidora-api/modules/video/model"
)

var (
	Cfg     *Config
	DB      *gorm.DB
	RDB     *redis.Client
	Storage storage.Client
)

func Init() error {
	var err error

	Cfg, err = LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	DB, err = NewMysqlDB(Cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	RDB, err = NewRedis(Cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	//Storage, err = storage.NewStorage(Cfg.Storage)
	//if err != nil {
	//	return fmt.Errorf("failed to init storage: %w", err)
	//}

	logging.Init()

	return nil
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&user.User{},
		&tagModel.Tag{},
		&tagModel.TagGroup{},
		&tagModel.VideoTag{},
		&vModel.Video{},
		&vModel.UploadFile{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to auto migrate: %v", err))
	}
	log.Println("Database tables migrated successfully")
}
