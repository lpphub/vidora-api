package service

import (
	"context"
	"errors"
	model2 "vidora-api/app/modules/tag/model"
	repository2 "vidora-api/app/modules/tag/repository"
	"vidora-api/app/shared/errs"

	"gorm.io/gorm"
)

type GroupService struct {
	repo    *repository2.GroupRepository
	tagRepo *repository2.Repository
}

func NewGroupService(repo *repository2.GroupRepository, tagRepo *repository2.Repository) *GroupService {
	return &GroupService{repo: repo, tagRepo: tagRepo}
}

type GroupReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

type ReorderGroupsReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

type TagGroupWithTags struct {
	model2.TagGroup
	TagList []model2.Tag `json:"tagList"`
}

func (s *GroupService) List(ctx context.Context) ([]TagGroupWithTags, error) {
	groups, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	allTags, err := s.tagRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[uint][]model2.Tag)
	for _, t := range allTags {
		tagMap[t.GroupID] = append(tagMap[t.GroupID], t)
	}

	result := make([]TagGroupWithTags, 0, len(groups)+1)
	result = append(result, TagGroupWithTags{
		TagGroup: model2.TagGroup{
			ID:        0,
			Name:      "默认",
			SortOrder: 0,
		},
		TagList: tagMap[0],
	})

	for _, g := range groups {
		result = append(result, TagGroupWithTags{
			TagGroup: g,
			TagList:  tagMap[g.ID],
		})
	}
	return result, nil
}

func (s *GroupService) Create(ctx context.Context, req GroupReq) (*TagGroupWithTags, error) {
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrTagGroupExists
	}
	maxOrder, err := s.repo.GetMaxSortOrder(ctx)
	if err != nil {
		return nil, err
	}
	group := &model2.TagGroup{
		Name:      req.Name,
		SortOrder: maxOrder + 1,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		return nil, err
	}
	return &TagGroupWithTags{TagGroup: *group, TagList: []model2.Tag{}}, nil
}

func (s *GroupService) Update(ctx context.Context, id uint, req GroupReq) error {
	group, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrTagGroupNotFound
	}
	if err != nil {
		return err
	}
	if req.Name != group.Name {
		exists, err := s.repo.ExistsByName(ctx, req.Name)
		if err != nil {
			return err
		}
		if exists {
			return errs.ErrTagGroupExists
		}
		return s.repo.Update(ctx, id, map[string]any{"name": req.Name})
	}
	return nil
}

func (s *GroupService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errs.ErrCannotDeleteDefaultGroup
	}
	_, err := s.repo.First(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.tagRepo.MoveToGroup(ctx, id, 0); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) Reorder(ctx context.Context, ids []uint) error {
	return s.repo.UpdateSortOrders(ctx, ids)
}
