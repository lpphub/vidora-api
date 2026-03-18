package handler

import (
	"github.com/gin-gonic/gin"
	"vidora-api/modules/tag/service"
	"vidora-api/server/helper"
)

type GroupHandler struct {
	svc *service.GroupService
}

func NewGroupHandler(svc *service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.svc.List(c.Request.Context())
	helper.Respond(c, err, groups)
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req service.CreateGroupReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	group, err := h.svc.Create(c.Request.Context(), req)
	helper.Respond(c, err, group)
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateGroupReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	group, err := h.svc.Update(c.Request.Context(), id, req)
	helper.Respond(c, err, group)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id, ok := helper.MustParseUintParam(c, "id")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), id)
	helper.Respond(c, err, gin.H{"success": true})
}

func (h *GroupHandler) Reorder(c *gin.Context) {
	var req service.ReorderGroupsReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	err := h.svc.Reorder(c.Request.Context(), req.IDs)
	helper.Respond(c, err, gin.H{"success": true})
}
