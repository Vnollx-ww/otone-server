package player

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
	g := rg.Group("/player")
	g.GET("/me", h.Me)
	g.PATCH("/me", h.UpdateMe)
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("unauthorized"))
		return
	}

	profile, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil {
		h.fail(c, err)
		return
	}
	httpx.OK(c, profile)
}

type updateMeRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.BadRequest("invalid request body"))
		return
	}

	profile, err := h.svc.UpdateNickname(c.Request.Context(), userID, req.Nickname)
	if err != nil {
		h.fail(c, err)
		return
	}
	httpx.OK(c, profile)
}

func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPlayerNotFound):
		httpx.Fail(c, apperr.NotFound("player not found"))
	case errors.Is(err, ErrInvalidNickname):
		httpx.Fail(c, apperr.BadRequest(err.Error()))
	default:
		httpx.Fail(c, err)
	}
}
