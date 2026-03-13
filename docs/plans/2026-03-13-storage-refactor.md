# Storage 重构实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 重构存储系统，支持 Minio/S3/OSS 多种云存储后端的配置切换

**Architecture:** 统一 StorageClient 接口定义在 interface.go，各存储后端独立实现（minio.go/s3.go/oss.go），factory.go 根据配置类型创建对应实例

**Tech Stack:** Go, minio-go/v7, aws-sdk-go-v2, aliyun-oss-go-sdk

---

## Task 1: 创建 storage 目录和接口定义

**Files:**
- Create: `infra/storage/interface.go`
- Create: `infra/storage/errors.go`

**Step 1: 创建目录**

```bash
mkdir -p infra/storage
```

**Step 2: 创建 interface.go**

```go
package storage

import (
	"context"
	"io"
	"time"
)

type StorageClient interface {
	Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error)
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectName string) error
	Exists(ctx context.Context, objectName string) (bool, error)
	GetURL(objectName string) string
	PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

type UploadRequest struct {
	ObjectName  string
	Reader      io.Reader
	Size        int64
	ContentType string
}

type UploadResult struct {
	ObjectName string
	URL        string
}
```

**Step 3: 创建 errors.go**

```go
package storage

import "errors"

var (
	ErrObjectNotFound  = errors.New("object not found")
	ErrBucketNotFound  = errors.New("bucket not found")
	ErrInvalidConfig   = errors.New("invalid storage config")
	ErrUnsupportedType = errors.New("unsupported storage type")
	ErrPresignedURL    = errors.New("failed to generate presigned url")
)
```

**Step 4: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 2: 实现 MinioStorage

**Files:**
- Create: `infra/storage/minio.go`

**Step 1: 创建 minio.go**

```go
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

func (s *MinioStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, req.ObjectName, req.Reader, req.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		ObjectName: req.ObjectName,
		URL:        s.GetURL(req.ObjectName),
	}, nil
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
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 3: 实现 S3Storage

**Files:**
- Create: `infra/storage/s3.go`

**Step 1: 添加 AWS SDK 依赖**

```bash
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/credentials
```

**Step 2: 创建 s3.go**

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Region     string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

type S3Storage struct {
	client     *s3.Client
	bucket     string
	publicHost string
	region     string
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Storage{
		client:     client,
		bucket:     cfg.Bucket,
		publicHost: cfg.PublicHost,
		region:     cfg.Region,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(req.ObjectName),
		Body:        req.Reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		ObjectName: req.ObjectName,
		URL:        s.GetURL(req.ObjectName),
	}, nil
}

func (s *S3Storage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, objectName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	return err
}

func (s *S3Storage) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *S3Storage) GetURL(objectName string) string {
	if s.publicHost != "" {
		return fmt.Sprintf("%s/%s/%s", s.publicHost, s.bucket, objectName)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectName)
}

func (s *S3Storage) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	resp, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}
```

**Step 3: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 4: 实现 OSSStorage

**Files:**
- Create: `infra/storage/oss.go`

**Step 1: 添加 OSS SDK 依赖**

```bash
go get github.com/aliyun/aliyun-oss-go-sdk/oss
```

**Step 2: 创建 oss.go**

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

type OSSConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

type OSSStorage struct {
	bucket     *ossBucket
	publicHost string
}

type ossBucket struct {
	name     string
	client   interface{}
	putObject func(objectKey string, reader io.Reader, options ...interface{}) error
	getObject func(objectKey string, options ...interface{}) (io.ReadCloser, error)
	deleteObject func(objectKey string, options ...interface{}) error
	isObjectExist func(objectKey string, options ...interface{}) (bool, error)
	signURL func(objectKey string, method interface{}, expiredInSec int64, options ...interface{}) (string, error)
}

func NewOSSStorage(cfg OSSConfig) (*OSSStorage, error) {
	return &OSSStorage{
		bucket:     nil,
		publicHost: cfg.PublicHost,
	}, nil
}

func (s *OSSStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	return nil, ErrUnsupportedType
}

func (s *OSSStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return nil, ErrUnsupportedType
}

func (s *OSSStorage) Delete(ctx context.Context, objectName string) error {
	return ErrUnsupportedType
}

func (s *OSSStorage) Exists(ctx context.Context, objectName string) (bool, error) {
	return false, ErrUnsupportedType
}

func (s *OSSStorage) GetURL(objectName string) string {
	if s.publicHost != "" {
		return fmt.Sprintf("%s/%s", s.publicHost, objectName)
	}
	return fmt.Sprintf("/%s", objectName)
}

