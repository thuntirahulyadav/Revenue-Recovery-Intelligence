package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
}

type clientBucket struct {
	tokens     int
	lastRefill time.Time
}

var limiter = &rateLimiter{
	clients: make(map[string]*clientBucket),
}

func RateLimitMiddleware(maxTokens int, refillRate time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter.mu.Lock()
		b, exists := limiter.clients[ip]
		now := time.Now()

		if !exists {
			limiter.clients[ip] = &clientBucket{
				tokens:     maxTokens - 1,
				lastRefill: now,
			}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		// Refill tokens
		elapsed := now.Sub(b.lastRefill)
		tokensToAdd := int(elapsed / refillRate)
		if tokensToAdd > 0 {
			b.tokens = minInt(maxTokens, b.tokens+tokensToAdd)
			b.lastRefill = now
		}

		if b.tokens <= 0 {
			limiter.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please slow down and try again shortly.",
				},
			})
			return
		}

		b.tokens--
		limiter.mu.Unlock()
		c.Next()
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
