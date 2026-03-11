// module/transcode/service.go
package transcode

import (
	"context"
	"errors"

	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

// Service 转码服务
type Service struct {
	repo *Repository
}

// NewService 创建转码服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create 创建转码任务
func (s *Service) Create(ctx context.Context, videoID uint, inputURL, resolution, bitrate string) (*Task, error) {
	task := &Task{
		VideoID:    videoID,
		InputURL:   inputURL,
		Resolution: resolution,
		Bitrate:    bitrate,
		Status:     StatusPending,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// List 获取转码任务列表
func (s *Service) List(ctx context.Context, status *int8, page, pageSize int) (*TranscodeListResp, error) {
	tasks, total, err := s.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &TranscodeListResp{Total: total, List: tasks}, nil
}

// Get 获取转码任务
func (s *Service) Get(ctx context.Context, id uint) (*Task, error) {
	task, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTranscodeNotFound
	}
	return task, err
}

// Retry 重试转码任务
func (s *Service) Retry(ctx context.Context, id uint) (*Task, error) {
	task, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.Status != StatusFailed {
		return nil, errs.ErrTranscodeNotFailed
	}

	s.repo.Update(ctx, id, map[string]interface{}{
		"status":        StatusPending,
		"progress":      0,
		"started_at":    nil,
		"completed_at":  nil,
		"error_message": "",
	})
	s.repo.IncrementRetry(ctx, id)

	return s.repo.First(ctx, id)
}

// GetStats 获取转码统计
func (s *Service) GetStats(ctx context.Context) (*TranscodeStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &TranscodeStats{
		Pending:    stats[StatusPending],
		Processing: stats[StatusProcessing],
		Success:    stats[StatusSuccess],
		Failed:     stats[StatusFailed],
	}, nil
}