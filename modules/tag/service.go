package tag

import (
	"context"
	"errors"

	"vidora-api/contract"
	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

var _ contract.TagBiz = (*Service)(nil)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateTagReq struct {
	Name      string    `json:"name" binding:"required,max=50"`
	Type      TagType   `json:"type"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sortOrder"`
	Status    TagStatus `json:"status"`
}

type UpdateTagReq struct {
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder *int      `json:"sortOrder"`
	Status    TagStatus `json:"status"`
}

func (s *Service) Create(ctx context.Context, req CreateTagReq) (*Tag, error) {
	exists, _ := s.repo.ExistsByName(ctx, req.Name)
	if exists {
		if req.Type == TypeCategory {
			return nil, errs.ErrCategoryExists
		}
		return nil, errs.ErrTagExists
	}

	status := StatusEnabled
	if req.Status != "" {
		status = req.Status
	}

	tag := &Tag{
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		Status:    status,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}

	tag.UsageCount = 0
	return tag, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	if err != nil {
		return nil, err
	}
	count, err := s.repo.GetUsageCount(ctx, id)
	if err != nil {
		return nil, err
	}
	tag.UsageCount = int(count)
	return tag, nil
}

func (s *Service) List(ctx context.Context, tagType *TagType) ([]Tag, error) {
	var tags []Tag
	var err error
	if tagType != nil {
		tags, err = s.repo.ListByType(ctx, *tagType)
	} else {
		tags, err = s.repo.List(ctx)
	}
	if err != nil {
		return nil, err
	}
	for i := range tags {
		count, _ := s.repo.GetUsageCount(ctx, tags[i].ID)
		tags[i].UsageCount = int(count)
	}
	return tags, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateTagReq) (*Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	if err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	if req.Name != "" && req.Name != tag.Name {
		exists, _ := s.repo.ExistsByName(ctx, req.Name)
		if exists {
			return nil, errs.ErrTagExists
		}
		updates["name"] = req.Name
	}
	if tag.Type == TypeNormal && req.Color != "" {
		updates["color"] = req.Color
	}
	if tag.Type == TypeCategory {
		if req.SortOrder != nil {
			updates["sort_order"] = *req.SortOrder
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return nil, err
		}
		tag, err = s.repo.First(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	count, err := s.repo.GetUsageCount(ctx, id)
	if err != nil {
		return nil, err
	}
	tag.UsageCount = int(count)
	return tag, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrTagNotFound
	}
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

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

func (s *Service) GetVideoTags(ctx context.Context, videoID uint) ([]Tag, error) {
	return s.repo.GetVideoTags(ctx, videoID)
}

func (s *Service) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return s.repo.SyncVideoTags(ctx, videoID, tagIDs)
}
