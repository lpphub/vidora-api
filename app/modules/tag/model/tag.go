package model

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"not null;default:0;index" json:"-"`
	Name      string         `gorm:"size:50;not null;index:idx_group_name" json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
