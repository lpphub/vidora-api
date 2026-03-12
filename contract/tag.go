// contract/tag.go
package contract

import "context"

// TagBiz 标签服务接口（供 Video 等模块调用）
type TagBiz interface {
	ExistByIDs(ctx context.Context, ids []uint) error
	SyncVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error
}

// TagDTO 标签数据传输对象
type TagDTO struct {
	ID   uint
	Name string
	Type int8
}
