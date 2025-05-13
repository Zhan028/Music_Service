package main

import (
	"context"
	grpc2 "github.com/Zhan028/Music_Service/playlistService/internal/delivery/grpc"
	"github.com/Zhan028/Music_Service/playlistService/internal/redis"
	mongodb2 "github.com/Zhan028/Music_Service/playlistService/internal/repository/mongodb"
	"github.com/Zhan028/Music_Service/playlistService/internal/usecase"
	kafkaconsumer "github.com/Zhan028/Music_Service/playlistService/kafka"
	pb "github.com/Zhan028/Music_Service/playlistService/proto"

	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
	}

	// Получаем конфигурацию из переменных окружения
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongoUser := getEnv("MONGO_USER", "")
	mongoPass := getEnv("MONGO_PASS", "")
	mongoDBName := getEnv("MONGO_DB", "playlist_service")
	grpcPort := getEnv("GRPC_PORT", "50051")

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подключаемся к MongoDB
	log.Println("Connecting to MongoDB...")
	db, err := mongodb2.NewClient(ctx, mongoURI, mongoUser, mongoPass, mongoDBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Подключаемся к Redis
	redisClient := redis.Init()

	// Создаем репозиторий
	playlistRepo := mongodb2.NewPlaylistRepository(db)

	// Создаем use case
	playlistUseCase := usecase.NewPlaylistUseCase(playlistRepo, redisClient)

	kafkaconsumer.StartTrackConsumer(*playlistUseCase)

	// Создаем gRPC сервер
	server := grpc2.NewPlaylistServer(playlistUseCase)

	// Настраиваем gRPC сервер
	grpcServer := grpc.NewServer()
	pb.RegisterPlaylistServiceServer(grpcServer, server)

	// Включаем reflection для удобства отладки (можно использовать grpcurl)
	reflection.Register(grpcServer)

	// Запускаем gRPC сервер
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Starting gRPC server on port %s...", grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Обработка сигналов для корректного завершения работы
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
