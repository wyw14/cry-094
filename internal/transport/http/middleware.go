package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := requestIDForStart(c.GetHeader("X-Request-ID"))
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
		if c.Writer.Status() >= http.StatusBadRequest {
			c.Header("X-Request-ID", uuid.NewString())
		}
	}
}

func requestIDForStart(candidate string) string {
	if candidate != "" {
		return candidate
	}
	return uuid.NewString()
}
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, _ any) {
		writeError(c, http.StatusInternalServerError, "PANIC", "request failed", nil)
		c.Abort()
	})
}
func Timeout(limit time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), limit)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() != nil && !c.Writer.Written() {
			writeError(c, http.StatusRequestTimeout, "REQUEST_CANCELED", "request context closed", nil)
		}
	}
}
