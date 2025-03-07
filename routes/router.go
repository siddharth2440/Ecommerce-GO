package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang/ecommerce/handlers"
	"github.com/golang/ecommerce/middlewares"
	"github.com/golang/ecommerce/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Setup Prometheus
	middlewares.PrometheusInit()

	// services
	authSevice := services.NewAuthService(db)
	userService := services.New_User_Service(db)
	productService := services.NewProductService(db)
	cartService := services.New_Cart_Service(db)
	orderService := services.NewOrderService(db)

	// handlers
	authhandler := handlers.New_Auth_Handler(authSevice)
	userHandler := handlers.New_User_Handler(userService)
	productHandler := handlers.New_Product_Handler(productService)
	cartHandler := handlers.New_Cart_Handler(cartService)
	orderHandler := handlers.NewOrderService(orderService)

	// Prometheus Metrics Endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Middleware to track Metrics
	router.Use(middlewares.TrackMetrics())

	// Public Routes  -- *** Modification ***
	publicAuthRoute := router.Group("/api/v1/auth")
	publicAuthRoute.Use(middlewares.Rate_lim())
	{
		publicAuthRoute.POST("/signup", authhandler.Signup)
		publicAuthRoute.POST("/login", authhandler.Login)
		publicAuthRoute.GET("/logout", authhandler.Logout)
	}

	// Private Routes
	user_private_routes := router.Group("/api/v1/user")
	user_private_routes.Use(middlewares.Chk_Auth())
	user_private_routes.Use(middlewares.Rate_lim())
	{
		user_private_routes.GET("/me", userHandler.Get_My_Profile)
		user_private_routes.PUT("/update_me", userHandler.Update_USER_Profile)
		user_private_routes.DELETE("/delete_me", userHandler.Delete_User_Profile)
		user_private_routes.GET("/user_info/:userID", userHandler.GET_USER_FROM_USERID)

		// TODO := to fix and Work this and also chk this
		user_private_routes.GET("/random_users", userHandler.GET_RANDOM_USERS)
		user_private_routes.GET("/recent_users", userHandler.GET_RECENT_USERS)
		user_private_routes.GET("/query_user", userHandler.SEARCH_FOR_USERS)
	}

	// Product Routes
	public_product_routes := router.Group("/api/v1/product")
	public_product_routes.Use(middlewares.Rate_lim())
	{
		// public_product_routes.GET("/query_product")
		public_product_routes.GET("/query", productHandler.Product_By_Query)
		// product information by product ID
		public_product_routes.GET("/:productId", productHandler.Get_Product_Details_By_ID)
		// get latest products
		public_product_routes.GET("/latest", productHandler.Latest_Products)
		// get random 2 or more products
	}

	private_product_routes := router.Group("/api/v1/products")
	private_product_routes.Use(middlewares.Chk_Auth())
	private_product_routes.Use(middlewares.Rate_lim())

	{
		// Create a product
		private_product_routes.POST("/add-product", productHandler.Add_Product_Handler)
		// Update the Product
		private_product_routes.PUT("/update-product/:productId", productHandler.Update_Product_Details)
		// Delete the Product
		private_product_routes.DELETE("/delete-product/:productId", productHandler.Delete_Products)
	}
	// router.SetTrustedProxies([]string{"<trusted_proxy_IP_address>"})

	// Cart
	cart_routes := router.Group("/api/v1/cart")
	cart_routes.Use(middlewares.Chk_Auth())
	cart_routes.Use(middlewares.Rate_lim())
	{
		cart_routes.POST("/add-to-cart", cartHandler.Create_Cart_Handler)
		cart_routes.GET("/my-cart", cartHandler.Get_My_Cart)
		cart_routes.DELETE("/delete-my-cart/:cartId", cartHandler.Delete_Cart_Handler)
		cart_routes.GET("/all-carts", cartHandler.Get_Carts_handler)
		// update cart details
		cart_routes.PUT("/update-cart/:cartID", cartHandler.Update_Cart_handler)
	}

	order_Routes := router.Group("/api/v1/orders")
	order_Routes.Use(middlewares.Chk_Auth())
	order_Routes.Use(middlewares.Rate_lim())

	{
		order_Routes.POST("/create-order", orderHandler.Create_Order_Handler)
		order_Routes.GET("/user-orders", orderHandler.Get_User_Orders_Handler)
		order_Routes.GET("/orders", orderHandler.Get_Orders_Handler)
		order_Routes.DELETE("/order/:orderId", orderHandler.Delete_Order_Handler)
		order_Routes.PUT("/order/:orderId", orderHandler.Update_Order_Handler)
	}

	return router
}
