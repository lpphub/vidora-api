package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"vidora-api/shared/errs"
)

type Client interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectName string) error
	Exists(ctx context.Context, objectName string) (bool, error)
	GetURL(objectName string) string
	PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	ComposeObject(ctx context.Context, sources []string, dest string) error
}

type Config struct {
	Type  string
	Minio MinioConfig
	OSS   OSSConfig
}

func NewStorage(cfg Config) (Client, error) {
	switch cfg.Type {
	case "minio":
		return NewMinioStorage(cfg.Minio)
	case "oss":
		return NewOSSStorage(cfg.OSS)
	default:
		return nil, fmt.Errorf("%w: %s", errs.ErrUnsupportedType, cfg.Type)
	}
}
