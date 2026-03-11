// module/tag/handler.go
package tag

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// Handler 标签处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建标签处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/tags", h.Create)
	r.GET("/tags", h.List)
}

// Create 创建标签/分类
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required,max=50"`
		Type      int8   `json:"type"`      // 0=标签, 1=分类
		SortOrder int    `json:"sortOrder"` // 排序（仅分类使用）
	}
	if !helper.MustBindJSON(c, &req) {
		return
	}

	var tag *Tag
	var err error
	if req.Type == int8(TypeCategory) {
		tag, err = h.svc.CreateCategory(c, req.Name, req.SortOrder)
	} else {
		tag, err = h.svc.CreateTag(c, req.Name)
	}
	helper.Respond(c, err, tag)
}

// List 标签/分类列表
func (h *Handler) List(c *gin.Context) {
	tagType := c.Query("type") // 空=全部, 0=标签, 1=分类

	var tags []Tag
	var err error
	switch tagType {
	case "0":
		tags, err = h.svc.ListTags(c)
	case "1":
		tags, err = h.svc.ListCategories(c)
	default:
		tags, err = h.svc.List(c)
	}
	helper.Respond(c, err, tags)
}
