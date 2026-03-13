# Video Upload Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement chunked video upload with instant upload (秒传) support

**Architecture:** 
- Upload session stored in Redis with 24h TTL for resumable uploads
- UploadFile table tracks MD5→URL mapping for instant upload detection
- S3 storage abstraction with TODO placeholder for actual implementation

**Tech Stack:** Go, Gin, GORM, Redis, multipart/form-data

---

## Task 1: Add UploadFile Model

**Files:**
- Modify: `module/video/model.go`

**Step 1: Add UploadFile model to model.go**

Add after VideoOutput struct:

```go
// UploadFile 上传文件记录（用于秒传）
type UploadFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MD5       string    `gorm:"uniqueIndex;size:32;not null" json:"md5"`
	FileName  string    `gorm:"size:255;not null" json:"fileName"`
	FileSize  int64     `gorm:"not null" json:"fileSize"`
	VideoURL  string    `gorm:"size:512" json:"videoUrl"`
	VideoID   uint      `json:"videoId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (*UploadFile) TableName() string {
	return "upload_files"
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 2: Add Upload DTOs

**Files:**
- Modify: `module/video/dto.go`

**Step 1: Add upload request/response DTOs**

Add at end of file:

```go
// UploadInitReq 初始化上传请求
type UploadInitReq struct {
	MD5         string `json:"md5" binding:"required,len=32"`
	FileName    string `json:"fileName" binding:"required,max=255"`
	FileSize    int64  `json:"fileSize" binding:"required,min=1"`
	TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

// UploadInitResp 初始化上传响应
type UploadInitResp struct {
	Exists          bool     `json:"exists"`
	URL             string   `json:"url,omitempty"`
	VideoID         uint     `json:"videoId,omitempty"`
	UploadID        string   `json:"uploadId"`
	UploadedChunks  []int    `json:"uploadedChunks"`
}

// UploadChunkReq 上传分片请求
type UploadChunkReq struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	UploadID    string                `form:"uploadId" binding:"required"`
	ChunkNumber int                   `form:"chunkNumber" binding:"required,min=1"`
	TotalChunks int                   `form:"totalChunks" binding:"required"`
	MD5         string                `form:"md5" binding:"required,len=32"`
}

// UploadChunkResp 上传分片响应
type UploadChunkResp struct {
	ChunkNumber int  `json:"chunkNumber"`
	Uploaded    bool `json:"uploaded"`
}

// UploadMergeReq 合并分片请求
type UploadMergeReq struct {
	UploadID    string `json:"uploadId" binding:"required"`
	MD5         string `json:"md5" binding:"required,len=32"`
	FileName    string `json:"fileName" binding:"required"`
	TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

// UploadMergeResp 合并分片响应
type UploadMergeResp struct {
	URL     string `json:"url"`
	VideoID uint   `json:"videoId,omitempty"`
}
```

**Step 2: Add missing import at top of file**

Add `"mime/multipart"` to imports.

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 3: Create Upload Store (Redis)

**Files:**
- Create: `module/video/upload_store.go`

**Step 1: Create upload_store.go**

```go
package video

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	uploadKeyPrefix = "upload:"
	uploadTTL       = 24 * time.Hour
)

// UploadSession 上传会话
type UploadSession struct {
	MD5            string `json:"md5"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	TotalChunks    int    `json:"totalChunks"`
	UploadedChunks []int  `json:"uploadedChunks"`
	CreatedAt      int64  `json:"createdAt"`
}

// UploadStore 上传会话存储
type UploadStore struct {
	rdb *redis.Client
}

func NewUploadStore(rdb *redis.Client) *UploadStore {
	return &UploadStore{rdb: rdb}
}

func (s *UploadStore) key(uploadID string) string {
	return uploadKeyPrefix + uploadID
}

func (s *UploadStore) Create(ctx context.Context, uploadID string, session *UploadSession) error {
	session.CreatedAt = time.Now().Unix()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(uploadID), data, uploadTTL).Err()
}

func (s *UploadStore) Get(ctx context.Context, uploadID string) (*UploadSession, error) {
	data, err := s.rdb.Get(ctx, s.key(uploadID)).Bytes()
	if err != nil {
		return nil, err
	}
	var session UploadSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *UploadStore) AddChunk(ctx context.Context, uploadID string, chunkNumber int) error {
	session, err := s.Get(ctx, uploadID)
	if err != nil {
		return err
	}
	session.UploadedChunks = append(session.UploadedChunks, chunkNumber)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(uploadID), data, uploadTTL).Err()
}

func (s *UploadStore) Delete(ctx context.Context, uploadID string) error {
	return s.rdb.Del(ctx, s.key(uploadID)).Err()
}

func (s *UploadStore) Exists(ctx context.Context, uploadID string) bool {
	return s.rdb.Exists(ctx, s.key(uploadID)).Val() > 0
}

