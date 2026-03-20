package handler

import (
	service2 "vidora-api/app/modules/tag/service"
	"vidora-api/app/server/helper"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc      *service2.Service
	groupSvc *service2.GroupService
}

func NewHandler(svc *service2.Service, groupSvc *service2.GroupService) *Handler {
	return &Handler{svc: svc, groupSvc: groupSvc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	gh := NewGroupHandler(h.groupSvc)
	r.GET("/tag-groups", gh.List)
	r.POST("/tag-groups", gh.Create)
	r.PUT("/tag-groups/reorder", gh.Reorder)
	r.PUT("/tag-groups/:id", gh.Update)
	r.DELETE("/tag-groups/:id", gh.Delete)
	r.POST("/tag-groups/:id/tags", h.CreateTag)
	r.PUT("/tag-groups/:id/tags/:tagId", h.UpdateTag)
	r.DELETE("/tag-groups/:id/tags/:tagId", h.DeleteTag)
}

func (h *Handler) CreateTag(c *gin.Context) {
	groupID, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	var req service2.TagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Create(c.Request.Context(), service2.CreateTagInGroupReq{
		Name:    req.Name,
		GroupID: groupID,
	})
	helper.Respond(c, err, tag)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	tagID, ok := helper.MustParseUintParam(c, "tagId")
	if !ok {
		return
	}
	var req service2.TagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	err := h.svc.Update(c.Request.Context(), tagID, req)
	helper.Respond(c, err, nil)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	tagID, ok := helper.MustParseUintParam(c, "tagId")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), tagID)
	helper.Respond(c, err, gin.H{"success": true})
}
