package handler

import (
	"github.com/gin-gonic/gin"
	"vidora-api/modules/tag/service"
	"vidora-api/server/helper"
)

type Handler struct {
	svc      *service.Service
	groupSvc *service.GroupService
}

func NewHandler(svc *service.Service, groupSvc *service.GroupService) *Handler {
	return &Handler{svc: svc, groupSvc: groupSvc}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	gh := NewGroupHandler(h.groupSvc)
	r.GET("/tag-groups", gh.List)
	r.POST("/tag-groups", gh.Create)
	r.PUT("/tag-groups/:id", gh.Update)
	r.DELETE("/tag-groups/:id", gh.Delete)
	r.PUT("/tag-groups/reorder", gh.Reorder)
	r.POST("/tag-groups/:groupId/tags", h.CreateTag)
	r.PUT("/tag-groups/:groupId/tags/:tagId", h.UpdateTag)
	r.DELETE("/tag-groups/:groupId/tags/:tagId", h.DeleteTag)
}

func (h *Handler) CreateTag(c *gin.Context) {
	groupID, ok := helper.MustParseUintParam(c, "groupId")
	if !ok {
		return
	}
	var req service.CreateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Create(c.Request.Context(), service.CreateTagInGroupReq{
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
	var req service.UpdateTagReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	tag, err := h.svc.Update(c.Request.Context(), tagID, req)
	helper.Respond(c, err, tag)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	tagID, ok := helper.MustParseUintParam(c, "tagId")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), tagID)
	helper.Respond(c, err, gin.H{"success": true})
}