func GenerateUploadID(md5 string) string {
	return fmt.Sprintf("%s_%d", md5, time.Now().UnixNano())
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 4: Create Upload Service

**Files:**
- Create: `module/video/upload_service.go`

**Step 1: Create upload_service.go**

```go
package video

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"vidora-api/shared/errs"
)

// UploadService 上传服务
type UploadService struct {
	db       *gorm.DB
	rdb      *redis.Client
	store    *UploadStore
	repo     *Repository
	tmpDir   string
	mu       sync.Mutex
}

func NewUploadService(db *gorm.DB, rdb *redis.Client, repo *Repository) *UploadService {
	return &UploadService{
		db:     db,
		rdb:    rdb,
		store:  NewUploadStore(rdb),
		repo:   repo,
		tmpDir: os.TempDir(),
	}
}

func (s *UploadService) Init(ctx context.Context, req UploadInitReq) (*UploadInitResp, error) {
	var uploadFile UploadFile
	err := s.db.WithContext(ctx).Where("md5 = ?", req.MD5).First(&uploadFile).Error
	if err == nil {
		return &UploadInitResp{
			Exists:   true,
			URL:      uploadFile.VideoURL,
			VideoID:  uploadFile.VideoID,
			UploadID: "",
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	uploadID := GenerateUploadID(req.MD5)
	session := &UploadSession{
		MD5:            req.MD5,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: []int{},
	}
	if err := s.store.Create(ctx, uploadID, session); err != nil {
		return nil, err
	}

	if err := s.ensureTempDir(uploadID); err != nil {
		return nil, err
	}

	return &UploadInitResp{
		Exists:         false,
		UploadID:       uploadID,
		UploadedChunks: []int{},
	}, nil
}

func (s *UploadService) UploadChunk(ctx context.Context, req UploadChunkReq) (*UploadChunkResp, error) {
	session, err := s.store.Get(ctx, req.UploadID)
	if err != nil {
		return nil, errs.ErrUploadSessionNotFound
	}

	if session.MD5 != req.MD5 {
		return nil, errs.ErrInvalidParam
	}

	chunkPath := s.chunkPath(req.UploadID, req.ChunkNumber)
	if err := s.saveChunk(req.File, chunkPath); err != nil {
		return nil, err
	}

	if err := s.store.AddChunk(ctx, req.UploadID, req.ChunkNumber-1); err != nil {
		return nil, err
	}

	return &UploadChunkResp{
		ChunkNumber: req.ChunkNumber,
		Uploaded:    true,
	}, nil
}

func (s *UploadService) Merge(ctx context.Context, req UploadMergeReq) (*UploadMergeResp, error) {
	session, err := s.store.Get(ctx, req.UploadID)
	if err != nil {
		return nil, errs.ErrUploadSessionNotFound
	}

	if session.MD5 != req.MD5 || session.TotalChunks != req.TotalChunks {
		return nil, errs.ErrInvalidParam
	}

	if len(session.UploadedChunks) != req.TotalChunks {
		return nil, errs.ErrChunksIncomplete
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	finalPath := s.finalPath(req.UploadID, session.FileName)
	if err := s.mergeChunks(req.UploadID, req.TotalChunks, finalPath); err != nil {
		return nil, err
	}

	videoURL, videoID, err := s.uploadToStorage(ctx, finalPath, session)
	if err != nil {
		os.Remove(finalPath)
		return nil, err
	}

	uploadFile := &UploadFile{
		MD5:      session.MD5,
		FileName: session.FileName,
		FileSize: session.FileSize,
		VideoURL: videoURL,
		VideoID:  videoID,
	}
	if err := s.db.WithContext(ctx).Create(uploadFile).Error; err != nil {
		return nil, err
	}

	s.cleanup(req.UploadID, req.TotalChunks)
	s.store.Delete(ctx, req.UploadID)

	return &UploadMergeResp{
		URL:     videoURL,
		VideoID: videoID,
	}, nil
}

func (s *UploadService) ensureTempDir(uploadID string) error {
	dir := s.tempDir(uploadID)
	return os.MkdirAll(dir, 0755)
}

func (s *UploadService) tempDir(uploadID string) string {
	return filepath.Join(s.tmpDir, "uploads", uploadID)
}

func (s *UploadService) chunkPath(uploadID string, chunkNumber int) string {
	return filepath.Join(s.tempDir(uploadID), fmt.Sprintf("chunk_%d", chunkNumber))
}

func (s *UploadService) finalPath(uploadID, fileName string) string {
	return filepath.Join(s.tempDir(uploadID), fileName)
}

func (s *UploadService) saveChunk(file *multipart.FileHeader, path string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (s *UploadService) mergeChunks(uploadID string, totalChunks int, finalPath string) error {
	out, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for i := 1; i <= totalChunks; i++ {
		chunkPath := s.chunkPath(uploadID, i)
		chunk, err := os.Open(chunkPath)
		if err != nil {
			return err
		}
		io.Copy(out, chunk)
		chunk.Close()
	}
	return nil
}

func (s *UploadService) uploadToStorage(ctx context.Context, filePath string, session *UploadSession) (string, uint, error) {
	// TODO: Implement S3 upload
	// For now, return a placeholder URL
	videoURL := fmt.Sprintf("https://cdn.example.com/videos/%s/%s", session.MD5, session.FileName)
	return videoURL, 0, nil
}

func (s *UploadService) cleanup(uploadID string, totalChunks int) {
	for i := 1; i <= totalChunks; i++ {
		os.Remove(s.chunkPath(uploadID, i))
	}
	os.Remove(s.finalPath(uploadID, ""))
	os.RemoveAll(s.tempDir(uploadID))
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success (will fail due to missing errors)

---

## Task 5: Add Upload Errors

**Files:**
- Modify: `shared/errs/errors.go`

**Step 1: Add upload-related errors**

Add after ErrTranscodeNotFailed:

```go
	// 上传错误
	ErrUploadSessionNotFound = base.NewError(2301, "上传会话不存在或已过期")
	ErrChunksIncomplete      = base.NewError(2302, "分片不完整")
	ErrChunkUploadFailed     = base.NewError(2303, "分片上传失败")
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 6: Create Upload Handler

**Files:**
- Create: `module/video/upload_handler.go`

**Step 1: Create upload_handler.go**

```go
package video

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	uploadSvc *UploadService
}

func NewUploadHandler(uploadSvc *UploadService) *UploadHandler {
	return &UploadHandler{uploadSvc: uploadSvc}
}

func (h *UploadHandler) Routes(r *gin.RouterGroup) {
	upload := r.Group("/upload")
	{
		upload.POST("/init", h.Init)
		upload.POST("/chunk", h.Chunk)
		upload.POST("/merge", h.Merge)
	}
}

func (h *UploadHandler) Init(c *gin.Context) {
	var req UploadInitReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.uploadSvc.Init(c.Request.Context(), req)
	helper.Respond(c, err, resp)
}

func (h *UploadHandler) Chunk(c *gin.Context) {
	var req UploadChunkReq
	if err := c.ShouldBind(&req); err != nil {
		helper.Respond(c, err, nil)
		return
	}

	resp, err := h.uploadSvc.UploadChunk(c.Request.Context(), req)
	helper.Respond(c, err, resp)
}

func (h *UploadHandler) Merge(c *gin.Context) {
	var req UploadMergeReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.uploadSvc.Merge(c.Request.Context(), req)
	helper.Respond(c, err, resp)
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 7: Update Module Init

**Files:**
- Modify: `module/video/init.go`

**Step 1: Add UploadService and UploadHandler to Module**

Replace entire file content:

```go
package video

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"vidora-api/contract"
	"vidora-api/shared/mod"
)

