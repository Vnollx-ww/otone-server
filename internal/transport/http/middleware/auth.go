package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"otome-server/internal/module/auth"
	"otome-server/internal/shared/apperr"
	"otome-server/internal/shared/httpx"
)

const bearerPrefix = "Bearer "

func Auth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			httpx.Fail(c, apperr.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}

		userID, err := tokens.ParseAccessToken(strings.TrimSpace(header[len(bearerPrefix):]))
		if err != nil {
			httpx.Fail(c, apperr.Unauthorized("invalid or expired token"))
			c.Abort()
			return
		}

		httpx.SetUserID(c, userID)
		c.Next()
	}
}
