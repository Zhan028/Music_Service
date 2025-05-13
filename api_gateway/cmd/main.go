package main

import (
	"github.com/Zhan028/Music_Service/api_gateway/internal/delivery/http"
	"github.com/Zhan028/Music_Service/api_gateway/internal/grpc"

	"github.com/Zhan028/Music_Service/api_gateway/internal/logger"
	"github.com/Zhan028/Music_Service/api_gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	logger.InitLogger()

	clients := grpc.NewClients()
	handler := http.NewHandler(clients)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.GinLoggerMiddleware())

	http.RegisterRoutes(r, handler)

	logger.InfoLogger.Println("API Gateway started on :8081")
	r.Run(":8081")
}
