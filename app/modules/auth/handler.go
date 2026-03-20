// modules/auth/handler.go
package auth

import (
	"vidora-api/app/server/helper"

	"github.com/gin-gonic/gin"
)

// Handler 认证处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建认证处理器
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) register(r *gin.RouterGroup) {
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.RefreshToken)
}

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req AuthReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	resp, err := h.svc.Register(c, req.Email, req.Password)
	helper.Respond(c, err, resp)
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req AuthReq
	if !helper.MustBindJSON(c, &req) {
		return
	}
	resp, err := h.svc.Login(c, req.Email, req.Password)
	helper.Respond(c, err, resp)
}

// RefreshToken 刷新令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if !helper.MustBindJSON(c, &req) {
		return
	}
	resp, err := h.svc.RefreshToken(c, req.RefreshToken)
	helper.Respond(c, err, resp)
}
