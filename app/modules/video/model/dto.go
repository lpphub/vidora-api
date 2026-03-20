package model

import "mime/multipart"

type UploadInitReq struct {
	MD5         string `json:"md5" binding:"required,len=32"`
	FileName    string `json:"fileName" binding:"required,max=255"`
	FileSize    int64  `json:"fileSize" binding:"required,min=1"`
	TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

type UploadInitResp struct {
	Exists         bool   `json:"exists"`
	FileKey        string `json:"fileKey,omitempty"`
	URL            string `json:"url,omitempty"`
	UploadID       string `json:"uploadId"`
	UploadedChunks []int  `json:"uploadedChunks"`
}

type UploadChunkReq struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	UploadID    string                `form:"uploadId" binding:"required"`
	ChunkNumber int                   `form:"chunkNumber" binding:"required,min=1"`
	TotalChunks int                   `form:"totalChunks" binding:"required"`
	MD5         string                `form:"md5" binding:"required,len=32"`
}

type UploadChunkResp struct {
	Uploaded    bool `json:"uploaded"`
	ChunkNumber int  `json:"chunkNumber"`
}

type UploadMergeReq struct {
	UploadID    string `json:"uploadId" binding:"required"`
	MD5         string `json:"md5" binding:"required,len=32"`
	FileName    string `json:"fileName" binding:"required"`
	TotalChunks int    `json:"totalChunks" binding:"required,min=1"`
}

type UploadMergeResp struct {
	FileKey string `json:"fileKey"`
	URL     string `json:"url"`
}

type UploadStatusResp struct {
	UploadID       string `json:"uploadId"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	TotalChunks    int    `json:"totalChunks"`
	UploadedChunks int    `json:"uploadedChunks"`
	Progress       int    `json:"progress"`
	ETA            int    `json:"eta"`
	Status         string `json:"status"`
}
