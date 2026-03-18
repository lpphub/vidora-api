package tag

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

type Repository struct {
	*dbx.BaseRepo[Tag]
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[Tag](db),
	}
}

func (r *Repository) List(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Order("name ASC").Find(&tags).Error
	return tags, err
}

func (r *Repository) ListByType(ctx context.Context, tagType TagType) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Where("type = ?", tagType).Order("sort_order ASC, name ASC").Find(&tags).Error
	return tags, err
}

func (r *Repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *Repository) SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	return r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Where("video_id = ?", videoID).Delete(&VideoTag{})
		if len(tagIDs) == 0 {
			return nil
		}
		var vts []VideoTag
		for _, tid := range tagIDs {
			vts = append(vts, VideoTag{VideoID: videoID, TagID: tid})
		}
		return tx.Create(&vts).Error
	})
}

func (r *Repository) GetVideoTags(ctx context.Context, videoID uint) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).
		Table("tags").
		Joins("JOIN video_tags ON video_tags.tag_id = tags.id").
		Where("video_tags.video_id = ?", videoID).
		Find(&tags).Error
	return tags, err
}

func (r *Repository) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.DB().WithContext(ctx).Model(&Tag{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.DB().WithContext(ctx).Delete(&Tag{}, id).Error
}

func (r *Repository) GetUsageCount(ctx context.Context, tagID uint) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&VideoTag{}).Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}
