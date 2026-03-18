// module/transcode/model.go
package transcode

import (
	"time"

	"gorm.io/gorm"
)

// Task 转码任务模型
type Task struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	VideoID      uint           `gorm:"index" json:"videoId"`
	Status       int8           `json:"status"`
	Progress     int            `json:"progress"`
	InputURL     string         `gorm:"size:512" json:"inputUrl"`
	OutputURL    string         `gorm:"size:512" json:"outputUrl"`
	Resolution   string         `gorm:"size:20" json:"resolution"`
	Bitrate      string         `gorm:"size:20" json:"bitrate"`
	ErrorMessage string         `gorm:"type:text" json:"errorMessage"`
	RetryCount   int            `json:"retryCount"`
	StartedAt    *time.Time     `json:"startedAt"`
	CompletedAt  *time.Time     `json:"completedAt"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (*Task) TableName() string {
	return "transcode_tasks"
}

// Status constants
const (
	StatusPending    int8 = 0
	StatusProcessing int8 = 1
	StatusSuccess    int8 = 2
	StatusFailed     int8 = 3
)