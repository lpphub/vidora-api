# 上传模块重构设计

## 概述

重构视频上传模块，优化代码结构和接口设计，支持秒传、断点续传、并发控制。

## 模块结构

```
module/video/
├── dto.go              # 请求/响应结构
├── model/
│   ├── model.go        # 数据库模型 (Video, UploadFile)
│   └── session.go      # 会话模型 (UploadSession)
├── handler/
│   └── upload.go       # HTTP 处理器
├── repository/
│   ├── session.go      # 会话管理 (Redis)
│   └── file.go         # 文件存储 (MinIO)
├── service/
│   ├── upload.go       # 上传核心逻辑
│   └── limiter.go      # 并发限制器
└── init.go             # 模块初始化
```

## API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /upload/init | 初始化上传，返回 uploadId + 秒传URL |
| POST | /upload/chunk | 上传分片，支持断点续传 |
| POST | /upload/merge | 合并分片 |
| GET | /upload/status/:uploadId | 获取上传进度 |

## 请求/响应结构

### POST /upload/init

请求:
```json
{
  "md5": "abc123...",
  "fileName": "video.mp4",
  "fileSize": 104857600,
  "totalChunks": 100
}
```

响应:
```json
{
  "uploadId": "abc123_1710300000000",
  "exists": false,
  "url": "",
  "uploadedChunks": []
}
```

秒传时:
```json
{
  "uploadId": "",
  "exists": true,
  "url": "https://storage.example.com/videos/...",
  "uploadedChunks": []
}
```

### POST /upload/chunk

请求: multipart/form-data
- file: 分片文件
- uploadId: 上传ID
- chunkNumber: 分片序号 (1-based)
- md5: 文件整体MD5

响应:
```json
{
  "uploaded": true,
  "chunkNumber": 1
}
```

### POST /upload/merge

请求:
```json
{
  "uploadId": "abc123_1710300000000",
  "md5": "abc123...",
  "fileName": "video.mp4",
  "totalChunks": 100
}
```

响应:
```json
{
  "fileKey": "v_abc123_20260313",
  "url": "https://storage.example.com/videos/..."
}
```

### GET /upload/status/:uploadId

响应:
```json
{
  "uploadId": "abc123_1710300000000",
  "fileName": "video.mp4",
  "fileSize": 104857600,
  "totalChunks": 100,
  "uploadedChunks": 50,
  "progress": 50,
  "eta": 120,
  "status": "uploading"
}
```

status 可能值: pending, uploading, completed, failed

## 核心流程

### Init 流程

1. MD5 查询 upload_files 表
   - 存在: 返回秒传结果
2. MD5 查询 Redis 会话
   - 存在: 返回续传结果
3. 创建新会话存入 Redis
4. 返回新 uploadId

### Chunk 流程

1. 验证会话存在且 MD5 匹配
2. 上传分片到 MinIO: `chunks/{uploadId}/{chunkNumber}`
3. 更新 Redis 会话的 uploadedChunks
4. 返回上传结果

### Merge 流程

1. 并发限制器获取令牌
2. 验证会话完整 (所有分片已上传)
3. 合并分片到 MinIO: `videos/{fileKey}/{fileName}`
4. 创建 upload_files 记录
5. 清理分片 + 删除会话
6. 返回最终 URL

## 数据模型

### UploadFile (数据库)

```go
type UploadFile struct {
    ID         uint      `gorm:"primaryKey"`
    FileKey    string    `gorm:"uniqueIndex;size:64;not null"`
    MD5        string    `gorm:"uniqueIndex;size:32;not null"`
    FileName   string    `gorm:"size:255;not null"`
    FileSize   int64     `gorm:"not null"`
    StorageURL string    `gorm:"size:512"`
    CreatedAt  time.Time
}
```

### UploadSession (Redis)

```go
type UploadSession struct {
    UploadID       string `json:"uploadId"`
    MD5            string `json:"md5"`
    FileName       string `json:"fileName"`
    FileSize       int64  `json:"fileSize"`
    TotalChunks    int    `json:"totalChunks"`
    UploadedChunks []int  `json:"uploadedChunks"`
    CreatedAt      int64  `json:"createdAt"`
    UpdatedAt      int64  `json:"updatedAt"`
}
```

Redis Key: `upload:{uploadId}`
TTL: 24 小时

## 并发控制

- 合并操作使用 semaphore 限制并发数
- 默认并发数: 5
- 可通过配置调整

## 存储路径

| 类型 | 路径格式 |
|------|---------|
| 分片 | `chunks/{uploadId}/{chunkNumber}` |
| 最终文件 | `videos/{fileKey}/{fileName}` |

## 错误码

| 错误码 | 说明 |
|--------|------|
| 2301 | 上传会话不存在或已过期 |
| 2302 | 分片不完整 |
| 2303 | 分片上传失败 |
| 2304 | 合并失败 |
| 2305 | 并发限制 |

## 改进点总结

| 改进点 | 重构前 | 重构后 |
|--------|--------|--------|
| 分片存储 | 本地临时目录 | 直接 MinIO |
| 进度查询 | 无 | GET /status 接口 |
| 响应详细度 | 简单 | 进度 + ETA |
| 并发控制 | 全局锁 | Semaphore |
| 清理策略 | TTL + 手动 | 自动立即清理 |