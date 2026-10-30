package story

import (
	"encoding/json"
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
	g := rg.Group("/story")
	g.GET("/progress", h.ListProgress)
	g.PUT("/progress", h.SaveProgress)
}

func (h *Handler) ListProgress(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("unauthorized"))
		return
	}

	items, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"items": items})
}

type saveProgressRequest struct {
	ChapterID      string          `json:"chapter_id" binding:"required,max=64"`
	NodeID         string          `json:"node_id" binding:"required,max=64"`
	ContentVersion string          `json:"content_version" binding:"max=32"`
	Flags          json.RawMessage `json:"flags"`
}

func (h *Handler) SaveProgress(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Fail(c, apperr.Unauthorized("unauthorized"))
		return
	}

	var req saveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, apperr.BadRequest("invalid request body"))
		return
	}

	progress, err := h.svc.Save(c.Request.Context(), userID, SaveInput{
		ChapterID:      req.ChapterID,
		NodeID:         req.NodeID,
		ContentVersion: req.ContentVersion,
		Flags:          req.Flags,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidProgress), errors.Is(err, ErrInvalidFlags):
			httpx.Fail(c, apperr.BadRequest(err.Error()))
		case errors.Is(err, ErrProgressNotFound):
			httpx.Fail(c, apperr.NotFound("story progress not found"))
		default:
			httpx.Fail(c, err)
		}
		return
	}
	httpx.OK(c, progress)
}
