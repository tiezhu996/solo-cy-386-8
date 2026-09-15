package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/util"
)

type bucket struct {
	count int
	until time.Time
}

// RateLimit 简单内存限流中间件：每 IP 每分钟最多 limit 次请求。
func RateLimit(limit int) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*bucket)
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		mu.Lock()
		b, ok := buckets[key]
		if !ok || now.After(b.until) {
			buckets[key] = &bucket{count: 1, until: now.Add(time.Minute)}
			mu.Unlock()
			c.Next()
			return
		}
		b.count++
		if b.count > limit {
			mu.Unlock()
			util.Fail(c, 429, constants.CodeBadRequest, "请求过于频繁，请稍后再试")
			return
		}
		mu.Unlock()
		c.Next()
	}
}
