package tag

import (
	"time"

	"gorm.io/gorm"
)

type TagType int8
type TagStatus string

const (
	TypeNormal   TagType = 0
	TypeCategory TagType = 1

	StatusEnabled  TagStatus = "enabled"
	StatusDisabled TagStatus = "disabled"
)

type Tag struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Type       TagType        `gorm:"default:0;index" json:"type"`
	Color      string         `gorm:"size:7" json:"color,omitempty"`
	SortOrder  int            `gorm:"default:0" json:"sortOrder,omitempty"`
	Status     TagStatus      `gorm:"size:10;default:'enabled'" json:"status,omitempty"`
	UsageCount int            `gorm:"default:0" json:"usageCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (*Tag) TableName() string {
	return "tags"
}

// IsCategory 是否是分类
func (t *Tag) IsCategory() bool {
	return t.Type == TypeCategory
}

// VideoTag 视频标签关联
type VideoTag struct {
	VideoID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}

// TableName 指定表名
func (*VideoTag) TableName() string {
	return "video_tags"
}
