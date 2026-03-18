package repository

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"vidora-api/modules/tag/model"

	"gorm.io/gorm"
)

type Repository struct {
	*dbx.BaseRepo[model.Tag]
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[model.Tag](db),
	}
}

func (r *Repository) ListByGroup(ctx context.Context, groupID uint) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.DB().WithContext(ctx).Where("group_id = ?", groupID).Order("created_at ASC").Find(&tags).Error
	return tags, err
}

func (r *Repository) ExistsByName(ctx context.Context, name string, groupID uint) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&model.Tag{}).Where("name = ? AND group_id = ?", name, groupID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Where("video_id = ?", videoID).Delete(&model.VideoTag{})
		if len(tagIDs) == 0 {
			return nil
		}
		var vts []model.VideoTag
		for _, tid := range tagIDs {
			vts = append(vts, model.VideoTag{VideoID: videoID, TagID: tid})
		}
		return tx.Create(&vts).Error
	})
}

func (r *Repository) GetVideoTags(ctx context.Context, videoID uint) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.DB().WithContext(ctx).
		Table("tags").
		Joins("JOIN video_tags ON video_tags.tag_id = tags.id").
		Where("video_tags.video_id = ?", videoID).
		Find(&tags).Error
	return tags, err
}

func (r *Repository) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.DB().WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.DB().WithContext(ctx).Delete(&model.Tag{}, id).Error
}

func (r *Repository) MoveToGroup(ctx context.Context, fromGroupID, toGroupID uint) error {
	return r.DB().WithContext(ctx).Model(&model.Tag{}).Where("group_id = ?", fromGroupID).Update("group_id", toGroupID).Error
}