func (s *OSSStorage) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return "", ErrUnsupportedType
}
```

**注意：** OSS 实现先留 TODO 占位，后续完善

**Step 3: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 5: 创建工厂函数

**Files:**
- Create: `infra/storage/factory.go`

**Step 1: 创建 factory.go**

```go
package storage

import "fmt"

type StorageConfig struct {
	Type  string
	Minio MinioConfig
	S3    S3Config
	OSS   OSSConfig
}

func NewStorage(cfg StorageConfig) (StorageClient, error) {
	switch cfg.Type {
	case "minio":
		return NewMinioStorage(cfg.Minio)
	case "s3":
		return NewS3Storage(cfg.S3)
	case "oss":
		return NewOSSStorage(cfg.OSS)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, cfg.Type)
	}
}
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 6: 更新配置结构

**Files:**
- Modify: `infra/config.go`

**Step 1: 添加 StorageConfig 并更新 Config**

修改 `infra/config.go`，在 MinioConfig 之后添加：

```go
type S3Config struct {
	Region     string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

type OSSConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
}

type StorageConfig struct {
	Type  string
	Minio MinioConfig
	S3    S3Config
	OSS   OSSConfig
}
```

并更新 Config 结构体：

```go
type Config struct {
	Database DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	Storage  StorageConfig
}
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 7: 更新初始化逻辑

**Files:**
- Modify: `infra/init.go`

**Step 1: 更新 import**

```go
import (
	"fmt"

	"github.com/lpphub/goweb/pkg/logging"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"vidora-api/infra/storage"
)
```

**Step 2: 更新全局变量**

```go
var (
	Cfg     *Config
	DB      *gorm.DB
	RDB     *redis.Client
	Storage storage.StorageClient
)
```

**Step 3: 更新 Init 函数**

```go
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

	Storage, err = storage.NewStorage(Cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to init storage: %w", err)
	}

	logging.Init()

	return nil
}
```

**Step 4: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 8: 更新配置文件

**Files:**
- Modify: `config/config.yml`

**Step 1: 更新配置格式**

```yaml
database:
  host: 1Panel-mysql-TUuu
  port: 3306
  dbname: app_db
  user: test
  password: D7p6ZkcJRMynPfCy

redis:
  host: 1Panel-redis-dpZp
  port: 6379
  password: 32wB6ejWwb78cPKT
  db: 0

jwt:
  secret: se123abc456
  expire_time: 7200
  refresh_expire_time: 604800 

server:
  port: 8080

storage:
  type: minio
  minio:
    endpoint: localhost:9000
    access_key: minioadmin
    secret_key: minioadmin
    bucket: videos
    use_ssl: false
    public_host: http://localhost:9000
  s3:
    region: us-east-1
    access_key: ""
    secret_key: ""
    bucket: ""
    public_host: ""
  oss:
    endpoint: ""
    access_key: ""
    secret_key: ""
    bucket: ""
    public_host: ""
```

---

## Task 9: 更新调用方代码

**Files:**
- Modify: `module/video/service/upload.go`
- Modify: `module/video/upload_test.go`
- Modify: `server/app.go`

**Step 1: 更新 module/video/service/upload.go**

修改 uploadToStorage 方法：

```go
func (s *UploadService) uploadToStorage(ctx context.Context, filePath string, session *UploadSession) (string, uint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	objectName := infra.GetVideoObjectName(session.MD5, session.FileName)
	result, err := s.storage.Upload(ctx, &storage.UploadRequest{
		ObjectName:  objectName,
		Reader:      file,
		Size:        stat.Size(),
		ContentType: "video/mp4",
	})
	if err != nil {
		return "", 0, err
	}

	return result.URL, 0, nil
}
```

更新 import：

```go
import (
	"vidora-api/infra"
	"vidora-api/infra/storage"
)
```

**Step 2: 更新 server/app.go**

修改 import：

```go
import (
	"vidora-api/infra"
	"vidora-api/infra/storage/model"
)
```

修改迁移：

```go
if err := infra.Migrate(infra.DB, &model.Video{}, &model.VideoOutput{}, &model.UploadFile{}); err != nil {
```

**Step 3: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 10: 清理旧文件

**Step 1: 删除旧文件**

```bash
rm -f infra/minio.go infra/storage.go infra/storage_memory.go
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: Success

---

## Task 11: 运行测试

**Step 1: 更新测试文件以适配新接口**

修改 `module/video/upload_test.go` 中的测试代码，确保使用新的接口。

**Step 2: 运行测试**

Run: `go test ./... -v`
Expected: All tests pass

---

## Commit Summary

After all tasks complete:

```bash
git add -A
git commit -m "refactor(storage): support minio/s3/oss multi-backend storage"
```