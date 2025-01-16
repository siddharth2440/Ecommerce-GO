package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/handlers"
	"github.com/golang/ecommerce/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(db *mongo.Client) *gin.Engine {
	router := gin.Default()
	// CORS Setup
	conf := cors.DefaultConfig()
	conf.AllowAllOrigins = true
	conf.AllowCredentials = true
	conf.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	router.Use(cors.New(conf))

	// services
	authSevice := services.NewAuthService(db)

	// handlers
	authhandler := handlers.New_Auth_Handler(authSevice)
	// route

	// public Routes
	{
		router.POST("/signup", authhandler.Signup)
		router.POST("/login", authhandler.Login)
	}

	// Private Routes

	// router.SetTrustedProxies([]string{"<trusted_proxy_IP_address>"})

	return router
}
