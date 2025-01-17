package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/handlers"
	"github.com/golang/ecommerce/middlewares"
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
	userService := services.New_User_Service(db)

	// handlers
	authhandler := handlers.New_Auth_Handler(authSevice)
	userHandler := handlers.New_User_Handler(userService)

	// Public Routes  -- *** Modification ***
	publicAuthRoute := router.Group("/api/v1/auth")
	{
		publicAuthRoute.POST("/signup", authhandler.Signup)
		publicAuthRoute.POST("/login", authhandler.Login)
		publicAuthRoute.GET("/logout", authhandler.Logout)
	}

	// Private Routes
	user_private_routes := router.Group("/api/v1/user")
	user_private_routes.Use(middlewares.Chk_Auth())
	{
		user_private_routes.GET("/me", userHandler.Get_My_Profile)
	}

	// router.SetTrustedProxies([]string{"<trusted_proxy_IP_address>"})

	return router
}
