package infra

import (
	"fmt"
	"vidora-api/app/infra/storage"

	"github.com/lpphub/goweb/pkg/logging"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
