package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"time"
	"vidora-api/app/infra/storage"
	"vidora-api/app/modules/video/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	uploadKeyPrefix = "upload:"
	uploadTTL       = 24 * time.Hour
)

type UploadRepository struct {
	db      *gorm.DB
	rdb     *redis.Client
	storage storage.Client
}

func NewUploadRepository(db *gorm.DB, rdb *redis.Client, st storage.Client) *UploadRepository {
	return &UploadRepository{
		db:      db,
		rdb:     rdb,
		storage: st,
	}
}

func (r *UploadRepository) FindUploadFileByMD5(ctx context.Context, md5 string) (*model.UploadFile, error) {
	var file model.UploadFile
	err := r.db.WithContext(ctx).Where("md5 = ?", md5).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *UploadRepository) CreateUploadFile(ctx context.Context, file *model.UploadFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *UploadRepository) CreateSession(ctx context.Context, uploadID string, session *model.UploadSession) error {
	now := time.Now().Unix()
	session.CreatedAt = now
	session.UpdatedAt = now
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, r.sessionKey(uploadID), data, uploadTTL).Err()
}

func (r *UploadRepository) GetSession(ctx context.Context, uploadID string) (*model.UploadSession, error) {
	data, err := r.rdb.Get(ctx, r.sessionKey(uploadID)).Bytes()
	if err != nil {
		return nil, err
	}
	var session model.UploadSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UploadRepository) UpdateSessionChunks(ctx context.Context, uploadID string, chunks []int) error {
	session, err := r.GetSession(ctx, uploadID)
	if err != nil {
		return err
	}
	session.UploadedChunks = chunks
	session.UpdatedAt = time.Now().Unix()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, r.sessionKey(uploadID), data, uploadTTL).Err()
}

func (r *UploadRepository) DeleteSession(ctx context.Context, uploadID string) error {
	return r.rdb.Del(ctx, r.sessionKey(uploadID)).Err()
}

func (r *UploadRepository) SaveChunk(ctx context.Context, uploadID string, chunkNumber int, file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	objectName := r.chunkObjectName(uploadID, chunkNumber)
	_, err = r.storage.Upload(ctx, objectName, src, file.Size, "application/octet-stream")
	return err
}

func (r *UploadRepository) MergeChunks(ctx context.Context, uploadID string, totalChunks int, destObjectName string) error {
	var sources []string
	for i := 1; i <= totalChunks; i++ {
		sources = append(sources, r.chunkObjectName(uploadID, i))
	}

	return r.storage.ComposeObject(ctx, sources, destObjectName)
}

func (r *UploadRepository) CleanupChunks(ctx context.Context, uploadID string, totalChunks int) {
	for i := 1; i <= totalChunks; i++ {
		r.storage.Delete(ctx, r.chunkObjectName(uploadID, i))
	}
}

func (r *UploadRepository) sessionKey(uploadID string) string {
	return uploadKeyPrefix + uploadID
}

func (r *UploadRepository) chunkObjectName(uploadID string, chunkNumber int) string {
	return fmt.Sprintf("chunks/%s/%d", uploadID, chunkNumber)
}

func (r *UploadRepository) SessionExists(ctx context.Context, uploadID string) bool {
	_, err := r.GetSession(ctx, uploadID)
	return err == nil
}

func (r *UploadRepository) FindExistingSessionByMD5(ctx context.Context, md5 string) (*model.UploadSession, string, error) {
	var uploadID string
	iter := r.rdb.Scan(ctx, 0, uploadKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := r.rdb.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var session model.UploadSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.MD5 == md5 {
			uploadID = session.UploadID
			return &session, uploadID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return nil, "", err
	}
	return nil, "", errors.New("session not found")
}
