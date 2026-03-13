package model

import (
	"time"

	"gorm.io/gorm"
)

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

func (*Video) TableName() string {
	return "videos"
}

func (v *Video) IsPublished() bool {
	return v.Status == 1
}

type UploadFile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FileKey    string    `gorm:"uniqueIndex;size:64;not null" json:"fileKey"`
	MD5        string    `gorm:"uniqueIndex;size:32;not null" json:"md5"`
	FileName   string    `gorm:"size:255;not null" json:"fileName"`
	FileSize   int64     `gorm:"not null" json:"fileSize"`
	StorageURL string    `gorm:"size:512" json:"storageUrl"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (*UploadFile) TableName() string {
	return "upload_files"
}

type UploadSession struct {
	UploadID       string `json:"uploadId"`
	MD5            string `json:"md5"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	TotalChunks    int    `json:"totalChunks"`
	UploadedChunks []int  `json:"uploadedChunks"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}
