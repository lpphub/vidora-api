package repository

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"vidora-api/modules/tag/model"

	"gorm.io/gorm"
)

type GroupRepository struct {
	*dbx.BaseRepo[model.TagGroup]
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{
		BaseRepo: dbx.NewBaseRepo[model.TagGroup](db),
	}
}

func (r *GroupRepository) ListWithTags(ctx context.Context) ([]model.TagGroup, error) {
	var groups []model.TagGroup
	err := r.DB().WithContext(ctx).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&groups).Error
	return groups, err
}

func (r *GroupRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&model.TagGroup{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *GroupRepository) UpdateSortOrders(ctx context.Context, ids []uint) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.TagGroup{}).Where("id = ?", id).Update("sort_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GroupRepository) GetMaxSortOrder(ctx context.Context) (int, error) {
	var maxOrder int
	err := r.DB().WithContext(ctx).Model(&model.TagGroup{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}
