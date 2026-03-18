package model

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
	GroupID    uint           `gorm:"not null;default:0;index:idx_group_name" json:"groupId"`
	Name       string         `gorm:"size:50;not null;index:idx_group_name" json:"name"`
	Type       TagType        `gorm:"default:0;index" json:"type,omitempty"`
	Color      string         `gorm:"size:7" json:"color,omitempty"`
	SortOrder  int            `gorm:"default:0" json:"sortOrder,omitempty"`
	Status     TagStatus      `gorm:"size:10;default:'enabled'" json:"status,omitempty"`
	UsageCount int            `gorm:"-" json:"usageCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (*Tag) TableName() string {
	return "tags"
}

type VideoTag struct {
	VideoID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}

func (*VideoTag) TableName() string {
	return "video_tags"
}
