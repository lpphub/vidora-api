// module/video/model.go
package video

import (
	"time"

	"gorm.io/gorm"
)

// Video 视频模型
type Video struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CoverURL    string         `gorm:"size:512" json:"coverUrl"`
	VideoURL    string         `gorm:"size:512;not null" json:"videoUrl"`
	CategoryID  uint           `gorm:"index" json:"categoryId"`
	Status      int8           `gorm:"default:0" json:"status"`
	PlayCount   uint64         `gorm:"default:0" json:"playCount"`
	Duration    int            `gorm:"default:0" json:"duration"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (*Video) TableName() string {
	return "videos"
}

// IsPublished 是否已发布
func (v *Video) IsPublished() bool {
	return v.Status == 1
}

// VideoOutput 视频输出
type VideoOutput struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	VideoID    uint      `gorm:"index" json:"videoId"`
	Resolution string    `gorm:"size:20" json:"resolution"`
	Bitrate    string    `gorm:"size:20" json:"bitrate"`
	VideoURL   string    `gorm:"size:512" json:"videoUrl"`
	FileSize   int64     `json:"fileSize"`
	Status     int8      `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TableName 指定表名
func (*VideoOutput) TableName() string {
	return "video_outputs"
}
