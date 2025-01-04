package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	// router.SetTrustedProxies([]string{"<trusted_proxy_IP_address>"})

	return router
}
