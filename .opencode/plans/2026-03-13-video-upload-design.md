# Video Upload 模块重构设计

## 概述

重构 `module/video/` 模块，专注于文件上传功能，支持秒传、分片上传、续传。

## 目标

1. 支持视频文件上传管理（秒传、分片上传、续传）
2. 简化代码结构，职责清晰
3. 目录结构优化：service/repository/handler 拆分为独立目录

## 不在本次范围

- 内容管理（电影/电视剧/短视频的 CRUD）
- 转码任务

---

## 目录结构

```
module/video/
├── model.go              # UploadFile, UploadSession
├── dto.go                # 上传相关 DTO
├── repository/
│   ├── video.go          # Content, Season, Episode（预留，暂不实现）
│   └── upload.go         # UploadFile 数据访问 + Redis Session 管理
├── service/
│   ├── video.go          # 内容管理（预留，暂不实现）
│   └── upload.go         # 上传核心逻辑
├── handler/
│   ├── video.go          # 内容接口（预留，暂不实现）
│   └── upload.go         # 上传接口
└── init.go
```

---

## Model 设计

### UploadFile

已上传完成的文件，用于秒传检测。

```go
type UploadFile struct {
    ID         uint      `gorm:"primaryKey"`
    FileKey    string    `gorm:"uniqueIndex;size:64;not null"`  // 唯一标识，前端用于关联内容
    MD5        string    `gorm:"uniqueIndex;size:32;not null"`  // 秒传检测
    FileName   string    `gorm:"size:255;not null"`
    FileSize   int64     `gorm:"not null"`
    StorageURL string    `gorm:"size:512"`                      // 存储后的 URL
    CreatedAt  time.Time
}

func (*UploadFile) TableName() string {
    return "upload_files"
}
```

### UploadSession

上传会话，存储在 Redis，用于续传。

```go
type UploadSession struct {
    UploadID       string `json:"uploadId"`
    MD5            string `json:"md5"`
    FileName       string `json:"fileName"`
    FileSize       int64  `json:"fileSize"`
    TotalChunks    int    `json:"totalChunks"`
    UploadedChunks []int  `json:"uploadedChunks"`  // 0-indexed
    CreatedAt      int64  `json:"createdAt"`
}
```

**Redis Key**: `upload:{uploadId}`
**TTL**: 24 小时

---

## DTO 设计

```go
// === Init ===
type UploadInitReq struct {
    MD5         string `json:"md5" binding:"required,len=32"`
    FileName    string `json:"fileName" binding:"required,max=255"`
    FileSize    int64  `json:"fileSize" binding:"required,min=1"`
    TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

type UploadInitResp struct {
    Exists         bool   `json:"exists"`                    // 秒传标记
    FileKey        string `json:"fileKey,omitempty"`         // 秒传时返回
    URL            string `json:"url,omitempty"`             // 秒传时返回
    UploadID       string `json:"uploadId"`                  // 新上传/续传时返回
    UploadedChunks []int  `json:"uploadedChunks"`            // 续传时返回已上传分片
}

// === Chunk ===
type UploadChunkReq struct {
    File        *multipart.FileHeader `form:"file" binding:"required"`
    UploadID    string                `form:"uploadId" binding:"required"`
    ChunkNumber int                   `form:"chunkNumber" binding:"required,min=1"`  // 1-indexed
    TotalChunks int                   `form:"totalChunks" binding:"required"`
    MD5         string                `form:"md5" binding:"required,len=32"`
}

type UploadChunkResp struct {
    Uploaded    bool `json:"uploaded"`
    ChunkNumber int  `json:"chunkNumber"`
}

// === Merge ===
type UploadMergeReq struct {
    UploadID    string `json:"uploadId" binding:"required"`
    MD5         string `json:"md5" binding:"required,len=32"`
    FileName    string `json:"fileName" binding:"required"`
    TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

type UploadMergeResp struct {
    FileKey string `json:"fileKey"`
    URL     string `json:"url"`
}
```

