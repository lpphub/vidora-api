package tag

import (
	"github.com/gin-gonic/gin"
	"vidora-api/server/helper"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) register(r *gin.RouterGroup) {
	r.POST("/tags", h.Create)
	r.GET("/tags", h.List)
	r.GET("/tags/:id", h.Get)
	r.PUT("/tags/:id", h.Update)
	r.DELETE("/tags/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Create(c.Request.Context(), req)
	helper.Respond(c, err, tag)
}

func (h *Handler) List(c *gin.Context) {
	typeStr := c.Query("type")
	var tagType *TagType
	if typeStr != "" {
		t := TagType(0)
		if typeStr == "1" {
			t = TypeCategory
		}
		tagType = &t
	}
	tags, err := h.svc.List(c.Request.Context(), tagType)
	helper.Respond(c, err, tags)
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	tag, err := h.svc.GetByID(c.Request.Context(), id)
	helper.Respond(c, err, tag)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	var req UpdateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Update(c.Request.Context(), id, req)
	helper.Respond(c, err, tag)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), id)
	helper.Respond(c, err, nil)
}
