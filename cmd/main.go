package main

import (
	"context"
	"github.com/Zhan028/Music_Service/internal/repository/mongodb"
	"github.com/Zhan028/Music_Service/internal/usecase"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc2 "github.com/Zhan028/Music_Service/internal/delivery/grpc"
	"github.com/joho/godotenv"

	pb "github.com/Zhan028/Music_Service/proto"
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
	db, err := mongodb.NewClient(ctx, mongoURI, mongoUser, mongoPass, mongoDBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Создаем репозиторий
	playlistRepo := mongodb.NewPlaylistRepository(db)

	// Создаем use case
	playlistUseCase := usecase.NewPlaylistUseCase(playlistRepo)

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
