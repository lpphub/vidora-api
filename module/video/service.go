// module/video/service.go
package video

import (
	"context"
	"errors"

	"vidora-api/port"
	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

// Service 视频服务
type Service struct {
	repo   *Repository
	tagSvc port.TagBiz
}

// NewService 创建视频服务
func NewService(repo *Repository, tagSvc port.TagBiz) *Service {
	return &Service{
		repo:   repo,
		tagSvc: tagSvc,
	}
}

// Create 创建视频
func (s *Service) Create(ctx context.Context, req port.CreateVideoReq) (*port.VideoDTO, error) {
	// 验证分类和标签
	ids := req.TagIDs
	if req.CategoryID > 0 {
		ids = append([]uint{req.CategoryID}, ids...)
	}
	if err := s.tagSvc.ExistByIDs(ctx, ids); err != nil {
		return nil, err
	}

	video := &Video{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		Status:      0,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return nil, err
	}

	if len(req.TagIDs) > 0 {
		_ = s.tagSvc.SyncVideoTags(ctx, video.ID, req.TagIDs)
	}

	return toDTO(video), nil
}

// Update 更新视频
func (s *Service) Update(ctx context.Context, id uint, req port.UpdateVideoReq) (*port.VideoDTO, error) {
	if _, err := s.repo.First(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrVideoNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CoverURL != "" {
		updates["cover_url"] = req.CoverURL
	}
	if req.VideoURL != "" {
		updates["video_url"] = req.VideoURL
	}
	if req.CategoryID > 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Duration > 0 {
		updates["duration"] = req.Duration
	}

	if len(updates) > 0 {
		s.repo.Update(ctx, id, updates)
	}

	if req.TagIDs != nil {
		s.tagSvc.SyncVideoTags(ctx, id, req.TagIDs)
	}

	video, err := s.repo.First(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDTO(video), nil
}

// Delete 删除视频
func (s *Service) Delete(ctx context.Context, id uint) error {
	exists, _ := s.repo.ExistsByID(ctx, id)
	if !exists {
		return errs.ErrVideoNotFound
	}
	return s.repo.Delete(ctx, id)
}

// Get 获取视频
func (s *Service) Get(ctx context.Context, id uint) (*port.VideoDTO, error) {
	video, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrVideoNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDTO(video), nil
}

// List 获取视频列表
func (s *Service) List(ctx context.Context, req port.VideoListReq) (*port.VideoListDTO, error) {
	videos, total, err := s.repo.ListWithFilter(ctx, req.CategoryID, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	dtos := make([]port.VideoDTO, len(videos))
	for i, v := range videos {
		dtos[i] = *toDTO(&v)
	}

	return &port.VideoListDTO{Total: total, List: dtos}, nil
}

// GetEntity 获取视频实体
func (s *Service) GetEntity(ctx context.Context, id uint) (*Video, error) {
	video, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrVideoNotFound
	}
	return video, err
}

// ListEntity 获取视频实体列表
func (s *Service) ListEntity(ctx context.Context, req VideoListReq) (*VideoListResp, error) {
	videos, total, err := s.repo.ListWithFilter(ctx, req.CategoryID, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &VideoListResp{Total: total, List: videos}, nil
}

func toDTO(video *Video) *port.VideoDTO {
	return &port.VideoDTO{
		ID:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		CoverURL:    video.CoverURL,
		VideoURL:    video.VideoURL,
		CategoryID:  video.CategoryID,
		Duration:    video.Duration,
		Status:      video.Status,
		PlayCount:   int(video.PlayCount),
	}
}