var _ mod.Module = (*Module)(nil)

type Module struct {
	Service    *Service
	UploadSvc  *UploadService
	handler    *Handler
	uploadH    *UploadHandler
}

func New(db *gorm.DB, rdb *redis.Client, tagSvc contract.TagBiz) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, tagSvc)
	h := NewHandler(svc)

	uploadSvc := NewUploadService(db, rdb, repo)
	uploadH := NewUploadHandler(uploadSvc)

	return &Module{
		Service:   svc,
		UploadSvc: uploadSvc,
		handler:   h,
		uploadH:   uploadH,
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.Routes(r)
	m.uploadH.Routes(r)
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Will fail - need to update server/app.go to pass redis

---

## Task 8: Update Server App

**Files:**
- Modify: `server/app.go`

**Step 1: Update initModules to pass Redis client**

Replace the initModules function:

```go
func (a *App) initModules() []mod.Module {
	userMod := user.New(infra.DB)
	authMod := auth.New(userMod.Service)
	tagMod := tag.New(infra.DB)
	videoMod := video.New(infra.DB, infra.RDB, tagMod.Service)
	transcodeMod := transcode.New(infra.DB)

	return []mod.Module{userMod, authMod, tagMod, videoMod, transcodeMod}
}
```

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: Success

---

## Task 9: Run Database Migration

**Files:**
- Modify: `infra/database.go`

**Step 1: Add UploadFile to auto-migration**

Check current file content first, then add UploadFile to migration if needed.

**Step 2: Run migration**

Run: `go run main.go &` then check logs for migration
Expected: Table `upload_files` created

---

## Task 10: Manual Testing

**Step 1: Test init endpoint**

```bash
curl -X POST http://localhost:8080/upload/init \
  -H "Content-Type: application/json" \
  -d '{"md5":"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6","fileName":"test.mp4","fileSize":10485760,"totalChunks":2}'
```

Expected: `{"code":0,"message":"success","data":{"exists":false,"uploadId":"...","uploadedChunks":[]}}`

**Step 2: Test chunk upload**

```bash
# Create test chunk files
dd if=/dev/zero of=/tmp/chunk1 bs=5M count=1

curl -X POST http://localhost:8080/upload/chunk \
  -F "file=@/tmp/chunk1" \
  -F "uploadId=<uploadId_from_step1>" \
  -F "chunkNumber=1" \
  -F "totalChunks=2" \
  -F "md5=a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"
```

Expected: `{"code":0,"message":"success","data":{"chunkNumber":1,"uploaded":true}}`

---

## Commit Summary

After all tasks complete:

```bash
git add -A
git commit -m "feat(video): add chunked upload with instant upload support"
```