package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/ecommerce/utils"
)

func Chk_Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie_string, err := ctx.Cookie("authCookie_golang")
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		if cookie_string == "" {
			ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		// Verify The Token which is stored in our Cookie
		token, err := utils.JWT_Verification(cookie_string)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		fmt.Println("Claims :-")
		fmt.Println(claims)
		if !ok || !token.Valid {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		// map[exp:1.73735283e+09 id:6789ef3f0747916de714e421 isAdmin:false username:itsmonday]
		ctx.Set("userId", claims["id"])
		ctx.Set("isAdmin", claims["isAdmin"])
		ctx.Set("username", claims["username"])
		ctx.Next()
	}
}
