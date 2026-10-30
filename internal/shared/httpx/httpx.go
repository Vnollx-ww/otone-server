package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"otome-server/internal/shared/apperr"
)

const (
	contextRequestID = "request_id"
	contextUserID    = "user_id"
)

// Envelope is the uniform JSON shape returned by every API endpoint.
type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func SetRequestID(c *gin.Context, id string) { c.Set(contextRequestID, id) }

func RequestID(c *gin.Context) string {
	if value, ok := c.Get(contextRequestID); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func SetUserID(c *gin.Context, userID string) { c.Set(contextUserID, userID) }

func UserID(c *gin.Context) (string, bool) {
	value, ok := c.Get(contextUserID)
	if !ok {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok && userID != ""
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data, RequestID: RequestID(c)})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Code: 0, Message: "ok", Data: data, RequestID: RequestID(c)})
}

// Fail writes an error response. Unknown/internal errors are attached to the
// gin context so the access log middleware records their details, while the
// client only sees a generic message.
func Fail(c *gin.Context, err error) {
	appErr := apperr.As(err)
	if appErr.Status >= http.StatusInternalServerError {
		_ = c.Error(err)
	}
	c.JSON(appErr.Status, Envelope{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: RequestID(c),
	})
}
