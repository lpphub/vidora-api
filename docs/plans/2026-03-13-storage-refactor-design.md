# Storage 重构设计文档

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 重构存储系统，支持 Minio/S3/OSS 多种云存储后端的配置切换

**Architecture:** 统一 StorageClient 接口 + 各存储后端独立实现，通过配置文件 `storage.type` 选择后端

**Tech Stack:** Go, minio-go, aws-sdk-go-v2, aliyun-oss-go-sdk

---

## 1. 配置结构

```yaml
# config/config.yml
storage:
  type: minio          # minio / s3 / oss
  
  minio:
    endpoint: localhost:9000
    access_key: minioadmin
    secret_key: minioadmin
    bucket: videos
    use_ssl: false
    public_host: http://localhost:9000
  
  s3:
    region: us-east-1
    access_key: xxx
    secret_key: xxx
    bucket: my-bucket
    public_host: https://s3.amazonaws.com
  
  oss:
    endpoint: oss-cn-hangzhou.aliyuncs.com
    access_key: xxx
    secret_key: xxx
    bucket: my-bucket
    public_host: https://my-bucket.oss-cn-hangzhou.aliyuncs.com
```

---

## 2. 接口设计

```go
// infra/storage/interface.go
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

---

## 3. 文件结构

```
infra/storage/
├── interface.go      # StorageClient 接口
├── factory.go        # NewStorage 工厂函数
├── minio.go          # MinioStorage 实现
├── s3.go             # S3Storage 实现
├── oss.go            # OSSStorage 实现
└── errors.go         # 统一错误定义
```

---

## 4. 初始化集成

**infra/config.go 修改：**
```go
type StorageConfig struct {
    Type  string
    Minio MinioConfig
    S3    S3Config
    OSS   OSSConfig
}
```

**infra/init.go 修改：**
```go
var Storage storage.StorageClient

func Init() error {
    // ...
    Storage, err = storage.NewStorage(Cfg.Storage)
    // ...
}
```

---

## 5. 错误定义

```go
// infra/storage/errors.go
var (
    ErrObjectNotFound    = errors.New("object not found")
    ErrBucketNotFound    = errors.New("bucket not found")
    ErrInvalidConfig     = errors.New("invalid storage config")
    ErrUnsupportedType   = errors.New("unsupported storage type")
    ErrPresignedURL      = errors.New("failed to generate presigned url")
)
```

---

## 6. 迁移影响

| 文件 | 修改内容 |
|------|---------|
| module/video/service/upload.go | Upload 参数改为 UploadRequest |
| module/video/upload_test.go | 适配新接口 |
| server/app.go | 更新引用 |

**删除文件：**
- `infra/minio.go`
- `infra/storage.go`
- `infra/storage_memory.go`

---

## 7. 依赖

```go
require (
    github.com/minio/minio-go/v7
    github.com/aws/aws-sdk-go-v2/service/s3
    github.com/aws/aws-sdk-go-v2/config
    github.com/aliyun/aliyun-oss-go-sdk/oss
)
```