package auth

import (
	"errors"

	"github.com/gin-gonic/gin"

	"otome-server/internal/shared/apperr"
	"otome-server/internal/shared/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/guest", h.GuestLogin)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout)
}

type guestLoginRequest struct {
	DeviceID string `json:"device_id" binding:"required,max=128"`
	Platform string `json:"platform" binding:"required,oneof=ios android pc web wechat_mini douyin_mini"`
}

type guestLoginResponse struct {
	User   *User      `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

func (h *Handler) GuestLogin(c *gin.Context) {
	var req guestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.BadRequest("invalid request body"))
		return
	}

	user, tokens, err := h.svc.GuestLogin(c.Request.Context(), req.DeviceID, req.Platform)
	if err != nil {
		h.fail(c, err)
		return
	}
	httpx.OK(c, guestLoginResponse{User: user, Tokens: tokens})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.BadRequest("invalid request body"))
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.fail(c, err)
		return
	}
	httpx.OK(c, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"revoked": true})
}

func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidDevice):
		httpx.Fail(c, apperr.BadRequest(err.Error()))
	case errors.Is(err, ErrInvalidRefreshToken):
		httpx.Fail(c, apperr.Unauthorized("invalid refresh token"))
	case errors.Is(err, ErrUserDisabled):
		httpx.Fail(c, apperr.Forbidden("user is disabled"))
	default:
		httpx.Fail(c, err)
	}
}
