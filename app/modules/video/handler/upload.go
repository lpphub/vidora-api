package handler

import (
	"vidora-api/app/modules/video/model"
	"vidora-api/app/modules/video/service"
	"vidora-api/app/server/helper"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) Register(r *gin.RouterGroup) {
	upload := r.Group("/upload")
	{
		upload.POST("/init", h.Init)
		upload.POST("/chunk", h.Chunk)
		upload.POST("/merge", h.Merge)
		upload.GET("/status/:uploadId", h.Status)
	}
}

func (h *UploadHandler) Init(c *gin.Context) {
	var req model.UploadInitReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.svc.Init(c.Request.Context(), &req)
	helper.Respond(c, err, resp)
}

func (h *UploadHandler) Chunk(c *gin.Context) {
	var req model.UploadChunkReq
	if err := c.ShouldBind(&req); err != nil {
		helper.Respond(c, err, nil)
		return
	}

	resp, err := h.svc.UploadChunk(c.Request.Context(), &req)
	helper.Respond(c, err, resp)
}

func (h *UploadHandler) Merge(c *gin.Context) {
	var req model.UploadMergeReq
	if !helper.MustBindJSON(c, &req) {
		return
	}

	resp, err := h.svc.Merge(c.Request.Context(), &req)
	helper.Respond(c, err, resp)
}

func (h *UploadHandler) Status(c *gin.Context) {
	uploadID := c.Param("uploadId")
	if uploadID == "" {
		helper.Respond(c, nil, nil)
		return
	}

	resp, err := h.svc.Status(c.Request.Context(), uploadID)
	helper.Respond(c, err, resp)
}
