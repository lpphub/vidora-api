// module/tag/repository.go
package tag

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

// Repository 标签仓储
type Repository struct {
	*dbx.BaseRepo[Tag]
}

// NewRepository 创建标签仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[Tag](db),
	}
}

// List 获取所有标签
func (r *Repository) List(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Order("name ASC").Find(&tags).Error
	return tags, err
}

// ListByType 按类型获取标签
func (r *Repository) ListByType(ctx context.Context, tagType TagType) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).Where("type = ?", tagType).Order("sort_order ASC, name ASC").Find(&tags).Error
	return tags, err
}

// ExistsByName 检查名称是否存在
func (r *Repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsCategoryByID 检查分类是否存在
func (r *Repository) ExistsCategoryByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&Tag{}).Where("id = ? AND type = ?", id, TypeCategory).Count(&count).Error
	return count > 0, err
}

// SyncVideoTags 同步视频标签 - 实现 port.TagRepository 接口
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

// GetVideoTags 获取视频的标签
func (r *Repository) GetVideoTags(ctx context.Context, videoID uint) ([]Tag, error) {
	var tags []Tag
	err := r.DB().WithContext(ctx).
		Table("tags").
		Joins("JOIN video_tags ON video_tags.tag_id = tags.id").
		Where("video_tags.video_id = ?", videoID).
		Find(&tags).Error
	return tags, err
}