---

## API 设计

### POST /upload/init

初始化上传会话，同时完成秒传检测和续传查询。

**流程**：
1. 检查 MD5 是否已存在 → 秒传
2. 检查 MD5 是否有未完成的上传会话 → 续传
3. 创建新的上传会话 → 新上传

**响应场景**：

| 场景 | exists | fileKey/url | uploadId | uploadedChunks |
|------|--------|-------------|----------|----------------|
| 秒传 | true | 有值 | 空 | 空 |
| 续传 | false | 无 | 已有ID | 已上传分片 |
| 新上传 | false | 无 | 新ID | 空 |

### POST /upload/chunk

上传单个分片。

**流程**：
1. 验证 uploadId 和 session 存在
2. 验证 MD5 匹配
3. 保存分片到临时目录
4. 更新 session 的 uploadedChunks
5. 返回上传结果

### POST /upload/merge

合并所有分片。

**流程**：
1. 验证 uploadId 和 session 存在
2. 验证所有分片已上传
3. 合并分片到临时文件
4. 上传到存储服务
5. 创建 UploadFile 记录
6. 清理临时文件和 session
7. 返回 fileKey 和 URL

---

## Repository 层

### repository/upload.go

```go
type UploadRepository struct {
    db     *gorm.DB
    rdb    *redis.Client
    tmpDir string
}

// === UploadFile 数据库操作 ===
func (r *UploadRepository) FindUploadFileByMD5(ctx, md5) (*UploadFile, error)
func (r *UploadRepository) CreateUploadFile(ctx, file *UploadFile) error

// === UploadSession Redis 操作 ===
func (r *UploadRepository) CreateSession(ctx, uploadID string, session *UploadSession) error
func (r *UploadRepository) GetSession(ctx, uploadID string) (*UploadSession, error)
func (r *UploadRepository) UpdateSessionChunks(ctx, uploadID string, chunks []int) error
func (r *UploadRepository) DeleteSession(ctx, uploadID string) error

// === 分片文件操作 ===
func (r *UploadRepository) SaveChunk(uploadID string, chunkNumber int, file *multipart.FileHeader) error
func (r *UploadRepository) MergeChunks(uploadID string, totalChunks int, destPath string) error
func (r *UploadRepository) CleanupChunks(uploadID string, totalChunks int) error
```

---

## Service 层

### service/upload.go

```go
type UploadService struct {
    repo    *repository.UploadRepository
    storage storage.StorageClient
    mu      sync.Mutex  // 合并时加锁
}

func (s *UploadService) Init(ctx, req *UploadInitReq) (*UploadInitResp, error)
func (s *UploadService) UploadChunk(ctx, req *UploadChunkReq) (*UploadChunkResp, error)
func (s *UploadService) Merge(ctx, req *UploadMergeReq) (*UploadMergeResp, error)
```

---

## Handler 层

### handler/upload.go

```go
type UploadHandler struct {
    svc *service.UploadService
}

func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup)
func (h *UploadHandler) Init(c *gin.Context)
func (h *UploadHandler) Chunk(c *gin.Context)
func (h *UploadHandler) Merge(c *gin.Context)
```

---

## 工具函数

```go
// 生成 uploadId
func GenerateUploadID(md5 string) string {
    return fmt.Sprintf("%s_%d", md5, time.Now().UnixNano())
}

// 生成 fileKey
func GenerateFileKey(md5 string) string {
    return fmt.Sprintf("v_%s_%s", md5[:8], time.Now().Format("20060102"))
}
```

---

## 删除文件

重构后删除以下文件：
- `service.go`
- `upload_service.go`
- `handler.go`
- `upload_handler.go`
- `repository.go`

---

## 迁移注意

1. 删除 `VideoOutput` 表（未使用）
2. `Video` 表暂时保留，后续内容模块重构时处理
3. `UploadFile` 表结构变化：添加 `file_key` 字段，删除 `video_id` 字段