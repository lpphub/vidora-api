// module/tag/model.go
package tag

import (
	"time"

	"gorm.io/gorm"
)

// TagType 标签类型
type TagType int8

const (
	TypeNormal   TagType = 0 // 普通标签
	TypeCategory TagType = 1 // 分类
)

// Tag 标签模型
type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Type      TagType        `gorm:"default:0;index" json:"type"` // 0=普通标签, 1=分类
	SortOrder int            `gorm:"default:0" json:"sortOrder"`  // 排序（仅分类使用）
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
