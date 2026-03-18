package service

import (
	"context"
	"errors"

	"vidora-api/modules/tag/model"
	"vidora-api/modules/tag/repository"
	"vidora-api/shared/errs"

	"gorm.io/gorm"
)

type GroupService struct {
	repo    *repository.GroupRepository
	tagRepo *repository.Repository
}

func NewGroupService(repo *repository.GroupRepository, tagRepo *repository.Repository) *GroupService {
	return &GroupService{repo: repo, tagRepo: tagRepo}
}

type CreateGroupReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type UpdateGroupReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type ReorderGroupsReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

const DefaultGroupID uint = 0

func (s *GroupService) List(ctx context.Context) ([]model.TagGroup, error) {
	groups, err := s.repo.ListWithTags(ctx)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		for j := range groups[i].Tags {
			count, _ := s.tagRepo.GetUsageCount(ctx, groups[i].Tags[j].ID)
			groups[i].Tags[j].UsageCount = int(count)
		}
	}
	return groups, nil
}

func (s *GroupService) Create(ctx context.Context, req CreateGroupReq) (*model.TagGroup, error) {
	exists, _ := s.repo.ExistsByName(ctx, req.Name)
	if exists {
		return nil, errs.ErrTagGroupExists
	}
	maxOrder, _ := s.repo.GetMaxSortOrder(ctx)
	group := &model.TagGroup{
		Name:      req.Name,
		SortOrder: maxOrder + 1,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		return nil, err
	}
	group.Tags = []model.Tag{}
	return group, nil
}

func (s *GroupService) Update(ctx context.Context, id uint, req UpdateGroupReq) (*model.TagGroup, error) {
	group, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrTagGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	if req.Name != group.Name {
		exists, _ := s.repo.ExistsByName(ctx, req.Name)
		if exists {
			return nil, errs.ErrTagGroupExists
		}
		if err := s.repo.Update(ctx, id, map[string]any{"name": req.Name}); err != nil {
			return nil, err
		}
		group.Name = req.Name
	}
	group.Tags, _ = s.tagRepo.ListByGroup(ctx, id)
	return group, nil
}

func (s *GroupService) Delete(ctx context.Context, id uint) error {
	if id == DefaultGroupID {
		return errs.ErrCannotDeleteDefaultGroup
	}
	_, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.tagRepo.MoveToGroup(ctx, id, DefaultGroupID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) Reorder(ctx context.Context, ids []uint) error {
	return s.repo.UpdateSortOrders(ctx, ids)
}
