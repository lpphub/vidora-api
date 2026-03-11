// module/video/dto.go
package video

// CreateVideoReq 创建视频请求
type CreateVideoReq struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	VideoURL    string `json:"videoUrl" binding:"required"`
	CategoryID  uint   `json:"categoryId" binding:"required"`
	Duration    int    `json:"duration"`
	TagIDs      []uint `json:"tagIds"`
}

// UpdateVideoReq 更新视频请求
type UpdateVideoReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	VideoURL    string `json:"videoUrl"`
	CategoryID  uint   `json:"categoryId"`
	Status      *int8  `json:"status"`
	Duration    int    `json:"duration"`
	TagIDs      []uint `json:"tagIds"`
}

// VideoListReq 视频列表请求
type VideoListReq struct {
	CategoryID uint   `form:"categoryId"`
	Status     *int8  `form:"status"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"pageSize" binding:"min=1,max=100"`
}

// VideoListResp 视频列表响应
type VideoListResp struct {
	Total int64   `json:"total"`
	List  []Video `json:"list"`
}