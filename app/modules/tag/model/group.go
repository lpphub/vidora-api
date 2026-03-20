package model

import (
	"time"

	"gorm.io/gorm"
)

type TagGroup struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (*TagGroup) TableName() string {
	return "tag_groups"
}
