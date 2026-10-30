package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"otome-server/internal/shared/httpx"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" || len(id) > 64 {
			id = uuid.NewString()
		}
		httpx.SetRequestID(c, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
