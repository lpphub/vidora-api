package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"vidora-api/infra/storage"
	"vidora-api/module/video/model"
	"vidora-api/module/video/repository"
	"vidora-api/shared/errs"
)

func GenerateUploadID(md5 string) string {
	return fmt.Sprintf("%s_%d", md5, time.Now().UnixNano())
}

func GenerateFileKey(md5 string) string {
	return fmt.Sprintf("v_%s_%s", md5[:8], time.Now().Format("20060102"))
}

type UploadService struct {
	repo         *repository.UploadRepository
	storage      storage.Client
	mergeLimiter *MergeLimiter
}

func NewUploadService(repo *repository.UploadRepository, st storage.Client) *UploadService {
	return &UploadService{
		repo:         repo,
		storage:      st,
		mergeLimiter: NewMergeLimiter(5),
	}
}

func (s *UploadService) Init(ctx context.Context, req *model.UploadInitReq) (*model.UploadInitResp, error) {
	uploadFile, err := s.repo.FindUploadFileByMD5(ctx, req.MD5)
	if err == nil {
		return &model.UploadInitResp{
			Exists:  true,
			FileKey: uploadFile.FileKey,
			URL:     uploadFile.StorageURL,
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	session, uploadID, err := s.repo.FindExistingSessionByMD5(ctx, req.MD5)
	if err == nil && session != nil {
		return &model.UploadInitResp{
			Exists:         false,
			UploadID:       uploadID,
			UploadedChunks: session.UploadedChunks,
		}, nil
	}

	uploadID = GenerateUploadID(req.MD5)
	newSession := &model.UploadSession{
		UploadID:       uploadID,
		MD5:            req.MD5,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		TotalChunks:    req.TotalChunks,
		UploadedChunks: []int{},
	}
	if err := s.repo.CreateSession(ctx, uploadID, newSession); err != nil {
		return nil, err
	}

	return &model.UploadInitResp{
		Exists:         false,
		UploadID:       uploadID,
		UploadedChunks: []int{},
	}, nil
}

func (s *UploadService) UploadChunk(ctx context.Context, req *model.UploadChunkReq) (*model.UploadChunkResp, error) {
	session, err := s.repo.GetSession(ctx, req.UploadID)
	if err != nil {
		return nil, errs.ErrUploadSessionNotFound
	}

	if session.MD5 != req.MD5 {
		return nil, errs.ErrInvalidParam
	}

	if err := s.repo.SaveChunk(ctx, req.UploadID, req.ChunkNumber, req.File); err != nil {
		return nil, errs.ErrChunkUploadFailed
	}

	uploadedChunks := append(session.UploadedChunks, req.ChunkNumber-1)
	if err := s.repo.UpdateSessionChunks(ctx, req.UploadID, uploadedChunks); err != nil {
		return nil, err
	}

	return &model.UploadChunkResp{
		Uploaded:    true,
		ChunkNumber: req.ChunkNumber,
	}, nil
}

func (s *UploadService) Merge(ctx context.Context, req *model.UploadMergeReq) (*model.UploadMergeResp, error) {
	session, err := s.repo.GetSession(ctx, req.UploadID)
	if err != nil {
		return nil, errs.ErrUploadSessionNotFound
	}

	if session.MD5 != req.MD5 || session.TotalChunks != req.TotalChunks {
		return nil, errs.ErrInvalidParam
	}

	if len(session.UploadedChunks) != req.TotalChunks {
		return nil, errs.ErrChunksIncomplete
	}

	if err := s.mergeLimiter.Acquire(ctx); err != nil {
		return nil, err
	}
	defer s.mergeLimiter.Release()

	fileKey := GenerateFileKey(session.MD5)
	destObjectName := fmt.Sprintf("videos/%s/%s", fileKey, session.FileName)

	if err := s.repo.MergeChunks(ctx, req.UploadID, req.TotalChunks, destObjectName); err != nil {
		return nil, err
	}

	storageURL := s.storage.GetURL(destObjectName)

	uploadFile := &model.UploadFile{
		FileKey:    fileKey,
		MD5:        session.MD5,
		FileName:   session.FileName,
		FileSize:   session.FileSize,
		StorageURL: storageURL,
	}
	if err := s.repo.CreateUploadFile(ctx, uploadFile); err != nil {
		return nil, err
	}

	s.repo.CleanupChunks(ctx, req.UploadID, req.TotalChunks)
	s.repo.DeleteSession(ctx, req.UploadID)

	return &model.UploadMergeResp{
		FileKey: fileKey,
		URL:     storageURL,
	}, nil
}

func (s *UploadService) Status(ctx context.Context, uploadID string) (*model.UploadStatusResp, error) {
	session, err := s.repo.GetSession(ctx, uploadID)
	if err != nil {
		return nil, errs.ErrUploadSessionNotFound
	}

	uploadedCount := len(session.UploadedChunks)
	progress := 0
	if session.TotalChunks > 0 {
		progress = uploadedCount * 100 / session.TotalChunks
	}

	eta := s.calculateETA(session)

	status := "pending"
	if uploadedCount > 0 && uploadedCount < session.TotalChunks {
		status = "uploading"
	} else if uploadedCount == session.TotalChunks {
		status = "completed"
	}

	return &model.UploadStatusResp{
		UploadID:       session.UploadID,
		FileName:       session.FileName,
		FileSize:       session.FileSize,
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedCount,
		Progress:       progress,
		ETA:            eta,
		Status:         status,
	}, nil
}

func (s *UploadService) calculateETA(session *model.UploadSession) int {
	if len(session.UploadedChunks) == 0 || session.UpdatedAt == session.CreatedAt {
		return 0
	}

	elapsed := session.UpdatedAt - session.CreatedAt
	if elapsed <= 0 {
		return 0
	}

	chunksPerSecond := float64(len(session.UploadedChunks)) / float64(elapsed)
	if chunksPerSecond <= 0 {
		return 0
	}

	remainingChunks := session.TotalChunks - len(session.UploadedChunks)
	eta := int(float64(remainingChunks) / chunksPerSecond)

	return eta
}
