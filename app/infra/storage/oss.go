package storage

import (
	"context"
	"fmt"
	"io"
	"time"
	"vidora-api/app/shared/errs"
)

type OSSConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

type OSSStorage struct {
	publicHost string
}

func NewOSSStorage(cfg OSSConfig) (*OSSStorage, error) {
	return &OSSStorage{
		publicHost: cfg.PublicHost,
	}, nil
}

func (s *OSSStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	return "", errs.ErrUnsupportedType
}

func (s *OSSStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return nil, errs.ErrUnsupportedType
}

func (s *OSSStorage) Delete(ctx context.Context, objectName string) error {
	return errs.ErrUnsupportedType
}

func (s *OSSStorage) Exists(ctx context.Context, objectName string) (bool, error) {
	return false, errs.ErrUnsupportedType
}

func (s *OSSStorage) GetURL(objectName string) string {
	if s.publicHost != "" {
		return fmt.Sprintf("%s/%s", s.publicHost, objectName)
	}
	return fmt.Sprintf("/%s", objectName)
}

func (s *OSSStorage) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return "", errs.ErrUnsupportedType
}

func (s *OSSStorage) ComposeObject(ctx context.Context, sources []string, dest string) error {
	return errs.ErrUnsupportedType
}
