// contracts/video.go
package contracts

import "context"

// VideoBiz 视频服务接口（供外部调用）
type VideoBiz interface {
	Create(ctx context.Context, req CreateVideoReq) (*VideoDTO, error)
	Update(ctx context.Context, id uint, req UpdateVideoReq) (*VideoDTO, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*VideoDTO, error)
	List(ctx context.Context, req VideoListReq) (*VideoListDTO, error)
}

// CreateVideoReq 创建视频请求
type CreateVideoReq struct {
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	TagIDs      []uint
}

// UpdateVideoReq 更新视频请求
type UpdateVideoReq struct {
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	Status      *int8
	TagIDs      []uint
}

// VideoListReq 视频列表请求
type VideoListReq struct {
	CategoryID uint
	Status     *int8
	Keyword    string
	Page       int
	PageSize   int
}

// VideoDTO 视频数据传输对象
type VideoDTO struct {
	ID          uint
	Title       string
	Description string
	CoverURL    string
	VideoURL    string
	CategoryID  uint
	Duration    int
	Status      int8
	PlayCount   int
}

// VideoListDTO 视频列表响应
type VideoListDTO struct {
	Total int64
	List  []VideoDTO
}
