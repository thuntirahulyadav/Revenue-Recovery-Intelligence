package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Idempotency-Key", "X-Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length", "Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
