package repository

import (
	"context"
	"vidora-api/app/modules/tag/model"

	"github.com/lpphub/goweb/ext/dbx"
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

func (r *Repository) ListAll(ctx context.Context) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.DB().WithContext(ctx).Order("created_at ASC").Find(&tags).Error
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
		vts := make([]model.VideoTag, 0, len(tagIDs))
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

func (r *Repository) MoveToGroup(ctx context.Context, fromGroupID, toGroupID uint) error {
	return r.DB().WithContext(ctx).Model(&model.Tag{}).Where("group_id = ?", fromGroupID).Update("group_id", toGroupID).Error
}
