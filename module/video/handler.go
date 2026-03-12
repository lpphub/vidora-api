// module/video/handler.go
package video

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"vidora-api/contract"
	"vidora-api/server/helper"
)

// Handler 视频处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建视频处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/videos", h.Create)
	r.GET("/videos", h.List)
	r.GET("/videos/:id", h.Get)
	r.PUT("/videos/:id", h.Update)
	r.DELETE("/videos/:id", h.Delete)
}

// Create 创建视频
func (h *Handler) Create(c *gin.Context) {
	var req CreateVideoReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	contractReq := contract.CreateVideoReq{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		TagIDs:      req.TagIDs,
	}

	dto, err := h.svc.Create(c, contractReq)
	helper.Respond(c, err, dto)
}

// Update 更新视频
func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateVideoReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	contractReq := contract.UpdateVideoReq{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		CategoryID:  req.CategoryID,
		Duration:    req.Duration,
		Status:      req.Status,
		TagIDs:      req.TagIDs,
	}

	dto, err := h.svc.Update(c, uint(id), contractReq)
	helper.Respond(c, err, dto)
}

// Delete 删除视频
func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := h.svc.Delete(c, uint(id))
	helper.Respond(c, err, nil)
}

// Get 获取视频
func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	video, err := h.svc.GetEntity(c, uint(id))
	helper.Respond(c, err, video)
}

// List 视频列表
func (h *Handler) List(c *gin.Context) {
	var req VideoListReq
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	resp, err := h.svc.ListEntity(c, req)
	helper.Respond(c, err, resp)
}
