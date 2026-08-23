package http

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type rateBucket struct {
	window time.Time
	count  int
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateBucket{}
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		bucket := buckets[key]
		if now.Sub(bucket.window) >= window {
			bucket = rateBucket{window: now}
		}
		bucket.count++
		buckets[key] = bucket
		allowed := bucket.count <= limit
		mu.Unlock()
		if !allowed {
			writeError(c, http.StatusTooManyRequests, "RATE_LIMITED", "request rate limit exceeded", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

type Metrics struct {
	requests atomic.Uint64
	errors   atomic.Uint64
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.requests.Add(1)
		c.Next()
		if c.Writer.Status() >= 500 {
			m.errors.Add(1)
		}
	}
}
func (m *Metrics) Handler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; version=0.0.4", []byte(fmt.Sprintf("scriptscope_http_requests_total %d\nscriptscope_http_errors_total %d\n", m.requests.Load(), m.errors.Load())))
}
