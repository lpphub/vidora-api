// module/tag/service.go
package tag

import (
	"context"
	"errors"

	"vidora-api/contract"
	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

// 确保实现 contract.TagBiz 接口
var _ contract.TagBiz = (*Service)(nil)

// Service 标签服务
type Service struct {
	repo *Repository
}

// NewService 创建标签服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateTag 创建普通标签
func (s *Service) CreateTag(ctx context.Context, name string) (*Tag, error) {
	exists, _ := s.repo.ExistsByName(ctx, name)
	if exists {
		return nil, errs.ErrTagExists
	}
	tag := &Tag{Name: name, Type: TypeNormal}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// CreateCategory 创建分类
func (s *Service) CreateCategory(ctx context.Context, name string, sortOrder int) (*Tag, error) {
	exists, _ := s.repo.ExistsByName(ctx, name)
	if exists {
		return nil, errs.ErrCategoryExists
	}
	tag := &Tag{Name: name, Type: TypeCategory, SortOrder: sortOrder}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// ListTags 获取普通标签列表
func (s *Service) ListTags(ctx context.Context) ([]Tag, error) {
	return s.repo.ListByType(ctx, TypeNormal)
}

// ListCategories 获取分类列表
func (s *Service) ListCategories(ctx context.Context) ([]Tag, error) {
	return s.repo.ListByType(ctx, TypeCategory)
}

// List 获取所有标签（包含分类）
func (s *Service) List(ctx context.Context) ([]Tag, error) {
	return s.repo.List(ctx)
}

// ExistByIDs 验证 ID 是否存在 - 实现 contract.TagBiz 接口
func (s *Service) ExistByIDs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	tags, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(tags) != len(ids) {
		return errs.ErrTagNotFound
	}
	return nil
}

// Get 获取标签
func (s *Service) Get(ctx context.Context, id uint) (*Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	return tag, err
}

// GetVideoTags 获取视频的标签
func (s *Service) GetVideoTags(ctx context.Context, videoID uint) ([]Tag, error) {
	return s.repo.GetVideoTags(ctx, videoID)
}

// SyncVideoTags 同步视频标签 - 实现 contract.TagBiz 接口
func (s *Service) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return s.repo.SyncVideoTags(ctx, videoID, tagIDs)
}
