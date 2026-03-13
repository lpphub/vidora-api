package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	PublicHost string
}

type MinioStorage struct {
	client     *minio.Client
	bucket     string
	publicHost string
}

func NewMinioStorage(cfg MinioConfig) (*MinioStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		client:     client,
		bucket:     cfg.Bucket,
		publicHost: cfg.PublicHost,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return s.GetURL(objectName), nil
}

func (s *MinioStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *MinioStorage) Delete(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinioStorage) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MinioStorage) GetURL(objectName string) string {
	if s.publicHost != "" {
		return fmt.Sprintf("%s/%s/%s", s.publicHost, s.bucket, objectName)
	}
	return fmt.Sprintf("/%s/%s", s.bucket, objectName)
}

func (s *MinioStorage) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *MinioStorage) ComposeObject(ctx context.Context, sources []string, dest string) error {
	var srcs []minio.CopySrcOptions
	for _, src := range sources {
		srcs = append(srcs, minio.CopySrcOptions{
			Bucket: s.bucket,
			Object: src,
		})
	}

	_, err := s.client.ComposeObject(ctx, minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: dest,
	}, srcs...)
	return err
}
