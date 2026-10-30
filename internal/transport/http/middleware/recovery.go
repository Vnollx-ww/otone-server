package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"otome-server/internal/shared/apperr"
	"otome-server/internal/shared/httpx"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			"request_id", httpx.RequestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, httpx.Envelope{
			Code:      apperr.ErrInternal.Code,
			Message:   apperr.ErrInternal.Message,
			RequestID: httpx.RequestID(c),
		})
	})
}
