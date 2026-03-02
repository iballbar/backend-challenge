package http

import (
	_ "backend-challenge/docs"
	"backend-challenge/internal/core/ports"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(userHandler *UserHandler, lotteryHandler *LotteryHandler, tokens ports.TokenProvider) *gin.Engine {
	router := gin.New()
	router.Use(LoggingMiddleware(), gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)

	users := router.Group("/users")
	users.Use(AuthMiddleware(tokens))
	users.GET("", userHandler.ListUsers)
	users.POST("", userHandler.CreateUser)
	users.GET("/:id", userHandler.GetUser)
	users.PATCH("/:id", userHandler.UpdateUser)
	users.DELETE("/:id", userHandler.DeleteUser)

	lottery := router.Group("/lottery")
	lottery.GET("/search", lotteryHandler.SearchLottery)

	return router
}
