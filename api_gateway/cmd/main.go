package main

import (
	"github.com/Zhan028/Music_Service/api_gateway/internal/delivery/http"
	"github.com/Zhan028/Music_Service/api_gateway/internal/grpc"
	"github.com/gin-gonic/gin"
)

func main() {
	clients := grpc.NewClients()
	handler := http.NewHandler(clients)

	r := gin.Default()
	http.RegisterRoutes(r, handler)

	r.Run(":8080") // API Gateway слушает на порту 8080
}
