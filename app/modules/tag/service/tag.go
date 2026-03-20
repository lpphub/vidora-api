package service

import (
	"context"
	"errors"
	"vidora-api/app/modules/tag/model"
	"vidora-api/app/modules/tag/repository"
	"vidora-api/app/shared/contracts"
	"vidora-api/app/shared/errs"

	"gorm.io/gorm"
)

var _ contracts.TagBiz = (*Service)(nil)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

type TagReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type CreateTagInGroupReq struct {
	Name    string `json:"name" binding:"required,max=50"`
	GroupID uint   `json:"-"`
}

func (s *Service) Create(ctx context.Context, req CreateTagInGroupReq) (*model.Tag, error) {
	exists, err := s.repo.ExistsByName(ctx, req.Name, req.GroupID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrTagExists
	}

	tag := &model.Tag{
		GroupID: req.GroupID,
		Name:    req.Name,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*model.Tag, error) {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagNotFound
	}
	return tag, err
}

func (s *Service) Update(ctx context.Context, id uint, req TagReq) error {
	tag, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrTagNotFound
	}
	if err != nil {
		return err
	}

	if req.Name != tag.Name {
		exists, err := s.repo.ExistsByName(ctx, req.Name, tag.GroupID)
		if err != nil {
			return err
		}
		if exists {
			return errs.ErrTagExists
		}
		return s.repo.Update(ctx, id, map[string]any{"name": req.Name})
	}
	return nil
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

func (s *Service) GetVideoTags(ctx context.Context, videoID uint) ([]model.Tag, error) {
	return s.repo.GetVideoTags(ctx, videoID)
}

func (s *Service) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return s.repo.SyncVideoTags(ctx, videoID, tagIDs)
}
