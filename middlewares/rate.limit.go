package middlewares

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func Rate_lim() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var limiter = rate.NewLimiter(1, 5)
		if !limiter.Allow() {
			ctx.AbortWithStatus(429) // Too Many Requests
			return
		}
		ctx.Next()
	}
}
