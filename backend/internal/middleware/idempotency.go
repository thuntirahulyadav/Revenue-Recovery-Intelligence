package middleware

import (
	"net/http"
	"time"

	"razorpay-recovery-intelligence/backend/internal/redis"

	"github.com/gin-gonic/gin"
)

func IdempotencyMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check mutations (POST, PUT, PATCH)
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			key = c.GetHeader("X-Idempotency-Key")
		}

		if key != "" {
			acquired, err := redisClient.SetIdempotencyKey(c.Request.Context(), key, 5*time.Minute)
			if err == nil && !acquired {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "DUPLICATE_REQUEST",
						"message": "Duplicate request detected with identical idempotency key. Request is already processed or in progress.",
					},
				})
				return
			}
		}

		c.Next()
	}
}
