// module/transcode/repository.go
package transcode

import (
	"context"
	"time"

	"github.com/lpphub/goweb/ext/dbx"
	"gorm.io/gorm"
)

// Repository 转码仓储
type Repository struct {
	*dbx.BaseRepo[Task]
}

// NewRepository 创建转码仓储
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepo: dbx.NewBaseRepo[Task](db),
	}
}

// ListByVideo 根据视频 ID 获取转码任务
func (r *Repository) ListByVideo(ctx context.Context, videoID uint) ([]Task, error) {
	var tasks []Task
	err := r.DB().WithContext(ctx).Where("video_id = ?", videoID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// ListByStatus 根据状态获取转码任务
func (r *Repository) ListByStatus(ctx context.Context, status *int8, page, pageSize int) ([]Task, int64, error) {
	var tasks []Task
	var total int64
	query := r.DB().WithContext(ctx).Model(&Task{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

// GetStats 获取转码统计
func (r *Repository) GetStats(ctx context.Context) (map[int8]int64, error) {
	type result struct {
		Status int8
		Count  int64
	}
	var results []result
	err := r.DB().WithContext(ctx).Model(&Task{}).
		Select("status, count(*) as count").
		Group("status").Scan(&results).Error

	stats := make(map[int8]int64)
	for _, res := range results {
		stats[res.Status] = res.Count
	}
	return stats, err
}

// MarkProcessing 标记为处理中
func (r *Repository) MarkProcessing(ctx context.Context, id uint) error {
	now := time.Now()
	return dbx.TxAwareDB(ctx, r.DB()).Model(&Task{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": StatusProcessing, "started_at": &now}).Error
}

// MarkCompleted 标记为完成
func (r *Repository) MarkCompleted(ctx context.Context, id uint, outputURL string) error {
	now := time.Now()
	return dbx.TxAwareDB(ctx, r.DB()).Model(&Task{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       StatusSuccess,
			"progress":     100,
			"output_url":   outputURL,
			"completed_at": &now,
		}).Error
}

// MarkFailed 标记为失败
func (r *Repository) MarkFailed(ctx context.Context, id uint, errMsg string) error {
	return dbx.TxAwareDB(ctx, r.DB()).Model(&Task{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": StatusFailed, "error_message": errMsg}).Error
}

// IncrementRetry 增加重试次数
func (r *Repository) IncrementRetry(ctx context.Context, id uint) error {
	return dbx.TxAwareDB(ctx, r.DB()).Model(&Task{}).Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}