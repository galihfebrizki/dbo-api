package main

import (
	"github.com/galihfebrizki/dbo-api/config"
	"github.com/galihfebrizki/dbo-api/internal/controllers"
	"github.com/galihfebrizki/dbo-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	healthController *controllers.HealthController,
	orderController *controllers.OrderController,
	userController *controllers.UserController,
	paymentController *controllers.PaymentController,
) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))

	if config.Get().Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r.Use(requestid.New())

	api := r.Group("/api")
	api.POST("/login", userController.Login)
	api.Use(middleware.JWTAuthMiddleware())

	// need jwt auth
	api.GET("/login", userController.GetLoginData)

	api.GET("/customer", userController.GetSelfData)
	api.GET("/customer/:userId", userController.GetCustomerData)
	api.GET("/list-customer", userController.GetListCustomerData)
	api.POST("/customer", userController.CreateUser)
	api.PUT("/customer", userController.UpdateUser)
	api.DELETE("/customer", userController.DeleteUser)
	api.GET("/search-customer", userController.SearchUser)

	api.GET("/order/:orderId", orderController.GetOrder)
	api.GET("/list-order", orderController.GetListOrder)
	api.POST("/order", orderController.CreateOrder)
	api.PUT("/order", orderController.UpdateOrder)
	api.DELETE("/order", orderController.DeleteOrder)
	api.GET("/search-order", orderController.SearchOrder)

	api.POST("/payment-order", paymentController.PaymentOrder)

	// free access
	r.GET("/health", healthController.Health)

	return r

}
