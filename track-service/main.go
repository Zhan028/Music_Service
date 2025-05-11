package main

import (
	"context"
	"github.com/Zhan028/Music_Service/track-service/services"
	"github.com/Zhan028/Music_Service/track-service/telemetry"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/Zhan028/Music_Service/track-service/proto"
	"github.com/Zhan028/Music_Service/track-service/repositories"
)

func main() {
	// Инициализация трейсера OpenTelemetry
	shutdown := telemetry.InitTracer("track-service")
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	// Подключение к MongoDB
	mongoURI := "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("MongoDB disconnect error: %v", err)
		}
	}()

	db := client.Database("trackdb")
	trackRepo := repositories.NewTrackRepo(db)
	trackService := services.NewTrackGRPCService(trackRepo)

	// Запуск gRPC-сервера
	grpcServer := grpc.NewServer()
	pb.RegisterTrackServiceServer(grpcServer, trackService)

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen on port 50052: %v", err)
	}

	go func() {
		log.Println("Track gRPC service is running on port 50052...")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down track gRPC service...")

	grpcServer.GracefulStop()
}
