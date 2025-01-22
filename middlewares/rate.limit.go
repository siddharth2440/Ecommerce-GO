package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// (Min,Max) :- (1,5)
var limiter = rate.NewLimiter(1, 5)

func Rate_lim() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Println("Rate Limiter middleware is called")
		fmt.Printf("Rate Limiter allow %v\n", limiter.Allow())
		if !limiter.Allow() {
			fmt.Println("Not Allowed")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		fmt.Println("Rate limimter is allowed")
		ctx.Next()
	}
}
