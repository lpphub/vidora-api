// module/transcode/handler.go
package transcode

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

// Handler 转码处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建转码处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes 注册路由
func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("/videos/:id/transcode", h.Create)
	r.GET("/transcodes", h.List)
	r.GET("/transcodes/:id", h.Get)
	r.POST("/transcodes/:id/retry", h.Retry)
	r.GET("/transcodes/stats", h.GetStats)
}

// Create 创建转码任务
func (h *Handler) Create(c *gin.Context) {
	videoID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req CreateTranscodeReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	task, err := h.svc.Create(c, uint(videoID), req.InputURL, req.Resolution, req.Bitrate)
	helper.Respond(c, err, task)
}

// List 转码任务列表
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		st, _ := strconv.ParseInt(s, 10, 8)
		st8 := int8(st)
		status = &st8
	}

	resp, err := h.svc.List(c, status, page, pageSize)
	helper.Respond(c, err, resp)
}

// Get 获取转码任务
func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	task, err := h.svc.Get(c, uint(id))
	helper.Respond(c, err, task)
}

// Retry 重试转码任务
func (h *Handler) Retry(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	task, err := h.svc.Retry(c, uint(id))
	helper.Respond(c, err, task)
}

// GetStats 获取转码统计
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c)
	helper.Respond(c, err, stats)
}