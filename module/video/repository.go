// module/video/repository.go
package video

import (
	"context"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

// Repository 视频仓储
type Repository struct {
	*dbx.BaseRepo[Video]
}

// NewRepository 创建视频仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[Video](db),
	}
}

// ExistsByID 检查视频是否存在
func (r *Repository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&Video{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ListWithFilter 带过滤条件的列表查询
func (r *Repository) ListWithFilter(ctx context.Context, categoryID uint, status *int8, keyword string, page, pageSize int) ([]Video, int64, error) {
	var videos []Video
	var total int64

	query := r.DB().WithContext(ctx).Model(&Video{})
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error
	return videos, total, err
}

// IncrementPlayCount 增加播放次数
func (r *Repository) IncrementPlayCount(ctx context.Context, id uint) error {
	return dbx.TxAwareDB(ctx, r.DB()).Model(&Video{}).Where("id = ?", id).
		UpdateColumn("play_count", gorm.Expr("play_count + 1")).Error
}